package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
	"toll_plaza/internal/cache"
	"toll_plaza/internal/config"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/repository/postgres"
	"toll_plaza/internal/validator"
	"toll_plaza/pkg/geometry"
)

type TollService interface {
	GetTollsBetweenPincodes(ctx context.Context, req domain.TollRequest) (*domain.TollResponse, error)
}

type tollService struct {
	cfg        *config.Config
	tollRepo   postgres.TollRepository
	geocoder   GeocodingService
	router     RoutingService
	spatial    SpatialService
	cache      cache.CacheService
}

func NewTollService(
	cfg *config.Config,
	tollRepo postgres.TollRepository,
	geocoder GeocodingService,
	router RoutingService,
	spatial SpatialService,
	cache cache.CacheService,
) TollService {
	return &tollService{
		cfg:        cfg,
		tollRepo:   tollRepo,
		geocoder:   geocoder,
		router:     router,
		spatial:    spatial,
		cache:      cache,
	}
}

func (s *tollService) GetTollsBetweenPincodes(ctx context.Context, req domain.TollRequest) (*domain.TollResponse, error) {
	// 1. Validate request
	if err := validator.ValidateTollRequest(&req); err != nil {
		return nil, err
	}

	srcPin := req.SourcePincode
	dstPin := req.DestinationPincode

	// 2. Check Cache
	if s.cache != nil {
		if cached, found := s.cache.GetRoute(ctx, srcPin, dstPin); found && cached != nil {
			log.Printf("Cache HIT for route %s -> %s", srcPin, dstPin)
			return cached, nil
		}
	}

	// 3. Geocode Source Pincode
	srcCoord, err := s.geocoder.GetCoordinates(ctx, srcPin)
	if err != nil {
		log.Printf("Geocoding failed for source pincode %s: %v", srcPin, err)
		return nil, domain.ErrInvalidPincode
	}

	// 4. Geocode Destination Pincode
	dstCoord, err := s.geocoder.GetCoordinates(ctx, dstPin)
	if err != nil {
		log.Printf("Geocoding failed for destination pincode %s: %v", dstPin, err)
		return nil, domain.ErrInvalidPincode
	}

	// 5. Calculate Driving Route
	routeResult, err := s.router.CalculateRoute(ctx, *srcCoord, *dstCoord)
	if err != nil {
		log.Printf("Routing failed between %s and %s: %v", srcPin, dstPin, err)
		return nil, domain.ErrRouteNotFound
	}

	// 6. Query Candidate Toll Plazas within Route Bounding Box (with 20km margin)
	minLat, minLon, maxLat, maxLon := geometry.BoundingBox(routeResult.PolylinePoints, 20.0)
	candidateTolls, err := s.tollRepo.GetTollPlazasInBoundingBox(ctx, minLat, minLon, maxLat, maxLon)
	if err != nil || len(candidateTolls) == 0 {
		log.Printf("Bounding box query returned 0 tolls, fallback to all tolls from DB: %v", err)
		candidateTolls, err = s.tollRepo.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch toll plazas: %w", err)
		}
	}

	// 7. Filter and Order Tolls along Highway Route Polyline
	matchedTolls := s.spatial.FilterAndOrderTollPlazas(
		candidateTolls,
		routeResult.PolylinePoints,
		s.cfg.TollBufferMeters,
		s.cfg.TollDeduplicationDist,
	)

	// 8. Build Response DTO
	tollPlazaResponses := make([]domain.TollPlazaResponse, 0, len(matchedTolls))
	for _, t := range matchedTolls {
		tollPlazaResponses = append(tollPlazaResponses, domain.TollPlazaResponse{
			Name:               t.Name,
			Latitude:           t.Latitude,
			Longitude:          t.Longitude,
			DistanceFromSource: int(math.Round(t.DistanceFromSource)),
		})
	}

	resp := &domain.TollResponse{
		Route: domain.RouteSummary{
			SourcePincode:      srcPin,
			DestinationPincode: dstPin,
			DistanceInKm:       int(math.Round(routeResult.DistanceInKm)),
		},
		TollPlazas: tollPlazaResponses,
	}

	// 9. Save to Cache asynchronously or with short timeout
	if s.cache != nil {
		ttl := time.Duration(s.cfg.RedisCacheTTLHours) * time.Hour
		s.cache.SetRoute(ctx, srcPin, dstPin, resp, ttl)
	}

	return resp, nil
}
