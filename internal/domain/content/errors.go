// Package content defines the content domain errors.
package content

import "errors"

var (
	// ErrMessageNotFound is returned when an audio message cannot be found.
	ErrMessageNotFound = errors.New("audio message not found")
	// ErrSeriesNotFound is returned when an audio series cannot be found.
	ErrSeriesNotFound = errors.New("audio series not found")
	// ErrMeditationNotFound is returned when a meditation cannot be found.
	ErrMeditationNotFound = errors.New("meditation not found")
)
