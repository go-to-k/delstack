# Delstack Test Data Generation

This directory contains tools for generating test data for Delstack. You can use these tools to easily generate test data for S3 buckets.

## How to Generate Test Data

You can generate test data using the following make command:

```bash
make testgen OPT="[options]"
```

### Options

- `-b <bucket>` : S3 bucket name to upload test data (required)
- `-p <profile>` : AWS CLI profile name to use
- `-r <region>` : AWS region (default: us-east-1)
- `-n <num>` : Number of files to generate (default: 100)
- `-t <type>` : Bucket type: general (standard S3) or directory (S3 Express) (default: general)
- `-prefix <prefix>` : Prefix for S3 object keys
- `-o <size>` : Size of each object in bytes (default: 1024)

## Usage Examples

### Generate test data for a standard S3 bucket

```bash
make testgen OPT="-b my-test-bucket -n 500 -o 2048 -prefix test-data/"
```

### Generate test data for an S3 Express Directory bucket

```bash
make testgen OPT="-b my-bucket--az4--x-s3 -t directory -n 20 -o 512"
```

### Use specific AWS profile and region

```bash
make testgen OPT="-b my-test-bucket -p devprofile -r us-west-2 -n 100"
```

## Display help for test data generation

```bash
make testgen_help
```

## Deploy Test Stacks

You can also deploy CloudFormation test stacks using `deploy.go`. This script deploys a CloudFormation stack containing test S3 buckets, IAM groups, ECR repositories, backup vaults, and more.

```bash
go run testdata/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name (required)
- `-p <profile>` : AWS CLI profile name (optional)

### Notes

- Due to AWS quota limitations, only up to 5 test stacks can be created simultaneously with this script.
- The script includes 2 `AWS::S3Express::DirectoryBucket` resources; an AWS account can have at most 10 directory buckets per region.
- The script includes 2 `AWS::IAM::Group` resources; one IAM user can only belong to 10 IAM groups.

## Areas for Improvement

The current `deploy.go` uses shell commands (AWS CLI, SAM CLI, Docker) for some operations, but the following could be improved by implementing directly with AWS SDK Go v2:

1. **S3 bucket operations** - Use `s3Client.PutObject` and `s3Client.DeleteObject` for uploading and deleting objects in S3 buckets
2. **ECR integration** - Implement ECR login and image pushing using ECR SDK and Docker engine API
3. **CloudFormation operations** - Use CloudFormation AWS SDK directly instead of SAM commands

These improvements would reduce dependency on shell commands and enhance cross-platform compatibility and error handling.
