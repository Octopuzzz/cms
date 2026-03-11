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

func TestGraphQLService_GenerateSchema(t *testing.T) {
	cfg := config.GetInstance()
	// Default config is mostly empty, so give it an sqlite database
	cfg.Database.Type = config.DatabaseType("sqlite")
	cfg.Database.Database = ":memory:"

	connManager := database.GetConnectionManager()
	err := connManager.Initialize(cfg)
	require.NoError(t, err)
	defer connManager.Close()

	// Ensure models exist
	defaultConn, _ := connManager.GetDefaultConnection()
	db := defaultConn.DB
	require.NoError(t, db.AutoMigrate(models.AllModels()...))

	svcService := services.NewServiceService(connManager)
	dynamicDataSvc := services.NewDynamicDataService(connManager, svcService)
	graphqlSvc := services.NewGraphQLService(dynamicDataSvc, svcService)

	ctx := context.Background()

	// Create a test service definition
	var defaultDbID uint = 1 // or nil depending on logic, let's leave it empty to use default db
	svcReq := services.CreateServiceRequest{
		Name:        "Posts",
		Description: "A test posts service",
		DatabaseConnectionID: &defaultDbID,
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: "string", IsRequired: true},
			{Name: "content", Label: "Content", Type: "text"},
			{Name: "published", Label: "Published", Type: "boolean"},
			{Name: "views", Label: "Views", Type: "integer"},
		},
	}

	createdSvc, err := svcService.CreateService(ctx, &svcReq, 1)
	require.NoError(t, err)
	require.NotNil(t, createdSvc)

	// Test schema generation
	schema, err := graphqlSvc.GenerateSchema(ctx, createdSvc.Slug)
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	// Inspect the schema to ensure expected fields exist
	queryType := schema.QueryType()
	assert.NotNil(t, queryType)

	// Queries
	getField := queryType.Fields()["get"]
	assert.NotNil(t, getField)
	listField := queryType.Fields()["list"]
	assert.NotNil(t, listField)

	// Mutations
	mutationType := schema.MutationType()
	assert.NotNil(t, mutationType)

	createField := mutationType.Fields()["create"]
	assert.NotNil(t, createField)
	updateField := mutationType.Fields()["update"]
	assert.NotNil(t, updateField)
	deleteField := mutationType.Fields()["delete"]
	assert.NotNil(t, deleteField)

	// Cleanup
	err = svcService.DeleteService(ctx, createdSvc.ID)
	assert.NoError(t, err)
}

func TestGraphQLService_GenerateSchema_NotFound(t *testing.T) {
	cfg := config.GetInstance()
	cfg.Database.Type = config.DatabaseType("sqlite")
	cfg.Database.Database = ":memory:"

	connManager := database.GetConnectionManager()
	connManager.Initialize(cfg)
	defer connManager.Close()
	defaultConn, _ := connManager.GetDefaultConnection()
	db := defaultConn.DB
	db.AutoMigrate(models.AllModels()...)

	svcService := services.NewServiceService(connManager)
	dynamicDataSvc := services.NewDynamicDataService(connManager, svcService)
	graphqlSvc := services.NewGraphQLService(dynamicDataSvc, svcService)

	ctx := context.Background()

	schema, err := graphqlSvc.GenerateSchema(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, schema)
}
