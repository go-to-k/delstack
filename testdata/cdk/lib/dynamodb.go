package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewDynamoDBResources creates DynamoDB resources
func NewDynamoDBResources(scope constructs.Construct, pjPrefix string) {
	// Create DynamoDB table for backup
	awsdynamodb.NewTable(scope, jsii.String("TableForBackup"), &awsdynamodb.TableProps{
		TableName: jsii.String(pjPrefix + "-Table"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PROVISIONED,
		ReadCapacity:  jsii.Number(5),
		WriteCapacity: jsii.Number(5),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
}
