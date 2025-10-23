package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// DB holds the database connection pool
type DB struct {
	Pool   *pgxpool.Pool
	Logger *logrus.Logger
}

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDB creates a new database connection pool
func NewDB(logger *logrus.Logger) (*DB, error) {
	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")

	// If DATABASE_URL not set, construct from individual env vars
	if databaseURL == "" {
		config := Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   getEnvOrDefault("DB_NAME", "nat_management"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "require"),
		}

		if config.Password == "" {
			return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
		}

		databaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode,
		)
	}

	logger.Info("🔌 Connecting to PostgreSQL database...")

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = 25                      // Maximum connections
	poolConfig.MinConns = 5                       // Minimum connections
	poolConfig.MaxConnLifetime = time.Hour        // Connection lifetime
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Idle connection timeout
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Health check interval

	// IMPORTANT: Disable prepared statement cache for Supabase Transaction Pooler
	// Transaction pooler doesn't support prepared statements well
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Create connection pool with longer timeout for Neon serverless cold start
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection with retry for Neon serverless cold start
	logger.Info("🔄 Testing database connection (this may take a moment for serverless cold start)...")
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.Infof("⏳ Retry %d/%d - Waiting for database to wake up...", i, maxRetries)
			time.Sleep(time.Duration(i*2) * time.Second) // Exponential backoff: 2s, 4s
		}

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := pool.Ping(pingCtx)
		pingCancel()

		if err == nil {
			break // Success!
		}

		lastErr = err
		logger.Warnf("⚠️ Ping attempt %d failed: %v", i+1, err)
	}

	if lastErr != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database after %d retries: %w", maxRetries, lastErr)
	}

	logger.Info("✅ PostgreSQL connection established successfully!")
	logger.Infof("📊 Connection pool: min=%d, max=%d", poolConfig.MinConns, poolConfig.MaxConns)

	return &DB{
		Pool:   pool,
		Logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Logger.Info("🔌 Closing database connection pool...")
		db.Pool.Close()
		db.Logger.Info("✅ Database connection closed")
	}
}

// Ping checks if the database is reachable
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns connection pool statistics
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// GetConnection returns a connection from the pool
func (db *DB) GetConnection(ctx context.Context) (*pgxpool.Conn, error) {
	return db.Pool.Acquire(ctx)
}

// Helper function to get environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RunMigration runs a SQL migration file (for development/testing)
func (db *DB) RunMigration(ctx context.Context, sqlContent string) error {
	db.Logger.Info("🔄 Running database migration...")

	_, err := db.Pool.Exec(ctx, sqlContent)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	db.Logger.Info("✅ Migration completed successfully")
	return nil
}

// HealthCheck performs a comprehensive database health check
func (db *DB) HealthCheck(ctx context.Context) error {
	// Check connection
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check pool stats
	stats := db.Stats()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no database connections available")
	}

	// Execute simple query
	var result int
	err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected test query result: %d", result)
	}

	db.Logger.Debugf("✅ Database health check passed (connections: %d)", stats.TotalConns())
	return nil
}
