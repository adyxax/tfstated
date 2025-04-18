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
	expires := session.Created.Add(12 * time.Hour) // 12 hours sessions
	return time.Now().After(expires)
}
