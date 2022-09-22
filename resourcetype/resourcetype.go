package resourcetype

const (
	CLOUDFORMATION_STACK = "AWS::CloudFormation::Stack"
	S3_BUCKET            = "AWS::S3::Bucket"
	IAM_ROLE             = "AWS::IAM::Role"
	ECR_REPOSITORY       = "AWS::ECR::Repository"
	BACKUP_VAULT         = "AWS::Backup::BackupVault"
	CUSTOM_RESOURCE      = "Custom::"
)

func GetResourceTypes() []string {
	return []string{
		CLOUDFORMATION_STACK,
		S3_BUCKET,
		IAM_ROLE,
		ECR_REPOSITORY,
		BACKUP_VAULT,
		CUSTOM_RESOURCE,
	}
}
