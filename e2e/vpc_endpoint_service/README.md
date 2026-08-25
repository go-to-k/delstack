# Delstack Test Environment - VPC Endpoint Service

This directory reproduces the real user-facing scenario of issue #656 and verifies that `delstack` can recover from it.

## Why the deletion fails

AWS refuses to delete a VPC endpoint service (AWS PrivateLink or Gateway Load Balancer endpoint service) while any endpoint connection is in the `Available` or `PendingAcceptance` state:

> You can't delete an endpoint service if there are any endpoints connected to the endpoint service that are in the `available` or `pending-acceptance` state.
> — [Delete an endpoint service](https://docs.aws.amazon.com/vpc/latest/privatelink/delete-endpoint-service.html)

Those endpoints belong to **service consumers**, so they are outside the stack that owns the endpoint service. CloudFormation cannot clear them, the delete fails with `ExistingVpcEndpointConnections`, and the stack ends in `DELETE_FAILED`. Only an operator that rejects the connections first can recover the stack, which is what the new `EC2VPCEndpointServiceOperator` does.

Rejecting a connection only severs it from this endpoint service. The consumer's endpoint itself is left untouched.

## What `deploy.go` does

1. `cdk deploy` a provider stack: VPC, an internal Network Load Balancer, and the endpoint service hosted on it (`AcceptanceRequired: false`). The stack outputs `EndpointServiceId` and `EndpointServiceName`.
2. **Out of band** (SDK only, so CloudFormation knows nothing about them): `ModifyVpcEndpointServicePermissions` to allow this account, then a small consumer VPC, a subnet in an Availability Zone the endpoint service supports, and an interface VPC endpoint pointing at the endpoint service. Since acceptance is not required, the connection reaches `Available` on its own.
3. Wait until the connection actually shows up on the endpoint service, so the stack is never deleted before anything blocks it.

Everything that holds the connection open has to live **outside** the CDK stack, in two separate ways, and getting either wrong makes the fixture silently stop reproducing the bug:

- **The consumer endpoint.** Inside the same stack it would be deleted by CloudFormation in the correct dependency order and never block anything; inside the stack's own VPC it would additionally block the Subnet and VPC deletion, which is a different (unsupported) failure.
- **The endpoint service permission.** Setting CDK's `allowedPrincipals` adds an `AWS::EC2::VPCEndpointServicePermissions` resource to the stack, and CloudFormation deletes it *before* the endpoint service. Revoking the principal makes AWS reject the consumer connection on its own, so the endpoint service then deletes cleanly and the stack never reaches `DELETE_FAILED`. The first version of this fixture did exactly that and passed while testing nothing, which is why the permission is granted with the SDK instead.

`delstack` itself drives `DeleteStack` and waits for `DELETE_FAILED` via its internal CloudFormation delete waiter, so this script does not call `DeleteStack`. CFN's first delete pass fails on the endpoint service, the stack lands in `DELETE_FAILED`, and the `EC2VPCEndpointServiceOperator` rejects the connection and deletes the endpoint service. CloudFormation then removes the NLB, subnets and VPC normally.

## Cleaning up the consumer resources

Since rejecting a connection does not delete the consumer's endpoint, the consumer VPC, subnet and endpoint survive the `delstack` run. Run the script again with `-d` to delete them (it only touches resources tagged `delstack-e2e-vpc-endpoint-service=<stage>`):

```bash
go run e2e/vpc_endpoint_service/deploy.go -s <stage> -d
```

`make e2e_vpc_endpoint_service` does this automatically after `delstack` finishes.

## Test Stack Deployment

You need AWS CDK installed:

```bash
npm install -g aws-cdk@latest
```

```bash
go run e2e/vpc_endpoint_service/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as the stack name (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)
- `-d` : Delete the out-of-band consumer resources of the stage instead of deploying

### Using the Makefile

```bash
# Deploy + connect a consumer endpoint with an auto-generated stage name
make testgen_vpc_endpoint_service

# Deploy + connect, delete with delstack, then remove the consumer resources
make e2e_vpc_endpoint_service

# Custom stage / profile
make e2e_vpc_endpoint_service STAGE=my-stage OPT="-p my-profile"
```
