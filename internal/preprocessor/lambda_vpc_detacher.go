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
	if len(resources) == 0 {
		var err error
		resources, err = d.cfnClient.ListStackResources(ctx, stackName)
		if err != nil {
			return err
		}
	}

	lambdaFunctions := FilterResourcesByType(resources, LambdaFunction)
	if len(lambdaFunctions) == 0 {
		return nil
	}

	for _, resource := range lambdaFunctions {
		functionName := resource.PhysicalResourceId
		if err := d.detachVPCFromFunction(ctx, functionName); err != nil {
			io.Logger.Warn().Msgf("[%v]: Failed to detach VPC from function %s: %v",
				aws.ToString(stackName), aws.ToString(functionName), err)
			continue
		}
	}

	return nil
}

func (d *LambdaVPCDetacher) detachVPCFromFunction(ctx context.Context, functionName *string) error {
	output, err := d.lambdaClient.GetFunction(ctx, functionName)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	if !d.isAttachedToVPC(output) {
		return nil
	}

	if d.isIPv6Enabled(output) {
		err := d.lambdaClient.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
			FunctionName: functionName,
			VpcConfig: &lambdatypes.VpcConfig{
				Ipv6AllowedForDualStack: aws.Bool(false),
			},
		})
		if err != nil {
			return fmt.Errorf("failed to disable IPv6: %w", err)
		}
	}

	err = d.lambdaClient.UpdateFunctionConfiguration(ctx, &lambda.UpdateFunctionConfigurationInput{
		FunctionName: functionName,
		VpcConfig: &lambdatypes.VpcConfig{
			SubnetIds:        []string{},
			SecurityGroupIds: []string{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to remove VPC config: %w", err)
	}

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
