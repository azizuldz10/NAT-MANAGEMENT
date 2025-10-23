package database

import (
	"context"
	"fmt"

	"github.com/lib/pq"
)

// RouterAccessControl represents router access control from database
type RouterAccessControl struct {
	ID          int      `json:"id"`
	Role        string   `json:"role"`
	RouterName  string   `json:"router_name"`
	Permissions []string `json:"permissions"`
	Description string   `json:"description"`
}

// AccessControlRepository handles database operations for router access control
type AccessControlRepository struct {
	db *DB
}

// NewAccessControlRepository creates a new access control repository
func NewAccessControlRepository(db *DB) *AccessControlRepository {
	return &AccessControlRepository{db: db}
}

// GetByRole retrieves all router access rules for a specific role
func (r *AccessControlRepository) GetByRole(ctx context.Context, role string) ([]RouterAccessControl, error) {
	query := `
		SELECT id, role, router_name, permissions, description
		FROM router_access_control
		WHERE role = $1
		ORDER BY router_name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get access control for role %s: %w", role, err)
	}
	defer rows.Close()

	var accessRules []RouterAccessControl
	for rows.Next() {
		var rule RouterAccessControl
		err := rows.Scan(
			&rule.ID,
			&rule.Role,
			&rule.RouterName,
			pq.Array(&rule.Permissions),
			&rule.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access control: %w", err)
		}
		accessRules = append(accessRules, rule)
	}

	return accessRules, nil
}

// GetRouterNamesByRole retrieves all router names accessible by a role
func (r *AccessControlRepository) GetRouterNamesByRole(ctx context.Context, role string) ([]string, error) {
	query := `
		SELECT router_name
		FROM router_access_control
		WHERE role = $1
		ORDER BY router_name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get router names for role %s: %w", role, err)
	}
	defer rows.Close()

	var routerNames []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan router name: %w", err)
		}
		routerNames = append(routerNames, name)
	}

	return routerNames, nil
}

// HasAccess checks if a role has access to a specific router
func (r *AccessControlRepository) HasAccess(ctx context.Context, role, routerName string) (bool, error) {
	// Check for wildcard access first
	wildcardQuery := `
		SELECT EXISTS(
			SELECT 1 FROM router_access_control
			WHERE role = $1 AND router_name = '*'
		)
	`

	var hasWildcard bool
	err := r.db.Pool.QueryRow(ctx, wildcardQuery, role).Scan(&hasWildcard)
	if err != nil {
		return false, fmt.Errorf("failed to check wildcard access: %w", err)
	}

	if hasWildcard {
		return true, nil
	}

	// Check for specific router access
	specificQuery := `
		SELECT EXISTS(
			SELECT 1 FROM router_access_control
			WHERE role = $1 AND router_name = $2
		)
	`

	var hasAccess bool
	err = r.db.Pool.QueryRow(ctx, specificQuery, role, routerName).Scan(&hasAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check router access: %w", err)
	}

	return hasAccess, nil
}

// GetPermissions retrieves permissions for a role on a specific router
func (r *AccessControlRepository) GetPermissions(ctx context.Context, role, routerName string) ([]string, error) {
	query := `
		SELECT permissions
		FROM router_access_control
		WHERE role = $1 AND (router_name = $2 OR router_name = '*')
		LIMIT 1
	`

	var permissions []string
	err := r.db.Pool.QueryRow(ctx, query, role, routerName).Scan(pq.Array(&permissions))
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return permissions, nil
}

// Create inserts a new access control rule
func (r *AccessControlRepository) Create(ctx context.Context, rule *RouterAccessControl) error {
	query := `
		INSERT INTO router_access_control (role, router_name, permissions, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query,
		rule.Role,
		rule.RouterName,
		pq.Array(rule.Permissions),
		rule.Description,
	).Scan(&rule.ID)

	if err != nil {
		return fmt.Errorf("failed to create access control rule: %w", err)
	}

	r.db.Logger.Infof("✅ Access control rule created: %s -> %s", rule.Role, rule.RouterName)
	return nil
}

// Update updates an existing access control rule
func (r *AccessControlRepository) Update(ctx context.Context, rule *RouterAccessControl) error {
	query := `
		UPDATE router_access_control
		SET router_name = $2, permissions = $3, description = $4
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query,
		rule.ID,
		rule.RouterName,
		pq.Array(rule.Permissions),
		rule.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update access control rule: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("access control rule not found: %d", rule.ID)
	}

	r.db.Logger.Infof("✅ Access control rule updated: ID %d", rule.ID)
	return nil
}

// Delete removes an access control rule
func (r *AccessControlRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM router_access_control WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete access control rule: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("access control rule not found: %d", id)
	}

	r.db.Logger.Infof("🗑️ Access control rule deleted: ID %d", id)
	return nil
}

// GetAll retrieves all access control rules
func (r *AccessControlRepository) GetAll(ctx context.Context) ([]RouterAccessControl, error) {
	query := `
		SELECT id, role, router_name, permissions, description
		FROM router_access_control
		ORDER BY role ASC, router_name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all access control rules: %w", err)
	}
	defer rows.Close()

	var rules []RouterAccessControl
	for rows.Next() {
		var rule RouterAccessControl
		err := rows.Scan(
			&rule.ID,
			&rule.Role,
			&rule.RouterName,
			pq.Array(&rule.Permissions),
			&rule.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access control rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}
