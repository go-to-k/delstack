package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewEcr(scope constructs.Construct) {
	ecr := awsecr.NewRepository(scope, jsii.String("Ecr"), &awsecr.RepositoryProps{
		EmptyOnDelete: jsii.Bool(false),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	ecr.AddLifecycleRule(&awsecr.LifecycleRule{
		MaxImageCount: jsii.Number(3),
	})
}
