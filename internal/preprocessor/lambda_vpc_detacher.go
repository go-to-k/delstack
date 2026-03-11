package preprocessor

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
)

const (
	LambdaFunction = "AWS::Lambda::Function"
)

var _ IPreprocessor = (*LambdaVPCDetacher)(nil)

type LambdaVPCDetacher struct {
	lambdaClient client.ILambda
	cfnClient    client.ICloudFormation
}

func NewLambdaVPCDetacher(lambdaClient client.ILambda, cfnClient client.ICloudFormation) *LambdaVPCDetacher {
	return &LambdaVPCDetacher{
		lambdaClient: lambdaClient,
		cfnClient:    cfnClient,
	}
}

func (d *LambdaVPCDetacher) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	return d.preprocessStack(ctx, stackName, resources)
}

func (d *LambdaVPCDetacher) preprocessStack(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	if len(resources) == 0 {
		var err error
		resources, err = d.cfnClient.ListStackResources(ctx, stackName)
		if err != nil {
			return err
		}
	}

	// Process Lambda functions in this stack
	lambdaFunctions := FilterResourcesByType(resources, LambdaFunction)
	if len(lambdaFunctions) > 0 {
		io.Logger.Debug().Msgf("[%v]: Found %d Lambda function(s), checking VPC attachment", aws.ToString(stackName), len(lambdaFunctions))

		for _, resource := range lambdaFunctions {
			functionName := resource.PhysicalResourceId
			if err := d.detachVPCFromFunction(ctx, stackName, functionName); err != nil {
				io.Logger.Warn().Msgf("[%v]: Failed to detach VPC from function %s: %v",
					aws.ToString(stackName), aws.ToString(functionName), err)
				continue
			}
		}
	}

	// Process nested stacks recursively
	nestedStacks := FilterResourcesByType(resources, "AWS::CloudFormation::Stack")
	for _, nestedStack := range nestedStacks {
		nestedStackName := nestedStack.PhysicalResourceId
		io.Logger.Debug().Msgf("[%v]: Processing nested stack %s", aws.ToString(stackName), aws.ToString(nestedStackName))
		if err := d.preprocessStack(ctx, nestedStackName, nil); err != nil {
			io.Logger.Warn().Msgf("[%v]: Failed to process nested stack %s: %v",
				aws.ToString(stackName), aws.ToString(nestedStackName), err)
			continue
		}
	}

	return nil
}

func (d *LambdaVPCDetacher) detachVPCFromFunction(ctx context.Context, stackName *string, functionName *string) error {
	output, err := d.lambdaClient.GetFunction(ctx, functionName)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	if !d.isAttachedToVPC(output) {
		return nil
	}

	vpcId := aws.ToString(output.Configuration.VpcConfig.VpcId)
	isIPv6 := d.isIPv6Enabled(output)

	if isIPv6 {
		io.Logger.Debug().Msgf("[%v]: Lambda function %s is attached to VPC %s with IPv6 enabled, detaching",
			aws.ToString(stackName), aws.ToString(functionName), vpcId)
	} else {
		io.Logger.Debug().Msgf("[%v]: Lambda function %s is attached to VPC %s, detaching",
			aws.ToString(stackName), aws.ToString(functionName), vpcId)
	}

	// Remove VPC configuration (and disable IPv6 if enabled)
	vpcConfig := &lambdatypes.VpcConfig{
		SubnetIds:        []string{},
		SecurityGroupIds: []string{},
	}
	if isIPv6 {
		vpcConfig.Ipv6AllowedForDualStack = aws.Bool(false)
	}

	io.Logger.Debug().Msgf("[%v]: Removing VPC configuration from Lambda function %s",
		aws.ToString(stackName), aws.ToString(functionName))
	err = d.lambdaClient.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: functionName,
		VpcConfig:    vpcConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to remove VPC config: %w", err)
	}

	io.Logger.Debug().Msgf("[%v]: Lambda function %s VPC detached successfully",
		aws.ToString(stackName), aws.ToString(functionName))
	return nil
}

func (d *LambdaVPCDetacher) isAttachedToVPC(output *lambda.GetFunctionOutput) bool {
	if output.Configuration == nil || output.Configuration.VpcConfig == nil {
		return false
	}

	return output.Configuration.VpcConfig.VpcId != nil &&
		*output.Configuration.VpcConfig.VpcId != ""
}

func (d *LambdaVPCDetacher) isIPv6Enabled(output *lambda.GetFunctionOutput) bool {
	if output.Configuration == nil || output.Configuration.VpcConfig == nil {
		return false
	}

	return output.Configuration.VpcConfig.Ipv6AllowedForDualStack != nil &&
		*output.Configuration.VpcConfig.Ipv6AllowedForDualStack
}
