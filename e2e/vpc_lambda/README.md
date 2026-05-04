# Delstack Test Environment - VPC Lambda Orphan ENI

This directory contains a tool for reproducing the orphan AWS Lambda VPC ENI scenario that blocks `AWS::EC2::Subnet` and `AWS::EC2::SecurityGroup` deletion (issue #637), and verifies that `delstack` can force-delete such a stuck stack.

## What this scenario reproduces

When a CloudFormation stack contains a VPC-attached Lambda function and the function is deleted (out-of-band, or by CloudFormation), AWS Lambda releases the function's VPC ENIs **asynchronously** (~30 minutes). During that window, the ENIs remain in `available` state attached to the original Subnet / SecurityGroup, blocking their deletion with errors like:

- `The subnet 'subnet-...' has dependencies and cannot be deleted.`
- `resource sg-... has a dependent object`

`deploy.go`:

1. `cdk deploy` a minimal VPC + Lambda VPC topology.
2. Invokes the Lambda once so AWS Lambda provisions the VPC ENIs.
3. Deletes the Lambda function via the Lambda SDK directly (NOT via CloudFormation), leaving orphan ENIs in `available` state.

After this, running `delstack -s <stage>` should detect the orphan Lambda VPC ENIs, delete them, and then delete the Subnet / SecurityGroup themselves so the stack can be removed.

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
