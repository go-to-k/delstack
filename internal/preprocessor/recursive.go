package preprocessor

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
)

type RecursivePreprocessor struct {
	cfnClient client.ICloudFormation
	pp        IPreprocessor
}

func NewRecursivePreprocessor(cfnClient client.ICloudFormation, pp IPreprocessor) *RecursivePreprocessor {
	return &RecursivePreprocessor{
		cfnClient: cfnClient,
		pp:        pp,
	}
}

func (r *RecursivePreprocessor) PreprocessRecursively(ctx context.Context, stackName *string) error {
	resources, err := r.cfnClient.ListStackResources(ctx, stackName)
	if err != nil {
		return fmt.Errorf("failed to list stack resources: %w", err)
	}

	nestedStacks := FilterResourcesByType(resources, resourcetype.CloudformationStack)

	eg, ctx := errgroup.WithContext(ctx)

	// Process current stack's resources with preprocessor
	eg.Go(func() error {
		return r.pp.Preprocess(ctx, stackName, resources)
	})

	// Process nested stacks recursively in parallel
	for _, nestedStack := range nestedStacks {
		nestedStackName := nestedStack.PhysicalResourceId
		eg.Go(func() error {
			io.Logger.Debug().Msgf("[%v]: Processing nested stack %s", aws.ToString(stackName), aws.ToString(nestedStackName))
			return r.PreprocessRecursively(ctx, nestedStackName)
		})
	}

	return eg.Wait()
}
