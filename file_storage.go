package safsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileStorage struct {
	path string
}

type jsonSession struct {
	ID        int64         `json:"id"`
	Token     string        `json:"token"`
	UserID    int64         `json:"user_id"`
	CreatedAt time.Time     `json:"created"`
	UpdatedAt time.Time     `json:"updated"`
	TTL       time.Duration `json:"ttl"`
}

func NewFileStorage(path string) (*FileStorage, error) {
	if info, err := os.Stat(path); err != nil {
		return nil, err
	} else {
		if !info.IsDir() {
			return nil, errors.New("path to not a dir")
		}
	}

	if ok, err := isWritablePath(path); err != nil || !ok {
		return nil, errors.New("path directory is not writable")
	}

	return &FileStorage{path: path}, nil
}

func (f *FileStorage) Close() error {
	return nil
}

func (f *FileStorage) InsertSession(session *Session) error {
	fname := nameFromToken(session.Token())
	file, err := os.Create(filepath.Join(f.path, fname))
	if err != nil {
		return err
	}
	defer file.Close()

	obj := jsonSession{
		ID:        session.ID(),
		Token:     session.Token(),
		UserID:    session.UserID(),
		CreatedAt: session.CreatedAt(),
		UpdatedAt: session.UpdatedAt(),
		TTL:       session.TTL(),
	}

	if err := json.NewEncoder(file).Encode(obj); err != nil {
		return fmt.Errorf("failed to encode session: %v", err)
	}

	return nil
}

func (f *FileStorage) FindSession(token string) (*Session, error) {
	fname := nameFromToken(token)
	file, err := os.Open(filepath.Join(f.path, fname))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoSession
		}
		return nil, fmt.Errorf("failed to open session file: %v", err)
	}
	defer file.Close()

	var v jsonSession
	err = json.NewDecoder(file).Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json session: %v", err)
	}

	return NewSession(v.ID, v.Token, v.UserID, v.CreatedAt, v.UpdatedAt, v.TTL), nil
}

func (f *FileStorage) RemoveSession(token string) error {
	fname := nameFromToken(token)
	if err := os.Remove(filepath.Join(f.path, fname)); err != nil {
		return fmt.Errorf("failed to remove session: %v", err)
	}
	return nil
}

func (fs *FileStorage) Each(f func(session *Session)) {
	entries, err := os.ReadDir(fs.path)
	if err != nil {
		fmt.Println(err)
		return // TODO: do something?
	}

	for _, v := range entries {
		if filepath.Ext(v.Name()) == ".json" {
			file, err := os.Open(filepath.Join(fs.path, v.Name()))
			if err != nil {
				continue // TODO: log something?
			}
			defer file.Close()
			var v jsonSession
			if err = json.NewDecoder(file).Decode(&v); err != nil {
				continue // TODO: log something?
			}
			session := NewSession(v.ID, v.Token, v.UserID, v.CreatedAt, v.UpdatedAt, v.TTL)
			f(session)
		}
	}
}

func isWritablePath(path string) (bool, error) {
	tmpFile := "tmpfile"

	file, err := os.CreateTemp(path, tmpFile)
	if err != nil {
		return false, err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	return true, nil
}

func nameFromToken(token string) string {
	return token[:12] + ".json"
}
