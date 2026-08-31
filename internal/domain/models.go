package domain

import "time"

// Point represents a geographical coordinate with optional cumulative distance from start
type Point struct {
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	CumulativeDistKm   float64 `json:"cumulativeDistKm"`
}

// TollPlaza represents a toll plaza record in the database
type TollPlaza struct {
	ID                 int64     `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Latitude           float64   `json:"latitude" db:"latitude"`
	Longitude          float64   `json:"longitude" db:"longitude"`
	GeoState           string    `json:"geoState" db:"geo_state"`
	DistanceFromSource float64   `json:"distanceFromSource" db:"-"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

// PincodeLocation represents cached or resolved pincode coordinates
type PincodeLocation struct {
	Pincode     string    `json:"pincode" db:"pincode"`
	Latitude    float64   `json:"latitude" db:"latitude"`
	Longitude   float64   `json:"longitude" db:"longitude"`
	DisplayName string    `json:"displayName" db:"display_name"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// RouteResult represents the output of a route calculation
type RouteResult struct {
	SourcePincode      string
	DestinationPincode string
	DistanceInKm       float64
	DurationInMinutes  float64
	PolylinePoints     []Point
}
