// Code generated by mockery. DO NOT EDIT.

package main

import (
	context "context"

	shared "github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
	mock "github.com/stretchr/testify/mock"
)

// mockStore is an autogenerated mock type for the Store type
type mockStore struct {
	mock.Mock
}

type mockStore_Expecter struct {
	mock *mock.Mock
}

func (_m *mockStore) EXPECT() *mockStore_Expecter {
	return &mockStore_Expecter{mock: &_m.Mock}
}

// GetList provides a mock function with given fields: ctx, uids
func (_m *mockStore) GetList(ctx context.Context, uids []string) ([]shared.Lpa, error) {
	ret := _m.Called(ctx, uids)

	if len(ret) == 0 {
		panic("no return value specified for GetList")
	}

	var r0 []shared.Lpa
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]shared.Lpa, error)); ok {
		return rf(ctx, uids)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []shared.Lpa); ok {
		r0 = rf(ctx, uids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]shared.Lpa)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, uids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockStore_GetList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetList'
type mockStore_GetList_Call struct {
	*mock.Call
}

// GetList is a helper method to define mock.On call
//   - ctx context.Context
//   - uids []string
func (_e *mockStore_Expecter) GetList(ctx interface{}, uids interface{}) *mockStore_GetList_Call {
	return &mockStore_GetList_Call{Call: _e.mock.On("GetList", ctx, uids)}
}

func (_c *mockStore_GetList_Call) Run(run func(ctx context.Context, uids []string)) *mockStore_GetList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]string))
	})
	return _c
}

func (_c *mockStore_GetList_Call) Return(_a0 []shared.Lpa, _a1 error) *mockStore_GetList_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockStore_GetList_Call) RunAndReturn(run func(context.Context, []string) ([]shared.Lpa, error)) *mockStore_GetList_Call {
	_c.Call.Return(run)
	return _c
}

// newMockStore creates a new instance of mockStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockStore {
	mock := &mockStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
