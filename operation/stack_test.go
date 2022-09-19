package operation

// var _ client.ICloudFormation = (*mockCloudFormation)(nil)
// var _ client.ICloudFormation = (*allErrorMockCloudFormation)(nil)
// var _ client.ICloudFormation = (*listRecoveryPointsErrorMockCloudFormation)(nil)
// var _ client.ICloudFormation = (*deleteRecoveryPointsErrorMockCloudFormation)(nil)
// var _ client.ICloudFormation = (*deleteCloudFormationVaultErrorMockCloudFormation)(nil)

// /*
// 	Mocks for client
// */
// type mockCloudFormation struct{}

// func NewMockCloudFormation() *mockCloudFormation {
// 	return &mockCloudFormation{}
// }

// func (m *mockCloudFormation) ListRecoveryPointsByCloudFormationVault(cloudformationVaultName *string) ([]types.RecoveryPointByCloudFormationVault, error) {
// 	output := []types.RecoveryPointByCloudFormationVault{
// 		{
// 			CloudFormationVaultName: aws.String("CloudFormationVaultName1"),
// 			CloudFormationVaultArn:  aws.String("CloudFormationVaultArn1"),
// 		},
// 		{
// 			CloudFormationVaultName: aws.String("CloudFormationVaultName2"),
// 			CloudFormationVaultArn:  aws.String("CloudFormationVaultArn2"),
// 		},
// 	}
// 	return output, nil
// }

// func (m *mockCloudFormation) DeleteRecoveryPoints(cloudformationVaultName *string, recoveryPoints []types.RecoveryPointByCloudFormationVault) error {
// 	return nil
// }

// func (m *mockCloudFormation) DeleteRecoveryPoint(cloudformationVaultName *string, recoveryPointArn *string) error {
// 	return nil
// }

// func (m *mockCloudFormation) DeleteCloudFormationVault(cloudformationVaultName *string) error {
// 	return nil
// }

// type allErrorMockCloudFormation struct{}

// func NewAllErrorMockCloudFormation() *allErrorMockCloudFormation {
// 	return &allErrorMockCloudFormation{}
// }

// func (m *allErrorMockCloudFormation) ListRecoveryPointsByCloudFormationVault(cloudformationVaultName *string) ([]types.RecoveryPointByCloudFormationVault, error) {
// 	return nil, fmt.Errorf("ListRecoveryPointsByCloudFormationVaultError")
// }

// func (m *allErrorMockCloudFormation) DeleteRecoveryPoints(cloudformationVaultName *string, recoveryPoints []types.RecoveryPointByCloudFormationVault) error {
// 	return fmt.Errorf("DeleteRecoveryPointsError")
// }

// func (m *allErrorMockCloudFormation) DeleteRecoveryPoint(cloudformationVaultName *string, recoveryPointArn *string) error {
// 	return fmt.Errorf("DeleteRecoveryPointError")
// }

// func (m *allErrorMockCloudFormation) DeleteCloudFormationVault(cloudformationVaultName *string) error {
// 	return fmt.Errorf("DeleteCloudFormationVaultError")
// }

// type listRecoveryPointsErrorMockCloudFormation struct{}

// func NewListRecoveryPointsErrorMockCloudFormation() *listRecoveryPointsErrorMockCloudFormation {
// 	return &listRecoveryPointsErrorMockCloudFormation{}
// }

// func (m *listRecoveryPointsErrorMockCloudFormation) ListRecoveryPointsByCloudFormationVault(cloudformationVaultName *string) ([]types.RecoveryPointByCloudFormationVault, error) {
// 	return nil, fmt.Errorf("ListRecoveryPointsByCloudFormationVaultError")
// }

// func (m *listRecoveryPointsErrorMockCloudFormation) DeleteRecoveryPoints(cloudformationVaultName *string, recoveryPoints []types.RecoveryPointByCloudFormationVault) error {
// 	return nil
// }

// func (m *listRecoveryPointsErrorMockCloudFormation) DeleteRecoveryPoint(cloudformationVaultName *string, recoveryPointArn *string) error {
// 	return nil
// }

// func (m *listRecoveryPointsErrorMockCloudFormation) DeleteCloudFormationVault(cloudformationVaultName *string) error {
// 	return nil
// }

// type deleteRecoveryPointsErrorMockCloudFormation struct{}

// func NewDeleteRecoveryPointsErrorMockCloudFormation() *deleteRecoveryPointsErrorMockCloudFormation {
// 	return &deleteRecoveryPointsErrorMockCloudFormation{}
// }

// func (m *deleteRecoveryPointsErrorMockCloudFormation) ListRecoveryPointsByCloudFormationVault(cloudformationVaultName *string) ([]types.RecoveryPointByCloudFormationVault, error) {
// 	return nil, nil
// }

// func (m *deleteRecoveryPointsErrorMockCloudFormation) DeleteRecoveryPoints(cloudformationVaultName *string, recoveryPoints []types.RecoveryPointByCloudFormationVault) error {
// 	return fmt.Errorf("DeleteRecoveryPointsError")
// }

// func (m *deleteRecoveryPointsErrorMockCloudFormation) DeleteRecoveryPoint(cloudformationVaultName *string, recoveryPointArn *string) error {
// 	return nil
// }

// func (m *deleteRecoveryPointsErrorMockCloudFormation) DeleteCloudFormationVault(cloudformationVaultName *string) error {
// 	return nil
// }

// type deleteCloudFormationVaultErrorMockCloudFormation struct{}

// func NewDeleteCloudFormationVaultErrorMockCloudFormation() *deleteCloudFormationVaultErrorMockCloudFormation {
// 	return &deleteCloudFormationVaultErrorMockCloudFormation{}
// }

// func (m *deleteCloudFormationVaultErrorMockCloudFormation) ListRecoveryPointsByCloudFormationVault(cloudformationVaultName *string) ([]types.RecoveryPointByCloudFormationVault, error) {
// 	return nil, nil
// }

// func (m *deleteCloudFormationVaultErrorMockCloudFormation) DeleteRecoveryPoints(cloudformationVaultName *string, recoveryPoints []types.RecoveryPointByCloudFormationVault) error {
// 	return nil
// }

// func (m *deleteCloudFormationVaultErrorMockCloudFormation) DeleteRecoveryPoint(cloudformationVaultName *string, recoveryPointArn *string) error {
// 	return nil
// }

// func (m *deleteCloudFormationVaultErrorMockCloudFormation) DeleteCloudFormationVault(cloudformationVaultName *string) error {
// 	return fmt.Errorf("DeleteCloudFormationVaultError")
// }

// /*
// 	Test Cases
// */
// func TestDeleteCloudFormationVault(t *testing.T) {
// 	logger.NewLogger()
// 	ctx := context.TODO()
// 	mock := NewMockCloudFormation()
// 	allErrorMock := NewAllErrorMockCloudFormation()
// 	listRecoveryPointsErrorMock := NewListRecoveryPointsErrorMockCloudFormation()
// 	deleteRecoveryPointsErrorMock := NewDeleteRecoveryPointsErrorMockCloudFormation()
// 	deleteCloudFormationVaultErrorMock := NewDeleteCloudFormationVaultErrorMockCloudFormation()

// 	type args struct {
// 		ctx                     context.Context
// 		cloudformationVaultName *string
// 		client                  client.ICloudFormation
// 	}

// 	cases := []struct {
// 		name    string
// 		args    args
// 		want    error
// 		wantErr bool
// 	}{
// 		{
// 			name: "delete cloudformation vault successfully",
// 			args: args{
// 				ctx:                     ctx,
// 				cloudformationVaultName: aws.String("test"),
// 				client:                  mock,
// 			},
// 			want:    nil,
// 			wantErr: false,
// 		},
// 		{
// 			name: "delete cloudformation vault failure for all errors",
// 			args: args{
// 				ctx:                     ctx,
// 				cloudformationVaultName: aws.String("test"),
// 				client:                  allErrorMock,
// 			},
// 			want:    fmt.Errorf("ListRecoveryPointsByCloudFormationVaultError"),
// 			wantErr: true,
// 		},
// 		{
// 			name: "delete cloudformation vault failure for list recovery points errors",
// 			args: args{
// 				ctx:                     ctx,
// 				cloudformationVaultName: aws.String("test"),
// 				client:                  listRecoveryPointsErrorMock,
// 			},
// 			want:    fmt.Errorf("ListRecoveryPointsByCloudFormationVaultError"),
// 			wantErr: true,
// 		},
// 		{
// 			name: "delete cloudformation vault failure for delete recovery points errors",
// 			args: args{
// 				ctx:                     ctx,
// 				cloudformationVaultName: aws.String("test"),
// 				client:                  deleteRecoveryPointsErrorMock,
// 			},
// 			want:    fmt.Errorf("DeleteRecoveryPointsError"),
// 			wantErr: true,
// 		},
// 		{
// 			name: "delete cloudformation vault failure for delete cloudformation vault errors",
// 			args: args{
// 				ctx:                     ctx,
// 				cloudformationVaultName: aws.String("test"),
// 				client:                  deleteCloudFormationVaultErrorMock,
// 			},
// 			want:    fmt.Errorf("DeleteCloudFormationVaultError"),
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range cases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cloudformationOperator := NewStackOperator(tt.args.client)

// 			err := cloudformationOperator.DeleteCloudFormationVault(tt.args.cloudformationVaultName)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
// 				return
// 			}
// 			if tt.wantErr && err.Error() != tt.want.Error() {
// 				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
// 				return
// 			}
// 		})
// 	}
// }

// func TestDeleteResourcesForCloudFormationVault(t *testing.T) {
// 	logger.NewLogger()
// 	ctx := context.TODO()
// 	mock := NewMockCloudFormation()
// 	allErrorMock := NewAllErrorMockCloudFormation()

// 	type args struct {
// 		ctx    context.Context
// 		client client.ICloudFormation
// 	}

// 	cases := []struct {
// 		name    string
// 		args    args
// 		want    error
// 		wantErr bool
// 	}{
// 		{
// 			name: "delete resources successfully",
// 			args: args{
// 				ctx:    ctx,
// 				client: mock,
// 			},
// 			want:    nil,
// 			wantErr: false,
// 		},
// 		{
// 			name: "delete resources failure",
// 			args: args{
// 				ctx:    ctx,
// 				client: allErrorMock,
// 			},
// 			want:    fmt.Errorf("ListRecoveryPointsByCloudFormationVaultError"),
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range cases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cloudformationOperator := NewStackOperator(tt.args.client)
// 			cloudformationOperator.AddResources(&types.StackResourceSummary{
// 				LogicalResourceId:  aws.String("LogicalResourceId1"),
// 				ResourceStatus:     "DELETE_FAILED",
// 				ResourceType:       aws.String("AWS::CloudFormation::CloudFormationVault"),
// 				PhysicalResourceId: aws.String("PhysicalResourceId1"),
// 			})

// 			err := cloudformationOperator.DeleteResources()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
// 				return
// 			}
// 			if tt.wantErr && err.Error() != tt.want.Error() {
// 				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
// 				return
// 			}
// 		})
// 	}
// }
