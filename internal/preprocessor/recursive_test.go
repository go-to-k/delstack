package preprocessor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestRecursivePreprocessor_PreprocessRecursively(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	io.Logger = &logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		name    string
		setup   func(*client.MockICloudFormation, *mockPreprocessor)
		wantErr bool
	}{
		{
			name: "no resources",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{}, nil,
				)
			},
			wantErr: false,
		},
		{
			name: "list resources error",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					nil, fmt.Errorf("list error"),
				)
			},
			wantErr: true,
		},
		{
			name: "resources without nested stacks",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::Lambda::Function"),
							PhysicalResourceId: aws.String("test-function"),
						},
					}, nil,
				)
			},
			wantErr: false,
		},
		{
			name: "with nested stack",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("nested-stack"),
						},
					}, nil,
				)
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("nested-stack")).Return(
					[]types.StackResourceSummary{}, nil,
				)
			},
			wantErr: false,
		},
		{
			name: "nested stack with DELETE_COMPLETE is skipped",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("deleted-stack"),
							ResourceStatus:     types.ResourceStatusDeleteComplete,
						},
					}, nil,
				)
			},
			wantErr: false,
		},
		{
			name: "preprocessor failure is propagated as error",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				pp.err = fmt.Errorf("preprocess failed")
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::Lambda::Function"),
							PhysicalResourceId: aws.String("test-function"),
						},
					}, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "nested stack preprocessor failure propagates error",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				pp.err = fmt.Errorf("nested preprocess failed")
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("nested-stack"),
						},
					}, nil,
				)
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("nested-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::Lambda::Function"),
							PhysicalResourceId: aws.String("nested-function"),
						},
					}, nil,
				)
			},
			wantErr: true,
		},
		{
			name: "nested stack list resources error propagates",
			setup: func(cfn *client.MockICloudFormation, pp *mockPreprocessor) {
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("test-stack")).Return(
					[]types.StackResourceSummary{
						{
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("nested-stack"),
						},
					}, nil,
				)
				cfn.EXPECT().ListStackResources(gomock.Any(), aws.String("nested-stack")).Return(
					nil, fmt.Errorf("nested list error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockCfn := client.NewMockICloudFormation(ctrl)
			pp := &mockPreprocessor{}
			tt.setup(mockCfn, pp)

			r := NewRecursivePreprocessor(mockCfn, pp)
			err := r.PreprocessRecursively(context.Background(), aws.String("test-stack"))

			if (err != nil) != tt.wantErr {
				t.Errorf("PreprocessRecursively() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
