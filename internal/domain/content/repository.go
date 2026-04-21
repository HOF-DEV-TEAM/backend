package content

import (
	"context"

	"github.com/google/uuid"
)

// MessageFilter carries optional query parameters for listing audio messages.
type MessageFilter struct {
	Search   string
	SeriesID *uuid.UUID
	IsFree   *bool
	Page     int
	PageSize int
}

// Repository defines every persistence operation the domain needs on content.
type Repository interface {
	// Audio messages
	CreateMessage(ctx context.Context, m *AudioMessage) error
	GetMessages(ctx context.Context, filter MessageFilter) ([]AudioMessage, int64, error)
	GetMessageByID(ctx context.Context, id uuid.UUID) (*AudioMessage, error)
	UpdateMessage(ctx context.Context, m *AudioMessage) error
	SoftDeleteMessage(ctx context.Context, id uuid.UUID) error

	// Audio series
	CreateSeries(ctx context.Context, s *AudioSeries) error
	GetAllSeries(ctx context.Context) ([]AudioSeries, int64, error)
	GetSeriesByID(ctx context.Context, id uuid.UUID) (*AudioSeries, error)
	UpdateSeries(ctx context.Context, s *AudioSeries) error
	SoftDeleteSeries(ctx context.Context, id uuid.UUID) error

	// Meditations
	CreateMeditation(ctx context.Context, m *Meditation) error
	GetMeditations(ctx context.Context, includeDeleted bool) ([]Meditation, error)
	GetMeditationByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*Meditation, error)
	UpdateMeditation(ctx context.Context, m *Meditation) error
	SoftDeleteMeditation(ctx context.Context, id uuid.UUID) error

	// Homepage
	GetHomepage(ctx context.Context) (*Homepage, error)
}
