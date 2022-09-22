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

func (factory *OperatorFactory) CreateStackOperator(targetResourceTypes []string) *StackOperator {
	return factory.stackOperatorFactory.CreateStackOperator(targetResourceTypes)
}

func (factory *OperatorFactory) CreateBackupVaultOperator() *BackupVaultOperator {
	return factory.backupVaultOperatorFactory.CreateBackupVaultOperator()
}

func (factory *OperatorFactory) CreateEcrOperator() *EcrOperator {
	return factory.ecrOperatorFactory.CreateEcrOperator()
}

func (factory *OperatorFactory) CreateRoleOperator() *RoleOperator {
	return factory.roleOperatorFactory.CreateRoleOperator()
}

func (factory *OperatorFactory) CreateBucketOperator() *BucketOperator {
	return factory.bucketOperatorFactory.CreateBucketOperator()
}

func (factory *OperatorFactory) CreateCustomOperator() *CustomOperator {
	return factory.customOperatorFactory.CreateCustomOperator()
}
