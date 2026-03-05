// Package models contains all data models for the CMS backend.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ==================== BASE MODEL ====================

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	CreatedBy uint           `json:"created_by" gorm:"index"`
	UpdatedBy uint           `json:"updated_by"`
}

// ==================== DATABASE CONNECTION ====================

// DatabaseConnection represents a configured database connection
type DatabaseConnection struct {
	BaseModel
	Name            string    `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Type            string    `json:"type" gorm:"size:50;not null"`
	Host            string    `json:"host" gorm:"size:255"`
	Port            int       `json:"port"`
	Username        string    `json:"username" gorm:"size:100"`
	Password        string    `json:"-" gorm:"size:255"`
	Database        string    `json:"database" gorm:"size:100;not null"`
	Schema          string    `json:"schema" gorm:"size:100;default:'public'"`
	SSLMode         string    `json:"ssl_mode" gorm:"size:50;default:'disable'"`
	SSLCert         string    `json:"-" gorm:"type:text"`
	SSLKey          string    `json:"-" gorm:"type:text"`
	SSLRootCert     string    `json:"-" gorm:"type:text"`
	CustomURL       string    `json:"-" gorm:"type:text"`
	IsActive        bool      `json:"is_active" gorm:"default:true;index"`
	IsDefault       bool      `json:"is_default" gorm:"default:false"`
	MaxOpenConns    int       `json:"max_open_conns" gorm:"default:25"`
	MaxIdleConns    int       `json:"max_idle_conns" gorm:"default:10"`
	ConnMaxLifetime int       `json:"conn_max_lifetime_seconds" gorm:"default:300"`
	ConnMaxIdleTime int       `json:"conn_max_idle_time_seconds" gorm:"default:60"`
	Metadata        JSONMap   `json:"metadata,omitempty" gorm:"type:text"`
	Services        []Service `json:"services,omitempty" gorm:"foreignKey:DatabaseConnectionID"`
}

func (DatabaseConnection) TableName() string { return "database_connections" }

// ==================== USER & AUTH ====================

// User represents a system user
type User struct {
	BaseModel
	Username     string     `json:"username" gorm:"uniqueIndex;size:100;not null"`
	Email        string     `json:"email" gorm:"uniqueIndex;size:255;not null"`
	Password     string     `json:"-" gorm:"size:255;not null"`
	FirstName    string     `json:"first_name" gorm:"size:100"`
	LastName     string     `json:"last_name" gorm:"size:100"`
	Phone        string     `json:"phone" gorm:"size:20"`
	Avatar       string     `json:"avatar" gorm:"size:500"`
	IsActive     bool       `json:"is_active" gorm:"default:true;index"`
	IsSuperAdmin bool       `json:"is_super_admin" gorm:"default:false"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	Roles        []Role     `json:"roles,omitempty" gorm:"many2many:user_roles;"`
}

func (User) TableName() string { return "users" }

func (u *User) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.Username
}

// RefreshToken stores JWT refresh tokens
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	User      *User     `json:"-" gorm:"foreignKey:UserID"`
	Token     string    `json:"token" gorm:"uniqueIndex;size:500;not null"`
	ExpiresAt time.Time `json:"expires_at"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

// ==================== ROLE & PERMISSION ====================

type Role struct {
	BaseModel
	Name        string       `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Description string       `json:"description" gorm:"size:500"`
	IsActive    bool         `json:"is_active" gorm:"default:true"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions;"`
	Users       []User       `json:"users,omitempty" gorm:"many2many:user_roles;"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	BaseModel
	Name        string `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Code        string `json:"code" gorm:"uniqueIndex;size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	Resource    string `json:"resource" gorm:"size:100;not null;index"`
	Action      string `json:"action" gorm:"size:50;not null;index"`
	Roles       []Role `json:"roles,omitempty" gorm:"many2many:role_permissions;"`
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	UserID    uint      `json:"user_id" gorm:"primaryKey"`
	RoleID    uint      `json:"role_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	RoleID       uint      `json:"role_id" gorm:"primaryKey"`
	PermissionID uint      `json:"permission_id" gorm:"primaryKey"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (RolePermission) TableName() string { return "role_permissions" }

// ==================== SERVICE ====================

// Service represents a dynamic CMS service (like a collection/table)
type Service struct {
	BaseModel
	Name                 string              `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Slug                 string              `json:"slug" gorm:"uniqueIndex;size:100;not null;index"`
	Description          string              `json:"description" gorm:"size:1000"`
	DbTableName          string              `json:"table_name" gorm:"uniqueIndex;column:db_table_name;size:100;not null"`
	IsActive             bool                `json:"is_active" gorm:"default:true;index"`
	IsPublic             bool                `json:"is_public" gorm:"default:false"`
	DatabaseConnectionID *uint               `json:"database_connection_id" gorm:"index"`
	DatabaseConnection   *DatabaseConnection `json:"database_connection,omitempty" gorm:"foreignKey:DatabaseConnectionID"`
	Fields               []Field             `json:"fields,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE;"`
	Permissions          []ServicePermission `json:"permissions,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE;"`
	Menu                 *Menu               `json:"menu,omitempty" gorm:"foreignKey:ServiceID"`
}

func (Service) TableName() string { return "services" }

// ==================== MENU ====================

type Menu struct {
	BaseModel
	Name      string   `json:"name" gorm:"size:100;not null"`
	Path      string   `json:"path" gorm:"size:200;index"`
	Icon      string   `json:"icon" gorm:"size:100"`
	ParentID  *uint    `json:"parent_id" gorm:"index"`
	Parent    *Menu    `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children  []Menu   `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	ServiceID *uint    `json:"service_id" gorm:"uniqueIndex;index"`
	Service   *Service `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
	SortOrder int      `json:"sort_order" gorm:"default:0"`
	IsActive  bool     `json:"is_active" gorm:"default:true;index"`
	IsVisible bool     `json:"is_visible" gorm:"default:true"`
}

func (Menu) TableName() string { return "menus" }

// ==================== FIELD ====================

type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeText     FieldType = "text"
	FieldTypeInteger  FieldType = "integer"
	FieldTypeFloat    FieldType = "float"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeJSON     FieldType = "json"
	FieldTypeUUID     FieldType = "uuid"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeFile     FieldType = "file"
	FieldTypeImage    FieldType = "image"
	FieldTypeRelation FieldType = "relation"
	FieldTypePassword FieldType = "password"
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
	FieldTypePhone    FieldType = "phone"
)

// Field represents a field configuration for a service
type Field struct {
	BaseModel
	ServiceID      uint             `json:"service_id" gorm:"index;not null"`
	Name           string           `json:"name" gorm:"size:100;not null;index"`
	Label          string           `json:"label" gorm:"size:200;not null"`
	Type           FieldType        `json:"type" gorm:"size:50;not null"`
	Description    string           `json:"description" gorm:"size:500"`
	IsRequired     bool             `json:"is_required" gorm:"default:false"`
	IsUnique       bool             `json:"is_unique" gorm:"default:false"`
	IsIndex        bool             `json:"is_index" gorm:"default:false"`
	IsNullable     bool             `json:"is_nullable" gorm:"default:true"`
	DefaultValue   string           `json:"default_value" gorm:"size:500"`
	Placeholder    string           `json:"placeholder" gorm:"size:200"`
	HelpText       string           `json:"help_text" gorm:"size:500"`
	SortOrder      int              `json:"sort_order" gorm:"default:0"`
	Options        FieldOptions     `json:"options,omitempty" gorm:"type:text"`
	Validations    FieldValidations `json:"validations,omitempty" gorm:"type:text"`
	RelationConfig *RelationConfig  `json:"relation_config,omitempty" gorm:"type:text"`
}

func (Field) TableName() string { return "fields" }

type FieldOptions map[string]interface{}

func (fo FieldOptions) Value() (driver.Value, error) {
	if fo == nil {
		return nil, nil
	}
	b, err := json.Marshal(fo)
	return string(b), err
}

func (fo *FieldOptions) Scan(value interface{}) error {
	if value == nil {
		*fo = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for FieldOptions")
	}
	return json.Unmarshal(bytes, fo)
}

// FieldValidations is a list of validation rules for a field
type FieldValidations []FieldValidationItem

type FieldValidationItem struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value,omitempty"`
	Message    string      `json:"message,omitempty"`
	CustomRule *uint       `json:"custom_rule,omitempty"`
}

func (fv FieldValidations) Value() (driver.Value, error) {
	if fv == nil {
		return nil, nil
	}
	b, err := json.Marshal(fv)
	return string(b), err
}

func (fv *FieldValidations) Scan(value interface{}) error {
	if value == nil {
		*fv = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for FieldValidations")
	}
	return json.Unmarshal(bytes, fv)
}

// RelationConfig represents configuration for relation fields
type RelationConfig struct {
	RelatedServiceID uint   `json:"related_service_id"`
	RelatedField     string `json:"related_field"`
	DisplayField     string `json:"display_field"`
	RelationType     string `json:"relation_type"` // belongs_to, has_one, has_many, many_to_many
	JoinTable        string `json:"join_table,omitempty"`
	CascadeDelete    bool   `json:"cascade_delete"`
}

func (rc *RelationConfig) Value() (driver.Value, error) {
	if rc == nil {
		return nil, nil
	}
	b, err := json.Marshal(rc)
	return string(b), err
}

func (rc *RelationConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for RelationConfig")
	}
	return json.Unmarshal(bytes, rc)
}

// ==================== SERVICE PERMISSION ====================

type ServicePermission struct {
	BaseModel
	ServiceID       uint             `json:"service_id" gorm:"index;not null"`
	RoleID          uint             `json:"role_id" gorm:"index;not null"`
	Role            *Role            `json:"role,omitempty" gorm:"foreignKey:RoleID"`
	CanCreate       bool             `json:"can_create" gorm:"default:false"`
	CanRead         bool             `json:"can_read" gorm:"default:true"`
	CanUpdate       bool             `json:"can_update" gorm:"default:false"`
	CanDelete       bool             `json:"can_delete" gorm:"default:false"`
	CanExport       bool             `json:"can_export" gorm:"default:false"`
	FieldLevelPerms FieldPermissions `json:"field_level_permissions,omitempty" gorm:"type:text"`
	RowFilter       string           `json:"row_filter" gorm:"type:text"`
}

func (ServicePermission) TableName() string { return "service_permissions" }

type FieldPermissions map[string]FieldPermission

type FieldPermission struct {
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
}

func (fp FieldPermissions) Value() (driver.Value, error) {
	if fp == nil {
		return nil, nil
	}
	b, err := json.Marshal(fp)
	return string(b), err
}

func (fp *FieldPermissions) Scan(value interface{}) error {
	if value == nil {
		*fp = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for FieldPermissions")
	}
	return json.Unmarshal(bytes, fp)
}

// ==================== CUSTOM VALIDATION ====================

// CustomValidation represents a reusable validation rule
type CustomValidation struct {
	BaseModel
	Name        string `json:"name" gorm:"uniqueIndex;size:100;not null"`
	Code        string `json:"code" gorm:"uniqueIndex;size:100;not null"`
	Description string `json:"description" gorm:"size:500"`
	Type        string `json:"type" gorm:"size:50;not null"` // regex, sql, js, range, enum
	Rule        string `json:"rule" gorm:"type:text;not null"`
	Message     string `json:"message" gorm:"size:500"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	IsGlobal    bool   `json:"is_global" gorm:"default:true"` // available to all services
	ServiceID   *uint  `json:"service_id,omitempty" gorm:"index"`
}

func (CustomValidation) TableName() string { return "custom_validations" }

// ==================== MIGRATION ====================

// Migration represents a schema migration record
type Migration struct {
	BaseModel
	ServiceID      uint       `json:"service_id" gorm:"index;not null"`
	Description    string     `json:"description" gorm:"size:500;not null"`
	Status         string     `json:"status" gorm:"size:50;not null;default:'pending'"` // pending, applied, failed, rolled_back
	SchemaSnapshot string     `json:"schema_snapshot,omitempty" gorm:"type:text"`
	ErrorMessage   string     `json:"error_message,omitempty" gorm:"type:text"`
	AppliedAt      *time.Time `json:"applied_at"`
}

func (Migration) TableName() string { return "migrations" }

// ==================== BACKUP ====================

// Backup represents a service data/schema backup
type Backup struct {
	BaseModel
	ServiceID  uint   `json:"service_id" gorm:"index;not null"`
	Type       string `json:"type" gorm:"size:50;not null"` // snapshot, schema, data
	Status     string `json:"status" gorm:"size:50;not null;default:'pending'"`
	SchemaData string `json:"schema_data,omitempty" gorm:"type:text"`
	TableData  string `json:"-" gorm:"type:text"` // hidden from JSON by default (may be large)
	RowCount   int    `json:"row_count"`
}

func (Backup) TableName() string { return "backups" }

// ==================== AUDIT LOG ====================

type AuditLog struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID        *uint     `json:"user_id" gorm:"index"`
	Action        string    `json:"action" gorm:"size:50;not null;index"`
	Resource      string    `json:"resource" gorm:"size:100;not null;index"`
	ResourceID    string    `json:"resource_id" gorm:"size:50;index"`
	OldValues     JSONMap   `json:"old_values,omitempty" gorm:"type:text"`
	NewValues     JSONMap   `json:"new_values,omitempty" gorm:"type:text"`
	IP            string    `json:"ip" gorm:"size:50"`
	UserAgent     string    `json:"user_agent" gorm:"size:500"`
	CorrelationID string    `json:"correlation_id" gorm:"size:100;index"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// ==================== API TOKEN ====================

type APIToken struct {
	BaseModel
	Name       string      `json:"name" gorm:"size:100;not null"`
	Token      string      `json:"-" gorm:"uniqueIndex;size:500;not null"`
	TokenHash  string      `json:"-" gorm:"size:255"`
	UserID     uint        `json:"user_id" gorm:"index;not null"`
	User       *User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Scopes     StringArray `json:"scopes" gorm:"type:text"`
	ExpiresAt  *time.Time  `json:"expires_at"`
	LastUsedAt *time.Time  `json:"last_used_at"`
	IsActive   bool        `json:"is_active" gorm:"default:true;index"`
}

func (APIToken) TableName() string { return "api_tokens" }

// ==================== HELPER TYPES ====================

type JSONMap map[string]interface{}

func (jm JSONMap) Value() (driver.Value, error) {
	if jm == nil {
		return nil, nil
	}
	b, err := json.Marshal(jm)
	return string(b), err
}

func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}
	return json.Unmarshal(bytes, jm)
}

type StringArray []string

func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	b, err := json.Marshal(sa)
	return string(b), err
}

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("unsupported type for StringArray")
	}
	return json.Unmarshal(bytes, sa)
}

// AllModels returns all model types for auto-migration
func AllModels() []interface{} {
	return []interface{}{
		&DatabaseConnection{},
		&User{},
		&RefreshToken{},
		&Role{},
		&Permission{},
		&UserRole{},
		&RolePermission{},
		&Service{},
		&Menu{},
		&Field{},
		&ServicePermission{},
		&CustomValidation{},
		&Migration{},
		&Backup{},
		&AuditLog{},
		&APIToken{},
	}
}
