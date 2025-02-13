package domain

import (
	"time"

	"github.com/google/uuid"
)

type OperationType int8

const (
	OperationTypeIncrease OperationType = iota + 1
	OperationTypeDecrease
)

type OperationSourceType int8

const (
	OperationSourceTypeDeposit OperationSourceType = iota + 1
	OperationSourceTypeShopPurchase
	OperationSourceTypeTransfer
)

type Operation struct {
	ID         uuid.UUID
	Type       OperationType
	AccountID  uuid.UUID
	Amount     int64
	SourceType OperationSourceType
	SourceID   *uuid.UUID
	CreatedAt  time.Time
}

func NewOperation(opertationType OperationType, accountID uuid.UUID, amount int64, sourceType OperationSourceType, sourceID *uuid.UUID) *Operation {
	return &Operation{
		ID:         uuid.New(),
		Type:       opertationType,
		AccountID:  accountID,
		Amount:     amount,
		SourceType: sourceType,
		SourceID:   sourceID,
		CreatedAt:  time.Now(),
	}
}
