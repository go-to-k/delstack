package resource

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewHostedZone creates a Route 53 HostedZone. Resource record set protections are created
// outside of CFn by deploy.go, since CloudFormation does not support them.
// The domain is not delegated, so the zone never serves real DNS traffic.
func NewHostedZone(scope constructs.Construct, resourcePrefix string) {
	awsroute53.NewCfnHostedZone(scope, jsii.String("HostedZone"), &awsroute53.CfnHostedZoneProps{
		Name: jsii.String(strings.ToLower(resourcePrefix) + ".delstack-e2e.com"),
	})
}

// NewExternalHostedZoneRecordSets creates record sets in a hosted zone outside the stack, created by deploy.go.
// Resource record set protections on them are created outside of CFn by deploy.go.
func NewExternalHostedZoneRecordSets(scope constructs.Construct, hostedZoneId string, hostedZoneName string) {
	awsroute53.NewCfnRecordSet(scope, jsii.String("ExternalRecordSet"), &awsroute53.CfnRecordSetProps{
		HostedZoneName:  jsii.String(hostedZoneName + "."),
		Name:            jsii.String("external-a." + hostedZoneName),
		Type:            jsii.String("A"),
		Ttl:             jsii.String("300"),
		ResourceRecords: jsii.Strings("192.0.2.1"),
	})

	awsroute53.NewCfnRecordSetGroup(scope, jsii.String("ExternalRecordSetGroup"), &awsroute53.CfnRecordSetGroupProps{
		HostedZoneId: jsii.String(hostedZoneId),
		RecordSets: &[]*awsroute53.CfnRecordSetGroup_RecordSetProperty{
			{
				Name:            jsii.String("external-b." + hostedZoneName),
				Type:            jsii.String("TXT"),
				Ttl:             jsii.String("300"),
				ResourceRecords: jsii.Strings("\"delstack-e2e\""),
			},
		},
	})
}
