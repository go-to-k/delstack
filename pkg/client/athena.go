//go:generate mockgen -source=$GOFILE -destination=athena_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
)

type IAthena interface {
	DeleteWorkGroup(ctx context.Context, workGroupName *string) error
	CheckAthenaWorkGroupExists(ctx context.Context, workGroupName *string) (bool, error)
}

var _ IAthena = (*Athena)(nil)

type Athena struct {
	client *athena.Client
}

func NewAthena(client *athena.Client) *Athena {
	return &Athena{
		client,
	}
}

func (a *Athena) DeleteWorkGroup(ctx context.Context, workGroupName *string) error {
	input := &athena.DeleteWorkGroupInput{
		WorkGroup:             workGroupName,
		RecursiveDeleteOption: aws.Bool(true),
	}

	_, err := a.client.DeleteWorkGroup(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: workGroupName,
			Err:          err,
		}
	}
	return nil
}

func (a *Athena) CheckAthenaWorkGroupExists(ctx context.Context, workGroupName *string) (bool, error) {
	input := &athena.GetWorkGroupInput{
		WorkGroup: workGroupName,
	}

	_, err := a.client.GetWorkGroup(ctx, input)
	if err != nil {
		if strings.Contains(err.Error(), "is not found") {
			return false, nil
		}
		return false, &ClientError{
			ResourceName: workGroupName,
			Err:          err,
		}
	}
	return true, nil
}
