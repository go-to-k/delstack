//go:generate mockgen -source=$GOFILE -destination=elbv2_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type IELBV2 interface {
	CheckLoadBalancerDeletionProtection(ctx context.Context, loadBalancerArn *string) (bool, error)
	DisableLoadBalancerDeletionProtection(ctx context.Context, loadBalancerArn *string) error
}

var _ IELBV2 = (*ELBV2)(nil)

type ELBV2 struct {
	client *elasticloadbalancingv2.Client
}

func NewELBV2(client *elasticloadbalancingv2.Client) *ELBV2 {
	return &ELBV2{
		client: client,
	}
}

func (e *ELBV2) CheckLoadBalancerDeletionProtection(ctx context.Context, loadBalancerArn *string) (bool, error) {
	input := &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: loadBalancerArn,
	}

	output, err := e.client.DescribeLoadBalancerAttributes(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: loadBalancerArn,
			Err:          err,
		}
	}

	for _, attr := range output.Attributes {
		if aws.ToString(attr.Key) == "deletion_protection.enabled" {
			return aws.ToString(attr.Value) == "true", nil
		}
	}

	return false, nil
}

func (e *ELBV2) DisableLoadBalancerDeletionProtection(ctx context.Context, loadBalancerArn *string) error {
	input := &elasticloadbalancingv2.ModifyLoadBalancerAttributesInput{
		LoadBalancerArn: loadBalancerArn,
		Attributes: []types.LoadBalancerAttribute{
			{
				Key:   aws.String("deletion_protection.enabled"),
				Value: aws.String("false"),
			},
		},
	}

	_, err := e.client.ModifyLoadBalancerAttributes(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: loadBalancerArn,
			Err:          err,
		}
	}

	return nil
}
