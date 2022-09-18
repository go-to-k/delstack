package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

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

func (factory *OperatorFactory) CreateStackOperator() Operator {
	return factory.stackOperatorFactory.CreateStackOperator()
}

func (factory *OperatorFactory) CreateBackupVaultOperator() Operator {
	return factory.backupVaultOperatorFactory.CreateBackupVaultOperator()
}

func (factory *OperatorFactory) CreateEcrOperator() Operator {
	return factory.ecrOperatorFactory.CreateEcrOperator()
}

func (factory *OperatorFactory) CreateRoleOperator() Operator {
	return factory.roleOperatorFactory.CreateRoleOperator()
}

func (factory *OperatorFactory) CreateBucketOperator() Operator {
	return factory.bucketOperatorFactory.CreateBucketOperator()
}

func (factory *OperatorFactory) CreateCustomOperator() Operator {
	return factory.customOperatorFactory.CreateCustomOperator()
}
