package core

import "errors"

var (
	ErrInvalidToken = errors.New("invalid page token")
	ErrNoPrevToken  = errors.New("no previous token found")
	ErrQueryFailed  = errors.New("failed to execute Cassandra query")
)
