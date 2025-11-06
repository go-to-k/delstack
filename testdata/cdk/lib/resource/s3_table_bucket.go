package resource

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3tables"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewS3TableBucket(scope constructs.Construct, bucketNamePrefix string) {
	tableBucket := awss3tables.NewCfnTableBucket(scope, jsii.String("TableBucket"), &awss3tables.CfnTableBucketProps{
		TableBucketName: jsii.String(strings.ToLower(bucketNamePrefix)),
	})

	// Create namespaces and tables using CDK (CloudFormation)
	// The deploy.go also creates additional namespaces and tables via SDK
	namespaceAmount := 2
	tableAmount := 2

	for i := 0; i < namespaceAmount; i++ {
		namespaceName := fmt.Sprintf("cfn_namespace_%d", i)
		namespaceId := fmt.Sprintf("Namespace%d", i)

		namespace := awscdk.NewCfnResource(scope, jsii.String(namespaceId), &awscdk.CfnResourceProps{
			Type: jsii.String("AWS::S3Tables::Namespace"),
			Properties: &map[string]interface{}{
				"TableBucketARN": tableBucket.GetAtt(jsii.String("TableBucketARN"), awscdk.ResolutionTypeHint_STRING),
				"Namespace":      namespaceName,
			},
		})

		// Add dependency to ensure namespace is created after the table bucket
		namespace.AddDependency(tableBucket)

		// Create tables within this namespace
		for j := 0; j < tableAmount; j++ {
			tableName := fmt.Sprintf("cfn_table_%d", j)
			tableId := fmt.Sprintf("Table%d%d", i, j)

			table := awscdk.NewCfnResource(scope, jsii.String(tableId), &awscdk.CfnResourceProps{
				Type: jsii.String("AWS::S3Tables::Table"),
				Properties: &map[string]interface{}{
					"TableBucketARN":  tableBucket.GetAtt(jsii.String("TableBucketARN"), awscdk.ResolutionTypeHint_STRING),
					"Namespace":       namespaceName,
					"TableName":       tableName,
					"OpenTableFormat": "ICEBERG",
					"IcebergMetadata": map[string]interface{}{
						"IcebergSchema": map[string]interface{}{
							"SchemaFieldList": []map[string]interface{}{
								{
									"Name":     "column",
									"Type":     "int",
									"Required": false,
								},
							},
						},
					},
				},
			})

			// Add dependency to ensure table is created after the namespace
			table.AddDependency(namespace)
		}
	}
}
