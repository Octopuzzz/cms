package graphql

import (
	"cms-backend/internal/models"
	"github.com/graphql-go/graphql"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// buildObjectType creates a GraphQL ObjectType for a Service
func (e *Engine) buildObjectType(svc *models.Service) *graphql.Object {
	fields := graphql.Fields{
		"id":         &graphql.Field{Type: graphql.ID},
		"created_at": &graphql.Field{Type: graphql.String},
		"updated_at": &graphql.Field{Type: graphql.String},
	}

	for _, field := range svc.Fields {
		fields[field.Name] = &graphql.Field{
			Type:        mapFieldType(field.Type),
			Description: field.Description,
		}
	}

	caser := cases.Title(language.Und)
	return graphql.NewObject(graphql.ObjectConfig{
		Name:        caser.String(svc.Name),
		Description: svc.Description,
		Fields:      fields,
	})
}

// buildInputType creates a GraphQL InputObject for Create/Update mutations
func (e *Engine) buildInputType(svc *models.Service, isUpdate bool) *graphql.InputObject {
	fields := graphql.InputObjectConfigFieldMap{}

	for _, field := range svc.Fields {
		// Auto-generated fields aren't input via mutations
		if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" {
			continue
		}

		fieldType := mapInputType(field.Type)
		if field.IsRequired && !isUpdate {
			fieldType = graphql.NewNonNull(fieldType)
		}

		fields[field.Name] = &graphql.InputObjectFieldConfig{
			Type:        fieldType,
			Description: field.Description,
		}
	}

	suffix := "Create"
	if isUpdate {
		suffix = "Update"
	}

	caser := cases.Title(language.Und)
	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name:        caser.String(svc.Name) + suffix + "Input",
		Description: "Input for " + svc.Name,
		Fields:      fields,
	})
}

// mapFieldType maps CMS field types to GraphQL output types
func mapFieldType(fType models.FieldType) graphql.Output {
	switch fType {
	case models.FieldTypeInteger:
		return graphql.Int
	case models.FieldTypeFloat:
		return graphql.Float
	case models.FieldTypeBoolean:
		return graphql.Boolean
	case models.FieldTypeString, models.FieldTypeText, models.FieldTypeEmail, models.FieldTypePhone:
		return graphql.String
	case models.FieldTypeJSON:
		return graphql.String // Can use custom scalar or simply return JSON string representation
	case models.FieldTypeUUID:
		return graphql.ID
	case models.FieldTypeDate, models.FieldTypeDateTime:
		return graphql.String
	default:
		return graphql.String
	}
}

// mapInputType maps CMS field types to GraphQL input types
func mapInputType(fType models.FieldType) graphql.Input {
	switch fType {
	case models.FieldTypeInteger:
		return graphql.Int
	case models.FieldTypeFloat:
		return graphql.Float
	case models.FieldTypeBoolean:
		return graphql.Boolean
	case models.FieldTypeString, models.FieldTypeText, models.FieldTypeEmail, models.FieldTypePhone:
		return graphql.String
	case models.FieldTypeJSON:
		return graphql.String
	case models.FieldTypeUUID:
		return graphql.ID
	case models.FieldTypeDate, models.FieldTypeDateTime:
		return graphql.String
	default:
		return graphql.String
	}
}
