package service

import (
	"math"
	"sort"
	"strings"
	"toll_plaza/internal/domain"
	"toll_plaza/pkg/geometry"
)

type SpatialService interface {
	FilterAndOrderTollPlazas(candidateTolls []domain.TollPlaza, routePoints []domain.Point, bufferMeters, dedupDistanceMeters float64) []domain.TollPlaza
}

type spatialService struct{}

func NewSpatialService() SpatialService {
	return &spatialService{}
}

type matchedToll struct {
	toll         domain.TollPlaza
	distFromSrc  float64
	perpDistance float64
}

func (s *spatialService) FilterAndOrderTollPlazas(
	candidateTolls []domain.TollPlaza,
	routePoints []domain.Point,
	bufferMeters,
	dedupDistanceMeters float64,
) []domain.TollPlaza {
	if len(candidateTolls) == 0 || len(routePoints) < 2 {
		return []domain.TollPlaza{}
	}

	var matched []matchedToll

	for _, toll := range candidateTolls {
		proj := geometry.ProjectPointOntoPolyline(toll.Latitude, toll.Longitude, routePoints, bufferMeters)
		if proj.IsOnRoute {
			toll.DistanceFromSource = proj.DistanceFromSourceKm
			matched = append(matched, matchedToll{
				toll:         toll,
				distFromSrc:  proj.DistanceFromSourceKm,
				perpDistance: proj.PerpendicularDistanceMeters,
			})
		}
	}

	if len(matched) == 0 {
		return []domain.TollPlaza{}
	}

	// Sort by distance from source ascending
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].distFromSrc < matched[j].distFromSrc
	})

	// Deduplicate twin plazas (e.g., opposite carriageway or dual entries within dedup distance)
	var filtered []domain.TollPlaza
	dedupDistKm := dedupDistanceMeters / 1000.0

	for i := 0; i < len(matched); i++ {
		current := matched[i]

		isDuplicate := false
		for _, existing := range filtered {
			// Check distance along route
			distDiffKm := math.Abs(current.distFromSrc - existing.DistanceFromSource)
			// Check physical distance between points
			physicalDistKm := geometry.HaversineDistanceKm(
				current.toll.Latitude, current.toll.Longitude,
				existing.Latitude, existing.Longitude,
			)

			// If tolls are within threshold km along route OR physically close with similar names
			if distDiffKm < dedupDistKm || physicalDistKm < dedupDistKm {
				// Also check name similarity or close proximity
				cleanCurrName := strings.ToLower(strings.TrimSpace(current.toll.Name))
				cleanExistName := strings.ToLower(strings.TrimSpace(existing.Name))

				if strings.Contains(cleanCurrName, cleanExistName) ||
					strings.Contains(cleanExistName, cleanCurrName) ||
					distDiffKm < 1.0 || physicalDistKm < 1.0 {
					isDuplicate = true
					break
				}
			}
		}

		if !isDuplicate {
			filtered = append(filtered, current.toll)
		}
	}

	return filtered
}
