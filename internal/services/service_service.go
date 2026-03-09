// Package services provides the core CMS service management logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"cms-backend/internal/database"
	"cms-backend/internal/models"

	"gorm.io/gorm"
)

// ServiceService manages CMS services (collections)
type ServiceService struct {
	connManager *database.ConnectionManager
}

// NewServiceService creates a new ServiceService
func NewServiceService(cm *database.ConnectionManager) *ServiceService {
	return &ServiceService{connManager: cm}
}

// ---- Request / Response Types ----

type CreateServiceRequest struct {
	Name                 string               `json:"name" binding:"required,min=2,max=100"`
	Description          string               `json:"description"`
	IsPublic             bool                 `json:"is_public"`
	DatabaseConnectionID *uint                `json:"database_connection_id"`
	Fields               []CreateFieldRequest `json:"fields"`
	MenuConfig           *MenuConfigRequest   `json:"menu_config"`
}

type UpdateServiceRequest struct {
	Name        string               `json:"name" binding:"required,min=2,max=100"`
	Description string               `json:"description"`
	IsPublic    bool                 `json:"is_public"`
	IsActive    bool                 `json:"is_active"`
	Fields      []UpdateFieldRequest `json:"fields"`
	MenuConfig  *MenuConfigRequest   `json:"menu_config"`
}

type CreateFieldRequest struct {
	Name           string                  `json:"name" binding:"required"`
	Label          string                  `json:"label" binding:"required"`
	Type           models.FieldType        `json:"type" binding:"required"`
	Description    string                  `json:"description"`
	IsRequired     bool                    `json:"is_required"`
	IsUnique       bool                    `json:"is_unique"`
	IsIndex        bool                    `json:"is_index"`
	IsNullable     bool                    `json:"is_nullable"`
	DefaultValue   string                  `json:"default_value"`
	Placeholder    string                  `json:"placeholder"`
	HelpText       string                  `json:"help_text"`
	SortOrder      int                     `json:"sort_order"`
	Options        models.FieldOptions     `json:"options"`
	Validations    models.FieldValidations `json:"validations"`
	RelationConfig *models.RelationConfig  `json:"relation_config"`
}

type UpdateFieldRequest struct {
	ID             *uint                   `json:"id"`
	Name           string                  `json:"name"`
	Label          string                  `json:"label"`
	Type           models.FieldType        `json:"type"`
	Description    string                  `json:"description"`
	IsRequired     bool                    `json:"is_required"`
	IsUnique       bool                    `json:"is_unique"`
	IsNullable     bool                    `json:"is_nullable"`
	DefaultValue   string                  `json:"default_value"`
	Placeholder    string                  `json:"placeholder"`
	HelpText       string                  `json:"help_text"`
	SortOrder      int                     `json:"sort_order"`
	Options        models.FieldOptions     `json:"options"`
	Validations    models.FieldValidations `json:"validations"`
	RelationConfig *models.RelationConfig  `json:"relation_config"`
	Deleted        bool                    `json:"deleted"`
}

type MenuConfigRequest struct {
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
	IsVisible bool   `json:"is_visible"`
	ParentID  *uint  `json:"parent_id"`
}

type ListServicesFilter struct {
	Search   string
	IsActive *bool
	IsPublic *bool
}

// CreateService creates a new service, its fields, and auto-generates the dynamic table
func (s *ServiceService) CreateService(ctx context.Context, req *CreateServiceRequest, userID uint) (*models.Service, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}

	slug := toSlug(req.Name)
	tableName := "svc_" + slug

	// Check uniqueness
	var existing models.Service
	if err := conn.DB.Where("slug = ? OR db_table_name = ?", slug, tableName).First(&existing).Error; err == nil {
		return nil, errors.New("service with this name already exists")
	}

	svc := &models.Service{
		Name:                 req.Name,
		Slug:                 slug,
		Description:          req.Description,
		DbTableName:          tableName,
		IsActive:             true,
		IsPublic:             req.IsPublic,
		DatabaseConnectionID: req.DatabaseConnectionID,
		BaseModel:            models.BaseModel{CreatedBy: userID, UpdatedBy: userID},
	}

	for i, f := range req.Fields {
		field := models.Field{
			Name:           f.Name,
			Label:          f.Label,
			Type:           f.Type,
			Description:    f.Description,
			IsRequired:     f.IsRequired,
			IsUnique:       f.IsUnique,
			IsIndex:        f.IsIndex,
			IsNullable:     f.IsNullable,
			DefaultValue:   f.DefaultValue,
			Placeholder:    f.Placeholder,
			HelpText:       f.HelpText,
			SortOrder:      i,
			Options:        f.Options,
			Validations:    f.Validations,
			RelationConfig: f.RelationConfig,
		}
		field.CreatedBy = userID
		svc.Fields = append(svc.Fields, field)
	}

	if err := conn.DB.WithContext(ctx).Create(svc).Error; err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	// Auto-create dynamic table
	targetConn := conn
	if req.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*req.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}
	if err := s.createDynamicTable(targetConn.DB, svc); err != nil {
		// Non-fatal: log but don't rollback service creation
		_ = conn.DB.Model(svc).Update("is_active", false)
		return nil, fmt.Errorf("service created but failed to create table: %w", err)
	}

	// Auto-create menu entry
	if req.MenuConfig != nil || true {
		menu := &models.Menu{
			Name:      req.Name,
			Path:      "/" + slug,
			IsActive:  true,
			IsVisible: true,
			ServiceID: &svc.ID,
		}
		if req.MenuConfig != nil {
			menu.Icon = req.MenuConfig.Icon
			menu.SortOrder = req.MenuConfig.SortOrder
			menu.IsVisible = req.MenuConfig.IsVisible
			menu.ParentID = req.MenuConfig.ParentID
		}
		menu.CreatedBy = userID
		conn.DB.Create(menu)
	}

	// Reload with preloads
	return s.GetService(ctx, svc.ID)
}

// GetService retrieves a service by ID with all relations
func (s *ServiceService) GetService(ctx context.Context, id uint) (*models.Service, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var svc models.Service
	if err := conn.DB.WithContext(ctx).
		Preload("Fields").
		Preload("Menu").
		Preload("DatabaseConnection").
		First(&svc, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("service not found")
		}
		return nil, err
	}
	return &svc, nil
}

// GetServiceBySlug retrieves a service by slug
func (s *ServiceService) GetServiceBySlug(ctx context.Context, slug string) (*models.Service, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var svc models.Service
	if err := conn.DB.WithContext(ctx).
		Preload("Fields").
		Preload("Permissions.Role").
		Where("slug = ?", slug).
		First(&svc).Error; err != nil {
		return nil, errors.New("service not found")
	}
	return &svc, nil
}

// ListServices returns a paginated list of services
func (s *ServiceService) ListServices(ctx context.Context, page, pageSize int, filter *ListServicesFilter) ([]models.Service, int64, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, 0, err
	}

	q := conn.DB.WithContext(ctx).Model(&models.Service{}).Preload("Menu")

	if filter != nil {
		if filter.Search != "" {
			q = q.Where("name LIKE ? OR description LIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
		}
		if filter.IsActive != nil {
			q = q.Where("is_active = ?", *filter.IsActive)
		}
		if filter.IsPublic != nil {
			q = q.Where("is_public = ?", *filter.IsPublic)
		}
	}

	var total int64
	q.Count(&total)

	offset := (page - 1) * pageSize
	var services []models.Service
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&services).Error; err != nil {
		return nil, 0, err
	}
	return services, total, nil
}

// UpdateService updates service details and fields
func (s *ServiceService) UpdateService(ctx context.Context, id uint, req *UpdateServiceRequest, userID uint) (*models.Service, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}

	svc, err := s.GetService(ctx, id)
	if err != nil {
		return nil, err
	}

	svc.Name = req.Name
	svc.Description = req.Description
	svc.IsPublic = req.IsPublic
	svc.IsActive = req.IsActive
	svc.UpdatedBy = userID

	if err := conn.DB.WithContext(ctx).Save(svc).Error; err != nil {
		return nil, err
	}

	// Handle field updates
	targetConn := conn
	if svc.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*svc.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}

	for _, f := range req.Fields {
		if f.Deleted && f.ID != nil {
			// Delete field and drop column
			conn.DB.Delete(&models.Field{}, *f.ID)
			_ = targetConn.DB.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s", svc.DbTableName, f.Name))
			continue
		}
		if f.ID != nil {
			// Update existing field
			conn.DB.Model(&models.Field{}).Where("id = ?", *f.ID).Updates(map[string]interface{}{
				"name": f.Name, "label": f.Label, "type": string(f.Type),
				"is_required": f.IsRequired, "updated_by": userID,
			})
		} else {
			// Add new field
			newField := models.Field{
				ServiceID: svc.ID, Name: f.Name, Label: f.Label,
				Type: f.Type, IsRequired: f.IsRequired, IsNullable: f.IsNullable,
				Validations: f.Validations, Options: f.Options,
			}
			newField.CreatedBy = userID
			if err := conn.DB.Create(&newField).Error; err == nil {
				colDef := fieldTypeToSQL(f.Type, f.IsNullable, f.DefaultValue)
				targetConn.DB.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", svc.DbTableName, f.Name, colDef))
			}
		}
	}

	// Update menu
	if req.MenuConfig != nil {
		conn.DB.Model(&models.Menu{}).Where("service_id = ?", id).Updates(map[string]interface{}{
			"icon": req.MenuConfig.Icon, "sort_order": req.MenuConfig.SortOrder, "is_visible": req.MenuConfig.IsVisible,
		})
	}

	return s.GetService(ctx, id)
}

// DeleteService soft-deletes a service and drops its dynamic table
func (s *ServiceService) DeleteService(ctx context.Context, id uint) error {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return err
	}
	svc, err := s.GetService(ctx, id)
	if err != nil {
		return err
	}

	// Drop dynamic table
	targetConn := conn
	if svc.DatabaseConnectionID != nil {
		if tc, err := s.connManager.GetConnection(*svc.DatabaseConnectionID); err == nil {
			targetConn = tc
		}
	}
	_ = targetConn.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", svc.DbTableName))

	// Delete menu
	conn.DB.Where("service_id = ?", id).Delete(&models.Menu{})

	return conn.DB.WithContext(ctx).Delete(svc).Error
}

// SetPermissions sets RBAC permissions for a service
func (s *ServiceService) SetPermissions(ctx context.Context, serviceID uint, perms []models.ServicePermission) error {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return err
	}
	// Remove existing
	conn.DB.Where("service_id = ?", serviceID).Delete(&models.ServicePermission{})
	for i := range perms {
		perms[i].ServiceID = serviceID
	}
	return conn.DB.CreateInBatches(perms, 10).Error
}

// GetPermissionsForRole checks if a role has access to a service action
func (s *ServiceService) GetPermissionsForRole(ctx context.Context, serviceID, roleID uint) (*models.ServicePermission, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}
	var perm models.ServicePermission
	if err := conn.DB.Where("service_id = ? AND role_id = ?", serviceID, roleID).First(&perm).Error; err != nil {
		return nil, err
	}
	return &perm, nil
}

// createDynamicTable creates the dynamic table for a service
func (s *ServiceService) createDynamicTable(db *gorm.DB, svc *models.Service) error {
	cols := []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"created_at DATETIME",
		"updated_at DATETIME",
		"deleted_at DATETIME",
		"created_by INTEGER",
		"updated_by INTEGER",
	}

	for _, f := range svc.Fields {
		colDef := fieldTypeToSQL(f.Type, f.IsNullable, f.DefaultValue)
		constraint := ""
		if f.IsRequired {
			constraint += " NOT NULL"
		}
		cols = append(cols, fmt.Sprintf("%s %s%s", f.Name, colDef, constraint))
	}

	// Use dialect-aware DDL
	dialectSQL := detectDialect(db)
	var ddl string
	switch dialectSQL {
	case "postgres":
		ddl = buildPostgresDDL(svc.DbTableName, svc.Fields)
	case "mysql":
		ddl = buildMySQLDDL(svc.DbTableName, svc.Fields)
	default:
		ddl = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", svc.DbTableName, strings.Join(cols, ", "))
	}

	return db.Exec(ddl).Error
}

func detectDialect(db *gorm.DB) string {
	name := db.Dialector.Name()
	return strings.ToLower(name)
}

func buildPostgresDDL(tableName string, fields []models.Field) string {
	cols := []string{
		"id SERIAL PRIMARY KEY",
		"created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()",
		"updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()",
		"deleted_at TIMESTAMP WITH TIME ZONE",
		"created_by INTEGER",
		"updated_by INTEGER",
	}
	for _, f := range fields {
		col := f.Name + " " + fieldTypeToPostgres(f.Type)
		if f.IsRequired {
			col += " NOT NULL"
		}
		if f.IsUnique {
			col += " UNIQUE"
		}
		cols = append(cols, col)
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(cols, ", "))
}

func buildMySQLDDL(tableName string, fields []models.Field) string {
	cols := []string{
		"id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY",
		"created_at DATETIME(3)",
		"updated_at DATETIME(3)",
		"deleted_at DATETIME(3)",
		"created_by BIGINT UNSIGNED",
		"updated_by BIGINT UNSIGNED",
	}
	for _, f := range fields {
		col := f.Name + " " + fieldTypeToMySQL(f.Type)
		if f.IsRequired {
			col += " NOT NULL"
		}
		cols = append(cols, col)
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", tableName, strings.Join(cols, ", "))
}

func fieldTypeToSQL(ft models.FieldType, nullable bool, defaultVal string) string {
	col := fieldTypeToSQLite(ft)
	if defaultVal != "" {
		col += fmt.Sprintf(" DEFAULT '%s'", defaultVal)
	}
	return col
}

func fieldTypeToSQLite(ft models.FieldType) string {
	switch ft {
	case models.FieldTypeInteger:
		return "INTEGER"
	case models.FieldTypeFloat:
		return "REAL"
	case models.FieldTypeBoolean:
		return "INTEGER" // 0/1
	case models.FieldTypeDate, models.FieldTypeDateTime:
		return "DATETIME"
	case models.FieldTypeText, models.FieldTypeJSON:
		return "TEXT"
	default:
		return "TEXT"
	}
}

func fieldTypeToPostgres(ft models.FieldType) string {
	switch ft {
	case models.FieldTypeInteger:
		return "INTEGER"
	case models.FieldTypeFloat:
		return "DOUBLE PRECISION"
	case models.FieldTypeBoolean:
		return "BOOLEAN"
	case models.FieldTypeDate:
		return "DATE"
	case models.FieldTypeDateTime:
		return "TIMESTAMP WITH TIME ZONE"
	case models.FieldTypeJSON:
		return "JSONB"
	case models.FieldTypeUUID:
		return "UUID"
	default:
		return "TEXT"
	}
}

func fieldTypeToMySQL(ft models.FieldType) string {
	switch ft {
	case models.FieldTypeInteger:
		return "INT"
	case models.FieldTypeFloat:
		return "DOUBLE"
	case models.FieldTypeBoolean:
		return "TINYINT(1)"
	case models.FieldTypeDate:
		return "DATE"
	case models.FieldTypeDateTime:
		return "DATETIME(3)"
	case models.FieldTypeJSON:
		return "JSON"
	case models.FieldTypeText:
		return "LONGTEXT"
	default:
		return "VARCHAR(255)"
	}
}

var (
	nonAlphanumRegex = regexp.MustCompile(`[^a-z0-9]+`)
	spaceRegex       = regexp.MustCompile(`\s+`)
)

func toSlug(name string) string {
	s := strings.ToLower(name)
	s = spaceRegex.ReplaceAllString(s, "_")
	s = nonAlphanumRegex.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	// Ensure starts with letter
	if len(s) > 0 && !unicode.IsLetter(rune(s[0])) {
		s = "s_" + s
	}
	// Truncate
	if len(s) > 50 {
		s = s[:50]
	}
	_ = time.Now() // suppress time import warning
	return s
}

// GetAllActiveServices returns all services that are currently active
func (s *ServiceService) GetAllActiveServices() ([]models.Service, error) {
	conn, err := s.connManager.GetDefaultConnection()
	if err != nil {
		return nil, err
	}

	var services []models.Service
	if err := conn.DB.Preload("Fields").Where("is_active = ?", true).Find(&services).Error; err != nil {
		return nil, err
	}

	return services, nil
}
