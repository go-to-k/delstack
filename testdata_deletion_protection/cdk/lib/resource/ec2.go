package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewEc2Instance creates an EC2 Instance with API termination protection enabled.
func NewEc2Instance(scope constructs.Construct, pjPrefix string, vpc awsec2.Vpc) awsec2.Instance {
	instance := awsec2.NewInstance(scope, jsii.String("Ec2Instance"), &awsec2.InstanceProps{
		InstanceName: jsii.String(pjPrefix + "-Ec2Instance"),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		MachineImage: awsec2.MachineImage_LatestAmazonLinux2023(&awsec2.AmazonLinux2023ImageSsmParameterProps{}),
		Vpc:          vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		// Enable API termination protection
		DisableApiTermination: jsii.Bool(true),
	})

	instance.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)

	return instance
}
