package core

import (
	"reflect"
	"strings"
)

// buildQueryWithFilters dynamically appends WHERE/AND clauses to the base query
// using the provided filters map. Supports operators like =, >, <, >=, <=, and IN.
// Example:
//
//	baseQuery: "SELECT * FROM users"
//	filters: map[string]interface{}{"age >": 25, "region IN": []string{"US", "CA"}}
//
// Result:
//
//	queryStr: "SELECT * FROM users WHERE age > ? AND region IN (?, ?)"
//	values:   [25, "US", "CA"]
func buildQueryWithFilters(baseQuery string, filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return baseQuery, nil
	}

	whereClauses := []string{}
	values := []interface{}{}

	for k, v := range filters {
		key := strings.TrimSpace(k)
		operator := "=" // default

		// Extract operator if present
		for _, op := range []string{">=", "<=", ">", "<", "IN"} {
			if strings.HasSuffix(strings.ToUpper(key), op) {
				operator = op
				key = strings.TrimSpace(strings.TrimSuffix(key, op))
				break
			}
		}

		switch strings.ToUpper(operator) {
		case "IN":
			// Handle slice or array values for IN
			valSlice, ok := anyToSlice(v)
			if !ok || len(valSlice) == 0 {
				continue // skip invalid or empty IN filters
			}
			placeholders := make([]string, len(valSlice))
			for i := range valSlice {
				placeholders[i] = "?"
			}
			whereClauses = append(whereClauses, key+" IN ("+strings.Join(placeholders, ", ")+")")
			values = append(values, valSlice...)

		default:
			// Basic comparison: =, >, <, >=, <=
			whereClauses = append(whereClauses, key+" "+operator+" ?")
			values = append(values, v)
		}
	}

	queryLower := strings.ToLower(baseQuery)
	if strings.Contains(queryLower, "where") {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	} else {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return baseQuery, values
}

// anyToSlice converts any slice/array into []interface{} for binding.
// Returns (nil, false) if the input is not slice-like.
func anyToSlice(v interface{}) ([]interface{}, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	out := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}
