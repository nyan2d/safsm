package safsm

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Safsm struct {
	storage Storage
}

func New(storage Storage) *Safsm {
	return &Safsm{
		storage: storage,
	}
}

func (s *Safsm) Close() error {
	return s.storage.Close()
}

func (s *Safsm) CreateSession(userID int64, ttl time.Duration) *Session {
	return &Session{
		id:        -1,
		token:     generateToken(),
		userID:    userID,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		ttl:       ttl,
		sm:        s,
	}
}

func (s *Safsm) ReadSession(r *http.Request) (*Session, error) {
	token, err := r.Cookie("auth-token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrNoSession
		}
		return nil, fmt.Errorf("failed to read session cookie: %v", err)
	}

	return s.storage.FindSession(token.Value)
}

func (s *Safsm) ReadBearerSession(r *http.Request) (*Session, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, ErrNoBearerToken
	}
	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, ErrNoBearerToken
	}

	return s.storage.FindSession(parts[1])
}

func (s *Safsm) RemoveInvalids() {
	s.storage.Each(func(session *Session) {
		if !session.Valid() {
			session.AssignTo(s).Remove()
		}
	})
}
