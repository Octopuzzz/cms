package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cms-backend/internal/database"
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// DynamicSchemaBuilder creates a GraphQL schema dynamically based on services
type DynamicSchemaBuilder struct {
	svcService  *services.ServiceService
	connManager *database.ConnectionManager
}

func NewDynamicSchemaBuilder(svcService *services.ServiceService, connManager *database.ConnectionManager) *DynamicSchemaBuilder {
	return &DynamicSchemaBuilder{
		svcService:  svcService,
		connManager: connManager,
	}
}

// BuildSchema generates the executable GraphQL schema
func (b *DynamicSchemaBuilder) BuildSchema() (graphql.ExecutableSchema, error) {
	// 1. Fetch all active services
	activeServices, err := b.svcService.GetAllActiveServices()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %v", err)
	}

	// 2. Build GraphQL AST
	schemaAST := &ast.Schema{
		Query:    &ast.Definition{Name: "Query", Kind: ast.Object},
		Mutation: &ast.Definition{Name: "Mutation", Kind: ast.Object},
		Types:    make(map[string]*ast.Definition),
	}
	schemaAST.Types["Query"] = schemaAST.Query
	schemaAST.Types["Mutation"] = schemaAST.Mutation

	// Built-in types
	schemaAST.Types["Int"] = &ast.Definition{Name: "Int", Kind: ast.Scalar}
	schemaAST.Types["Float"] = &ast.Definition{Name: "Float", Kind: ast.Scalar}
	schemaAST.Types["String"] = &ast.Definition{Name: "String", Kind: ast.Scalar}
	schemaAST.Types["Boolean"] = &ast.Definition{Name: "Boolean", Kind: ast.Scalar}
	schemaAST.Types["ID"] = &ast.Definition{Name: "ID", Kind: ast.Scalar}
	schemaAST.Types["JSON"] = &ast.Definition{Name: "JSON", Kind: ast.Scalar}
	schemaAST.Types["DateTime"] = &ast.Definition{Name: "DateTime", Kind: ast.Scalar}

	for _, svc := range activeServices {
		typeName := b.toPascalCase(svc.Name)

		// Create object type for the service
		objDef := &ast.Definition{
			Name: typeName,
			Kind: ast.Object,
		}

		// Add basic ID and timestamp fields (assuming every table has them or checking could be done)
		objDef.Fields = append(objDef.Fields, &ast.FieldDefinition{
			Name: "id",
			Type: ast.NamedType("ID", nil),
		})

		inputTypeName := typeName + "Input"
		inputDef := &ast.Definition{
			Name: inputTypeName,
			Kind: ast.InputObject,
		}

		for _, field := range svc.Fields {
			gqlType := b.mapFieldTypeToGQL(field.Type)

			// Non-nullable if required
			var fieldType *ast.Type
			if field.IsRequired {
				fieldType = ast.NonNullNamedType(gqlType, nil)
			} else {
				fieldType = ast.NamedType(gqlType, nil)
			}

			objDef.Fields = append(objDef.Fields, &ast.FieldDefinition{
				Name: field.Name,
				Type: fieldType,
			})

			// For input type, only allow string, int, float, boolean, json, datetime
			inputDef.Fields = append(inputDef.Fields, &ast.FieldDefinition{
				Name: field.Name,
				Type: fieldType,
			})
		}

		schemaAST.Types[typeName] = objDef
		schemaAST.Types[inputTypeName] = inputDef

		// Add queries
		// Get by ID
		schemaAST.Query.Fields = append(schemaAST.Query.Fields, &ast.FieldDefinition{
			Name: "get" + typeName,
			Arguments: ast.ArgumentDefinitionList{
				{Name: "id", Type: ast.NonNullNamedType("ID", nil)},
			},
			Type: ast.NamedType(typeName, nil),
		})

		// List
		schemaAST.Query.Fields = append(schemaAST.Query.Fields, &ast.FieldDefinition{
			Name: "list" + typeName + "s",
			Arguments: ast.ArgumentDefinitionList{
				{Name: "limit", Type: ast.NamedType("Int", nil)},
				{Name: "offset", Type: ast.NamedType("Int", nil)},
			},
			Type: ast.ListType(ast.NamedType(typeName, nil), nil),
		})

		// Add mutations
		// Create
		schemaAST.Mutation.Fields = append(schemaAST.Mutation.Fields, &ast.FieldDefinition{
			Name: "create" + typeName,
			Arguments: ast.ArgumentDefinitionList{
				{Name: "input", Type: ast.NonNullNamedType(inputTypeName, nil)},
			},
			Type: ast.NamedType(typeName, nil),
		})

		// Update
		schemaAST.Mutation.Fields = append(schemaAST.Mutation.Fields, &ast.FieldDefinition{
			Name: "update" + typeName,
			Arguments: ast.ArgumentDefinitionList{
				{Name: "id", Type: ast.NonNullNamedType("ID", nil)},
				{Name: "input", Type: ast.NonNullNamedType(inputTypeName, nil)},
			},
			Type: ast.NamedType(typeName, nil),
		})

		// Delete
		schemaAST.Mutation.Fields = append(schemaAST.Mutation.Fields, &ast.FieldDefinition{
			Name: "delete" + typeName,
			Arguments: ast.ArgumentDefinitionList{
				{Name: "id", Type: ast.NonNullNamedType("ID", nil)},
			},
			Type: ast.NamedType("Boolean", nil),
		})
	}

	// Dynamic Executable Schema requires custom implementations
	// Due to gqlgen architecture, true dynamic schema without code generation is complex.
	// Here we implement a custom ExecutableSchema wrapper.

	execSchema := &dynamicExecutableSchema{
		schemaAST:   schemaAST,
		svcService:  b.svcService,
		connManager: b.connManager,
		services:    activeServices,
	}

	return execSchema, nil
}

func (b *DynamicSchemaBuilder) mapFieldTypeToGQL(t models.FieldType) string {
	switch t {
	case models.FieldTypeString, models.FieldTypeText, models.FieldTypeUUID, models.FieldTypeEmail, models.FieldTypeURL:
		return "String"
	case models.FieldTypeInteger:
		return "Int"
	case models.FieldTypeFloat:
		return "Float"
	case models.FieldTypeBoolean:
		return "Boolean"
	case models.FieldTypeDate, models.FieldTypeDateTime:
		return "DateTime"
	case models.FieldTypeJSON:
		return "JSON"
	default:
		return "String"
	}
}

func (b *DynamicSchemaBuilder) toPascalCase(s string) string {
	parts := strings.Split(strings.ReplaceAll(s, "-", "_"), "_")
	var result string
	for _, p := range parts {
		if len(p) > 0 {
			result += strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return result
}

// dynamicExecutableSchema implements graphql.ExecutableSchema
type dynamicExecutableSchema struct {
	schemaAST   *ast.Schema
	svcService  *services.ServiceService
	connManager *database.ConnectionManager
	services    []models.Service
}

func (e *dynamicExecutableSchema) Schema() *ast.Schema {
	return e.schemaAST
}

func (e *dynamicExecutableSchema) Complexity(typeName, field string, childComplexity int, args map[string]interface{}) (int, bool) {
	return 1, true
}

func (e *dynamicExecutableSchema) Exec(ctx context.Context) graphql.ResponseHandler {
	return func(ctx context.Context) *graphql.Response {
		rc := graphql.GetOperationContext(ctx)

		// Very simplified execution logic for dynamic schema
		// In a production system, you'd use a more robust GraphQL execution engine like graphql-go
		// or properly hook into gqlgen's execution path.

		if len(rc.Operation.SelectionSet) == 0 {
			return &graphql.Response{Errors: gqlerror.List{{Message: "empty selection set"}}}
		}

		op := rc.Operation

		var data map[string]interface{}
		var err error

		if op.Operation == ast.Query {
			data, err = e.executeQuery(ctx, op, rc.Variables)
		} else if op.Operation == ast.Mutation {
			data, err = e.executeMutation(ctx, op, rc.Variables)
		}

		if err != nil {
			return &graphql.Response{Errors: gqlerror.List{{Message: err.Error()}}}
		}

		b, err := json.Marshal(data)
		if err != nil {
			return &graphql.Response{Errors: gqlerror.List{{Message: "Failed to marshal data"}}}
		}
		return &graphql.Response{Data: b}
	}
}

func (e *dynamicExecutableSchema) executeQuery(ctx context.Context, op *ast.OperationDefinition, vars map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	dynamicService := services.NewDynamicDataService(e.connManager, e.svcService)

	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}

		// Handle list query
		if strings.HasPrefix(field.Name, "list") {
			// Extract service name: listUsers -> user
			svcName := strings.TrimPrefix(field.Name, "list")
			svcName = strings.TrimSuffix(svcName, "s")
			slug := toKebabCase(svcName)

			// Extract arguments
			limit := 10
			offset := 0

			limitArg := field.Arguments.ForName("limit")
			if limitArg != nil && limitArg.Value != nil {
				fmt.Sscanf(limitArg.Value.Raw, "%d", &limit)
			}
			offsetArg := field.Arguments.ForName("offset")
			if offsetArg != nil && offsetArg.Value != nil {
				fmt.Sscanf(offsetArg.Value.Raw, "%d", &offset)
			}

			page := (offset / limit) + 1

			req := &services.ListDataRequest{
				PageSize:  limit,
				Page: page,
			}

			data, _, err := dynamicService.ListData(ctx, slug, req)
			if err != nil {
				return nil, err
			}

			result[field.Alias] = data

		// Handle get by ID query
		} else if strings.HasPrefix(field.Name, "get") {
			svcName := strings.TrimPrefix(field.Name, "get")
			slug := toKebabCase(svcName)

			idArg := field.Arguments.ForName("id")
			var id uint
			if idArg != nil && idArg.Value != nil {
				fmt.Sscanf(idArg.Value.Raw, "%d", &id)
			}

			data, err := dynamicService.GetData(ctx, slug, id, nil)
			if err != nil {
				return nil, err
			}
			result[field.Alias] = data
		}
	}

	return result, nil
}

func (e *dynamicExecutableSchema) executeMutation(ctx context.Context, op *ast.OperationDefinition, vars map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	dynamicService := services.NewDynamicDataService(e.connManager, e.svcService)

	for _, sel := range op.SelectionSet {
		field, ok := sel.(*ast.Field)
		if !ok {
			continue
		}

		if strings.HasPrefix(field.Name, "create") {
			svcName := strings.TrimPrefix(field.Name, "create")
			slug := toKebabCase(svcName)

			inputData := map[string]interface{}{}
			inputArg := field.Arguments.ForName("input")
			if inputArg != nil && inputArg.Value != nil {
				// Simplified parsing, a real implementation needs to map variables correctly
				// based on AST variables definition and operation
				if inputArg.Value.Kind == ast.Variable {
					if val, ok := vars[inputArg.Value.Raw]; ok {
						if m, ok := val.(map[string]interface{}); ok {
							inputData = m
						}
					}
				}
			}

			data, err := dynamicService.CreateData(ctx, slug, inputData, 1) // default admin ID
			if err != nil {
				return nil, err
			}
			result[field.Alias] = data
		} else if strings.HasPrefix(field.Name, "update") {
			svcName := strings.TrimPrefix(field.Name, "update")
			slug := toKebabCase(svcName)

			var id uint
			idArg := field.Arguments.ForName("id")
			if idArg != nil && idArg.Value != nil {
				fmt.Sscanf(idArg.Value.Raw, "%d", &id)
			}

			inputData := map[string]interface{}{}
			inputArg := field.Arguments.ForName("input")
			if inputArg != nil && inputArg.Value != nil {
				if inputArg.Value.Kind == ast.Variable {
					if val, ok := vars[inputArg.Value.Raw]; ok {
						if m, ok := val.(map[string]interface{}); ok {
							inputData = m
						}
					}
				}
			}

			data, err := dynamicService.UpdateData(ctx, slug, id, inputData, 1) // default admin ID
			if err != nil {
				return nil, err
			}
			result[field.Alias] = data
		} else if strings.HasPrefix(field.Name, "delete") {
			svcName := strings.TrimPrefix(field.Name, "delete")
			slug := toKebabCase(svcName)

			var id uint
			idArg := field.Arguments.ForName("id")
			if idArg != nil && idArg.Value != nil {
				fmt.Sscanf(idArg.Value.Raw, "%d", &id)
			}

			err := dynamicService.DeleteData(ctx, slug, id)
			if err != nil {
				return nil, err
			}
			result[field.Alias] = true
		}
	}

	return result, nil
}

func toKebabCase(s string) string {
	// Simple PascalCase to kebab-case
	var result string
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "-"
		}
		result += strings.ToLower(string(r))
	}
	return result
}
