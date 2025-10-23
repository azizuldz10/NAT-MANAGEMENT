package database

import (
	"context"
	"fmt"
	"time"

	"nat-management-app/internal/models"
)

// ONTWiFiRepository handles database operations for ONT WiFi information
type ONTWiFiRepository struct {
	db *DB
}

// NewONTWiFiRepository creates a new ONT WiFi repository
func NewONTWiFiRepository(db *DB) *ONTWiFiRepository {
	return &ONTWiFiRepository{db: db}
}

// SaveWiFiInfo saves or updates WiFi information in the database
func (r *ONTWiFiRepository) SaveWiFiInfo(ctx context.Context, info *models.ONTWiFiInfo) error {
	query := `
		INSERT INTO ont_wifi_info (
			pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		info.PPPoEUsername,
		info.Router,
		info.SSID,
		info.Password,
		info.Security,
		info.Encryption,
		info.Authentication,
		info.ONTURL,
		info.ONTModel,
		info.ExtractedAt,
		info.ExtractedBy,
		time.Now(),
	).Scan(&info.ID)

	if err != nil {
		return fmt.Errorf("failed to save WiFi info: %w", err)
	}

	r.db.Logger.Debugf("Saved WiFi info with ID: %d", info.ID)
	return nil
}

// GetWiFiInfo retrieves WiFi information by ID
func (r *ONTWiFiRepository) GetWiFiInfo(ctx context.Context, id int) (*models.ONTWiFiInfo, error) {
	query := `
		SELECT id, pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
		FROM ont_wifi_info
		WHERE id = $1
	`

	info := &models.ONTWiFiInfo{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&info.ID,
		&info.PPPoEUsername,
		&info.Router,
		&info.SSID,
		&info.Password,
		&info.Security,
		&info.Encryption,
		&info.Authentication,
		&info.ONTURL,
		&info.ONTModel,
		&info.ExtractedAt,
		&info.ExtractedBy,
		&info.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get WiFi info: %w", err)
	}

	return info, nil
}

// GetLatestWiFiInfoByPPPoE retrieves the most recent WiFi info for a PPPoE username
func (r *ONTWiFiRepository) GetLatestWiFiInfoByPPPoE(ctx context.Context, pppoeUsername string) (*models.ONTWiFiInfo, error) {
	query := `
		SELECT id, pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
		FROM ont_wifi_info
		WHERE pppoe_username = $1
		ORDER BY extracted_at DESC
		LIMIT 1
	`

	info := &models.ONTWiFiInfo{}
	err := r.db.Pool.QueryRow(ctx, query, pppoeUsername).Scan(
		&info.ID,
		&info.PPPoEUsername,
		&info.Router,
		&info.SSID,
		&info.Password,
		&info.Security,
		&info.Encryption,
		&info.Authentication,
		&info.ONTURL,
		&info.ONTModel,
		&info.ExtractedAt,
		&info.ExtractedBy,
		&info.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get latest WiFi info: %w", err)
	}

	return info, nil
}

// GetWiFiHistory retrieves WiFi extraction history with optional filters
func (r *ONTWiFiRepository) GetWiFiHistory(ctx context.Context, req models.ONTWiFiHistoryRequest) ([]models.ONTWiFiInfo, int, error) {
	// Build query with filters
	baseQuery := `
		FROM ont_wifi_info
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) " + baseQuery
	selectQuery := `
		SELECT id, pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
	` + baseQuery

	// Add filters
	args := []interface{}{}
	argCount := 1

	if req.PPPoEUsername != "" {
		baseQuery += fmt.Sprintf(" AND pppoe_username = $%d", argCount)
		args = append(args, req.PPPoEUsername)
		argCount++
	}

	if req.Router != "" {
		baseQuery += fmt.Sprintf(" AND router = $%d", argCount)
		args = append(args, req.Router)
		argCount++
	}

	// Update queries with filters
	countQuery = "SELECT COUNT(*) " + baseQuery
	selectQuery = `
		SELECT id, pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
	` + baseQuery

	// Get total count
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count WiFi history: %w", err)
	}

	// Add ordering and pagination
	selectQuery += " ORDER BY extracted_at DESC"

	if req.Limit > 0 {
		selectQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, req.Limit)
		argCount++
	} else {
		selectQuery += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, 50) // Default limit
		argCount++
	}

	if req.Offset > 0 {
		selectQuery += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, req.Offset)
	}

	// Execute query
	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query WiFi history: %w", err)
	}
	defer rows.Close()

	// Scan results
	var history []models.ONTWiFiInfo
	for rows.Next() {
		info := models.ONTWiFiInfo{}
		err := rows.Scan(
			&info.ID,
			&info.PPPoEUsername,
			&info.Router,
			&info.SSID,
			&info.Password,
			&info.Security,
			&info.Encryption,
			&info.Authentication,
			&info.ONTURL,
			&info.ONTModel,
			&info.ExtractedAt,
			&info.ExtractedBy,
			&info.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan WiFi info: %w", err)
		}
		history = append(history, info)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating WiFi history: %w", err)
	}

	return history, total, nil
}

// SearchWiFiBySSID searches for WiFi information by SSID (fuzzy match)
func (r *ONTWiFiRepository) SearchWiFiBySSID(ctx context.Context, ssid string, limit int) ([]models.ONTWiFiInfo, error) {
	query := `
		SELECT id, pppoe_username, router, ssid, password, security, encryption,
			authentication, ont_url, ont_model, extracted_at, extracted_by, created_at
		FROM ont_wifi_info
		WHERE ssid ILIKE $1
		ORDER BY extracted_at DESC
		LIMIT $2
	`

	if limit <= 0 {
		limit = 50
	}

	rows, err := r.db.Pool.Query(ctx, query, "%"+ssid+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search WiFi by SSID: %w", err)
	}
	defer rows.Close()

	var results []models.ONTWiFiInfo
	for rows.Next() {
		info := models.ONTWiFiInfo{}
		err := rows.Scan(
			&info.ID,
			&info.PPPoEUsername,
			&info.Router,
			&info.SSID,
			&info.Password,
			&info.Security,
			&info.Encryption,
			&info.Authentication,
			&info.ONTURL,
			&info.ONTModel,
			&info.ExtractedAt,
			&info.ExtractedBy,
			&info.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WiFi info: %w", err)
		}
		results = append(results, info)
	}

	return results, nil
}

// DeleteOldWiFiInfo deletes WiFi information older than specified duration
func (r *ONTWiFiRepository) DeleteOldWiFiInfo(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM ont_wifi_info
		WHERE extracted_at < $1
	`

	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.db.Pool.Exec(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old WiFi info: %w", err)
	}

	rowsAffected := result.RowsAffected()
	r.db.Logger.Infof("Deleted %d old WiFi info records (older than %v)", rowsAffected, olderThan)
	return rowsAffected, nil
}

// GetWiFiInfoStats returns statistics about WiFi info records
func (r *ONTWiFiRepository) GetWiFiInfoStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_records,
			COUNT(DISTINCT pppoe_username) as unique_users,
			COUNT(DISTINCT router) as unique_routers,
			COUNT(DISTINCT ont_model) as unique_models,
			MAX(extracted_at) as latest_extraction,
			MIN(extracted_at) as oldest_extraction
		FROM ont_wifi_info
	`

	var stats struct {
		TotalRecords      int
		UniqueUsers       int
		UniqueRouters     int
		UniqueModels      int
		LatestExtraction  *time.Time
		OldestExtraction  *time.Time
	}

	err := r.db.Pool.QueryRow(ctx, query).Scan(
		&stats.TotalRecords,
		&stats.UniqueUsers,
		&stats.UniqueRouters,
		&stats.UniqueModels,
		&stats.LatestExtraction,
		&stats.OldestExtraction,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get WiFi info stats: %w", err)
	}

	result := map[string]interface{}{
		"total_records":      stats.TotalRecords,
		"unique_users":       stats.UniqueUsers,
		"unique_routers":     stats.UniqueRouters,
		"unique_models":      stats.UniqueModels,
		"latest_extraction":  stats.LatestExtraction,
		"oldest_extraction":  stats.OldestExtraction,
	}

	return result, nil
}
