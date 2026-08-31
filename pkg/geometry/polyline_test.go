package geometry_test

import (
	"testing"
	"toll_plaza/internal/domain"
	"toll_plaza/pkg/geometry"

	"github.com/stretchr/testify/assert"
)

func TestComputeCumulativeDistances(t *testing.T) {
	points := []domain.Point{
		{Latitude: 12.9716, Longitude: 77.5946},
		{Latitude: 13.0000, Longitude: 77.6000},
		{Latitude: 13.0500, Longitude: 77.6200},
	}

	result := geometry.ComputeCumulativeDistances(points)
	assert.Len(t, result, 3)
	assert.Equal(t, 0.0, result[0].CumulativeDistKm)
	assert.Greater(t, result[1].CumulativeDistKm, 0.0)
	assert.Greater(t, result[2].CumulativeDistKm, result[1].CumulativeDistKm)
}

func TestProjectPointOntoPolyline(t *testing.T) {
	// A straight line going North: (12.0, 77.0) -> (13.0, 77.0)
	route := geometry.ComputeCumulativeDistances([]domain.Point{
		{Latitude: 12.0, Longitude: 77.0},
		{Latitude: 13.0, Longitude: 77.0},
	})

	// Point right on the route halfway: (12.5, 77.0)
	midPoint := geometry.ProjectPointOntoPolyline(12.5, 77.0, route, 500.0)
	assert.True(t, midPoint.IsOnRoute)
	assert.InDelta(t, 0.0, midPoint.PerpendicularDistanceMeters, 50.0)
	assert.InDelta(t, route[1].CumulativeDistKm/2.0, midPoint.DistanceFromSourceKm, 1.0)

	// Point 200 meters east of the route: (12.5, 77.0018)
	nearPoint := geometry.ProjectPointOntoPolyline(12.5, 77.0018, route, 500.0)
	assert.True(t, nearPoint.IsOnRoute)
	assert.Less(t, nearPoint.PerpendicularDistanceMeters, 500.0)

	// Point 10 km away from the route
	farPoint := geometry.ProjectPointOntoPolyline(12.5, 77.1, route, 500.0)
	assert.False(t, farPoint.IsOnRoute)
	assert.Greater(t, farPoint.PerpendicularDistanceMeters, 5000.0)
}

func TestBoundingBox(t *testing.T) {
	points := []domain.Point{
		{Latitude: 12.0, Longitude: 77.0},
		{Latitude: 13.0, Longitude: 78.0},
	}

	minLat, minLon, maxLat, maxLon := geometry.BoundingBox(points, 10.0)
	assert.Less(t, minLat, 12.0)
	assert.Less(t, minLon, 77.0)
	assert.Greater(t, maxLat, 13.0)
	assert.Greater(t, maxLon, 78.0)
}
