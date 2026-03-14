//go:generate mockgen -source=$GOFILE -destination=cloudwatchlogs_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

type ICloudWatchLogs interface {
	CheckLogGroupDeletionProtection(ctx context.Context, logGroupName *string) (bool, error)
	DisableLogGroupDeletionProtection(ctx context.Context, logGroupName *string) error
}

var _ ICloudWatchLogs = (*CloudWatchLogs)(nil)

type CloudWatchLogs struct {
	client *cloudwatchlogs.Client
}

func NewCloudWatchLogs(client *cloudwatchlogs.Client) *CloudWatchLogs {
	return &CloudWatchLogs{
		client: client,
	}
}

func (c *CloudWatchLogs) CheckLogGroupDeletionProtection(ctx context.Context, logGroupName *string) (bool, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: logGroupName,
	}

	output, err := c.client.DescribeLogGroups(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: logGroupName,
			Err:          err,
		}
	}

	// Find exact match since we used prefix filter
	for _, lg := range output.LogGroups {
		if aws.ToString(lg.LogGroupName) == aws.ToString(logGroupName) {
			return aws.ToBool(lg.DeletionProtectionEnabled), nil
		}
	}

	return false, nil
}

func (c *CloudWatchLogs) DisableLogGroupDeletionProtection(ctx context.Context, logGroupName *string) error {
	input := &cloudwatchlogs.PutLogGroupDeletionProtectionInput{
		LogGroupIdentifier:        logGroupName,
		DeletionProtectionEnabled: aws.Bool(false),
	}

	_, err := c.client.PutLogGroupDeletionProtection(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: logGroupName,
			Err:          err,
		}
	}

	return nil
}
