package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewAgentCoreRuntimeVpcAttached creates a Bedrock AgentCore Runtime attached to VPC.
// NOTE: Deploying this resource requires a container image in ECR.
// Set the CDK context variable "AGENTCORE_CONTAINER_URI" to the container image URI
// (e.g., "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-agent:latest").
func NewAgentCoreRuntimeVpcAttached(scope constructs.Construct, pjPrefix string, vpc awsec2.Vpc) {
	containerUri := scope.Node().TryGetContext(jsii.String("AGENTCORE_CONTAINER_URI"))
	if containerUri == nil {
		return // Skip if container URI is not provided
	}

	role := awsiam.NewRole(scope, jsii.String("AgentCoreRuntimeRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-AgentCoreRuntimeRole"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("bedrock.amazonaws.com"), nil),
	})

	sg := awsec2.NewSecurityGroup(scope, jsii.String("AgentCoreSG"), &awsec2.SecurityGroupProps{
		Vpc:               vpc,
		SecurityGroupName: jsii.String(pjPrefix + "-AgentCoreSG"),
		Description:       jsii.String("Security group for AgentCore Runtime"),
	})

	subnets := vpc.SelectSubnets(&awsec2.SubnetSelection{
		SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
	})

	// AWS::BedrockAgentCore::Runtime (L1 construct, no L2 available yet)
	awscdk.NewCfnResource(scope, jsii.String("AgentCoreRuntime"), &awscdk.CfnResourceProps{
		Type: jsii.String("AWS::BedrockAgentCore::Runtime"),
		Properties: &map[string]interface{}{
			"AgentRuntimeName": pjPrefix + "-AgentCoreRuntime",
			"RoleArn":          role.RoleArn(),
			"AgentRuntimeArtifact": map[string]interface{}{
				"ContainerConfiguration": map[string]interface{}{
					"ContainerUri": containerUri.(string),
				},
			},
			"NetworkConfiguration": map[string]interface{}{
				"NetworkMode": "VPC",
				"NetworkModeConfig": map[string]interface{}{
					"SecurityGroups": []*string{sg.SecurityGroupId()},
					"Subnets":        subnets.SubnetIds,
				},
			},
		},
	})
}
