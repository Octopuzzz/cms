// Package middleware provides Gin HTTP middleware.
package middleware

import (
	"context"
	"strings"

	"cms-backend/internal/contextutils"
	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Auth is the JWT authentication middleware
func Auth(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := authSvc.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(string(contextutils.ContextUserID), claims.UserID)
		c.Set(string(contextutils.ContextUsername), claims.Username)
		c.Set(string(contextutils.ContextRoles), claims.Roles)
		c.Set(string(contextutils.ContextIsSuperAdmin), claims.IsSuperAdmin)
		c.Set(string(contextutils.ContextClaims), claims)

		ctx := context.WithValue(c.Request.Context(), contextutils.ContextUserID, claims.UserID)
		ctx = context.WithValue(ctx, contextutils.ContextUsername, claims.Username)
		ctx = context.WithValue(ctx, contextutils.ContextRoles, claims.Roles)
		ctx = context.WithValue(ctx, contextutils.ContextIsSuperAdmin, claims.IsSuperAdmin)
		ctx = context.WithValue(ctx, contextutils.ContextClaims, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalAuth tries to authenticate but doesn't block if no token
func OptionalAuth(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				if claims, err := authSvc.ValidateToken(parts[1]); err == nil {
					c.Set(string(contextutils.ContextUserID), claims.UserID)
					c.Set(string(contextutils.ContextRoles), claims.Roles)
					c.Set(string(contextutils.ContextIsSuperAdmin), claims.IsSuperAdmin)

					ctx := context.WithValue(c.Request.Context(), contextutils.ContextUserID, claims.UserID)
					ctx = context.WithValue(ctx, contextutils.ContextRoles, claims.Roles)
					ctx = context.WithValue(ctx, contextutils.ContextIsSuperAdmin, claims.IsSuperAdmin)
					c.Request = c.Request.WithContext(ctx)
				}
			}
		}
		c.Next()
	}
}

// RequireRole ensures the user has at least one of the specified roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperAdmin, _ := c.Get(string(contextutils.ContextIsSuperAdmin))
		if sa, ok := isSuperAdmin.(bool); ok && sa {
			c.Next()
			return
		}

		userRoles, _ := c.Get(string(contextutils.ContextRoles))
		roleList, ok := userRoles.([]string)
		if !ok {
			response.Forbidden(c, "")
			c.Abort()
			return
		}

		for _, required := range roles {
			for _, userRole := range roleList {
				if userRole == required {
					c.Next()
					return
				}
			}
		}

		response.Forbidden(c, "insufficient permissions")
		c.Abort()
	}
}

// SuperAdmin ensures only super admins can access
func SuperAdmin() gin.HandlerFunc {
	return RequireRole("super_admin")
}

// GetUserID extracts the user ID from the context
func GetUserID(c *gin.Context) uint {
	if v, ok := c.Get(string(contextutils.ContextUserID)); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetRoles extracts roles from the context
func GetRoles(c *gin.Context) []string {
	if v, ok := c.Get(string(contextutils.ContextRoles)); ok {
		if roles, ok := v.([]string); ok {
			return roles
		}
	}
	return nil
}

// IsSuperAdmin checks if current user is super admin
func IsSuperAdmin(c *gin.Context) bool {
	if v, ok := c.Get(string(contextutils.ContextIsSuperAdmin)); ok {
		if sa, ok := v.(bool); ok {
			return sa
		}
	}
	return false
}
