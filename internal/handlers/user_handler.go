package handlers

import (
	"strconv"

	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user management
type UserHandler struct {
	userSvc *services.UserService
}

func NewUserHandler(userSvc *services.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// ListUsers godoc
// @Summary      List users
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int    false "Page" default(1)
// @Param        page_size query int    false "Page size" default(20)
// @Param        search    query string false "Search"
// @Success      200  {object}  response.Response
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	users, total, err := h.userSvc.ListUsers(c.Request.Context(), page, pageSize, c.Query("search"))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Paginated(c, users, total, page, pageSize)
}

// CreateUser godoc
// @Summary      Create user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreateUserRequest true "User data"
// @Success      201  {object}  response.Response
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.userSvc.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, user)
}

// GetUser godoc
// @Summary      Get user
// @Tags         users
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200  {object}  response.Response
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	user, err := h.userSvc.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "user")
		return
	}
	response.OK(c, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Param        id path int true "User ID"
// @Success      200  {object}  response.Response
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, err := h.userSvc.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, user)
}

// DeleteUser godoc
// @Summary      Delete user
// @Tags         users
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200  {object}  response.Response
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.userSvc.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKMessage(c, "user deleted", nil)
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Success      200  {object}  response.Response
// @Router       /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID := middleware.GetUserID(c)
	if err := h.userSvc.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKMessage(c, "password changed", nil)
}

// RoleHandler handles role management
type RoleHandler struct {
	roleSvc *services.RoleService
}

func NewRoleHandler(roleSvc *services.RoleService) *RoleHandler {
	return &RoleHandler{roleSvc: roleSvc}
}

// ListRoles godoc
// @Summary      List roles
// @Tags         roles
// @Security     BearerAuth
// @Success      200  {object}  response.Response
// @Router       /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.roleSvc.ListRoles(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, roles)
}

// CreateRole godoc
// @Summary      Create role
// @Tags         roles
// @Security     BearerAuth
// @Accept       json
// @Success      201  {object}  response.Response
// @Router       /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req services.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.roleSvc.CreateRole(c.Request.Context(), &req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, role)
}

// GetRole godoc
// @Summary      Get role
// @Tags         roles
// @Security     BearerAuth
// @Param        id path int true "Role ID"
// @Success      200  {object}  response.Response
// @Router       /roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	role, err := h.roleSvc.GetRole(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "role")
		return
	}
	response.OK(c, role)
}

// UpdateRole godoc
// @Summary      Update role
// @Tags         roles
// @Security     BearerAuth
// @Param        id path int true "Role ID"
// @Accept       json
// @Success      200  {object}  response.Response
// @Router       /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req services.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.roleSvc.UpdateRole(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, role)
}

// DeleteRole godoc
// @Summary      Delete role
// @Tags         roles
// @Security     BearerAuth
// @Param        id path int true "Role ID"
// @Success      200  {object}  response.Response
// @Router       /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.roleSvc.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OKMessage(c, "role deleted", nil)
}

// ListPermissions godoc
// @Summary      List all permissions
// @Tags         roles
// @Security     BearerAuth
// @Success      200  {object}  response.Response
// @Router       /permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.roleSvc.ListPermissions(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, perms)
}
