package services_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/handlers"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGraphQLTestEnv(t *testing.T) (*gin.Engine, *services.ServiceService, *services.DynamicDataService, *services.GraphQLService) {
	gin.SetMode(gin.TestMode)

	cfg := config.GetInstance()
	cfg.Database.Type = config.DatabaseType("sqlite")
	cfg.Database.Database = ":memory:"

	connManager := database.GetConnectionManager()
	err := connManager.Initialize(cfg)
	require.NoError(t, err)

	defaultConn, _ := connManager.GetDefaultConnection()
	db := defaultConn.DB
	require.NoError(t, db.AutoMigrate(models.AllModels()...))

	svcService := services.NewServiceService(connManager)
	dynamicDataSvc := services.NewDynamicDataService(connManager, svcService)
	graphqlSvc := services.NewGraphQLService(dynamicDataSvc, svcService)
	graphqlH := handlers.NewGraphQLHandler(graphqlSvc)

	router := gin.New()

	// Mock auth middleware putting user ID in context
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Next()
	})

	router.POST("/data/:slug/graphql", graphqlH.HandleGraphQL)

	return router, svcService, dynamicDataSvc, graphqlSvc
}

func TestGraphQLHandler_CreateAndList(t *testing.T) {
	router, svcService, _, _ := setupGraphQLTestEnv(t)
	ctx := context.Background()

	// Create a service. The mock database might not use ID=1 so we let the service builder fallback
	svcReq := services.CreateServiceRequest{
		Name:                 "Comments",
		Description:          "User comments",
		Fields: []services.CreateFieldRequest{
			{Name: "author", Label: "Author", Type: "string", IsRequired: true},
			{Name: "text", Label: "Text", Type: "text", IsRequired: true},
		},
	}
	createdSvc, err := svcService.CreateService(ctx, &svcReq, 1)
	require.NoError(t, err)

	// 1. Execute Create Mutation
	createMutation := `
		mutation {
			create(author: "Alice", text: "Hello World!") {
				id
				author
				text
			}
		}
	`
	reqBody := map[string]interface{}{"query": createMutation}
	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/data/"+createdSvc.Slug+"/graphql", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Expected OK, got %s", w.Body.String())
	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	require.Contains(t, resp, "data", "Expected response to contain 'data' field, instead got %v", resp)
	data := resp["data"].(map[string]interface{})
	require.Contains(t, data, "create", "Expected 'create' field in data, instead got %v", data)

	if data["create"] != nil {
		createData := data["create"].(map[string]interface{})
		assert.Equal(t, "Alice", createData["author"])
		assert.Equal(t, "Hello World!", createData["text"])
	}

	// 2. Execute List Query
	listQuery := `
		query {
			list {
				id
				author
				text
			}
		}
	`
	reqBody2 := map[string]interface{}{"query": listQuery}
	bodyBytes2, _ := json.Marshal(reqBody2)

	req2, _ := http.NewRequest(http.MethodPost, "/data/"+createdSvc.Slug+"/graphql", bytes.NewBuffer(bodyBytes2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code, "Expected OK, got %s", w2.Body.String())
	var resp2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &resp2)
	require.NoError(t, err)

	require.Contains(t, resp2, "data", "Expected response to contain 'data', got %s", w2.Body.String())
	data2, ok := resp2["data"].(map[string]interface{})
	require.True(t, ok, "data field should be a map")

	listData, ok := data2["list"].([]interface{})
	require.True(t, ok, "list field should be an array, got %v", data2["list"])

	if len(listData) > 0 {
		item := listData[0].(map[string]interface{})
		assert.Equal(t, "Alice", item["author"])
		assert.Equal(t, "Hello World!", item["text"])
	}
}

func TestGraphQLHandler_MissingSlug(t *testing.T) {
	router, _, _, _ := setupGraphQLTestEnv(t)
	// We map the route directly for this test to trigger missing slug logic
	// But in standard routing `/:slug` requires the param.
	router.POST("/graphql", func(c *gin.Context) {
		h := handlers.NewGraphQLHandler(nil)
		h.HandleGraphQL(c)
	})

	req, _ := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer([]byte(`{"query": "{ list { id } }"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
