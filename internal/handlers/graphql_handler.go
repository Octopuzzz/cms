package handlers

import (
	"context"

	"cms-backend/internal/contextutils"
	"cms-backend/internal/graphql"
	"cms-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// GraphQLHandler handles incoming GraphQL queries and mutations
type GraphQLHandler struct {
	engine *graphql.Engine
}

func NewGraphQLHandler(engine *graphql.Engine) *GraphQLHandler {
	return &GraphQLHandler{engine: engine}
}

// Handle executes a GraphQL query
// @Summary      Execute GraphQL Query
// @Description  Executes a dynamic GraphQL query/mutation based on generated schemas
// @Tags         graphql
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        query body object true "GraphQL query, variables, operationName"
// @Success      200  {object}  map[string]interface{}
// @Router       /graphql [post]
func (h *GraphQLHandler) Handle(c *gin.Context) {
	var body struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid graphql request body")
		return
	}

	ctx := c.Request.Context()

	// Add Gin context info (like user ID) to the context for resolvers
	if userID, exists := c.Get("userID"); exists {
		ctx = context.WithValue(ctx, contextutils.UserIDKey, userID)
	}

	result := h.engine.ExecuteQuery(body.Query, body.Variables, body.OperationName, ctx)

	if len(result.Errors) > 0 {
		c.JSON(400, result)
		return
	}

	c.JSON(200, result)
}

// ReloadSchema triggers a schema reload
// @Summary      Reload GraphQL Schema
// @Description  Rebuilds the GraphQL schema from current active services
// @Tags         graphql
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /graphql/reload [post]
func (h *GraphQLHandler) ReloadSchema(c *gin.Context) {
	if err := h.engine.ReloadSchema(c.Request.Context()); err != nil {
		response.InternalError(c, err)
		return
	}
	response.OKMessage(c, "graphql schema reloaded successfully", nil)
}
