# safsm - Simple As F⭐️ck Session Manager

## Installation

```bash
go get -u github.com/nyan2d/safsm
```

## Session storage

We have several session storage types: **MemoryStorage**, **FileStorage**, **SQLiteStorage**.

But you can implement your own type using the interface below:

```go
type Storage interface {
    Close() error
    InsertSession(session *Session) error
    FindSession(token string) (*Session, error)
    RemoveSession(token string) error
    Each(f func(session *Session))
}
```

## Cooking

### Create a session manager

``` go
// open sqlite database
db, err := sql.Open("sqlite", "sessions.db")
if err != nil {
    ...
}

// create storage
storage, err := safsm.NewSQLiteStorage(db)
if err != nil {
    ...
}

// create session manager
sm := safsm.New(storage)
```

### Create a session

```go
ttl := time.Minute * 60
session := sm.CreateSession(userID, ttl)
```

### Playing with cookies

#### Write a session to cookies

```go
func demoHandler (w http.ResponseWriter, r *http.Request) {
    session.SetCookie(w)
}
```

#### Read session from cookies
```go
func demoHandler (w http.ResponseWriter, r *http.Request) {
    session, err := safsm.ReadSession(r)
    if err == safsm.ErrNoSession {
        // there is no session
    }
}
```

## Caching

Caching is implemented by a storage-wrapper with the type CachedStorage

```go
cacheSize := 1024
cachedStorage := safsm.NewCachedStorage(originalStorage, cacheSize)
```

**Warning!** Cache invalidation works very simply: if the number of sessions hit by the cache exceeds the cache size, the entire cache is just invalidated.

This behavior can be reworked in the future.

## Assigning session to session manager

If you want to save a session from a storage you work with, you need to assign the session to some session manager

```go
session.AssignTo(sm)
session.Save()
```

## Session functions chaining

Some of session funcs can be chained. For example:

```go
session.AssignTo(sm).
    Update().
    SetTTL(time.Minute * 60).
    Save()
```

## Session lifetime

A session has the following time parameters:

+ CreatedAt
+ UpdatedAt
+ TTL

If the current time exceeds the session update time + TTL, the session is considered invalid. If you need to update the session, you can use the following approach:

```go
if session.TimeToLive().Minutes() < 60 {
    session.Update().AssignTo(sm).Save()
}
```

## License

The code is distributed under the **MIT** license.