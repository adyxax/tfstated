package model

import (
	"time"
)

type SessionContextKey struct{}

type Session struct {
	Id        string
	AccountId int
	Created   time.Time
	Updated   time.Time
	Data      any
}

func (session *Session) IsExpired() bool {
	// TODO
	return false
}
