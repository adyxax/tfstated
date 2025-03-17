package model

import (
	"time"

	"go.n16f.net/uuid"
)

type SessionContextKey struct{}

type Session struct {
	Id        string
	AccountId uuid.UUID
	Created   time.Time
	Updated   time.Time
	Data      any
}

func (session *Session) IsExpired() bool {
	// TODO
	return false
}
