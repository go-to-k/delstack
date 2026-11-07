package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3files"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewClientStack creates a VPC with a public subnet and an EC2 instance that
// deploy.go drives via SSM to mount the file system and keep writing to it.
func NewClientStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(1),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	role := awsiam.NewRole(stack, jsii.String("InstanceRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ec2.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonSSMManagedInstanceCore")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonS3FilesClientFullAccess")),
		},
	})

	userData := awsec2.UserData_ForLinux(nil)
	userData.AddCommands(
		jsii.String("dnf install -y amazon-efs-utils python3-pip"),
		jsii.String("python3 -m pip install botocore"),
	)

	instance := awsec2.NewInstance(stack, jsii.String("Instance"), &awsec2.InstanceProps{
		Vpc:          vpc,
		VpcSubnets:   &awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC},
		InstanceType: awsec2.NewInstanceType(jsii.String("t4g.small")),
		MachineImage: awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{
			CpuType: awsec2.AmazonLinuxCpuType_ARM_64,
		}),
		Role:     role,
		UserData: userData,
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcId"), &awscdk.CfnOutputProps{Value: vpc.VpcId()})
	awscdk.NewCfnOutput(stack, jsii.String("VpcCidr"), &awscdk.CfnOutputProps{Value: vpc.VpcCidrBlock()})
	awscdk.NewCfnOutput(stack, jsii.String("SubnetId"), &awscdk.CfnOutputProps{Value: (*vpc.PublicSubnets())[0].SubnetId()})
	awscdk.NewCfnOutput(stack, jsii.String("InstanceId"), &awscdk.CfnOutputProps{Value: instance.InstanceId()})

	return stack
}

// NewTargetStack creates the stack delstack deletes: a versioned bucket, an S3
// Files file system on it, and a mount target in the client stack's subnet.
func NewTargetStack(scope constructs.Construct, id string, props *awscdk.StackProps, vpcId, subnetId, vpcCidr string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		Versioned:     jsii.Bool(true),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	role := awsiam.NewRole(stack, jsii.String("FileSystemRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("elasticfilesystem.amazonaws.com"), &awsiam.ServicePrincipalOpts{
			Conditions: &map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"aws:SourceAccount": stack.Account(),
				},
				"ArnLike": map[string]interface{}{
					"aws:SourceArn": awscdk.Fn_Sub(jsii.String("arn:aws:s3files:${AWS::Region}:${AWS::AccountId}:file-system/*"), nil),
				},
			},
		}),
	})
	bucket.GrantReadWrite(role, nil)
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings("s3:ListBucketVersions"),
		Resources: jsii.Strings(*bucket.BucketArn()),
	}))
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings(
			"events:DeleteRule",
			"events:DisableRule",
			"events:EnableRule",
			"events:PutRule",
			"events:PutTargets",
			"events:RemoveTargets",
		),
		Resources: jsii.Strings("arn:aws:events:*:*:rule/DO-NOT-DELETE-S3-Files*"),
	}))
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: jsii.Strings(
			"events:DescribeRule",
			"events:ListRuleNamesByTarget",
			"events:ListRules",
			"events:ListTargetsByRule",
		),
		Resources: jsii.Strings("*"),
	}))

	fileSystem := awss3files.NewCfnFileSystem(stack, jsii.String("FileSystem"), &awss3files.CfnFileSystemProps{
		Bucket:              bucket.BucketArn(),
		RoleArn:             role.RoleArn(),
		AcceptBucketWarning: jsii.Bool(true),
	})
	// The file system validates the role on creation, so wait for its policy.
	fileSystem.Node().AddDependency(role)

	sg := awsec2.NewCfnSecurityGroup(stack, jsii.String("MountTargetSg"), &awsec2.CfnSecurityGroupProps{
		GroupDescription: jsii.String("S3 Files mount target"),
		VpcId:            jsii.String(vpcId),
		SecurityGroupIngress: &[]interface{}{
			&awsec2.CfnSecurityGroup_IngressProperty{
				IpProtocol: jsii.String("tcp"),
				FromPort:   jsii.Number(2049),
				ToPort:     jsii.Number(2049),
				CidrIp:     jsii.String(vpcCidr),
			},
		},
	})

	awss3files.NewCfnMountTarget(stack, jsii.String("MountTarget"), &awss3files.CfnMountTargetProps{
		FileSystemId:   fileSystem.AttrFileSystemId(),
		SubnetId:       jsii.String(subnetId),
		SecurityGroups: jsii.Strings(*sg.AttrGroupId()),
	})

	awscdk.NewCfnOutput(stack, jsii.String("FileSystemId"), &awscdk.CfnOutputProps{Value: fileSystem.AttrFileSystemId()})

	return stack
}
