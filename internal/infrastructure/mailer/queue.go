package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	jobStatusPending = "pending"
	jobStatusSent    = "sent"
	jobStatusFailed  = "failed"
	maxAttempts      = 4 // 3 retries (1 min, 5 min, 15 min) + 1 initial attempt
	chanBuffer       = 200
	workerCount      = 5 // fixed SMTP concurrency limit
)

// retryDelays[attempt-1] is how long to wait before the next try.
// All three delays are used: attempts 1→2 wait 1 min, 2→3 wait 5 min, 3→4 wait 15 min.
var retryDelays = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

// emailJobRecord is the DB row for a queued email (infrastructure-only, not a domain entity).
type emailJobRecord struct {
	ID              string `gorm:"primaryKey"`
	To              string `gorm:"column:to;not null"` // explicit column name: 'to' is a reserved SQL keyword
	Subject         string `gorm:"not null"`
	Template        string `gorm:"not null"`
	DataJSON        string `gorm:"type:text;not null"`
	Status          string `gorm:"default:'pending'"`
	Attempts        int    `gorm:"default:0"`
	LastError       string
	ScheduledAt     time.Time `gorm:"not null"`
	CreatedAt       time.Time
	LastAttemptedAt *time.Time
}

func (emailJobRecord) TableName() string { return "email_jobs" }

type queuedJob struct {
	record emailJobRecord
	data   map[string]any
}

// EmailQueue wraps Mailer with async delivery, DB persistence, and retry.
// Implements EmailSender — swap it in wherever *Mailer is used.
type EmailQueue struct {
	db   *gorm.DB
	jobs chan queuedJob
	log  *zap.Logger
	// sendFunc is the actual SMTP call. Overridable in tests.
	sendFunc func(to, subject, template string, data map[string]any) error
	// inFlight tracks IDs of jobs currently being processed so loadPending skips them.
	inFlight sync.Map
}

// NewEmailQueue creates an EmailQueue backed by the given Mailer for actual SMTP delivery.
//
// Connection to mailer.go: m.send (defined in mailer.go) is stored as sendFunc. Every
// subsequent call to process() will invoke that method value — EmailQueue never imports
// SMTP details directly; it delegates through this single captured reference.
//
// Call Start before sending any emails.
func NewEmailQueue(m *Mailer, db *gorm.DB, log *zap.Logger) *EmailQueue {
	return &EmailQueue{
		db:       db,
		jobs:     make(chan queuedJob, chanBuffer),
		log:      log,
		sendFunc: m.send,
	}
}

// Start is the startup sequence — call it once after NewEmailQueue, before any sends.
//
// Launches workerCount fixed goroutines that each block on the jobs channel
// (capping SMTP concurrency to workerCount), plus one poller goroutine that
// runs the 5-minute DB recovery tick.
// Cancel ctx to stop all goroutines gracefully; in-flight sends finish first.
func (q *EmailQueue) Start(ctx context.Context) {
	q.loadPending(ctx)
	for range workerCount {
		go q.worker(ctx)
	}
	go q.poller(ctx)
}

// SendPasswordReset implements the EmailSender interface (defined in mailer.go).
//
// This is the first call in the delivery chain from the application layer.
// It formats the template data and hands off to enqueue — the caller returns
// immediately without waiting for SMTP.
func (q *EmailQueue) SendPasswordReset(to, name, otp string) error {
	return q.enqueue(to, "Password Reset", "reset_password.page.tmpl", map[string]any{
		"User": name, "OTP": otp, "ExpiresIn": "5",
	})
}

// SendEmailVerification implements the EmailSender interface (defined in mailer.go).
//
// Mirrors SendPasswordReset — same non-blocking contract, different template and data.
func (q *EmailQueue) SendEmailVerification(to, name, link string) error {
	return q.enqueue(to, "Verify Your Email", "verify_email.page.tmpl", map[string]any{
		"User": name, "VerificationLink": link, "ExpiresIn": "24 hours",
	})
}

// enqueue is the second step in the delivery chain, called by SendPasswordReset /
// SendEmailVerification. It does two things before returning to the caller:
//  1. Persist — writes an emailJobRecord to the DB so the job survives a crash.
//  2. Push — sends the job to the in-memory channel so worker/process can act on it
//     immediately. If the channel is full the job is already safe in the DB and will
//     be picked up by the next loadPending poll.
//
// Returns an error only if both the DB write and the channel push fail — in that case
// the job would be silently lost and the caller must decide whether to retry.
func (q *EmailQueue) enqueue(to, subject, template string, data map[string]any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling email data: %w", err)
	}

	rec := emailJobRecord{
		ID:          uuid.NewString(),
		To:          to,
		Subject:     subject,
		Template:    template,
		DataJSON:    string(raw),
		Status:      jobStatusPending,
		ScheduledAt: time.Now(),
	}

	var dbOK bool
	if q.db != nil {
		if err := q.db.WithContext(context.Background()).Create(&rec).Error; err != nil {
			q.log.Error("failed to persist email job",
				zap.String("to", to), zap.Error(err))
		} else {
			dbOK = true
		}
	}

	select {
	case q.jobs <- queuedJob{record: rec, data: data}:
	default:
		if !dbOK {
			// No DB record and no channel slot — the job would be permanently lost.
			return fmt.Errorf("email queue full and job could not be persisted for %s", to)
		}
		q.log.Warn("email channel full — job persisted in DB and will be retried by poll",
			zap.String("to", to))
	}
	return nil
}

// worker is one member of the fixed pool started by Start.
// It blocks on the jobs channel, processes one email at a time, and exits when ctx is canceled.
// Concurrency is bounded to workerCount across the whole pool — no per-job goroutines are spawned.
func (q *EmailQueue) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-q.jobs:
			q.inFlight.Store(job.record.ID, struct{}{})
			q.process(ctx, &job)
			q.inFlight.Delete(job.record.ID)
		}
	}
}

// poller runs the 5-minute safety-net tick that calls loadPending.
// Separated from worker so the fixed pool goroutines are never blocked by a DB poll.
func (q *EmailQueue) poller(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.loadPending(ctx)
		}
	}
}

// process is the core of the delivery chain — the step where the email actually leaves
// the server. It is called in a goroutine by worker for every job pulled from the channel.
//
// Connection to mailer.go: q.sendFunc is the method value Mailer.send captured at
// construction time (see NewEmailQueue). Calling q.sendFunc here is therefore identical
// to calling m.send in mailer.go — it renders the HTML template and dials the Brevo
// SMTP server. mailer.go:send logs "email sent" on success; this function does not
// duplicate that log.
//
// Retry logic:
//   - Success → status = "sent", DB updated.
//   - Failure, attempts < maxAttempts → status stays "pending", ScheduledAt pushed
//     forward by retryDelays[attempt-1]. A goroutine re-queues the job after that delay.
//   - Failure, attempts == maxAttempts → status = "failed", logged as permanent failure.
func (q *EmailQueue) process(ctx context.Context, job *queuedJob) {
	rec := job.record
	now := time.Now()

	err := q.sendFunc(rec.To, rec.Subject, rec.Template, job.data)
	rec.Attempts++
	rec.LastAttemptedAt = &now

	if err == nil {
		rec.Status = jobStatusSent
		rec.LastError = ""
		// mailer.go:send already logs "email sent"; no duplicate log here.
	} else {
		rec.LastError = err.Error()
		q.log.Warn("email send failed",
			zap.String("to", rec.To),
			zap.String("subject", rec.Subject),
			zap.Int("attempt", rec.Attempts),
			zap.Error(err))

		if rec.Attempts < maxAttempts {
			delay := retryDelays[rec.Attempts-1]
			rec.Status = jobStatusPending
			rec.ScheduledAt = now.Add(delay)

			// Re-queue after the delay in a goroutine so the worker stays unblocked.
			go func(j queuedJob, d time.Duration) {
				timer := time.NewTimer(d)
				defer timer.Stop()
				select {
				case <-ctx.Done():
				case <-timer.C:
					select {
					case q.jobs <- j:
					default:
						// Channel full — DB poll will pick it up when ScheduledAt is due.
					}
				}
			}(queuedJob{record: rec, data: job.data}, delay)
		} else {
			rec.Status = jobStatusFailed
			q.log.Error("email permanently failed after max attempts",
				zap.String("to", rec.To),
				zap.String("subject", rec.Subject),
				zap.String("last_error", rec.LastError))
		}
	}

	if q.db != nil {
		if err := q.db.WithContext(ctx).
			Model(&emailJobRecord{}).
			Where("id = ?", rec.ID).
			Updates(map[string]any{
				"status":            rec.Status,
				"attempts":          rec.Attempts,
				"last_error":        rec.LastError,
				"last_attempted_at": rec.LastAttemptedAt,
				"scheduled_at":      rec.ScheduledAt,
			}).Error; err != nil {
			q.log.Error("failed to update email job status", zap.String("id", rec.ID), zap.Error(err))
		}
	}
}

// loadPending is the DB-backed recovery path. It is called at startup (via Start) and
// every 5 minutes by worker's ticker.
//
// It queries for rows with status = 'pending' AND scheduled_at <= now — this covers:
//   - Jobs not yet attempted (enqueued before a restart when the channel was full).
//   - Retry jobs whose delay has elapsed (process set their ScheduledAt into the future).
//
// Jobs whose IDs are already tracked in inFlight are skipped to prevent double delivery.
// Results are capped at chanBuffer rows so a large backlog does not spike memory.
// If the channel is full the remainder will be caught by the next tick — no job is lost.
func (q *EmailQueue) loadPending(ctx context.Context) {
	if q.db == nil {
		return
	}
	var records []emailJobRecord
	if err := q.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= ?", jobStatusPending, time.Now()).
		Limit(chanBuffer).
		Find(&records).Error; err != nil {
		q.log.Error("loading pending email jobs from DB", zap.Error(err))
		return
	}

	for i := range records {
		if _, ok := q.inFlight.Load(records[i].ID); ok {
			continue // already being processed; skip to avoid double delivery
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(records[i].DataJSON), &data); err != nil {
			q.log.Error("unmarshaling email job data",
				zap.String("id", records[i].ID), zap.Error(err))
			continue
		}
		select {
		case q.jobs <- queuedJob{record: records[i], data: data}:
		default:
			return // Channel full; next poll will catch the remainder.
		}
	}
}
