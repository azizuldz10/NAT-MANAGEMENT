package database

import (
	"context"
	"fmt"
	"time"

	"nat-management-app/internal/models"

	"github.com/jackc/pgx/v5"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			username, password, full_name, email, role, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query,
		user.Username,
		user.Password,
		user.FullName,
		user.Email,
		user.Role,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.db.Logger.Infof("✅ User created in database: %s (ID: %d)", user.Username, user.ID)
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, username, password, full_name, email, role, is_active,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, password, full_name, email, role, is_active,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.FullName,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetAll retrieves all users
func (r *UserRepository) GetAll(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, username, password, full_name, email, role, is_active,
		       created_at, updated_at, last_login_at
		FROM users
		ORDER BY username ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.FullName,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $2, password = $3, full_name = $4, email = $5,
		    role = $6, is_active = $7, updated_at = $8, last_login_at = $9
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Password,
		user.FullName,
		user.Email,
		user.Role,
		user.IsActive,
		user.UpdatedAt,
		user.LastLoginAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	r.db.Logger.Infof("✅ User updated in database: %s (ID: %d)", user.Username, user.ID)
	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2`

	now := time.Now()
	result, err := r.db.Pool.Exec(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", userID)
	}

	return nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", id)
	}

	r.db.Logger.Infof("🗑️ User deleted from database: %d", id)
	return nil
}

// Exists checks if a user with the given username exists
func (r *UserRepository) Exists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// GetActiveUsers retrieves all active users
func (r *UserRepository) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT id, username, password, full_name, email, role, is_active,
		       created_at, updated_at, last_login_at
		FROM users
		WHERE is_active = true
		ORDER BY username ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Password,
			&user.FullName,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
