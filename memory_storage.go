package safsm

import (
	"errors"
	"sync"
)

type MemoryStorage struct {
	sessions map[string]*Session
	counter  int64
	m        *sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		sessions: map[string]*Session{},
		counter:  0,
		m:        &sync.Mutex{},
	}
}

func (m *MemoryStorage) Close() error {
	m.m.Lock()
	m.sessions = nil
	m.m.Unlock()
	return nil
}

func (m *MemoryStorage) InsertSession(session *Session) error {
	if m.sessions == nil {
		return errors.New("session storage is closed")
	}
	m.m.Lock()
	session.id = m.counter
	m.counter++
	m.sessions[session.Token()] = session
	m.m.Unlock()
	return nil
}

func (m *MemoryStorage) FindSession(token string) (*Session, error) {
	if m.sessions == nil {
		return nil, errors.New("session storage is closed")
	}

	if session, ok := m.sessions[token]; ok {
		return session, nil
	}
	return nil, ErrNoSession
}

func (m *MemoryStorage) RemoveSession(token string) error {
	m.m.Lock()
	delete(m.sessions, token)
	m.m.Unlock()
	return nil
}

func (m *MemoryStorage) Each(f func(session *Session)) {
	for _, v := range m.sessions {
		f(v.Copy())
	}
}
