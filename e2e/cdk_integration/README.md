# Delstack Test Environment - CDK Integration

This directory contains a test environment for the `delstack cdk` subcommand. It verifies that `delstack` can synthesize a CDK app and delete all stacks with dependency resolution.

## Stack Structure

```
BaseStack (Export: ExportFromBase) -- S3 Bucket (non-empty)
AppStack  (Import: ExportFromBase) -- S3 Bucket (non-empty)

Dependency: AppStack -> BaseStack
```

## Test Workflow

### 1. Deploy test stacks

```bash
go run e2e/cdk_integration/deploy.go [-s <stage>] [-p <profile>] [-r]
```

This deploys 2 stacks and uploads objects to their S3 buckets (making them non-empty so normal deletion fails).

#### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, defaults to random name)
- `-p <profile>` : AWS CLI profile name to use (optional)
- `-r` : Make all resources RETAIN to test `-f` option for delstack (optional)

### 2. Delete with `delstack cdk`

```bash
cd e2e/cdk_integration/cdk
delstack cdk -c PJ_PREFIX=<stage> -c RETAIN_MODE=false -f -y
```

For RETAIN mode:

```bash
delstack cdk -c PJ_PREFIX=<stage> -c RETAIN_MODE=true -f -y
```

### Expected behavior

1. `delstack cdk` runs `npx cdk synth --quiet -c PJ_PREFIX=<stage>`
2. Parses `cdk.out/manifest.json` to discover 2 stacks
3. Deletes AppStack first (dependent), then BaseStack
4. Force mode empties S3 buckets before deletion
5. With RETAIN mode, DeletionPolicy is removed from templates before deletion

### Using the Makefile

```bash
make testgen_cdk_integration
make testgen_cdk_integration OPT="-s my-stage -p my-profile"

# Combined deploy + delete
make e2e_cdk_integration
make e2e_cdk_integration OPT="-p my-profile"

# With RETAIN resources
make e2e_cdk_integration_retain
make e2e_cdk_integration_retain OPT="-p my-profile"
```
