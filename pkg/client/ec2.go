//go:generate mockgen -source=$GOFILE -destination=ec2_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	ENIDetachmentWaitTime = 90 * time.Second // Maximum wait time for ENI detachment
)

type IEC2 interface {
	DescribeNetworkInterfaces(ctx context.Context, filters []types.Filter) ([]types.NetworkInterface, error)
	DeleteNetworkInterface(ctx context.Context, networkInterfaceId *string) error
	DeleteSubnet(ctx context.Context, subnetId *string) error
	DeleteSecurityGroup(ctx context.Context, securityGroupId *string) error
	CheckTerminationProtection(ctx context.Context, instanceId *string) (bool, error)
	DisableTerminationProtection(ctx context.Context, instanceId *string) error
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

func (c *EC2Client) DeleteSubnet(ctx context.Context, subnetId *string) error {
	input := &ec2.DeleteSubnetInput{
		SubnetId: subnetId,
	}

	_, err := c.client.DeleteSubnet(ctx, input)
	if err != nil {
		// If the subnet is already gone, treat as success.
		if strings.Contains(err.Error(), "InvalidSubnetID.NotFound") {
			return nil
		}
		return &ClientError{
			ResourceName: subnetId,
			Err:          err,
		}
	}

	return nil
}

func (c *EC2Client) DeleteSecurityGroup(ctx context.Context, securityGroupId *string) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: securityGroupId,
	}

	_, err := c.client.DeleteSecurityGroup(ctx, input)
	if err != nil {
		// If the security group is already gone, treat as success.
		if strings.Contains(err.Error(), "InvalidGroup.NotFound") {
			return nil
		}
		return &ClientError{
			ResourceName: securityGroupId,
			Err:          err,
		}
	}

	return nil
}

func (c *EC2Client) CheckTerminationProtection(ctx context.Context, instanceId *string) (bool, error) {
	input := &ec2.DescribeInstanceAttributeInput{
		InstanceId: instanceId,
		Attribute:  types.InstanceAttributeNameDisableApiTermination,
	}

	output, err := c.client.DescribeInstanceAttribute(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: instanceId,
			Err:          err,
		}
	}

	if output.DisableApiTermination != nil && output.DisableApiTermination.Value != nil {
		return aws.ToBool(output.DisableApiTermination.Value), nil
	}

	return false, nil
}

func (c *EC2Client) DisableTerminationProtection(ctx context.Context, instanceId *string) error {
	input := &ec2.ModifyInstanceAttributeInput{
		InstanceId: instanceId,
		DisableApiTermination: &types.AttributeBooleanValue{
			Value: aws.Bool(false),
		},
	}

	_, err := c.client.ModifyInstanceAttribute(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: instanceId,
			Err:          err,
		}
	}

	return nil
}
