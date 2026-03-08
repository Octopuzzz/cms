package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"cms-backend/internal/models"
	"cms-backend/internal/services"
)

// GenerateGraphQL generating the gqlgen project files based on DB services
func GenerateGraphQL(ctx context.Context, svcService *services.ServiceService) error {
	svcs, _, err := svcService.ListServices(ctx, 1, 1000, nil)
	if err != nil {
		return err
	}

	// 1. Write schema.graphqls
	schema := "scalar JSON\n\n"

	schema += "type Query {\n"
	if len(svcs) == 0 {
		schema += "  _empty: String\n"
	}
	for _, svc := range svcs {
		schema += fmt.Sprintf("  %s(id: ID!): %s\n", svc.Slug, svc.Name)
		schema += fmt.Sprintf("  %sList(page: Int, limit: Int): [%s]\n", svc.Slug, svc.Name)
	}
	schema += "}\n\n"

	schema += "type Mutation {\n"
	if len(svcs) == 0 {
		schema += "  _empty: String\n"
	}
	for _, svc := range svcs {
		schema += fmt.Sprintf("  create%s(data: JSON!): %s\n", svc.Name, svc.Name)
		schema += fmt.Sprintf("  update%s(id: ID!, data: JSON!): %s\n", svc.Name, svc.Name)
		schema += fmt.Sprintf("  delete%s(id: ID!): Boolean\n", svc.Name)
	}
	schema += "}\n\n"

	for _, svc := range svcs {
		schema += fmt.Sprintf("type %s {\n", svc.Name)
		hasId := false
		for _, f := range svc.Fields {
			if f.Name == "id" { hasId = true; break }
		}
		if !hasId {
			schema += "  id: ID!\n"
		}
		for _, f := range svc.Fields {
			gqlType := "String"
			switch f.Type {
			case models.FieldTypeInteger: gqlType = "Int"
			case models.FieldTypeFloat: gqlType = "Float"
			case models.FieldTypeBoolean: gqlType = "Boolean"
			case models.FieldTypeJSON: gqlType = "JSON"
			}
			if f.Name == "id" { gqlType = "ID" }
			if !f.IsNullable {
				gqlType += "!"
			}
			schema += fmt.Sprintf("  %s: %s\n", f.Name, gqlType)
		}
		schema += "}\n\n"
	}

	if err := os.WriteFile("internal/graphql/schema.graphqls", []byte(schema), 0644); err != nil {
		return err
	}

	// 2. Write gqlgen.yml
	yml := `schema:
  - internal/graphql/*.graphqls
exec:
  filename: internal/graphql/generated/generated.go
  package: generated
model:
  filename: internal/graphql/model/models_gen.go
  package: model
resolver:
  layout: follow-schema
  dir: internal/graphql
  package: graphql
models:
  JSON:
    model: github.com/99designs/gqlgen/graphql.Map
`
	for _, svc := range svcs {
		yml += fmt.Sprintf("  %s:\n    model: github.com/99designs/gqlgen/graphql.Map\n", svc.Name)
	}

	if err := os.WriteFile("gqlgen.yml", []byte(yml), 0644); err != nil {
		return err
	}

	// 3. Write resolvers
	resolvers := `package graphql

import (
	"context"
	"strconv"
	"fmt"

	"cms-backend/internal/graphql/generated"
	"cms-backend/internal/services"
	"cms-backend/internal/middleware"
)

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

func parseID(id string) uint {
	v, _ := strconv.ParseUint(id, 10, 64)
	return uint(v)
}

func getUserID(ctx context.Context) uint {
	if uid, ok := ctx.Value(middleware.UserIDKey).(uint); ok {
		return uid
	}
	return 0
}

`
	if len(svcs) == 0 {
		resolvers += `
func (r *queryResolver) Empty(ctx context.Context) (*string, error) { return nil, nil }
func (r *mutationResolver) Empty(ctx context.Context) (*string, error) { return nil, nil }
`
	}

	for _, svc := range svcs {
		// Queries
		resolvers += fmt.Sprintf(`
func (r *queryResolver) %s(ctx context.Context, id string) (map[string]interface{}, error) {
	return r.DataSvc.GetData(ctx, "%s", parseID(id), nil)
}
func (r *queryResolver) %sList(ctx context.Context, page *int, limit *int) ([]map[string]interface{}, error) {
	p, l := 1, 20
	if page != nil { p = *page }
	if limit != nil { l = *limit }
	rows, _, err := r.DataSvc.ListData(ctx, "%s", &services.ListDataRequest{Page: p, PageSize: l})
	return rows, err
}
`, svc.Name, svc.Slug, svc.Name, svc.Slug)

		// Mutations
		resolvers += fmt.Sprintf(`
func (r *mutationResolver) Create%s(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	return r.DataSvc.CreateData(ctx, "%s", data, getUserID(ctx))
}
func (r *mutationResolver) Update%s(ctx context.Context, id string, data map[string]interface{}) (map[string]interface{}, error) {
	return r.DataSvc.UpdateData(ctx, "%s", parseID(id), data, getUserID(ctx))
}
func (r *mutationResolver) Delete%s(ctx context.Context, id string) (*bool, error) {
	err := r.DataSvc.DeleteData(ctx, "%s", parseID(id))
	success := err == nil
	return &success, err
}
`, svc.Name, svc.Slug, svc.Name, svc.Slug, svc.Name, svc.Slug)
	}

	if err := os.WriteFile("internal/graphql/schema.resolvers.go", []byte(resolvers), 0644); err != nil {
		return err
	}

	// 4. Run gqlgen
	cmd := exec.Command("go", "run", "github.com/99designs/gqlgen", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
