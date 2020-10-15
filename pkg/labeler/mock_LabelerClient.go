// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

// Re-generate by running 'make mock'

package labeler

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// MockLabelerClient is an autogenerated mock type for the LabelerClient type
type MockLabelerClient struct {
	mock.Mock
}

// SetLabels provides a mock function with given fields: ctx, in, opts
func (_m *MockLabelerClient) SetLabels(ctx context.Context, in *SetLabelsRequest, opts ...grpc.CallOption) (*SetLabelsReply, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *SetLabelsReply
	if rf, ok := ret.Get(0).(func(context.Context, *SetLabelsRequest, ...grpc.CallOption) *SetLabelsReply); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*SetLabelsReply)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *SetLabelsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
