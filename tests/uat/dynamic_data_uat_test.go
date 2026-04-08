package uat_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUAT_DynamicDataAPI(t *testing.T) {
	router := setupTestServer(t)
	token := loginAsAdmin(t, router)

	// 1. Create a service
	serviceData := map[string]interface{}{
		"name":        "UAT Dynamic Data Test",
		"description": "Products for UAT dynamic data testing",
		"fields": []interface{}{
			map[string]interface{}{
				"name": "title", "label": "Title", "type": "string", "is_required": true,
			},
			map[string]interface{}{
				"name": "price", "label": "Price", "type": "float",
			},
		},
		"permissions": []interface{}{
			map[string]interface{}{
				"role_id": 1, // Assume 1 is the default admin role created
				"can_create": true,
				"can_read": true,
				"can_update": true,
				"can_delete": true,
			},
		},
	}

	body, _ := json.Marshal(serviceData)
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/v1/services", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	svcData, _ := createResp["data"].(map[string]interface{})
	slug, _ := svcData["slug"].(string)
	require.NotEmpty(t, slug)

	// 2. Create data
	dataPayload := map[string]interface{}{
		"title": "Test Item 1",
		"price": 19.99,
	}
	body, _ = json.Marshal(dataPayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "POST", "/api/v1/data/"+slug, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Body.String())

	var dataResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &dataResp)
	createdItem, _ := dataResp["data"].(map[string]interface{})
	assert.Equal(t, "Test Item 1", createdItem["title"])

	var itemID uint
	if idF, ok := createdItem["id"].(float64); ok {
		itemID = uint(idF)
	}

	// 3. Get Data by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("/api/v1/data/%s/%d", slug, itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &dataResp)
	fetchedItem, _ := dataResp["data"].(map[string]interface{})
	assert.Equal(t, "Test Item 1", fetchedItem["title"])

	// 4. Update Data
	updatePayload := map[string]interface{}{
		"title": "Updated Test Item 1",
		"price": 29.99,
	}
	body, _ = json.Marshal(updatePayload)
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "PUT", fmt.Sprintf("/api/v1/data/%s/%d", slug, itemID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &dataResp)
	updatedItem, _ := dataResp["data"].(map[string]interface{})
	assert.Equal(t, "Updated Test Item 1", updatedItem["title"])

	// 5. List Data
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", "/api/v1/data/"+slug+"?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &dataResp)
	items, ok := dataResp["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, items, 1)

	// 6. Delete Data
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "DELETE", fmt.Sprintf("/api/v1/data/%s/%d", slug, itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Ensure it's deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("/api/v1/data/%s/%d", slug, itemID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}
