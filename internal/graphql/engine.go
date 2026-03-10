package graphql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"cms-backend/internal/services"
	"cms-backend/pkg/logger"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Engine represents the dynamic GraphQL engine
type Engine struct {
	dataSvc    *services.DynamicDataService
	svcService *services.ServiceService
	schema     *graphql.Schema
	schemaMu   sync.RWMutex
	log        logger.Logger
}

// NewEngine creates a new GraphQL engine
func NewEngine(dataSvc *services.DynamicDataService, svcService *services.ServiceService) *Engine {
	return &Engine{
		dataSvc:    dataSvc,
		svcService: svcService,
		log:        logger.GetLogger(),
	}
}

// ReloadSchema rebuilds the GraphQL schema from the current services
func (e *Engine) ReloadSchema(ctx context.Context) error {
	svcs, _, err := e.svcService.ListServices(ctx, 1, 1000, nil) // Fetch all active services
	if err != nil {
		return fmt.Errorf("failed to fetch services for schema: %w", err)
	}

	queryFields := graphql.Fields{}
	mutationFields := graphql.Fields{}

	for _, svc := range svcs {
		if !svc.IsActive {
			continue
		}

		objType := e.buildObjectType(&svc)

		// List query
		queryFields[svc.Slug] = &graphql.Field{
			Type:        graphql.NewList(objType),
			Description: fmt.Sprintf("List of %s", svc.Name),
			Args: graphql.FieldConfigArgument{
				"page":  &graphql.ArgumentConfig{Type: graphql.Int},
				"limit": &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: e.resolveList(svc.Slug),
		}

		// Get by ID query
		singleName := svc.Slug
		if strings.HasSuffix(singleName, "s") {
			singleName = singleName[:len(singleName)-1]
		}

		queryFields[singleName] = &graphql.Field{
			Type:        objType,
			Description: fmt.Sprintf("Get %s by ID", svc.Name),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: e.resolveGet(svc.Slug),
		}

		caser := cases.Title(language.Und)

		// Create mutation
		mutationFields["create"+caser.String(singleName)] = &graphql.Field{
			Type: objType,
			Args: graphql.FieldConfigArgument{
				"data": &graphql.ArgumentConfig{Type: e.buildInputType(&svc, false)},
			},
			Resolve: e.resolveCreate(svc.Slug),
		}

		// Update mutation
		mutationFields["update"+caser.String(singleName)] = &graphql.Field{
			Type: objType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				"data": &graphql.ArgumentConfig{Type: e.buildInputType(&svc, true)},
			},
			Resolve: e.resolveUpdate(svc.Slug),
		}

		// Delete mutation
		mutationFields["delete"+caser.String(singleName)] = &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
			},
			Resolve: e.resolveDelete(svc.Slug),
		}
	}

	// Always provide at least a dummy query to avoid schema errors if no services exist
	if len(queryFields) == 0 {
		queryFields["_empty"] = &graphql.Field{
			Type:    graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) { return "No services defined", nil },
		}
	}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(graphql.ObjectConfig{
			Name:   "Query",
			Fields: queryFields,
		}),
	}

	if len(mutationFields) > 0 {
		schemaConfig.Mutation = graphql.NewObject(graphql.ObjectConfig{
			Name: "Mutation",
			Fields: mutationFields,
		})
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return fmt.Errorf("failed to build GraphQL schema: %w", err)
	}

	e.schemaMu.Lock()
	e.schema = &schema
	e.schemaMu.Unlock()

	e.log.Info("GraphQL schema reloaded successfully", "services_count", len(svcs))
	return nil
}

func (e *Engine) Schema() *graphql.Schema {
	e.schemaMu.RLock()
	defer e.schemaMu.RUnlock()
	return e.schema
}

// ExecuteQuery executes a GraphQL query against the current schema
func (e *Engine) ExecuteQuery(query string, variables map[string]interface{}, operationName string, ctx context.Context) *graphql.Result {
	e.schemaMu.RLock()
	schema := e.schema
	e.schemaMu.RUnlock()

	if schema == nil || schema.QueryType() == nil {
		return &graphql.Result{
			Errors: []gqlerrors.FormattedError{gqlerrors.FormatError(fmt.Errorf("Schema not initialized"))},
		}
	}

	params := graphql.Params{
		Schema:         *schema,
		RequestString:  query,
		VariableValues: variables,
		OperationName:  operationName,
		Context:        ctx,
	}

	return graphql.Do(params)
}
