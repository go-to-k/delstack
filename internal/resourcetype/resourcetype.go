package resourcetype

// For Force Deletion
const (
	S3Bucket                     = "AWS::S3::Bucket"
	S3DirectoryBucket            = "AWS::S3Express::DirectoryBucket"
	S3TableBucket                = "AWS::S3Tables::TableBucket"
	S3TableNamespace             = "AWS::S3Tables::Namespace"
	S3VectorBucket               = "AWS::S3Vectors::VectorBucket"
	IamGroup                     = "AWS::IAM::Group"
	IamUser                      = "AWS::IAM::User"
	EcrRepository                = "AWS::ECR::Repository"
	BackupVault                  = "AWS::Backup::BackupVault"
	AthenaWorkGroup              = "AWS::Athena::WorkGroup"
	CloudformationStack          = "AWS::CloudFormation::Stack"
	CloudformationCustomResource = "AWS::CloudFormation::CustomResource"
	CustomResource               = "Custom::"
)

// For Force Deletion and Preprocessors
const (
	LambdaFunction = "AWS::Lambda::Function"
)

// For Deletion Protection Check
const (
	Ec2Instance       = "AWS::EC2::Instance"
	RdsDBInstance     = "AWS::RDS::DBInstance"
	RdsDBCluster      = "AWS::RDS::DBCluster"
	CognitoUserPool   = "AWS::Cognito::UserPool"
	LogsLogGroup      = "AWS::Logs::LogGroup"
	Elbv2LoadBalancer = "AWS::ElasticLoadBalancingV2::LoadBalancer"
)

var ResourceTypes = []string{
	S3Bucket,
	S3DirectoryBucket,
	S3TableBucket,
	S3TableNamespace,
	S3VectorBucket,
	IamGroup,
	IamUser,
	EcrRepository,
	BackupVault,
	AthenaWorkGroup,
	LambdaFunction,
	CloudformationStack,
	CloudformationCustomResource,
	CustomResource,
}
