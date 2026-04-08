package integration_test

import (
	"cms-backend/internal/models"
	"cms-backend/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamicDataCRUD tests the full CRUD lifecycle of dynamic data
func (s *IntegrationTestSuite) TestDynamicDataCRUD() {
	// Create a test service
	req := &services.CreateServiceRequest{
		Name:        "Test Dynamic Data Service",
		Description: "Service for testing dynamic data",
		Fields: []services.CreateFieldRequest{
			{
				Name:       "title",
				Label:      "Title",
				Type:       models.FieldTypeString,
				IsRequired: true,
			},
			{
				Name:  "content",
				Label: "Content",
				Type:  models.FieldTypeString,
			},
			{
				Name:  "rating",
				Label: "Rating",
				Type:  models.FieldTypeInteger,
			},
		},
	}

	svc, err := s.svcSvc.CreateService(s.ctx, req, 1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), svc)

	// Test CreateData
	data := map[string]interface{}{
		"title":   "Test Record 1",
		"content": "This is test record 1",
		"rating":  5,
	}
	createdData, err := s.dataSvc.CreateData(s.ctx, svc.Slug, data, 1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), createdData)

	idFloat, ok := createdData["id"].(float64) // SQLite returns float64 for numeric in maps typically, or int64 from LAST_INSERT_ROWID
	var id uint
	if !ok {
		idInt, ok2 := createdData["id"].(int64)
		require.True(s.T(), ok2, "ID should be a number")
		id = uint(idInt)
	} else {
		id = uint(idFloat)
	}

	assert.Equal(s.T(), "Test Record 1", createdData["title"])
	assert.Equal(s.T(), "This is test record 1", createdData["content"])

	// Test GetData
	fetchedData, err := s.dataSvc.GetData(s.ctx, svc.Slug, id, nil)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Test Record 1", fetchedData["title"])

	// Create a second record
	data2 := map[string]interface{}{
		"title":   "Test Record 2",
		"content": "This is test record 2",
		"rating":  4,
	}
	_, err = s.dataSvc.CreateData(s.ctx, svc.Slug, data2, 1)
	require.NoError(s.T(), err)

	// Test ListData
	listReq := &services.ListDataRequest{
		Page:      1,
		PageSize:  10,
		SortBy:    "id",
		SortOrder: "ASC",
	}
	listData, total, err := s.dataSvc.ListData(s.ctx, svc.Slug, listReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(2), total)
	assert.Len(s.T(), listData, 2)
	assert.Equal(s.T(), "Test Record 1", listData[0]["title"])
	assert.Equal(s.T(), "Test Record 2", listData[1]["title"])

	// Test UpdateData
	updateData := map[string]interface{}{
		"title":  "Updated Test Record 1",
		"rating": 10,
	}
	updatedData, err := s.dataSvc.UpdateData(s.ctx, svc.Slug, id, updateData, 1)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "Updated Test Record 1", updatedData["title"])

	// Use float64 for JSON unmarshaled numbers or interface{} to int/float64 mapping
	var ratingVal float64
	switch v := updatedData["rating"].(type) {
	case float64:
		ratingVal = v
	case int64:
		ratingVal = float64(v)
	case int:
		ratingVal = float64(v)
	}
	assert.Equal(s.T(), float64(10), ratingVal)
	assert.Equal(s.T(), "This is test record 1", updatedData["content"]) // shouldn't be overwritten if not specified

	// Test DeleteData
	err = s.dataSvc.DeleteData(s.ctx, svc.Slug, id)
	require.NoError(s.T(), err)

	// Verify DeleteData
	_, err = s.dataSvc.GetData(s.ctx, svc.Slug, id, nil)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), "record not found", err.Error())

	// List should only return 1 now
	listDataAfterDelete, totalAfterDelete, err := s.dataSvc.ListData(s.ctx, svc.Slug, listReq)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(1), totalAfterDelete)
	assert.Len(s.T(), listDataAfterDelete, 1)
}

func (s *IntegrationTestSuite) TestDynamicDataValidation() {
	// Create a test service
	req := &services.CreateServiceRequest{
		Name:        "Validation Test Service",
		Description: "Service for testing validations",
		Fields: []services.CreateFieldRequest{
			{
				Name:       "required_field",
				Label:      "Required Field",
				Type:       models.FieldTypeString,
				IsRequired: true,
			},
			{
				Name:  "email_field",
				Label: "Email Field",
				Type:  models.FieldTypeString,
				Validations: models.FieldValidations{
					{
						Type: "email",
					},
				},
			},
			{
				Name:  "min_field",
				Label: "Min Field",
				Type:  models.FieldTypeString,
				Validations: models.FieldValidations{
					{
						Type: "min",
						Value: 3,
					},
				},
			},
		},
	}

	svc, err := s.svcSvc.CreateService(s.ctx, req, 1)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), svc)

	// Test missing required field
	data1 := map[string]interface{}{
		"email_field": "test@test.com",
	}
	_, err = s.dataSvc.CreateData(s.ctx, svc.Slug, data1, 1)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "field 'Required Field' is required")

	// Test invalid email format
	data2 := map[string]interface{}{
		"required_field": "valid",
		"email_field": "invalid-email",
	}
	_, err = s.dataSvc.CreateData(s.ctx, svc.Slug, data2, 1)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "field 'Email Field' must be a valid email")

	// Test min length
	data3 := map[string]interface{}{
		"required_field": "valid",
		"min_field": "ab",
	}
	_, err = s.dataSvc.CreateData(s.ctx, svc.Slug, data3, 1)
	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "field 'Min Field' must be at least 3 characters")

	// Test successful creation
	data4 := map[string]interface{}{
		"required_field": "valid",
		"email_field": "test@example.com",
		"min_field": "abcd",
	}
	_, err = s.dataSvc.CreateData(s.ctx, svc.Slug, data4, 1)
	assert.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestDynamicDataRelations() {
	// 1. Create authors service
	authorsReq := &services.CreateServiceRequest{
		Name: "Authors",
		Fields: []services.CreateFieldRequest{
			{Name: "name", Label: "Name", Type: models.FieldTypeString},
		},
	}
	authorsSvc, err := s.svcSvc.CreateService(s.ctx, authorsReq, 1)
	require.NoError(s.T(), err)

	// 2. Create an author record
	authorData := map[string]interface{}{
		"name": "Jane Doe",
	}
	createdAuthor, err := s.dataSvc.CreateData(s.ctx, authorsSvc.Slug, authorData, 1)
	require.NoError(s.T(), err)

	authorID := getID(createdAuthor["id"])

	// 3. Create books service with relation to authors
	booksReq := &services.CreateServiceRequest{
		Name: "Books",
		Fields: []services.CreateFieldRequest{
			{Name: "title", Label: "Title", Type: models.FieldTypeString},
			{
				Name:  "author_id",
				Label: "Author",
				Type:  models.FieldTypeRelation,
				RelationConfig: &models.RelationConfig{
					RelatedServiceID: authorsSvc.ID,
					RelatedField:     "id",
					DisplayField:     "name",
					RelationType:     "belongs_to",
				},
			},
		},
	}
	booksSvc, err := s.svcSvc.CreateService(s.ctx, booksReq, 1)
	require.NoError(s.T(), err)

	// 4. Create a book record referencing the author
	bookData := map[string]interface{}{
		"title":     "Go Programming",
		"author_id": authorID,
	}
	createdBook, err := s.dataSvc.CreateData(s.ctx, booksSvc.Slug, bookData, 1)
	require.NoError(s.T(), err)
	bookID := getID(createdBook["id"])

	// 5. Test GetData with joins
	bookWithAuthor, err := s.dataSvc.GetData(s.ctx, booksSvc.Slug, uint(bookID), []string{"author_id"})
	require.NoError(s.T(), err)

	authorDataInBook, ok := bookWithAuthor["author_id_data"].(map[string]interface{})
	require.True(s.T(), ok, "author_id_data should be a map")
	assert.Equal(s.T(), "Jane Doe", authorDataInBook["name"])

	// 6. Test ListData with joins
	listReq := &services.ListDataRequest{
		Page:     1,
		PageSize: 10,
		Joins:    []string{"author_id"},
	}
	listData, _, err := s.dataSvc.ListData(s.ctx, booksSvc.Slug, listReq)
	require.NoError(s.T(), err)
	require.Len(s.T(), listData, 1)

	authorDataInList, ok := listData[0]["author_id_data"].(map[string]interface{})
	require.True(s.T(), ok, "author_id_data should be a map in list")
	assert.Equal(s.T(), "Jane Doe", authorDataInList["name"])
}

func getID(val interface{}) uint {
	switch v := val.(type) {
	case float64:
		return uint(v)
	case int64:
		return uint(v)
	case int:
		return uint(v)
	case uint:
		return v
	default:
		return 0
	}
}
