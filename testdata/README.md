# Delstack Test Environment

This directory contains a tool for creating a test environment for `delstack`. This tool (`deploy.go`) deploys AWS CloudFormation stacks with various resources that can be used to test the stack deletion functionality of `delstack`.

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script. This script creates a CloudFormation stack containing various resources that typically cause deletion issues, including:

- S3 buckets with contents
- S3 Express Directory buckets
- IAM groups
- ECR repositories with images
- AWS Backup vaults with recovery points
- And more

```bash
go run testdata/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (required)
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
- The script includes 2 `AWS::S3Express::DirectoryBucket` resources; an AWS account can have at most 10 directory buckets per region.
- The script includes 2 `AWS::IAM::Group` resources; one IAM user can only belong to 10 IAM groups.

## Testing Delstack with the Test Environment

After deploying test stacks, you can test the `delstack` functionality by attempting to delete these stacks. The resources created are specifically designed to simulate real-world scenarios where CloudFormation deletion fails.

Common resources that will cause DELETE_FAILED in CloudFormation, but can be force-deleted by `delstack`:

1. **S3 buckets with content**
2. **IAM Groups with users**
3. **ECR repositories with images**
4. **Backup vaults with recovery points**
5. **Nested stacks with DELETE_FAILED resources**

## Implementation Details

The `deploy.go` script uses AWS SDK Go v2 for most operations, but also uses some shell commands (AWS CLI, SAM CLI, Docker) for certain operations. The script:

1. Deploys CloudFormation stacks using SAM CLI
2. Creates test resources across various services
3. Populates S3 buckets with objects
4. Pushes images to ECR repositories
5. Creates backup recovery points

## Areas for Improvement

The current implementation could be enhanced by:

1. **Replacing SAM commands** - Use CloudFormation AWS SDK directly instead of SAM commands
2. **Streamlining Docker interactions** - Implement ECR login and image pushing using ECR SDK and Docker engine API integration
3. **Adding more resource types** - Extend with additional resource types that are challenging to delete

## Cleaning Up

When you're done testing, you can use `delstack` itself to clean up the test environment:

```bash
delstack -s dev-<stage>-TestStack -p <profile>
```

Or use the interactive mode:

```bash
delstack -i -p <profile>
```
