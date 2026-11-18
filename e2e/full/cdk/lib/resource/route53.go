package resource

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Record sets are created and DNSSEC signing is enabled outside of CFn by deploy.go, so the HostedZone fails to delete.
// The domain is not delegated, so the zone never serves real DNS traffic.
func NewRoute53HostedZone(scope constructs.Construct, resourcePrefix string) {
	hostedZone := awsroute53.NewCfnHostedZone(scope, jsii.String("HostedZone"), &awsroute53.CfnHostedZoneProps{
		Name: jsii.String(strings.ToLower(resourcePrefix) + ".delstack-e2e.com"),
	})

	// The key used by the key-signing key that deploy.go creates for DNSSEC signing
	key := awskms.NewKey(scope, jsii.String("HostedZoneDNSSECKey"), &awskms.KeyProps{
		KeySpec:       awskms.KeySpec_ECC_NIST_P256,
		KeyUsage:      awskms.KeyUsage_SIGN_VERIFY,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		PendingWindow: awscdk.Duration_Days(jsii.Number(7)),
	})
	// Delete the hosted zone before the key, so that DNSSEC signing is disabled while the key is still available.
	hostedZone.AddDependency(key.Node().DefaultChild().(awscdk.CfnResource))
	key.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Principals: &[]awsiam.IPrincipal{awsiam.NewServicePrincipal(jsii.String("dnssec-route53.amazonaws.com"), nil)},
		Actions:    jsii.Strings("kms:DescribeKey", "kms:GetPublicKey", "kms:Sign"),
		Resources:  jsii.Strings("*"),
	}), nil)
	key.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Principals: &[]awsiam.IPrincipal{awsiam.NewServicePrincipal(jsii.String("dnssec-route53.amazonaws.com"), nil)},
		Actions:    jsii.Strings("kms:CreateGrant"),
		Resources:  jsii.Strings("*"),
		Conditions: &map[string]any{
			"Bool": map[string]any{"kms:GrantIsForAWSResource": true},
		},
	}), nil)
}
