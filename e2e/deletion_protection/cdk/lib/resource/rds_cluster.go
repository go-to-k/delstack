package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewRdsCluster creates an Aurora MySQL DBCluster with deletion protection enabled.
func NewRdsCluster(scope constructs.Construct, vpc awsec2.Vpc) awsrds.DatabaseCluster {
	dbCluster := awsrds.NewDatabaseCluster(scope, jsii.String("RdsCluster"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraMysql(&awsrds.AuroraMysqlClusterEngineProps{
			Version: awsrds.AuroraMysqlEngineVersion_VER_3_08_0(),
		}),
		Writer: awsrds.ClusterInstance_Provisioned(jsii.String("Writer"), &awsrds.ProvisionedClusterInstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MEDIUM),
		}),
		Vpc: vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		// Enable deletion protection
		DeletionProtection: jsii.Bool(true),
		// Set RemovalPolicy to DESTROY so CDK allows creating the resource
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		// Minimal backup retention
		Backup: &awsrds.BackupProps{
			Retention: awscdk.Duration_Days(jsii.Number(1)),
		},
		// No storage encryption for simplicity
		StorageEncrypted: jsii.Bool(false),
	})

	return dbCluster
}
