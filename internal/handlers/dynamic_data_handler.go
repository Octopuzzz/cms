package handlers

import (
	"strconv"
	"strings"

	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// DynamicDataHandler handles CRUD on dynamic service tables
type DynamicDataHandler struct {
	dataSvc    *services.DynamicDataService
	svcService *services.ServiceService
}

func NewDynamicDataHandler(dataSvc *services.DynamicDataService, svcService *services.ServiceService) *DynamicDataHandler {
	return &DynamicDataHandler{dataSvc: dataSvc, svcService: svcService}
}

// checkServiceAccess ensures the role has the required permission on the service
func (h *DynamicDataHandler) checkServiceAccess(c *gin.Context, slug, action string) bool {
	if middleware.IsSuperAdmin(c) {
		return true
	}

	ctx := c.Request.Context()
	svc, err := h.svcService.GetServiceBySlug(ctx, slug)
	if err != nil {
		response.NotFound(c, "service")
		return false
	}

	// Public services allow reads without auth
	if svc.IsPublic && action == "read" {
		return true
	}

	roles := middleware.GetRoles(c)
	for _, perm := range svc.Permissions {
		if perm.Role == nil {
			continue
		}
		for _, role := range roles {
			if perm.Role.Name == role {
				switch action {
				case "create":
					return perm.CanCreate
				case "read":
					return perm.CanRead
				case "update":
					return perm.CanUpdate
				case "delete":
					return perm.CanDelete
				}
			}
		}
	}

	// If no permissions defined, default allow for admin roles
	for _, role := range roles {
		if role == "admin" || role == "super_admin" {
			return true
		}
	}

	return false
}

// ListData godoc
// @Summary      List records
// @Description  List records from a dynamic service table with optional filters, sorting, and relation joins
// @Tags         data
// @Security     BearerAuth
// @Produce      json
// @Param        slug      path  string false "Service slug"
// @Param        page      query int    false "Page" default(1)
// @Param        page_size query int    false "Page size" default(20)
// @Param        sort_by   query string false "Sort field"
// @Param        sort_order query string false "asc or desc"
// @Param        joins     query string false "Comma-separated relation fields to join"
// @Success      200  {object}  response.Response
// @Router       /data/{slug} [get]
func (h *DynamicDataHandler) ListData(c *gin.Context) {
	slug := c.Param("slug")
	if !h.checkServiceAccess(c, slug, "read") {
		response.Forbidden(c, "")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	var joins []string
	if j := c.Query("joins"); j != "" {
		joins = strings.Split(j, ",")
	}

	req := &services.ListDataRequest{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
		Search:    c.Query("search"),
		Joins:     joins,
	}

	rows, total, err := h.dataSvc.ListData(c.Request.Context(), slug, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Paginated(c, rows, total, page, pageSize)
}

// CreateData godoc
// @Summary      Create record
// @Description  Create a new record in the service's dynamic table
// @Tags         data
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        slug path string true "Service slug"
// @Param        body body object true "Record data"
// @Success      201  {object}  response.Response
// @Router       /data/{slug} [post]
func (h *DynamicDataHandler) CreateData(c *gin.Context) {
	slug := c.Param("slug")
	if !h.checkServiceAccess(c, slug, "create") {
		response.Forbidden(c, "")
		return
	}

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	record, err := h.dataSvc.CreateData(c.Request.Context(), slug, data, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, record)
}

// GetDataByID godoc
// @Summary      Get record
// @Tags         data
// @Security     BearerAuth
// @Produce      json
// @Param        slug  path string true "Service slug"
// @Param        id    path int    true "Record ID"
// @Param        joins query string false "Comma-separated relation fields to join"
// @Success      200  {object}  response.Response
// @Router       /data/{slug}/{id} [get]
func (h *DynamicDataHandler) GetDataByID(c *gin.Context) {
	slug := c.Param("slug")
	if !h.checkServiceAccess(c, slug, "read") {
		response.Forbidden(c, "")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var joins []string
	if j := c.Query("joins"); j != "" {
		joins = strings.Split(j, ",")
	}

	record, err := h.dataSvc.GetData(c.Request.Context(), slug, uint(id), joins)
	if err != nil {
		response.NotFound(c, "record")
		return
	}
	response.OK(c, record)
}

// UpdateData godoc
// @Summary      Update record
// @Tags         data
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        slug path string true "Service slug"
// @Param        id   path int    true "Record ID"
// @Success      200  {object}  response.Response
// @Router       /data/{slug}/{id} [put]
func (h *DynamicDataHandler) UpdateData(c *gin.Context) {
	slug := c.Param("slug")
	if !h.checkServiceAccess(c, slug, "update") {
		response.Forbidden(c, "")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	var data map[string]any
	if err := c.ShouldBindJSON(&data); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	record, err := h.dataSvc.UpdateData(c.Request.Context(), slug, uint(id), data, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, record)
}

// DeleteData godoc
// @Summary      Delete record
// @Tags         data
// @Security     BearerAuth
// @Param        slug path string true "Service slug"
// @Param        id   path int    true "Record ID"
// @Success      200  {object}  response.Response
// @Router       /data/{slug}/{id} [delete]
func (h *DynamicDataHandler) DeleteData(c *gin.Context) {
	slug := c.Param("slug")
	if !h.checkServiceAccess(c, slug, "delete") {
		response.Forbidden(c, "")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := h.dataSvc.DeleteData(c.Request.Context(), slug, uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKMessage(c, "record deleted", nil)
}
