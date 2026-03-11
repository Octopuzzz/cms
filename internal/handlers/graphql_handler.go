package handlers

import (
	"context"
	"net/http"

	"cms-backend/internal/services"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/graphql"
)

// GraphQLHandler handles incoming GraphQL requests
type GraphQLHandler struct {
	graphqlService *services.GraphQLService
}

// NewGraphQLHandler creates a new GraphQLHandler
func NewGraphQLHandler(gs *services.GraphQLService) *GraphQLHandler {
	return &GraphQLHandler{graphqlService: gs}
}

// Request struct for parsing incoming graphql post body
type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operationName"`
	Variables map[string]interface{} `json:"variables"`
}

// HandleGraphQL processes a GraphQL query/mutation for a specific service
func (h *GraphQLHandler) HandleGraphQL(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Service slug is required")
		return
	}

	var pData postData
	if err := c.ShouldBindJSON(&pData); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Extract user ID from context for auth tracking in mutations
	var userID uint
	if uid, exists := c.Get("userID"); exists {
		if id, ok := uid.(uint); ok {
			userID = id
		}
	}

	// Generate or retrieve the schema
	schema, err := h.graphqlService.GenerateSchema(c.Request.Context(), slug)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Add userID to context so resolves can access it
	ctx := context.WithValue(c.Request.Context(), "userID", userID)

	// Execute the query
	result := graphql.Do(graphql.Params{
		Schema:         *schema,
		RequestString:  pData.Query,
		VariableValues: pData.Variables,
		OperationName:  pData.Operation,
		Context:        ctx,
	})

	if len(result.Errors) > 0 {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	c.JSON(http.StatusOK, result)
}
