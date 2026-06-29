package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var _ IOperator = (*CognitoUserPoolUICustomizationAttachmentOperator)(nil)

type CognitoUserPoolUICustomizationAttachmentOperator struct {
	resources []*types.StackResourceSummary
}

func NewCognitoUserPoolUICustomizationAttachmentOperator() *CognitoUserPoolUICustomizationAttachmentOperator {
	return &CognitoUserPoolUICustomizationAttachmentOperator{
		resources: []*types.StackResourceSummary{},
	}
}

func (o *CognitoUserPoolUICustomizationAttachmentOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *CognitoUserPoolUICustomizationAttachmentOperator) GetResourcesLength() int {
	return len(o.resources)
}

// A UserPoolUICustomizationAttachment has no standalone AWS resource to delete: it is only UI
// customization (CSS/logo) on a user pool, applied via SetUICustomization, which requires an
// existing UserPoolDomain. It reaches DELETE_FAILED only as a phantom (e.g. its create failed
// because no domain existed, so nothing was ever set), where there is nothing to force-delete.
// Hence this is a no-op and the CFN delete loop retains the logical resource to drop it from the
// stack. (Implicit implement; these resources will be deleted on its own.)
func (o *CognitoUserPoolUICustomizationAttachmentOperator) DeleteResources(ctx context.Context) error {
	return nil
}
