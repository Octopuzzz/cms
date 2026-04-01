package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cms-backend/internal/database"
	"cms-backend/internal/models"

	"gorm.io/gorm"
)

// DynamicDataService handles CRUD on dynamically created service tables
type DynamicDataService struct {
	connManager    *database.ConnectionManager
	serviceService *ServiceService
}

// NewDynamicDataService creates a new DynamicDataService
func NewDynamicDataService(cm *database.ConnectionManager, ss *ServiceService) *DynamicDataService {
	return &DynamicDataService{connManager: cm, serviceService: ss}
}

// ListDataRequest holds query params for listing data
type ListDataRequest struct {
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
	Search    string
	Filters   map[string]interface{}
	Joins     []string // field names to join (relation fields)
}

// CreateData inserts a new record in the service's dynamic table
func (s *DynamicDataService) CreateData(ctx context.Context, slug string, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	svc, db, err := s.resolveServiceDB(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Validate data against field definitions
	if err := s.validateData(ctx, svc, data); err != nil {
		return nil, err
	}

	// Build insert
	data["created_by"] = userID
	data["updated_by"] = userID

	cols := make([]string, 0)
	vals := make([]interface{}, 0)
	placeholders := make([]string, 0)

	for k, v := range data {
		cols = append(cols, k)
		vals = append(vals, v)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		svc.DbTableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	result := db.WithContext(ctx).Exec(query, vals...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create data: %w", result.Error)
	}

	// Get the inserted record
	var row map[string]interface{}
	db.WithContext(ctx).Raw(fmt.Sprintf("SELECT * FROM %s WHERE id = last_insert_rowid()", svc.DbTableName)).Scan(&row)
	if row == nil {
		// For non-SQLite, query by rowid
		row = data
	}

	return row, nil
}

// GetData retrieves a single record
func (s *DynamicDataService) GetData(ctx context.Context, slug string, id uint, joins []string) (map[string]interface{}, error) {
	svc, db, err := s.resolveServiceDB(ctx, slug)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ? AND deleted_at IS NULL", svc.DbTableName)
	var row map[string]interface{}
	if err := db.WithContext(ctx).Raw(query, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row == nil || len(row) == 0 {
		return nil, errors.New("record not found")
	}

	// Handle relation joins
	if len(joins) > 0 {
		row, err = s.applyJoins(ctx, db, svc, row, joins)
		if err != nil {
			return nil, err
		}
	}

	return row, nil
}

// ListData returns paginated records with optional joins and filters
func (s *DynamicDataService) ListData(ctx context.Context, slug string, req *ListDataRequest) ([]map[string]interface{}, int64, error) {
	svc, db, err := s.resolveServiceDB(ctx, slug)
	if err != nil {
		return nil, 0, err
	}

	// Count query
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE deleted_at IS NULL", svc.DbTableName)
	db.WithContext(ctx).Raw(countQuery).Scan(&total)

	// Data query
	sortBy := "id"
	if req.SortBy != "" {
		// Validate against allowed fields
		validFields := map[string]bool{
			"id":         true,
			"created_at": true,
			"updated_at": true,
			"deleted_at": true,
			"created_by": true,
			"updated_by": true,
		}
		for _, field := range svc.Fields {
			validFields[field.Name] = true
		}

		if !validFields[req.SortBy] {
			return nil, 0, fmt.Errorf("invalid sort field: %s", req.SortBy)
		}
		sortBy = req.SortBy
	}
	sortOrder := "DESC"
	if strings.ToUpper(req.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	query := fmt.Sprintf("SELECT * FROM %s WHERE deleted_at IS NULL ORDER BY %s %s LIMIT ? OFFSET ?",
		svc.DbTableName, sortBy, sortOrder)

	var rows []map[string]interface{}
	if err := db.WithContext(ctx).Raw(query, req.PageSize, offset).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	// Apply joins
	if len(req.Joins) > 0 && len(rows) > 0 {
		for i, row := range rows {
			rows[i], _ = s.applyJoins(ctx, db, svc, row, req.Joins)
		}
	}

	return rows, total, nil
}

// UpdateData updates a record in the dynamic table
func (s *DynamicDataService) UpdateData(ctx context.Context, slug string, id uint, data map[string]interface{}, userID uint) (map[string]interface{}, error) {
	svc, db, err := s.resolveServiceDB(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Validate data
	if err := s.validateData(ctx, svc, data); err != nil {
		return nil, err
	}

	data["updated_by"] = userID

	setClauses := make([]string, 0)
	vals := make([]interface{}, 0)
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", k))
		vals = append(vals, v)
	}
	vals = append(vals, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", svc.DbTableName, strings.Join(setClauses, ", "))
	if err := db.WithContext(ctx).Exec(query, vals...).Error; err != nil {
		return nil, err
	}

	return s.GetData(ctx, slug, id, nil)
}

// DeleteData soft-deletes a record
func (s *DynamicDataService) DeleteData(ctx context.Context, slug string, id uint) error {
	svc, db, err := s.resolveServiceDB(ctx, slug)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", svc.DbTableName)
	return db.WithContext(ctx).Exec(query, id).Error
}

// resolveServiceDB fetches the service and its target DB connection
func (s *DynamicDataService) resolveServiceDB(ctx context.Context, slug string) (*models.Service, *gorm.DB, error) {
	svc, err := s.serviceService.GetServiceBySlug(ctx, slug)
	if err != nil {
		return nil, nil, err
	}

	conn, err := s.connManager.GetConnectionForService(svc)
	if err != nil {
		return nil, nil, err
	}
	return svc, conn.DB, nil
}

// validateData validates incoming data against field definitions
func (s *DynamicDataService) validateData(_ context.Context, svc *models.Service, data map[string]interface{}) error {
	var validationErrors []string
	for _, field := range svc.Fields {
		val, exists := data[field.Name]
		if field.IsRequired && (!exists || val == nil || val == "") {
			validationErrors = append(validationErrors, fmt.Sprintf("field '%s' is required", field.Label))
			continue
		}
		if !exists {
			continue
		}

		// Run field validations
		for _, v := range field.Validations {
			switch v.Type {
			case "email":
				if str, ok := val.(string); ok && !isEmail(str) {
					msg := v.Message
					if msg == "" {
						msg = fmt.Sprintf("field '%s' must be a valid email", field.Label)
					}
					validationErrors = append(validationErrors, msg)
				}
			case "min":
				if v.Value != nil {
					if str, ok := val.(string); ok {
						minVal, _ := toFloat(v.Value)
						if float64(len(str)) < minVal {
							validationErrors = append(validationErrors, fmt.Sprintf("field '%s' must be at least %.0f characters", field.Label, minVal))
						}
					}
				}
			case "max":
				if v.Value != nil {
					if str, ok := val.(string); ok {
						maxVal, _ := toFloat(v.Value)
						if float64(len(str)) > maxVal {
							validationErrors = append(validationErrors, fmt.Sprintf("field '%s' must be at most %.0f characters", field.Label, maxVal))
						}
					}
				}
			}
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}
	return nil
}

// applyJoins enriches records with related data
func (s *DynamicDataService) applyJoins(ctx context.Context, db *gorm.DB, svc *models.Service, row map[string]interface{}, joinFields []string) (map[string]interface{}, error) {
	joinSet := make(map[string]bool)
	for _, j := range joinFields {
		joinSet[j] = true
	}

	for _, field := range svc.Fields {
		if field.Type != models.FieldTypeRelation || !joinSet[field.Name] {
			continue
		}
		if field.RelationConfig == nil {
			continue
		}

		relatedVal, ok := row[field.Name]
		if !ok || relatedVal == nil {
			continue
		}

		// Get the related service
		var relatedSvc models.Service
		if err := db.First(&relatedSvc, field.RelationConfig.RelatedServiceID).Error; err != nil {
			continue
		}

		var relatedRow map[string]interface{}
		db.WithContext(ctx).Raw(
			fmt.Sprintf("SELECT * FROM %s WHERE %s = ? AND deleted_at IS NULL LIMIT 1",
				relatedSvc.DbTableName, field.RelationConfig.RelatedField),
			relatedVal,
		).Scan(&relatedRow)

		if relatedRow != nil {
			row[field.Name+"_data"] = relatedRow
		}
	}

	return row, nil
}

func isEmail(s string) bool {
	return strings.Contains(s, "@") && strings.Contains(s, ".")
}

func toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case float32:
		return float64(val), true
	}
	return 0, false
}
