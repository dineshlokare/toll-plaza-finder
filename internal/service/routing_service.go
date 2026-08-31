package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"toll_plaza/internal/domain"
	"toll_plaza/pkg/geometry"
)

type RoutingService interface {
	CalculateRoute(ctx context.Context, src, dst domain.Point) (*domain.RouteResult, error)
}

type routingService struct {
	baseURL    string
	httpClient *http.Client
}

type osrmResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64      `json:"distance"` // in meters
	Duration float64      `json:"duration"` // in seconds
	Geometry osrmGeometry `json:"geometry"`
}

type osrmGeometry struct {
	Coordinates [][]float64 `json:"coordinates"` // [[lon, lat], ...]
	Type        string      `json:"type"`
}

func NewRoutingService(baseURL string) RoutingService {
	return &routingService{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *routingService) CalculateRoute(ctx context.Context, src, dst domain.Point) (*domain.RouteResult, error) {
	// Format: {baseURL}/route/v1/driving/{lon1},{lat1};{lon2},{lat2}?overview=full&geometries=geojson&steps=true
	endpoint := fmt.Sprintf("%s/route/v1/driving/%.6f,%.6f;%.6f,%.6f?overview=full&geometries=geojson&steps=true",
		s.baseURL, src.Longitude, src.Latitude, dst.Longitude, dst.Latitude)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build route request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call routing service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrRouteNotFound
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read route response: %w", err)
	}

	var osrmResp osrmResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, fmt.Errorf("failed to parse route response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, domain.ErrRouteNotFound
	}

	route := osrmResp.Routes[0]
	coords := route.Geometry.Coordinates

	if len(coords) == 0 {
		return nil, domain.ErrRouteNotFound
	}

	rawPoints := make([]domain.Point, len(coords))
	for i, coord := range coords {
		if len(coord) >= 2 {
			rawPoints[i] = domain.Point{
				Longitude: coord[0],
				Latitude:  coord[1],
			}
		}
	}

	// Compute cumulative distances along the polyline route
	polylineWithDistances := geometry.ComputeCumulativeDistances(rawPoints)

	return &domain.RouteResult{
		DistanceInKm:      route.Distance / 1000.0,
		DurationInMinutes: route.Duration / 60.0,
		PolylinePoints:    polylineWithDistances,
	}, nil
}
