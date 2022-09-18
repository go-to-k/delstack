package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

type CustomOperatorFactory struct {
	config aws.Config
}

func NewCustomOperatorFactory(config aws.Config) *CustomOperatorFactory {
	return &CustomOperatorFactory{config}
}

func (factory *CustomOperatorFactory) CreateCustomOperator() Operator {
	return NewCustomOperator() // Implicit instances that do not actually delete resources
}
