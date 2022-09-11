package option

import (
	"context"
	"log"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/jessevdk/go-flags"
)

var CONCURRENCY_NUM = runtime.NumCPU()

type Option struct {
	Profile   string `short:"p" long:"profile" description:"AWS profile name"`
	StackName string `short:"s" long:"stackName" description:"Stack name" required:"true"`
	Region    string `short:"r" long:"region" description:"AWS Region" default:"ap-northeast-1"`
}

// Arguments are passed by go-flags module
func NewOption() *Option {
	return &Option{}
}

func (option *Option) Parse() ([]string, error) {
	result, err := flags.Parse(option)
	return result, err
}

func (option *Option) LoadAwsConfig() (aws.Config, error) {
	var (
		cfg aws.Config
		err error
	)

	if option.Profile != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(option.Region), config.WithSharedConfigProfile(option.Profile))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(option.Region))
	}

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
		return cfg, err
	}

	return cfg, nil
}
