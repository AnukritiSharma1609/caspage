package core_test

import (
	"errors"
	"testing"

	"github.com/AnukritiSharma1609/caspage/core"
)

// ---- Mock Implementations ----                       {}

type mockSession struct{}

func (m *mockSession) Query(q string, args ...interface{}) core.CassandraQuery {
	return &mockQuery{}
}

type mockQuery struct{}

func (q *mockQuery) PageSize(n int) core.CassandraQuery              { return q }
func (q *mockQuery) PageState(b []byte) core.CassandraQuery          { return q }
func (q *mockQuery) WithContext(ctx interface{}) core.CassandraQuery { return q }
func (q *mockQuery) Iter() core.CassandraIter                        { return &mockIter{} }

type mockIter struct {
	called bool
}

func (i *mockIter) MapScan(m map[string]interface{}) bool {
	if i.called {
		return false
	}
	i.called = true
	m["name"] = "Anukriti"
	m["count"] = 1
	return true
}

func (i *mockIter) PageState() []byte { return []byte("next_page") }
func (i *mockIter) Close() error      { return nil }

// ---- Tests ----

func TestPaginator_NextWithToken(t *testing.T) {
	p := &core.Paginator{
		Session:  &mockSession{}, // ✅ interface mock works now
		Query:    "SELECT * FROM users",
		PageSize: 10,
		Cache:    core.NewTokenCache(10),
		Opts:     core.Options{},
	}

	results, token, err := p.NextWithToken("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || token == "" {
		t.Errorf("unexpected result or token: %v, %s", results, token)
	}
}

func TestPaginator_Previous_TokenNotFound(t *testing.T) {
	p := &core.Paginator{Cache: core.NewTokenCache(5)}
	_, _, err := p.Previous("not_found")
	if !errors.Is(err, core.ErrNoPrevToken) {
		t.Errorf("expected ErrPreviousTokenNotFound, got %v", err)
	}
}

func TestPaginator_Next(t *testing.T) {
	p := &core.Paginator{
		Session:  &mockSession{},
		Query:    "SELECT * FROM users",
		PageSize: 10,
		Cache:    core.NewTokenCache(10),
		Opts:     core.Options{},
	}

	results, token, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 || token == "" {
		t.Errorf("expected valid results and token, got %+v, %s", results, token)
	}
}
