package resource

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsathena"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewAthena(scope constructs.Construct, resourcePrefix string) {
	bucketName := strings.ToLower(resourcePrefix) + "-athena-results"

	awss3.NewCfnBucket(scope, jsii.String("AthenaResultsBucket"), &awss3.CfnBucketProps{
		BucketName: jsii.String(bucketName),
	})

	awsathena.NewCfnWorkGroup(scope, jsii.String("AthenaWorkGroup"), &awsathena.CfnWorkGroupProps{
		Name:  jsii.String(resourcePrefix + "-AthenaWorkGroup"),
		State: jsii.String("ENABLED"),
		WorkGroupConfiguration: &awsathena.CfnWorkGroup_WorkGroupConfigurationProperty{
			ResultConfiguration: &awsathena.CfnWorkGroup_ResultConfigurationProperty{
				OutputLocation: jsii.String(fmt.Sprintf("s3://%s/", bucketName)),
			},
		},
	})
}
