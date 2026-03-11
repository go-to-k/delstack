package resourcetype

// For Force Deletion
const (
	S3Bucket            = "AWS::S3::Bucket"
	S3DirectoryBucket   = "AWS::S3Express::DirectoryBucket"
	S3TableBucket       = "AWS::S3Tables::TableBucket"
	S3TableNamespace    = "AWS::S3Tables::Namespace"
	S3VectorBucket      = "AWS::S3Vectors::VectorBucket"
	IamGroup            = "AWS::IAM::Group"
	EcrRepository       = "AWS::ECR::Repository"
	BackupVault         = "AWS::Backup::BackupVault"
	CloudformationStack = "AWS::CloudFormation::Stack"
	CustomResource      = "Custom::"
)

// For Preprocessors to filter resources by type
const (
	LambdaFunction = "AWS::Lambda::Function"
)

var ResourceTypes = []string{
	S3Bucket,
	S3DirectoryBucket,
	S3TableBucket,
	S3TableNamespace,
	S3VectorBucket,
	IamGroup,
	EcrRepository,
	BackupVault,
	CloudformationStack,
	CustomResource,
}
