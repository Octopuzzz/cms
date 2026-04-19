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
	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler(t *testing.T) {
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
			Secret:             "test-secret",
			AccessTokenExpire:  15 * time.Minute,
			RefreshTokenExpire: 7 * 24 * time.Hour,
		},
	}
	err := connMgr.Initialize(cfg)
	require.NoError(t, err)

	conn, err := connMgr.GetDefaultConnection()
	require.NoError(t, err)

	authSvc := services.NewAuthService(conn.DB, &cfg.JWT)
	authH := handlers.NewAuthHandler(authSvc)

	router := gin.New()
	router.POST("/auth/register", authH.Register)
	router.POST("/auth/login", authH.Login)
	router.POST("/auth/refresh", authH.Refresh)

	protected := router.Group("/auth")
	protected.Use(middleware.Auth(authSvc))
	protected.GET("/me", authH.Me)

	// Register
	body, _ := json.Marshal(map[string]interface{}{
		"username": "handleruser",
		"email":    "handler@test.com",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Login
	body, _ = json.Marshal(map[string]interface{}{
		"username": "handleruser",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	data := loginResp["data"].(map[string]interface{})
	token := data["access_token"].(string)
	refresh := data["refresh_token"].(string)

	// Refresh
	body, _ = json.Marshal(map[string]interface{}{
		"refresh_token": refresh,
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Me
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
