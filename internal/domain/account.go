package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (a *Account) GeneretePasswordHash(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	a.PasswordHash = string(hash)

	return nil
}

func (a *Account) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password))
	return err == nil
}

// New account with text password, hash will be generated automatically
func NewAccount(username string, password string) (*Account, error) {
	account := &Account{
		ID:        uuid.New(),
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := account.GeneretePasswordHash(password)
	if err != nil {
		return nil, err
	}

	return account, nil
}
