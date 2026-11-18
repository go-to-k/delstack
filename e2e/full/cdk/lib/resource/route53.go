package resource

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Record sets are created outside of CFn by deploy.go, so the HostedZone fails to delete.
// The domain is not delegated, so the zone never serves real DNS traffic.
func NewRoute53HostedZone(scope constructs.Construct, resourcePrefix string) {
	awsroute53.NewCfnHostedZone(scope, jsii.String("HostedZone"), &awsroute53.CfnHostedZoneProps{
		Name: jsii.String(strings.ToLower(resourcePrefix) + ".delstack-e2e.com"),
	})
}
