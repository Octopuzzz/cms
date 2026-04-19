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

func TestDynamicDataService_CRUD(t *testing.T) {
	// Initialize logger
	logger.InitLogger("error", "json")

	// Set up the connection manager
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
	dataSvc := services.NewDynamicDataService(connMgr, svcSvc)

	// Create a dynamic service
	req := &services.CreateServiceRequest{
		Name: "Test Items",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
			{Name: "price", Label: "Price", Type: models.FieldTypeFloat},
		},
	}
	svc, err := svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test CreateData
	data := map[string]interface{}{
		"title": "Item 1",
		"price": 9.99,
	}
	created, err := dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)
	assert.NotNil(t, created["id"])

	// Parse ID handling varied DB outputs
	var id uint
	switch v := created["id"].(type) {
	case int64:
		id = uint(v)
	case float64:
		id = uint(v)
	default:
		t.Fatalf("Unexpected ID type: %T", v)
	}

	// Test GetData
	fetched, err := dataSvc.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	assert.Equal(t, "Item 1", fetched["title"])

	// Test UpdateData
	updateData := map[string]interface{}{
		"title": "Item 1 Updated",
	}
	updated, err := dataSvc.UpdateData(ctx, svc.Slug, id, updateData, 1)
	require.NoError(t, err)
	assert.Equal(t, "Item 1 Updated", updated["title"])

	// Test ListData
	reqList := &services.ListDataRequest{Page: 1, PageSize: 10}
	list, total, err := dataSvc.ListData(ctx, svc.Slug, reqList)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	// Test DeleteData
	err = dataSvc.DeleteData(ctx, svc.Slug, id)
	require.NoError(t, err)

	_, err = dataSvc.GetData(ctx, svc.Slug, id, nil)
	assert.Error(t, err) // Should be not found after soft delete
}
