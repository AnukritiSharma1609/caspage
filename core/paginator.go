package core

import (
	"fmt"

	"github.com/gocql/gocql"
)

type Paginator struct {
	session  *gocql.Session
	query    string
	pageSize int
	cache    *TokenCache
}

// NewPaginator now initializes a cache too
func NewPaginator(session *gocql.Session, query string, opts Options) *Paginator {
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	return &Paginator{
		session:  session,
		query:    query,
		pageSize: pageSize,
		cache:    NewTokenCache(10), // keep history of 10 tokens
	}
}

func (p *Paginator) NextWithToken(token string) ([]map[string]interface{}, string, error) {
	results, nextToken, err := p.fetchWithToken(token)
	if err != nil {
		return nil, "", err
	}
	p.cache.Add(nextToken)
	return results, nextToken, nil
}

// Internal shared logic
func (p *Paginator) fetchWithToken(token string) ([]map[string]interface{}, string, error) {
	var state []byte
	if token != "" {
		s, err := DecodeToken(token)
		if err != nil {
			return nil, "", err
		}
		state = s
	}

	q := p.session.Query(p.query).PageSize(p.pageSize)
	if len(state) > 0 {
		q = q.PageState(state)
	}

	iter := q.Iter()
	results := []map[string]interface{}{}
	row := map[string]interface{}{}
	for iter.MapScan(row) {
		results = append(results, row)
		row = map[string]interface{}{}
	}

	next := iter.PageState()
	if err := iter.Close(); err != nil {
		return nil, "", err
	}

	return results, EncodeToken(next), nil
}

// Stateful convenience wrapper (calls the stateless version internally)
func (p *Paginator) Next() ([]map[string]interface{}, string, error) {
	// empty token = start from beginning
	return p.NextWithToken("")
}

func (p *Paginator) Previous(currentToken string) ([]map[string]interface{}, string, error) {
	prevToken, ok := p.cache.Previous(currentToken)
	if !ok {
		return nil, "", fmt.Errorf("no previous token found for %q", currentToken)
	}

	results, nextToken, err := p.fetchWithToken(prevToken)
	if err != nil {
		return nil, "", err
	}

	return results, nextToken, nil
}
