package model

import (
	"crypto/subtle"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
	"go.n16f.net/uuid"
)

type AccountContextKey struct{}

type Account struct {
	Id            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Salt          []byte     `json:"salt"`
	PasswordHash  []byte     `json:"password_hash"`
	IsAdmin       bool       `json:"is_admin"`
	Created       time.Time  `json:"created"`
	LastLogin     time.Time  `json:"last_login"`
	Settings      *Settings  `json:"settings"`
	PasswordReset *uuid.UUID `json:"password_reset"`
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
