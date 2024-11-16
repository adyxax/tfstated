package model

import (
	"crypto/subtle"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/helpers"
)

type AccountContextKey struct{}

type Account struct {
	Id           int
	Username     string
	Salt         []byte
	PasswordHash []byte
	IsAdmin      bool
	Created      time.Time
	LastLogin    time.Time
	Settings     any
}

func (account *Account) CheckPassword(password string) bool {
	hash := helpers.HashPassword(password, account.Salt)
	return subtle.ConstantTimeCompare(hash, account.PasswordHash) == 1
}
