package mailer

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// newTestQueue builds an EmailQueue with no DB and a custom send function.
// Retries are shortened so tests don't wait minutes.
func newTestQueue(sendFn func(to, subject, template string, data map[string]any) error) *EmailQueue {
	log, _ := zap.NewDevelopment()
	return &EmailQueue{
		db:       nil, // no DB needed for unit tests
		jobs:     make(chan queuedJob, chanBuffer),
		log:      log,
		sendFunc: sendFn,
	}
}

func TestEmailQueue_SucceedsOnSecondAttempt(t *testing.T) {
	// Shorten retry delays so the test completes in milliseconds.
	original := retryDelays
	retryDelays = []time.Duration{10 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond}
	defer func() { retryDelays = original }()

	var calls atomic.Int32

	q := newTestQueue(func(to, subject, template string, data map[string]any) error {
		n := calls.Add(1)
		if n == 1 {
			return errors.New("smtp: connection refused") // first attempt fails
		}
		return nil // second attempt succeeds
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q.Start(ctx)

	if err := q.SendPasswordReset("user@example.com", "Test User", "123456"); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	// Wait for both attempts to complete.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if calls.Load() >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if got := calls.Load(); got < 2 {
		t.Fatalf("expected at least 2 send attempts, got %d", got)
	}
	t.Logf("send called %d time(s) — failed once, succeeded on retry ✓", calls.Load())
}

func TestEmailQueue_PermanentlyFailsAfterMaxAttempts(t *testing.T) {
	original := retryDelays
	retryDelays = []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond}
	defer func() { retryDelays = original }()

	var calls atomic.Int32

	q := newTestQueue(func(to, subject, template string, data map[string]any) error {
		calls.Add(1)
		return errors.New("smtp: always failing")
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q.Start(ctx)

	if err := q.SendEmailVerification("user@example.com", "Test User", "https://example.com/verify"); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if calls.Load() >= int32(maxAttempts) {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if got := calls.Load(); got != int32(maxAttempts) {
		t.Fatalf("expected exactly %d attempts, got %d", maxAttempts, got)
	}
	t.Logf("send called %d time(s) — all failed, gave up after max attempts ✓", calls.Load())
}
