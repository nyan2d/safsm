package safsm

import (
	"net/http"
	"time"
)

type Session struct {
	id        int64
	token     string
	userID    int64
	createdAt time.Time
	updatedAt time.Time
	ttl       time.Duration

	sm *Safsm
}

func NewSession(id int64, token string, userID int64, createdAt, updatedAt time.Time, ttl time.Duration) *Session {
	return &Session{
		id:        id,
		token:     token,
		userID:    userID,
		createdAt: createdAt,
		updatedAt: updatedAt,
		ttl:       ttl,

		sm: nil,
	}
}

func (s *Session) ID() int64 {
	return s.id
}

func (s *Session) SetID(id int64) {
	s.id = id
}

func (s *Session) Token() string {
	return s.token
}

func (s *Session) UserID() int64 {
	return s.userID
}

func (s *Session) SetUserID(id int64) *Session {
	s.userID = id
	return s
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Session) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Session) Update() *Session {
	s.updatedAt = time.Now()
	return s
}

func (s *Session) TimeToLive() time.Duration {
	return time.Until(s.updatedAt.Add(s.ttl))
}

func (s *Session) TTL() time.Duration {
	return s.ttl
}

func (s *Session) SetTTL(ttl time.Duration) *Session {
	s.ttl = ttl
	return s
}

func (s *Session) Save() error {
	if s.sm == nil {
		return ErrNoSessionManager
	}
	return s.sm.storage.InsertSession(s)
}

func (s *Session) Remove() error {
	if s.sm == nil {
		return ErrNoSessionManager
	}
	return s.sm.storage.RemoveSession(s.token)
}

func (s *Session) AssignTo(ss *Safsm) *Session {
	s.sm = ss
	return s
}

func (s *Session) SetCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    "auth-token",
		Value:   s.Token(),
		Path:    "/",
		Expires: s.createdAt.Add(s.ttl),
	}
	http.SetCookie(w, &cookie)
}

func (s *Session) Valid() bool {
	return time.Now().Before(s.updatedAt.Add(s.ttl))
}

func (s *Session) Copy() *Session {
	return NewSession(s.id, s.token, s.userID, s.createdAt, s.updatedAt, s.ttl)
}
