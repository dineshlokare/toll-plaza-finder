package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/repository/postgres"
)

type GeocodingService interface {
	GetCoordinates(ctx context.Context, pincode string) (*domain.Point, error)
}

type geocodingService struct {
	pincodeRepo postgres.PincodeRepository
	baseURL     string
	httpClient  *http.Client
}

type nominatimResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func NewGeocodingService(pincodeRepo postgres.PincodeRepository, baseURL string) GeocodingService {
	return &geocodingService{
		pincodeRepo: pincodeRepo,
		baseURL:     strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *geocodingService) GetCoordinates(ctx context.Context, pincode string) (*domain.Point, error) {
	cleanPin := strings.TrimSpace(pincode)

	// 1. Check local DB cache
	if s.pincodeRepo != nil {
		cached, err := s.pincodeRepo.Get(ctx, cleanPin)
		if err == nil && cached != nil {
			return &domain.Point{
				Latitude:  cached.Latitude,
				Longitude: cached.Longitude,
			}, nil
		}
	}

	// 2. Query Nominatim API with postalcode filter
	endpoint := fmt.Sprintf("%s/search?postalcode=%s&country=India&format=json&limit=1",
		s.baseURL, url.QueryEscape(cleanPin))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create geocoding request: %w", err)
	}
	req.Header.Set("User-Agent", "TollPlazaFinderApp/1.0 (contact@tollplaza.local)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Geocoding HTTP error for pin %s: %v", cleanPin, err)
		return s.queryFallback(ctx, cleanPin)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var results []nominatimResult
			if err := json.Unmarshal(body, &results); err == nil && len(results) > 0 {
				lat, errLat := strconv.ParseFloat(results[0].Lat, 64)
				lon, errLon := strconv.ParseFloat(results[0].Lon, 64)
				if errLat == nil && errLon == nil {
					// Save to DB cache
					if s.pincodeRepo != nil {
						_ = s.pincodeRepo.Save(ctx, &domain.PincodeLocation{
							Pincode:     cleanPin,
							Latitude:    lat,
							Longitude:   lon,
							DisplayName: results[0].DisplayName,
						})
					}
					return &domain.Point{
						Latitude:  lat,
						Longitude: lon,
					}, nil
				}
			}
		}
	}

	// 3. Fallback query with general query string
	return s.queryFallback(ctx, cleanPin)
}

func (s *geocodingService) queryFallback(ctx context.Context, cleanPin string) (*domain.Point, error) {
	endpoint := fmt.Sprintf("%s/search?q=%s,+India&format=json&limit=1",
		s.baseURL, url.QueryEscape(cleanPin))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, domain.ErrInvalidPincode
	}
	req.Header.Set("User-Agent", "TollPlazaFinderApp/1.0 (contact@tollplaza.local)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, domain.ErrInvalidPincode
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var results []nominatimResult
		if err := json.Unmarshal(body, &results); err == nil && len(results) > 0 {
			lat, errLat := strconv.ParseFloat(results[0].Lat, 64)
			lon, errLon := strconv.ParseFloat(results[0].Lon, 64)
			if errLat == nil && errLon == nil {
				if s.pincodeRepo != nil {
					_ = s.pincodeRepo.Save(ctx, &domain.PincodeLocation{
						Pincode:     cleanPin,
						Latitude:    lat,
						Longitude:   lon,
						DisplayName: results[0].DisplayName,
					})
				}
				return &domain.Point{
					Latitude:  lat,
					Longitude: lon,
				}, nil
			}
		}
	}

	return nil, domain.ErrInvalidPincode
}
