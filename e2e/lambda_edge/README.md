# Delstack Test Environment - Lambda@Edge

This directory contains a tool for creating a test environment specifically for `delstack`'s Lambda@Edge deletion feature. This tool (`deploy.go`) deploys AWS CloudFormation stacks with CloudFront distributions using Lambda@Edge functions to test the automatic retry-based deletion functionality.

## Why a Separate E2E Directory

This test environment is separated from `e2e/full/` for the following reasons:

- **Deployment time**: CloudFront distribution creation takes 3-5 minutes, significantly slower than other resources
- **Deletion wait time**: Lambda@Edge replicas take ~15-20 minutes to be cleaned up by AWS after CloudFront distribution deletion. The operator retries deletion every 30 seconds for up to 60 minutes
- **Region constraint**: Lambda@Edge requires `us-east-1`, which may differ from other test environments

Including this in the full E2E test would make the entire test suite significantly slower.

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script **with AWS CDK for Go**. So you need to install AWS CDK.

```bash
npm install -g aws-cdk@latest
```

This script creates a CloudFormation stack containing:

- S3 bucket as CloudFront origin
- CloudFront distribution with Lambda@Edge function (origin-response)
- Lambda@Edge function (Node.js)

```bash
go run e2e/lambda_edge/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)

### Using the Makefile

For convenience, you can also use the Makefile target:

```bash
# Deploy with auto-generated stage name
make testgen_lambda_edge

# Deploy with custom stage and profile
make testgen_lambda_edge OPT="-s my-stage -p my-profile"
```

### Testing Lambda@Edge Deletion

After deployment, you can test the Lambda@Edge deletion feature by running:

```bash
delstack -s <stack-name>
```

The operator will:
1. Attempt to delete the Lambda function after CloudFormation's initial deletion fails
2. Detect the "replicated function" error from Lambda@Edge
3. Log a message about waiting for replica cleanup
4. Retry deletion every 30 seconds until successful (timeout: 60 minutes)

### Notes

- This test environment focuses on Lambda@Edge replica cleanup only
- Stack must be deployed to `us-east-1` (Lambda@Edge requirement)
- Expect ~15-20 minutes for the full deletion cycle after `delstack` starts
