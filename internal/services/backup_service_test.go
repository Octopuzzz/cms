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

func TestBackupService(t *testing.T) {
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
	backupSvc := services.NewBackupService(connMgr)

	req := &services.CreateServiceRequest{
		Name: "Backup Items",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test CreateBackup
	backup, err := backupSvc.CreateBackup(ctx, svc.ID, 1)
	require.NoError(t, err)
	assert.Equal(t, "completed", backup.Status)

	// Test ListBackups
	list, err := backupSvc.ListBackups(ctx, svc.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Test GetBackup
	fetched, err := backupSvc.GetBackup(ctx, backup.ID)
	require.NoError(t, err)
	assert.Equal(t, backup.ID, fetched.ID)

	// Test RestoreBackup
	err = backupSvc.RestoreBackup(ctx, backup.ID)
	require.NoError(t, err)
}
