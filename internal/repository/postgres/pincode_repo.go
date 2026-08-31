package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"toll_plaza/internal/domain"
)

type PincodeRepository interface {
	Get(ctx context.Context, pincode string) (*domain.PincodeLocation, error)
	Save(ctx context.Context, loc *domain.PincodeLocation) error
}

type pincodeRepository struct {
	db *DB
}

func NewPincodeRepository(db *DB) PincodeRepository {
	return &pincodeRepository{db: db}
}

func (r *pincodeRepository) Get(ctx context.Context, pincode string) (*domain.PincodeLocation, error) {
	query := `
		SELECT pincode, latitude, longitude, display_name, updated_at
		FROM pincode_cache
		WHERE pincode = $1
	`
	var loc domain.PincodeLocation
	err := r.db.QueryRowContext(ctx, query, pincode).Scan(
		&loc.Pincode, &loc.Latitude, &loc.Longitude, &loc.DisplayName, &loc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not cached yet
		}
		return nil, fmt.Errorf("failed to query pincode cache: %w", err)
	}
	return &loc, nil
}

func (r *pincodeRepository) Save(ctx context.Context, loc *domain.PincodeLocation) error {
	query := `
		INSERT INTO pincode_cache (pincode, latitude, longitude, display_name, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pincode)
		DO UPDATE SET
			latitude = EXCLUDED.latitude,
			longitude = EXCLUDED.longitude,
			display_name = EXCLUDED.display_name,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, loc.Pincode, loc.Latitude, loc.Longitude, loc.DisplayName, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save pincode to cache: %w", err)
	}
	return nil
}
