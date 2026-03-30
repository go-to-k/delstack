package app

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/pkg/client"
)

// IConfigLoader loads AWS SDK configuration for a given region and profile.
type IConfigLoader interface {
	LoadConfig(ctx context.Context, region, profile string) (aws.Config, error)
}

type ConfigLoader struct{}

func (l *ConfigLoader) LoadConfig(ctx context.Context, region, profile string) (aws.Config, error) {
	return client.LoadAWSConfig(ctx, region, profile)
}
