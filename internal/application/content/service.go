// Package content provides the content application service and DTOs.
package content

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domainContent "bitbucket.org/hofng/hofApp/internal/domain/content"
	"bitbucket.org/hofng/hofApp/internal/domain/shared"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var validate = validator.New()

// isAccessAllowed checks if a viewer role can access content with a given required access level.
// Access hierarchy (from broadest to narrowest):
//
//	"members" = all roles (members, stewards, leaders)
//	"stewards" = stewards and leaders only
//	"leaders" = leaders only
func isAccessAllowed(viewerRole, requiredAccess string) bool {
	viewerRole = strings.ToLower(strings.TrimSpace(viewerRole))
	requiredAccess = strings.ToLower(strings.TrimSpace(requiredAccess))

	if requiredAccess == "" {
		requiredAccess = "members"
	}

	// Rank: members=1, stewards=2, leaders=3
	rank := func(r string) int {
		switch r {
		case "leaders":
			return 3
		case "stewards":
			return 2
		case "members":
			return 1
		default:
			return 0
		}
	}

	return rank(viewerRole) >= rank(requiredAccess)
}

// Service exposes all content-management use cases.
type Service interface {
	// Messages
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*domainContent.AudioMessage, error)
	ListMessages(ctx context.Context, filter MessageListFilter) ([]domainContent.AudioMessage, int64, error)
	GetMessage(ctx context.Context, id uuid.UUID, viewerRole string) (*domainContent.AudioMessage, error)
	UpdateMessage(ctx context.Context, id uuid.UUID, req *UpdateMessageRequest) (*domainContent.AudioMessage, error)
	DeleteMessage(ctx context.Context, id uuid.UUID) error

	// Series
	CreateSeries(ctx context.Context, req *CreateSeriesRequest) (*domainContent.AudioSeries, error)
	ListSeries(ctx context.Context) ([]domainContent.AudioSeries, int64, error)
	GetSeries(ctx context.Context, id uuid.UUID) (*domainContent.AudioSeries, error)
	UpdateSeries(ctx context.Context, id uuid.UUID, req *UpdateSeriesRequest) (*domainContent.AudioSeries, error)
	DeleteSeries(ctx context.Context, id uuid.UUID) error

	// Meditations
	CreateMeditation(ctx context.Context, req *CreateMeditationRequest) (*domainContent.Meditation, error)
	ListMeditations(ctx context.Context, admin bool) ([]domainContent.Meditation, error)
	GetMeditation(ctx context.Context, id uuid.UUID, admin bool) (*domainContent.Meditation, error)
	UpdateMeditation(ctx context.Context, id uuid.UUID, req *UpdateMeditationRequest) (*domainContent.Meditation, error)
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

func (s *contentService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*domainContent.AudioMessage, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
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

	// Trim whitespace from audio and image URLs
	m.AudioURL = strings.TrimSpace(m.AudioURL)
	m.ImageURL = strings.TrimSpace(m.ImageURL)

	if req.SeriesID != "" {
		sid, err := uuid.Parse(req.SeriesID)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "series_id", Message: "invalid UUID"}
		}
		m.SeriesID = &sid
	}

	if req.DateReleased != "" {
		t, err := time.Parse("02/01/2006", req.DateReleased)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "date_released", Message: "must be DD/MM/YYYY format (e.g., 10/11/2022)"}
		}
		m.DateReleased = &t
	}

	// Determine and validate access level
	// Allowed values: leaders, stewards, members
	access := ""
	if strings.TrimSpace(req.Access) != "" {
		access = strings.ToLower(strings.TrimSpace(req.Access))
		switch access {
		case "leaders", "stewards", "members":
			// ok
		default:
			return nil, shared.ErrInvalidInput{Field: "access", Message: "must be one of: leaders, stewards, members"}
		}
	} else if req.AllowSteward {
		access = "stewards"
	} else {
		access = "members"
	}
	m.AccessLevel = access

	// Uniqueness check: ensure audio_url isn't already present
	if m.AudioURL != "" {
		if existing, err := s.repo.GetMessageByAudioURL(ctx, m.AudioURL); err == nil && existing != nil {
			return nil, shared.ErrAlreadyExists{Resource: "audio message", Field: "audio_url", Value: m.AudioURL}
		} else if err != nil {
			if _, ok := errors.AsType[shared.ErrNotFound](err); !ok {
				return nil, fmt.Errorf("checking audio_url uniqueness: %w", err)
			}
		}
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
	if f.Access != "" {
		// Map viewer role to allowed message access levels based on hierarchy:
		// viewer 'leaders' -> can see ["leaders", "stewards", "members"]
		// viewer 'stewards' -> can see ["stewards", "members"]
		// viewer 'members' -> can see ["members"]
		var accessIn []string
		switch strings.ToLower(f.Access) {
		case "leaders":
			accessIn = []string{"leaders", "stewards", "members"}
		case "stewards":
			accessIn = []string{"stewards", "members"}
		case "members":
			accessIn = []string{"members"}
		default:
			// unknown viewer role: no filter applied
			accessIn = nil
		}
		if len(accessIn) > 0 {
			filter.AccessIn = accessIn
		}
	}
	return s.repo.GetMessages(ctx, filter)
}

func (s *contentService) GetMessage(ctx context.Context, id uuid.UUID, viewerRole string) (*domainContent.AudioMessage, error) {
	m, err := s.repo.GetMessageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// Default to "members" if viewerRole is empty so access check always applies
	if viewerRole == "" {
		viewerRole = "members"
	}
	if !isAccessAllowed(viewerRole, m.AccessLevel) {
		return nil, shared.ErrForbidden{Message: "access denied"}
	}
	return m, nil
}

func (s *contentService) UpdateMessage(ctx context.Context, id uuid.UUID, req *UpdateMessageRequest) (*domainContent.AudioMessage, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
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
		newAudio := strings.TrimSpace(req.AudioURL)
		// If audio URL is changing, ensure uniqueness
		if newAudio != "" && newAudio != m.AudioURL {
			if existing, err := s.repo.GetMessageByAudioURL(ctx, newAudio); err == nil && existing != nil {
				// if existing message is a different record, conflict
				if existing.ID != m.ID {
					return nil, shared.ErrAlreadyExists{Resource: "audio message", Field: "audio_url", Value: newAudio}
				}
			} else if err != nil {
				if _, ok := errors.AsType[shared.ErrNotFound](err); !ok {
					return nil, fmt.Errorf("checking audio_url uniqueness: %w", err)
				}
			}
		}
		m.AudioURL = newAudio
	}
	if req.ImageURL != "" {
		m.ImageURL = strings.TrimSpace(req.ImageURL)
	}

	// Access update: can change via Access string (preferred) or legacy AllowSteward pointer
	if req.Access != nil {
		acc := strings.ToLower(strings.TrimSpace(*req.Access))
		switch acc {
		case "leaders", "stewards", "members":
			m.AccessLevel = acc
		default:
			return nil, shared.ErrInvalidInput{Field: "access", Message: "must be one of: leaders, stewards, members"}
		}
	} else if req.AllowSteward != nil {
		if *req.AllowSteward {
			m.AccessLevel = "stewards"
		} else {
			m.AccessLevel = "members"
		}
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
		t, err := time.Parse("02/01/2006", req.DateReleased)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "date_released", Message: "must be DD/MM/YYYY format (e.g., 10/11/2022)"}
		}
		m.DateReleased = &t
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

func (s *contentService) CreateSeries(ctx context.Context, req *CreateSeriesRequest) (*domainContent.AudioSeries, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
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

	// Trim whitespace from image URL
	series.ImageURL = strings.TrimSpace(series.ImageURL)

	if req.DateReleased != "" {
		t, err := time.Parse("02/01/2006", req.DateReleased)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "date_released", Message: "must be DD/MM/YYYY format (e.g., 10/11/2022)"}
		}
		series.DateReleased = &t
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

func (s *contentService) UpdateSeries(ctx context.Context, id uuid.UUID, req *UpdateSeriesRequest) (*domainContent.AudioSeries, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
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
		series.ImageURL = strings.TrimSpace(req.ImageURL)
	}
	if req.Description != "" {
		series.Description = req.Description
	}
	if req.OfTheMonth != nil {
		series.OfTheMonth = *req.OfTheMonth
	}
	if req.DateReleased != "" {
		t, err := time.Parse("02/01/2006", req.DateReleased)
		if err != nil {
			return nil, shared.ErrInvalidInput{Field: "date_released", Message: "must be DD/MM/YYYY format (e.g., 10/11/2022)"}
		}
		series.DateReleased = &t
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

func (s *contentService) CreateMeditation(ctx context.Context, req *CreateMeditationRequest) (*domainContent.Meditation, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
	if err := validate.Struct(req); err != nil {
		return nil, shared.ErrInvalidInput{Message: err.Error()}
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	m := &domainContent.Meditation{
		Name:   req.Name,
		Image:  strings.TrimSpace(req.Image),
		Text:   req.Text,
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

func (s *contentService) UpdateMeditation(ctx context.Context, id uuid.UUID, req *UpdateMeditationRequest) (*domainContent.Meditation, error) {
	if req == nil {
		return nil, shared.ErrInvalidInput{Message: "request body is required"}
	}
	m, err := s.repo.GetMeditationByID(ctx, id, true)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Image != "" {
		m.Image = strings.TrimSpace(req.Image)
	}
	if req.Text != "" {
		m.Text = req.Text
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
