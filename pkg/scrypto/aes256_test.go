package scrypto

import (
	"encoding/hex"
	"slices"
	"strings"
	"testing"
)

func TestAES256KeyHex(t *testing.T) {
	testKeyHex := "28278b7c0a25f01d3cab639633b9487f9ea1e9a2176dc9595a3f01323aa44284"
	testKey, _ := hex.DecodeString(testKeyHex)

	var key AES256Key
	if err := key.FromHex(testKeyHex); err != nil {
		t.Errorf("got unexpected error %+v", err)
	}
	if slices.Compare(testKey, key[:]) != 0 {
		t.Errorf("got %v, wanted %v", testKey, key[:])
	}
	if strings.Compare(testKeyHex, key.Hex()) != 0 {
		t.Errorf("got %v, wanted %v", testKeyHex, key.Hex())
	}
}

func TestAES256(t *testing.T) {
	keyHex := "28278b7c0a25f01d3cab639633b9487f9ea1e9a2176dc9595a3f01323aa44284"
	var key AES256Key
	if err := key.FromHex(keyHex); err != nil {
		t.Errorf("got unexpected error %+v", err)
	}

	data := []byte("Hello world!")
	encryptedData, err := key.EncryptAES256(data)
	if err != nil {
		t.Errorf("got unexpected error when encrypting data %+v", err)
	}

	decryptedData, err := key.DecryptAES256(encryptedData)
	if err != nil {
		t.Errorf("got unexpected error when decrypting data %+v", err)
	}

	if slices.Compare(data, decryptedData) != 0 {
		t.Errorf("got %v, wanted %v", decryptedData, data)
	}
}

func TestAES256InvalidData(t *testing.T) {
	keyHex := "28278b7c0a25f01d3cab639633b9487f9ea1e9a2176dc9595a3f01323aa44284"
	var key AES256Key
	if err := key.FromHex(keyHex); err != nil {
		t.Errorf("got unexpected error when converting data from base64: %+v", err)
	}

	iv := make([]byte, AES256IVSize)
	if _, err := key.DecryptAES256(append(iv, []byte("foo")...)); err == nil {
		t.Error("decrypting operation should have failed")
	}
}
