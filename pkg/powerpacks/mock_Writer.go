// Code generated by mockery v2.46.3. DO NOT EDIT.

package powerpacks

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// MockWriter is an autogenerated mock type for the Writer type
type MockWriter struct {
	mock.Mock
}

type MockWriter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockWriter) EXPECT() *MockWriter_Expecter {
	return &MockWriter_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: name
func (_m *MockWriter) Execute(name string) (io.Writer, error) {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 io.Writer
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (io.Writer, error)); ok {
		return rf(name)
	}
	if rf, ok := ret.Get(0).(func(string) io.Writer); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockWriter_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockWriter_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - name string
func (_e *MockWriter_Expecter) Execute(name interface{}) *MockWriter_Execute_Call {
	return &MockWriter_Execute_Call{Call: _e.mock.On("Execute", name)}
}

func (_c *MockWriter_Execute_Call) Run(run func(name string)) *MockWriter_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockWriter_Execute_Call) Return(_a0 io.Writer, _a1 error) *MockWriter_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockWriter_Execute_Call) RunAndReturn(run func(string) (io.Writer, error)) *MockWriter_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockWriter creates a new instance of MockWriter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockWriter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWriter {
	mock := &MockWriter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
