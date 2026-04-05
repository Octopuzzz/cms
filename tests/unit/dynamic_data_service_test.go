package services_test

import (
	"context"
	"testing"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDynamicDataTestEnv(t *testing.T) (*services.DynamicDataService, *services.ServiceService, *database.ConnectionManager) {
	db := setupTestDB(t)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "sqlite",
			Database: ":memory:",
		},
	}

	// Pre-populate default connection for testing if needed
	db.Create(&models.DatabaseConnection{
		Name:      "default",
		Type:      "sqlite",
		Database:  ":memory:",
		IsDefault: true,
	})

	cm := database.GetConnectionManager()

	// Let's rely on the real DB but set it to our memory DB
	// Since database.GetConnectionManager() is a singleton in main, we should initialize it.
	err := cm.Initialize(cfg)
	require.NoError(t, err)

	ss := services.NewServiceService(cm)
	ds := services.NewDynamicDataService(cm, ss)

	return ds, ss, cm
}

func TestDynamicDataService_ListData_SortValidation(t *testing.T) {
	db := setupTestDB(t)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "sqlite",
			Database: ":memory:",
		},
	}

	cm := database.GetConnectionManager()
	// Overwrite the DB inside the default connection after initialization
	err := cm.Initialize(cfg)
	require.NoError(t, err)

	defConn, err := cm.GetDefaultConnection()
	require.NoError(t, err)
	// Replace the physical connection with our gorm mapped DB
	defConn.DB = db

	ss := services.NewServiceService(cm)
	ds := services.NewDynamicDataService(cm, ss)
	ctx := context.Background()

	// 1. Create a service
	svcReq := &services.CreateServiceRequest{
		Name:        "Test Sort Service",
		Description: "A service to test sort validation",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString},
			{Name: "age", Label: "Age", Type: models.FieldTypeInteger},
		},
	}
	svc, err := ss.CreateService(ctx, svcReq, 1)
	require.NoError(t, err)

	// 2. Insert test data
	_, err = ds.CreateData(ctx, svc.Slug, map[string]interface{}{"title": "A", "age": 20}, 1)
	require.NoError(t, err)
	_, err = ds.CreateData(ctx, svc.Slug, map[string]interface{}{"title": "B", "age": 30}, 1)
	require.NoError(t, err)

	// 3. Test valid sort fields
	reqValid := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "age",
		SortOrder: "ASC",
	}
	rows, _, err := ds.ListData(ctx, svc.Slug, reqValid)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, int64(20), rows[0]["age"]) // SQLite integer comes back as int64

	// 4. Test common field sort
	reqCommon := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "created_at",
		SortOrder: "DESC",
	}
	rowsCommon, _, err := ds.ListData(ctx, svc.Slug, reqCommon)
	require.NoError(t, err)
	require.Len(t, rowsCommon, 2)

	// 5. Test invalid sort field (SQL Injection attempt)
	reqInvalid := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "id; DROP TABLE users; --",
		SortOrder: "ASC",
	}
	rowsInvalid, _, err := ds.ListData(ctx, svc.Slug, reqInvalid)
	require.NoError(t, err)
	require.Len(t, rowsInvalid, 2)
	// It should default to 'id' sort order because of the validation
	// ID 1 should be first because it defaults to ASC for valid fields but 'id' sortOrder might default to DESC or ASC depending on req
	// The implementation sets default sortOrder to DESC, but we provided "ASC" in reqInvalid, so it's `ORDER BY id ASC`
	assert.Equal(t, int64(1), rowsInvalid[0]["id"])
}
