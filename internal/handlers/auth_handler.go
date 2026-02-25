// Package handlers provides Gin HTTP handlers for authentication.
package handlers

import (
	"cms-backend/internal/middleware"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles auth endpoints
type AuthHandler struct {
	authSvc *services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authSvc *services.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Login godoc
// @Summary      Login
// @Description  Authenticate and get access tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body services.LoginRequest true "Login credentials"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, user, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, gin.H{
		"user":          user,
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
	})
}

// Register godoc
// @Summary      Register
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body services.RegisterRequest true "Registration data"
// @Success      201  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Conflict(c, err.Error())
		return
	}

	response.Created(c, user)
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Get a new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, err := h.authSvc.RefreshTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, tokens)
}

// Me godoc
// @Summary      Get current user
// @Description  Get the currently authenticated user
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authSvc.GetUserByID(userID)
	if err != nil {
		response.NotFound(c, "user")
		return
	}
	response.OK(c, user)
}
