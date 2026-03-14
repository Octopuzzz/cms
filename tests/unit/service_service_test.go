package services_test

import (
	"context"
	"testing"

	"cms-backend/internal/database"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceService_CRUD(t *testing.T) {
	db := setupTestDB(t)

	// Since NewConnectionManager needs a config, we'll mock it roughly or use a test instance
	cm := database.NewTestConnectionManager(db) // Assuming we can mock or inject DB

	svc := services.NewServiceService(cm)
	ctx := context.Background()

	// Create Service
	req := &services.CreateServiceRequest{
		Name:        "Test Service",
		Description: "A test service",
		IsPublic:    true,
		Fields: []services.CreateFieldRequest{
			{Name: "title", Type: "string", IsRequired: true},
		},
	}
	modelSvc, err := svc.CreateService(ctx, req, 1)
	require.NoError(t, err)
	assert.NotNil(t, modelSvc)
	assert.Equal(t, "test_service", modelSvc.Slug)

	// Get Service
	fetched, err := svc.GetService(ctx, modelSvc.ID)
	require.NoError(t, err)
	assert.Equal(t, "test_service", fetched.Slug)
	assert.Len(t, fetched.Fields, 1)

	// Get Service By Slug
	fetchedSlug, err := svc.GetServiceBySlug(ctx, "test_service")
	require.NoError(t, err)
	assert.Equal(t, modelSvc.ID, fetchedSlug.ID)

	// List Services
	list, _, err := svc.ListServices(ctx, 1, 10, nil)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Update Service
	upReq := &services.UpdateServiceRequest{
		Name:        "Updated Service",
		Description: "Updated desc",
		IsPublic:    false,
		IsActive:    true,
		Fields: []services.UpdateFieldRequest{
			{ID: &fetched.Fields[0].ID, Name: "title_updated", Type: "string"},
		},
	}
	updated, err := svc.UpdateService(ctx, modelSvc.ID, upReq, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Service", updated.Name)

	// Delete Service
	err = svc.DeleteService(ctx, modelSvc.ID)
	require.NoError(t, err)

	_, err = svc.GetService(ctx, modelSvc.ID)
	assert.Error(t, err)
}
