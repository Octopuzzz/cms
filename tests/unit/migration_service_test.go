package services_test

import (
	"context"
	"testing"

	"cms-backend/internal/database"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationService_CRUD(t *testing.T) {
	db := setupTestDB(t)
	cm := database.NewTestConnectionManager(db)
	svcService := services.NewServiceService(cm)
	migService := services.NewMigrationService(cm)
	ctx := context.Background()

	// 1. Create a service
	req := &services.CreateServiceRequest{
		Name:        "MigTest",
		Description: "Migration test",
		IsPublic:    true,
		Fields: []services.CreateFieldRequest{
			{Name: "first_name", Type: "string"},
		},
	}
	modelSvc, err := svcService.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// 2. Create Migration
	migReq := &services.MigrationRequest{
		ServiceID:   modelSvc.ID,
		Description: "Add last name",
		Operations: []services.MigrationOperation{
			{
				Type:       "add_column",
				Column:     "last_name",
				ColumnType: "string",
				Nullable:   true,
			},
		},
	}
	mig, err := migService.CreateMigration(ctx, migReq, 1)
	require.NoError(t, err)
	assert.Equal(t, "applied", mig.Status)

	// Verify field is created
	updatedSvc, _ := svcService.GetService(ctx, modelSvc.ID)
	found := false
	for _, f := range updatedSvc.Fields {
		if f.Name == "last_name" {
			found = true
			break
		}
	}
	assert.True(t, found, "last_name field should be added")

	// 3. Rollback
	err = migService.RollbackMigration(ctx, mig.ID)
	require.NoError(t, err)
}
