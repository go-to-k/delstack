//go:generate mockgen -source=$GOFILE -destination=cognito_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"

	cognitotypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type ICognito interface {
	CheckUserPoolDeletionProtection(ctx context.Context, userPoolId *string) (bool, error)
	DisableUserPoolDeletionProtection(ctx context.Context, userPoolId *string) error
}

var _ ICognito = (*Cognito)(nil)

type Cognito struct {
	client *cognitoidentityprovider.Client
}

func NewCognito(client *cognitoidentityprovider.Client) *Cognito {
	return &Cognito{
		client: client,
	}
}

func (c *Cognito) CheckUserPoolDeletionProtection(ctx context.Context, userPoolId *string) (bool, error) {
	input := &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: userPoolId,
	}

	output, err := c.client.DescribeUserPool(ctx, input)
	if err != nil {
		return false, &ClientError{
			ResourceName: userPoolId,
			Err:          err,
		}
	}

	if output.UserPool == nil {
		return false, nil
	}

	return output.UserPool.DeletionProtection == cognitotypes.DeletionProtectionTypeActive, nil
}

func (c *Cognito) DisableUserPoolDeletionProtection(ctx context.Context, userPoolId *string) error {
	input := &cognitoidentityprovider.UpdateUserPoolInput{
		UserPoolId:         userPoolId,
		DeletionProtection: cognitotypes.DeletionProtectionTypeInactive,
	}

	_, err := c.client.UpdateUserPool(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: userPoolId,
			Err:          err,
		}
	}

	return nil
}
