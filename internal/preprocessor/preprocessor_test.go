package preprocessor

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func TestFilterResourcesByType(t *testing.T) {
	type args struct {
		resources    []types.StackResourceSummary
		resourceType string
	}

	cases := []struct {
		name string
		args args
		want int
	}{
		{
			name: "empty resources",
			args: args{
				resources:    []types.StackResourceSummary{},
				resourceType: "AWS::Lambda::Function",
			},
			want: 0,
		},
		{
			name: "single matching resource",
			args: args{
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function"),
					},
				},
				resourceType: "AWS::Lambda::Function",
			},
			want: 1,
		},
		{
			name: "multiple matching resources",
			args: args{
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function-1"),
					},
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("test-bucket"),
					},
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function-2"),
					},
				},
				resourceType: "AWS::Lambda::Function",
			},
			want: 2,
		},
		{
			name: "no matching resources",
			args: args{
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("test-bucket"),
					},
					{
						ResourceType:       aws.String("AWS::IAM::Role"),
						PhysicalResourceId: aws.String("test-role"),
					},
				},
				resourceType: "AWS::Lambda::Function",
			},
			want: 0,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterResourcesByType(tt.args.resources, tt.args.resourceType)
			if len(got) != tt.want {
				t.Errorf("FilterResourcesByType() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
