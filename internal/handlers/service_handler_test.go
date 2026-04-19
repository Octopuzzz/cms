package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
    "fmt"

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

func TestServiceHandler(t *testing.T) {
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
	svcH := handlers.NewServiceHandler(svcSvc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	router.POST("/services", svcH.CreateService)
	router.GET("/services", svcH.ListServices)
	router.GET("/services/:id", svcH.GetService)
	router.PUT("/services/:id", svcH.UpdateService)
	router.DELETE("/services/:id", svcH.DeleteService)
	router.POST("/services/:id/permissions", svcH.SetPermissions)

	// Create
	req := &services.CreateServiceRequest{
		Name: "Handler Items",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	router.ServeHTTP(w, reqPost)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	idFloat := data["id"].(float64)
	id := int(idFloat)
	idStr := fmt.Sprintf("%d", id)

	// Get
	w = httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/services/"+idStr, nil)
	router.ServeHTTP(w, reqGet)
	assert.Equal(t, http.StatusOK, w.Code)

	// List
	w = httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/services", nil)
	router.ServeHTTP(w, reqList)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	updateReq := &services.UpdateServiceRequest{
		Name: "Updated Handler Items",
	}
	bodyUpdate, _ := json.Marshal(updateReq)
	w = httptest.NewRecorder()
	reqPut, _ := http.NewRequest("PUT", "/services/"+idStr, bytes.NewBuffer(bodyUpdate))
	router.ServeHTTP(w, reqPut)
	assert.Equal(t, http.StatusOK, w.Code)

	// Permissions
	perms := []models.ServicePermission{
		{RoleID: 1, CanRead: true, CanCreate: true},
	}
	bodyPerms, _ := json.Marshal(perms)
	w = httptest.NewRecorder()
	reqPerms, _ := http.NewRequest("POST", "/services/"+idStr+"/permissions", bytes.NewBuffer(bodyPerms))
	router.ServeHTTP(w, reqPerms)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete
	w = httptest.NewRecorder()
	reqDelete, _ := http.NewRequest("DELETE", "/services/"+idStr, nil)
	router.ServeHTTP(w, reqDelete)
	assert.Equal(t, http.StatusOK, w.Code)
}
