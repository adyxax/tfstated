package model

import (
	"crypto/sha256"
	"crypto/subtle"
	"time"

	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2Iterations = 600000
	SaltSize         = 32
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
	hash := HashPassword(password, account.Salt)
	return subtle.ConstantTimeCompare(hash, account.PasswordHash) == 1
}

func GenerateSalt() []byte {
	return scrypto.RandomBytes(SaltSize)
}

func HashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, 32, sha256.New)
}
