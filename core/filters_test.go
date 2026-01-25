package core

import (
	"strings"
	"testing"
)

func TestBuildQueryWithFilters(t *testing.T) {
	query := "SELECT * FROM users"
	filters := map[string]interface{}{
		"age >":     30,
		"region IN": []string{"US", "CA"},
		"active":    true,
	}

	gotQuery, values := buildQueryWithFilters(query, filters)

	if !strings.Contains(strings.ToLower(gotQuery), "where") {
		t.Errorf("expected WHERE clause in query: %s", gotQuery)
	}

	if len(values) != 4 {
		t.Errorf("expected 4 bound values, got %d", len(values))
	}
}

func TestBuildQueryWithFilters_Empty(t *testing.T) {
	query, vals := buildQueryWithFilters("SELECT * FROM users", nil)
	if query != "SELECT * FROM users" {
		t.Errorf("expected original query unchanged, got %s", query)
	}
	if len(vals) != 0 {
		t.Errorf("expected no values, got %v", vals)
	}
}
