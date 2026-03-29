# Delstack Test Environment - CDK Stage

This directory contains a test environment for the `delstack cdk` subcommand's CDK Stage support. When CDK apps use `Stage` constructs, stacks are placed inside nested Cloud Assemblies (`assembly-*/`). `delstack cdk` must recursively parse these nested manifests to discover all stacks.

## Stack Structure

```text
MyStage1/AppStack (us-east-1)        -- S3 Bucket (non-empty)
MyStage2/AppStack (ap-northeast-1)   -- S3 Bucket (non-empty)
```

Each stack is inside a CDK Stage, resulting in nested Cloud Assemblies:

```text
cdk.out/
  manifest.json                          <- top-level, contains cdk:cloud-assembly refs
  assembly-MyStage1/
    manifest.json                        <- contains MyStage1/AppStack
  assembly-MyStage2/
    manifest.json                        <- contains MyStage2/AppStack
```

## Test Workflow

### 1. Deploy test stacks

```bash
go run e2e/cdk_stage/deploy.go [-s <stage>] [-p <profile>]
```

### 2. Delete with `delstack cdk`

```bash
cd e2e/cdk_stage/cdk
delstack cdk -c PJ_PREFIX=<stage> -f -y
```

### Expected behavior

1. `delstack cdk` runs `npx cdk synth --quiet -c PJ_PREFIX=<stage>`
2. Parses top-level `manifest.json`, finds `cdk:cloud-assembly` artifacts
3. Recursively parses `assembly-MyStage1/manifest.json` and `assembly-MyStage2/manifest.json`
4. Discovers 2 stacks in different regions
5. Deletes both stacks in parallel (no dependencies between them)

### Using the Makefile

```bash
make testgen_cdk_stage
make testgen_cdk_stage OPT="-s my-stage -p my-profile"

# Combined deploy + delete
make e2e_cdk_stage
make e2e_cdk_stage OPT="-p my-profile"
```
