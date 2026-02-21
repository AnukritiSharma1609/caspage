package core_test

import (
	"testing"

	"github.com/AnukritiSharma1609/caspage/core"
)

func TestEncodeDecodeToken(t *testing.T) {
	data := []byte("sample_page_state")
	prev := "prev_token_value"

	// 1️⃣ Encode token with current state + previous token
	encoded := core.EncodeToken(data, prev)
	if encoded == "" {
		t.Fatal("expected encoded token, got empty string")
	}

	// 2️⃣ Decode back to TokenEnvelope
	env, err := core.DecodeToken(encoded)
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}

	// 3️⃣ Validate decoded fields
	if string(env.State) != string(data) {
		t.Fatalf("expected state %q, got %q", data, env.State)
	}

	if env.Prev != prev {
		t.Fatalf("expected prev %q, got %q", prev, env.Prev)
	}

	// 4️⃣ Invalid token should return error
	_, err = core.DecodeToken("invalid-base64!")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}

	// 5️⃣ Empty token should decode safely to an empty envelope
	env, err = core.DecodeToken("")
	if err != nil {
		t.Fatalf("unexpected error decoding empty token: %v", err)
	}
	if len(env.State) != 0 || env.Prev != "" {
		t.Fatalf("expected empty envelope, got %+v", env)
	}
}
