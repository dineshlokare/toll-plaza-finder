package service_test

import (
	"testing"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/service"
	"toll_plaza/pkg/geometry"

	"github.com/stretchr/testify/assert"
)

func TestFilterAndOrderTollPlazas(t *testing.T) {
	spatial := service.NewSpatialService()

	// Route from South to North: 12.0 -> 13.0 -> 14.0 latitude along 77.0 longitude
	route := geometry.ComputeCumulativeDistances([]domain.Point{
		{Latitude: 12.0, Longitude: 77.0},
		{Latitude: 13.0, Longitude: 77.0},
		{Latitude: 14.0, Longitude: 77.0},
	})

	candidates := []domain.TollPlaza{
		{
			ID:        1,
			Name:      "Toll Far Away",
			Latitude:  12.5,
			Longitude: 78.0, // ~100 km away
		},
		{
			ID:        2,
			Name:      "Second Toll Plaza",
			Latitude:  13.5,
			Longitude: 77.0, // On route, further along
		},
		{
			ID:        3,
			Name:      "First Toll Plaza",
			Latitude:  12.5,
			Longitude: 77.0, // On route, closer to start
		},
		{
			ID:        4,
			Name:      "First Toll Plaza Twin (Opposite Lane)",
			Latitude:  12.5001,
			Longitude: 77.0001, // ~15 meters away from Toll 3
		},
	}

	results := spatial.FilterAndOrderTollPlazas(candidates, route, 500.0, 1000.0)

	// Should match only First Toll Plaza and Second Toll Plaza (Far away skipped, Twin deduplicated)
	assert.Len(t, results, 2)

	// Order check: First Toll Plaza then Second Toll Plaza
	assert.Equal(t, "First Toll Plaza", results[0].Name)
	assert.Equal(t, "Second Toll Plaza", results[1].Name)
	assert.Less(t, results[0].DistanceFromSource, results[1].DistanceFromSource)
}
