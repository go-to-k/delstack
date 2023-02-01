// Code generated by MockGen. DO NOT EDIT.
// Source: ./operator.go

// Package operation is a generated GoMock package.
package operation

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	gomock "github.com/golang/mock/gomock"
)

// MockIOperator is a mock of IOperator interface.
type MockIOperator struct {
	ctrl     *gomock.Controller
	recorder *MockIOperatorMockRecorder
}

// MockIOperatorMockRecorder is the mock recorder for MockIOperator.
type MockIOperatorMockRecorder struct {
	mock *MockIOperator
}

// NewMockIOperator creates a new mock instance.
func NewMockIOperator(ctrl *gomock.Controller) *MockIOperator {
	mock := &MockIOperator{ctrl: ctrl}
	mock.recorder = &MockIOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIOperator) EXPECT() *MockIOperatorMockRecorder {
	return m.recorder
}

// AddResource mocks base method.
func (m *MockIOperator) AddResource(resource *types.StackResourceSummary) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddResource", resource)
}

// AddResource indicates an expected call of AddResource.
func (mr *MockIOperatorMockRecorder) AddResource(resource interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddResource", reflect.TypeOf((*MockIOperator)(nil).AddResource), resource)
}

// DeleteResources mocks base method.
func (m *MockIOperator) DeleteResources(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteResources", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteResources indicates an expected call of DeleteResources.
func (mr *MockIOperatorMockRecorder) DeleteResources(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteResources", reflect.TypeOf((*MockIOperator)(nil).DeleteResources), ctx)
}

// GetResourcesLength mocks base method.
func (m *MockIOperator) GetResourcesLength() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResourcesLength")
	ret0, _ := ret[0].(int)
	return ret0
}

// GetResourcesLength indicates an expected call of GetResourcesLength.
func (mr *MockIOperatorMockRecorder) GetResourcesLength() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourcesLength", reflect.TypeOf((*MockIOperator)(nil).GetResourcesLength))
}
