//go:generate mockgen -source=$GOFILE -destination=route53resolver_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
)

type IRoute53Resolver interface {
	GetResolverConfig(ctx context.Context, resourceId *string) (*types.ResolverConfig, bool, error)
	UpdateResolverConfig(ctx context.Context, resourceId *string, flag types.AutodefinedReverseFlag) error
}

var _ IRoute53Resolver = (*Route53Resolver)(nil)

type Route53Resolver struct {
	client *route53resolver.Client
}

func NewRoute53Resolver(client *route53resolver.Client) *Route53Resolver {
	return &Route53Resolver{
		client,
	}
}

// GetResolverConfig returns exists=false when the target resource (e.g. VPC) no longer exists.
func (r *Route53Resolver) GetResolverConfig(ctx context.Context, resourceId *string) (*types.ResolverConfig, bool, error) {
	input := &route53resolver.GetResolverConfigInput{
		ResourceId: resourceId,
	}

	output, err := r.client.GetResolverConfig(ctx, input)
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil, false, nil
		}
		return nil, false, &ClientError{
			ResourceName: resourceId,
			Err:          err,
		}
	}
	return output.ResolverConfig, true, nil
}

// UpdateResolverConfig succeeds without doing anything when the target resource (e.g. VPC) no longer exists.
func (r *Route53Resolver) UpdateResolverConfig(ctx context.Context, resourceId *string, flag types.AutodefinedReverseFlag) error {
	input := &route53resolver.UpdateResolverConfigInput{
		ResourceId:             resourceId,
		AutodefinedReverseFlag: flag,
	}

	_, err := r.client.UpdateResolverConfig(ctx, input)
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: resourceId,
			Err:          err,
		}
	}
	return nil
}
