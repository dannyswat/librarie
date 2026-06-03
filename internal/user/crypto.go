package user

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type credentialCipher struct {
	aead cipher.AEAD
}

func newCredentialCipher(rawKey string) (*credentialCipher, error) {
	key, err := parseEncryptionKey(rawKey)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create AES-GCM: %w", err)
	}

	return &credentialCipher{aead: aead}, nil
}

func (c *credentialCipher) encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

func parseEncryptionKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, fmt.Errorf("LIBRARIE_ENCRYPTION_KEY is required")
	}

	if key, err := hex.DecodeString(raw); err == nil {
		if len(key) == 32 {
			return key, nil
		}
	}

	if key, err := base64.StdEncoding.DecodeString(raw); err == nil {
		if len(key) == 32 {
			return key, nil
		}
	}

	if key, err := base64.RawStdEncoding.DecodeString(raw); err == nil {
		if len(key) == 32 {
			return key, nil
		}
	}

	key := []byte(raw)
	if len(key) == 32 {
		return key, nil
	}

	return nil, fmt.Errorf("LIBRARIE_ENCRYPTION_KEY must decode to 32 bytes")
}
