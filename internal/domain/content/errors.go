package content

import "errors"

var (
	ErrMessageNotFound    = errors.New("audio message not found")
	ErrSeriesNotFound     = errors.New("audio series not found")
	ErrMeditationNotFound = errors.New("meditation not found")
)
