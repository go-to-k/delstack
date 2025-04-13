package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsbackup"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewBackup(scope constructs.Construct, resourcePrefix string) {
	backupRole := awsiam.NewRole(scope, jsii.String("AWSBackupServiceRole"), &awsiam.RoleProps{
		Path:      jsii.String("/service-role/"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("backup.amazonaws.com"), nil),
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSBackupServiceRolePolicyForBackup")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSBackupServiceRolePolicyForRestores")),
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonDynamoDBReadOnlyAccess")),
		},
	})

	backupVault := awsbackup.NewBackupVault(scope, jsii.String("BackupVaultWithThinBackups"), &awsbackup.BackupVaultProps{
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	backupPlan := awsbackup.NewCfnBackupPlan(scope, jsii.String("BackupPlanWithThinBackups"), &awsbackup.CfnBackupPlanProps{
		BackupPlan: &awsbackup.CfnBackupPlan_BackupPlanResourceTypeProperty{
			BackupPlanName: jsii.String(resourcePrefix + "-Backup-Plan"),
			BackupPlanRule: &[]*awsbackup.CfnBackupPlan_BackupRuleResourceTypeProperty{
				{
					RuleName:           jsii.String("RuleForDailyBackups"),
					TargetBackupVault:  backupVault.BackupVaultName(),
					ScheduleExpression: jsii.String("cron(0 18 * * ? *)"),
					StartWindowMinutes: jsii.Number(60),
					Lifecycle: &awsbackup.CfnBackupPlan_LifecycleResourceTypeProperty{
						DeleteAfterDays: jsii.Number(3),
					},
				},
			},
		},
	})

	selectionName := resourcePrefix + "-Backup-Selection"
	awsbackup.NewCfnBackupSelection(scope, jsii.String("TagBasedBackupSelection"), &awsbackup.CfnBackupSelectionProps{
		BackupPlanId: backupPlan.AttrBackupPlanId(),
		BackupSelection: &awsbackup.CfnBackupSelection_BackupSelectionResourceTypeProperty{
			SelectionName: jsii.String(selectionName),
			IamRoleArn:    backupRole.RoleArn(),
			ListOfTags: &[]*awsbackup.CfnBackupSelection_ConditionResourceTypeProperty{
				{
					ConditionType:  jsii.String("STRINGEQUALS"),
					ConditionKey:   jsii.String("Test"),
					ConditionValue: jsii.String("Test"),
				},
			},
		},
	})
}
