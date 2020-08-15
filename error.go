// This file defines all domain errors for this service.

package streaming

import "errors"

var (
	ErrDataNotFoundInStorage = errors.New("data was not found in storage")
	ErrDataCurrentlyUnavailable = errors.New("data is currently unavailable")
)
