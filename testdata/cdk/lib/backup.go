package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsbackup"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewBackupResources creates required AWS Backup resources
func NewBackupResources(scope constructs.Construct, pjPrefix string, iamResources map[string]awscdk.IResource) {
	resources := make(map[string]interface{})

	// Get the AWS Backup role
	backupRole, ok := iamResources["AWSBackupServiceRole"].(awsiam.Role)
	if !ok {
		panic("AWSBackupServiceRole not found or not a Role")
	}

	// Create backup vault 1
	backupVault1 := awsbackup.NewBackupVault(scope, jsii.String("BackupVaultWithThinBackups1"), &awsbackup.BackupVaultProps{
		BackupVaultName: jsii.String(pjPrefix + "-Backup-Vault1"),
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
	})
	resources["BackupVaultWithThinBackups1"] = backupVault1

	// Create backup vault 2
	backupVault2 := awsbackup.NewBackupVault(scope, jsii.String("BackupVaultWithThinBackups2"), &awsbackup.BackupVaultProps{
		BackupVaultName: jsii.String(pjPrefix + "-Backup-Vault2"),
		RemovalPolicy:   awscdk.RemovalPolicy_DESTROY,
	})
	resources["BackupVaultWithThinBackups2"] = backupVault2

	// Create backup plan 1 using L1 constructs directly
	backupPlan1 := awsbackup.NewCfnBackupPlan(scope, jsii.String("BackupPlanWithThinBackups1"), &awsbackup.CfnBackupPlanProps{
		BackupPlan: &awsbackup.CfnBackupPlan_BackupPlanResourceTypeProperty{
			BackupPlanName: jsii.String(pjPrefix + "-Backup-Plan1"),
			BackupPlanRule: &[]*awsbackup.CfnBackupPlan_BackupRuleResourceTypeProperty{
				{
					RuleName:           jsii.String("RuleForDailyBackups1"),
					TargetBackupVault:  backupVault1.BackupVaultName(),
					ScheduleExpression: jsii.String("cron(0 18 * * ? *)"),
					StartWindowMinutes: jsii.Number(60),
					Lifecycle: &awsbackup.CfnBackupPlan_LifecycleResourceTypeProperty{
						DeleteAfterDays: jsii.Number(3),
					},
				},
			},
		},
	})

	// Create backup plan 2 using L1 constructs directly
	backupPlan2 := awsbackup.NewCfnBackupPlan(scope, jsii.String("BackupPlanWithThinBackups2"), &awsbackup.CfnBackupPlanProps{
		BackupPlan: &awsbackup.CfnBackupPlan_BackupPlanResourceTypeProperty{
			BackupPlanName: jsii.String(pjPrefix + "-Backup-Plan2"),
			BackupPlanRule: &[]*awsbackup.CfnBackupPlan_BackupRuleResourceTypeProperty{
				{
					RuleName:           jsii.String("RuleForDailyBackups2"),
					TargetBackupVault:  backupVault2.BackupVaultName(),
					ScheduleExpression: jsii.String("cron(0 18 * * ? *)"),
					StartWindowMinutes: jsii.Number(60),
					Lifecycle: &awsbackup.CfnBackupPlan_LifecycleResourceTypeProperty{
						DeleteAfterDays: jsii.Number(3),
					},
				},
			},
		},
	})

	// Create backup selection 1 using L1 constructs
	selectionName1 := pjPrefix + "-Backup-Selection1"
	awsbackup.NewCfnBackupSelection(scope, jsii.String("TagBasedBackupSelection1"), &awsbackup.CfnBackupSelectionProps{
		BackupPlanId: backupPlan1.AttrBackupPlanId(),
		BackupSelection: &awsbackup.CfnBackupSelection_BackupSelectionResourceTypeProperty{
			SelectionName: jsii.String(selectionName1),
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

	// Create backup selection 2 using L1 constructs
	selectionName2 := pjPrefix + "-Backup-Selection2"
	awsbackup.NewCfnBackupSelection(scope, jsii.String("TagBasedBackupSelection2"), &awsbackup.CfnBackupSelectionProps{
		BackupPlanId: backupPlan2.AttrBackupPlanId(),
		BackupSelection: &awsbackup.CfnBackupSelection_BackupSelectionResourceTypeProperty{
			SelectionName: jsii.String(selectionName2),
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
