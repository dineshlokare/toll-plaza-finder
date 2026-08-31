package domain

import "errors"

var (
	ErrInvalidPincode = errors.New("Invalid source or destination pincode")
	ErrSamePincode    = errors.New("Source and destination pincodes cannot be the same")
	ErrRouteNotFound  = errors.New("Could not find a valid driving route between the specified pincodes")
	ErrGeocodingFail  = errors.New("Unable to locate one or both pincodes")
)
