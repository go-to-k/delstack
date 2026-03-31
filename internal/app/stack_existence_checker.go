package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
)

// IStackExistenceChecker checks whether a stack exists in AWS.
type IStackExistenceChecker interface {
	Check(ctx context.Context, region, stackName string) (operation.StackCheckResult, error)
}

type StackExistenceChecker struct {
	profile   string
	forceMode bool
	// cache operators per region
	operatorCache map[string]*operation.CloudFormationStackOperator
}

func NewStackExistenceChecker(profile string, forceMode bool) *StackExistenceChecker {
	return &StackExistenceChecker{
		profile:       profile,
		forceMode:     forceMode,
		operatorCache: make(map[string]*operation.CloudFormationStackOperator),
	}
}

func (c *StackExistenceChecker) Check(ctx context.Context, region, stackName string) (operation.StackCheckResult, error) {
	op, ok := c.operatorCache[region]
	if !ok {
		cfg, err := client.LoadAWSConfig(ctx, region, c.profile)
		if err != nil {
			return operation.StackCheckResult{}, fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
		}
		factory := operation.NewOperatorFactory(cfg, c.forceMode)
		op = factory.CreateCloudFormationStackOperator()
		c.operatorCache[region] = op
	}

	return op.CheckStack(ctx, aws.String(stackName))
}
