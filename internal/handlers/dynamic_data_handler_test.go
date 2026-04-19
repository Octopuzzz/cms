package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/handlers"
	"cms-backend/internal/models"
	"cms-backend/internal/services"
	"cms-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicDataHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger.InitLogger("error", "json")

	connMgr := database.GetConnectionManager()
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:            config.SQLite,
			Database:        ":memory:",
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 1 * time.Minute,
		},
		Redis: config.RedisConfig{Enabled: false},
	}
	err := connMgr.Initialize(cfg)
	require.NoError(t, err)

	svcSvc := services.NewServiceService(connMgr)
	dataSvc := services.NewDynamicDataService(connMgr, svcSvc)
	dataH := handlers.NewDynamicDataHandler(dataSvc, svcSvc)

	// Set up router with mocked user info middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("roles", []string{"super_admin"}) // Superadmin bypasses normal permission checks
		c.Next()
	})

	router.POST("/data/:slug", dataH.CreateData)
	router.GET("/data/:slug", dataH.ListData)
	router.GET("/data/:slug/:id", dataH.GetDataByID)
	router.PUT("/data/:slug/:id", dataH.UpdateData)
	router.DELETE("/data/:slug/:id", dataH.DeleteData)

	// Create a test service directly using the service layer
	ctx := context.Background()
	req := &services.CreateServiceRequest{
		Name: "API Items",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := svcSvc.CreateService(ctx, req, 1)
	require.NoError(t, err)

	// Test POST /data/:slug
	body, _ := json.Marshal(map[string]interface{}{"name": "API Item 1"})
	w := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("POST", "/data/"+svc.Slug, bytes.NewBuffer(body))
	router.ServeHTTP(w, reqPost)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	idFloat := data["id"].(float64)
	id := int(idFloat)

	// Test GET /data/:slug/:id
	w = httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/data/"+svc.Slug+"/1", nil)
	router.ServeHTTP(w, reqGet)
	assert.Equal(t, http.StatusOK, w.Code)

	_ = id

	// Test PUT /data/:slug/:id
	bodyPut, _ := json.Marshal(map[string]interface{}{"name": "Updated API Item"})
	w = httptest.NewRecorder()
	reqPut, _ := http.NewRequest("PUT", "/data/"+svc.Slug+"/1", bytes.NewBuffer(bodyPut))
	router.ServeHTTP(w, reqPut)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test DELETE /data/:slug/:id
	w = httptest.NewRecorder()
	reqDelete, _ := http.NewRequest("DELETE", "/data/"+svc.Slug+"/1", nil)
	router.ServeHTTP(w, reqDelete)
	assert.Equal(t, http.StatusOK, w.Code)
}
