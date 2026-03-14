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
		name      string
		checkers  []IPreprocessor
		modifiers []IPreprocessor
		wantErr   bool
	}{
		{
			name:      "empty lists",
			checkers:  []IPreprocessor{},
			modifiers: []IPreprocessor{},
			wantErr:   false,
		},
		{
			name: "single checker success",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			modifiers: []IPreprocessor{},
			wantErr:   false,
		},
		{
			name:     "single modifier success",
			checkers: []IPreprocessor{},
			modifiers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			wantErr: false,
		},
		{
			name: "checker failure returns error",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("checker failed")},
			},
			modifiers: []IPreprocessor{},
			wantErr:   true,
		},
		{
			name: "multiple checker failures returns combined error",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("checker 1 failed")},
				&mockPreprocessor{err: fmt.Errorf("checker 2 failed")},
			},
			modifiers: []IPreprocessor{},
			wantErr:   true,
		},
		{
			name:     "modifier failure does not return error",
			checkers: []IPreprocessor{},
			modifiers: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("modifier failed")},
			},
			wantErr: false,
		},
		{
			name: "checker failure prevents modifiers from running",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("checker failed")},
			},
			modifiers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			wantErr: true,
		},
		{
			name: "both checkers and modifiers succeed",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			modifiers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			wantErr: false,
		},
		{
			name: "checker success then modifier failure returns no error",
			checkers: []IPreprocessor{
				&mockPreprocessor{err: nil},
			},
			modifiers: []IPreprocessor{
				&mockPreprocessor{err: fmt.Errorf("modifier failed")},
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			composite := NewCompositePreprocessor(tt.checkers, tt.modifiers)
			err := composite.Preprocess(context.Background(), aws.String("test-stack"), resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
