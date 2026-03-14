package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewLogGroup creates a CloudWatch LogGroup with deletion protection enabled.
// Note: CDK v2.176.0 does not support DeletionProtectionEnabled in LogGroupProps or
// CfnLogGroupProps directly. We use AddPropertyOverride on the underlying CfnLogGroup
// to set the DeletionProtectionEnabled property at the CloudFormation level.
func NewLogGroup(scope constructs.Construct) awslogs.LogGroup {
	logGroup := awslogs.NewLogGroup(scope, jsii.String("LogGroup"), &awslogs.LogGroupProps{
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     awslogs.RetentionDays_ONE_DAY,
	})

	// Enable deletion protection via CloudFormation property override
	cfnLogGroup := logGroup.Node().DefaultChild().(awscdk.CfnResource)
	cfnLogGroup.AddPropertyOverride(jsii.String("DeletionProtectionEnabled"), jsii.Bool(true))

	return logGroup
}
