package helpers

import (
	"crypto/sha256"

	"git.adyxax.org/adyxax/tfstated/pkg/scrypto"
	"golang.org/x/crypto/pbkdf2"
)

const (
	PBKDF2Iterations = 600000
	SaltSize         = 32
)

func GenerateSalt() []byte {
	return scrypto.RandomBytes(SaltSize)
}

func HashPassword(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, PBKDF2Iterations, 32, sha256.New)
}
