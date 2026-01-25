package core_test

import (
	"testing"

	"github.com/AnukritiSharma1609/caspage/core"
)

func TestEncodeDecodeToken(t *testing.T) {
	data := []byte("sample_page_state")

	encoded := core.EncodeToken(data)
	if encoded == "" {
		t.Fatal("expected encoded token, got empty string")
	}

	decoded, err := core.DecodeToken(encoded)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	if string(decoded) != string(data) {
		t.Fatalf("expected %q, got %q", data, decoded)
	}

	// Invalid token
	_, err = core.DecodeToken("invalid-base64!")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
