// Code generated by mockery v2.50.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// URL is an autogenerated mock type for the URL type
type URL struct {
	mock.Mock
}

// ResolveURL provides a mock function with given fields: ctx, shorted
func (_m *URL) ResolveURL(ctx context.Context, shorted string) (string, error) {
	ret := _m.Called(ctx, shorted)

	if len(ret) == 0 {
		panic("no return value specified for ResolveURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, shorted)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, shorted)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, shorted)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ShortenURL provides a mock function with given fields: ctx, original
func (_m *URL) ShortenURL(ctx context.Context, original string) (string, error) {
	ret := _m.Called(ctx, original)

	if len(ret) == 0 {
		panic("no return value specified for ShortenURL")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, original)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, original)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, original)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewURL creates a new instance of URL. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewURL(t interface {
	mock.TestingT
	Cleanup(func())
}) *URL {
	mock := &URL{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
