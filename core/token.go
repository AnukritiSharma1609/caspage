package core

import (
	"encoding/base64"
	"fmt"
)

// EncodeToken converts the raw Cassandra page state into a Base64 URL-safe token.
// This can be safely used in URLs or API responses.
func EncodeToken(state []byte) string {
	if len(state) == 0 {
		return ""
	}
	return base64.URLEncoding.EncodeToString(state)
}

// DecodeToken converts a Base64 token back into Cassandra-compatible page state bytes.
func DecodeToken(token string) ([]byte, error) {
	if token == "" {
		return nil, nil
	}

	state, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination token: %w", err)
	}
	return state, nil
}
