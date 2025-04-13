package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewS3Bucket(scope constructs.Construct) {
	awss3.NewBucket(scope, jsii.String("Bucket"), &awss3.BucketProps{
		Versioned:     jsii.Bool(true),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
}
