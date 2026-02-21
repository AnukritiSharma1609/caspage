package core

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// NextAs returns typed results (e.g., []User) instead of []map[string]interface{}.
func NextAs[T any](p *Paginator) ([]T, string, error) {
	results, token, err := p.Next()
	if err != nil {
		return nil, "", err
	}
	return MapTo[T](results, token)
}

// NextWithTokenAs returns typed results (e.g., []User) instead of []map[string]interface{}.
func NextWithTokenAs[T any](p *Paginator, token string) ([]T, string, error) {
	results, nextToken, err := p.NextWithToken(token)
	if err != nil {
		return nil, "", err
	}
	return MapTo[T](results, nextToken)
}

// mapTo decodes a slice of map[string]interface{} into a typed slice using struct tags.
func MapTo[T any](input []map[string]interface{}, token string) ([]T, string, error) {
	var typed []T

	if len(input) == 0 {
		return typed, token, nil
	}

	for _, m := range input {
		var t T
		if err := mapstructure.Decode(m, &t); err != nil {
			return nil, "", fmt.Errorf("mapstructure decode failed: %w", err)
		}
		typed = append(typed, t)
	}

	return typed, token, nil
}
