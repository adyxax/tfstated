package scrypto

import (
	"crypto/rand"
	"fmt"
)

func RandomBytes(n int) []byte {
	data := make([]byte, n)

	if _, err := rand.Read(data); err != nil {
		panic(fmt.Sprintf("cannot generate random data: %+v", err))
	}

	return data
}
