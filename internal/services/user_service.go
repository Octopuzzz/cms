package services

import (
	"context"
	"errors"
	"fmt"

	"cms-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService handles user management
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new UserService
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

type CreateUserRequest struct {
	Username     string `json:"username" binding:"required,min=3,max=50"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	RoleIDs      []uint `json:"role_ids"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
	IsActive  bool   `json:"is_active"`
	RoleIDs   []uint `json:"role_ids"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	var existing models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("username or email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		Password:     string(hash),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		IsActive:     true,
		IsSuperAdmin: req.IsSuperAdmin,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}

	if len(req.RoleIDs) > 0 {
		userRoles := make([]models.UserRole, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, models.UserRole{UserID: user.ID, RoleID: roleID})
		}
		if err := s.db.CreateInBatches(userRoles, 100).Error; err != nil {
			return nil, err
		}
	} else {
		var viewerRole models.Role
		if err := s.db.Where("name = ?", "viewer").First(&viewerRole).Error; err == nil {
			if err := s.db.Create(&models.UserRole{UserID: user.ID, RoleID: viewerRole.ID}).Error; err != nil {
				return nil, err
			}
		}
	}

	return s.GetUser(ctx, user.ID)
}

func (s *UserService) GetUser(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Roles.Permissions").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int, search string) ([]models.User, int64, error) {
	q := s.db.WithContext(ctx).Model(&models.User{}).Preload("Roles")
	if search != "" {
		q = q.Where("username LIKE ? OR email LIKE ? OR first_name LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	q.Count(&total)
	var users []models.User
	if err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, total, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) (*models.User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone
	user.Avatar = req.Avatar
	user.IsActive = req.IsActive

	if err := s.db.Save(user).Error; err != nil {
		return nil, err
	}

	// Update roles
	if len(req.RoleIDs) > 0 {
		if err := s.db.Where("user_id = ?", id).Delete(&models.UserRole{}).Error; err != nil {
			return nil, err
		}
		userRoles := make([]models.UserRole, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			userRoles = append(userRoles, models.UserRole{UserID: id, RoleID: roleID})
		}
		if err := s.db.CreateInBatches(userRoles, 100).Error; err != nil {
			return nil, err
		}
	}

	return s.GetUser(ctx, id)
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, req *ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.db.Model(&user).Update("password", string(hash)).Error
}

// ---- Role Service ----

// RoleService manages roles and permissions
type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

type CreateRoleRequest struct {
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

type UpdateRoleRequest = CreateRoleRequest

func (s *RoleService) CreateRole(ctx context.Context, req *CreateRoleRequest) (*models.Role, error) {
	var existing models.Role
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("role '%s' already exists", req.Name)
	}
	role := &models.Role{Name: req.Name, Description: req.Description, IsActive: true}
	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}
	if len(req.PermissionIDs) > 0 {
		rolePerms := make([]models.RolePermission, 0, len(req.PermissionIDs))
		for _, pid := range req.PermissionIDs {
			rolePerms = append(rolePerms, models.RolePermission{RoleID: role.ID, PermissionID: pid})
		}
		if err := s.db.CreateInBatches(rolePerms, 100).Error; err != nil {
			return nil, err
		}
	}
	return s.GetRole(ctx, role.ID)
}

func (s *RoleService) GetRole(ctx context.Context, id uint) (*models.Role, error) {
	var role models.Role
	if err := s.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error; err != nil {
		return nil, errors.New("role not found")
	}
	return &role, nil
}

func (s *RoleService) ListRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	if err := s.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, id uint, req *UpdateRoleRequest) (*models.Role, error) {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	role.Name = req.Name
	role.Description = req.Description
	if err := s.db.Save(role).Error; err != nil {
		return nil, err
	}

	if len(req.PermissionIDs) > 0 {
		if err := s.db.Where("role_id = ?", id).Delete(&models.RolePermission{}).Error; err != nil {
			return nil, err
		}
		rolePerms := make([]models.RolePermission, 0, len(req.PermissionIDs))
		for _, pid := range req.PermissionIDs {
			rolePerms = append(rolePerms, models.RolePermission{RoleID: id, PermissionID: pid})
		}
		if err := s.db.CreateInBatches(rolePerms, 100).Error; err != nil {
			return nil, err
		}
	}
	return s.GetRole(ctx, id)
}

func (s *RoleService) DeleteRole(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.Role{}, id).Error
}

func (s *RoleService) ListPermissions(ctx context.Context) ([]models.Permission, error) {
	var perms []models.Permission
	s.db.WithContext(ctx).Find(&perms)
	return perms, nil
}

// ---- Database Connection Service ----

// DBConnectionService manages DB connections from the API
type DBConnectionService struct {
	db          *gorm.DB
	connManager interface {
		CreateConnection(ctx context.Context, dbConn *models.DatabaseConnection, password string) (interface{}, error)
	}
}

type DBConnService struct {
	db *gorm.DB
}

func NewDBConnService(db *gorm.DB) *DBConnService {
	return &DBConnService{db: db}
}

type CreateDBConnectionRequest struct {
	Name            string                 `json:"name" binding:"required,min=2,max=100"`
	Type            string                 `json:"type" binding:"required,oneof=postgresql mysql sqlite sqlserver"`
	Host            string                 `json:"host"`
	Port            int                    `json:"port"`
	Username        string                 `json:"username"`
	Password        string                 `json:"password"`
	Database        string                 `json:"database" binding:"required"`
	Schema          string                 `json:"schema"`
	SSLMode         string                 `json:"ssl_mode"`
	CustomURL       string                 `json:"custom_url"`
	IsDefault       bool                   `json:"is_default"`
	MaxOpenConns    int                    `json:"max_open_conns"`
	MaxIdleConns    int                    `json:"max_idle_conns"`
	ConnMaxLifetime int                    `json:"conn_max_lifetime_seconds"`
	ConnMaxIdleTime int                    `json:"conn_max_idle_time_seconds"`
	Metadata        map[string]interface{} `json:"metadata"`
}

func (s *DBConnService) Create(ctx context.Context, req *CreateDBConnectionRequest, userID uint) (*models.DatabaseConnection, error) {
	var existing models.DatabaseConnection
	if err := s.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("connection '%s' already exists", req.Name)
	}

	meta := models.JSONMap{}
	for k, v := range req.Metadata {
		meta[k] = v
	}

	conn := &models.DatabaseConnection{
		Name:            req.Name,
		Type:            req.Type,
		Host:            req.Host,
		Port:            req.Port,
		Username:        req.Username,
		Password:        req.Password,
		Database:        req.Database,
		Schema:          req.Schema,
		SSLMode:         req.SSLMode,
		CustomURL:       req.CustomURL,
		IsDefault:       req.IsDefault,
		IsActive:        true,
		MaxOpenConns:    max(req.MaxOpenConns, 5),
		MaxIdleConns:    max(req.MaxIdleConns, 2),
		ConnMaxLifetime: max(req.ConnMaxLifetime, 60),
		ConnMaxIdleTime: max(req.ConnMaxIdleTime, 30),
		Metadata:        meta,
		BaseModel:       models.BaseModel{CreatedBy: userID},
	}

	if err := s.db.WithContext(ctx).Create(conn).Error; err != nil {
		return nil, err
	}
	conn.Password = ""
	return conn, nil
}

func (s *DBConnService) Get(ctx context.Context, id uint) (*models.DatabaseConnection, error) {
	var conn models.DatabaseConnection
	if err := s.db.WithContext(ctx).First(&conn, id).Error; err != nil {
		return nil, errors.New("connection not found")
	}
	conn.Password = ""
	return &conn, nil
}

func (s *DBConnService) List(ctx context.Context) ([]models.DatabaseConnection, error) {
	var conns []models.DatabaseConnection
	s.db.WithContext(ctx).Find(&conns)
	for i := range conns {
		conns[i].Password = ""
	}
	return conns, nil
}

func (s *DBConnService) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.DatabaseConnection{}, id).Error
}

// ---- Custom Validation Service ----

type ValidationService struct {
	db *gorm.DB
}

func NewValidationService(db *gorm.DB) *ValidationService {
	return &ValidationService{db: db}
}

type CreateValidationRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=regex sql range enum"`
	Rule        string `json:"rule" binding:"required"`
	Description string `json:"description"`
	Message     string `json:"message"`
	IsGlobal    bool   `json:"is_global"`
	ServiceID   *uint  `json:"service_id"`
}

func (s *ValidationService) Create(ctx context.Context, req *CreateValidationRequest, userID uint) (*models.CustomValidation, error) {
	v := &models.CustomValidation{
		Name: req.Name, Code: req.Code, Type: req.Type, Rule: req.Rule,
		Description: req.Description, Message: req.Message,
		IsGlobal: req.IsGlobal, ServiceID: req.ServiceID, IsActive: true,
		BaseModel: models.BaseModel{CreatedBy: userID},
	}
	if err := s.db.Create(v).Error; err != nil {
		return nil, err
	}
	return v, nil
}

func (s *ValidationService) List(ctx context.Context, serviceID *uint) ([]models.CustomValidation, error) {
	var validations []models.CustomValidation
	q := s.db.WithContext(ctx).Where("is_active = true AND (is_global = true")
	if serviceID != nil {
		q = q.Where("OR service_id = ?", *serviceID)
	}
	q = q.Where(")")
	q.Find(&validations)
	return validations, nil
}

// MenuService handles menu management
type MenuService struct {
	db *gorm.DB
}

func NewMenuService(db *gorm.DB) *MenuService {
	return &MenuService{db: db}
}

func (s *MenuService) GetMenuTree(ctx context.Context) ([]models.Menu, error) {
	var menus []models.Menu
	if err := s.db.WithContext(ctx).
		Preload("Children").
		Preload("Service").
		Where("parent_id IS NULL AND is_active = true").
		Order("sort_order ASC").
		Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}
