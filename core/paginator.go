package core

import (
	"fmt"
	"strings"
	"time"
)

type Paginator struct {
	Session  CassandraSession
	Query    string
	PageSize int
	Opts     Options
}

// NewPaginator now initializes a cache too
func NewPaginator(session CassandraSession, query string, Opts Options) *Paginator {
	pageSize := Opts.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	return &Paginator{
		Session:  session,
		Query:    query,
		PageSize: pageSize,
		Opts:     Opts,
	}
}

func (p *Paginator) NextWithToken(token string) ([]map[string]interface{}, string, error) {
	results, nextToken, err := p.fetchWithToken(token)
	if err != nil {
		return nil, "", err
	}

	return results, nextToken, nil
}

// fetchWithToken executes the paginated Cassandra query and returns results and the next page token.
func (p *Paginator) fetchWithToken(token string) ([]map[string]interface{}, string, error) {
	var env *TokenEnvelope
	var err error
	var prev string

	// 1️⃣ Decode the page token if provided
	if token != "" {
		env, err = DecodeToken(token)
		if err != nil {
			p.log("invalid_token", map[string]interface{}{
				"token": token,
				"error": err.Error(),
			})
			if p.Opts.Metrics != nil {
				p.Opts.Metrics.ObserveError(ErrInvalidToken)
			}
			return nil, "", ErrInvalidToken
		}
		prev = token // current token becomes "prev" for the next page
	} else {
		env = &TokenEnvelope{}
	}

	// 2️⃣ Build the query string dynamically (columns + filters)
	queryStr := p.Query

	// Replace "*" with selected columns if provided
	if len(p.Opts.Columns) > 0 {
		queryStr = strings.Replace(queryStr, "*", strings.Join(p.Opts.Columns, ", "), 1)
	}

	// Use helper to build WHERE/AND clauses dynamically
	queryStr, bindValues := buildQueryWithFilters(queryStr, p.Opts.Filters)

	// Initialize query with optional bound values
	q := p.Session.Query(queryStr, bindValues...).PageSize(p.PageSize)

	// 3️⃣ Apply page state if resuming from token
	if len(env.State) > 0 {
		q = q.PageState(env.State)
	}

	// 4️⃣ Apply context if present (for timeout/cancellation)
	if p.Opts.Context != nil {
		q = q.WithContext(p.Opts.Context)
	}

	start := time.Now()
	iter := q.Iter()

	results := []map[string]interface{}{}
	row := map[string]interface{}{}
	count := 0

	for iter.MapScan(row) {
		results = append(results, row)
		row = map[string]interface{}{}
		count++
		if count >= p.PageSize {
			break
		}
	}

	duration := time.Since(start)
	nextState := iter.PageState()

	// 5️⃣ Handle query errors
	if err := iter.Close(); err != nil {
		p.log("query_failed", map[string]interface{}{
			"query":     queryStr,
			"error":     err.Error(),
			"filters":   p.Opts.Filters,
			"duration":  duration.Milliseconds(),
			"page_size": p.PageSize,
		})
		if p.Opts.Metrics != nil {
			p.Opts.Metrics.ObserveError(ErrQueryFailed)
		}
		return nil, "", ErrQueryFailed
	}

	// 6️⃣ Log success
	p.log("page_fetched", map[string]interface{}{
		"rows_fetched":  len(results),
		"next_token":    len(nextState) > 0,
		"duration_ms":   duration.Milliseconds(),
		"query_filters": p.Opts.Filters,
	})

	// 7️⃣ Record metrics
	if p.Opts.Metrics != nil {
		p.Opts.Metrics.ObservePageFetch(len(results), duration)
	}

	// 8️⃣ Encode next token with embedded "prev"
	nextToken := EncodeToken(nextState, prev)

	return results, nextToken, nil
}

// log safely invokes the optional logger hook.
func (p *Paginator) log(event string, data map[string]interface{}) {
	if p.Opts.Logger != nil {
		p.Opts.Logger(event, data)
	}
}

// Stateful convenience wrapper (calls the stateless version internally)
func (p *Paginator) Next() ([]map[string]interface{}, string, error) {
	// empty token = start from beginning
	return p.NextWithToken("")
}

// Previous navigates one page backward using the embedded "prev" token.
// It decodes the given token, extracts the previous token inside it, and fetches that page.
func (p *Paginator) Previous(token string) ([]map[string]interface{}, string, error) {
	env, err := DecodeToken(token)
	if err != nil {
		return nil, "", err
	}

	if env.Prev == "" {
		return nil, "", fmt.Errorf("no previous page available")
	}

	// Directly fetch the previous page using the embedded previous token.
	return p.fetchWithToken(env.Prev)
}
