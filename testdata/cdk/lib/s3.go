package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3express"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewS3Resources creates required S3 resources
func NewS3Resources(scope constructs.Construct, pjPrefix string) {
	// Create regular S3 bucket
	awss3.NewBucket(scope, jsii.String("RootS3Bucket"), &awss3.BucketProps{
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	// Create S3 Express Directory Bucket using the correct package
	awss3express.NewCfnDirectoryBucket(scope, jsii.String("RootS3DirectoryBucket"), &awss3express.CfnDirectoryBucketProps{
		BucketName:     jsii.String(pjPrefix + "-root--use1-az4--x-s3"),
		DataRedundancy: jsii.String("SingleAvailabilityZone"),
		LocationName:   jsii.String("use1-az4"),
	})
}
