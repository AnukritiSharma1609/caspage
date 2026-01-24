package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type Paginator struct {
	session  *gocql.Session
	query    string
	pageSize int
	cache    *TokenCache
	opts     Options
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
		opts:     opts,
	}
}

func (p *Paginator) NextWithToken(token string) ([]map[string]interface{}, string, error) {
	results, nextToken, err := p.fetchWithToken(token)
	if err != nil {
		return nil, "", err
	}
	p.cache.Add(nextToken)

	if p.opts.Metrics != nil {
		p.opts.Metrics.ObserveActiveTokens(p.cache.Size())
	}
	return results, nextToken, nil
}

// fetchWithToken executes the paginated Cassandra query and returns results and the next page token.
func (p *Paginator) fetchWithToken(token string) ([]map[string]interface{}, string, error) {
	var state []byte
	var err error

	// 1️⃣ Decode the page token if provided
	if token != "" {
		state, err = DecodeToken(token)
		if err != nil {
			p.log("invalid_token", map[string]interface{}{
				"token": token,
				"error": err.Error(),
			})
			if p.opts.Metrics != nil {
				p.opts.Metrics.ObserveError(ErrInvalidToken)
			}
			return nil, "", ErrInvalidToken
		}
	}

	// 2️⃣ Build the query string dynamically (columns + filters)
	queryStr := p.query

	// Replace "*" with selected columns if provided
	if len(p.opts.Columns) > 0 {
		queryStr = strings.Replace(queryStr, "*", strings.Join(p.opts.Columns, ", "), 1)
	}

	// Use helper to build WHERE/AND clauses dynamically
	queryStr, bindValues := buildQueryWithFilters(queryStr, p.opts.Filters)

	// Initialize query with optional bound values
	q := p.session.Query(queryStr, bindValues...).PageSize(p.pageSize)

	// 3️⃣ Apply page state if resuming from token
	if len(state) > 0 {
		q = q.PageState(state)
	}

	// 4️⃣ Apply context if present (for timeout/cancellation)
	if p.opts.Context != nil {
		q = q.WithContext(p.opts.Context)
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
		if count >= p.pageSize {
			break
		}
	}

	duration := time.Since(start)
	next := iter.PageState()

	// 5️⃣ Handle query errors
	if err := iter.Close(); err != nil {
		p.log("query_failed", map[string]interface{}{
			"query":     queryStr,
			"error":     err.Error(),
			"filters":   p.opts.Filters,
			"duration":  duration.Milliseconds(),
			"page_size": p.pageSize,
		})
		if p.opts.Metrics != nil {
			p.opts.Metrics.ObserveError(ErrQueryFailed)
		}
		return nil, "", ErrQueryFailed
	}

	// 6️⃣ Log success
	p.log("page_fetched", map[string]interface{}{
		"rows_fetched":  len(results),
		"next_token":    len(next) > 0,
		"duration_ms":   duration.Milliseconds(),
		"query_filters": p.opts.Filters,
	})

	// 7️⃣ Record metrics
	if p.opts.Metrics != nil {
		p.opts.Metrics.ObservePageFetch(len(results), duration)
	}

	return results, EncodeToken(next), nil
}

// log safely invokes the optional logger hook.
func (p *Paginator) log(event string, data map[string]interface{}) {
	if p.opts.Logger != nil {
		p.opts.Logger(event, data)
	}
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

// buildQueryWithFilters dynamically appends WHERE/AND clauses to the base query
// using the provided filters map. Supports operators like =, >, <, >=, <=, and IN.
// Example:
//
//	baseQuery: "SELECT * FROM users"
//	filters: map[string]interface{}{"age >": 25, "region IN": []string{"US", "CA"}}
//
// Result:
//
//	queryStr: "SELECT * FROM users WHERE age > ? AND region IN (?, ?)"
//	values:   [25, "US", "CA"]
func buildQueryWithFilters(baseQuery string, filters map[string]interface{}) (string, []interface{}) {
	if len(filters) == 0 {
		return baseQuery, nil
	}

	whereClauses := []string{}
	values := []interface{}{}

	for k, v := range filters {
		key := strings.TrimSpace(k)
		operator := "=" // default

		// Extract operator if present
		for _, op := range []string{">=", "<=", ">", "<", "IN"} {
			if strings.HasSuffix(strings.ToUpper(key), op) {
				operator = op
				key = strings.TrimSpace(strings.TrimSuffix(key, op))
				break
			}
		}

		switch strings.ToUpper(operator) {
		case "IN":
			// Handle slice or array values for IN
			valSlice, ok := anyToSlice(v)
			if !ok || len(valSlice) == 0 {
				continue // skip invalid or empty IN filters
			}
			placeholders := make([]string, len(valSlice))
			for i := range valSlice {
				placeholders[i] = "?"
			}
			whereClauses = append(whereClauses, key+" IN ("+strings.Join(placeholders, ", ")+")")
			values = append(values, valSlice...)

		default:
			// Basic comparison: =, >, <, >=, <=
			whereClauses = append(whereClauses, key+" "+operator+" ?")
			values = append(values, v)
		}
	}

	queryLower := strings.ToLower(baseQuery)
	if strings.Contains(queryLower, "where") {
		baseQuery += " AND " + strings.Join(whereClauses, " AND ")
	} else {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	return baseQuery, values
}

// anyToSlice converts any slice/array into []interface{} for binding.
// Returns (nil, false) if the input is not slice-like.
func anyToSlice(v interface{}) ([]interface{}, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	out := make([]interface{}, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}
