package domain

// TollResponse represents the exact JSON response required by the specification
type TollResponse struct {
	Route      RouteSummary        `json:"route"`
	TollPlazas []TollPlazaResponse `json:"tollPlazas"`
}

// RouteSummary holds the route summary details
type RouteSummary struct {
	SourcePincode      string `json:"sourcePincode"`
	DestinationPincode string `json:"destinationPincode"`
	DistanceInKm       int    `json:"distanceInKm"`
}

// TollPlazaResponse represents a single toll plaza along the route
type TollPlazaResponse struct {
	Name               string  `json:"name"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	DistanceFromSource int     `json:"distanceFromSource"`
}

// ErrorResponse represents standard error JSON output
type ErrorResponse struct {
	Error string `json:"error"`
}
