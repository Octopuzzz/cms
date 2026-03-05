package handlers

import (
	"strconv"

	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// MigrationHandler handles schema migration endpoints
type MigrationHandler struct {
	migSvc *services.MigrationService
}

// NewMigrationHandler creates a new MigrationHandler
func NewMigrationHandler(migSvc *services.MigrationService) *MigrationHandler {
	return &MigrationHandler{migSvc: migSvc}
}

// CreateMigration godoc
// @Summary      Create and apply a schema migration
// @Description  Apply schema changes (add_column, drop_column, rename_column, change_type)
// @Tags         migrations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.MigrationRequest true "Migration operations"
// @Success      201  {object}  response.Response
// @Router       /cms/migrations [post]
func (h *MigrationHandler) CreateMigration(c *gin.Context) {
	var req services.MigrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := uint(0)
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}

	migration, err := h.migSvc.CreateMigration(c.Request.Context(), &req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, migration)
}

// ListMigrations godoc
// @Summary      List migrations for a service
// @Tags         migrations
// @Security     BearerAuth
// @Param        service_id path int true "Service ID"
// @Success      200  {object}  response.Response
// @Router       /cms/migrations/service/{service_id} [get]
func (h *MigrationHandler) ListMigrations(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("service_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service_id")
		return
	}

	migrations, listErr := h.migSvc.ListMigrations(c.Request.Context(), uint(serviceID))
	if listErr != nil {
		response.InternalError(c, listErr)
		return
	}
	response.OK(c, migrations)
}

// RollbackMigration godoc
// @Summary      Rollback a migration
// @Tags         migrations
// @Security     BearerAuth
// @Param        id path int true "Migration ID"
// @Success      200  {object}  response.Response
// @Router       /cms/migrations/{id}/rollback [post]
func (h *MigrationHandler) RollbackMigration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if rollbackErr := h.migSvc.RollbackMigration(c.Request.Context(), uint(id)); rollbackErr != nil {
		response.BadRequest(c, rollbackErr.Error())
		return
	}
	response.OKMessage(c, "migration rolled back", nil)
}
