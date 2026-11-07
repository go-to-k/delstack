package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix, _ := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackS3FilesTest"
	}

	// The client stack (VPC + EC2 that mounts the file system) is deployed first and
	// lives outside the target stack, so the writes keep running while delstack
	// deletes the target stack. Its VPC/Subnet IDs are passed to the target stack
	// via context instead of a CFn Export, because an Export imported by a
	// non-target stack makes delstack refuse the deletion.
	resource.NewClientStack(app, pjPrefix+"-client", &awscdk.StackProps{
		Env:       env(),
		StackName: jsii.String(pjPrefix + "-client"),
	})

	vpcId, _ := app.Node().TryGetContext(jsii.String("VPC_ID")).(string)
	subnetId, _ := app.Node().TryGetContext(jsii.String("SUBNET_ID")).(string)
	vpcCidr, _ := app.Node().TryGetContext(jsii.String("VPC_CIDR")).(string)
	if vpcId != "" && subnetId != "" && vpcCidr != "" {
		resource.NewTargetStack(app, pjPrefix, &awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix),
		}, vpcId, subnetId, vpcCidr)
	}

	app.Synth(nil)
}

func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	if region == "" {
		region = "us-east-1"
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
