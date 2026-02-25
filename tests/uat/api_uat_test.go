package uat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cms-backend/internal/config"
	"cms-backend/internal/database"
	"cms-backend/internal/handlers"
	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a full test HTTP server
func setupTestServer(t *testing.T) *gin.Engine {
	t.Helper()
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
		JWT: config.JWTConfig{
			Secret:             "uat-test-secret",
			AccessTokenExpire:  15 * time.Minute,
			RefreshTokenExpire: 7 * 24 * time.Hour,
		},
	}

	err := connMgr.Initialize(cfg)
	require.NoError(t, err)

	conn, err := connMgr.GetDefaultConnection()
	require.NoError(t, err)
	db := conn.DB

	authSvc := services.NewAuthService(db, &cfg.JWT)
	svcService := services.NewServiceService(connMgr)
	dynamicDataSvc := services.NewDynamicDataService(connMgr, svcService)
	userSvc := services.NewUserService(db)
	roleSvc := services.NewRoleService(db)
	dbConnSvc := services.NewDBConnService(db)
	valSvc := services.NewValidationService(db)
	menuSvc := services.NewMenuService(db)

	router := gin.New()
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	authH := handlers.NewAuthHandler(authSvc)
	svcH := handlers.NewServiceHandler(svcService)
	dataH := handlers.NewDynamicDataHandler(dynamicDataSvc, svcService)
	userH := handlers.NewUserHandler(userSvc)
	roleH := handlers.NewRoleHandler(roleSvc)
	dbConnH := handlers.NewDBConnectionHandler(dbConnSvc, connMgr)
	menuH := handlers.NewMenuHandler(menuSvc)
	valH := handlers.NewValidationHandler(valSvc)

	auth := v1.Group("/auth")
	auth.POST("/login", authH.Login)
	auth.POST("/register", authH.Register)
	auth.POST("/refresh", authH.Refresh)

	protected := v1.Group("")
	protected.Use(middleware.Auth(authSvc))
	protected.GET("/auth/me", authH.Me)
	protected.GET("/services", svcH.ListServices)
	protected.POST("/services", svcH.CreateService)
	protected.GET("/services/:id", svcH.GetService)
	protected.PUT("/services/:id", svcH.UpdateService)
	protected.DELETE("/services/:id", svcH.DeleteService)
	protected.GET("/data/:slug", dataH.ListData)
	protected.POST("/data/:slug", dataH.CreateData)
	protected.GET("/data/:slug/:id", dataH.GetDataByID)
	protected.PUT("/data/:slug/:id", dataH.UpdateData)
	protected.DELETE("/data/:slug/:id", dataH.DeleteData)
	protected.GET("/users", userH.ListUsers)
	protected.POST("/users", userH.CreateUser)
	protected.GET("/roles", roleH.ListRoles)
	protected.GET("/menu", menuH.GetMenuTree)
	protected.GET("/validations", valH.ListValidations)
	protected.POST("/validations", valH.CreateValidation)
	_ = dbConnH

	return router
}

func loginAsAdmin(t *testing.T, router *gin.Engine) string {
	t.Helper()

	// First register an admin
	body, _ := json.Marshal(map[string]interface{}{
		"username": "admin",
		"email":    "admin@test.com",
		"password": "admin_password_123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Login
	body, _ = json.Marshal(map[string]interface{}{
		"username": "admin",
		"password": "admin_password_123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data, _ := resp["data"].(map[string]interface{})
	return fmt.Sprintf("%v", data["access_token"])
}

// TestUAT_AuthFlow tests the full authentication flow
func TestUAT_AuthFlow(t *testing.T) {
	router := setupTestServer(t)

	// Register
	body, _ := json.Marshal(map[string]interface{}{
		"username": "uatuser",
		"email":    "uat@example.com",
		"password": "uatpassword123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	body, _ = json.Marshal(map[string]interface{}{
		"username": "uatuser",
		"password": "uatpassword123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	data, _ := loginResp["data"].(map[string]interface{})
	token := fmt.Sprintf("%v", data["access_token"])
	assert.NotEmpty(t, token)

	// Me endpoint
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUAT_ServiceLifecycle tests service creation and menu auto-generation
func TestUAT_ServiceLifecycle(t *testing.T) {
	router := setupTestServer(t)
	token := loginAsAdmin(t, router)

	// Create service
	serviceData := map[string]interface{}{
		"name":        "UAT Products",
		"description": "Products for UAT testing",
		"fields": []interface{}{
			map[string]interface{}{
				"name": "name", "label": "Name", "type": "string", "is_required": true,
			},
			map[string]interface{}{
				"name": "price", "label": "Price", "type": "float",
			},
		},
		"menu_config": map[string]interface{}{
			"icon": "shopping-cart", "sort_order": 1, "is_visible": true,
		},
	}

	body, _ := json.Marshal(serviceData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v1/services", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	svcData, _ := createResp["data"].(map[string]interface{})
	slug, _ := svcData["slug"].(string)
	assert.NotEmpty(t, slug)

	// List services
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", "/api/v1/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Menu should have entry
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", "/api/v1/menu", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUAT_UnauthorizedAccess tests that protected routes reject unauthenticated requests
func TestUAT_UnauthorizedAccess(t *testing.T) {
	router := setupTestServer(t)

	endpoints := []struct {
		Method string
		Path   string
	}{
		{"GET", "/api/v1/services"},
		{"POST", "/api/v1/services"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/menu"},
		{"GET", "/api/v1/auth/me"},
	}

	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), ep.Method, ep.Path, nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"expected 401 for %s %s", ep.Method, ep.Path)
	}
}
