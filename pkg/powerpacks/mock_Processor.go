// Code generated by mockery v2.46.3. DO NOT EDIT.

package powerpacks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// MockProcessor is an autogenerated mock type for the Processor type
type MockProcessor struct {
	mock.Mock
}

type MockProcessor_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProcessor) EXPECT() *MockProcessor_Expecter {
	return &MockProcessor_Expecter{mock: &_m.Mock}
}

// Collect provides a mock function with given fields: _a0, rel, r
func (_m *MockProcessor) Collect(_a0 context.Context, rel string, r io.Reader) error {
	ret := _m.Called(_a0, rel, r)

	if len(ret) == 0 {
		panic("no return value specified for Collect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, io.Reader) error); ok {
		r0 = rf(_a0, rel, r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProcessor_Collect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Collect'
type MockProcessor_Collect_Call struct {
	*mock.Call
}

// Collect is a helper method to define mock.On call
//   - _a0 context.Context
//   - rel string
//   - r io.Reader
func (_e *MockProcessor_Expecter) Collect(_a0 interface{}, rel interface{}, r interface{}) *MockProcessor_Collect_Call {
	return &MockProcessor_Collect_Call{Call: _e.mock.On("Collect", _a0, rel, r)}
}

func (_c *MockProcessor_Collect_Call) Run(run func(_a0 context.Context, rel string, r io.Reader)) *MockProcessor_Collect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(io.Reader))
	})
	return _c
}

func (_c *MockProcessor_Collect_Call) Return(_a0 error) *MockProcessor_Collect_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProcessor_Collect_Call) RunAndReturn(run func(context.Context, string, io.Reader) error) *MockProcessor_Collect_Call {
	_c.Call.Return(run)
	return _c
}

// Write provides a mock function with given fields: ctx, path
func (_m *MockProcessor) Write(ctx context.Context, path string) error {
	ret := _m.Called(ctx, path)

	if len(ret) == 0 {
		panic("no return value specified for Write")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, path)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockProcessor_Write_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Write'
type MockProcessor_Write_Call struct {
	*mock.Call
}

// Write is a helper method to define mock.On call
//   - ctx context.Context
//   - path string
func (_e *MockProcessor_Expecter) Write(ctx interface{}, path interface{}) *MockProcessor_Write_Call {
	return &MockProcessor_Write_Call{Call: _e.mock.On("Write", ctx, path)}
}

func (_c *MockProcessor_Write_Call) Run(run func(ctx context.Context, path string)) *MockProcessor_Write_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProcessor_Write_Call) Return(_a0 error) *MockProcessor_Write_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockProcessor_Write_Call) RunAndReturn(run func(context.Context, string) error) *MockProcessor_Write_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProcessor creates a new instance of MockProcessor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProcessor {
	mock := &MockProcessor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
