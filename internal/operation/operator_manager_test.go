package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-to-k/delstack/internal/io"
	gomock "github.com/golang/mock/gomock"
)

/*
	Test Cases
*/

func TestOperatorManager_getOperatorResourcesLength(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(c *gomock.Controller, m *MockIOperatorCollection)
		want          int
	}{
		{
			name: "get operator resources length successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(c *gomock.Controller, m *MockIOperatorCollection) {
				var operators []IOperator

				cloudformationStackOperatorMock := NewMockIOperator(c)
				s3BucketOperatorMock := NewMockIOperator(c)
				iamRoleOperatorMock := NewMockIOperator(c)
				ecrRepositoryOperatorMock := NewMockIOperator(c)
				backupVaultOperatorMock := NewMockIOperator(c)
				customOperatorMock := NewMockIOperator(c)

				cloudformationStackOperatorMock.EXPECT().GetResourcesLength().Return(1)
				s3BucketOperatorMock.EXPECT().GetResourcesLength().Return(1)
				iamRoleOperatorMock.EXPECT().GetResourcesLength().Return(1)
				ecrRepositoryOperatorMock.EXPECT().GetResourcesLength().Return(1)
				backupVaultOperatorMock.EXPECT().GetResourcesLength().Return(1)
				customOperatorMock.EXPECT().GetResourcesLength().Return(1)

				operators = append(operators, cloudformationStackOperatorMock)
				operators = append(operators, s3BucketOperatorMock)
				operators = append(operators, iamRoleOperatorMock)
				operators = append(operators, ecrRepositoryOperatorMock)
				operators = append(operators, backupVaultOperatorMock)
				operators = append(operators, customOperatorMock)

				m.EXPECT().GetOperators().Return(operators)
			},
			want: 6,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			collectionMock := NewMockIOperatorCollection(ctrl)
			tt.prepareMockFn(ctrl, collectionMock)

			operatorManager := NewOperatorManager(collectionMock)

			got := operatorManager.getOperatorResourcesLength()
			if got != tt.want {
				t.Errorf("got = %#v, want %#v", got, tt.want)
				return
			}
		})
	}
}

func TestOperatorManager_CheckResourceCounts(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(c *gomock.Controller, m *MockIOperatorCollection)
		want          error
		wantErr       bool
	}{
		{
			name: "check resource counts successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(c *gomock.Controller, m *MockIOperatorCollection) {
				var operators []IOperator

				cloudformationStackOperatorMock := NewMockIOperator(c)
				s3BucketOperatorMock := NewMockIOperator(c)
				iamRoleOperatorMock := NewMockIOperator(c)
				ecrRepositoryOperatorMock := NewMockIOperator(c)
				backupVaultOperatorMock := NewMockIOperator(c)
				customOperatorMock := NewMockIOperator(c)

				cloudformationStackOperatorMock.EXPECT().GetResourcesLength().Return(1)
				s3BucketOperatorMock.EXPECT().GetResourcesLength().Return(1)
				iamRoleOperatorMock.EXPECT().GetResourcesLength().Return(1)
				ecrRepositoryOperatorMock.EXPECT().GetResourcesLength().Return(1)
				backupVaultOperatorMock.EXPECT().GetResourcesLength().Return(1)
				customOperatorMock.EXPECT().GetResourcesLength().Return(1)

				operators = append(operators, cloudformationStackOperatorMock)
				operators = append(operators, s3BucketOperatorMock)
				operators = append(operators, iamRoleOperatorMock)
				operators = append(operators, ecrRepositoryOperatorMock)
				operators = append(operators, backupVaultOperatorMock)
				operators = append(operators, customOperatorMock)

				m.EXPECT().GetOperators().Return(operators)

				m.EXPECT().GetLogicalResourceIds().Return(
					[]string{
						"logicalResourceId1",
						"logicalResourceId2",
						"logicalResourceId3",
						"logicalResourceId4",
						"logicalResourceId5",
						"logicalResourceId6",
					},
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "check resource counts failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(c *gomock.Controller, m *MockIOperatorCollection) {
				var operators []IOperator

				cloudformationStackOperatorMock := NewMockIOperator(c)
				s3BucketOperatorMock := NewMockIOperator(c)
				iamRoleOperatorMock := NewMockIOperator(c)
				ecrRepositoryOperatorMock := NewMockIOperator(c)
				backupVaultOperatorMock := NewMockIOperator(c)
				customOperatorMock := NewMockIOperator(c)

				cloudformationStackOperatorMock.EXPECT().GetResourcesLength().Return(1)
				s3BucketOperatorMock.EXPECT().GetResourcesLength().Return(1)
				iamRoleOperatorMock.EXPECT().GetResourcesLength().Return(1)
				ecrRepositoryOperatorMock.EXPECT().GetResourcesLength().Return(1)
				backupVaultOperatorMock.EXPECT().GetResourcesLength().Return(1)
				customOperatorMock.EXPECT().GetResourcesLength().Return(1)

				operators = append(operators, cloudformationStackOperatorMock)
				operators = append(operators, s3BucketOperatorMock)
				operators = append(operators, iamRoleOperatorMock)
				operators = append(operators, ecrRepositoryOperatorMock)
				operators = append(operators, backupVaultOperatorMock)
				operators = append(operators, customOperatorMock)

				m.EXPECT().GetOperators().Return(operators)

				m.EXPECT().GetLogicalResourceIds().Return(
					[]string{
						"logicalResourceId1",
						"logicalResourceId2",
					},
				)

				m.EXPECT().RaiseUnsupportedResourceError().Return(fmt.Errorf("UnsupportedResourceError"))
			},
			want:    fmt.Errorf("UnsupportedResourceError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			collectionMock := NewMockIOperatorCollection(ctrl)
			tt.prepareMockFn(ctrl, collectionMock)

			operatorManager := NewOperatorManager(collectionMock)

			err := operatorManager.CheckResourceCounts()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestOperatorManager_DeleteResourceCollection(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(c *gomock.Controller, m *MockIOperatorCollection)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resource collection successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(c *gomock.Controller, m *MockIOperatorCollection) {
				var operators []IOperator

				cloudformationStackOperatorMock := NewMockIOperator(c)
				s3BucketOperatorMock := NewMockIOperator(c)
				iamRoleOperatorMock := NewMockIOperator(c)
				ecrRepositoryOperatorMock := NewMockIOperator(c)
				backupVaultOperatorMock := NewMockIOperator(c)
				customOperatorMock := NewMockIOperator(c)

				cloudformationStackOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				s3BucketOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				iamRoleOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				ecrRepositoryOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				backupVaultOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				customOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)

				operators = append(operators, cloudformationStackOperatorMock)
				operators = append(operators, s3BucketOperatorMock)
				operators = append(operators, iamRoleOperatorMock)
				operators = append(operators, ecrRepositoryOperatorMock)
				operators = append(operators, backupVaultOperatorMock)
				operators = append(operators, customOperatorMock)

				m.EXPECT().GetOperators().Return(operators)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resource collection failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(c *gomock.Controller, m *MockIOperatorCollection) {
				var operators []IOperator

				cloudformationStackOperatorMock := NewMockIOperator(c)
				s3BucketOperatorMock := NewMockIOperator(c)
				iamRoleOperatorMock := NewMockIOperator(c)
				ecrRepositoryOperatorMock := NewMockIOperator(c)
				backupVaultOperatorMock := NewMockIOperator(c)
				customOperatorMock := NewMockIOperator(c)

				cloudformationStackOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(fmt.Errorf("ErrorDeleteResources"))
				s3BucketOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				iamRoleOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				ecrRepositoryOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				backupVaultOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)
				customOperatorMock.EXPECT().DeleteResources(gomock.Any()).Return(nil)

				operators = append(operators, cloudformationStackOperatorMock)
				operators = append(operators, s3BucketOperatorMock)
				operators = append(operators, iamRoleOperatorMock)
				operators = append(operators, ecrRepositoryOperatorMock)
				operators = append(operators, backupVaultOperatorMock)
				operators = append(operators, customOperatorMock)

				m.EXPECT().GetOperators().Return(operators)
			},
			want:    fmt.Errorf("ErrorDeleteResources"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			collectionMock := NewMockIOperatorCollection(ctrl)
			tt.prepareMockFn(ctrl, collectionMock)

			operatorManager := NewOperatorManager(collectionMock)

			err := operatorManager.DeleteResourceCollection(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
