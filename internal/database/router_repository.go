package database

import (
	"context"
	"fmt"
	"time"

	"nat-management-app/internal/models"

	"github.com/jackc/pgx/v5"
)

// RouterRepository handles database operations for routers
type RouterRepository struct {
	db *DB
}

// NewRouterRepository creates a new router repository
func NewRouterRepository(db *DB) *RouterRepository {
	return &RouterRepository{db: db}
}

// Create inserts a new router into the database
func (r *RouterRepository) Create(ctx context.Context, router *models.Router) error {
	query := `
		INSERT INTO routers (
			id, name, host, port, username, password,
			tunnel_endpoint, public_ont_url, enabled, description,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		router.ID,
		router.Name,
		router.Host,
		router.Port,
		router.Username,
		router.Password,
		router.TunnelEndpoint,
		router.PublicONTURL,
		router.Enabled,
		router.Description,
		router.CreatedAt,
		router.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	r.db.Logger.Infof("✅ Router created in database: %s (ID: %s)", router.Name, router.ID)
	return nil
}

// GetByID retrieves a router by ID
func (r *RouterRepository) GetByID(ctx context.Context, id string) (*models.Router, error) {
	query := `
		SELECT id, name, host, port, username, password,
		       tunnel_endpoint, public_ont_url, enabled, description,
		       created_at, updated_at
		FROM routers
		WHERE id = $1
	`

	router := &models.Router{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&router.ID,
		&router.Name,
		&router.Host,
		&router.Port,
		&router.Username,
		&router.Password,
		&router.TunnelEndpoint,
		&router.PublicONTURL,
		&router.Enabled,
		&router.Description,
		&router.CreatedAt,
		&router.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("router not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}

	return router, nil
}

// GetByName retrieves a router by name
func (r *RouterRepository) GetByName(ctx context.Context, name string) (*models.Router, error) {
	query := `
		SELECT id, name, host, port, username, password,
		       tunnel_endpoint, public_ont_url, enabled, description,
		       created_at, updated_at
		FROM routers
		WHERE name = $1
	`

	router := &models.Router{}
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&router.ID,
		&router.Name,
		&router.Host,
		&router.Port,
		&router.Username,
		&router.Password,
		&router.TunnelEndpoint,
		&router.PublicONTURL,
		&router.Enabled,
		&router.Description,
		&router.CreatedAt,
		&router.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("router not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get router: %w", err)
	}

	return router, nil
}

// GetAll retrieves all routers
func (r *RouterRepository) GetAll(ctx context.Context) ([]models.Router, error) {
	query := `
		SELECT id, name, host, port, username, password,
		       tunnel_endpoint, public_ont_url, enabled, description,
		       created_at, updated_at
		FROM routers
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get routers: %w", err)
	}
	defer rows.Close()

	var routers []models.Router
	for rows.Next() {
		var router models.Router
		err := rows.Scan(
			&router.ID,
			&router.Name,
			&router.Host,
			&router.Port,
			&router.Username,
			&router.Password,
			&router.TunnelEndpoint,
			&router.PublicONTURL,
			&router.Enabled,
			&router.Description,
			&router.CreatedAt,
			&router.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan router: %w", err)
		}
		routers = append(routers, router)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routers: %w", err)
	}

	return routers, nil
}

// GetEnabled retrieves all enabled routers
func (r *RouterRepository) GetEnabled(ctx context.Context) ([]models.Router, error) {
	query := `
		SELECT id, name, host, port, username, password,
		       tunnel_endpoint, public_ont_url, enabled, description,
		       created_at, updated_at
		FROM routers
		WHERE enabled = true
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled routers: %w", err)
	}
	defer rows.Close()

	var routers []models.Router
	for rows.Next() {
		var router models.Router
		err := rows.Scan(
			&router.ID,
			&router.Name,
			&router.Host,
			&router.Port,
			&router.Username,
			&router.Password,
			&router.TunnelEndpoint,
			&router.PublicONTURL,
			&router.Enabled,
			&router.Description,
			&router.CreatedAt,
			&router.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan router: %w", err)
		}
		routers = append(routers, router)
	}

	return routers, nil
}

// Update updates an existing router
func (r *RouterRepository) Update(ctx context.Context, router *models.Router) error {
	query := `
		UPDATE routers
		SET name = $2, host = $3, port = $4, username = $5, password = $6,
		    tunnel_endpoint = $7, public_ont_url = $8, enabled = $9,
		    description = $10, updated_at = $11
		WHERE id = $1
	`

	router.UpdatedAt = time.Now()

	result, err := r.db.Pool.Exec(ctx, query,
		router.ID,
		router.Name,
		router.Host,
		router.Port,
		router.Username,
		router.Password,
		router.TunnelEndpoint,
		router.PublicONTURL,
		router.Enabled,
		router.Description,
		router.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update router: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("router not found: %s", router.ID)
	}

	r.db.Logger.Infof("✅ Router updated in database: %s (ID: %s)", router.Name, router.ID)
	return nil
}

// Delete removes a router from the database
func (r *RouterRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM routers WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete router: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("router not found: %s", id)
	}

	r.db.Logger.Infof("🗑️ Router deleted from database: %s", id)
	return nil
}

// Exists checks if a router with the given name exists
func (r *RouterRepository) Exists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM routers WHERE name = $1)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check router existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of routers
func (r *RouterRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM routers`

	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count routers: %w", err)
	}

	return count, nil
}

// CountEnabled returns the number of enabled routers
func (r *RouterRepository) CountEnabled(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM routers WHERE enabled = true`

	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count enabled routers: %w", err)
	}

	return count, nil
}
