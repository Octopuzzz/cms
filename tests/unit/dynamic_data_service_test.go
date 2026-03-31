package services_test

import (
	"context"
	"testing"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/require"
)

type DynamicDataTestEnv struct {
	db      *database.DBConnection
	connMgr *database.ConnectionManager
	svcSvc  *services.ServiceService
	dataSvc *services.DynamicDataService
}

func setupDynamicDataTest(t *testing.T) *DynamicDataTestEnv {
	t.Helper()

	// Initialize config
	cfg := config.GetInstance()
	cfg.Database.Type = config.SQLite
	cfg.Database.Database = ":memory:"

	// Initialize connection manager
	connMgr := database.GetConnectionManager()
	err := connMgr.Initialize(cfg)
	require.NoError(t, err)

	// Get default connection
	defaultConn, err := connMgr.GetDefaultConnection()
	require.NoError(t, err)

	// Initialize services
	svcSvc := services.NewServiceService(connMgr)
	dataSvc := services.NewDynamicDataService(connMgr, svcSvc)

	return &DynamicDataTestEnv{
		db:      defaultConn,
		connMgr: connMgr,
		svcSvc:  svcSvc,
		dataSvc: dataSvc,
	}
}

func TestDynamicDataService_CreateData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create a service first
	req := &services.CreateServiceRequest{
		Name:        "Products",
		Description: "A list of products",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
			{Name: "price", Label: "Price", Type: models.FieldTypeInteger},
		},
	}
	svc, err := env.svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Insert valid data
	data := map[string]interface{}{
		"title": "Smartphone",
		"price": 999,
	}

	// Assuming userID is 1 for testing
	result, err := env.dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check if title is set correctly
	require.Equal(t, "Smartphone", result["title"])
	// SQLite returns numeric types dynamically, usually as int64 for integers
	require.EqualValues(t, 999, result["price"])

	// 3. Insert invalid data (missing required field)
	invalidData := map[string]interface{}{
		"price": 100,
	}
	_, err = env.dataSvc.CreateData(ctx, svc.Slug, invalidData, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "field 'Title' is required")

	// 4. Test missing service
	_, err = env.dataSvc.CreateData(ctx, "non_existent", data, 1)
	require.Error(t, err)
}

func TestDynamicDataService_GetData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create a service
	req := &services.CreateServiceRequest{
		Name:        "Categories",
		Description: "Product categories",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := env.svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Insert data
	data := map[string]interface{}{
		"name": "Electronics",
	}
	result, err := env.dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)

	idVal, ok := result["id"]
	require.True(t, ok)
	var id uint
	switch v := idVal.(type) {
	case int64:
		id = uint(v)
	case uint:
		id = v
	case float64:
		id = uint(v)
	default:
		t.Fatalf("unexpected id type: %T", v)
	}

	// 3. Get existing data
	fetched, err := env.dataSvc.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	require.Equal(t, "Electronics", fetched["name"])

	// 4. Get non-existent data
	_, err = env.dataSvc.GetData(ctx, svc.Slug, 99999, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "record not found")
}

func TestDynamicDataService_UpdateData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create a service
	req := &services.CreateServiceRequest{
		Name:        "Tags",
		Description: "Tags",
		Fields: []services.CreateFieldRequest{
			{Name: "label", Label: "Label", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := env.svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Insert data
	data := map[string]interface{}{
		"label": "OldTag",
	}
	result, err := env.dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)

	var id uint
	switch v := result["id"].(type) {
	case int64: id = uint(v)
	case float64: id = uint(v)
	default: t.Fatalf("unexpected id type: %T", v)
	}

	// 3. Update data
	updateData := map[string]interface{}{
		"label": "NewTag",
	}
	updated, err := env.dataSvc.UpdateData(ctx, svc.Slug, id, updateData, 2) // assuming updated by user 2
	require.NoError(t, err)
	require.Equal(t, "NewTag", updated["label"])

	// 4. Verify update via GetData
	fetched, err := env.dataSvc.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	require.Equal(t, "NewTag", fetched["label"])
}

func TestDynamicDataService_ListData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create related service "Authors"
	authorReq := &services.CreateServiceRequest{
		Name: "Authors",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	authorSvc, err := env.svcSvc.CreateService(ctx, authorReq, 1)
	require.NoError(t, err)

	// 2. Insert authors
	authorResult, err := env.dataSvc.CreateData(ctx, authorSvc.Slug, map[string]interface{}{"name": "Jane Doe"}, 1)
	require.NoError(t, err)
	var authorID uint
	switch v := authorResult["id"].(type) {
	case int64: authorID = uint(v)
	case float64: authorID = uint(v)
	}

	// 3. Create service "Books" with relation to Authors
	bookReq := &services.CreateServiceRequest{
		Name: "Books",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
			{
				Name: "author_id",
				Label: "Author",
				Type: models.FieldTypeRelation,
				RelationConfig: &models.RelationConfig{
					RelatedServiceID: authorSvc.ID,
					RelatedField: "id",
					DisplayField: "name",
					RelationType: "belongs_to",
				},
			},
		},
	}
	bookSvc, err := env.svcSvc.CreateService(ctx, bookReq, 1)
	require.NoError(t, err)

	// 4. Insert books
	env.dataSvc.CreateData(ctx, bookSvc.Slug, map[string]interface{}{"title": "Book A", "author_id": authorID}, 1)
	env.dataSvc.CreateData(ctx, bookSvc.Slug, map[string]interface{}{"title": "Book B", "author_id": authorID}, 1)
	env.dataSvc.CreateData(ctx, bookSvc.Slug, map[string]interface{}{"title": "Book C"}, 1) // no author

	// 5. Test ListData
	req := &services.ListDataRequest{
		Page: 1,
		PageSize: 10,
	}
	items, total, err := env.dataSvc.ListData(ctx, bookSvc.Slug, req)
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, items, 3)

	// 6. Test Joins
	reqJoins := &services.ListDataRequest{
		Page: 1,
		PageSize: 10,
		Joins: []string{"author_id"},
	}
	joinedItems, _, err := env.dataSvc.ListData(ctx, bookSvc.Slug, reqJoins)
	require.NoError(t, err)
	require.Len(t, joinedItems, 3)

	// Verify join populated correctly for at least one item
	foundJoin := false
	for _, item := range joinedItems {
		if val, ok := item["author_id_data"]; ok && val != nil {
			authorData, ok := val.(map[string]interface{})
			if ok && authorData["name"] == "Jane Doe" {
				foundJoin = true
			}
		}
	}
	require.True(t, foundJoin, "Join data should be populated")
}

func TestDynamicDataService_DeleteData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create a service
	req := &services.CreateServiceRequest{
		Name:        "Comments",
		Fields: []services.CreateFieldRequest{
			{Name: "content", Label: "Content", Type: models.FieldTypeString},
		},
	}
	svc, err := env.svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Insert data
	data := map[string]interface{}{"content": "Nice post!"}
	result, err := env.dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)

	var id uint
	switch v := result["id"].(type) {
	case int64: id = uint(v)
	case float64: id = uint(v)
	}

	// 3. Delete data
	err = env.dataSvc.DeleteData(ctx, svc.Slug, id)
	require.NoError(t, err)

	// 4. Verify it's not found
	_, err = env.dataSvc.GetData(ctx, svc.Slug, id, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "record not found")

	// 5. Verify it's excluded from ListData
	reqList := &services.ListDataRequest{Page: 1, PageSize: 10}
	items, total, err := env.dataSvc.ListData(ctx, svc.Slug, reqList)
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Len(t, items, 0)
}

func TestDynamicDataService_ValidateData(t *testing.T) {
	env := setupDynamicDataTest(t)
	ctx := context.Background()

	// 1. Create a service with validation
	req := &services.CreateServiceRequest{
		Name: "UsersValid",
		Fields: []services.CreateFieldRequest{
			{
				Name: "email",
				Label: "Email",
				Type: models.FieldTypeString,
				Validations: models.FieldValidations{
					{Type: "email", Message: "Invalid email"},
				},
			},
			{
				Name: "username",
				Label: "Username",
				Type: models.FieldTypeString,
				Validations: models.FieldValidations{
					{Type: "min", Value: 3.0, Message: "Too short"},
					{Type: "max", Value: 10.0, Message: "Too long"},
				},
			},
			{
				Name: "age",
				Label: "Age",
				Type: models.FieldTypeInteger,
			},
		},
	}
	svc, err := env.svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Test valid data
	validData := map[string]interface{}{
		"email": "test@example.com",
		"username": "tester",
		"age": 25,
	}
	_, err = env.dataSvc.CreateData(ctx, svc.Slug, validData, 1)
	require.NoError(t, err)

	// 3. Test invalid email
	invalidEmail := map[string]interface{}{
		"email": "not-an-email",
		"username": "tester",
	}
	_, err = env.dataSvc.CreateData(ctx, svc.Slug, invalidEmail, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid email")

	// 4. Test min length
	invalidMin := map[string]interface{}{
		"email": "test@example.com",
		"username": "ab",
	}
	_, err = env.dataSvc.CreateData(ctx, svc.Slug, invalidMin, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be at least 3 characters")

	// 5. Test max length
	invalidMax := map[string]interface{}{
		"email": "test@example.com",
		"username": "thisiswaytoolongtobeausername",
	}
	_, err = env.dataSvc.CreateData(ctx, svc.Slug, invalidMax, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be at most 10 characters")
}
