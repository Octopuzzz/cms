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

func setupDynamicDataTest(t *testing.T) (*services.DynamicDataService, *services.ServiceService, *database.ConnectionManager) {
	// Need to initialize a ConnectionManager and a default DB
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     config.SQLite,
			Database: ":memory:",
		},
	}

	cm := database.GetConnectionManager()
	err := cm.Initialize(cfg)
	require.NoError(t, err)

	svcService := services.NewServiceService(cm)
	dynamicDataSvc := services.NewDynamicDataService(cm, svcService)

	return dynamicDataSvc, svcService, cm
}

func createTestService(t *testing.T, ctx context.Context, svcService *services.ServiceService, name string) *models.Service {
	req := &services.CreateServiceRequest{
		Name:        name,
		Description: "A test service for dynamic data",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
			{Name: "content", Label: "Content", Type: models.FieldTypeText},
			{Name: "price", Label: "Price", Type: models.FieldTypeFloat},
			{Name: "is_active", Label: "Active", Type: models.FieldTypeBoolean},
			{Name: "author_email", Label: "Email", Type: models.FieldTypeString, Validations: []models.FieldValidationItem{{Type: "email"}}},
			{Name: "short_code", Label: "Code", Type: models.FieldTypeString, Validations: []models.FieldValidationItem{{Type: "min", Value: 3}, {Type: "max", Value: 5}}},
		},
	}

	svc, err := svcService.CreateService(ctx, req, 1)
	require.NoError(t, err)
	return svc
}

func TestDynamicDataService_CreateAndGet(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()
	svc := createTestService(t, ctx, svcService, "Posts")

	// Create Data
	data := map[string]interface{}{
		"title":        "Hello World",
		"content":      "This is a test post.",
		"price":        9.99,
		"is_active":    1,
		"author_email": "test@example.com",
		"short_code":   "ABCD",
	}

	created, err := dds.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created["id"])
	assert.Equal(t, "Hello World", created["title"])

	idFloat, ok := created["id"].(float64)
	if !ok {
		idInt, ok := created["id"].(int64)
		if !ok {
			idInt2, ok := created["id"].(int)
			require.True(t, ok, "id should be an int, int64 or float64")
			idFloat = float64(idInt2)
		} else {
			idFloat = float64(idInt)
		}
	}
	id := uint(idFloat)

	// Get Data
	fetched, err := dds.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	assert.Equal(t, created["id"], fetched["id"])
	assert.Equal(t, "Hello World", fetched["title"])
}

func TestDynamicDataService_ListData(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()
	svc := createTestService(t, ctx, svcService, "Products")

	// Create 3 records
	for i := 0; i < 3; i++ {
		_, err := dds.CreateData(ctx, svc.Slug, map[string]interface{}{
			"title": "Product " + string(rune('A'+i)),
		}, 1)
		require.NoError(t, err)
	}

	// List
	req := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "id",
		SortOrder: "ASC",
	}

	rows, total, err := dds.ListData(ctx, svc.Slug, req)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, rows, 3)
	assert.Equal(t, "Product A", rows[0]["title"])
}

func TestDynamicDataService_UpdateData(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()
	svc := createTestService(t, ctx, svcService, "Articles")

	created, err := dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"title": "Old Title",
	}, 1)
	require.NoError(t, err)

	idFloat, ok := created["id"].(float64)
	if !ok {
		idInt, ok := created["id"].(int64)
		if !ok {
			idInt2, ok := created["id"].(int)
			require.True(t, ok, "id should be an int, int64 or float64")
			idFloat = float64(idInt2)
		} else {
			idFloat = float64(idInt)
		}
	}
	id := uint(idFloat)

	updated, err := dds.UpdateData(ctx, svc.Slug, id, map[string]interface{}{
		"title": "New Title",
	}, 2)
	require.NoError(t, err)
	assert.Equal(t, "New Title", updated["title"])

	// Verify DB update
	fetched, err := dds.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	assert.Equal(t, "New Title", fetched["title"])
}

func TestDynamicDataService_DeleteData(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()
	svc := createTestService(t, ctx, svcService, "Comments")

	created, err := dds.CreateData(ctx, svc.Slug, map[string]interface{}{
		"title": "Comment 1",
	}, 1)
	require.NoError(t, err)

	idFloat, ok := created["id"].(float64)
	if !ok {
		idInt, ok := created["id"].(int64)
		if !ok {
			idInt2, ok := created["id"].(int)
			require.True(t, ok, "id should be an int, int64 or float64")
			idFloat = float64(idInt2)
		} else {
			idFloat = float64(idInt)
		}
	}
	id := uint(idFloat)

	err = dds.DeleteData(ctx, svc.Slug, id)
	require.NoError(t, err)

	// Fetch should fail
	_, err = dds.GetData(ctx, svc.Slug, id, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestDynamicDataService_Validations(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()
	svc := createTestService(t, ctx, svcService, "Validations")

	tests := []struct {
		name        string
		data        map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "Missing required field",
			data: map[string]interface{}{
				"content": "No title provided",
			},
			expectError: true,
			errorMsg:    "field 'Title' is required",
		},
		{
			name: "Invalid email",
			data: map[string]interface{}{
				"title":        "Valid Title",
				"author_email": "not-an-email",
			},
			expectError: true,
			errorMsg:    "must be a valid email",
		},
		{
			name: "String too short",
			data: map[string]interface{}{
				"title":      "Valid",
				"short_code": "AB",
			},
			expectError: true,
			errorMsg:    "at least 3 characters",
		},
		{
			name: "String too long",
			data: map[string]interface{}{
				"title":      "Valid",
				"short_code": "ABCDEF",
			},
			expectError: true,
			errorMsg:    "at most 5 characters",
		},
		{
			name: "Valid data",
			data: map[string]interface{}{
				"title":        "Valid",
				"author_email": "test@example.com",
				"short_code":   "ABCD",
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dds.CreateData(ctx, svc.Slug, tc.data, 1)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDynamicDataService_ApplyJoins(t *testing.T) {
	dds, svcService, cm := setupDynamicDataTest(t)
	defer cm.Close()

	ctx := context.Background()

	// Create Author Service
	authorSvc, err := svcService.CreateService(ctx, &services.CreateServiceRequest{
		Name: "Authors",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString},
		},
	}, 1)
	require.NoError(t, err)

	// Create Book Service with relation to Author
	bookSvc, err := svcService.CreateService(ctx, &services.CreateServiceRequest{
		Name: "Books",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString},
			{
				Name:  "author_id",
				Label: "Author",
				Type:  models.FieldTypeRelation,
				RelationConfig: &models.RelationConfig{
					RelatedServiceID: authorSvc.ID,
					RelatedField:     "id",
				},
			},
		},
	}, 1)
	require.NoError(t, err)

	// Create Author
	authorData, err := dds.CreateData(ctx, authorSvc.Slug, map[string]interface{}{
		"name": "J.K. Rowling",
	}, 1)
	require.NoError(t, err)

	// Create Book
	bookData, err := dds.CreateData(ctx, bookSvc.Slug, map[string]interface{}{
		"title":     "Harry Potter",
		"author_id": authorData["id"],
	}, 1)
	require.NoError(t, err)

	idFloat, ok := bookData["id"].(float64)
	if !ok {
		idInt, ok := bookData["id"].(int64)
		if !ok {
			idInt2, ok := bookData["id"].(int)
			require.True(t, ok, "id should be an int, int64 or float64")
			idFloat = float64(idInt2)
		} else {
			idFloat = float64(idInt)
		}
	}
	bookID := uint(idFloat)

	// Get Book WITH join
	fetched, err := dds.GetData(ctx, bookSvc.Slug, bookID, []string{"author_id"})
	require.NoError(t, err)

	assert.Equal(t, "Harry Potter", fetched["title"])
	require.NotNil(t, fetched["author_id_data"])

	authorMap := fetched["author_id_data"].(map[string]interface{})
	assert.Equal(t, "J.K. Rowling", authorMap["name"])
}
