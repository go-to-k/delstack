package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewCognitoUserPool creates a Cognito UserPool with deletion protection active.
func NewCognitoUserPool(scope constructs.Construct) awscognito.UserPool {
	userPool := awscognito.NewUserPool(scope, jsii.String("CognitoUserPool"), &awscognito.UserPoolProps{
		// Enable deletion protection
		DeletionProtection: jsii.Bool(true),
		RemovalPolicy:      awscdk.RemovalPolicy_DESTROY,
	})

	return userPool
}
