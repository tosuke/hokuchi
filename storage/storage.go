package storage

import (
	"context"
	"errors"
	"io"
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
	ErrNotfound = errors.New("storage: not found")
	ErrExists   = errors.New("storage: already exists")
)
