package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/tosuke/hokuchi/syncmap"
)

type fsStorage struct {
	dataDir string
	tempDir string

	running syncmap.M[string, *fsTx]
}
type fsTx struct {
	s        *fsStorage
	key      string
	tempFile *os.File
}

func NewFSStorage(dataDir string, tempDir string) Storage {
	return &fsStorage{
		dataDir: dataDir,
		tempDir: tempDir,
	}
}

var _ Storage = (*fsStorage)(nil)

func (s *fsStorage) Get(ctx context.Context, key string) (int64, io.ReadCloser, error) {
	path := s.pathForKey(key)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, &fs.PathError{}) || errors.Is(err, fs.ErrNotExist) {
			return 0, nil, ErrNotfound
		}
		return 0, nil, fmt.Errorf("failed to open storage data file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get stat for data file: %w", err)
	}
	size := stat.Size()

	return size, file, nil
}

func (s *fsStorage) Add(ctx context.Context, key string) (TxWriter, error) {
	if _, err := os.Stat(s.pathForKey(key)); err == nil {
		return nil, ErrExists
	}

	temp, err := os.CreateTemp(s.tempDir, "hokuchi-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for tx: %w", err)
	}
	tx := &fsTx{
		s: s,
		key: key,
		tempFile: temp,
	}

	if _, loaded := s.running.LoadOrStore(key, tx); loaded {
		temp.Close()
		os.Remove(temp.Name())
		return nil, ErrExists
	}

	return tx, nil
}

func (s *fsStorage) Close() error {
	var txs []*fsTx
	s.running.Range(func(_ string, tx *fsTx) bool {
		txs = append(txs, tx)
		return true
	})

	var errs []error
	for _, tx := range txs {
		if err := s.rollback(tx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *fsStorage) pathForKey(key string) string {
	return filepath.Join(s.dataDir, key)
}

func (s *fsStorage) close(key string) {
	s.running.Delete(key)
}

func (s *fsStorage) rollback(tx *fsTx) error {
	defer s.running.Delete(tx.key)
	tx.tempFile.Close()

	path := tx.tempFile.Name()
	if err := os.Remove(path); err != nil && !errors.Is(err, &fs.PathError{}) {
		return fmt.Errorf("faield to remove tmpfile: %w", err)
	}
	return nil
}

func (s *fsStorage) commit(tx *fsTx) error {
	defer func() {
		tx.s.close(tx.key)
		tx.tempFile.Close()
	}()
	tmpPath := tx.tempFile.Name()
	desiredPath := s.pathForKey(tx.key)

	if err := tx.tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync tmpfile: %w", err)
	}
	tx.tempFile.Close()

	if err := os.Rename(tmpPath, desiredPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}

var _ TxWriter = (*fsTx)(nil)

func (tx *fsTx) Write(b []byte) (int, error) {
	return tx.tempFile.Write(b)
}

func (tx *fsTx) Rollback() error {
	return tx.s.rollback(tx)
}

func (tx *fsTx) Commit(ctx context.Context) error {
	return tx.s.commit(tx)
}
