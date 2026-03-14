package preprocessor

import (
	"context"
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
)

var _ IPreprocessor = (*CompositePreprocessor)(nil)

type CompositePreprocessor struct {
	checkers  []IPreprocessor
	modifiers []IPreprocessor
}

func NewCompositePreprocessor(checkers []IPreprocessor, modifiers []IPreprocessor) *CompositePreprocessor {
	return &CompositePreprocessor{
		checkers:  checkers,
		modifiers: modifiers,
	}
}

func (c *CompositePreprocessor) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	// Phase 1: Run checkers in parallel, collect errors
	// Checker errors are fatal and abort the process
	if err := c.runCheckers(ctx, stackName, resources); err != nil {
		return err
	}

	// Phase 2: Run modifiers in parallel, log warnings
	// Modifier errors are logged but not returned
	c.runModifiers(ctx, stackName, resources)

	return nil
}

func (c *CompositePreprocessor) runCheckers(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	if len(c.checkers) == 0 {
		return nil
	}

	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup

	for _, checker := range c.checkers {
		wg.Add(1)
		go func(ch IPreprocessor) {
			defer wg.Done()
			if err := ch.Preprocess(ctx, stackName, resources); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(checker)
	}
	wg.Wait()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (c *CompositePreprocessor) runModifiers(ctx context.Context, stackName *string, resources []types.StackResourceSummary) {
	if len(c.modifiers) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, modifier := range c.modifiers {
		wg.Add(1)
		go func(m IPreprocessor) {
			defer wg.Done()
			if err := m.Preprocess(ctx, stackName, resources); err != nil {
				io.Logger.Warn().Msgf("[%v]: Preprocessor failed: %v", aws.ToString(stackName), err)
			}
		}(modifier)
	}
	wg.Wait()
}
