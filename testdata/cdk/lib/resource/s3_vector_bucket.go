package resource

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewS3VectorBucket(scope constructs.Construct, bucketNamePrefix string) {
	// Note: AWS::S3Vectors::VectorBucket is not yet available in CDK, so we use CfnResource
	vectorBucket := awscdk.NewCfnResource(scope, jsii.String("VectorBucket"), &awscdk.CfnResourceProps{
		Type: jsii.String("AWS::S3Vectors::VectorBucket"),
		Properties: &map[string]interface{}{
			"VectorBucketName": strings.ToLower(bucketNamePrefix),
		},
	})

	// Create indexes using CloudFormation
	// Additional indexes and vectors are created in the deploy.go file using SDK
	for i := 0; i < 5; i++ {
		indexId := fmt.Sprintf("VectorIndex%d", i)
		indexName := fmt.Sprintf("cfn-index-%d", i)
		index := awscdk.NewCfnResource(scope, jsii.String(indexId), &awscdk.CfnResourceProps{
			Type: jsii.String("AWS::S3Vectors::Index"),
			Properties: &map[string]interface{}{
				"VectorBucketName": strings.ToLower(bucketNamePrefix),
				"IndexName":        indexName,
				"DataType":         "float32",
				"Dimension":        128,
				"DistanceMetric":   "cosine",
			},
		})
		// Add dependency to ensure indexes are created after the bucket
		index.AddDependency(vectorBucket)
	}
}
