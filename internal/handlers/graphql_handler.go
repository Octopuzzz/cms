package handlers

import (
	ourgraphql "cms-backend/internal/graphql"
	"cms-backend/internal/services"
	"cms-backend/internal/database"

	"github.com/gin-gonic/gin"
	gqlgenHandler "github.com/99designs/gqlgen/graphql/handler"
	gqlgenPlayground "github.com/99designs/gqlgen/graphql/playground"
)

import "sync"
import gqlgengraphql "github.com/99designs/gqlgen/graphql"

type GraphQLHandler struct {
	svcService    *services.ServiceService
	connManager   *database.ConnectionManager
	schema        gqlgengraphql.ExecutableSchema
	schemaBuilder *ourgraphql.DynamicSchemaBuilder
	mu            sync.RWMutex
}

func NewGraphQLHandler(svcService *services.ServiceService, connManager *database.ConnectionManager) *GraphQLHandler {
	return &GraphQLHandler{
		svcService:  svcService,
		connManager: connManager,
	}
}

// RefreshSchema forces a rebuild of the GraphQL schema
func (h *GraphQLHandler) RefreshSchema() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	schemaBuilder := ourgraphql.NewDynamicSchemaBuilder(h.svcService, h.connManager)
	schema, err := schemaBuilder.BuildSchema()
	if err != nil {
		return err
	}
	h.schema = schema
	return nil
}

// Handler returns the GraphQL handler
func (h *GraphQLHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.mu.RLock()
		schema := h.schema
		h.mu.RUnlock()

		if schema == nil {
			// Try to build it if it's nil
			err := h.RefreshSchema()
			if err != nil {
				c.JSON(200, gin.H{"errors": []gin.H{{"message": "Database not initialized for GraphQL schema build: " + err.Error()}}})
				return
			}
			h.mu.RLock()
			schema = h.schema
			h.mu.RUnlock()
		}

		server := gqlgenHandler.NewDefaultServer(schema)
		server.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler returns the GraphQL Playground handler
func (h *GraphQLHandler) PlaygroundHandler() gin.HandlerFunc {
	playground := gqlgenPlayground.Handler("GraphQL playground", "/api/v1/graphql")

	return func(c *gin.Context) {
		playground.ServeHTTP(c.Writer, c.Request)
	}
}
