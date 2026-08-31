package geometry

import (
	"math"
)

const (
	// EarthRadiusMeters is the approximate mean radius of the Earth in meters
	EarthRadiusMeters = 6371000.0
	// EarthRadiusKm is the radius in kilometers
	EarthRadiusKm = 6371.0
)

// ToRadians converts degrees to radians
func ToRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// ToDegrees converts radians to degrees
func ToDegrees(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// HaversineDistanceMeters calculates the great-circle distance between two coordinates in meters
func HaversineDistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := ToRadians(lat2 - lat1)
	dLon := ToRadians(lon2 - lon1)

	rLat1 := ToRadians(lat1)
	rLat2 := ToRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusMeters * c
}

// HaversineDistanceKm calculates the great-circle distance in kilometers
func HaversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	return HaversineDistanceMeters(lat1, lon1, lat2, lon2) / 1000.0
}
