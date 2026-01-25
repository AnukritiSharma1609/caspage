package core

// CassandraSession defines a minimal interface that both real and mock sessions can implement.
type CassandraSession interface {
	Query(string, ...interface{}) CassandraQuery
}

type CassandraQuery interface {
	PageSize(int) CassandraQuery
	PageState([]byte) CassandraQuery
	WithContext(interface{}) CassandraQuery
	Iter() CassandraIter
}

type CassandraIter interface {
	MapScan(map[string]interface{}) bool
	PageState() []byte
	Close() error
}
