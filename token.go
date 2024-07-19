package safsm

import (
	"crypto/rand"
	"fmt"
)

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
