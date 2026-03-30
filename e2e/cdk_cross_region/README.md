# Delstack Test Environment - CDK Cross-Region

This directory contains a test environment for the `delstack cdk` subcommand's cross-region deletion capability. It tests two different cross-region dependency patterns with 4 stacks total.

## Stack Structure

### Pattern A: AddDependency + Export/Import

```text
EdgeStackA (us-east-1)        -- S3 Bucket, Export: ExportFromEdgeA
MainStackA (ap-northeast-1)   -- S3 Bucket, Import: ExportFromEdgeA

Dependency: MainStackA -> EdgeStackA (via AddDependency + Fn::ImportValue)
```

### Pattern B: crossRegionReferences (SSM-backed)

```text
EdgeStackB (us-east-1)        -- S3 Bucket
MainStackB (ap-northeast-1)   -- S3 Bucket, references EdgeStackB bucket ARN

Dependency: MainStackB -> EdgeStackB (auto-resolved by CDK via SSM parameters)
```

## Test Workflow

### 1. Deploy test stacks

```bash
go run e2e/cdk_cross_region/deploy.go [-s <stage>] [-p <profile>] [-r]
```

#### Options

- `-s <stage>` : Stage name (optional, defaults to random name)
- `-p <profile>` : AWS CLI profile name (optional)
- `-r` : Make all resources RETAIN to test `-f` option (optional)

### 2. Delete with `delstack cdk`

```bash
cd e2e/cdk_cross_region/cdk
delstack cdk -c PJ_PREFIX=<stage> -c RETAIN_MODE=false -f -y
```

### Expected behavior

1. Parses `cdk.out/manifest.json` to discover 4 stacks in 2 regions
2. Detects cross-region dependencies for both patterns
3. Deletes MainStackA and MainStackB first (ap-northeast-1)
4. Deletes EdgeStackA and EdgeStackB after (us-east-1)

### Using the Makefile

```bash
make e2e_cdk_cross_region
make e2e_cdk_cross_region OPT="-p my-profile"

# With RETAIN resources
make e2e_cdk_cross_region_retain
```
