package resourcetype

const (
	S3Bucket            = "AWS::S3::Bucket"
	S3DirectoryBucket   = "AWS::S3Express::DirectoryBucket"
	IamRole             = "AWS::IAM::Role"
	EcrRepository       = "AWS::ECR::Repository"
	BackupVault         = "AWS::Backup::BackupVault"
	CloudformationStack = "AWS::CloudFormation::Stack"
	CustomResource      = "Custom::"
)

func GetResourceTypes() []string {
	return []string{
		S3Bucket,
		S3DirectoryBucket,
		IamRole,
		EcrRepository,
		BackupVault,
		CloudformationStack,
		CustomResource,
	}
}
