package core

import (
	"context"

	"github.com/gocql/gocql"
)

// RealSession wraps a *gocql.Session to implement CassandraSession.
type RealSession struct {
	*gocql.Session
}

func (s *RealSession) Query(stmt string, values ...interface{}) CassandraQuery {
	return &RealQuery{s.Session.Query(stmt, values...)}
}

// RealQuery wraps a *gocql.Query.
type RealQuery struct {
	*gocql.Query
}

func (q *RealQuery) PageSize(n int) CassandraQuery {
	q.Query = q.Query.PageSize(n)
	return q
}

func (q *RealQuery) PageState(b []byte) CassandraQuery {
	q.Query = q.Query.PageState(b)
	return q
}

func (q *RealQuery) WithContext(ctx interface{}) CassandraQuery {
	if c, ok := ctx.(context.Context); ok {
		q.Query = q.Query.WithContext(c)
	}
	return q
}

func (q *RealQuery) Iter() CassandraIter {
	return &RealIter{q.Query.Iter()}
}

// RealIter wraps a *gocql.Iter.
type RealIter struct {
	*gocql.Iter
}

func (i *RealIter) MapScan(m map[string]interface{}) bool { return i.Iter.MapScan(m) }
func (i *RealIter) PageState() []byte                    { return i.Iter.PageState() }
func (i *RealIter) Close() error                         { return i.Iter.Close() }
