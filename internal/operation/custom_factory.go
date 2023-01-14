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

func (f *CustomOperatorFactory) CreateCustomOperator() *CustomOperator {
	return NewCustomOperator() // Implicit instances that do not actually delete resources
}
