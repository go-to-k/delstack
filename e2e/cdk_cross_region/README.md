# Delstack Test Environment - CDK Cross-Region

This directory contains a test environment for the `delstack cdk` subcommand's cross-region deletion capability. It uses CDK's `crossRegionReferences` feature to create stacks in different regions with cross-region SSM parameter-backed references.

## Stack Structure

```text
EdgeStack (us-east-1)        -- S3 Bucket (non-empty)
MainStack (ap-northeast-1)   -- S3 Bucket (non-empty), tagged with EdgeStack bucket ARN via crossRegionReferences

Dependency: MainStack -> EdgeStack (cross-region)
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

For RETAIN mode:

```bash
delstack cdk -c PJ_PREFIX=<stage> -c RETAIN_MODE=true -f -y
```

### Expected behavior

1. `delstack cdk` runs `npx cdk synth --quiet -c PJ_PREFIX=<stage>`
2. Parses `cdk.out/manifest.json` to discover 2 stacks in different regions
3. Detects cross-region dependency (MainStack depends on EdgeStack)
4. Deletes MainStack (ap-northeast-1) first, then EdgeStack (us-east-1)
5. Creates separate AWS sessions for each region

### Using the Makefile

```bash
make testgen_cdk_cross_region
make testgen_cdk_cross_region OPT="-s my-stage -p my-profile"

# Combined deploy + delete
make e2e_cdk_cross_region
make e2e_cdk_cross_region OPT="-p my-profile"

# With RETAIN resources
make e2e_cdk_cross_region_retain
make e2e_cdk_cross_region_retain OPT="-p my-profile"
```
