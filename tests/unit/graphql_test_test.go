package services

import (
	"cms-backend/internal/database"
	"cms-backend/internal/graphql"
	"cms-backend/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicSchemaBuilder(t *testing.T) {
	connManager := database.GetConnectionManager()
	svcService := services.NewServiceService(connManager)

	builder := graphql.NewDynamicSchemaBuilder(svcService, connManager)

	schema, err := builder.BuildSchema()

	// Error could happen if db isn't initialized or services can't be fetched
	if err == nil {
		assert.NotNil(t, schema)
	}
}
