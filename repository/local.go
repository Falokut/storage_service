package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Falokut/storage_service/domain"
	"github.com/Falokut/storage_service/entity"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type lockEntry struct {
	mu    sync.Mutex
	count int
}

type LocalImageStorage struct {
	mu        sync.Mutex
	fileLocks map[string]*lockEntry
	basePath  string
}

func NewLocalStorage(baseStoragePath string) *LocalImageStorage {
	return &LocalImageStorage{
		basePath:  baseStoragePath,
		fileLocks: make(map[string]*lockEntry),
	}
}

func (s *LocalImageStorage) Shutdown() {}

func (s *LocalImageStorage) UploadFile(ctx context.Context, req entity.File) error {
	path := filepath.Clean(fmt.Sprintf("%s/%s/%s", s.basePath, req.Category, req.Filename))
	lock := s.lockFile(path)
	lock.Lock()
	defer func() {
		lock.Unlock()
		s.unlockFile(path)
	}()

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return errors.WithMessage(err, "make directory")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0660)
	if err != nil {
		return errors.WithMessage(err, "create file")
	}

	content, err := msgpack.Marshal(req)
	if err != nil {
		return errors.WithMessage(err, "marshal file")
	}
	_, err = f.Write(content)
	if err != nil {
		return errors.WithMessage(err, "write into file")
	}
	return nil
}

func (s *LocalImageStorage) GetFile(ctx context.Context, filename string, category string) (*entity.File, error) {
	path := filepath.Clean(fmt.Sprintf("%s/%s/%s", s.basePath, category, filename))
	lock := s.lockFile(path)
	lock.Lock()
	defer func() {
		lock.Unlock()
		s.unlockFile(path)
	}()

	fileBody, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil, domain.ErrFileNotFound
	case err != nil:
		return nil, errors.WithMessage(err, "read file")
	}

	file := &entity.File{}
	err = msgpack.Unmarshal(fileBody, file)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal file")
	}
	return file, nil
}

func (s *LocalImageStorage) IsFileExist(ctx context.Context, filename string, category string) (bool, error) {
	path := filepath.Clean(fmt.Sprintf("%s/%s/%s", s.basePath, category, filename))
	lock := s.lockFile(path)
	lock.Lock()
	defer func() {
		lock.Unlock()
		s.unlockFile(path)
	}()

	_, err := os.Stat(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		return false, errors.WithMessage(err, "os stat")
	default:
		return true, nil
	}
}

func (s *LocalImageStorage) DeleteFile(ctx context.Context, filename string, category string) error {
	path := filepath.Clean(fmt.Sprintf("%s/%s/%s", s.basePath, category, filename))
	lock := s.lockFile(path)
	lock.Lock()
	defer func() {
		lock.Unlock()
		s.unlockFile(path)
	}()

	err := os.Remove(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return domain.ErrFileNotFound
	default:
		return errors.WithMessage(err, "remove file")
	}
}

func (s *LocalImageStorage) lockFile(path string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.fileLocks[path]; exists {
		entry.count++
		return &entry.mu
	}

	entry := &lockEntry{count: 1}
	s.fileLocks[path] = entry
	return &entry.mu
}

func (s *LocalImageStorage) unlockFile(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.fileLocks[path]; exists {
		entry.count--
		if entry.count == 0 {
			delete(s.fileLocks, path)
		}
	}
}
