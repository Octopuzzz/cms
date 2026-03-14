package services_test

import (
	"context"
	"testing"

	"cms-backend/internal/database"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicDataService_CRUD(t *testing.T) {
	db := setupTestDB(t)
	cm := database.NewTestConnectionManager(db)
	svcService := services.NewServiceService(cm)
	dataSvc := services.NewDynamicDataService(cm, svcService)
	ctx := context.Background()

	// 1. Create a service to hold data
	req := &services.CreateServiceRequest{
		Name:        "Posts",
		Description: "Blog posts",
		IsPublic:    true,
		Fields: []services.CreateFieldRequest{
			{Name: "title", Type: "string", IsRequired: true},
			{Name: "content", Type: "string", IsRequired: false},
			{Name: "views", Type: "integer", DefaultValue: "0"},
		},
	}
	modelSvc, err := svcService.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Create Data
	postData := map[string]interface{}{
		"title":   "My First Post",
		"content": "Hello world",
		"views":   100,
	}
	created, err := dataSvc.CreateData(ctx, modelSvc.Slug, postData, 1)
	require.NoError(t, err)
	assert.NotNil(t, created["id"])
	var postID uint
	switch v := created["id"].(type) {
	case float64:
		postID = uint(v)
	case int64:
		postID = uint(v)
	case int:
		postID = uint(v)
	default:
		t.Fatalf("unexpected type for id: %T", v)
	}
	assert.Equal(t, "My First Post", created["title"])

	// 3. Get Data
	fetched, err := dataSvc.GetData(ctx, modelSvc.Slug, postID, nil)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", fetched["content"])

	// 4. Update Data
	updateData := map[string]interface{}{
		"title": "Updated Post",
	}
	updated, err := dataSvc.UpdateData(ctx, modelSvc.Slug, postID, updateData, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Post", updated["title"])
	assert.Equal(t, "Hello world", updated["content"]) // should not overwrite if not provided?
	// Note: GORM dynamic updates map[string]interface{} will update specified fields.

	// 5. List Data
	listReq := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
	}
	listRes, total, err := dataSvc.ListData(ctx, modelSvc.Slug, listReq)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, listRes, 1)

	// 6. Delete Data
	err = dataSvc.DeleteData(ctx, modelSvc.Slug, postID)
	require.NoError(t, err)

	_, err = dataSvc.GetData(ctx, modelSvc.Slug, postID, nil)
	assert.Error(t, err) // Should be deleted
}
