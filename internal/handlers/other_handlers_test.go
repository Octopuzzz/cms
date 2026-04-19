package handlers_test

import (
	"bytes"
	"context"
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

func TestOtherHandlers(t *testing.T) {
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

	roleSvc := services.NewRoleService(db)
	roleH := handlers.NewRoleHandler(roleSvc)

	backupSvc := services.NewBackupService(connMgr)
	backupH := handlers.NewBackupHandler(backupSvc)

	migSvc := services.NewMigrationService(connMgr)
	migH := handlers.NewMigrationHandler(migSvc)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	// Roles
	router.POST("/roles", roleH.CreateRole)
	router.GET("/roles", roleH.ListRoles)
	router.GET("/roles/:id", roleH.GetRole)
	router.PUT("/roles/:id", roleH.UpdateRole)
	router.DELETE("/roles/:id", roleH.DeleteRole)
	router.GET("/permissions", roleH.ListPermissions)

	// Role Tests
	roleReq := &services.CreateRoleRequest{
		Name:        "HandlerRole",
		Description: "Role from handler",
	}
	body, _ := json.Marshal(roleReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/roles", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	idFloat := data["id"].(float64)
	idStr := fmt.Sprintf("%d", int(idFloat))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/roles", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/roles/"+idStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	roleUpdateReq := &services.UpdateRoleRequest{
		Name: "UpdatedHandlerRole",
	}
	body, _ = json.Marshal(roleUpdateReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/roles/"+idStr, bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/permissions", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/roles/"+idStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)


	// Service setup for Migration/Backup
	ctx := context.Background()
	svcSvc := services.NewServiceService(connMgr)
	svcReq := &services.CreateServiceRequest{
		Name: "OtherItems",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString, IsRequired: true},
		},
	}
	svc, err := svcSvc.CreateService(ctx, svcReq, 1)
	require.NoError(t, err)
	svcIDStr := fmt.Sprintf("%d", svc.ID)

	// Backup Tests
	router.POST("/backup/service/:service_id", backupH.CreateBackup)
	router.GET("/backup/service/:service_id", backupH.ListBackups)
	router.POST("/restore/:id", backupH.RestoreBackup)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/backup/service/"+svcIDStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp["data"].(map[string]interface{})
	backupIDFloat := data["id"].(float64)
	backupIDStr := fmt.Sprintf("%d", int(backupIDFloat))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/backup/service/"+svcIDStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/restore/"+backupIDStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Migration Tests
	router.POST("/migrations", migH.CreateMigration)
	router.GET("/migrations/service/:service_id", migH.ListMigrations)
	router.POST("/migrations/:id/rollback", migH.RollbackMigration)

	migReq := &services.MigrationRequest{
		ServiceID:   svc.ID,
		Description: "Add desc",
		Operations: []services.MigrationOperation{
			{Type: "add_column", Column: "desc", ColumnType: "TEXT", Nullable: true},
		},
	}
	body, _ = json.Marshal(migReq)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/migrations", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	json.Unmarshal(w.Body.Bytes(), &resp)
	data = resp["data"].(map[string]interface{})
	migIDFloat := data["id"].(float64)
	migIDStr := fmt.Sprintf("%d", int(migIDFloat))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/migrations/service/"+svcIDStr, nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/migrations/"+migIDStr+"/rollback", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

}
