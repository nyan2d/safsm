package safsm

import "sync"

type CachedStorage struct {
	original Storage
	cache    map[string]*Session
	limit    int
	m        *sync.Mutex
}

func NewCachedStorage(storage Storage, limit int) *CachedStorage {
	return &CachedStorage{
		original: storage,
		cache:    map[string]*Session{},
		limit:    limit,
		m:        &sync.Mutex{},
	}
}

func (c *CachedStorage) Close() error {
	c.cache = nil
	return c.original.Close()
}

func (c *CachedStorage) InsertSession(session *Session) error {
	if len(c.cache) > c.limit {
		c.m.Lock()
		c.cache = map[string]*Session{}
		c.m.Unlock()
	}

	if err := c.original.InsertSession(session); err != nil {
		return err
	}
	c.m.Lock()
	c.cache[session.token] = session
	c.m.Unlock()
	return nil
}

func (c *CachedStorage) FindSession(token string) (*Session, error) {
	if len(c.cache) > c.limit {
		c.m.Lock()
		c.cache = map[string]*Session{}
		c.m.Unlock()
	}

	if session, ok := c.cache[token]; ok {
		if session.Valid() {
			return session, nil
		} else {
			c.m.Lock()
			delete(c.cache, token)
			c.m.Unlock()
		}
	}

	session, err := c.original.FindSession(token)
	if err == nil {
		c.m.Lock()
		c.cache[session.token] = session
		c.m.Unlock()
		return session, nil
	}

	return session, err
}

func (c *CachedStorage) RemoveSession(token string) error {
	c.m.Lock()
	delete(c.cache, token)
	c.m.Unlock()
	return c.original.RemoveSession(token)
}

func (c *CachedStorage) Each(f func(session *Session)) {
	// TODO: do something
	// HACK: as we use this function to remove invalid session, we must to invalidate the cache D:
	c.cache = map[string]*Session{}
	c.original.Each(f)
}
