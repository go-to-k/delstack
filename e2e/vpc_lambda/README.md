# Delstack Test Environment - VPC Lambda Orphan ENI

This directory reproduces the real user-facing scenario of issue #637 and verifies that `delstack` can recover from it.

## The two scenarios (and which one this E2E covers)

`delstack` already handles VPC Lambda fine **when the user runs `delstack` first**: the `LambdaVPCDetacher` preprocessor detaches each Lambda from its VPC before CloudFormation deletes the function, so the ENIs are released up front and the Subnet / SecurityGroup deletion never gets blocked.

The hard case is the other order:

1. The user first runs an ordinary `cdk destroy` / `aws cloudformation delete-stack`.
2. CloudFormation deletes the Lambda function successfully (`DELETE_COMPLETE`).
3. CloudFormation immediately tries to delete the Subnet / SecurityGroup, but AWS Lambda has not yet released the function's VPC ENIs (release is asynchronous, can take ~30 min). The ENIs stay in `available` state and block deletion with messages like:
   - `The subnet 'subnet-...' has dependencies and cannot be deleted.`
   - `resource sg-... has a dependent object`
4. The stack ends in `DELETE_FAILED`. The Lambda is gone, so the preprocessor can no longer help — only an operator that can clean up orphan ENIs by Subnet ID / SecurityGroup ID can recover this stack.

This E2E reproduces **scenario 4** and then expects `delstack` to clean the stack via the new `EC2SubnetOperator` / `EC2SecurityGroupOperator`.

## What `deploy.go` does

We do **not** deploy a real VPC Lambda. AWS Lambda's Hyperplane ENI release is non-deterministic — sometimes the ENIs are released before CFN gets to the Subnet / SecurityGroup, the stack reaches `DELETE_COMPLETE` cleanly, and the new operator code never runs. To exercise the operator code path on every run we instead inject **synthetic ENIs** that look exactly like orphan Lambda VPC ENIs from the operator's point of view (same `description` prefix, `available` state).

1. `cdk deploy` a minimal VPC + private Subnet + SecurityGroup. The stack outputs `PrivateSubnetId` and `LambdaSgId`.
2. `CreateNetworkInterface` (x2) on that Subnet+SG with description `AWS Lambda VPC ENI-<stage>-<n>`. The ENIs are unattached and stay in `available` state.

`delstack` itself drives `DeleteStack` and waits for `DELETE_FAILED` via its internal CloudFormation delete waiter (~75 min cap), so this script does not call `DeleteStack` itself. CFN's first delete pass hits `DependencyViolation` on the SecurityGroup / Subnet because of the synthetic ENIs, the stack lands in `DELETE_FAILED`, and the new `EC2SubnetOperator` / `EC2SecurityGroupOperator` remove the synthetic ENIs (matched by the description prefix) and then the Subnet / SecurityGroup themselves.

## Test Stack Deployment

You need AWS CDK installed:

```bash
npm install -g aws-cdk@latest
```

```bash
go run e2e/vpc_lambda/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as the stack name (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)

### Using the Makefile

```bash
# Deploy + setup orphan ENI with auto-generated stage name
make testgen_vpc_lambda

# Deploy + setup, then force delete with delstack
make e2e_vpc_lambda

# Custom stage / profile
make e2e_vpc_lambda STAGE=my-stage OPT="-p my-profile"
```
