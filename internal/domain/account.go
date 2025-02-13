package domain

import (
	"github.com/google/uuid"
	"github.com/m11ano/avito-shop/pkg/cryptopass"
)

type Account struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
}

func (a *Account) GeneretePasswordHash(password string) error {
	salt, err := cryptopass.GenerateSalt()
	if err != nil {
		return err
	}

	hash, err := cryptopass.HashPasswordArgon2(password, salt)
	if err != nil {
		return err
	}

	a.PasswordHash = hash

	return nil
}

func (a *Account) VerifyPassword(password string) (bool, error) {
	return cryptopass.VerifyPasswordArgon2(password, a.PasswordHash)
}

// New account with text password, hash will be generated automatically
func NewAccount(username string, password string) (*Account, error) {
	account := &Account{
		ID:       uuid.New(),
		Username: username,
	}

	err := account.GeneretePasswordHash(password)
	if err != nil {
		return nil, err
	}

	return account, nil
}
