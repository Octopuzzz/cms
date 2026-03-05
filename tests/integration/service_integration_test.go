package integration_test

import (
	"context"
	"testing"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/models"
	"cms-backend/internal/services"
	"cms-backend/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite tests services against a real in-memory SQLite DB
type IntegrationTestSuite struct {
	suite.Suite
	ctx       context.Context
	connMgr   *database.ConnectionManager
	svcSvc    *services.ServiceService
	dataSvc   *services.DynamicDataService
	authSvc   *services.AuthService
	migSvc    *services.MigrationService
	backupSvc *services.BackupService
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()
	logger.InitLogger("error", "json") // suppress logs in tests

	// Reset singleton for tests
	s.connMgr = database.GetConnectionManager()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:            config.SQLite,
			Database:        ":memory:",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
		},
		Redis: config.RedisConfig{Enabled: false},
	}
	err := s.connMgr.Initialize(cfg)
	s.Require().NoError(err)

	s.svcSvc = services.NewServiceService(s.connMgr)
	s.dataSvc = services.NewDynamicDataService(s.connMgr, s.svcSvc)
	s.migSvc = services.NewMigrationService(s.connMgr)
	s.backupSvc = services.NewBackupService(s.connMgr)

	conn, err := s.connMgr.GetDefaultConnection()
	s.Require().NoError(err)
	s.authSvc = services.NewAuthService(conn.DB, &config.JWTConfig{
		Secret:             "integration-test-secret",
		AccessTokenExpire:  15 * time.Minute,
		RefreshTokenExpire: 7 * 24 * time.Hour,
	})
}

// TestCreateService verifies service creation, slug generation, and table creation
func (s *IntegrationTestSuite) TestCreateService() {
	req := &services.CreateServiceRequest{
		Name:        "Integration Test Products",
		Description: "Product catalog for testing",
		Fields: []services.CreateFieldRequest{
			{
				Name:       "name",
				Label:      "Product Name",
				Type:       models.FieldTypeString,
				IsRequired: true,
			},
			{
				Name:  "price",
				Label: "Price",
				Type:  models.FieldTypeFloat,
			},
		},
		MenuConfig: &services.MenuConfigRequest{
			Icon:      "package",
			SortOrder: 1,
			IsVisible: true,
		},
	}

	svc, err := s.svcSvc.CreateService(s.ctx, req, 1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), svc)

	assert.Equal(s.T(), "Integration Test Products", svc.Name)
	assert.NotEmpty(s.T(), svc.Slug)
	assert.NotEmpty(s.T(), svc.DbTableName)
	assert.Len(s.T(), svc.Fields, 2)
	assert.Equal(s.T(), "name", svc.Fields[0].Name)
	assert.Equal(s.T(), "price", svc.Fields[1].Name)
}

// TestServiceCRUD tests full service lifecycle
func (s *IntegrationTestSuite) TestServiceCRUD() {
	// Create
	svc, err := s.svcSvc.CreateService(s.ctx, &services.CreateServiceRequest{
		Name: "CRUD Test Service",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString},
		},
	}, 1)
	require.NoError(s.T(), err)
	id := svc.ID

	// Read
	fetched, err := s.svcSvc.GetService(s.ctx, id)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), id, fetched.ID)

	// List
	svcs, total, err := s.svcSvc.ListServices(s.ctx, 1, 10, nil)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), total, int64(1))
	assert.GreaterOrEqual(s.T(), len(svcs), 1)

	// Update
	updated, err := s.svcSvc.UpdateService(s.ctx, id, &services.UpdateServiceRequest{
		Name:        "Updated Service",
		Description: "Updated",
		IsActive:    true,
	}, 1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Service", updated.Name)

	// Delete
	err = s.svcSvc.DeleteService(s.ctx, id)
	require.NoError(s.T(), err)

	_, err = s.svcSvc.GetService(s.ctx, id)
	assert.Error(s.T(), err)
}

// TestAuthIntegration tests full login flow
func (s *IntegrationTestSuite) TestAuthIntegration() {
	conn, err := s.connMgr.GetDefaultConnection()
	require.NoError(s.T(), err)

	authSvc := services.NewAuthService(conn.DB, &config.JWTConfig{
		Secret:             "integration-test-secret",
		AccessTokenExpire:  15 * time.Minute,
		RefreshTokenExpire: 7 * 24 * time.Hour,
	})

	// Register
	user, err := authSvc.Register(s.ctx, &services.RegisterRequest{
		Username:  "integrationuser",
		Email:     "integration@example.com",
		Password:  "securepassword123",
		FirstName: "Integration",
	})
	require.NoError(s.T(), err)
	assert.NotZero(s.T(), user.ID)

	// Login
	tokens, loggedInUser, err := authSvc.Login(s.ctx, &services.LoginRequest{
		Username: "integrationuser",
		Password: "securepassword123",
	})
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), tokens.AccessToken)
	assert.NotNil(s.T(), loggedInUser)

	// Validate token
	claims, err := authSvc.ValidateToken(tokens.AccessToken)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "integrationuser", claims.Username)

	// Refresh tokens
	newTokens, err := authSvc.RefreshTokens(s.ctx, tokens.RefreshToken)
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), newTokens.AccessToken)
}

// TestMigrationService tests schema migration operations
func (s *IntegrationTestSuite) TestMigrationService() {
	// Create a service first
	svc, err := s.svcSvc.CreateService(s.ctx, &services.CreateServiceRequest{
		Name: "Migration Test Service",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString},
		},
	}, 1)
	require.NoError(s.T(), err)

	// Apply a migration to add a column
	migration, err := s.migSvc.CreateMigration(s.ctx, &services.MigrationRequest{
		ServiceID:   svc.ID,
		Description: "Add description column",
		Operations: []services.MigrationOperation{
			{Type: "add_column", Column: "description", ColumnType: "TEXT", Nullable: true},
		},
	}, 1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "applied", migration.Status)
	assert.NotNil(s.T(), migration.AppliedAt)

	// List migrations
	migrations, err := s.migSvc.ListMigrations(s.ctx, svc.ID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(migrations), 1)

	// Rollback
	err = s.migSvc.RollbackMigration(s.ctx, migration.ID)
	require.NoError(s.T(), err)
}

// TestBackupService tests backup and restore operations
func (s *IntegrationTestSuite) TestBackupService() {
	// Create a service first
	svc, err := s.svcSvc.CreateService(s.ctx, &services.CreateServiceRequest{
		Name: "Backup Test Service",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString},
		},
	}, 1)
	require.NoError(s.T(), err)

	// Create a backup
	backup, err := s.backupSvc.CreateBackup(s.ctx, svc.ID, 1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "completed", backup.Status)
	assert.Equal(s.T(), "snapshot", backup.Type)

	// List backups
	backups, err := s.backupSvc.ListBackups(s.ctx, svc.ID)
	require.NoError(s.T(), err)
	assert.GreaterOrEqual(s.T(), len(backups), 1)

	// Get backup
	fetched, err := s.backupSvc.GetBackup(s.ctx, backup.ID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), backup.ID, fetched.ID)

	// Restore backup
	err = s.backupSvc.RestoreBackup(s.ctx, backup.ID)
	require.NoError(s.T(), err)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
