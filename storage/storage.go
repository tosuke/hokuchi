package storage

import (
	"context"
	"io"

	"braces.dev/errtrace"
)

type TxWriter interface {
	io.Writer

	Rollback() error
	Commit(ctx context.Context) error
}

type Storage interface {
	Get(ctx context.Context, key string) (size int64, r io.ReadCloser, err error)
	Add(ctx context.Context, key string) (TxWriter, error)
	Close() error
}

var (
	ErrNotfound = errtrace.New("storage: not found")
	ErrExists   = errtrace.New("storage: already exists")
)
