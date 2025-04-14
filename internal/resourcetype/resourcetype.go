package resourcetype

const (
	S3Bucket            = "AWS::S3::Bucket"
	S3DirectoryBucket   = "AWS::S3Express::DirectoryBucket"
	S3TableBucket       = "AWS::S3Tables::TableBucket"
	IamGroup            = "AWS::IAM::Group"
	EcrRepository       = "AWS::ECR::Repository"
	BackupVault         = "AWS::Backup::BackupVault"
	CloudformationStack = "AWS::CloudFormation::Stack"
	CustomResource      = "Custom::"
)

func GetResourceTypes() []string {
	return []string{
		S3Bucket,
		S3DirectoryBucket,
		S3TableBucket,
		IamGroup,
		EcrRepository,
		BackupVault,
		CloudformationStack,
		CustomResource,
	}
}
