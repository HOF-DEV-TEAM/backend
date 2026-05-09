// Package content defines the content domain entities and value objects.
package content

import (
	"time"

	"github.com/google/uuid"
)

// AudioSeries is a named collection that groups related audio messages.
type AudioSeries struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title        string     `gorm:"type:varchar(200);not null"`
	Author       string     `gorm:"type:varchar(200)"`
	ImageURL     string     `gorm:"column:image_url;type:varchar(500)"`
	Description  string     `gorm:"type:text"`
	OfTheMonth   bool       `gorm:"column:of_the_month;default:false"`
	DateReleased *time.Time `gorm:"column:date_released"`
	CreatedAt    time.Time  `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`

	Messages []AudioMessage `gorm:"foreignKey:SeriesID"`
}

// TableName returns the database table for audio series.
func (AudioSeries) TableName() string { return "audio_series" }

// AudioMessage is a single preach-able audio content item.
// AllowSteward controls whether steward-role users may access this message
// even when it is not free and the listener has no active subscription.
type AudioMessage struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title        string    `gorm:"type:varchar(200);not null"`
	Author       string    `gorm:"type:varchar(200)"`
	ImageURL     string    `gorm:"column:image_url;type:varchar(500)"`
	AudioURL     string    `gorm:"column:audio_url;type:varchar(500);not null"`
	Description  string    `gorm:"type:text"`
	IsFree       bool      `gorm:"column:is_free;default:false"`
	AllowSteward bool      `gorm:"column:allow_steward;default:false"`
	// AccessLevel controls who may access this message.
	// Valid values: "leaders", "stewards", "members" (members includes stewards and leaders).
	AccessLevel  string     `gorm:"column:access_level;type:varchar(50);default:'members'"`
	SeriesID     *uuid.UUID `gorm:"column:series_id;type:uuid"`
	DateReleased *time.Time `gorm:"column:date_released"`
	CreatedAt    time.Time  `gorm:"column:date_added;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:last_updated;autoUpdateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`

	Series *AudioSeries `gorm:"foreignKey:SeriesID"`
}

// TableName returns the database table for audio messages.
func (AudioMessage) TableName() string { return "audio_messages" }

// Meditation is a guided meditation content item.
type Meditation struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `gorm:"type:varchar(200);not null"`
	Image     string     `gorm:"type:varchar(500)"`
	Text      string     `gorm:"type:text"`
	Status    string     `gorm:"type:varchar(50);default:'active'"`
	CreatedAt time.Time  `gorm:"column:date_added;autoCreateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

// TableName returns the database table for meditations.
func (Meditation) TableName() string { return "meditations" }

// Homepage aggregates the content shown on the app home screen.
type Homepage struct {
	Series      []AudioSeries `json:"audio_series"`
	Meditations []Meditation  `json:"meditations"`
}
