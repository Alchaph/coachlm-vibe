package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Gear struct {
	ID          int64
	ExternalID  string
	Name        string
	BrandName   string
	ModelName   string
	Description string
	Distance    float64
	IsPrimary   bool
	IsRetired   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (db *DB) SaveGear(gear *Gear) error {
	if gear == nil {
		return errors.New("gear is nil")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	isPrimary := 0
	if gear.IsPrimary {
		isPrimary = 1
	}
	isRetired := 0
	if gear.IsRetired {
		isRetired = 1
	}

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO gear
			(external_id, name, brand_name, model_name, description, distance,
			 is_primary, is_retired, created_at, updated_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?,
			 COALESCE((SELECT created_at FROM gear WHERE external_id = ?), CURRENT_TIMESTAMP),
			 CURRENT_TIMESTAMP)`,
		gear.ExternalID,
		gear.Name,
		gear.BrandName,
		gear.ModelName,
		gear.Description,
		gear.Distance,
		isPrimary,
		isRetired,
		gear.ExternalID,
	)
	if err != nil {
		return fmt.Errorf("save gear: %w", err)
	}
	return nil
}

func (db *DB) GetGearByExternalID(externalID string) (*Gear, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	g := &Gear{}
	var isPrimary, isRetired int
	err := db.conn.QueryRow(`
		SELECT id, external_id, name, brand_name, model_name, description,
		       distance, is_primary, is_retired, created_at, updated_at
		FROM gear
		WHERE external_id = ?`, externalID).Scan(
		&g.ID,
		&g.ExternalID,
		&g.Name,
		&g.BrandName,
		&g.ModelName,
		&g.Description,
		&g.Distance,
		&isPrimary,
		&isRetired,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get gear by external id: %w", err)
	}
	g.IsPrimary = isPrimary == 1
	g.IsRetired = isRetired == 1
	return g, nil
}

func (db *DB) ListGear() ([]*Gear, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(`
		SELECT id, external_id, name, brand_name, model_name, description,
		       distance, is_primary, is_retired, created_at, updated_at
		FROM gear
		WHERE is_retired = 0
		ORDER BY distance DESC`)
	if err != nil {
		return nil, fmt.Errorf("list gear: %w", err)
	}
	defer rows.Close()

	result := make([]*Gear, 0)
	for rows.Next() {
		g := &Gear{}
		var isPrimary, isRetired int
		if err := rows.Scan(
			&g.ID,
			&g.ExternalID,
			&g.Name,
			&g.BrandName,
			&g.ModelName,
			&g.Description,
			&g.Distance,
			&isPrimary,
			&isRetired,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan gear: %w", err)
		}
		g.IsPrimary = isPrimary == 1
		g.IsRetired = isRetired == 1
		result = append(result, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gear: %w", err)
	}
	return result, nil
}
