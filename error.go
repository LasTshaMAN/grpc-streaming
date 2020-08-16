// This file defines all domain errors for this service.

package streaming

import "errors"

var (
	// ErrDataNotFoundInStorage is returned by DataStorage when there is no requested data in DataStorage storage.
	ErrDataNotFoundInStorage = errors.New("data was not found in storage")
	// ErrDataCurrentlyUnavailable is returned by DataProvider when data is not available at this very moment,
	// and will not be available in the near future (for example, during the following 30 seconds - that depends on data source).
	ErrDataCurrentlyUnavailable = errors.New("data is currently unavailable")
)
