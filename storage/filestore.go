package storage

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gofrs/flock"
)

var (
	ErrLocked = errors.New("storage locked")
)

type FileStore struct {
	db   string
	lock *flock.Flock
}

func NewFileStore(ctx context.Context, db string, lockFile string) (*FileStore, error) {
	f := flock.New(lockFile)
	_, err := f.TryRLockContext(ctx, time.Second)
	if err != nil {
		return nil, err
	}

	if f.Locked() {
		return nil, fmt.Errorf("%w by %s", ErrLocked, lockFile)
	}

	if err := f.Lock(); err != nil {
		return nil, fmt.Errorf("failed to lock storage: %w", err)
	}

	return &FileStore{db: db, lock: f}, nil
}

func (s FileStore) Read() ([]byte, error) {
	b, err := ioutil.ReadFile(s.db)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (s FileStore) Write(data []byte) error {
	return ioutil.WriteFile(s.db, data, 0600)
}

func (s FileStore) Close() error {
	return s.lock.Unlock()
}
