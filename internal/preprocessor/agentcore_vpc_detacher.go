package preprocessor

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
)

var _ IPreprocessor = (*AgentCoreVPCDetacher)(nil)

type AgentCoreVPCDetacher struct {
	agentCoreClient client.IAgentCore
	ec2Client       client.IEC2
}

func NewAgentCoreVPCDetacher(agentCoreClient client.IAgentCore, ec2Client client.IEC2) *AgentCoreVPCDetacher {
	return &AgentCoreVPCDetacher{
		agentCoreClient: agentCoreClient,
		ec2Client:       ec2Client,
	}
}

func (d *AgentCoreVPCDetacher) Preprocess(ctx context.Context, stackName *string, resources []cfntypes.StackResourceSummary) error {
	runtimes := FilterResourcesByType(resources, resourcetype.BedrockAgentCoreRuntime)

	if len(runtimes) == 0 {
		return nil
	}

	io.Logger.Debug().Msgf("[%v]: Found %d AgentCore Runtime(s), checking VPC attachment", aws.ToString(stackName), len(runtimes))

	var wg sync.WaitGroup
	for _, resource := range runtimes {
		runtimeId := resource.PhysicalResourceId
		wg.Add(1)
		go func(id *string) {
			defer wg.Done()
			if err := d.detachVPCFromRuntime(ctx, stackName, id); err != nil {
				io.Logger.Warn().Msgf("[%v]: Failed to detach VPC from AgentCore Runtime %s: %v",
					aws.ToString(stackName), aws.ToString(id), err)
			}
		}(runtimeId)
	}

	wg.Wait()

	return nil
}

func (d *AgentCoreVPCDetacher) detachVPCFromRuntime(ctx context.Context, stackName *string, runtimeId *string) error {
	output, err := d.agentCoreClient.GetAgentRuntime(ctx, runtimeId)
	if err != nil {
		return fmt.Errorf("failed to get runtime: %w", err)
	}

	if !d.isAttachedToVPC(output) {
		return nil
	}

	sgIDs := d.getSecurityGroupIDs(output)

	io.Logger.Debug().Msgf("[%v]: AgentCore Runtime %s is attached to VPC, detaching",
		aws.ToString(stackName), aws.ToString(runtimeId))

	// Update network mode from VPC to PUBLIC to detach VPC
	err = d.agentCoreClient.UpdateAgentRuntime(ctx, &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
		AgentRuntimeId:       runtimeId,
		AgentRuntimeArtifact: output.AgentRuntimeArtifact,
		RoleArn:              output.RoleArn,
		NetworkConfiguration: &types.NetworkConfiguration{
			NetworkMode: types.NetworkModePublic,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update network configuration: %w", err)
	}

	io.Logger.Debug().Msgf("[%v]: AgentCore Runtime %s VPC detached successfully",
		aws.ToString(stackName), aws.ToString(runtimeId))

	// Clean up ENIs associated with this runtime
	if err := d.cleanupENIs(ctx, stackName, runtimeId, sgIDs); err != nil {
		io.Logger.Warn().Msgf("[%v]: Failed to clean up ENIs for AgentCore Runtime %s (continuing): %v",
			aws.ToString(stackName), aws.ToString(runtimeId), err)
	}

	return nil
}

func (d *AgentCoreVPCDetacher) cleanupENIs(ctx context.Context, stackName *string, runtimeId *string, sgIDs []string) error {
	if len(sgIDs) == 0 {
		return nil
	}

	// Find ENIs by interface-type=agentic_ai and matching security groups
	filters := []ec2types.Filter{
		{
			Name:   aws.String("interface-type"),
			Values: []string{"agentic_ai"},
		},
		{
			Name:   aws.String("group-id"),
			Values: sgIDs,
		},
	}

	enis, err := d.ec2Client.DescribeNetworkInterfaces(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to describe ENIs: %w", err)
	}

	if len(enis) == 0 {
		io.Logger.Debug().Msgf("[%v]: No ENIs found for AgentCore Runtime %s",
			aws.ToString(stackName), aws.ToString(runtimeId))
		return nil
	}

	io.Logger.Debug().Msgf("[%v]: Found %d ENI(s) for AgentCore Runtime %s, deleting",
		aws.ToString(stackName), len(enis), aws.ToString(runtimeId))

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

func (d *AgentCoreVPCDetacher) isAttachedToVPC(output *bedrockagentcorecontrol.GetAgentRuntimeOutput) bool {
	if output.NetworkConfiguration == nil {
		return false
	}

	return output.NetworkConfiguration.NetworkMode == types.NetworkModeVpc
}

func (d *AgentCoreVPCDetacher) getSecurityGroupIDs(output *bedrockagentcorecontrol.GetAgentRuntimeOutput) []string {
	if output.NetworkConfiguration == nil || output.NetworkConfiguration.NetworkModeConfig == nil {
		return nil
	}

	return output.NetworkConfiguration.NetworkModeConfig.SecurityGroups
}
