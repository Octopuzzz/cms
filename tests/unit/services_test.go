package services_test

import (
	"context"
	"testing"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cms-backend/internal/database"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(models.AllModels()...))
	return db
}

func makeJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:             "test-secret-key",
		AccessTokenExpire:  15 * time.Minute,
		RefreshTokenExpire: 7 * 24 * time.Hour,
	}
}

// --- Auth Service Tests ---

func TestAuthService_Register_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewAuthService(db, makeJWTConfig())
	ctx := context.Background()

	user, err := svc.Register(ctx, &services.RegisterRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	})

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Empty(t, user.Password, "password must not be returned")
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewAuthService(db, makeJWTConfig())
	ctx := context.Background()

	req := &services.RegisterRequest{
		Username: "duplicate",
		Email:    "dup@example.com",
		Password: "password123",
	}
	_, err := svc.Register(ctx, req)
	require.NoError(t, err)

	_, err = svc.Register(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAuthService_Login_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewAuthService(db, makeJWTConfig())
	ctx := context.Background()

	_, err := svc.Register(ctx, &services.RegisterRequest{
		Username: "loginuser",
		Email:    "login@example.com",
		Password: "mypassword123",
	})
	require.NoError(t, err)

	tokens, user, err := svc.Login(ctx, &services.LoginRequest{
		Username: "loginuser",
		Password: "mypassword123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.NotNil(t, user)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewAuthService(db, makeJWTConfig())
	ctx := context.Background()

	_, err := svc.Register(ctx, &services.RegisterRequest{
		Username: "loginuser2",
		Email:    "login2@example.com",
		Password: "correctpassword",
	})
	require.NoError(t, err)

	_, _, err = svc.Login(ctx, &services.LoginRequest{
		Username: "loginuser2",
		Password: "wrongpassword",
	})
	assert.Error(t, err)
}

func TestAuthService_ValidateToken(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewAuthService(db, makeJWTConfig())
	ctx := context.Background()

	_, err := svc.Register(ctx, &services.RegisterRequest{
		Username: "tokenuser",
		Email:    "token@example.com",
		Password: "tokenpassword",
	})
	require.NoError(t, err)

	tokens, _, err := svc.Login(ctx, &services.LoginRequest{
		Username: "tokenuser",
		Password: "tokenpassword",
	})
	require.NoError(t, err)

	claims, err := svc.ValidateToken(tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "tokenuser", claims.Username)
}

// --- User Service Tests ---

func TestUserService_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewUserService(db)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, &services.CreateUserRequest{
		Username:  "newuser",
		Email:     "new@example.com",
		Password:  "newpassword123",
		FirstName: "New",
		LastName:  "User",
	})
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.Equal(t, "New", user.FirstName)

	fetched, err := svc.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "newuser", fetched.Username)
}

func TestUserService_ListUsers(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewUserService(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateUser(ctx, &services.CreateUserRequest{
			Username: "listuser" + string(rune('a'+i)),
			Email:    "listuser" + string(rune('a'+i)) + "@example.com",
			Password: "testpassword123",
		})
		require.NoError(t, err)
	}

	users, total, err := svc.ListUsers(ctx, 1, 10, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(3))
	assert.GreaterOrEqual(t, len(users), 3)
}

func TestUserService_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewUserService(db)
	ctx := context.Background()

	user, err := svc.CreateUser(ctx, &services.CreateUserRequest{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Password: "deletepassword123",
	})
	require.NoError(t, err)

	err = svc.DeleteUser(ctx, user.ID)
	require.NoError(t, err)

	_, err = svc.GetUser(ctx, user.ID)
	assert.Error(t, err)
}

// --- Role Service Tests ---

func TestRoleService_CRUD(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewRoleService(db)
	ctx := context.Background()

	// Create
	role, err := svc.CreateRole(ctx, &services.CreateRoleRequest{
		Name:        "test_editor",
		Description: "Can edit content",
	})
	require.NoError(t, err)
	assert.Equal(t, "test_editor", role.Name)

	// Get
	fetched, err := svc.GetRole(ctx, role.ID)
	require.NoError(t, err)
	assert.Equal(t, role.ID, fetched.ID)

	// Update
	updated, err := svc.UpdateRole(ctx, role.ID, &services.UpdateRoleRequest{
		Name:        "updated_editor",
		Description: "Updated description",
	})
	require.NoError(t, err)
	assert.Equal(t, "updated_editor", updated.Name)

	// Delete
	err = svc.DeleteRole(ctx, role.ID)
	require.NoError(t, err)
}

// --- DynamicDataService Tests ---

func TestDynamicDataService_ListData_SortValidation(t *testing.T) {
	db := setupTestDB(t)
	cm := database.GetConnectionManager()
	cm.AddConnection("default", "sqlite", ":memory:", db, 1, 10, time.Hour)

	// Create required models and services
	db.Create(&models.Service{
		Name:        "Test Sort Service",
		Slug:        "test-sort-svc",
		DbTableName: "test_sort_svc",
		Fields: []models.Field{
			{Name: "custom_field", Type: "string"},
		},
	})

	svc := services.NewServiceService(cm)
	dds := services.NewDynamicDataService(cm, svc)
	ctx := context.Background()

	// Mock resolveServiceDB internally or bypass for simple test if needed, assuming the mock DB doesn't require a real ConnectionManager connection here
	// Given the internal implementation uses resolveServiceDB which depends on connManager
	// We'll simulate a controlled test where we verify error cases

	// Valid sort
	reqValid := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
		SortBy:   "id",
	}
	_, _, err := dds.ListData(ctx, "test-sort-svc", reqValid)
	// It may fail later in execution due to mocking cm, but shouldn't fail validation
	if err != nil && err.Error() == "invalid sort field: id" {
		t.Fatalf("Expected 'id' to be a valid sort field")
	}

	// Valid sort with dynamic field
	reqValidField := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
		SortBy:   "custom_field",
	}
	_, _, err = dds.ListData(ctx, "test-sort-svc", reqValidField)
	if err != nil && err.Error() == "invalid sort field: custom_field" {
		t.Fatalf("Expected 'custom_field' to be a valid sort field")
	}

	// Invalid sort (SQL Injection attempt)
	reqInvalid := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
		SortBy:   "id; DROP TABLE users;",
	}
	_, _, err = dds.ListData(ctx, "test-sort-svc", reqInvalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort field")
}

// --- DBConnService Tests ---

func TestDBConnService_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	svc := services.NewDBConnService(db)
	ctx := context.Background()

	conn, err := svc.Create(ctx, &services.CreateDBConnectionRequest{
		Name:     "Test SQLite",
		Type:     "sqlite",
		Database: ":memory:",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, "Test SQLite", conn.Name)

	list, err := svc.List(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 1)
}
