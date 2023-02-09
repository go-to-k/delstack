package resourcetype

const (
	S3Bucket            = "AWS::S3::Bucket"
	IamRole             = "AWS::IAM::Role"
	EcrRepository       = "AWS::ECR::Repository"
	BackupVault         = "AWS::Backup::BackupVault"
	CloudformationStack = "AWS::CloudFormation::Stack"
	CustomResource      = "Custom::"
)

func GetResourceTypes() []string {
	return []string{
		S3Bucket,
		IamRole,
		EcrRepository,
		BackupVault,
		CloudformationStack,
		CustomResource,
	}
}
