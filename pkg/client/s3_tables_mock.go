// Code generated by MockGen. DO NOT EDIT.
// Source: s3_tables.go
//
// Generated by this command:
//
//	mockgen -source=s3_tables.go -destination=s3_tables_mock.go -package=client -write_package_comment=false
package client

import (
	context "context"
	reflect "reflect"

	types "github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	gomock "go.uber.org/mock/gomock"
)

// MockIS3Tables is a mock of IS3Tables interface.
type MockIS3Tables struct {
	ctrl     *gomock.Controller
	recorder *MockIS3TablesMockRecorder
}

// MockIS3TablesMockRecorder is the mock recorder for MockIS3Tables.
type MockIS3TablesMockRecorder struct {
	mock *MockIS3Tables
}

// NewMockIS3Tables creates a new mock instance.
func NewMockIS3Tables(ctrl *gomock.Controller) *MockIS3Tables {
	mock := &MockIS3Tables{ctrl: ctrl}
	mock.recorder = &MockIS3TablesMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIS3Tables) EXPECT() *MockIS3TablesMockRecorder {
	return m.recorder
}

// CheckTableBucketExists mocks base method.
func (m *MockIS3Tables) CheckTableBucketExists(ctx context.Context, tableBucketARN *string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckTableBucketExists", ctx, tableBucketARN)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckTableBucketExists indicates an expected call of CheckTableBucketExists.
func (mr *MockIS3TablesMockRecorder) CheckTableBucketExists(ctx, tableBucketARN any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckTableBucketExists", reflect.TypeOf((*MockIS3Tables)(nil).CheckTableBucketExists), ctx, tableBucketARN)
}

// DeleteNamespace mocks base method.
func (m *MockIS3Tables) DeleteNamespace(ctx context.Context, namespace, tableBucketARN *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteNamespace", ctx, namespace, tableBucketARN)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteNamespace indicates an expected call of DeleteNamespace.
func (mr *MockIS3TablesMockRecorder) DeleteNamespace(ctx, namespace, tableBucketARN any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteNamespace", reflect.TypeOf((*MockIS3Tables)(nil).DeleteNamespace), ctx, namespace, tableBucketARN)
}

// DeleteTable mocks base method.
func (m *MockIS3Tables) DeleteTable(ctx context.Context, tableName, namespace, tableBucketARN *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTable", ctx, tableName, namespace, tableBucketARN)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTable indicates an expected call of DeleteTable.
func (mr *MockIS3TablesMockRecorder) DeleteTable(ctx, tableName, namespace, tableBucketARN any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTable", reflect.TypeOf((*MockIS3Tables)(nil).DeleteTable), ctx, tableName, namespace, tableBucketARN)
}

// DeleteTableBucket mocks base method.
func (m *MockIS3Tables) DeleteTableBucket(ctx context.Context, tableBucketARN *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTableBucket", ctx, tableBucketARN)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTableBucket indicates an expected call of DeleteTableBucket.
func (mr *MockIS3TablesMockRecorder) DeleteTableBucket(ctx, tableBucketARN any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTableBucket", reflect.TypeOf((*MockIS3Tables)(nil).DeleteTableBucket), ctx, tableBucketARN)
}

// ListNamespacesByPage mocks base method.
func (m *MockIS3Tables) ListNamespacesByPage(ctx context.Context, tableBucketARN, continuationToken *string) (*ListNamespacesByPageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListNamespacesByPage", ctx, tableBucketARN, continuationToken)
	ret0, _ := ret[0].(*ListNamespacesByPageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListNamespacesByPage indicates an expected call of ListNamespacesByPage.
func (mr *MockIS3TablesMockRecorder) ListNamespacesByPage(ctx, tableBucketARN, continuationToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListNamespacesByPage", reflect.TypeOf((*MockIS3Tables)(nil).ListNamespacesByPage), ctx, tableBucketARN, continuationToken)
}

// ListTableBuckets mocks base method.
func (m *MockIS3Tables) ListTableBuckets(ctx context.Context) ([]types.TableBucketSummary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTableBuckets", ctx)
	ret0, _ := ret[0].([]types.TableBucketSummary)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTableBuckets indicates an expected call of ListTableBuckets.
func (mr *MockIS3TablesMockRecorder) ListTableBuckets(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTableBuckets", reflect.TypeOf((*MockIS3Tables)(nil).ListTableBuckets), ctx)
}

// ListTablesByPage mocks base method.
func (m *MockIS3Tables) ListTablesByPage(ctx context.Context, tableBucketARN, namespace, continuationToken *string) (*ListTablesByPageOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListTablesByPage", ctx, tableBucketARN, namespace, continuationToken)
	ret0, _ := ret[0].(*ListTablesByPageOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTablesByPage indicates an expected call of ListTablesByPage.
func (mr *MockIS3TablesMockRecorder) ListTablesByPage(ctx, tableBucketARN, namespace, continuationToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTablesByPage", reflect.TypeOf((*MockIS3Tables)(nil).ListTablesByPage), ctx, tableBucketARN, namespace, continuationToken)
}
