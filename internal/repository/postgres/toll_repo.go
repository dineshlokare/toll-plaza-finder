package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
	"toll_plaza/internal/domain"
)

type TollRepository interface {
	UpsertBatch(ctx context.Context, tolls []domain.TollPlaza) error
	GetTollPlazasInBoundingBox(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]domain.TollPlaza, error)
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context) ([]domain.TollPlaza, error)
}

type tollRepository struct {
	db *DB
}

func NewTollRepository(db *DB) TollRepository {
	return &tollRepository{db: db}
}

func (r *tollRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM toll_plazas").Scan(&count)
	return count, err
}

func (r *tollRepository) GetAll(ctx context.Context) ([]domain.TollPlaza, error) {
	query := `
		SELECT id, name, latitude, longitude, COALESCE(geo_state, ''), created_at, updated_at
		FROM toll_plazas
		ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query toll plazas: %w", err)
	}
	defer rows.Close()

	var tolls []domain.TollPlaza
	for rows.Next() {
		var t domain.TollPlaza
		if err := rows.Scan(&t.ID, &t.Name, &t.Latitude, &t.Longitude, &t.GeoState, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan toll plaza: %w", err)
		}
		tolls = append(tolls, t)
	}
	return tolls, rows.Err()
}

func (r *tollRepository) GetTollPlazasInBoundingBox(ctx context.Context, minLat, minLon, maxLat, maxLon float64) ([]domain.TollPlaza, error) {
	query := `
		SELECT id, name, latitude, longitude, COALESCE(geo_state, ''), created_at, updated_at
		FROM toll_plazas
		WHERE latitude BETWEEN $1 AND $2
		  AND longitude BETWEEN $3 AND $4
	`
	rows, err := r.db.QueryContext(ctx, query, minLat, maxLat, minLon, maxLon)
	if err != nil {
		return nil, fmt.Errorf("failed to query toll plazas in bounding box: %w", err)
	}
	defer rows.Close()

	var tolls []domain.TollPlaza
	for rows.Next() {
		var t domain.TollPlaza
		if err := rows.Scan(&t.ID, &t.Name, &t.Latitude, &t.Longitude, &t.GeoState, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan toll plaza: %w", err)
		}
		tolls = append(tolls, t)
	}
	return tolls, rows.Err()
}

func (r *tollRepository) UpsertBatch(ctx context.Context, tolls []domain.TollPlaza) error {
	if len(tolls) == 0 {
		return nil
	}

	batchSize := 200
	for i := 0; i < len(tolls); i += batchSize {
		end := i + batchSize
		if end > len(tolls) {
			end = len(tolls)
		}
		batch := tolls[i:end]

		if err := r.upsertChunk(ctx, batch); err != nil {
			return err
		}
	}

	log.Printf("Successfully upserted %d toll plazas into database", len(tolls))
	return nil
}

func (r *tollRepository) upsertChunk(ctx context.Context, batch []domain.TollPlaza) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	var valueStrings []string
	var valueArgs []interface{}
	now := time.Now()

	for idx, t := range batch {
		baseParam := idx*5 + 1
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
			baseParam, baseParam+1, baseParam+2, baseParam+3, baseParam+4, baseParam+4))
		valueArgs = append(valueArgs, t.Name, t.Latitude, t.Longitude, t.GeoState, now)
	}

	stmt := fmt.Sprintf(`
		INSERT INTO toll_plazas (name, latitude, longitude, geo_state, created_at, updated_at)
		VALUES %s
		ON CONFLICT (name, latitude, longitude)
		DO UPDATE SET
			geo_state = EXCLUDED.geo_state,
			updated_at = EXCLUDED.updated_at
	`, strings.Join(valueStrings, ","))

	if _, err := tx.ExecContext(ctx, stmt, valueArgs...); err != nil {
		return fmt.Errorf("failed to execute batch upsert: %w", err)
	}

	return tx.Commit()
}
