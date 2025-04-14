package resource

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awss3tables"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewS3TableBucket(scope constructs.Construct, bucketNamePrefix string) {
	// Namespaces and tables are created in the deploy.go file because they are not supported in the CDK and CFn yet
	awss3tables.NewCfnTableBucket(scope, jsii.String("TableBucket"), &awss3tables.CfnTableBucketProps{
		TableBucketName: jsii.String(strings.ToLower(bucketNamePrefix)),
	})
}
