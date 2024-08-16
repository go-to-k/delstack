// Code generated by MockGen. DO NOT EDIT.
// Source: iam.go
//
// Generated by this command:
//
//	mockgen -source=iam.go -destination=iam_mock.go -package=client -write_package_comment=false
package client

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	gomock "go.uber.org/mock/gomock"
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

// CheckGroupExists mocks base method.
func (m *MockIIam) CheckGroupExists(ctx context.Context, groupName *string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckGroupExists", ctx, groupName)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckGroupExists indicates an expected call of CheckGroupExists.
func (mr *MockIIamMockRecorder) CheckGroupExists(ctx, groupName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckGroupExists", reflect.TypeOf((*MockIIam)(nil).CheckGroupExists), ctx, groupName)
}

// DeleteGroup mocks base method.
func (m *MockIIam) DeleteGroup(ctx context.Context, groupName *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGroup", ctx, groupName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGroup indicates an expected call of DeleteGroup.
func (mr *MockIIamMockRecorder) DeleteGroup(ctx, groupName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGroup", reflect.TypeOf((*MockIIam)(nil).DeleteGroup), ctx, groupName)
}

// GetGroupUsers mocks base method.
func (m *MockIIam) GetGroupUsers(ctx context.Context, groupName *string) ([]types.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGroupUsers", ctx, groupName)
	ret0, _ := ret[0].([]types.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGroupUsers indicates an expected call of GetGroupUsers.
func (mr *MockIIamMockRecorder) GetGroupUsers(ctx, groupName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGroupUsers", reflect.TypeOf((*MockIIam)(nil).GetGroupUsers), ctx, groupName)
}

// RemoveUsersFromGroup mocks base method.
func (m *MockIIam) RemoveUsersFromGroup(ctx context.Context, groupName *string, users []types.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveUsersFromGroup", ctx, groupName, users)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveUsersFromGroup indicates an expected call of RemoveUsersFromGroup.
func (mr *MockIIamMockRecorder) RemoveUsersFromGroup(ctx, groupName, users any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveUsersFromGroup", reflect.TypeOf((*MockIIam)(nil).RemoveUsersFromGroup), ctx, groupName, users)
}
