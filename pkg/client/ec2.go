//go:generate mockgen -source=$GOFILE -destination=ec2_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	ENIDetachmentWaitTime = 90 * time.Second // Maximum wait time for ENI detachment
)

type IEC2 interface {
	DescribeNetworkInterfaces(ctx context.Context, filters []types.Filter) ([]types.NetworkInterface, error)
	DeleteNetworkInterface(ctx context.Context, networkInterfaceId *string) error
}

var _ IEC2 = (*EC2Client)(nil)

type EC2Client struct {
	client *ec2.Client
}

func NewEC2Client(client *ec2.Client) *EC2Client {
	return &EC2Client{
		client: client,
	}
}

func (c *EC2Client) DescribeNetworkInterfaces(ctx context.Context, filters []types.Filter) ([]types.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: filters,
	}

	output, err := c.client.DescribeNetworkInterfaces(ctx, input)
	if err != nil {
		return nil, &ClientError{
			Err: err,
		}
	}

	return output.NetworkInterfaces, nil
}

func (c *EC2Client) DeleteNetworkInterface(ctx context.Context, networkInterfaceId *string) error {
	input := &ec2.DeleteNetworkInterfaceInput{
		NetworkInterfaceId: networkInterfaceId,
	}

	// Wait for ENIs to transition to "available" state after VPC detach
	// Lambda waiter confirms function update, but ENI detachment is asynchronous
	time.Sleep(10 * time.Second)

	startTime := time.Now()
	for {
		_, err := c.client.DeleteNetworkInterface(ctx, input)
		if err == nil {
			return nil
		}

		// If ENI is already deleted, consider it success
		if strings.Contains(err.Error(), "InvalidNetworkInterfaceID.NotFound") {
			return nil
		}

		// If ENI is in use, wait and retry
		if time.Since(startTime) < ENIDetachmentWaitTime {
			time.Sleep(5 * time.Second)
			continue
		}

		return &ClientError{
			ResourceName: networkInterfaceId,
			Err:          fmt.Errorf("timeout waiting for ENI to be deletable: %w", err),
		}
	}
}
