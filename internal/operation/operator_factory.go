package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

type IOperatorFactory interface {
	CreateStackOperator(targetResourceTypes []string) *StackOperator
	CreateBackupVaultOperator() *BackupVaultOperator
	CreateEcrOperator() *EcrOperator
	CreateRoleOperator() *RoleOperator
	CreateBucketOperator() *BucketOperator
	CreateCustomOperator() *CustomOperator
}

var _ IOperatorFactory = (*OperatorFactory)(nil)

type OperatorFactory struct {
	config                     aws.Config
	stackOperatorFactory       *StackOperatorFactory
	backupVaultOperatorFactory *BackupVaultOperatorFactory
	ecrOperatorFactory         *EcrOperatorFactory
	roleOperatorFactory        *RoleOperatorFactory
	bucketOperatorFactory      *BucketOperatorFactory
	customOperatorFactory      *CustomOperatorFactory
}

func NewOperatorFactory(config aws.Config) *OperatorFactory {
	return &OperatorFactory{
		config,
		NewStackOperatorFactory(config),
		NewBackupVaultOperatorFactory(config),
		NewEcrOperatorFactory(config),
		NewRoleOperatorFactory(config),
		NewBucketOperatorFactory(config),
		NewCustomOperatorFactory(config),
	}
}

func (f *OperatorFactory) CreateStackOperator(targetResourceTypes []string) *StackOperator {
	return f.stackOperatorFactory.CreateStackOperator(targetResourceTypes)
}

func (f *OperatorFactory) CreateBackupVaultOperator() *BackupVaultOperator {
	return f.backupVaultOperatorFactory.CreateBackupVaultOperator()
}

func (f *OperatorFactory) CreateEcrOperator() *EcrOperator {
	return f.ecrOperatorFactory.CreateEcrOperator()
}

func (f *OperatorFactory) CreateRoleOperator() *RoleOperator {
	return f.roleOperatorFactory.CreateRoleOperator()
}

func (f *OperatorFactory) CreateBucketOperator() *BucketOperator {
	return f.bucketOperatorFactory.CreateBucketOperator()
}

func (f *OperatorFactory) CreateCustomOperator() *CustomOperator {
	return f.customOperatorFactory.CreateCustomOperator()
}
