package safsm

type Storage interface {
	Close() error
	InsertSession(session *Session) error
	FindSession(token string) (*Session, error)
	RemoveSession(token string) error
	Each(f func(session *Session))
}
