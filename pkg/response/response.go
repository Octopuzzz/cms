// Package response provides standardized API response helpers.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta holds pagination information
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// OK sends a 200 response
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

// OKMessage sends a 200 response with a message
func OKMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

// Created sends a 201 response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

// Paginated sends a paginated 200 response
func Paginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// BadRequest sends a 400 response
func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, Response{Success: false, Error: err})
}

// Unauthorized sends a 401 response
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "unauthorized"
	}
	c.JSON(http.StatusUnauthorized, Response{Success: false, Error: msg})
}

// Forbidden sends a 403 response
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "forbidden"
	}
	c.JSON(http.StatusForbidden, Response{Success: false, Error: msg})
}

// NotFound sends a 404 response
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, Response{Success: false, Error: resource + " not found"})
}

// Conflict sends a 409 response
func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, Response{Success: false, Error: err})
}

// UnprocessableEntity sends a 422 response with validation errors
func UnprocessableEntity(c *gin.Context, errs []ValidationError) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"errors":  errs,
	})
}

// InternalError sends a 500 response
func InternalError(c *gin.Context, err error) {
	msg := "internal server error"
	if err != nil {
		msg = err.Error()
	}
	c.JSON(http.StatusInternalServerError, Response{Success: false, Error: msg})
}
