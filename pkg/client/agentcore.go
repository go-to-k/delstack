//go:generate mockgen -source=$GOFILE -destination=agentcore_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
)

const (
	AgentRuntimeUpdateWaitTime    = 5 * time.Minute
	AgentRuntimeUpdatePollInterval = 5 * time.Second
)

type IAgentCore interface {
	GetAgentRuntime(ctx context.Context, runtimeId *string) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error)
	UpdateAgentRuntime(ctx context.Context, input *bedrockagentcorecontrol.UpdateAgentRuntimeInput) error
}

var _ IAgentCore = (*AgentCoreClient)(nil)

type AgentCoreClient struct {
	client *bedrockagentcorecontrol.Client
}

func NewAgentCoreClient(client *bedrockagentcorecontrol.Client) *AgentCoreClient {
	return &AgentCoreClient{
		client: client,
	}
}

func (c *AgentCoreClient) GetAgentRuntime(ctx context.Context, runtimeId *string) (*bedrockagentcorecontrol.GetAgentRuntimeOutput, error) {
	output, err := c.client.GetAgentRuntime(ctx, &bedrockagentcorecontrol.GetAgentRuntimeInput{
		AgentRuntimeId: runtimeId,
	})
	if err != nil {
		return nil, &ClientError{
			ResourceName: runtimeId,
			Err:          err,
		}
	}
	return output, nil
}

func (c *AgentCoreClient) UpdateAgentRuntime(ctx context.Context, input *bedrockagentcorecontrol.UpdateAgentRuntimeInput) error {
	_, err := c.client.UpdateAgentRuntime(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: input.AgentRuntimeId,
			Err:          err,
		}
	}

	if err := c.waitForRuntimeUpdated(ctx, input.AgentRuntimeId); err != nil {
		return &ClientError{
			ResourceName: input.AgentRuntimeId,
			Err:          err,
		}
	}

	return nil
}

func (c *AgentCoreClient) waitForRuntimeUpdated(ctx context.Context, runtimeId *string) error {
	startTime := time.Now()
	for {
		output, err := c.client.GetAgentRuntime(ctx, &bedrockagentcorecontrol.GetAgentRuntimeInput{
			AgentRuntimeId: runtimeId,
		})
		if err != nil {
			return fmt.Errorf("failed to poll runtime status: %w", err)
		}

		switch output.Status {
		case types.AgentRuntimeStatusReady:
			return nil
		case types.AgentRuntimeStatusCreateFailed, types.AgentRuntimeStatusUpdateFailed:
			// Ignore failure states to continue the deletion process
			return nil
		}

		if time.Since(startTime) >= AgentRuntimeUpdateWaitTime {
			return fmt.Errorf("timeout waiting for runtime update to complete (status: %s)", output.Status)
		}

		time.Sleep(AgentRuntimeUpdatePollInterval)
	}
}
