package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// TokenEnvelope wraps both current Cassandra page state and previous token
type TokenEnvelope struct {
	State []byte `json:"state,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// EncodeToken converts a TokenEnvelope into a base64-encoded JSON string
func EncodeToken(state []byte, prev string) string {
	if len(state) == 0 && prev == "" {
		return ""
	}

	env := TokenEnvelope{
		State: state,
		Prev:  prev,
	}

	b, err := json.Marshal(env)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeToken decodes a base64 token into a TokenEnvelope
func DecodeToken(token string) (*TokenEnvelope, error) {
	if token == "" {
		return &TokenEnvelope{}, nil
	}

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 token: %w", err)
	}

	var env TokenEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("invalid token structure: %w", err)
	}

	return &env, nil
}
