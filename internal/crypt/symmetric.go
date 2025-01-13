package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
)

func SymmetricEncrypt(password string, data []byte) ([]byte, error) {

	key := sha256.Sum256([]byte(password))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot crate new cipher.Block: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("cannot crate new block cipher wrapped in Galois Counter Mode: %w", err)
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	dst := aesgcm.Seal(nil, nonce, data, nil)

	return dst, nil
}

func SymmetricDecrypt(password string, encryptData []byte) ([]byte, error) {

	key := sha256.Sum256([]byte(password))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot crate new cipher.Block: %w", err)
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, fmt.Errorf("cannot crate new block cipher wrapped in Galois Counter Mode: %w", err)
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	data, err := aesgcm.Open(nil, nonce, encryptData, nil) // расшифровываем
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt data: %w", err)
	}

	return data, nil
}
