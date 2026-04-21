package content

import (
	"context"
	"fmt"
	"time"

	domainContent "bitbucket.org/hofng/hofApp/internal/domain/content"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var validate = validator.New()

// Service exposes all content-management use cases.
type Service interface {
	// Messages
	CreateMessage(ctx context.Context, req CreateMessageRequest) (*domainContent.AudioMessage, error)
	ListMessages(ctx context.Context, filter MessageListFilter) ([]domainContent.AudioMessage, int64, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*domainContent.AudioMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, req UpdateMessageRequest) (*domainContent.AudioMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error

	// Series
	CreateSeries(ctx context.Context, req CreateSeriesRequest) (*domainContent.AudioSeries, error)
	ListSeries(ctx context.Context) ([]domainContent.AudioSeries, int64, error)
	GetSeries(ctx context.Context, id uuid.UUID) (*domainContent.AudioSeries, error)
	UpdateSeries(ctx context.Context, id uuid.UUID, req UpdateSeriesRequest) (*domainContent.AudioSeries, error)
	DeleteSeries(ctx context.Context, id uuid.UUID) error

	// Meditations
	CreateMeditation(ctx context.Context, req CreateMeditationRequest) (*domainContent.Meditation, error)
	ListMeditations(ctx context.Context, admin bool) ([]domainContent.Meditation, error)
	GetMeditation(ctx context.Context, id uuid.UUID, admin bool) (*domainContent.Meditation, error)
	UpdateMeditation(ctx context.Context, id uuid.UUID, req UpdateMeditationRequest) (*domainContent.Meditation, error)
	DeleteMeditation(ctx context.Context, id uuid.UUID) error

	// Homepage
	GetHomepage(ctx context.Context) (*domainContent.Homepage, error)
}

type contentService struct {
	repo domainContent.Repository
	log  *zap.Logger
}

// NewService creates the content application service.
func NewService(repo domainContent.Repository, log *zap.Logger) Service {
	return &contentService{repo: repo, log: log}
}

// ── Messages ──────────────────────────────────────────────────────────────────

func (s *contentService) CreateMessage(ctx context.Context, req CreateMessageRequest) (*domainContent.AudioMessage, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	m := &domainContent.AudioMessage{
		Title:        req.Title,
		Author:       req.Author,
		AudioURL:     req.AudioURL,
		ImageURL:     req.ImageURL,
		Description:  req.Description,
		IsFree:       req.IsFree,
		AllowSteward: req.AllowSteward,
	}

	if req.SeriesID != "" {
		sid, err := uuid.Parse(req.SeriesID)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "series_id", Message: "invalid UUID"}
		}
		m.SeriesID = &sid
	}

	if req.DateReleased != "" {
		t, err := time.Parse(time.RFC3339, req.DateReleased)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "date_released", Message: "must be RFC3339"}
		}
		m.DateReleased = &t
	}

	if err := s.repo.CreateMessage(ctx, m); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}
	return m, nil
}

func (s *contentService) ListMessages(ctx context.Context, f MessageListFilter) ([]domainContent.AudioMessage, int64, error) {
	filter := domainContent.MessageFilter{
		Search:   f.Search,
		IsFree:   f.IsFree,
		Page:     f.Page,
		PageSize: f.PageSize,
	}
	if f.SeriesID != "" {
		sid, err := uuid.Parse(f.SeriesID)
		if err == nil {
			filter.SeriesID = &sid
		}
	}
	return s.repo.GetMessages(ctx, filter)
}

func (s *contentService) GetMessage(ctx context.Context, id uuid.UUID) (*domainContent.AudioMessage, error) {
	return s.repo.GetMessageByID(ctx, id)
}

func (s *contentService) UpdateMessage(ctx context.Context, id uuid.UUID, req UpdateMessageRequest) (*domainContent.AudioMessage, error) {
	m, err := s.repo.GetMessageByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		m.Title = req.Title
	}
	if req.Author != "" {
		m.Author = req.Author
	}
	if req.AudioURL != "" {
		m.AudioURL = req.AudioURL
	}
	if req.ImageURL != "" {
		m.ImageURL = req.ImageURL
	}
	if req.Description != "" {
		m.Description = req.Description
	}
	if req.IsFree != nil {
		m.IsFree = *req.IsFree
	}
	if req.AllowSteward != nil {
		m.AllowSteward = *req.AllowSteward
	}
	if req.SeriesID != "" {
		sid, err := uuid.Parse(req.SeriesID)
		if err == nil {
			m.SeriesID = &sid
		}
	}
	if req.DateReleased != "" {
		t, err := time.Parse(time.RFC3339, req.DateReleased)
		if err == nil {
			m.DateReleased = &t
		}
	}

	if err := s.repo.UpdateMessage(ctx, m); err != nil {
		return nil, fmt.Errorf("update message: %w", err)
	}
	return m, nil
}

func (s *contentService) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteMessage(ctx, id)
}

// ── Series ────────────────────────────────────────────────────────────────────

func (s *contentService) CreateSeries(ctx context.Context, req CreateSeriesRequest) (*domainContent.AudioSeries, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	series := &domainContent.AudioSeries{
		Title:       req.Title,
		Author:      req.Author,
		ImageURL:    req.ImageURL,
		Description: req.Description,
		OfTheMonth:  req.OfTheMonth,
	}

	if req.DateReleased != "" {
		t, err := time.Parse(time.RFC3339, req.DateReleased)
		if err == nil {
			series.DateReleased = &t
		}
	}

	if err := s.repo.CreateSeries(ctx, series); err != nil {
		return nil, fmt.Errorf("create series: %w", err)
	}
	return series, nil
}

func (s *contentService) ListSeries(ctx context.Context) ([]domainContent.AudioSeries, int64, error) {
	return s.repo.GetAllSeries(ctx)
}

func (s *contentService) GetSeries(ctx context.Context, id uuid.UUID) (*domainContent.AudioSeries, error) {
	return s.repo.GetSeriesByID(ctx, id)
}

func (s *contentService) UpdateSeries(ctx context.Context, id uuid.UUID, req UpdateSeriesRequest) (*domainContent.AudioSeries, error) {
	series, err := s.repo.GetSeriesByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		series.Title = req.Title
	}
	if req.Author != "" {
		series.Author = req.Author
	}
	if req.ImageURL != "" {
		series.ImageURL = req.ImageURL
	}
	if req.Description != "" {
		series.Description = req.Description
	}
	if req.OfTheMonth != nil {
		series.OfTheMonth = *req.OfTheMonth
	}
	if req.DateReleased != "" {
		t, err := time.Parse(time.RFC3339, req.DateReleased)
		if err == nil {
			series.DateReleased = &t
		}
	}

	if err := s.repo.UpdateSeries(ctx, series); err != nil {
		return nil, fmt.Errorf("update series: %w", err)
	}
	return series, nil
}

func (s *contentService) DeleteSeries(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteSeries(ctx, id)
}

// ── Meditations ───────────────────────────────────────────────────────────────

func (s *contentService) CreateMeditation(ctx context.Context, req CreateMeditationRequest) (*domainContent.Meditation, error) {
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	m := &domainContent.Meditation{
		Name:   req.Name,
		Image:  req.Image,
		Status: status,
	}

	if err := s.repo.CreateMeditation(ctx, m); err != nil {
		return nil, fmt.Errorf("create meditation: %w", err)
	}
	return m, nil
}

func (s *contentService) ListMeditations(ctx context.Context, admin bool) ([]domainContent.Meditation, error) {
	return s.repo.GetMeditations(ctx, admin)
}

func (s *contentService) GetMeditation(ctx context.Context, id uuid.UUID, admin bool) (*domainContent.Meditation, error) {
	return s.repo.GetMeditationByID(ctx, id, admin)
}

func (s *contentService) UpdateMeditation(ctx context.Context, id uuid.UUID, req UpdateMeditationRequest) (*domainContent.Meditation, error) {
	m, err := s.repo.GetMeditationByID(ctx, id, true)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Image != "" {
		m.Image = req.Image
	}
	if req.Status != "" {
		m.Status = req.Status
	}

	if err := s.repo.UpdateMeditation(ctx, m); err != nil {
		return nil, fmt.Errorf("update meditation: %w", err)
	}
	return m, nil
}

func (s *contentService) DeleteMeditation(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteMeditation(ctx, id)
}

// ── Homepage ──────────────────────────────────────────────────────────────────

func (s *contentService) GetHomepage(ctx context.Context) (*domainContent.Homepage, error) {
	return s.repo.GetHomepage(ctx)
}
