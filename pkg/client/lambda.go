//go:generate mockgen -source=$GOFILE -destination=lambda_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

const LambdaFunctionUpdatedWaitNanoSecTime = time.Duration(300000000000) // 5 minutes

type ILambda interface {
	GetFunction(ctx context.Context, functionName *string) (*lambda.GetFunctionOutput, error)
	UpdateFunctionConfiguration(ctx context.Context, input *lambda.UpdateFunctionConfigurationInput) error
	DeleteFunction(ctx context.Context, functionName *string) error
	CheckLambdaFunctionExists(ctx context.Context, functionName *string) (bool, error)
}

var _ ILambda = (*LambdaClient)(nil)

type LambdaClient struct {
	client                *lambda.Client
	functionUpdatedWaiter *lambda.FunctionUpdatedV2Waiter
}

func NewLambdaClient(client *lambda.Client, functionUpdatedWaiter *lambda.FunctionUpdatedV2Waiter) *LambdaClient {
	return &LambdaClient{
		client,
		functionUpdatedWaiter,
	}
}

func (c *LambdaClient) GetFunction(ctx context.Context, functionName *string) (*lambda.GetFunctionOutput, error) {
	output, err := c.client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: functionName,
	})
	if err != nil {
		return nil, &ClientError{
			ResourceName: functionName,
			Err:          err,
		}
	}
	return output, nil
}

func (c *LambdaClient) UpdateFunctionConfiguration(ctx context.Context, input *lambda.UpdateFunctionConfigurationInput) error {
	_, err := c.client.UpdateFunctionConfiguration(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: input.FunctionName,
			Err:          err,
		}
	}

	if err := c.waitForFunctionUpdated(ctx, input.FunctionName); err != nil {
		return &ClientError{
			ResourceName: input.FunctionName,
			Err:          err,
		}
	}

	return nil
}

func (c *LambdaClient) DeleteFunction(ctx context.Context, functionName *string) error {
	_, err := c.client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: functionName,
	})
	if err != nil {
		return &ClientError{
			ResourceName: functionName,
			Err:          err,
		}
	}
	return nil
}

func (c *LambdaClient) CheckLambdaFunctionExists(ctx context.Context, functionName *string) (bool, error) {
	_, err := c.client.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: functionName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Function not found") {
			return false, nil
		}
		return false, &ClientError{
			ResourceName: functionName,
			Err:          err,
		}
	}
	return true, nil
}

func (c *LambdaClient) waitForFunctionUpdated(ctx context.Context, functionName *string) error {
	input := &lambda.GetFunctionInput{
		FunctionName: functionName,
	}

	err := c.functionUpdatedWaiter.Wait(ctx, input, LambdaFunctionUpdatedWaitNanoSecTime)
	if err != nil {
		// Waiter returns "waiter state transitioned to Failure" when the function update fails,
		// but we want to ignore this error to continue the deletion process
		if strings.Contains(err.Error(), "waiter state transitioned to Failure") {
			return nil
		}
		return err // return non wrapping error because wrap in public callers
	}

	return nil
}
