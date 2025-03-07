// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/m11ano/avito-shop/internal/domain"
	mock "github.com/stretchr/testify/mock"

	usecase "github.com/m11ano/avito-shop/internal/usecase"

	uuid "github.com/google/uuid"
)

// ShopPurchase is an autogenerated mock type for the ShopPurchase type
type ShopPurchase struct {
	mock.Mock
}

// GetInventory provides a mock function with given fields: ctx, accountID
func (_m *ShopPurchase) GetInventory(ctx context.Context, accountID uuid.UUID) ([]usecase.ShopPurchaseGetInventoryItem, error) {
	ret := _m.Called(ctx, accountID)

	if len(ret) == 0 {
		panic("no return value specified for GetInventory")
	}

	var r0 []usecase.ShopPurchaseGetInventoryItem
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) ([]usecase.ShopPurchaseGetInventoryItem, error)); ok {
		return rf(ctx, accountID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) []usecase.ShopPurchaseGetInventoryItem); ok {
		r0 = rf(ctx, accountID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]usecase.ShopPurchaseGetInventoryItem)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, accountID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MakePurchase provides a mock function with given fields: ctx, shopItemName, ownerAccountID, quantity, identityKey
func (_m *ShopPurchase) MakePurchase(ctx context.Context, shopItemName string, ownerAccountID uuid.UUID, quantity int64, identityKey *uuid.UUID) (*domain.ShopPurchase, error) {
	ret := _m.Called(ctx, shopItemName, ownerAccountID, quantity, identityKey)

	if len(ret) == 0 {
		panic("no return value specified for MakePurchase")
	}

	var r0 *domain.ShopPurchase
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, int64, *uuid.UUID) (*domain.ShopPurchase, error)); ok {
		return rf(ctx, shopItemName, ownerAccountID, quantity, identityKey)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, uuid.UUID, int64, *uuid.UUID) *domain.ShopPurchase); ok {
		r0 = rf(ctx, shopItemName, ownerAccountID, quantity, identityKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.ShopPurchase)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, uuid.UUID, int64, *uuid.UUID) error); ok {
		r1 = rf(ctx, shopItemName, ownerAccountID, quantity, identityKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewShopPurchase creates a new instance of ShopPurchase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewShopPurchase(t interface {
	mock.TestingT
	Cleanup(func())
}) *ShopPurchase {
	mock := &ShopPurchase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
