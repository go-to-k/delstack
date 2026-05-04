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

1. `cdk deploy` a minimal VPC + VPC-attached Lambda topology.
2. Invokes the Lambda once so AWS Lambda provisions the VPC (Hyperplane) ENIs.
3. Triggers an ordinary CloudFormation `DeleteStack` (not `delstack`) so CFN itself drives Lambda deletion and then fails on Subnet / SecurityGroup.
4. Polls until the stack reaches `DELETE_FAILED`. If AWS Lambda happens to release the ENIs in time and the stack reaches `DELETE_COMPLETE` instead, the script reports failure (the orphan ENI condition was not actually reproduced, so the test would not be meaningful).

After step 4, the stack is in the exact state described in issue #637. Running `delstack -s <stage>` then exercises the new operators end-to-end.

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
