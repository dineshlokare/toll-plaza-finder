package geometry

import (
	"math"
	"toll_plaza/internal/domain"
)

// ComputeCumulativeDistances populates cumulative distances in kilometers for all points along a route
func ComputeCumulativeDistances(points []domain.Point) []domain.Point {
	if len(points) == 0 {
		return points
	}

	result := make([]domain.Point, len(points))
	result[0] = points[0]
	result[0].CumulativeDistKm = 0.0

	var totalDistKm float64
	for i := 1; i < len(points); i++ {
		distKm := HaversineDistanceKm(points[i-1].Latitude, points[i-1].Longitude, points[i].Latitude, points[i].Longitude)
		totalDistKm += distKm
		result[i] = points[i]
		result[i].CumulativeDistKm = totalDistKm
	}

	return result
}

// ProjectionResult stores information about the projected point on the polyline
type ProjectionResult struct {
	PerpendicularDistanceMeters float64
	DistanceFromSourceKm        float64
	SegmentIndex                int
	IsOnRoute                   bool
}

// ProjectPointOntoPolyline finds the minimum perpendicular distance from a target point (toll)
// to a polyline route, and calculates the cumulative distance from the route origin if within buffer.
func ProjectPointOntoPolyline(pointLat, pointLon float64, routePoints []domain.Point, bufferMeters float64) ProjectionResult {
	if len(routePoints) < 2 {
		return ProjectionResult{
			PerpendicularDistanceMeters: math.MaxFloat64,
			DistanceFromSourceKm:        0,
			SegmentIndex:                -1,
			IsOnRoute:                   false,
		}
	}

	minDistMeters := math.MaxFloat64
	bestDistFromSourceKm := 0.0
	bestSegmentIdx := -1

	for i := 0; i < len(routePoints)-1; i++ {
		p1 := routePoints[i]
		p2 := routePoints[i+1]

		segDistMeters := HaversineDistanceMeters(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude)
		if segDistMeters < 1e-6 {
			// Negligible segment length
			distMeters := HaversineDistanceMeters(pointLat, pointLon, p1.Latitude, p1.Longitude)
			if distMeters < minDistMeters {
				minDistMeters = distMeters
				bestDistFromSourceKm = p1.CumulativeDistKm
				bestSegmentIdx = i
			}
			continue
		}

		// Equirectangular projection for point-to-segment distance
		midLatRad := ToRadians((p1.Latitude + p2.Latitude) / 2.0)
		cosMidLat := math.Cos(midLatRad)

		dx := (p2.Longitude - p1.Longitude) * cosMidLat
		dy := p2.Latitude - p1.Latitude

		px := (pointLon - p1.Longitude) * cosMidLat
		py := pointLat - p1.Latitude

		lenSq := dx*dx + dy*dy
		var t float64
		if lenSq > 0 {
			t = (px*dx + py*dy) / lenSq
		}

		// Clamp t to segment [0, 1]
		if t < 0.0 {
			t = 0.0
		} else if t > 1.0 {
			t = 1.0
		}

		projLat := p1.Latitude + t*(p2.Latitude-p1.Latitude)
		projLon := p1.Longitude + t*(p2.Longitude-p1.Longitude)

		distMeters := HaversineDistanceMeters(pointLat, pointLon, projLat, projLon)

		if distMeters < minDistMeters {
			minDistMeters = distMeters
			bestSegmentIdx = i
			// Distance from source = cumulative distance to p1 + fraction of segment distance
			bestDistFromSourceKm = p1.CumulativeDistKm + (t * (segDistMeters / 1000.0))
		}
	}

	return ProjectionResult{
		PerpendicularDistanceMeters: minDistMeters,
		DistanceFromSourceKm:        bestDistFromSourceKm,
		SegmentIndex:                bestSegmentIdx,
		IsOnRoute:                   minDistMeters <= bufferMeters,
	}
}

// BoundingBox calculates the [minLat, minLon, maxLat, maxLon] envelope for a route with a buffer in degrees
func BoundingBox(points []domain.Point, bufferKm float64) (minLat, minLon, maxLat, maxLon float64) {
	if len(points) == 0 {
		return 0, 0, 0, 0
	}

	minLat = points[0].Latitude
	maxLat = points[0].Latitude
	minLon = points[0].Longitude
	maxLon = points[0].Longitude

	for _, pt := range points {
		if pt.Latitude < minLat {
			minLat = pt.Latitude
		}
		if pt.Latitude > maxLat {
			maxLat = pt.Latitude
		}
		if pt.Longitude < minLon {
			minLon = pt.Longitude
		}
		if pt.Longitude > maxLon {
			maxLon = pt.Longitude
		}
	}

	// Approximate 1 deg latitude ~ 111 km
	latBuffer := bufferKm / 111.0
	// Longitudinal distance varies with cosine of latitude
	avgLatRad := ToRadians((minLat + maxLat) / 2.0)
	cosLat := math.Cos(avgLatRad)
	if cosLat < 0.1 {
		cosLat = 0.1
	}
	lonBuffer := bufferKm / (111.0 * cosLat)

	return minLat - latBuffer, minLon - lonBuffer, maxLat + latBuffer, maxLon + lonBuffer
}
