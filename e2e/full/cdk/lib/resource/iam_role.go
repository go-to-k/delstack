package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewIamRole creates an EC2 role. The role itself does not block deletion; deploy.go adds it to
// an instance profile created outside of CFn to trigger DELETE_FAILED (issue #662).
func NewIamRole(scope constructs.Construct) {
	awsiam.NewRole(scope, jsii.String("IamRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"DelstackTestRoleInlinePolicy": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Effect:    awsiam.Effect_ALLOW,
						Actions:   jsii.Strings("s3:GetObject"),
						Resources: jsii.Strings("*"),
					}),
				},
			}),
		},
	})
}
