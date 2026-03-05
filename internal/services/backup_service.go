// Package services provides the backup engine for the CMS backend.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cms-backend/internal/database"
	"cms-backend/internal/models"
)

// BackupService handles backup and restore operations
type BackupService struct {
	connManager *database.ConnectionManager
}

// NewBackupService creates a new BackupService
func NewBackupService(cm *database.ConnectionManager) *BackupService {
	return &BackupService{connManager: cm}
}

// BackupResult contains the result of a backup operation
type BackupResult struct {
	BackupID  uint                   `json:"backup_id"`
	ServiceID uint                   `json:"service_id"`
	Type      string                 `json:"type"`
	Schema    map[string]interface{} `json:"schema,omitempty"`
	Data      []map[string]interface{} `json:"data,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// CreateBackup creates a full snapshot backup of a service (schema + data)
func (s *BackupService) CreateBackup(ctx context.Context, serviceID uint, userID uint) (*models.Backup, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}

	// Fetch service
	var svc models.Service
	if err := conn.DB.WithContext(ctx).Preload("Fields").First(&svc, serviceID).Error; err != nil {
		return nil, errors.New("service not found")
	}

	// Determine target DB for data
	targetConn := conn
	if svc.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*svc.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}

	// Snapshot schema
	schemaBytes, _ := json.Marshal(svc.Fields)

	// Snapshot data
	var rows []map[string]interface{}
	if err := targetConn.DB.Table(svc.DbTableName).
		Where("deleted_at IS NULL").
		Find(&rows).Error; err != nil {
		rows = []map[string]interface{}{}
	}
	dataBytes, _ := json.Marshal(rows)

	backup := &models.Backup{
		ServiceID:  serviceID,
		Type:       "snapshot",
		Status:     "completed",
		SchemaData: string(schemaBytes),
		TableData:  string(dataBytes),
		RowCount:   len(rows),
		BaseModel:  models.BaseModel{CreatedBy: userID, UpdatedBy: userID},
	}

	if err := conn.DB.WithContext(ctx).Create(backup).Error; err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	return backup, nil
}

// ListBackups returns all backups for a service
func (s *BackupService) ListBackups(ctx context.Context, serviceID uint) ([]models.Backup, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var backups []models.Backup
	if err := conn.DB.WithContext(ctx).
		Where("service_id = ?", serviceID).
		Order("created_at DESC").
		Find(&backups).Error; err != nil {
		return nil, err
	}
	return backups, nil
}

// GetBackup retrieves a single backup by ID
func (s *BackupService) GetBackup(ctx context.Context, backupID uint) (*models.Backup, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var backup models.Backup
	if err := conn.DB.WithContext(ctx).First(&backup, backupID).Error; err != nil {
		return nil, errors.New("backup not found")
	}
	return &backup, nil
}

// RestoreBackup restores data from a backup into the service's table
func (s *BackupService) RestoreBackup(ctx context.Context, backupID uint) error {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return err
	}

	var backup models.Backup
	if err := conn.DB.WithContext(ctx).First(&backup, backupID).Error; err != nil {
		return errors.New("backup not found")
	}

	var svc models.Service
	if err := conn.DB.WithContext(ctx).First(&svc, backup.ServiceID).Error; err != nil {
		return errors.New("service not found")
	}

	// Determine target DB
	targetConn := conn
	if svc.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*svc.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}

	// Validate table name to prevent SQL injection
	if !identifierRegex.MatchString(svc.DbTableName) {
		return fmt.Errorf("invalid table name: %q", svc.DbTableName)
	}

	// Parse data
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(backup.TableData), &rows); err != nil {
		return fmt.Errorf("failed to parse backup data: %w", err)
	}

	// Clear existing data (soft-delete approach: mark as deleted)
	targetConn.DB.Exec(fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE deleted_at IS NULL", svc.DbTableName), time.Now())

	// Insert restored rows, collecting errors
	var restoreErrors int
	for _, row := range rows {
		delete(row, "id")
		delete(row, "deleted_at")
		if err := targetConn.DB.Table(svc.DbTableName).Create(row).Error; err != nil {
			restoreErrors++
		}
	}

	if restoreErrors > 0 {
		return fmt.Errorf("restore completed with %d error(s) out of %d rows", restoreErrors, len(rows))
	}

	return nil
}
