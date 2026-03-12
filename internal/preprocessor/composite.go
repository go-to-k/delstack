package preprocessor

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
)

var _ IPreprocessor = (*CompositePreprocessor)(nil)

type CompositePreprocessor struct {
	preprocessors []IPreprocessor
}

func NewCompositePreprocessor(preprocessors ...IPreprocessor) *CompositePreprocessor {
	return &CompositePreprocessor{
		preprocessors: preprocessors,
	}
}

func (c *CompositePreprocessor) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	var wg sync.WaitGroup
	for _, p := range c.preprocessors {
		wg.Add(1)
		go func(pp IPreprocessor) {
			defer wg.Done()
			if err := pp.Preprocess(ctx, stackName, resources); err != nil {
				io.Logger.Warn().Msgf("[%v]: Preprocessor failed: %v", aws.ToString(stackName), err)
			}
		}(p)
	}
	wg.Wait()
	return nil
}
