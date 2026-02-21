package core_test

import (
	"testing"

	"github.com/AnukritiSharma1609/caspage/core"
)

func TestMapTo_User(t *testing.T) {
	type User struct {
		ID   string `mapstructure:"user_id"`
		Name string `mapstructure:"name"`
	}

	// sample Cassandra-style results
	results := []map[string]interface{}{
		{"user_id": "123", "name": "Alice"},
		{"user_id": "456", "name": "Bob"},
	}

	// call exported helper
	typed, token, err := core.MapTo[User](results, "nextToken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(typed) != 2 {
		t.Fatalf("expected 2 users, got %d", len(typed))
	}
	if typed[0].Name != "Alice" || typed[1].ID != "456" {
		t.Fatalf("unexpected mapped values: %+v", typed)
	}
	if token != "nextToken" {
		t.Fatalf("expected token 'nextToken', got %q", token)
	}
}
