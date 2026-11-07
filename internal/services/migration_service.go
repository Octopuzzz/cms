// Package services provides the schema migration engine for the CMS backend.
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"cms-backend/internal/database"
	"cms-backend/internal/models"

	"gorm.io/gorm"
)

// identifierRegex validates SQL identifiers (table/column names)
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,63}$`)

// validateIdentifier checks that a string is a safe SQL identifier
func validateIdentifier(name string) error {
	if !identifierRegex.MatchString(name) {
		return fmt.Errorf("invalid identifier: %q", name)
	}
	return nil
}

// MigrationService handles schema migration operations
type MigrationService struct {
	connManager *database.ConnectionManager
}

// NewMigrationService creates a new MigrationService
func NewMigrationService(cm *database.ConnectionManager) *MigrationService {
	return &MigrationService{connManager: cm}
}

// MigrationRequest describes a schema migration to apply
type MigrationRequest struct {
	ServiceID   uint                 `json:"service_id" binding:"required"`
	Description string               `json:"description" binding:"required"`
	Operations  []MigrationOperation `json:"operations" binding:"required,min=1"`
}

// MigrationOperation is a single schema change
type MigrationOperation struct {
	Type       string `json:"type" binding:"required,oneof=add_column drop_column rename_column change_type"`
	Column     string `json:"column" binding:"required"`
	NewName    string `json:"new_name,omitempty"`
	ColumnType string `json:"column_type,omitempty"`
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default,omitempty"`
}

// CreateMigration plans and executes a schema migration for a service
func (s *MigrationService) CreateMigration(ctx context.Context, req *MigrationRequest, userID uint) (*models.Migration, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}

	// Fetch the service
	var svc models.Service
	if err := conn.DB.WithContext(ctx).Preload("Fields").First(&svc, req.ServiceID).Error; err != nil {
		return nil, errors.New("service not found")
	}

	// Snapshot current schema
	schemaSnapshot, _ := json.Marshal(svc.Fields)

	migration := &models.Migration{
		ServiceID:      req.ServiceID,
		Description:    req.Description,
		Status:         "pending",
		SchemaSnapshot: string(schemaSnapshot),
		BaseModel:      models.BaseModel{CreatedBy: userID, UpdatedBy: userID},
	}

	if err := conn.DB.WithContext(ctx).Create(migration).Error; err != nil {
		return nil, fmt.Errorf("failed to create migration record: %w", err)
	}

	// Determine target DB
	targetConn := conn
	if svc.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*svc.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}

	// Apply operations
	var applyErr error
	for _, op := range req.Operations {
		if execErr := s.applyOperation(targetConn.DB, svc.DbTableName, &op); execErr != nil {
			applyErr = execErr
			break
		}
	}

	if applyErr != nil {
		migration.Status = "failed"
		migration.ErrorMessage = applyErr.Error()
		conn.DB.Save(migration)
		return migration, fmt.Errorf("migration failed: %w", applyErr)
	}

	// Sync field metadata for add_column operations
	for _, op := range req.Operations {
		if op.Type == "add_column" {
			field := models.Field{
				ServiceID:  req.ServiceID,
				Name:       op.Column,
				Label:      op.Column,
				Type:       sqlTypeToFieldType(op.ColumnType),
				IsNullable: op.Nullable,
			}
			field.CreatedBy = userID
			conn.DB.Create(&field)
		}
		if op.Type == "drop_column" {
			conn.DB.Where("service_id = ? AND name = ?", req.ServiceID, op.Column).Delete(&models.Field{})
		}
		if op.Type == "rename_column" && op.NewName != "" {
			conn.DB.Model(&models.Field{}).
				Where("service_id = ? AND name = ?", req.ServiceID, op.Column).
				Update("name", op.NewName)
		}
	}

	migration.Status = "applied"
	now := time.Now()
	migration.AppliedAt = &now
	conn.DB.Save(migration)
	return migration, nil
}

// applyOperation executes a single DDL operation with identifier validation
func (s *MigrationService) applyOperation(db *gorm.DB, tableName string, op *MigrationOperation) error {
	// Validate table name
	if err := validateIdentifier(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	// Validate column name
	if err := validateIdentifier(op.Column); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	switch op.Type {
	case "add_column":
		colType := op.ColumnType
		if colType == "" {
			colType = "TEXT"
		}
		if err := validateIdentifier(colType); colType != "TEXT" && err != nil {
			// Allow compound types like "TIMESTAMP WITH TIME ZONE" via whitelist
			allowed := map[string]bool{
				"TEXT": true, "INTEGER": true, "INT": true, "BIGINT": true,
				"REAL": true, "FLOAT": true, "DOUBLE": true, "DOUBLE PRECISION": true,
				"BOOLEAN": true, "DATE": true, "DATETIME": true, "TIMESTAMP": true,
				"TIMESTAMP WITH TIME ZONE": true, "JSON": true, "JSONB": true,
				"UUID": true, "VARCHAR(255)": true, "TINYINT(1)": true,
				"DATETIME(3)": true, "LONGTEXT": true,
			}
			if !allowed[colType] {
				return fmt.Errorf("unsupported column type: %q", colType)
			}
		}
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, op.Column, colType)
		if !op.Nullable {
			sql += " NOT NULL"
		}
		if op.Default != "" {
			// Use parameterized default value to prevent injection
			sql += " DEFAULT ?"
			return db.Exec(sql, op.Default).Error
		}
		return db.Exec(sql).Error

	case "drop_column":
		return db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, op.Column)).Error

	case "rename_column":
		if op.NewName == "" {
			return errors.New("new_name is required for rename_column")
		}
		if err := validateIdentifier(op.NewName); err != nil {
			return fmt.Errorf("invalid new column name: %w", err)
		}
		return db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, op.Column, op.NewName)).Error

	case "change_type":
		if op.ColumnType == "" {
			return errors.New("column_type is required for change_type")
		}
		return db.Exec(fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", tableName, op.Column, op.ColumnType)).Error

	default:
		return fmt.Errorf("unsupported operation type: %s", op.Type)
	}
}

// ListMigrations returns all migrations for a service
func (s *MigrationService) ListMigrations(ctx context.Context, serviceID uint) ([]models.Migration, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var migrations []models.Migration
	if err := conn.DB.WithContext(ctx).
		Where("service_id = ?", serviceID).
		Order("created_at DESC").
		Find(&migrations).Error; err != nil {
		return nil, err
	}
	return migrations, nil
}

// RollbackMigration marks a migration as rolled back.
// Note: This updates the migration status but does not automatically reverse DDL changes.
// For full schema rollback, apply a new migration with the inverse operations.
func (s *MigrationService) RollbackMigration(ctx context.Context, migrationID uint) error {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return err
	}

	var migration models.Migration
	if err := conn.DB.WithContext(ctx).First(&migration, migrationID).Error; err != nil {
		return errors.New("migration not found")
	}

	if migration.Status != "applied" {
		return errors.New("only applied migrations can be rolled back")
	}

	migration.Status = "rolled_back"
	return conn.DB.Save(&migration).Error
}

// sqlTypeToFieldType converts a SQL type string to a FieldType
func sqlTypeToFieldType(sqlType string) models.FieldType {
	switch sqlType {
	case "INTEGER", "INT", "BIGINT":
		return models.FieldTypeInteger
	case "REAL", "DOUBLE", "FLOAT", "DOUBLE PRECISION":
		return models.FieldTypeFloat
	case "BOOLEAN", "TINYINT(1)":
		return models.FieldTypeBoolean
	case "DATE":
		return models.FieldTypeDate
	case "DATETIME", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE":
		return models.FieldTypeDateTime
	case "JSON", "JSONB":
		return models.FieldTypeJSON
	case "UUID":
		return models.FieldTypeUUID
	default:
		return models.FieldTypeString
	}
}
