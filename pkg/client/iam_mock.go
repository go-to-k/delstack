// Code generated by MockGen. DO NOT EDIT.
// Source: iam.go

package client

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	gomock "github.com/golang/mock/gomock"
)

// MockIIam is a mock of IIam interface.
type MockIIam struct {
	ctrl     *gomock.Controller
	recorder *MockIIamMockRecorder
}

// MockIIamMockRecorder is the mock recorder for MockIIam.
type MockIIamMockRecorder struct {
	mock *MockIIam
}

// NewMockIIam creates a new mock instance.
func NewMockIIam(ctrl *gomock.Controller) *MockIIam {
	mock := &MockIIam{ctrl: ctrl}
	mock.recorder = &MockIIamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIIam) EXPECT() *MockIIamMockRecorder {
	return m.recorder
}

// CheckRoleExists mocks base method.
func (m *MockIIam) CheckRoleExists(ctx context.Context, roleName *string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckRoleExists", ctx, roleName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckRoleExists indicates an expected call of CheckRoleExists.
func (mr *MockIIamMockRecorder) CheckRoleExists(ctx, roleName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckRoleExists", reflect.TypeOf((*MockIIam)(nil).CheckRoleExists), ctx, roleName)
}

// DeleteRole mocks base method.
func (m *MockIIam) DeleteRole(ctx context.Context, roleName *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRole", ctx, roleName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRole indicates an expected call of DeleteRole.
func (mr *MockIIamMockRecorder) DeleteRole(ctx, roleName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRole", reflect.TypeOf((*MockIIam)(nil).DeleteRole), ctx, roleName)
}

// DetachRolePolicies mocks base method.
func (m *MockIIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetachRolePolicies", ctx, roleName, policies)
	ret0, _ := ret[0].(error)
	return ret0
}

// DetachRolePolicies indicates an expected call of DetachRolePolicies.
func (mr *MockIIamMockRecorder) DetachRolePolicies(ctx, roleName, policies interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetachRolePolicies", reflect.TypeOf((*MockIIam)(nil).DetachRolePolicies), ctx, roleName, policies)
}

// ListAttachedRolePolicies mocks base method.
func (m *MockIIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAttachedRolePolicies", ctx, roleName)
	ret0, _ := ret[0].([]types.AttachedPolicy)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAttachedRolePolicies indicates an expected call of ListAttachedRolePolicies.
func (mr *MockIIamMockRecorder) ListAttachedRolePolicies(ctx, roleName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAttachedRolePolicies", reflect.TypeOf((*MockIIam)(nil).ListAttachedRolePolicies), ctx, roleName)
}
