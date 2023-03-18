// Code generated by MockGen. DO NOT EDIT.
// Source: ./cloudformation.go

package client

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	gomock "github.com/golang/mock/gomock"
)

// MockICloudFormation is a mock of ICloudFormation interface.
type MockICloudFormation struct {
	ctrl     *gomock.Controller
	recorder *MockICloudFormationMockRecorder
}

// MockICloudFormationMockRecorder is the mock recorder for MockICloudFormation.
type MockICloudFormationMockRecorder struct {
	mock *MockICloudFormation
}

// NewMockICloudFormation creates a new mock instance.
func NewMockICloudFormation(ctrl *gomock.Controller) *MockICloudFormation {
	mock := &MockICloudFormation{ctrl: ctrl}
	mock.recorder = &MockICloudFormationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICloudFormation) EXPECT() *MockICloudFormationMockRecorder {
	return m.recorder
}

// DeleteStack mocks base method.
func (m *MockICloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteStack", ctx, stackName, retainResources)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteStack indicates an expected call of DeleteStack.
func (mr *MockICloudFormationMockRecorder) DeleteStack(ctx, stackName, retainResources interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStack", reflect.TypeOf((*MockICloudFormation)(nil).DeleteStack), ctx, stackName, retainResources)
}

// DescribeStacks mocks base method.
func (m *MockICloudFormation) DescribeStacks(ctx context.Context, stackName *string) ([]types.Stack, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeStacks", ctx, stackName)
	ret0, _ := ret[0].([]types.Stack)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeStacks indicates an expected call of DescribeStacks.
func (mr *MockICloudFormationMockRecorder) DescribeStacks(ctx, stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeStacks", reflect.TypeOf((*MockICloudFormation)(nil).DescribeStacks), ctx, stackName)
}

// ListStackResources mocks base method.
func (m *MockICloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListStackResources", ctx, stackName)
	ret0, _ := ret[0].([]types.StackResourceSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListStackResources indicates an expected call of ListStackResources.
func (mr *MockICloudFormationMockRecorder) ListStackResources(ctx, stackName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListStackResources", reflect.TypeOf((*MockICloudFormation)(nil).ListStackResources), ctx, stackName)
}

// ListStacks mocks base method.
func (m *MockICloudFormation) ListStacks(ctx context.Context, stackStatusFilter []types.StackStatus) ([]types.StackSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListStacks", ctx, stackStatusFilter)
	ret0, _ := ret[0].([]types.StackSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListStacks indicates an expected call of ListStacks.
func (mr *MockICloudFormationMockRecorder) ListStacks(ctx, stackStatusFilter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListStacks", reflect.TypeOf((*MockICloudFormation)(nil).ListStacks), ctx, stackStatusFilter)
}
