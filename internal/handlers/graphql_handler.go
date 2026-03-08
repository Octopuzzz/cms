package handlers

import (
	"context"

	"cms-backend/internal/graphql"
	"cms-backend/internal/graphql/generated"
	"cms-backend/internal/services"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

// GraphQLHandler sets up the GraphQL server
type GraphQLHandler struct {
	resolver *graphql.Resolver
}

func NewGraphQLHandler(dataSvc *services.DynamicDataService) *GraphQLHandler {
	return &GraphQLHandler{
		resolver: &graphql.Resolver{
			DataSvc: dataSvc,
		},
	}
}

// Serve returns a gin handler for the GraphQL endpoint
func (h *GraphQLHandler) Serve() gin.HandlerFunc {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: h.resolver}))

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		importContext := context.WithValue(ctx, "user_id", uint(1)) // default
		if id, exists := c.Get("user_id"); exists {
			if uid, ok := id.(uint); ok {
				importContext = context.WithValue(ctx, "user_id", uid)
			}
		}

		srv.ServeHTTP(c.Writer, c.Request.WithContext(importContext))
	}
}

// Playground returns a gin handler for the GraphQL Playground
func (h *GraphQLHandler) Playground() gin.HandlerFunc {
	hnd := playground.Handler("GraphQL Playground", "/api/v1/graphql")

	return func(c *gin.Context) {
		hnd.ServeHTTP(c.Writer, c.Request)
	}
}
