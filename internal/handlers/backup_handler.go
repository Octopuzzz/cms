package handlers

import (
	"strconv"

	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// BackupHandler handles backup and restore endpoints
type BackupHandler struct {
	backupSvc *services.BackupService
}

// NewBackupHandler creates a new BackupHandler
func NewBackupHandler(backupSvc *services.BackupService) *BackupHandler {
	return &BackupHandler{backupSvc: backupSvc}
}

// CreateBackup godoc
// @Summary      Create a backup for a service
// @Description  Create a full snapshot backup (schema + data) of a service
// @Tags         backups
// @Security     BearerAuth
// @Param        service_id path int true "Service ID"
// @Success      201  {object}  response.Response
// @Router       /cms/backup/service/{service_id} [post]
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("service_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service_id")
		return
	}

	userID := uint(0)
	if v, ok := c.Get("user_id"); ok {
		if id, ok := v.(uint); ok {
			userID = id
		}
	}

	backup, createErr := h.backupSvc.CreateBackup(c.Request.Context(), uint(serviceID), userID)
	if createErr != nil {
		response.BadRequest(c, createErr.Error())
		return
	}
	response.Created(c, backup)
}

// ListBackups godoc
// @Summary      List backups for a service
// @Tags         backups
// @Security     BearerAuth
// @Param        service_id path int true "Service ID"
// @Success      200  {object}  response.Response
// @Router       /cms/backup/service/{service_id} [get]
func (h *BackupHandler) ListBackups(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("service_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid service_id")
		return
	}

	backups, listErr := h.backupSvc.ListBackups(c.Request.Context(), uint(serviceID))
	if listErr != nil {
		response.InternalError(c, listErr)
		return
	}
	response.OK(c, backups)
}

// RestoreBackup godoc
// @Summary      Restore from a backup
// @Description  Restore data from a backup into the service's table
// @Tags         backups
// @Security     BearerAuth
// @Param        id path int true "Backup ID"
// @Success      200  {object}  response.Response
// @Router       /cms/restore/{id} [post]
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if restoreErr := h.backupSvc.RestoreBackup(c.Request.Context(), uint(id)); restoreErr != nil {
		response.BadRequest(c, restoreErr.Error())
		return
	}
	response.OKMessage(c, "backup restored successfully", nil)
}
