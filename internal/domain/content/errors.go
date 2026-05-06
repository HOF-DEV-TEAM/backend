// Package content defines the content domain errors.
package content

import "bitbucket.org/hofng/hofApp/internal/domain/shared"

var (
	// ErrMessageNotFound is returned when an audio message cannot be found.
	ErrMessageNotFound = shared.ErrNotFound{Resource: "audio message", ID: ""}
	// ErrSeriesNotFound is returned when an audio series cannot be found.
	ErrSeriesNotFound = shared.ErrNotFound{Resource: "audio series", ID: ""}
	// ErrMeditationNotFound is returned when a meditation cannot be found.
	ErrMeditationNotFound = shared.ErrNotFound{Resource: "meditation", ID: ""}
)
