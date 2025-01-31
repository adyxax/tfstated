package model

import (
	"crypto/subtle"
	"encoding/json"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
)

type AccountContextKey struct{}

type Account struct {
	Id           string
	Username     string
	Salt         []byte
	PasswordHash []byte
	IsAdmin      bool
	Created      time.Time
	LastLogin    time.Time
	Settings     json.RawMessage
}

func (account *Account) CheckPassword(password string) bool {
	hash := helpers.HashPassword(password, account.Salt)
	return subtle.ConstantTimeCompare(hash, account.PasswordHash) == 1
}
