package preprocessor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/rs/zerolog"
)

type mockPreprocessor struct {
	err error
}

func (m *mockPreprocessor) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	return m.err
}

func TestCompositePreprocessor_Preprocess(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	io.Logger = &logger

	resources := []types.StackResourceSummary{
		{
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("test-bucket"),
		},
	}

	cases := []struct {
		name          string
		preprocessors []IPreprocessor
		wantErr       bool
	}{
		{
			name:          "empty list",
			preprocessors: []IPreprocessor{},
			wantErr:       false,
		},
		{
			name: "single preprocessor success",
			preprocessors: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			wantErr: false,
		},
		{
			name: "multiple preprocessors success",
			preprocessors: []IPreprocessor{
				&mockPreprocessor{err: nil},
				&mockPreprocessor{err: nil},
			},
			wantErr: false,
		},
		{
			name: "one failure does not affect others",
			preprocessors: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("preprocessor failed")},
				&mockPreprocessor{err: nil},
			},
			wantErr: false,
		},
		{
			name: "all failures still returns nil",
			preprocessors: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("preprocessor 1 failed")},
				&mockPreprocessor{err: fmt.Errorf("preprocessor 2 failed")},
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			composite := NewCompositePreprocessor(tt.preprocessors...)
			err := composite.Preprocess(context.Background(), aws.String("test-stack"), resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
