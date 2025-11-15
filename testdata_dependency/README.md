# Delstack Test Environment - Dependency Graph

This directory contains a tool for creating a test environment to verify `delstack`'s dependency graph handling and concurrent deletion capabilities. This tool (`deploy.go`) deploys 6 CloudFormation stacks with complex multi-level dependencies.

## Stack Dependency Structure

The test environment creates the following stack dependency structure:

```
Stack Configuration:
  A (Export: ExportA)
  B (Export: ExportB)
  C (Import: ExportA, Export: ExportC)
  D (Import: ExportA, Export: ExportD)
  E (Import: ExportB, Export: ExportE)
  F (Import: ExportC, ExportD, ExportE)

Dependencies:
  C → A
  D → A
  E → B
  F → C
  F → D
  F → E
```

### Expected Deletion Flow

When deleting all 6 stacks with `delstack`, the following concurrent deletion flow should occur:

1. **Step 1**: Start deleting Stack F (reverse in-degree 0)
2. **Step 2**: F completes → C, D, E become ready (reverse in-degree 0)
3. **Step 3**: Start deleting C, D, E concurrently
4. **Step 4**: E completes → B becomes ready
5. **Step 5**: Start deleting B (doesn't wait for C, D!)
6. **Step 6**: C completes → A's reverse in-degree decreases (2→1)
7. **Step 7**: D completes → A becomes ready (reverse in-degree 0)
8. **Step 8**: Start deleting A
9. **Step 9**: B and A complete → all done

This demonstrates `delstack`'s ability to:
- Build correct dependency graphs from CloudFormation exports/imports
- Delete stacks in the correct order based on dependencies
- Maximize parallelism by deleting independent stacks concurrently

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script **with AWS CDK for Go**. So you need to install AWS CDK.

```bash
npm install -g aws-cdk@latest
```

This script creates 6 CloudFormation stacks (A, B, C, D, E, F) with the dependency structure described above. Each stack contains:

- A minimal S3 bucket resource
- CloudFormation Exports (for stacks A, B, C, D, E)
- CloudFormation Imports using `Fn::ImportValue` (for stacks C, D, E, F)
- S3 objects for testing deletion

```bash
go run testdata_dependency/deploy.go -s <stage> [-p <profile>]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, defaults to random name)
- `-p <profile>` : AWS CLI profile name to use (optional)
- `-r` : Make all resources RETAIN to test `-f` option for delstack (optional)

### Using the Makefile

For convenience, you can also use the Makefile target:

```bash
# Deploy with default stage and profile
make testgen_dependency

# Deploy with custom stage and profile
make testgen_dependency OPT="-s my-stage -p my-profile"

# Deploy the stacks with all RETAIN resources
make testgen_dependency OPT="-r"
```
