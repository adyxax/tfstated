package helpers

import (
	"crypto/sha256"

	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2PasswordIterations = 600000
	PBKDF2SessionIterations  = 12
	SaltSize                 = 32
)

func GenerateSalt() []byte {
	return scrypto.RandomBytes(SaltSize)
}

func HashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2PasswordIterations, 32, sha256.New)
}

func HashSessionId(id []byte, salt []byte) []byte {
	return pbkdf2.Key(id, salt, PBKDF2SessionIterations, 32, sha256.New)
}
