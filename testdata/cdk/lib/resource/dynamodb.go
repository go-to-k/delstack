package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewDynamoDB(scope constructs.Construct, resourcePrefix string) {
	awsdynamodb.NewTable(scope, jsii.String("TableForBackup"), &awsdynamodb.TableProps{
		TableName: jsii.String(resourcePrefix + "-Table"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
}
