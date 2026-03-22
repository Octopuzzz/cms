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

// setupConnectionManager creates an initialized ConnectionManager for testing
func setupConnectionManager(t *testing.T) *database.ConnectionManager {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.LogLevel = "error"
	logger.InitLogger("error", "console")

	cfg.Database = config.DatabaseConfig{
		Type:            "sqlite",
		Database:        ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	cm := database.GetConnectionManager()
	err := cm.Initialize(cfg)
	require.NoError(t, err)

	return cm
}

func TestDynamicDataService_CRUD(t *testing.T) {
	cm := setupConnectionManager(t)
	// Important: We must not Close the singleton ConnectionManager because it breaks other tests!
	// We just rely on SQLite in-memory isolation, or we might need to reset tables if needed,
	// but setupConnectionManager overwrites singleton each time. Wait, no, GetConnectionManager returns singleton.
	// That means it's already initialized if another test ran it.

	// Let's get the DB and clear it out if needed, but in-memory DB per dialector might be new if ":memory:" is used?
	// Actually for shared singleton, we shouldn't re-initialize.
	// Let's manually get a connection and pass it to services if possible.

	svcService := services.NewServiceService(cm)
	dataSvc := services.NewDynamicDataService(cm, svcService)
	ctx := context.Background()

	// 1. Create a dynamic service
	req := &services.CreateServiceRequest{
		Name:        "Test Posts",
		Description: "A test service for posts",
		IsPublic:    true,
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true, IsNullable: false},
			{Name: "content", Label: "Content", Type: models.FieldTypeText, IsRequired: false, IsNullable: true},
		},
	}
	svc, err := svcService.CreateService(ctx, req, 1)
	require.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, "test_posts", svc.Slug)

	// 2. Create data
	data := map[string]interface{}{
		"title":   "Hello World",
		"content": "This is a test post.",
	}
	createdData, err := dataSvc.CreateData(ctx, svc.Slug, data, 1)
	require.NoError(t, err)
	assert.NotNil(t, createdData)
	assert.Equal(t, "Hello World", createdData["title"])
	assert.Equal(t, "This is a test post.", createdData["content"])

	// Convert createdData ID to uint
	var id uint
	switch v := createdData["id"].(type) {
	case float64:
		id = uint(v)
	case int64:
		id = uint(v)
	case uint:
		id = v
	case int:
		id = uint(v)
	default:
		t.Fatalf("Unexpected ID type: %T", v)
	}

	// 3. Get data
	fetchedData, err := dataSvc.GetData(ctx, svc.Slug, id, nil)
	require.NoError(t, err)
	assert.Equal(t, createdData["title"], fetchedData["title"])

	// 4. Update data
	updateData := map[string]interface{}{
		"title": "Updated Hello World",
	}
	updatedData, err := dataSvc.UpdateData(ctx, svc.Slug, id, updateData, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Hello World", updatedData["title"])
	assert.Equal(t, "This is a test post.", updatedData["content"]) // Should remain unchanged if we just updated title?
	// Note: the test will show if UpdateData is a partial update or replaces everything.
	// Looking at the implementation, it's a partial update: UPDATE table SET title = ? WHERE id = ?

	// 5. List data
	// Create another one
	_, err = dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{
		"title":   "Second Post",
		"content": "Another content",
	}, 1)
	require.NoError(t, err)

	listReq := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
	}
	rows, total, err := dataSvc.ListData(ctx, svc.Slug, listReq)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, rows, 2)

	// 6. Delete data
	err = dataSvc.DeleteData(ctx, svc.Slug, id)
	require.NoError(t, err)

	// 7. Verify deletion (GetData should fail or not return it)
	_, err = dataSvc.GetData(ctx, svc.Slug, id, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")

	// Verify ListData doesn't return deleted
	rowsAfterDelete, totalAfterDelete, err := dataSvc.ListData(ctx, svc.Slug, listReq)
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalAfterDelete)
	assert.Len(t, rowsAfterDelete, 1)
}

func TestDynamicDataService_Validation(t *testing.T) {
	cm := setupConnectionManager(t)
	svcService := services.NewServiceService(cm)
	dataSvc := services.NewDynamicDataService(cm, svcService)
	ctx := context.Background()

	minLen := float64(5)
	maxLen := float64(20)

	req := &services.CreateServiceRequest{
		Name:        "Test Users Validation",
		Description: "A test service for users",
		Fields: []services.CreateFieldRequest{
			{
				Name:       "email",
				Label:      "Email",
				Type:       models.FieldTypeString,
				IsRequired: true,
				Validations: []models.FieldValidationItem{
					{Type: "email", Message: "invalid email format"},
				},
			},
			{
				Name:       "username",
				Label:      "Username",
				Type:       models.FieldTypeString,
				IsRequired: false,
				Validations: []models.FieldValidationItem{
					{Type: "min", Value: minLen},
					{Type: "max", Value: maxLen},
				},
			},
		},
	}
	svc, err := svcService.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test required field missing
	_, err = dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Email' is required")

	// Test invalid email
	_, err = dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{
		"email": "invalid-email",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// Test valid email, invalid min length
	_, err = dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{
		"email":    "test@example.com",
		"username": "abc",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Username' must be at least 5 characters")

	// Test valid email, invalid max length
	_, err = dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{
		"email":    "test@example.com",
		"username": "this_username_is_way_too_long",
	}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field 'Username' must be at most 20 characters")

	// Test valid data
	validData, err := dataSvc.CreateData(ctx, svc.Slug, map[string]interface{}{
		"email":    "test@example.com",
		"username": "validname",
	}, 1)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", validData["email"])
	assert.Equal(t, "validname", validData["username"])
}

func TestDynamicDataService_ApplyJoins(t *testing.T) {
	cm := setupConnectionManager(t)
	svcService := services.NewServiceService(cm)
	dataSvc := services.NewDynamicDataService(cm, svcService)
	ctx := context.Background()

	// Create Authors service
	authorReq := &services.CreateServiceRequest{
		Name: "Authors",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	authorSvc, err := svcService.CreateService(ctx, authorReq, 1)
	require.NoError(t, err)

	authorData, err := dataSvc.CreateData(ctx, authorSvc.Slug, map[string]interface{}{
		"name": "Jane Doe",
	}, 1)
	require.NoError(t, err)

	var authorID float64
	switch v := authorData["id"].(type) {
	case float64:
		authorID = v
	case int64:
		authorID = float64(v)
	case uint:
		authorID = float64(v)
	case int:
		authorID = float64(v)
	}

	// Create Books service
	bookReq := &services.CreateServiceRequest{
		Name: "Books",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
			{
				Name:  "author_id",
				Label: "Author",
				Type:  models.FieldTypeRelation,
				RelationConfig: &models.RelationConfig{
					RelatedServiceID: authorSvc.ID,
					RelatedField:     "id",
					DisplayField:     "name",
					RelationType:     "belongs_to",
				},
			},
		},
	}
	bookSvc, err := svcService.CreateService(ctx, bookReq, 1)
	require.NoError(t, err)

	// Create a book linked to author
	bookData, err := dataSvc.CreateData(ctx, bookSvc.Slug, map[string]interface{}{
		"title":     "Go Programming",
		"author_id": authorID,
	}, 1)
	require.NoError(t, err)

	var bookID uint
	switch v := bookData["id"].(type) {
	case float64:
		bookID = uint(v)
	case int64:
		bookID = uint(v)
	case uint:
		bookID = v
	case int:
		bookID = uint(v)
	}

	// Fetch book with join
	fetchedBook, err := dataSvc.GetData(ctx, bookSvc.Slug, bookID, []string{"author_id"})
	require.NoError(t, err)

	assert.Equal(t, "Go Programming", fetchedBook["title"])

	// Join result should be populated
	authorJoinData, ok := fetchedBook["author_id_data"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Jane Doe", authorJoinData["name"])
}
