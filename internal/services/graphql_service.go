package services

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/graphql-go/graphql"
)

// GraphQLService generates and executes GraphQL queries/mutations
type GraphQLService struct {
	dynamicDataService *DynamicDataService
	serviceService     *ServiceService

	schemaCache        map[string]*graphql.Schema
	schemaMutex        sync.RWMutex
}

// NewGraphQLService creates a new GraphQLService
func NewGraphQLService(dds *DynamicDataService, ss *ServiceService) *GraphQLService {
	return &GraphQLService{
		dynamicDataService: dds,
		serviceService:     ss,
		schemaCache:        make(map[string]*graphql.Schema),
	}
}

// ClearSchemaCache removes a cached schema for a service slug, forcing regeneration
func (s *GraphQLService) ClearSchemaCache(slug string) {
	s.schemaMutex.Lock()
	defer s.schemaMutex.Unlock()
	delete(s.schemaCache, slug)
}

// mapFieldType maps a CMS field type to a GraphQL type
func (s *GraphQLService) mapFieldType(fieldType string) graphql.Type {
	switch strings.ToLower(fieldType) {
	case "integer":
		return graphql.Int
	case "float":
		return graphql.Float
	case "boolean":
		return graphql.Boolean
	case "string", "uuid", "json", "text":
		return graphql.String
	case "timestamp", "date", "datetime":
		return graphql.DateTime
	default:
		return graphql.String
	}
}

// GenerateSchema generates or returns a cached GraphQL schema for a given service slug
func (s *GraphQLService) GenerateSchema(ctx context.Context, slug string) (*graphql.Schema, error) {
	// 1. Check Cache
	s.schemaMutex.RLock()
	cachedSchema, exists := s.schemaCache[slug]
	s.schemaMutex.RUnlock()

	if exists && cachedSchema != nil {
		return cachedSchema, nil
	}

	// 2. Not in cache, build schema
	svc, err := s.serviceService.GetServiceBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// 3. Build the base GraphQL Object Type for this service
	fields := graphql.Fields{
		"id":         &graphql.Field{Type: graphql.Int},
		"created_at": &graphql.Field{Type: graphql.DateTime},
		"updated_at": &graphql.Field{Type: graphql.DateTime},
		"deleted_at": &graphql.Field{Type: graphql.DateTime},
		"created_by": &graphql.Field{Type: graphql.Int},
		"updated_by": &graphql.Field{Type: graphql.Int},
	}

	for _, field := range svc.Fields {
		fields[field.Name] = &graphql.Field{
			Type: s.mapFieldType(string(field.Type)),
		}
	}

	objType := graphql.NewObject(graphql.ObjectConfig{
		Name:   strings.Title(slug),
		Fields: fields,
	})

	// 2. Build Queries (Get, List)
	queryFields := graphql.Fields{
		"get": &graphql.Field{
			Type: objType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, ok := p.Args["id"].(int)
				if !ok {
					return nil, fmt.Errorf("id is required and must be an integer")
				}

				// Optional: get joins from GraphQL query context if needed, empty for now
				return s.dynamicDataService.GetData(p.Context, slug, uint(id), []string{})
			},
		},
		"list": &graphql.Field{
			Type: graphql.NewList(objType),
			Args: graphql.FieldConfigArgument{
				"page":     &graphql.ArgumentConfig{Type: graphql.Int},
				"pageSize": &graphql.ArgumentConfig{Type: graphql.Int},
				"sortBy":   &graphql.ArgumentConfig{Type: graphql.String},
				"sortOrder": &graphql.ArgumentConfig{Type: graphql.String},
				"search":   &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				req := ListDataRequest{
					Filters: make(map[string]interface{}),
				}
				if page, ok := p.Args["page"].(int); ok {
					req.Page = page
				}
				if pageSize, ok := p.Args["pageSize"].(int); ok {
					req.PageSize = pageSize
				}
				if sortBy, ok := p.Args["sortBy"].(string); ok {
					req.SortBy = sortBy
				}
				if sortOrder, ok := p.Args["sortOrder"].(string); ok {
					req.SortOrder = sortOrder
				}
				if search, ok := p.Args["search"].(string); ok {
					req.Search = search
				}

				res, _, err := s.dynamicDataService.ListData(p.Context, slug, &req)
				if err != nil {
					return nil, err
				}
				return res, nil
			},
		},
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	})

	// 3. Build Mutations (Create, Update, Delete)
	mutationArgs := graphql.FieldConfigArgument{}
	for _, field := range svc.Fields {
		mutationArgs[field.Name] = &graphql.ArgumentConfig{
			Type: s.mapFieldType(string(field.Type)),
		}
	}

	mutationFields := graphql.Fields{
		"create": &graphql.Field{
			Type: objType,
			Args: mutationArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// We expect user ID to be in context
				var userID uint
				if id, ok := p.Context.Value("userID").(uint); ok {
					userID = id
				}

				// Data has to be massaged from p.Args map to match DB insertion
				data := make(map[string]interface{})
				for k, v := range p.Args {
					data[k] = v
				}

				// The create method needs standard map structure
				res, err := s.dynamicDataService.CreateData(p.Context, slug, data, userID)
				if err != nil {
					return nil, err
				}

				// Optional: retrieve record back if the create method returned it in a nested way
				return res, nil
			},
		},
		"update": &graphql.Field{
			Type: objType,
			Args: appendArgs(mutationArgs, "id", &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)}),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, ok := p.Args["id"].(int)
				if !ok {
					return nil, fmt.Errorf("id is required and must be an integer")
				}

				var userID uint
				if uid, ok := p.Context.Value("userID").(uint); ok {
					userID = uid
				}

				// Filter out 'id' from data
				data := make(map[string]interface{})
				for k, v := range p.Args {
					if k != "id" {
						data[k] = v
					}
				}

				return s.dynamicDataService.UpdateData(p.Context, slug, uint(id), data, userID)
			},
		},
		"delete": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, ok := p.Args["id"].(int)
				if !ok {
					return nil, fmt.Errorf("id is required and must be an integer")
				}
				err := s.dynamicDataService.DeleteData(p.Context, slug, uint(id))
				if err != nil {
					return nil, err
				}
				return true, nil
			},
		},
	}

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: mutationFields,
	})

	// 4. Create Schema
	schemaConfig := graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create graphql schema: %w", err)
	}

	// Save to cache
	s.schemaMutex.Lock()
	s.schemaCache[slug] = &schema
	s.schemaMutex.Unlock()

	return &schema, nil
}

// Helper to append args without mutating the original
func appendArgs(base graphql.FieldConfigArgument, key string, val *graphql.ArgumentConfig) graphql.FieldConfigArgument {
	res := graphql.FieldConfigArgument{}
	for k, v := range base {
		res[k] = v
	}
	res[key] = val
	return res
}
