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

type IOperatorFactory interface {
	CreateCloudFormationStackOperator(targetResourceTypes []string) *CloudFormationStackOperator
	CreateBackupVaultOperator() *BackupVaultOperator
	CreateEcrRepositoryOperator() *EcrRepositoryOperator
	CreateRoleOperator() *RoleOperator
	CreateBucketOperator() *BucketOperator
	CreateCustomOperator() *CustomOperator
}

var _ IOperatorFactory = (*OperatorFactory)(nil)

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

func (f *OperatorFactory) CreateRoleOperator() *RoleOperator {
	sdkIamClient := iam.NewFromConfig(f.config, func(o *iam.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewRoleOperator(
		client.NewIam(
			sdkIamClient,
		),
	)
}

func (f *OperatorFactory) CreateBucketOperator() *BucketOperator {
	sdkS3Client := s3.NewFromConfig(f.config, func(o *s3.Options) {
		o.RetryMaxAttempts = SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewBucketOperator(
		client.NewS3(
			sdkS3Client,
		),
	)
}

func (f *OperatorFactory) CreateCustomOperator() *CustomOperator {
	return NewCustomOperator() // Implicit instances that do not actually delete resources
}
