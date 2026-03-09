package handlers_test

import (
	"cms-backend/internal/database"
	"cms-backend/internal/handlers"
	"cms-backend/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGraphQLHandler(t *testing.T) {
	// Setup minimum dependencies
	gin.SetMode(gin.TestMode)
	connManager := database.GetConnectionManager()
	svcService := services.NewServiceService(connManager)
	handler := handlers.NewGraphQLHandler(svcService, connManager)

	r := gin.New()
	r.POST("/graphql", handler.Handler())
	r.GET("/playground", handler.PlaygroundHandler())

	t.Run("Playground is accessible", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/playground", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "GraphQL playground")
	})

	t.Run("GraphQL endpoint accepts POST", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/graphql", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// We expect 200 or 400 because database is not set up correctly in this isolated unit test
		// but at least it shouldn't 404
		assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, w.Code)
	})
}
