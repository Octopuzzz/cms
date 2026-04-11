package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDynamicDataService(t *testing.T) (*services.DynamicDataService, *services.ServiceService, *gorm.DB) {
	t.Helper()

	cm := database.GetConnectionManager()

	// Create a dummy config to init cm if not already done
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:            config.SQLite,
			Database:        ":memory:",
			MaxOpenConns:    1,
			MaxIdleConns:    1,
			ConnMaxLifetime: 1 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
		},
		Redis: config.RedisConfig{Enabled: false},
	}
	_ = cm.Initialize(cfg)

	defConn, err := cm.GetDefaultConnection()
	require.NoError(t, err)

	db := defConn.DB
	require.NoError(t, db.AutoMigrate(models.AllModels()...))

	ss := services.NewServiceService(cm)
	dds := services.NewDynamicDataService(cm, ss)

	return dds, ss, db
}

func TestDynamicDataService_CRUD(t *testing.T) {
	dds, ss, db := setupTestDynamicDataService(t)
	ctx := context.Background()

	// 1. Create a service first
	req := &services.CreateServiceRequest{
		Name: "Test Products " + fmt.Sprintf("%d", time.Now().UnixNano()),
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString, IsRequired: true},
			{Name: "price", Label: "Price", Type: models.FieldTypeFloat},
		},
	}
	svc, err := ss.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Verify table exists
	assert.True(t, db.Migrator().HasTable(svc.DbTableName))

	// 2. Create Data
	data := map[string]interface{}{
		"name":  "Test Product 1",
		"price": 19.99,
	}

	created, err := dds.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Retrieve id dynamically
	var id uint
	switch v := created["id"].(type) {
	case int64:
		id = uint(v)
	case float64:
		id = uint(v)
	default:
		var row map[string]interface{}
		err = db.Table(svc.DbTableName).Select("id").First(&row).Error
		require.NoError(t, err)
		if val, ok := row["id"].(int64); ok {
			id = uint(val)
		} else if val, ok := row["id"].(float64); ok {
			id = uint(val)
		} else {
			t.Fatalf("Could not determine ID type: %T", row["id"])
		}
	}
	require.NotZero(t, id)


	// 3. GetData
	fetched, err := dds.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "Test Product 1", fetched["name"])

	// 4. UpdateData
	updateData := map[string]interface{}{
		"name": "Updated Product",
	}
	updated, err := dds.UpdateData(ctx, svc.Slug, id, updateData, 2)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Product", updated["name"])

	switch v := updated["updated_by"].(type) {
	case int64:
		assert.Equal(t, int64(2), v)
	case float64:
		assert.Equal(t, float64(2), v)
	}

	// 5. DeleteData
	err = dds.DeleteData(ctx, svc.Slug, id)
	require.NoError(t, err)

	// Ensure deleted (soft delete)
	_, err = dds.GetData(ctx, svc.Slug, id, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestDynamicDataService_ListDataAndJoins(t *testing.T) {
	dds, ss, db := setupTestDynamicDataService(t)
	ctx := context.Background()

	// 1. Create a Category service (Related service)
	catReq := &services.CreateServiceRequest{
		Name: "Test Categories " + fmt.Sprintf("%d", time.Now().UnixNano()),
		Fields: []services.CreateFieldRequest{
			{Name: "cat_name", Label: "Category Name", Type: models.FieldTypeString},
		},
	}
	catSvc, err := ss.CreateService(ctx, catReq, 1)
	require.NoError(t, err)

	// Create a category record
	catData := map[string]interface{}{"cat_name": "Electronics"}
	catRecord, err := dds.CreateData(ctx, catSvc.Slug, catData, 1)
	require.NoError(t, err)
	var catID uint
	if v, ok := catRecord["id"].(int64); ok {
		catID = uint(v)
	} else if v, ok := catRecord["id"].(float64); ok {
		catID = uint(v)
	}

	// 2. Create an Item service with Relation
	itemReq := &services.CreateServiceRequest{
		Name: "Test Items " + fmt.Sprintf("%d", time.Now().UnixNano()),
		Fields: []services.CreateFieldRequest{
			{Name: "item_name", Label: "Item Name", Type: models.FieldTypeString},
			{
				Name:  "category_id",
				Label: "Category",
				Type:  models.FieldTypeRelation,
				RelationConfig: &models.RelationConfig{
					RelatedServiceID: catSvc.ID,
					RelatedField:     "id",
				},
			},
		},
	}
	itemSvc, err := ss.CreateService(ctx, itemReq, 1)
	require.NoError(t, err)

	// Ensure the RelationConfig is saved in DB for the field
	var field models.Field
	err = db.Where("service_id = ? AND name = ?", itemSvc.ID, "category_id").First(&field).Error
	require.NoError(t, err)
	if field.RelationConfig == nil {
		rc := models.RelationConfig{RelatedServiceID: catSvc.ID, RelatedField: "id"}
		field.RelationConfig = &rc
		db.Save(&field)
	}

	// Refresh service from DB
	itemSvc, err = ss.GetService(ctx, itemSvc.ID)
	require.NoError(t, err)

	// 3. Create items
	_, err = dds.CreateData(ctx, itemSvc.Slug, map[string]interface{}{"item_name": "Item A", "category_id": catID}, 1)
	require.NoError(t, err)
	_, err = dds.CreateData(ctx, itemSvc.Slug, map[string]interface{}{"item_name": "Item C", "category_id": catID}, 1)
	require.NoError(t, err)
	_, err = dds.CreateData(ctx, itemSvc.Slug, map[string]interface{}{"item_name": "Item B", "category_id": catID}, 1)
	require.NoError(t, err)

	// 4. Test ListData with pagination and sorting
	req := &services.ListDataRequest{
		Page:      1,
		PageSize:  2,
		SortBy:    "item_name",
		SortOrder: "ASC",
	}
	rows, total, err := dds.ListData(ctx, itemSvc.Slug, req)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, rows, 2)
	assert.Equal(t, "Item A", rows[0]["item_name"])
	assert.Equal(t, "Item B", rows[1]["item_name"])

	// Page 2
	req.Page = 2
	rows, total, err = dds.ListData(ctx, itemSvc.Slug, req)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, rows, 1)
	assert.Equal(t, "Item C", rows[0]["item_name"])

	// 5. Test ListData with Joins
	reqJoin := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "item_name",
		SortOrder: "ASC",
		Joins:     []string{"category_id"},
	}
	rowsJoin, totalJoin, err := dds.ListData(ctx, itemSvc.Slug, reqJoin)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalJoin)
	assert.Len(t, rowsJoin, 3)

	// Check if joined data is present
	for _, row := range rowsJoin {
		catDataRaw, exists := row["category_id_data"]
		assert.True(t, exists, "Join data should exist")

		catData, ok := catDataRaw.(map[string]interface{})
		assert.True(t, ok, "Join data should be a map")
		assert.Equal(t, "Electronics", catData["cat_name"])
	}

	// 6. Test GetData with Joins
	itemAIdRaw := rowsJoin[0]["id"]
	var itemAId uint
	if v, ok := itemAIdRaw.(int64); ok {
		itemAId = uint(v)
	} else if v, ok := itemAIdRaw.(float64); ok {
		itemAId = uint(v)
	}

	fetched, err := dds.GetData(ctx, itemSvc.Slug, itemAId, []string{"category_id"})
	require.NoError(t, err)
	assert.NotNil(t, fetched)

	fetchedCatDataRaw, exists := fetched["category_id_data"]
	assert.True(t, exists)
	fetchedCatData := fetchedCatDataRaw.(map[string]interface{})
	assert.Equal(t, "Electronics", fetchedCatData["cat_name"])
}

func TestDynamicDataService_Validations(t *testing.T) {
	dds, ss, _ := setupTestDynamicDataService(t)
	ctx := context.Background()

	// Create service with validations
	req := &services.CreateServiceRequest{
		Name: "Test Validations " + fmt.Sprintf("%d", time.Now().UnixNano()),
		Fields: []services.CreateFieldRequest{
			{Name: "req_field", Label: "Required Field", Type: models.FieldTypeString, IsRequired: true},
			{
				Name:  "email_field",
				Label: "Email Field",
				Type:  models.FieldTypeString,
				Validations: models.FieldValidations{
					{Type: "email", Message: "invalid email format"},
				},
			},
			{
				Name:  "min_field",
				Label: "Min Field",
				Type:  models.FieldTypeString,
				Validations: models.FieldValidations{
					{Type: "min", Value: 3.0}, // Value as float64 usually from JSON
				},
			},
			{
				Name:  "max_field",
				Label: "Max Field",
				Type:  models.FieldTypeString,
				Validations: models.FieldValidations{
					{Type: "max", Value: 5.0},
				},
			},
		},
	}
	svc, err := ss.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 1. Missing Required Field
	_, err = dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"email_field": "valid@email.com",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")

	// 2. Invalid Email
	_, err = dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"req_field":   "exists",
		"email_field": "invalidemail",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// 3. Invalid Min length
	_, err = dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"req_field": "exists",
		"min_field": "ab",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be at least")

	// 4. Invalid Max length
	_, err = dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"req_field": "exists",
		"max_field": "123456",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be at most")

	// 5. Valid Data
	_, err = dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"req_field":   "exists",
		"email_field": "valid@email.com",
		"min_field":   "123",
		"max_field":   "12345",
	}, 1)
	assert.NoError(t, err)
}
