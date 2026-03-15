# Delstack Test Environment - Preprocessor (Lambda VPC)

This directory contains a tool for creating a test environment specifically for `delstack`'s Lambda VPC detachment preprocessing feature. This tool (`deploy.go`) deploys AWS CloudFormation stacks with VPC-attached Lambda functions to test the automatic VPC detachment functionality.

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script **with AWS CDK for Go**. So you need to install AWS CDK.

```bash
npm install -g aws-cdk@latest
```

This script creates a CloudFormation stack containing:

- VPC with isolated subnets (2 AZs)
- Lambda functions attached to VPC (IPv6 disabled)
- Lambda functions attached to VPC (IPv6 enabled)
- Lambda functions without VPC
- Nested stack with VPC-attached Lambda functions (both IPv6 enabled and disabled)

```bash
go run e2e/testdata_preprocessor/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)

### Using the Makefile

For convenience, you can also use the Makefile target:

```bash
# Deploy with auto-generated stage name
make testgen_preprocessor

# Deploy with custom stage and profile
make testgen_preprocessor OPT="-s my-stage -p my-profile"
```

### Testing Lambda VPC Detachment

After deployment, you can test the Lambda VPC auto-detachment feature by running:

```bash
delstack -s <stack-name>
```

The preprocessor will automatically:
1. Detect Lambda functions attached to VPC
2. Disable IPv6 for dual-stack Lambda functions (if enabled)
3. Remove VPC configuration from all Lambda functions
4. Proceed with stack deletion (much faster without ENI cleanup wait time)

### Notes

- This test environment focuses on Lambda VPC configurations only
- The stack includes both root and nested stacks with VPC-attached Lambda functions
- VPC uses isolated subnets with no NAT Gateway to minimize costs
- The preprocessor runs automatically without any additional flags
