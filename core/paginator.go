package core

import (
	"fmt"

	"github.com/gocql/gocql"
)

// Paginator encapsulates Cassandra pagination logic.
// It abstracts away PageState handling and exposes a clean Next() API.
type Paginator struct {
	session   *gocql.Session
	query     string
	pageSize  int
	pageState []byte
}

// NewPaginator initializes a paginator for a given query and session.
func NewPaginator(session *gocql.Session, query string, opts Options) *Paginator {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 100 // default fallback
	}

	return &Paginator{
		session:  session,
		query:    query,
		pageSize: pageSize,
	}
}

// Next fetches the next page of results from Cassandra.
// It returns the result slice, a new PageState for the next page, and an error (if any).
func (p *Paginator) Next() ([]map[string]interface{}, []byte, error) {
	q := p.session.Query(p.query).PageSize(p.pageSize)
	if p.pageState != nil {
		q = q.PageState(p.pageState)
	}

	iter := q.Iter()

	// Read all rows from this page.
	results := []map[string]interface{}{}
	m := map[string]interface{}{}
	for iter.MapScan(m) {
		results = append(results, m)
		m = map[string]interface{}{}
	}

	nextPageState := iter.PageState()
	err := iter.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("pagination query failed: %w", err)
	}

	// Store state for chaining calls (if used directly).
	p.pageState = nextPageState

	return results, nextPageState, nil
}
