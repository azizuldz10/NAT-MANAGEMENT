package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nat-management-app/internal/database"
	"nat-management-app/internal/models"

	"github.com/sirupsen/logrus"
)

// ActivityLogService handles activity log operations
type ActivityLogService struct {
	db     *database.DB
	logger *logrus.Logger
}

// NewActivityLogService creates a new activity log service
func NewActivityLogService(db *database.DB, logger *logrus.Logger) *ActivityLogService {
	return &ActivityLogService{
		db:     db,
		logger: logger,
	}
}

// CreateLog creates a new activity log entry
func (s *ActivityLogService) CreateLog(log *models.ActivityLogCreate) error {
	// Set default status if not provided
	if log.Status == "" {
		log.Status = models.StatusSuccess
	}

	// Marshal metadata to JSON
	var metadataJSON []byte
	var err error
	if log.Metadata != nil {
		metadataJSON, err = json.Marshal(log.Metadata)
		if err != nil {
			s.logger.Errorf("Failed to marshal metadata: %v", err)
			metadataJSON = nil
		}
	}

	// Marshal device info to JSON
	var deviceInfoJSON []byte
	if log.DeviceInfo != nil {
		deviceInfoJSON, err = json.Marshal(log.DeviceInfo)
		if err != nil {
			s.logger.Errorf("Failed to marshal device info: %v", err)
			deviceInfoJSON = nil
		}
	}

	query := `
		INSERT INTO activity_logs (
			user_id, username, user_role, action_type, resource_type,
			resource_id, description, ip_address, user_agent,
			status, error_message, duration_ms, device_info, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at
	`

	var id int
	var createdAt time.Time

	err = s.db.Pool.QueryRow(
		context.Background(),
		query,
		log.UserID,
		log.Username,
		log.UserRole,
		log.ActionType,
		log.ResourceType,
		log.ResourceID,
		log.Description,
		log.IPAddress,
		log.UserAgent,
		log.Status,
		log.ErrorMessage,
		log.DurationMs,
		deviceInfoJSON,
		metadataJSON,
	).Scan(&id, &createdAt)

	if err != nil {
		s.logger.Errorf("Failed to create activity log: %v", err)
		return err
	}

	s.logger.Debugf("Activity log created: ID=%d, User=%s, Action=%s, Duration=%vms",
		id, log.Username, log.ActionType, log.DurationMs)
	return nil
}

// GetLogs retrieves activity logs with optional filters
func (s *ActivityLogService) GetLogs(filter *models.ActivityLogFilter) ([]models.ActivityLog, int, error) {
	// Build query with filters
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT
			id, user_id, username, user_role, action_type, resource_type,
			resource_id, description, ip_address, user_agent,
			status, error_message, metadata, created_at
		FROM activity_logs
		WHERE 1=1
	`)

	countQueryBuilder := strings.Builder{}
	countQueryBuilder.WriteString("SELECT COUNT(*) FROM activity_logs WHERE 1=1")

	args := []interface{}{}
	argCount := 1

	// Add filters
	if filter.UserID != nil {
		queryBuilder.WriteString(fmt.Sprintf(" AND user_id = $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND user_id = $%d", argCount))
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.Username != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND username ILIKE $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND username ILIKE $%d", argCount))
		args = append(args, "%"+filter.Username+"%")
		argCount++
	}

	if filter.ActionType != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND action_type = $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND action_type = $%d", argCount))
		args = append(args, filter.ActionType)
		argCount++
	}

	if filter.ResourceType != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND resource_type = $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND resource_type = $%d", argCount))
		args = append(args, filter.ResourceType)
		argCount++
	}

	if filter.Status != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND status = $%d", argCount))
		args = append(args, filter.Status)
		argCount++
	}

	if !filter.StartDate.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND created_at >= $%d", argCount))
		args = append(args, filter.StartDate)
		argCount++
	}

	if !filter.EndDate.IsZero() {
		queryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argCount))
		countQueryBuilder.WriteString(fmt.Sprintf(" AND created_at <= $%d", argCount))
		args = append(args, filter.EndDate)
		argCount++
	}

	// Get total count
	var totalCount int
	err := s.db.Pool.QueryRow(context.Background(), countQueryBuilder.String(), args...).Scan(&totalCount)
	if err != nil {
		s.logger.Errorf("Failed to count activity logs: %v", err)
		return nil, 0, err
	}

	// Add ordering
	queryBuilder.WriteString(" ORDER BY created_at DESC")

	// Add pagination
	if filter.Limit > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", argCount))
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", argCount))
		args = append(args, filter.Offset)
		argCount++
	}

	// Execute query
	rows, err := s.db.Pool.Query(context.Background(), queryBuilder.String(), args...)
	if err != nil {
		s.logger.Errorf("Failed to query activity logs: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	logs := []models.ActivityLog{}
	for rows.Next() {
		var log models.ActivityLog
		var metadataJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.UserRole,
			&log.ActionType,
			&log.ResourceType,
			&log.ResourceID,
			&log.Description,
			&log.IPAddress,
			&log.UserAgent,
			&log.Status,
			&log.ErrorMessage,
			&metadataJSON,
			&log.CreatedAt,
		)

		if err != nil {
			s.logger.Errorf("Failed to scan activity log row: %v", err)
			continue
		}

		// Unmarshal metadata
		if metadataJSON != nil {
			err = json.Unmarshal(metadataJSON, &log.Metadata)
			if err != nil {
				s.logger.Warnf("Failed to unmarshal metadata for log %d: %v", log.ID, err)
			}
		}

		logs = append(logs, log)
	}

	return logs, totalCount, nil
}

// GetLogByID retrieves a single activity log by ID
func (s *ActivityLogService) GetLogByID(id int) (*models.ActivityLog, error) {
	query := `
		SELECT
			id, user_id, username, user_role, action_type, resource_type,
			resource_id, description, ip_address, user_agent,
			status, error_message, duration_ms, device_info, metadata, created_at
		FROM activity_logs
		WHERE id = $1
	`

	var log models.ActivityLog
	var metadataJSON []byte
	var deviceInfoJSON []byte

	err := s.db.Pool.QueryRow(context.Background(), query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.Username,
		&log.UserRole,
		&log.ActionType,
		&log.ResourceType,
		&log.ResourceID,
		&log.Description,
		&log.IPAddress,
		&log.UserAgent,
		&log.Status,
		&log.ErrorMessage,
		&log.DurationMs,
		&deviceInfoJSON,
		&metadataJSON,
		&log.CreatedAt,
	)

	if err != nil {
		s.logger.Errorf("Failed to get activity log by ID %d: %v", id, err)
		return nil, err
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		err = json.Unmarshal(metadataJSON, &log.Metadata)
		if err != nil {
			s.logger.Warnf("Failed to unmarshal metadata for log %d: %v", log.ID, err)
		}
	}

	// Unmarshal device info
	if deviceInfoJSON != nil {
		var deviceInfo models.DeviceInfo
		err = json.Unmarshal(deviceInfoJSON, &deviceInfo)
		if err != nil {
			s.logger.Warnf("Failed to unmarshal device info for log %d: %v", log.ID, err)
		} else {
			log.DeviceInfo = &deviceInfo
		}
	}

	return &log, nil
}

// GetLogStats retrieves statistics about activity logs
func (s *ActivityLogService) GetLogStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total logs
	var totalLogs int
	err := s.db.Pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM activity_logs").Scan(&totalLogs)
	if err != nil {
		return nil, err
	}
	stats["total_logs"] = totalLogs

	// Logs by action type
	actionQuery := `
		SELECT action_type, COUNT(*) as count
		FROM activity_logs
		GROUP BY action_type
		ORDER BY count DESC
	`
	rows, err := s.db.Pool.Query(context.Background(), actionQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actionStats := make(map[string]int)
	for rows.Next() {
		var actionType string
		var count int
		if err := rows.Scan(&actionType, &count); err == nil {
			actionStats[actionType] = count
		}
	}
	stats["by_action_type"] = actionStats

	// Recent activity (last 24 hours)
	var recentCount int
	err = s.db.Pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM activity_logs WHERE created_at >= NOW() - INTERVAL '24 hours'",
	).Scan(&recentCount)
	if err == nil {
		stats["last_24h"] = recentCount
	}

	// Failed actions
	var failedCount int
	err = s.db.Pool.QueryRow(
		context.Background(),
		"SELECT COUNT(*) FROM activity_logs WHERE status IN ('FAILED', 'ERROR')",
	).Scan(&failedCount)
	if err == nil {
		stats["failed_actions"] = failedCount
	}

	return stats, nil
}

// DeleteOldLogs deletes logs older than specified days (for maintenance)
func (s *ActivityLogService) DeleteOldLogs(daysToKeep int) (int, error) {
	query := `
		DELETE FROM activity_logs
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := s.db.Pool.Exec(context.Background(), query, daysToKeep)
	if err != nil {
		s.logger.Errorf("Failed to delete old logs: %v", err)
		return 0, err
	}

	rowsAffected := result.RowsAffected()
	s.logger.Infof("Deleted %d old activity logs (older than %d days)", rowsAffected, daysToKeep)
	return int(rowsAffected), nil
}
