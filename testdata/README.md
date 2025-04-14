# Delstack Test Environment

This directory contains a tool for creating a test environment for `delstack`. This tool (`deploy.go`) deploys AWS CloudFormation stacks with various resources that can be used to test the stack deletion functionality of `delstack`.

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script **with AWS CDK for Go**. So you need to install AWS CDK.

```bash
npm install -g aws-cdk@latest
```

This script creates a CloudFormation stack containing various resources that typically cause deletion issues, including:

- S3 buckets with contents
- S3 Express Directory buckets with contents
- S3 Table buckets with namespaces and tables
- IAM groups with users
- ECR repositories with images
- AWS Backup vaults with recovery points
- And more

```bash
go run testdata/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional)
- `-p <profile>` : AWS CLI profile name to use (optional)

### Using the Makefile

For convenience, you can also use the Makefile target:

```bash
# Deploy with default stage and profile
make testgen

# Deploy with custom stage and profile
make testgen OPT="-s my-stage -p my-profile"
```

### Notes

- Due to AWS quota limitations, only up to 5 test stacks can be created simultaneously with this script.
- The script includes 2 `AWS::IAM::Group` resources only; one IAM user (`DelstackTestUser`) can only belong to 10 IAM groups, and we want to be able to make up to 5 stacks.
- The script includes 2 `AWS::S3Tables::TableBucket` resources; an AWS account can have at most 10 table buckets per region.
