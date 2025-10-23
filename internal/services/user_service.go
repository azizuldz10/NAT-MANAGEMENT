package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nat-management-app/internal/database"
	"nat-management-app/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user management operations
type UserService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewUserService creates a new UserService instance
func NewUserService(db *database.DB, logger *logrus.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// UserWithRouters represents a user with their assigned routers
type UserWithRouters struct {
	models.User
	Routers []string `json:"routers"`
}

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Routers  []string `json:"routers"` // List of router names
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	FullName string   `json:"full_name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password,omitempty"` // Optional: only if changing password
	Routers  []string `json:"routers"`            // List of router names
	IsActive bool     `json:"is_active"`
}

// CreateUser creates a new user with router assignments
func (s *UserService) CreateUser(req *CreateUserRequest) (*UserWithRouters, error) {
	// Check if username already exists
	var exists bool
	err := s.db.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		s.logger.Errorf("Error checking username existence: %v", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	err = s.db.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		s.logger.Errorf("Error checking email existence: %v", err)
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Error hashing password: %v", err)
		return nil, err
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(context.Background())
	if err != nil {
		s.logger.Errorf("Error starting transaction: %v", err)
		return nil, err
	}
	ctx := context.Background()
	defer tx.Rollback(ctx)

	// Insert user (role is deprecated, use empty string or default)
	var userID int
	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, password, full_name, email, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, req.Username, string(hashedPassword), req.FullName, req.Email, "User", true, time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		s.logger.Errorf("Error inserting user: %v", err)
		return nil, err
	}

	// Insert router assignments
	for _, routerName := range req.Routers {
		_, err = tx.Exec(ctx, `
			INSERT INTO user_routers (user_id, router_name, created_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, router_name) DO NOTHING
		`, userID, routerName, time.Now())
		if err != nil {
			s.logger.Errorf("Error assigning router %s to user: %v", routerName, err)
			return nil, err
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		s.logger.Errorf("Error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Infof("✅ User '%s' created successfully with %d router assignments", req.Username, len(req.Routers))

	// Return created user with routers
	return s.GetUserByID(userID)
}

// GetUserByID retrieves a user by ID with their router assignments
func (s *UserService) GetUserByID(userID int) (*UserWithRouters, error) {
	var user models.User

	// Get user
	err := s.db.Pool.QueryRow(context.Background(), `
		SELECT id, username, full_name, email, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Role,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)

	if err == pgx.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		s.logger.Errorf("Error getting user: %v", err)
		return nil, err
	}

	// Get user routers
	routers, err := s.GetUserRouters(userID)
	if err != nil {
		s.logger.Errorf("Error getting user routers: %v", err)
		return nil, err
	}

	return &UserWithRouters{
		User:    user,
		Routers: routers,
	}, nil
}

// GetUserByUsername retrieves a user by username with their router assignments
func (s *UserService) GetUserByUsername(username string) (*UserWithRouters, error) {
	var user models.User

	// Get user
	err := s.db.Pool.QueryRow(context.Background(), `
		SELECT id, username, full_name, email, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Role,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)

	if err == pgx.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		s.logger.Errorf("Error getting user: %v", err)
		return nil, err
	}

	// Get user routers
	routers, err := s.GetUserRouters(user.ID)
	if err != nil {
		s.logger.Errorf("Error getting user routers: %v", err)
		return nil, err
	}

	return &UserWithRouters{
		User:    user,
		Routers: routers,
	}, nil
}

// ListUsers retrieves all active users with pagination
func (s *UserService) ListUsers(limit, offset int) ([]UserWithRouters, int, error) {
	// Get total count of active users only
	var total int
	err := s.db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&total)
	if err != nil {
		s.logger.Errorf("Error counting users: %v", err)
		return nil, 0, err
	}

	// Get active users only
	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT id, username, full_name, email, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		s.logger.Errorf("Error listing users: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var users []UserWithRouters
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Role,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
		if err != nil {
			s.logger.Errorf("Error scanning user: %v", err)
			continue
		}

		// Get routers for this user
		routers, err := s.GetUserRouters(user.ID)
		if err != nil {
			s.logger.Warnf("Error getting routers for user %d: %v", user.ID, err)
			routers = []string{}
		}

		users = append(users, UserWithRouters{
			User:    user,
			Routers: routers,
		})
	}

	return users, total, nil
}

// UpdateUser updates a user's information and router assignments
func (s *UserService) UpdateUser(userID int, req *UpdateUserRequest) (*UserWithRouters, error) {
	// Check if user exists
	var exists bool
	err := s.db.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		s.logger.Errorf("Error checking user existence: %v", err)
		return nil, err
	}
	if !exists {
		return nil, errors.New("user not found")
	}

	// Start transaction
	tx, err := s.db.Pool.Begin(context.Background())
	if err != nil {
		s.logger.Errorf("Error starting transaction: %v", err)
		return nil, err
	}
	ctx := context.Background()
	defer tx.Rollback(ctx)

	// Update user info
	if req.Password != "" {
		// Update with new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Errorf("Error hashing password: %v", err)
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			UPDATE users
			SET full_name = $1, email = $2, password = $3, is_active = $4, updated_at = $5
			WHERE id = $6
		`, req.FullName, req.Email, string(hashedPassword), req.IsActive, time.Now(), userID)
		if err != nil {
			s.logger.Errorf("Error updating user: %v", err)
			return nil, err
		}
	} else {
		// Update without changing password
		_, err = tx.Exec(ctx, `
			UPDATE users
			SET full_name = $1, email = $2, is_active = $3, updated_at = $4
			WHERE id = $5
		`, req.FullName, req.Email, req.IsActive, time.Now(), userID)
		if err != nil {
			s.logger.Errorf("Error updating user: %v", err)
			return nil, err
		}
	}

	// Update router assignments: delete all and re-insert
	_, err = tx.Exec(ctx, "DELETE FROM user_routers WHERE user_id = $1", userID)
	if err != nil {
		s.logger.Errorf("Error deleting old router assignments: %v", err)
		return nil, err
	}

	for _, routerName := range req.Routers {
		_, err = tx.Exec(ctx, `
			INSERT INTO user_routers (user_id, router_name, created_at)
			VALUES ($1, $2, $3)
		`, userID, routerName, time.Now())
		if err != nil {
			s.logger.Errorf("Error assigning router %s to user: %v", routerName, err)
			return nil, err
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		s.logger.Errorf("Error committing transaction: %v", err)
		return nil, err
	}

	s.logger.Infof("✅ User %d updated successfully", userID)

	// Return updated user
	return s.GetUserByID(userID)
}

// DeleteUser soft deletes a user (sets is_active = false)
func (s *UserService) DeleteUser(userID int) error {
	// Check if user exists
	var exists bool
	err := s.db.Pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
	if err != nil {
		s.logger.Errorf("Error checking user existence: %v", err)
		return err
	}
	if !exists {
		return errors.New("user not found")
	}

	// Soft delete (set is_active = false)
	_, err = s.db.Pool.Exec(context.Background(), `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`, time.Now(), userID)
	if err != nil {
		s.logger.Errorf("Error deleting user: %v", err)
		return err
	}

	s.logger.Infof("✅ User %d deleted successfully", userID)
	return nil
}

// HardDeleteUser permanently deletes a user (use with caution)
func (s *UserService) HardDeleteUser(userID int) error {
	// Delete user (cascade will delete user_routers)
	result, err := s.db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		s.logger.Errorf("Error hard deleting user: %v", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	s.logger.Warnf("⚠️ User %d permanently deleted", userID)
	return nil
}

// GetUserRouters retrieves all router names assigned to a user
func (s *UserService) GetUserRouters(userID int) ([]string, error) {
	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT router_name
		FROM user_routers
		WHERE user_id = $1
		ORDER BY router_name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routers []string
	for rows.Next() {
		var routerName string
		if err := rows.Scan(&routerName); err != nil {
			s.logger.Errorf("Error scanning router name: %v", err)
			continue
		}
		routers = append(routers, routerName)
	}

	return routers, nil
}

// CanAccessRouter checks if a user can access a specific router
func (s *UserService) CanAccessRouter(userID int, routerName string) (bool, error) {
	// Check if user is active
	var isActive bool
	err := s.db.Pool.QueryRow(context.Background(), "SELECT is_active FROM users WHERE id = $1", userID).Scan(&isActive)
	if err != nil {
		return false, err
	}
	if !isActive {
		return false, nil
	}

	// Check if user has access to this router
	var exists bool
	err = s.db.Pool.QueryRow(context.Background(), `
		SELECT EXISTS(
			SELECT 1 FROM user_routers
			WHERE user_id = $1 AND router_name = $2
		)
	`, userID, routerName).Scan(&exists)

	return exists, err
}

// GetRouterUsers retrieves all users who have access to a specific router
func (s *UserService) GetRouterUsers(routerName string) ([]UserWithRouters, error) {
	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT DISTINCT u.id, u.username, u.full_name, u.email, u.role, u.is_active,
		       u.created_at, u.updated_at, u.last_login_at
		FROM users u
		INNER JOIN user_routers ur ON u.id = ur.user_id
		WHERE ur.router_name = $1 AND u.is_active = true
		ORDER BY u.username ASC
	`, routerName)
	if err != nil {
		s.logger.Errorf("Error getting router users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []UserWithRouters
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Role,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
		if err != nil {
			s.logger.Errorf("Error scanning user: %v", err)
			continue
		}

		routers, _ := s.GetUserRouters(user.ID)
		users = append(users, UserWithRouters{
			User:    user,
			Routers: routers,
		})
	}

	return users, nil
}

// GetUserStats returns statistics about a user
func (s *UserService) GetUserStats(userID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get router count
	var routerCount int
	err := s.db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM user_routers WHERE user_id = $1", userID).Scan(&routerCount)
	if err != nil {
		return nil, err
	}
	stats["router_count"] = routerCount

	// Add more stats as needed (login count, last activity, etc.)
	stats["user_id"] = userID

	return stats, nil
}

// SearchUsers searches users by username, full name, or email
func (s *UserService) SearchUsers(query string) ([]UserWithRouters, error) {
	searchPattern := fmt.Sprintf("%%%s%%", query)

	rows, err := s.db.Pool.Query(context.Background(), `
		SELECT id, username, full_name, email, role, is_active, created_at, updated_at, last_login_at
		FROM users
		WHERE username ILIKE $1 OR full_name ILIKE $1 OR email ILIKE $1
		ORDER BY username ASC
		LIMIT 50
	`, searchPattern)
	if err != nil {
		s.logger.Errorf("Error searching users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []UserWithRouters
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Email, &user.Role,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
		if err != nil {
			s.logger.Errorf("Error scanning user: %v", err)
			continue
		}

		routers, _ := s.GetUserRouters(user.ID)
		users = append(users, UserWithRouters{
			User:    user,
			Routers: routers,
		})
	}

	return users, nil
}
