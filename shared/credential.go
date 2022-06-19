package shared

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadAwsConfig(profile string) (aws.Config, error) {
	var (
		cfg aws.Config
		err error
	)

	if profile != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO())
	}

	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
		return cfg, err
	}

	return cfg, nil
}
