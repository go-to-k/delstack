package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewCreateFailedResources builds a stack whose CREATE deliberately fails and leaves a
// "phantom" DELETE_FAILED resource: one that never actually existed in AWS but that
// CloudFormation cannot cleanly delete on rollback, leaving the stack in ROLLBACK_FAILED.
//
// The vehicle is AWS::Cognito::UserPoolUICustomizationAttachment. Its underlying
// SetUICustomization API requires the user pool to have a UserPoolDomain. We create the
// attachment WITHOUT a domain on purpose, so its CREATE fails and, on rollback, its DELETE
// also fails (there is no domain / nothing to clear), leaving it in DELETE_FAILED.
//
// This scenario is intentionally generic: it stands in for the whole class of
// "create-failed phantom" resources (Custom Resources and Attachment/Association/Settings
// style types whose delete handlers are not idempotent). The Cognito attachment is just a
// cheap, reliable way to manufacture that condition.
func NewCreateFailedResources(scope constructs.Construct) {
	userPool := awscognito.NewCfnUserPool(scope, jsii.String("UserPool"), &awscognito.CfnUserPoolProps{})

	userPoolClient := awscognito.NewCfnUserPoolClient(scope, jsii.String("UserPoolClient"), &awscognito.CfnUserPoolClientProps{
		UserPoolId: userPool.Ref(),
	})

	// No UserPoolDomain is created on purpose: SetUICustomization fails without one, so this
	// resource's CREATE fails and it ends up as a phantom DELETE_FAILED on rollback.
	awscognito.NewCfnUserPoolUICustomizationAttachment(scope, jsii.String("UICustomization"), &awscognito.CfnUserPoolUICustomizationAttachmentProps{
		UserPoolId: userPool.Ref(),
		ClientId:   userPoolClient.Ref(),
		Css:        jsii.String(".label-customizable {color: #000000;}"),
	})
}
