package preprocessor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
)

var _ IPreprocessor = (*DeletionProtectionRemover)(nil)

type protectedResource struct {
	resourceType       string
	logicalResourceId  string
	physicalResourceId string
}

type DeletionProtectionRemover struct {
	forceMode     bool
	ec2Client     client.IEC2
	rdsClient     client.IRDS
	cognitoClient client.ICognito
	logsClient    client.ICloudWatchLogs
	elbv2Client   client.IELBV2
}

func NewDeletionProtectionRemover(
	forceMode bool,
	ec2Client client.IEC2,
	rdsClient client.IRDS,
	cognitoClient client.ICognito,
	logsClient client.ICloudWatchLogs,
	elbv2Client client.IELBV2,
) *DeletionProtectionRemover {
	return &DeletionProtectionRemover{
		forceMode:     forceMode,
		ec2Client:     ec2Client,
		rdsClient:     rdsClient,
		cognitoClient: cognitoClient,
		logsClient:    logsClient,
		elbv2Client:   elbv2Client,
	}
}

func (r *DeletionProtectionRemover) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	protectedResources := r.findProtectedResources(ctx, stackName, resources)

	if len(protectedResources) == 0 {
		return nil
	}

	if !r.forceMode {
		return r.buildProtectionError(protectedResources)
	}

	return r.disableProtections(ctx, stackName, protectedResources)
}

func (r *DeletionProtectionRemover) findProtectedResources(ctx context.Context, stackName *string, resources []types.StackResourceSummary) []protectedResource {
	type checkResult struct {
		resource  types.StackResourceSummary
		protected bool
	}

	var mu sync.Mutex
	var results []checkResult
	var wg sync.WaitGroup

	checkResource := func(res types.StackResourceSummary) {
		defer wg.Done()
		protected, err := r.checkProtection(ctx, res)
		if err != nil {
			// Treat check errors as protected to be safe
			io.Logger.Warn().Msgf("[%v]: Failed to check deletion protection for %s (%s), treating as protected: %v",
				aws.ToString(stackName), aws.ToString(res.LogicalResourceId), aws.ToString(res.PhysicalResourceId), err)
			protected = true
		}
		if protected {
			mu.Lock()
			results = append(results, checkResult{resource: res, protected: true})
			mu.Unlock()
		}
	}

	for _, res := range resources {
		if !r.isTargetResourceType(aws.ToString(res.ResourceType)) {
			continue
		}
		if res.ResourceStatus == types.ResourceStatusDeleteComplete {
			continue
		}
		wg.Add(1)
		go checkResource(res)
	}
	wg.Wait()

	protected := make([]protectedResource, 0, len(results))
	for _, result := range results {
		protected = append(protected, protectedResource{
			resourceType:       aws.ToString(result.resource.ResourceType),
			logicalResourceId:  aws.ToString(result.resource.LogicalResourceId),
			physicalResourceId: aws.ToString(result.resource.PhysicalResourceId),
		})
	}
	return protected
}

func (r *DeletionProtectionRemover) checkProtection(ctx context.Context, resource types.StackResourceSummary) (bool, error) {
	physicalId := resource.PhysicalResourceId
	switch aws.ToString(resource.ResourceType) {
	case resourcetype.Ec2Instance:
		return r.ec2Client.CheckTerminationProtection(ctx, physicalId)
	case resourcetype.RdsDBInstance:
		return r.rdsClient.CheckDBInstanceDeletionProtection(ctx, physicalId)
	case resourcetype.RdsDBCluster:
		return r.rdsClient.CheckDBClusterDeletionProtection(ctx, physicalId)
	case resourcetype.CognitoUserPool:
		return r.cognitoClient.CheckUserPoolDeletionProtection(ctx, physicalId)
	case resourcetype.LogsLogGroup:
		return r.logsClient.CheckLogGroupDeletionProtection(ctx, physicalId)
	case resourcetype.Elbv2LoadBalancer:
		return r.elbv2Client.CheckLoadBalancerDeletionProtection(ctx, physicalId)
	default:
		return false, nil
	}
}

func (r *DeletionProtectionRemover) disableProtections(ctx context.Context, stackName *string, resources []protectedResource) error {
	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup

	for _, res := range resources {
		wg.Add(1)
		go func(pr protectedResource) {
			defer wg.Done()
			if err := r.disableProtection(ctx, pr); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %s (physical: %s): %w", pr.resourceType, pr.logicalResourceId, pr.physicalResourceId, err))
				mu.Unlock()
				return
			}
			io.Logger.Info().Msgf("[%v]: Disabled deletion protection for %s: %s (physical: %s)",
				aws.ToString(stackName), pr.resourceType, pr.logicalResourceId, pr.physicalResourceId)
		}(res)
	}
	wg.Wait()

	if len(errs) > 0 {
		messages := make([]string, 0, len(errs))
		for _, err := range errs {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("DeletionProtectionError: failed to disable deletion protection:\n  %s", strings.Join(messages, "\n  "))
	}

	return nil
}

func (r *DeletionProtectionRemover) disableProtection(ctx context.Context, pr protectedResource) error {
	switch pr.resourceType {
	case resourcetype.Ec2Instance:
		return r.ec2Client.DisableTerminationProtection(ctx, aws.String(pr.physicalResourceId))
	case resourcetype.RdsDBInstance:
		return r.rdsClient.DisableDBInstanceDeletionProtection(ctx, aws.String(pr.physicalResourceId))
	case resourcetype.RdsDBCluster:
		return r.rdsClient.DisableDBClusterDeletionProtection(ctx, aws.String(pr.physicalResourceId))
	case resourcetype.CognitoUserPool:
		return r.cognitoClient.DisableUserPoolDeletionProtection(ctx, aws.String(pr.physicalResourceId))
	case resourcetype.LogsLogGroup:
		return r.logsClient.DisableLogGroupDeletionProtection(ctx, aws.String(pr.physicalResourceId))
	case resourcetype.Elbv2LoadBalancer:
		return r.elbv2Client.DisableLoadBalancerDeletionProtection(ctx, aws.String(pr.physicalResourceId))
	default:
		return nil
	}
}

func (r *DeletionProtectionRemover) isTargetResourceType(resourceType string) bool {
	switch resourceType {
	case resourcetype.Ec2Instance,
		resourcetype.RdsDBInstance,
		resourcetype.RdsDBCluster,
		resourcetype.CognitoUserPool,
		resourcetype.LogsLogGroup,
		resourcetype.Elbv2LoadBalancer:
		return true
	default:
		return false
	}
}

func (r *DeletionProtectionRemover) buildProtectionError(resources []protectedResource) error {
	lines := make([]string, 0, len(resources))
	for _, res := range resources {
		lines = append(lines, fmt.Sprintf("- %s: %s (physical: %s)", res.resourceType, res.logicalResourceId, res.physicalResourceId))
	}
	return fmt.Errorf("DeletionProtectionError: the following resources have deletion protection enabled:\n  %s\nuse the -f option to force disable deletion protection and delete the stack", strings.Join(lines, "\n  "))
}
