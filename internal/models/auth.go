package models

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthDTO struct {
	Login    string `json:"login"`    //Логин пользователя
	Password string `json:"password"` //Пароль пользователя
}

func (m *AuthDTO) Validate() error {
	if m.Login == `` {
		return fmt.Errorf("login required")
	}
	if m.Password == `` {
		return fmt.Errorf("password required")
	}

	return nil
}

func (m *AuthDTO) CheckPassword(passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(m.Password))
	return err == nil
}

func (m *AuthDTO) GeneratePasswordHash() error {

	if m.Password == "" {
		return fmt.Errorf("password required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(m.Password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("failed to hash password due to error %w", err)
	}

	m.Password = string(hash)
	return nil
}
