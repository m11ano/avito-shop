// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	domain "github.com/m11ano/avito-shop/internal/domain"
	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// OperationRepository is an autogenerated mock type for the OperationRepository type
type OperationRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, operation
func (_m *OperationRepository) Create(ctx context.Context, operation *domain.Operation) (int64, error) {
	ret := _m.Called(ctx, operation)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Operation) (int64, error)); ok {
		return rf(ctx, operation)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *domain.Operation) int64); ok {
		r0 = rf(ctx, operation)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *domain.Operation) error); ok {
		r1 = rf(ctx, operation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBalanceByAccountID provides a mock function with given fields: ctx, accountID
func (_m *OperationRepository) GetBalanceByAccountID(ctx context.Context, accountID uuid.UUID) (int64, bool, error) {
	ret := _m.Called(ctx, accountID)

	if len(ret) == 0 {
		panic("no return value specified for GetBalanceByAccountID")
	}

	var r0 int64
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) (int64, bool, error)); ok {
		return rf(ctx, accountID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) int64); ok {
		r0 = rf(ctx, accountID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) bool); ok {
		r1 = rf(ctx, accountID)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, uuid.UUID) error); ok {
		r2 = rf(ctx, accountID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewOperationRepository creates a new instance of OperationRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOperationRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *OperationRepository {
	mock := &OperationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
