package graphql

import (
	"fmt"
	"strconv"

	"cms-backend/internal/contextutils"
	"cms-backend/internal/services"
	"github.com/graphql-go/graphql"
)

// List resolve list data
func (e *Engine) resolveList(slug string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		page, _ := p.Args["page"].(int)
		if page < 1 {
			page = 1
		}

		limit, _ := p.Args["limit"].(int)
		if limit < 1 || limit > 100 {
			limit = 20
		}

		req := &services.ListDataRequest{
			Page:     page,
			PageSize: limit,
		}

		rows, _, err := e.dataSvc.ListData(p.Context, slug, req)
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
}

// Get resolve single data by ID
func (e *Engine) resolveGet(slug string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		idStr, ok := p.Args["id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid id")
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id format")
		}

		record, err := e.dataSvc.GetData(p.Context, slug, uint(id), nil)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
}

// Create resolve create data
func (e *Engine) resolveCreate(slug string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		data, ok := p.Args["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid data payload")
		}

		// Currently, this bypasses the standard Gin auth extraction.
		// For a full implementation, the Context should contain the UserID.
		// Since it's dynamic generation, we extract userID from context if set.
		var userID uint = 1 // default for now, can be extracted from p.Context if set
		ctxUserID := p.Context.Value(contextutils.UserIDKey)
		if val, ok := ctxUserID.(uint); ok {
			userID = val
		}

		record, err := e.dataSvc.CreateData(p.Context, slug, data, userID)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
}

// Update resolve update data
func (e *Engine) resolveUpdate(slug string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		idStr, ok := p.Args["id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid id")
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id format")
		}

		data, ok := p.Args["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid data payload")
		}

		var userID uint = 1
		ctxUserID := p.Context.Value(contextutils.UserIDKey)
		if val, ok := ctxUserID.(uint); ok {
			userID = val
		}

		record, err := e.dataSvc.UpdateData(p.Context, slug, uint(id), data, userID)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
}

// Delete resolve delete data
func (e *Engine) resolveDelete(slug string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		idStr, ok := p.Args["id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid id")
		}

		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id format")
		}

		err = e.dataSvc.DeleteData(p.Context, slug, uint(id))
		if err != nil {
			return false, err
		}
		return true, nil
	}
}
