package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// ErrNoKey is returned when an encryption operation is requested but no key is configured.
var ErrNoKey = errors.New("encryption key not configured")

// AEAD returns a cipher.AEAD derived from the provided key using Argon2.
func AEAD(key string) (cipher.AEAD, error) {
	if key == "" {
		return nil, ErrNoKey
	}
	salt := make([]byte, 16)
	copy(salt, key)
	k := argon2.IDKey([]byte(key), salt, 1, 64*1024, 4, 32)
	block, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// Encrypt returns a base64-encoded ciphertext of value using key.
func Encrypt(value, key string) (string, error) {
	if value == "" {
		return "", nil
	}
	if key == "" {
		return value, nil
	}
	aead, err := AEAD(key)
	if err != nil {
		return "", fmt.Errorf("init aead: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt returns the plaintext of a value encrypted by Encrypt.
func Decrypt(value, key string) (string, error) {
	if value == "" {
		return "", nil
	}
	if key == "" {
		return value, nil
	}
	aead, err := AEAD(key)
	if err != nil {
		return "", fmt.Errorf("init aead: %w", err)
	}
	ct, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("decode ciphertext: %w", err)
	}
	if len(ct) < aead.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := ct[:aead.NonceSize()], ct[aead.NonceSize():]
	pt, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(pt), nil
}
