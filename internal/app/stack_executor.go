package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/internal/preprocessor"
)

// IStackExecutor executes the deletion of a single CloudFormation stack.
type IStackExecutor interface {
	Execute(ctx context.Context, stack string, config aws.Config, operatorFactory *operation.OperatorFactory, forceMode bool, isRootStack bool) error
}

type StackExecutor struct{}

func (e *StackExecutor) Execute(
	ctx context.Context,
	stack string,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
	forceMode bool,
	isRootStack bool,
) error {
	operatorCollection := operation.NewOperatorCollection(config, operatorFactory)
	operatorManager := operation.NewOperatorManager(operatorCollection)
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()

	io.Logger.Info().Msgf("[%v]: Start deletion. Please wait a few minutes...", stack)

	if forceMode {
		if err := cloudformationStackOperator.RemoveDeletionPolicy(ctx, aws.String(stack)); err != nil {
			return fmt.Errorf("[%v]: Failed to remove deletion policy: %w", stack, err)
		}
	}

	pp := preprocessor.NewRecursivePreprocessorFromConfig(config, forceMode)
	if err := pp.PreprocessRecursively(ctx, aws.String(stack)); err != nil {
		return fmt.Errorf("[%v]: %w", stack, err)
	}

	if err := cloudformationStackOperator.DeleteCloudFormationStack(ctx, aws.String(stack), isRootStack, operatorManager); err != nil {
		return fmt.Errorf("[%v]: Failed to delete: %w", stack, err)
	}

	io.Logger.Info().Msgf("[%v]: Successfully deleted!!", stack)
	return nil
}
