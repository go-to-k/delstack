package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewECRRepositories creates required ECR repositories
func NewECRRepositories(scope constructs.Construct, pjPrefix string) map[string]awsecr.Repository {
	repositories := make(map[string]awsecr.Repository)

	// Create ECR1
	ecr1 := awsecr.NewRepository(scope, jsii.String("ECR1"), &awsecr.RepositoryProps{
		RepositoryName:   jsii.String(pjPrefix + "-ecr1"),
		EmptyOnDelete:    jsii.Bool(false),
		RemovalPolicy:    awscdk.RemovalPolicy_DESTROY,
		AutoDeleteImages: jsii.Bool(false),
	})

	// Add lifecycle policy
	ecr1.AddLifecycleRule(&awsecr.LifecycleRule{
		Description:   jsii.String("Delete more than 3 images"),
		MaxImageCount: jsii.Number(3),
		TagStatus:     awsecr.TagStatus_ANY,
	})

	// Add tags
	awscdk.Tags_Of(ecr1).Add(
		jsii.String("Name"),
		jsii.String(pjPrefix+"-ECR1"),
		nil,
	)

	repositories["ECR1"] = ecr1

	// Create ECR2
	ecr2 := awsecr.NewRepository(scope, jsii.String("ECR2"), &awsecr.RepositoryProps{
		RepositoryName:   jsii.String(pjPrefix + "-ecr2"),
		EmptyOnDelete:    jsii.Bool(false),
		RemovalPolicy:    awscdk.RemovalPolicy_DESTROY,
		AutoDeleteImages: jsii.Bool(false),
	})

	// Add lifecycle policy
	ecr2.AddLifecycleRule(&awsecr.LifecycleRule{
		Description:   jsii.String("Delete more than 3 images"),
		MaxImageCount: jsii.Number(3),
		TagStatus:     awsecr.TagStatus_ANY,
	})

	// Add tags
	awscdk.Tags_Of(ecr2).Add(
		jsii.String("Name"),
		jsii.String(pjPrefix+"-ECR2"),
		nil,
	)

	repositories["ECR2"] = ecr2

	return repositories
}
