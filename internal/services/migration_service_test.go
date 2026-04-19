package services_test

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
)

func TestMigrationService(t *testing.T) {
	logger.InitLogger("error", "json")

	connMgr := database.GetConnectionManager()
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
	err := connMgr.Initialize(cfg)
	require.NoError(t, err)
	defer connMgr.Close()

	ctx := context.Background()
	svcSvc := services.NewServiceService(connMgr)
	migSvc := services.NewMigrationService(connMgr)

	req := &services.CreateServiceRequest{
		Name: "Migration Items",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test CreateMigration
	migReq := &services.MigrationRequest{
		ServiceID:   svc.ID,
		Description: "Add desc",
		Operations: []services.MigrationOperation{
			{Type: "add_column", Column: "desc", ColumnType: "TEXT", Nullable: true},
		},
	}
	migration, err := migSvc.CreateMigration(ctx, migReq, 1)
	require.NoError(t, err)
	assert.Equal(t, "applied", migration.Status)

	// Test ListMigrations
	list, err := migSvc.ListMigrations(ctx, svc.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Test RollbackMigration
	err = migSvc.RollbackMigration(ctx, migration.ID)
	require.NoError(t, err)
}
