package handlers

import (
	"strconv"

	"cms-backend/internal/database"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// DBConnectionHandler manages database connection endpoints
type DBConnectionHandler struct {
	dbSvc       *services.DBConnService
	connManager *database.ConnectionManager
}

func NewDBConnectionHandler(dbSvc *services.DBConnService, connManager *database.ConnectionManager) *DBConnectionHandler {
	return &DBConnectionHandler{dbSvc: dbSvc, connManager: connManager}
}

// ListConnections godoc
// @Summary      List DB connections
// @Tags         database-connections
// @Security     BearerAuth
// @Success      200  {object}  response.Response
// @Router       /database-connections [get]
func (h *DBConnectionHandler) ListConnections(c *gin.Context) {
	conns, err := h.dbSvc.List(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, conns)
}

// CreateConnection godoc
// @Summary      Create DB connection
// @Description  Register a new database connection (cloud or on-premise)
// @Tags         database-connections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreateDBConnectionRequest true "Connection data"
// @Success      201  {object}  response.Response
// @Router       /database-connections [post]
func (h *DBConnectionHandler) CreateConnection(c *gin.Context) {
	var req services.CreateDBConnectionRequest
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

	conn, err := h.dbSvc.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}
	response.Created(c, conn)
}

// GetConnection godoc
// @Summary      Get DB connection
// @Tags         database-connections
// @Security     BearerAuth
// @Param        id path int true "Connection ID"
// @Success      200  {object}  response.Response
// @Router       /database-connections/{id} [get]
func (h *DBConnectionHandler) GetConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	conn, connErr := h.dbSvc.Get(c.Request.Context(), uint(id))
	if connErr != nil {
		response.NotFound(c, "connection")
		return
	}
	response.OK(c, conn)
}

// DeleteConnection godoc
// @Summary      Delete DB connection
// @Tags         database-connections
// @Security     BearerAuth
// @Param        id path int true "Connection ID"
// @Success      200  {object}  response.Response
// @Router       /database-connections/{id} [delete]
func (h *DBConnectionHandler) DeleteConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if delErr := h.dbSvc.Delete(c.Request.Context(), uint(id)); delErr != nil {
		response.BadRequest(c, delErr.Error())
		return
	}
	response.OKMessage(c, "connection deleted", nil)
}

// TestConnection godoc
// @Summary      Test DB connection
// @Description  Test a database connection without saving it
// @Tags         database-connections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /database-connections/test [post]
func (h *DBConnectionHandler) TestConnection(c *gin.Context) {
	var req services.CreateDBConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if testErr := database.TestConnection(req.Type, req.Host, req.Port, req.Username, req.Password, req.Database, req.SSLMode, req.CustomURL); testErr != nil {
		response.BadRequest(c, "connection failed: "+testErr.Error())
		return
	}
	response.OK(c, gin.H{"status": "connected"})
}

// MenuHandler handles menu endpoints
type MenuHandler struct {
	menuSvc *services.MenuService
}

func NewMenuHandler(menuSvc *services.MenuService) *MenuHandler {
	return &MenuHandler{menuSvc: menuSvc}
}

// GetMenuTree godoc
// @Summary      Get menu tree
// @Description  Get the complete menu tree (automatically reflects created services)
// @Tags         menu
// @Security     BearerAuth
// @Success      200  {object}  response.Response
// @Router       /menu [get]
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	menu, err := h.menuSvc.GetMenuTree(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, menu)
}

// ValidationHandler handles custom validation endpoints
type ValidationHandler struct {
	valSvc *services.ValidationService
}

func NewValidationHandler(valSvc *services.ValidationService) *ValidationHandler {
	return &ValidationHandler{valSvc: valSvc}
}

// ListValidations godoc
// @Summary      List custom validations
// @Tags         validations
// @Security     BearerAuth
// @Success      200  {object}  response.Response
// @Router       /validations [get]
func (h *ValidationHandler) ListValidations(c *gin.Context) {
	vals, err := h.valSvc.List(c.Request.Context(), nil)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, vals)
}

// CreateValidation godoc
// @Summary      Create custom validation
// @Tags         validations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body services.CreateValidationRequest true "Validation data"
// @Success      201  {object}  response.Response
// @Router       /validations [post]
func (h *ValidationHandler) CreateValidation(c *gin.Context) {
	var req services.CreateValidationRequest
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
	val, err := h.valSvc.Create(c.Request.Context(), &req, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Created(c, val)
}
