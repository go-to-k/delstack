package preprocessor

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
)

var _ IPreprocessor = (*LambdaVPCDetacher)(nil)

type LambdaVPCDetacher struct {
	lambdaClient client.ILambda
	ec2Client    client.IEC2
}

func NewLambdaVPCDetacher(lambdaClient client.ILambda, ec2Client client.IEC2) *LambdaVPCDetacher {
	return &LambdaVPCDetacher{
		lambdaClient: lambdaClient,
		ec2Client:    ec2Client,
	}
}

func (d *LambdaVPCDetacher) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	// Filter Lambda functions from the provided resources
	lambdaFunctions := FilterResourcesByType(resources, resourcetype.LambdaFunction)

	if len(lambdaFunctions) == 0 {
		return nil
	}

	io.Logger.Debug().Msgf("[%v]: Found %d Lambda function(s), checking VPC attachment", aws.ToString(stackName), len(lambdaFunctions))

	// Process all Lambda functions in parallel
	var wg sync.WaitGroup
	for _, resource := range lambdaFunctions {
		functionName := resource.PhysicalResourceId
		wg.Add(1)
		go func(name *string) {
			defer wg.Done()
			if err := d.detachVPCFromFunction(ctx, stackName, name); err != nil {
				io.Logger.Warn().Msgf("[%v]: Failed to detach VPC from function %s: %v",
					aws.ToString(stackName), aws.ToString(name), err)
			}
		}(functionName)
	}

	wg.Wait()

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

	// Clean up ENIs associated with this Lambda function
	if err := d.cleanupENIs(ctx, stackName, functionName); err != nil {
		io.Logger.Warn().Msgf("[%v]: Failed to clean up ENIs for function %s (continuing): %v",
			aws.ToString(stackName), aws.ToString(functionName), err)
	}

	return nil
}

func (d *LambdaVPCDetacher) cleanupENIs(ctx context.Context, stackName *string, functionName *string) error {
	// Find ENIs associated with this Lambda function
	// Lambda creates ENIs with description "AWS Lambda VPC ENI-<function-name>"
	// Note: We don't filter by status because immediately after VPC detach,
	// ENIs may still be "in-use" and transitioning to "available"
	filters := []ec2types.Filter{
		{
			Name:   aws.String("description"),
			Values: []string{"AWS Lambda VPC ENI-" + aws.ToString(functionName)},
		},
	}

	enis, err := d.ec2Client.DescribeNetworkInterfaces(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to describe ENIs: %w", err)
	}

	if len(enis) == 0 {
		io.Logger.Debug().Msgf("[%v]: No ENIs found for Lambda function %s",
			aws.ToString(stackName), aws.ToString(functionName))
		return nil
	}

	io.Logger.Debug().Msgf("[%v]: Found %d ENI(s) for Lambda function %s, deleting",
		aws.ToString(stackName), len(enis), aws.ToString(functionName))

	var wg sync.WaitGroup
	for _, eni := range enis {
		eniId := eni.NetworkInterfaceId
		wg.Add(1)
		go func(id *string) {
			defer wg.Done()
			if err := d.ec2Client.DeleteNetworkInterface(ctx, id); err != nil {
				io.Logger.Warn().Msgf("[%v]: Failed to delete ENI %s: %v",
					aws.ToString(stackName), aws.ToString(id), err)
				return
			}
			io.Logger.Debug().Msgf("[%v]: Deleted ENI %s",
				aws.ToString(stackName), aws.ToString(id))
		}(eniId)
	}

	wg.Wait()

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
