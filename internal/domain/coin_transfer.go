package domain

import (
	"time"

	"github.com/google/uuid"
)

type CoinTransferType int8

const (
	CoinTransferTypeSending CoinTransferType = iota + 1
	CoinTransferTypeReciving
)

type CoinTransfer struct {
	ID                    uuid.UUID
	Type                  CoinTransferType
	OwnerAccountID        uuid.UUID
	CounterpartyAccountID uuid.UUID
	Amount                int64
	CreatedAt             time.Time
	IdentityKey           *uuid.UUID
}

func NewCoinTransfer(transferType CoinTransferType, counterpartyAccountID uuid.UUID, ownerAccountID uuid.UUID, amount int64, identityKey *uuid.UUID) *CoinTransfer {
	return &CoinTransfer{
		ID:                    uuid.New(),
		Type:                  transferType,
		OwnerAccountID:        ownerAccountID,
		CounterpartyAccountID: counterpartyAccountID,
		Amount:                amount,
		CreatedAt:             time.Now(),
		IdentityKey:           identityKey,
	}
}
