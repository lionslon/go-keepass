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

// Encryptor хранит ключ шифрования и реализует метод шифрования.
type Encryptor struct {
	openkey *rsa.PublicKey // ключ шифрования
}

// NewEncryptor разбирает файл с ключом и инициализирует синглтон encrypt.
func NewEncryptor(file string) (*Encryptor, error) {

	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read open key from file: %w", err)
	}

	keyBlock, _ := pem.Decode(b)
	if keyBlock == nil {
		return nil, fmt.Errorf("bad open key blob: %w", err)
	}

	pubKey, err := x509.ParsePKCS1PublicKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse open key: %w", err)
	}

	return &Encryptor{
		openkey: pubKey,
	}, nil
}

func (m *Encryptor) Encrypt(message []byte) ([]byte, error) {

	hash := sha512.New()
	random := rand.Reader

	msgLen := len(message)
	step := m.openkey.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, m.openkey, message[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("encrypt part message process error: %w", err)
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}
