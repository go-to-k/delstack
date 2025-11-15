package resource

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awss3express"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewS3DirectoryBucket(scope constructs.Construct, bucketNamePrefix string) {
	awss3express.NewCfnDirectoryBucket(scope, jsii.String("DirectoryBucket"), &awss3express.CfnDirectoryBucketProps{
		BucketName:     jsii.String(strings.ToLower(bucketNamePrefix) + "--use1-az4--x-s3"),
		DataRedundancy: jsii.String("SingleAvailabilityZone"),
		LocationName:   jsii.String("use1-az4"),
	})
}
