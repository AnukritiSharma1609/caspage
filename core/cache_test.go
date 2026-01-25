package core_test

import (
	"testing"

	"github.com/AnukritiSharma1609/caspage/core"
)

func TestTokenCache_AddAndSize(t *testing.T) {
	cache := core.NewTokenCache(3)

	cache.Add("t1")
	cache.Add("t2")

	if got := cache.Size(); got != 2 {
		t.Errorf("expected 2 tokens, got %d", got)
	}

	cache.Add("t3")
	cache.Add("t4") // should evict t1 if limit=3

	if got := cache.Size(); got != 3 {
		t.Errorf("expected cache size 3, got %d", got)
	}
}

func TestTokenCache_Previous(t *testing.T) {
	cache := core.NewTokenCache(5)
	cache.Add("t1")
	cache.Add("t2")
	cache.Add("t3")

	prev, ok := cache.Previous("t3")
	if !ok || prev != "t2" {
		t.Errorf("expected previous token t2, got %v (ok=%v)", prev, ok)
	}

	_, ok = cache.Previous("unknown")
	if ok {
		t.Error("expected not found for unknown token")
	}
}

func TestTokenCache_Last(t *testing.T) {
	cache := core.NewTokenCache(3)
	cache.Add("t1")
	cache.Add("t2")

	if last := cache.Last(); last != "t2" {
		t.Errorf("expected last token 't2', got %s", last)
	}
}

