package handlers

import (
	"strconv"

	"cms-backend/internal/middleware"
	"cms-backend/internal/models"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// ServiceHandler handles service management endpoints
type ServiceHandler struct {
	svcService *services.ServiceService
}

func NewServiceHandler(svcService *services.ServiceService) *ServiceHandler {
	return &ServiceHandler{svcService: svcService}
}

// CreateService godoc
// @Summary      Create service
// @Description  Create a new CMS service with fields. Automatically creates a database table and menu entry.
// @Tags         services
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreateServiceRequest true "Service data"
// @Success      201  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req services.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	svc, err := h.svcService.CreateService(c.Request.Context(), &req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, svc)
}

// GetService godoc
// @Summary      Get service
// @Tags         services
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Service ID"
// @Success      200  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /services/{id} [get]
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	svc, err := h.svcService.GetService(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "service")
		return
	}
	response.OK(c, svc)
}

// ListServices godoc
// @Summary      List services
// @Tags         services
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Page number" default(1)
// @Param        page_size query int    false "Items per page" default(20)
// @Param        search    query string false "Search term"
// @Success      200  {object}  response.Response
// @Router       /services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := &services.ListServicesFilter{Search: search}
	svcs, total, err := h.svcService.ListServices(c.Request.Context(), page, pageSize, filter)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Paginated(c, svcs, total, page, pageSize)
}

// UpdateService godoc
// @Summary      Update service
// @Tags         services
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Service ID"
// @Param        body body services.UpdateServiceRequest true "Update data"
// @Success      200  {object}  response.Response
// @Router       /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req services.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	svc, err := h.svcService.UpdateService(c.Request.Context(), uint(id), &req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, svc)
}

// DeleteService godoc
// @Summary      Delete service
// @Tags         services
// @Security     BearerAuth
// @Param        id path int true "Service ID"
// @Success      200  {object}  response.Response
// @Router       /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svcService.DeleteService(c.Request.Context(), uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKMessage(c, "service deleted", nil)
}

// SetPermissions godoc
// @Summary      Set service permissions
// @Description  Set role-based access permissions for a service
// @Tags         services
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "Service ID"
// @Success      200  {object}  response.Response
// @Router       /services/{id}/permissions [post]
func (h *ServiceHandler) SetPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var perms []models.ServicePermission
	if err := c.ShouldBindJSON(&perms); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svcService.SetPermissions(c.Request.Context(), uint(id), perms); err != nil {
		response.InternalError(c, err)
		return
	}
	response.OKMessage(c, "permissions updated", nil)
}
