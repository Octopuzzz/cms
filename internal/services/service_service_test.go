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

func TestServiceService(t *testing.T) {
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

	// CreateService is tested indirectly, test more
	req := &services.CreateServiceRequest{
		Name: "Svc Items",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test GetServiceBySlug
	fetchedSlug, err := svcSvc.GetServiceBySlug(ctx, svc.Slug)
	require.NoError(t, err)
	assert.Equal(t, svc.ID, fetchedSlug.ID)

	// Test ListServices
	filter := &services.ListServicesFilter{Search: "Svc Items"}
	list, total, err := svcSvc.ListServices(ctx, 1, 10, filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	// Test UpdateService
	idUint := new(uint)
	*idUint = svc.Fields[0].ID
	updateReq := &services.UpdateServiceRequest{
		Name: "Updated Svc Items",
		Fields: []services.UpdateFieldRequest{
			{ID: idUint, Name: "title", Label: "New Title", Type: models.FieldTypeString},
		},
	}
	updated, err := svcSvc.UpdateService(ctx, svc.ID, updateReq, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Svc Items", updated.Name)

	// Test SetPermissions
	perms := []models.ServicePermission{
		{RoleID: 1, CanRead: true, CanCreate: true},
	}
	err = svcSvc.SetPermissions(ctx, svc.ID, perms)
	require.NoError(t, err)

	// Test DeleteService
	err = svcSvc.DeleteService(ctx, svc.ID)
	require.NoError(t, err)

	_, err = svcSvc.GetService(ctx, svc.ID)
	assert.Error(t, err)
}
