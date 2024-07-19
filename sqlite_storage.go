package safsm

import (
	"database/sql"
	"fmt"
	"time"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(db *sql.DB) (*SQLiteStorage, error) {
	s := &SQLiteStorage{
		db: db,
	}

	// check: is connection alive
	if err := s.db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to check db connection: %v", err)
	}

	// create tables
	if err := s.createTables(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SQLiteStorage) createTables() error {
	q := `
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT UNIQUE,
		user_id INTEGER,
		created_at INTEGER,
		updated_at INTEGER,
		ttl INTEGER
	) STRICT;`
	if _, err := s.db.Exec(q); err != nil {
		return fmt.Errorf("failed to create sessions table: %v", err)
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close db: %v", err)
	}
	return nil
}

func (s *SQLiteStorage) InsertSession(session *Session) error {
	if session.id > -1 {
		q := `UPDATE sessions SET user_id = ?, updated_at = ?, ttl = ? WHERE id=?`
		if _, err := s.db.Exec(q, session.UserID(), session.updatedAt.Unix(), session.TTL(), session.ID()); err != nil {
			return fmt.Errorf("failed to update session: %v", err)
		}
		return nil
	}

	q := `INSERT INTO sessions (token, user_id, created_at, updated_at, ttl) VALUES (?,?,?,?,?)`
	if re, err := s.db.Exec(q, session.Token(), session.UserID(), session.CreatedAt().Unix(), session.UpdatedAt().Unix(), session.TTL()); err == nil {
		insertedID, _ := re.LastInsertId()
		session.SetID(insertedID)
		return nil
	} else {
		return fmt.Errorf("failed to insert session: %v", err)
	}
}

func (s *SQLiteStorage) FindSession(token string) (*Session, error) {
	q := `SELECT * FROM sessions WHERE token=?`
	row := s.db.QueryRow(q, token)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	return s.scanSession(row.Scan)
}

func (s *SQLiteStorage) RemoveSession(token string) error {
	q := `DELETE FROM sessions WHERE token=?`
	if _, err := s.db.Exec(q, token); err != nil {
		return fmt.Errorf("failed to remove session: %v", err)
	}
	return nil
}

func (s *SQLiteStorage) Each(f func(session *Session)) {
	// HACK to avoid db lock
	sessions := []*Session{}
	rows, err := s.db.Query(`SELECT * FROM sessions`)
	if err != nil {
		return
	}
	for rows.Next() {
		session, err := s.scanSession(rows.Scan)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	for _, v := range sessions {
		f(v)
	}
}

func (s *SQLiteStorage) scanSession(scanf func(dest ...any) error) (*Session, error) {
	var (
		sessionID        int64
		sessionToken     string
		sessionUserID    int64
		sessionCreatedAt int64
		sessionUpdatedAt int64
		sessionTTL       int64
	)
	if err := scanf(&sessionID, &sessionToken, &sessionUserID, &sessionCreatedAt,
		&sessionUpdatedAt, &sessionTTL); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoSession
		}
		return nil, fmt.Errorf("failed to scan session: %v", err)
	}
	return NewSession(
		sessionID,
		sessionToken,
		sessionUserID,
		time.Unix(sessionCreatedAt, 0),
		time.Unix(sessionUpdatedAt, 0),
		time.Duration(sessionTTL),
	), nil
}
