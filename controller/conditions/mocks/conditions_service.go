// Code generated by mockery v2.32.3. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionsService is an autogenerated mock type for the ConditionsService type
type ConditionsService struct {
	mock.Mock
}

type ConditionsService_Expecter struct {
	mock *mock.Mock
}

func (_m *ConditionsService) EXPECT() *ConditionsService_Expecter {
	return &ConditionsService_Expecter{mock: &_m.Mock}
}

// GetStatus provides a mock function with given fields: objectMeta, _a1
func (_m *ConditionsService) GetStatus(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition) *v1.ConditionStatus {
	ret := _m.Called(objectMeta, _a1)

	var r0 *v1.ConditionStatus
	if rf, ok := ret.Get(0).(func(v1.ObjectMeta, *[]v1.Condition) *v1.ConditionStatus); ok {
		r0 = rf(objectMeta, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConditionStatus)
		}
	}

	return r0
}

// ConditionsService_GetStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetStatus'
type ConditionsService_GetStatus_Call struct {
	*mock.Call
}

// GetStatus is a helper method to define mock.On call
//   - objectMeta v1.ObjectMeta
//   - _a1 *[]v1.Condition
func (_e *ConditionsService_Expecter) GetStatus(objectMeta interface{}, _a1 interface{}) *ConditionsService_GetStatus_Call {
	return &ConditionsService_GetStatus_Call{Call: _e.mock.On("GetStatus", objectMeta, _a1)}
}

func (_c *ConditionsService_GetStatus_Call) Run(run func(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition)) *ConditionsService_GetStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(v1.ObjectMeta), args[1].(*[]v1.Condition))
	})
	return _c
}

func (_c *ConditionsService_GetStatus_Call) Return(_a0 *v1.ConditionStatus) *ConditionsService_GetStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ConditionsService_GetStatus_Call) RunAndReturn(run func(v1.ObjectMeta, *[]v1.Condition) *v1.ConditionStatus) *ConditionsService_GetStatus_Call {
	_c.Call.Return(run)
	return _c
}

// SetFalse provides a mock function with given fields: objectMeta, _a1, reason, message
func (_m *ConditionsService) SetFalse(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition, reason string, message string) {
	_m.Called(objectMeta, _a1, reason, message)
}

// ConditionsService_SetFalse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetFalse'
type ConditionsService_SetFalse_Call struct {
	*mock.Call
}

// SetFalse is a helper method to define mock.On call
//   - objectMeta v1.ObjectMeta
//   - _a1 *[]v1.Condition
//   - reason string
//   - message string
func (_e *ConditionsService_Expecter) SetFalse(objectMeta interface{}, _a1 interface{}, reason interface{}, message interface{}) *ConditionsService_SetFalse_Call {
	return &ConditionsService_SetFalse_Call{Call: _e.mock.On("SetFalse", objectMeta, _a1, reason, message)}
}

func (_c *ConditionsService_SetFalse_Call) Run(run func(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition, reason string, message string)) *ConditionsService_SetFalse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(v1.ObjectMeta), args[1].(*[]v1.Condition), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *ConditionsService_SetFalse_Call) Return() *ConditionsService_SetFalse_Call {
	_c.Call.Return()
	return _c
}

func (_c *ConditionsService_SetFalse_Call) RunAndReturn(run func(v1.ObjectMeta, *[]v1.Condition, string, string)) *ConditionsService_SetFalse_Call {
	_c.Call.Return(run)
	return _c
}

// SetTrue provides a mock function with given fields: objectMeta, _a1, reason, message
func (_m *ConditionsService) SetTrue(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition, reason string, message string) {
	_m.Called(objectMeta, _a1, reason, message)
}

// ConditionsService_SetTrue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetTrue'
type ConditionsService_SetTrue_Call struct {
	*mock.Call
}

// SetTrue is a helper method to define mock.On call
//   - objectMeta v1.ObjectMeta
//   - _a1 *[]v1.Condition
//   - reason string
//   - message string
func (_e *ConditionsService_Expecter) SetTrue(objectMeta interface{}, _a1 interface{}, reason interface{}, message interface{}) *ConditionsService_SetTrue_Call {
	return &ConditionsService_SetTrue_Call{Call: _e.mock.On("SetTrue", objectMeta, _a1, reason, message)}
}

func (_c *ConditionsService_SetTrue_Call) Run(run func(objectMeta v1.ObjectMeta, _a1 *[]v1.Condition, reason string, message string)) *ConditionsService_SetTrue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(v1.ObjectMeta), args[1].(*[]v1.Condition), args[2].(string), args[3].(string))
	})
	return _c
}

func (_c *ConditionsService_SetTrue_Call) Return() *ConditionsService_SetTrue_Call {
	_c.Call.Return()
	return _c
}

func (_c *ConditionsService_SetTrue_Call) RunAndReturn(run func(v1.ObjectMeta, *[]v1.Condition, string, string)) *ConditionsService_SetTrue_Call {
	_c.Call.Return(run)
	return _c
}

// NewConditionsService creates a new instance of ConditionsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConditionsService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConditionsService {
	mock := &ConditionsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
