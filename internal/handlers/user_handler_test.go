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
	"cms-backend/internal/services"
	"cms-backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler(t *testing.T) {
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

	conn, err := connMgr.GetDefaultConnection()
	require.NoError(t, err)
	db := conn.DB

	userSvc := services.NewUserService(db)
	userH := handlers.NewUserHandler(userSvc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	router.POST("/users", userH.CreateUser)
	router.GET("/users", userH.ListUsers)
	router.GET("/users/:id", userH.GetUser)
	router.PUT("/users/:id", userH.UpdateUser)
	router.DELETE("/users/:id", userH.DeleteUser)

	// Create User
	req := &services.CreateUserRequest{
		Username: "newhandleruser",
		Email:    "newhandleruser@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	router.ServeHTTP(w, reqPost)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	idFloat := data["id"].(float64)
	id := int(idFloat)
	idStr := fmt.Sprintf("%d", id)

	// Get User
	w = httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/users/"+idStr, nil)
	router.ServeHTTP(w, reqGet)
	assert.Equal(t, http.StatusOK, w.Code)

	// List Users
	w = httptest.NewRecorder()
	reqList, _ := http.NewRequest("GET", "/users", nil)
	router.ServeHTTP(w, reqList)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update User
	updateReq := &services.UpdateUserRequest{
		FirstName: "UpdatedName",
	}
	bodyUpdate, _ := json.Marshal(updateReq)
	w = httptest.NewRecorder()
	reqPut, _ := http.NewRequest("PUT", "/users/"+idStr, bytes.NewBuffer(bodyUpdate))
	router.ServeHTTP(w, reqPut)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete User
	w = httptest.NewRecorder()
	reqDelete, _ := http.NewRequest("DELETE", "/users/"+idStr, nil)
	router.ServeHTTP(w, reqDelete)
	assert.Equal(t, http.StatusOK, w.Code)
}
