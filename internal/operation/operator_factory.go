package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-to-k/delstack/pkg/client"
)

const SDKRetryMaxAttempts = 3

type OperatorFactory struct {
	config aws.Config
}

func NewOperatorFactory(config aws.Config) *OperatorFactory {
	return &OperatorFactory{
		config,
	}
}

func (f *OperatorFactory) CreateCloudFormationStackOperator(targetResourceTypes []string) *CloudFormationStackOperator {
	sdkCfnClient := cloudformation.NewFromConfig(f.config, func(o *cloudformation.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	sdkCfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(sdkCfnClient)

	return NewCloudFormationStackOperator(
		f.config,
		client.NewCloudFormation(
			sdkCfnClient,
			sdkCfnWaiter,
		),
		targetResourceTypes,
	)
}

func (f *OperatorFactory) CreateBackupVaultOperator() *BackupVaultOperator {
	sdkBackupClient := backup.NewFromConfig(f.config, func(o *backup.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewBackupVaultOperator(
		client.NewBackup(
			sdkBackupClient,
		),
	)
}

func (f *OperatorFactory) CreateEcrRepositoryOperator() *EcrRepositoryOperator {
	sdkEcrClient := ecr.NewFromConfig(f.config, func(o *ecr.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewEcrRepositoryOperator(
		client.NewEcr(
			sdkEcrClient,
		),
	)
}

func (f *OperatorFactory) CreateIamRoleOperator() *IamRoleOperator {
	sdkIamClient := iam.NewFromConfig(f.config, func(o *iam.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewIamRoleOperator(
		client.NewIam(
			sdkIamClient,
		),
	)
}

func (f *OperatorFactory) CreateIamGroupOperator() *IamGroupOperator {
	sdkIamClient := iam.NewFromConfig(f.config, func(o *iam.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewIamGroupOperator(
		client.NewIam(
			sdkIamClient,
		),
	)
}

func (f *OperatorFactory) CreateS3BucketOperator() *S3BucketOperator {
	sdkS3Client := s3.NewFromConfig(f.config, func(o *s3.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewS3BucketOperator(
		client.NewS3(
			sdkS3Client,
			false,
		),
	)
}

func (f *OperatorFactory) CreateS3DirectoryBucketOperator() *S3BucketOperator {
	sdkS3Client := s3.NewFromConfig(f.config, func(o *s3.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	// Basically, a separate operator should be defined for each resource type,
	// but the S3DirectoryBucket uses the same operator as the S3BucketOperator
	// since the process is almost the same.
	operator := NewS3BucketOperator(
		client.NewS3(
			sdkS3Client,
			true,
		),
	)

	return operator
}

func (f *OperatorFactory) CreateCustomOperator() *CustomOperator {
	return NewCustomOperator() // Implicit instances that do not actually delete resources
}
