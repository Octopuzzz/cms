package graphql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
)

// DynamicExecutableSchema is a placeholder that implements graphql.ExecutableSchema
// In a real application, this would dynamically generate an AST based on service models
// and resolve queries using the DatabaseConnectionManager or DynamicDataService.
type DynamicExecutableSchema struct{}

func (s *DynamicExecutableSchema) Schema() *ast.Schema {
	// Provide a dummy static schema for the gateway
	schema := &ast.Schema{
		Query: &ast.Definition{
			Name: "Query",
			Fields: []*ast.FieldDefinition{
				{
					Name: "service",
					Type: ast.NonNullListType(ast.NonNullNamedType("JSON", nil), nil),
					Arguments: []*ast.ArgumentDefinition{
						{
							Name: "name",
							Type: ast.NonNullNamedType("String", nil),
						},
					},
				},
			},
		},
		Types: map[string]*ast.Definition{
			"Query": {
				Name: "Query",
				Kind: ast.Object,
				Fields: []*ast.FieldDefinition{
					{
						Name: "service",
						Type: ast.NonNullListType(ast.NonNullNamedType("JSON", nil), nil),
					},
				},
			},
			"JSON": {
				Name: "JSON",
				Kind: ast.Scalar,
			},
			"String": {
				Name: "String",
				Kind: ast.Scalar,
			},
		},
	}
	return schema
}

func (s *DynamicExecutableSchema) Complexity(typeName, field string, childComplexity int, args map[string]any) (int, bool) {
	return 1, false
}

func (s *DynamicExecutableSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	_ = graphql.GetOperationContext(ctx)

	// Create a generic response
	return func(ctx context.Context) *graphql.Response {
		return &graphql.Response{
			Data: []byte(`{"service": [{"id": 1, "status": "dynamic schema generation pending"}]}`),
		}
	}
}

// NewGateway returns a gin.HandlerFunc representing the GraphQL endpoint
func NewGateway() gin.HandlerFunc {
	// Initialize a new graphql handler using the DynamicExecutableSchema
	h := handler.NewDefaultServer(&DynamicExecutableSchema{})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// PlaygroundHandler returns a gin.HandlerFunc representing the GraphQL Playground
func PlaygroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL playground", "/api/v1/graphql")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
