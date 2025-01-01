package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// deprypt - глобальный объект через который работает Middleware.
var deprypt *decryptor

// Decryptor хранит ключ расшифровывания данных и реализует метод расшифровывания
type decryptor struct {
	privateKey *rsa.PrivateKey //ключ расшифровывания
}

// NewDecryptor разбирает файл с ключом и инициализирует синглтон deprypt.
func NewDecryptor(file string) error {

	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cannot read private key from file: %w", err)
	}

	keyBlock, _ := pem.Decode(b)
	if keyBlock == nil {
		return fmt.Errorf("bad private key blob: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("cannot parse private key: %w", err)
	}

	deprypt = &decryptor{
		privateKey: privateKey,
	}

	return nil
}

func (m *decryptor) Decrypt(message []byte) ([]byte, error) {
	msgLen := len(message)
	hash := sha512.New()
	random := rand.Reader

	step := m.privateKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, m.privateKey, message[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("decrypt part message process error: %w", err)
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}
