// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// TenantIdReader is an autogenerated mock type for the TenantIdReader type
type TenantIdReader struct {
	mock.Mock
}

type TenantIdReader_Expecter struct {
	mock *mock.Mock
}

func (_m *TenantIdReader) EXPECT() *TenantIdReader_Expecter {
	return &TenantIdReader_Expecter{mock: &_m.Mock}
}

// Read provides a mock function with given fields: parentCtx
func (_m *TenantIdReader) Read(parentCtx context.Context) (string, error) {
	ret := _m.Called(parentCtx)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (string, error)); ok {
		return rf(parentCtx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) string); ok {
		r0 = rf(parentCtx)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(parentCtx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TenantIdReader_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type TenantIdReader_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - parentCtx context.Context
func (_e *TenantIdReader_Expecter) Read(parentCtx interface{}) *TenantIdReader_Read_Call {
	return &TenantIdReader_Read_Call{Call: _e.mock.On("Read", parentCtx)}
}

func (_c *TenantIdReader_Read_Call) Run(run func(parentCtx context.Context)) *TenantIdReader_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *TenantIdReader_Read_Call) Return(_a0 string, _a1 error) *TenantIdReader_Read_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TenantIdReader_Read_Call) RunAndReturn(run func(context.Context) (string, error)) *TenantIdReader_Read_Call {
	_c.Call.Return(run)
	return _c
}

// NewTenantIdReader creates a new instance of TenantIdReader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTenantIdReader(t interface {
	mock.TestingT
	Cleanup(func())
}) *TenantIdReader {
	mock := &TenantIdReader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
