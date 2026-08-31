package geometry_test

import (
	"testing"
	"toll_plaza/pkg/geometry"

	"github.com/stretchr/testify/assert"
)

func TestHaversineDistance(t *testing.T) {
	// Delhi (28.6139, 77.2090) to Mumbai (19.0760, 72.8777) ~ 1148 km
	delhiLat, delhiLon := 28.6139, 77.2090
	mumbaiLat, mumbaiLon := 19.0760, 72.8777

	distKm := geometry.HaversineDistanceKm(delhiLat, delhiLon, mumbaiLat, mumbaiLon)
	assert.InDelta(t, 1148.0, distKm, 20.0, "Delhi to Mumbai distance should be ~1148 km")

	distMeters := geometry.HaversineDistanceMeters(delhiLat, delhiLon, mumbaiLat, mumbaiLon)
	assert.InDelta(t, 1148000.0, distMeters, 20000.0)

	// Zero distance for identical points
	zeroDist := geometry.HaversineDistanceKm(delhiLat, delhiLon, delhiLat, delhiLon)
	assert.Equal(t, 0.0, zeroDist)
}
