package core

import "sync"

// TokenCache stores the last N tokens for backward navigation.
// It's a simple, thread-safe in-memory list.
type TokenCache struct {
	mu     sync.Mutex
	tokens []string
	limit  int
}

// NewTokenCache creates a cache that can hold up to `limit` tokens.
func NewTokenCache(limit int) *TokenCache {
	if limit <= 0 {
		limit = 5 // default to 5 pages of history
	}
	return &TokenCache{
		tokens: make([]string, 0, limit),
		limit:  limit,
	}
}

// Add stores a new token, evicting the oldest if needed.
func (c *TokenCache) Add(token string) {
	if token == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Prevent duplicates if the token already exists
	for _, t := range c.tokens {
		if t == token {
			return
		}
	}

	if len(c.tokens) == c.limit {
		// Remove oldest
		c.tokens = c.tokens[1:]
	}
	c.tokens = append(c.tokens, token)
}

// Previous returns the token before the given one, if available.
func (c *TokenCache) Previous(current string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, t := range c.tokens {
		if t == current && i > 0 {
			return c.tokens[i-1], true
		}
	}
	return "", false
}

// Last returns the most recent token (for quick access).
func (c *TokenCache) Last() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.tokens) == 0 {
		return ""
	}
	return c.tokens[len(c.tokens)-1]
}
