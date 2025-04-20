package model

import (
	"crypto/subtle"
	"encoding/json"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"go.n16f.net/uuid"
)

type AccountContextKey struct{}

type Account struct {
	Id            uuid.UUID
	Username      string
	Salt          []byte
	PasswordHash  []byte
	IsAdmin       bool
	Created       time.Time
	LastLogin     time.Time
	Settings      json.RawMessage
	PasswordReset *uuid.UUID
}

func (account *Account) CheckPassword(password string) bool {
	hash := helpers.HashPassword(password, account.Salt)
	return subtle.ConstantTimeCompare(hash, account.PasswordHash) == 1
}

func (account *Account) SetPassword(password string) {
	account.Salt = helpers.GenerateSalt()
	account.PasswordHash = helpers.HashPassword(password, account.Salt)
	account.PasswordReset = nil
}
