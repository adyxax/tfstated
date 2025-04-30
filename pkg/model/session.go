package model

import (
	"fmt"
	"time"

	"go.n16f.net/uuid"
)

type SessionData struct {
	Account   *Account  `json:"account"`
	CsrfToken uuid.UUID `json:"csrf_token"`
	Settings  *Settings `json:"settings"`
}

func NewSessionData(account *Account, previousSessionSettings *Settings) (*SessionData, error) {
	data := SessionData{Account: account}
	if err := data.CsrfToken.Generate(uuid.V4); err != nil {
		return nil, fmt.Errorf("failed to generate csrf token uuid: %w", err)
	}
	if account != nil {
		data.Settings = account.Settings
	} else if previousSessionSettings != nil {
		data.Settings = previousSessionSettings
	} else {
		data.Settings = &Settings{}
	}
	return &data, nil
}

type SessionContextKey struct{}

type Session struct {
	Id      []byte
	Created time.Time
	Updated time.Time
	Data    *SessionData
}

func (session *Session) IsExpired() bool {
	expires := session.Created.Add(12 * time.Hour) // 12 hours sessions
	return time.Now().After(expires)
}
