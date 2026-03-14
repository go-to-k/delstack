package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewRdsInstance creates an RDS DBInstance with deletion protection enabled.
func NewRdsInstance(scope constructs.Construct, vpc awsec2.Vpc) awsrds.DatabaseInstance {
	dbInstance := awsrds.NewDatabaseInstance(scope, jsii.String("RdsInstance"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Mysql(&awsrds.MySqlInstanceEngineProps{
			Version: awsrds.MysqlEngineVersion_VER_8_0(),
		}),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_T3, awsec2.InstanceSize_MICRO),
		Vpc:          vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		// Enable deletion protection
		DeletionProtection: jsii.Bool(true),
		// Set RemovalPolicy to DESTROY so CDK allows creating the resource
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		// Disable multi-AZ to minimize costs
		MultiAz: jsii.Bool(false),
		// Disable automated backups to speed up deletion
		BackupRetention: awscdk.Duration_Days(jsii.Number(0)),
		// No storage encryption for simplicity
		StorageEncrypted: jsii.Bool(false),
		// Minimal storage
		AllocatedStorage: jsii.Number(20),
	})

	return dbInstance
}
