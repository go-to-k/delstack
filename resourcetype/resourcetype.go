package resourcetype

const (
	S3_BUCKET            = "AWS::S3::Bucket"
	IAM_ROLE             = "AWS::IAM::Role"
	ECR_REPOSITORY       = "AWS::ECR::Repository"
	BACKUP_VAULT         = "AWS::Backup::BackupVault"
	CLOUDFORMATION_STACK = "AWS::CloudFormation::Stack"
	CUSTOM_RESOURCE      = "Custom::"
)

func GetResourceTypes() []string {
	return []string{
		S3_BUCKET,
		IAM_ROLE,
		ECR_REPOSITORY,
		BACKUP_VAULT,
		CLOUDFORMATION_STACK,
		CUSTOM_RESOURCE,
	}
}
