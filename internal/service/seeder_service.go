package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"toll_plaza/internal/domain"
	"toll_plaza/internal/repository/postgres"
)

type SeederService interface {
	SeedFromCSV(ctx context.Context, filePath string) error
	AutoSeedIfEmpty(ctx context.Context, filePath string) error
}

type seederService struct {
	tollRepo postgres.TollRepository
}

func NewSeederService(tollRepo postgres.TollRepository) SeederService {
	return &seederService{tollRepo: tollRepo}
}

func (s *seederService) AutoSeedIfEmpty(ctx context.Context, filePath string) error {
	count, err := s.tollRepo.Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count existing toll records: %w", err)
	}

	if count > 0 {
		log.Printf("Database already contains %d toll plazas. Skipping auto-seeding.", count)
		return nil
	}

	log.Printf("Toll plazas table is empty. Starting auto-seeding from %s...", filePath)
	return s.SeedFromCSV(ctx, filePath)
}

func (s *seederService) SeedFromCSV(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file %s: %w", filePath, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable fields if trailing commas exist
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Identify column indices
	lonIdx, latIdx, nameIdx, stateIdx := -1, -1, -1, -1
	for idx, col := range header {
		cleanCol := strings.ToLower(strings.TrimSpace(col))
		switch cleanCol {
		case "longitude", "lon", "lng":
			lonIdx = idx
		case "latitude", "lat":
			latIdx = idx
		case "toll_name", "name", "toll_plaza_name", "plaza_name":
			nameIdx = idx
		case "geo_state", "state":
			stateIdx = idx
		}
	}

	// Default to standard 0,1,2,3 if not matched
	if lonIdx == -1 {
		lonIdx = 0
	}
	if latIdx == -1 {
		latIdx = 1
	}
	if nameIdx == -1 {
		nameIdx = 2
	}
	if stateIdx == -1 {
		stateIdx = 3
	}

	var tolls []domain.TollPlaza
	seen := make(map[string]bool)
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Skipping invalid CSV line %d: %v", lineNum, err)
			lineNum++
			continue
		}
		lineNum++

		if len(record) <= latIdx || len(record) <= lonIdx || len(record) <= nameIdx {
			continue
		}

		lonStr := strings.TrimSpace(record[lonIdx])
		latStr := strings.TrimSpace(record[latIdx])
		name := strings.TrimSpace(record[nameIdx])
		state := ""
		if stateIdx >= 0 && len(record) > stateIdx {
			state = strings.TrimSpace(record[stateIdx])
		}

		if lonStr == "" || latStr == "" || name == "" {
			continue
		}

		lon, errLon := strconv.ParseFloat(lonStr, 64)
		lat, errLat := strconv.ParseFloat(latStr, 64)
		if errLon != nil || errLat != nil {
			continue
		}

		// Basic bounds check for coordinates in India / world
		if lat < -90.0 || lat > 90.0 || lon < -180.0 || lon > 180.0 {
			continue
		}

		// Deduplicate exact duplicate records in CSV before batching
		key := fmt.Sprintf("%s|%.6f|%.6f", strings.ToLower(name), lat, lon)
		if seen[key] {
			continue
		}
		seen[key] = true

		tolls = append(tolls, domain.TollPlaza{
			Name:      name,
			Latitude:  lat,
			Longitude: lon,
			GeoState:  state,
		})
	}

	log.Printf("Parsed %d unique toll plazas from CSV. Beginning database batch upsert...", len(tolls))
	return s.tollRepo.UpsertBatch(ctx, tolls)
}
