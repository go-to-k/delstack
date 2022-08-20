package operations

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadAwsConfig(profile string, region string) (aws.Config, error) {
	var (
		cfg aws.Config
		err error
	)

	if profile != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region), config.WithSharedConfigProfile(profile))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	}

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
		return cfg, err
	}

	return cfg, nil
}
