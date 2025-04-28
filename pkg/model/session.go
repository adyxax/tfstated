package model

import (
	"encoding/json"
	"time"

	"go.n16f.net/uuid"
)

type SessionContextKey struct{}

type Session struct {
	Id        []byte
	AccountId *uuid.UUID
	Created   time.Time
	Updated   time.Time
	Settings  json.RawMessage
}

func (session *Session) IsExpired() bool {
	expires := session.Created.Add(12 * time.Hour) // 12 hours sessions
	return time.Now().After(expires)
}
