package app

import (
	"crypto/rand"
	"fmt"
	"sync"
)

var (
	Users  sync.Map // username -> hashed password
	Todos  sync.Map // username -> []*Todo
	JwtKey []byte   // global JWT secret
)

func Init(secret string) {
	JwtKey = []byte(secret)
}

func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
