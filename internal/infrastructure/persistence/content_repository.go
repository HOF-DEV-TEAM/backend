package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	domainContent "bitbucket.org/hofng/hofApp/internal/domain/content"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
)

type contentRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// NewContentRepository returns a GORM-backed implementation of content.Repository.
func NewContentRepository(db *gorm.DB, log *zap.Logger) domainContent.Repository {
	return &contentRepository{db: db, log: log}
}

// ── Audio messages ────────────────────────────────────────────────────────────

func (r *contentRepository) CreateMessage(ctx context.Context, m *domainContent.AudioMessage) error {
	if result := r.db.WithContext(ctx).Create(m); result.Error != nil {
		if isUniqueViolation(result.Error) {
			// Try to map unique constraint to a specific field. If the DB error mentions audio_url
			// prefer returning an ErrAlreadyExists for the audio_url field; otherwise fall back to title.
			errMsg := result.Error.Error()
			if strings.Contains(errMsg, "audio_url") || strings.Contains(errMsg, "idx_audio_messages_audio_url") {
				return shared.ErrAlreadyExists{Resource: "audio message", Field: "audio_url", Value: m.AudioURL}
			}
			return shared.ErrAlreadyExists{Resource: "audio message", Field: "title", Value: m.Title}
		}
		return fmt.Errorf("creating audio message: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) GetMessageByAudioURL(ctx context.Context, audioURL string) (*domainContent.AudioMessage, error) {
	var m domainContent.AudioMessage
	result := r.db.WithContext(ctx).Preload("Series").
		Where("deleted_at IS NULL").
		First(&m, "audio_url = ?", audioURL)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "audio message", ID: audioURL}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting message by audio_url: %w", result.Error)
	}
	return &m, nil
}

func (r *contentRepository) GetMessages(ctx context.Context, filter domainContent.MessageFilter) ([]domainContent.AudioMessage, int64, error) {
	q := r.db.WithContext(ctx).Model(&domainContent.AudioMessage{}).
		Where("deleted_at IS NULL")

	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		q = q.Where("title ILIKE ? OR author ILIKE ?", pattern, pattern)
	}
	if filter.SeriesID != nil {
		q = q.Where("series_id = ?", filter.SeriesID)
	}
	if len(filter.AccessIn) > 0 {
		q = q.Where("access_level IN ?", filter.AccessIn)
	}
	if filter.ExcludePrivate {
		q = q.Where("is_private = false")
	}
	if filter.IsFree != nil {
		q = q.Where("is_free = ?", *filter.IsFree)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting messages: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	var messages []domainContent.AudioMessage
	result := q.Preload("Series").
		Order("date_added DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&messages)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("listing messages: %w", result.Error)
	}

	return messages, total, nil
}

func (r *contentRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*domainContent.AudioMessage, error) {
	var m domainContent.AudioMessage
	result := r.db.WithContext(ctx).Preload("Series").
		Where("deleted_at IS NULL").
		First(&m, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "audio message", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting message by id: %w", result.Error)
	}
	return &m, nil
}

func (r *contentRepository) UpdateMessage(ctx context.Context, m *domainContent.AudioMessage) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(m).Updates(map[string]any{
		"title":         m.Title,
		"author":        m.Author,
		"audio_url":     m.AudioURL,
		"image_url":     m.ImageURL,
		"description":   m.Description,
		"is_free":       m.IsFree,
		"access_level":  m.AccessLevel,
		"is_private":    m.IsPrivate,
		"series_id":     m.SeriesID,
		"date_released": m.DateReleased,
		"last_updated":  now,
	})
	if result.Error != nil {
		return fmt.Errorf("updating audio message: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) SoftDeleteMessage(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domainContent.AudioMessage{}).
		Where("id = ?", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("soft-deleting audio message: %w", result.Error)
	}
	return nil
}

// ── Audio series ──────────────────────────────────────────────────────────────

func (r *contentRepository) CreateSeries(ctx context.Context, s *domainContent.AudioSeries) error {
	if result := r.db.WithContext(ctx).Create(s); result.Error != nil {
		if isUniqueViolation(result.Error) {
			return shared.ErrAlreadyExists{Resource: "audio series", Field: "title", Value: s.Title}
		}
		return fmt.Errorf("creating audio series: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) GetAllSeries(ctx context.Context) ([]domainContent.AudioSeries, int64, error) {
	var series []domainContent.AudioSeries
	q := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("date_added DESC")

	var total int64
	if err := q.Model(&domainContent.AudioSeries{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting series: %w", err)
	}

	if result := q.Find(&series); result.Error != nil {
		return nil, 0, fmt.Errorf("listing series: %w", result.Error)
	}
	return series, total, nil
}

func (r *contentRepository) GetSeriesByID(ctx context.Context, id uuid.UUID) (*domainContent.AudioSeries, error) {
	var s domainContent.AudioSeries
	result := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Preload("Messages", "deleted_at IS NULL").
		First(&s, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "audio series", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting series by id: %w", result.Error)
	}
	return &s, nil
}

func (r *contentRepository) UpdateSeries(ctx context.Context, s *domainContent.AudioSeries) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(s).Updates(map[string]any{
		"title":         s.Title,
		"author":        s.Author,
		"image_url":     s.ImageURL,
		"description":   s.Description,
		"of_the_month":  s.OfTheMonth,
		"date_released": s.DateReleased,
		"last_updated":  now,
	})
	if result.Error != nil {
		return fmt.Errorf("updating audio series: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) SoftDeleteSeries(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domainContent.AudioSeries{}).
		Where("id = ?", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("soft-deleting audio series: %w", result.Error)
	}
	return nil
}

// ── Meditations ───────────────────────────────────────────────────────────────

func (r *contentRepository) CreateMeditation(ctx context.Context, m *domainContent.Meditation) error {
	if result := r.db.WithContext(ctx).Create(m); result.Error != nil {
		if isUniqueViolation(result.Error) {
			return shared.ErrAlreadyExists{Resource: "meditation", Field: "name", Value: m.Name}
		}
		return fmt.Errorf("creating meditation: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) GetMeditations(ctx context.Context, includeDeleted bool) ([]domainContent.Meditation, error) {
	q := r.db.WithContext(ctx)
	if !includeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	var meditations []domainContent.Meditation
	if result := q.Find(&meditations); result.Error != nil {
		return nil, fmt.Errorf("listing meditations: %w", result.Error)
	}
	return meditations, nil
}

func (r *contentRepository) GetMeditationByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*domainContent.Meditation, error) {
	q := r.db.WithContext(ctx)
	if !includeDeleted {
		q = q.Where("deleted_at IS NULL")
	}
	var m domainContent.Meditation
	result := q.First(&m, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, shared.ErrNotFound{Resource: "meditation", ID: id.String()}
	}
	if result.Error != nil {
		return nil, fmt.Errorf("getting meditation by id: %w", result.Error)
	}
	return &m, nil
}

func (r *contentRepository) UpdateMeditation(ctx context.Context, m *domainContent.Meditation) error {
	result := r.db.WithContext(ctx).Model(m).Updates(map[string]any{
		"name":   m.Name,
		"image":  m.Image,
		"text":   m.Text,
		"status": m.Status,
	})
	if result.Error != nil {
		return fmt.Errorf("updating meditation: %w", result.Error)
	}
	return nil
}

func (r *contentRepository) SoftDeleteMeditation(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&domainContent.Meditation{}).
		Where("id = ?", id).
		Update("deleted_at", now)
	if result.Error != nil {
		return fmt.Errorf("soft-deleting meditation: %w", result.Error)
	}
	return nil
}

// ── Homepage ──────────────────────────────────────────────────────────────────

func (r *contentRepository) GetHomepage(ctx context.Context) (*domainContent.Homepage, error) {
	var series []domainContent.AudioSeries
	if result := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND of_the_month = true").
		Order("date_released DESC").
		Limit(10).
		Find(&series); result.Error != nil {
		return nil, fmt.Errorf("loading homepage series: %w", result.Error)
	}

	meditations, err := r.GetMeditations(ctx, false)
	if err != nil {
		return nil, err
	}

	return &domainContent.Homepage{
		Series:      series,
		Meditations: meditations,
	}, nil
}
