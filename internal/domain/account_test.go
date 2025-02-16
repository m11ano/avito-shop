package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountGeneretePasswordHash(t *testing.T) {
	account, err := NewAccount("test", "test")
	assert.NoError(t, err)

	currentPasswordHash := account.PasswordHash
	err = account.GeneretePasswordHash("test_another")
	assert.NoError(t, err)

	assert.NotEqual(t, currentPasswordHash, account.PasswordHash)
}

func TestAccountVerifyPassword(t *testing.T) {
	account, err := NewAccount("test", "test")
	assert.NoError(t, err)

	result := account.VerifyPassword("test")
	assert.True(t, result)

	result = account.VerifyPassword("test_another")
	assert.False(t, result)
}
