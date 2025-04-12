package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewIAMResources creates required IAM resources
func NewIAMResources(scope constructs.Construct, pjPrefix string) map[string]awscdk.IResource {
	resources := make(map[string]awscdk.IResource)

	// Create IAM group
	rootIamGroup := awsiam.NewGroup(scope, jsii.String("RootIamGroup"), &awsiam.GroupProps{
		GroupName: jsii.String(pjPrefix + "-root-group"),
	})
	resources["RootIamGroup"] = rootIamGroup

	// Create IAM role for Lambda
	rootLambdaRole := awsiam.NewRole(scope, jsii.String("RootLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-root-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	resources["RootLambdaRole"] = rootLambdaRole

	// Create a second IAM role for Lambda
	rootLambdaRole2 := awsiam.NewRole(scope, jsii.String("RootLambdaRole2"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-root-role-2"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	resources["RootLambdaRole2"] = rootLambdaRole2

	// Create Lambda policy and attach to both roles
	rootLambdaPolicy := awsiam.NewPolicy(scope, jsii.String("RootLambdaPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String(pjPrefix + "-root-policy"),
		Roles:      &[]awsiam.IRole{rootLambdaRole, rootLambdaRole2},
		Document: awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
			Statements: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Actions: jsii.Strings(
						"logs:CreateLogGroup",
						"logs:CreateLogStream",
						"logs:PutLogEvents",
						"logs:PutResourcePolicy",
						"logs:DeleteResourcePolicy",
					),
					Resources: jsii.Strings("*"),
				}),
			},
		}),
	})
	resources["RootLambdaPolicy"] = rootLambdaPolicy

	// Create service role for AWS Backup
	awsBackupServiceRole := awsiam.NewRole(scope, jsii.String("AWSBackupServiceRole"), &awsiam.RoleProps{
		RoleName:    jsii.String(pjPrefix + "-AWSBackupServiceRole"),
		Description: jsii.String("for AWS Backup"),
		Path:        jsii.String("/service-role/"),
		AssumedBy:   awsiam.NewServicePrincipal(jsii.String("backup.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSBackupServiceRolePolicyForBackup")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSBackupServiceRolePolicyForRestores")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBReadOnlyAccess")),
		},
	})
	resources["AWSBackupServiceRole"] = awsBackupServiceRole

	return resources
}
