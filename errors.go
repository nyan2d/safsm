package safsm

import "errors"

var (
	ErrNoSession        = errors.New("session not found")
	ErrNoSessionManager = errors.New("not assigned session manager")
	ErrNoBearerToken    = errors.New("no bearer token set")
)
