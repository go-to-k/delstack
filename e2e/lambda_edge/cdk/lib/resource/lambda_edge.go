package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewLambdaEdge(scope constructs.Construct) {
	bucket := awss3.NewBucket(scope, jsii.String("OriginBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	edgeFn := awscloudfront.NewFunction(scope, jsii.String("EdgeFn"), &awscloudfront.FunctionProps{
		Code: awscloudfront.FunctionCode_FromInline(jsii.String(
			`function handler(event) { return event.response; }`,
		)),
	})

	awscloudfront.NewDistribution(scope, jsii.String("Distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{}),
			FunctionAssociations: &[]*awscloudfront.FunctionAssociation{
				{
					Function:  edgeFn,
					EventType: awscloudfront.FunctionEventType_VIEWER_RESPONSE,
				},
			},
		},
	})

	// Create a Lambda@Edge function (actual edge function using awslambda)
	lambdaEdgeFn := awslambda.NewFunction(scope, jsii.String("LambdaEdgeFn"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_NODEJS_20_X(),
		Handler: jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(
			`exports.handler = async (event) => { return event.Records[0].cf.response; };`,
		)),
	})

	// Create a separate distribution with Lambda@Edge
	awscloudfront.NewDistribution(scope, jsii.String("EdgeDistribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(bucket, &awscloudfrontorigins.S3BucketOriginWithOACProps{}),
			EdgeLambdas: &[]*awscloudfront.EdgeLambda{
				{
					FunctionVersion: lambdaEdgeFn.CurrentVersion(),
					EventType:       awscloudfront.LambdaEdgeEventType_ORIGIN_RESPONSE,
				},
			},
		},
	})
}
