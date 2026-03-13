package preprocessor

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
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

	var wg sync.WaitGroup

	// Process current stack's resources with preprocessor (in parallel with nested stacks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.pp.Preprocess(ctx, stackName, resources); err != nil {
			io.Logger.Warn().Msgf("[%v]: Failed to preprocess stack: %v", aws.ToString(stackName), err)
		}
	}()

	// Process nested stacks recursively in parallel
	for _, nestedStack := range nestedStacks {
		nestedStackName := nestedStack.PhysicalResourceId
		wg.Add(1)
		go func(name *string) {
			defer wg.Done()
			io.Logger.Debug().Msgf("[%v]: Processing nested stack %s", aws.ToString(stackName), aws.ToString(name))
			if err := r.PreprocessRecursively(ctx, name); err != nil {
				io.Logger.Warn().Msgf("[%v]: Failed to preprocess nested stack %s: %v",
					aws.ToString(stackName), aws.ToString(name), err)
			}
		}(nestedStackName)
	}

	// Wait for all preprocessing (current stack + nested stacks) to complete
	wg.Wait()

	return nil
}
