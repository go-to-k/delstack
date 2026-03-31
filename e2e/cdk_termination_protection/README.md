# Delstack Test Environment - CDK TerminationProtection

This directory contains a test environment for the `delstack cdk` subcommand with TerminationProtection-enabled stacks. It verifies that `delstack cdk -f` can disable TerminationProtection and delete stacks.

## Stack Structure

```
TPStack     -- SNS Topic (TerminationProtection: enabled)
NormalStack -- SNS Topic (TerminationProtection: disabled)
```

## Test Workflow

### 1. Deploy test stacks

```bash
go run e2e/cdk_termination_protection/deploy.go [-s <stage>] [-p <profile>]
```

#### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, defaults to random name)
- `-p <profile>` : AWS CLI profile name to use (optional)

### 2. Delete with `delstack cdk`

```bash
cd e2e/cdk_termination_protection/cdk
delstack cdk -c PJ_PREFIX=<stage> -f -y
```

### Expected behavior

1. `delstack cdk` runs `npx cdk synth --quiet -c PJ_PREFIX=<stage>`
2. Parses `cdk.out/manifest.json` to discover 2 stacks
3. Detects TPStack has TerminationProtection enabled
4. With `-f`, disables TerminationProtection and deletes both stacks
5. Without `-f`, returns TerminationProtectionError for TPStack

### Using the Makefile

```bash
make testgen_cdk_termination_protection
make testgen_cdk_termination_protection OPT="-s my-stage -p my-profile"

# Combined deploy + delete
make e2e_cdk_termination_protection
make e2e_cdk_termination_protection OPT="-p my-profile"
```
