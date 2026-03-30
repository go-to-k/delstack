# delstack

[![Go Report Card](https://goreportcard.com/badge/github.com/go-to-k/delstack)](https://goreportcard.com/report/github.com/go-to-k/delstack) ![GitHub](https://img.shields.io/github/license/go-to-k/delstack) ![GitHub](https://img.shields.io/github/v/release/go-to-k/delstack) [![ci](https://github.com/go-to-k/delstack/actions/workflows/ci.yml/badge.svg)](https://github.com/go-to-k/delstack/actions/workflows/ci.yml)

## What is

A CLI tool for deleting AWS CloudFormation stacks. Handles everything from routine deletions to stacks containing resources that fail to delete. Unlike CloudFormation's built-in `FORCE_DELETE_STACK` which leaves failed resources behind, **delstack** actually cleans them up with **no orphaned resources**.

Works with stacks from **AWS CDK**, **AWS SAM**, **AWS Amplify**, **Serverless Framework**, and any IaC tool using CloudFormation.

![delstack](https://github.com/user-attachments/assets/4f02526d-536c-4a23-81fd-10484902133f)

## Features

- **Force delete undeletable resources**: Automatically cleans up resources blocking deletion, such as non-empty S3 buckets, and [more resource types](#resource-types-that-can-be-forced-to-delete)
- **Parallel deletion with dependency resolution**: Deletes multiple stacks with maximum parallelism while respecting inter-stack dependencies
- **Interactive stack selection**: Search and select stacks in a TUI with case-insensitive filtering
- **Deletion protection handling**: Detects resource-level protection (EC2, RDS, Cognito, etc.) and stack TerminationProtection. With `-f`, automatically disables them before deletion
- **Pre-deletion optimization**: Detaches Lambda VPC configurations in parallel to eliminate ENI cleanup wait time
- **Retain policy override**: Force deletes resources with `Retain` or `RetainExceptOnCreate` deletion policies via `-f`
- **GitHub Actions support**: Available as a [GitHub Actions](#github-actions) workflow for CI/CD stack cleanup
- **[CDK integration](#cdk-integration)**: Run `delstack cdk` in a CDK app directory to synthesize, discover all stacks (including cross-region), and delete them with dependency resolution

## Install

- Homebrew

  ```bash
  brew install go-to-k/tap/delstack
  ```

- Linux, Darwin (macOS) and Windows

  ```bash
  curl -fsSL https://raw.githubusercontent.com/go-to-k/delstack/main/install.sh | sh
  delstack -h

  # To install a specific version of delstack
  # e.g. version 1.2.0
  curl -fsSL https://raw.githubusercontent.com/go-to-k/delstack/main/install.sh | sh -s "v1.2.0"
  delstack -h
  ```

- aqua

  ```bash
  aqua g -i go-to-k/delstack
  ```

- Binary
  - [Releases](https://github.com/go-to-k/delstack/releases)
- Git Clone and install (for developers)

  ```bash
  git clone https://github.com/go-to-k/delstack.git
  cd delstack
  make install
  ```

## How to use

  ```bash
  delstack [-s <stackName>] [-p <profile>] [-r <region>] [-i|--interactive] [-f|--force] [-y|--yes] [-n <concurrencyNumber>]
  ```

- -s, --stackName: optional
  - CloudFormation stack name
    - Required in non-interactive mode
    - In interactive mode, you can select stacks from the UI instead
  - **Multiple stack names can be specified.**
    - `delstack -s test1 -s test2`
    - **Multiple stacks are deleted in parallel by default, taking dependencies between stacks into account.**
    - You can limit the number of parallel deletions with the `-n` option (e.g., `delstack -s test1 -s test2 -s test3 -n 2`).
- -p, --profile: optional
  - AWS profile name
- -r, --region: optional(default: `us-east-1`)
  - AWS Region
- -i, --interactive: optional
  - Interactive Mode
- -f, --force: optional
  - Force Mode to delete stacks including resources with **deletion policy `Retain`/`RetainExceptOnCreate`**, resources with **deletion protection**, and stacks with **TerminationProtection**
- -y, --yes: optional
  - Skip confirmation prompts (e.g., TerminationProtection disable confirmation)
- -n, --concurrencyNumber: optional(default: unlimited)
  - Specify the number of parallel stack deletions. Default is unlimited (delete all stacks in parallel).

### CDK Integration

  ```bash
  delstack cdk [-s <stackName>] [-a <cdkOutPath>] [-c <key=value>] [-p <profile>] [-i] [-f] [-y] [-n <concurrencyNumber>]
  ```

- -a, --app: optional
  - Path to an existing `cdk.out` directory. When specified, `npx cdk synth` is skipped and the manifest is read directly.
- -c, --context: optional (repeatable)
  - CDK context values in `key=value` format, passed to `npx cdk synth -c key=value`.
- All global options (`-s`, `-p`, `-r`, `-i`, `-f`, `-y`, `-n`) also work with the `cdk` subcommand.
- **Requires**: [AWS CDK CLI](https://docs.aws.amazon.com/cdk/v2/guide/cli.html) installed (unless using `-a`).

  ```bash
  delstack cdk                         # Synthesize and delete all stacks
  delstack cdk -c env=dev              # With CDK context
  delstack cdk -a ./cdk.out            # Use existing cdk.out (skip synthesis)
  delstack cdk -s MyStack              # Delete specific stack
  delstack cdk -i                      # Interactive selection
  delstack cdk -f -y                   # Force delete, skip confirmation
  ```

For details on cross-region deletion, see [CDK Integration Details](#cdk-integration-details).

## Resource Types that can be forced to delete

This tool supports force deletion of the following resource types that cause `DELETE_FAILED` in normal CloudFormation stack deletion. All other resources are deleted normally, so you can use this tool for any stack.

If you need support for additional resource types, please create an issue at [GitHub](https://github.com/go-to-k/delstack/issues).

|  RESOURCE TYPE  |  DETAILS  |
| ---- | ---- |
|  AWS::S3::Bucket  |  S3 Buckets, including **non-empty buckets or buckets with Versioning enabled**.  |
|  AWS::S3Express::DirectoryBucket  |  S3 Directory Buckets for S3 Express One Zone, including non-empty buckets.  |
|  AWS::S3Tables::TableBucket  |  S3 Table Buckets, including buckets with any namespaces or tables.  |
|  AWS::S3Tables::Namespace  |  S3 Table Namespaces, including namespaces with any tables.  |
|  AWS::S3Vectors::VectorBucket  |  S3 Vector Buckets, including buckets with any indexes.  |
|  AWS::IAM::Group  |  IAM Groups, including groups **with IAM users from outside the stack.** In that case, this tool detaches the IAM users and then deletes the IAM group (but not the IAM users themselves).  |
|  AWS::ECR::Repository  |  ECR Repositories, including repositories that contain images and where **the `EmptyOnDelete` is not true.**  |
|  AWS::Backup::BackupVault  |  Backup Vaults, including vaults **containing recovery points**.  |
|  AWS::Athena::WorkGroup  |  Athena WorkGroups, including workgroups containing **named queries or prepared statements**.  |
|  AWS::Lambda::Function  |  Lambda Functions, including **Lambda@Edge functions with replicas** still being cleaned up by AWS. Waits for AWS to finish removing edge replicas.  |
|  AWS::CloudFormation::Stack  |  **Nested Child Stacks** that failed to delete. If any of the other resources are included in the child stack, **they too will be deleted**.  |
|  AWS::CloudFormation::CustomResource  |  Custom Resources (AWS::CloudFormation::CustomResource), including resources that **do not return a SUCCESS status.**  |
|  Custom::Xxx  |  Custom Resources (Custom::Xxx), including resources that **do not return a SUCCESS status.**  |

> **Note**: If there are resources other than those listed above that result in DELETE_FAILED, the deletion will fail. Resources that are **referenced by stacks outside the deletion targets** are not supported for force deletion. However, if all dependent stacks are included in the deletion targets, they are [deleted in the correct dependency order automatically](#parallel-stack-deletion-with-automatic-dependency-resolution).

## Pre-deletion Processing

This tool automatically performs the following processing **before CloudFormation deletion starts**.

### Deletion Protection Check

Checks for resource-level deletion protection before attempting stack deletion.

**Without `-f` option**: If any resources have deletion protection enabled, the tool reports them and aborts without attempting deletion.

**With `-f` option**: Deletion protection is automatically disabled before deletion proceeds.

|  RESOURCE TYPE  |
| ---- |
|  AWS::EC2::Instance  |
|  AWS::RDS::DBInstance  |
|  AWS::RDS::DBCluster  |
|  AWS::Cognito::UserPool  |
|  AWS::Logs::LogGroup  |
|  AWS::ElasticLoadBalancingV2::LoadBalancer  |

### Performance Optimization

The following resources do not fail during normal deletion, but are optimized in advance to improve deletion speed.

|  RESOURCE TYPE  |  DETAILS  |
| ---- | ---- |
|  AWS::Lambda::Function  |  Automatically detaches VPC configurations from Lambda functions and deletes their ENIs in parallel, **eliminating ENI cleanup wait time**. All Lambda functions within a stack (including nested stacks) are processed in parallel for maximum performance.  |

## Interactive Mode

### Stack Name Selection

When using interactive mode (`-i, --interactive`) without specifying stack names, you can search stack names **case-insensitively** and select stacks.

The filter keyword can be **empty**.

```bash
❯ delstack -i
Filter a keyword of stack names(case-insensitive): goto
```

Then you select stack names in the UI.

```bash
? Select StackNames.
Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed.
  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [x]  dev-goto-04-TestStack
  [ ]  dev-GOTO-03-TestStack
> [x]  dev-Goto-02-TestStack
  [ ]  dev-goto-01-TestStack
```

### Stacks excluded from interactive selection

- **Nested child stacks**: Parent stacks should generally be deleted as a whole. If you need to delete a child stack directly, specify its name with the `-s` option instead of using interactive mode.
- **`XXX_IN_PROGRESS` stacks** (e.g. `ROLLBACK_IN_PROGRESS`): Multiple CloudFormation operations should not run on the same stack simultaneously.
- **TerminationProtection stacks**: Not displayed without `-f`. With `-f`, they appear with a **`*` prefix marker**.

## Force Mode

The `-f, --force` option enables deletion of stacks that would otherwise be blocked:

- **Retain/RetainExceptOnCreate deletion policies**: Resources with these policies are force deleted instead of being retained
- **Resource-level deletion protection** (EC2, RDS, Cognito, etc.): Protection is automatically disabled before deletion
- **Stack TerminationProtection**: After a confirmation prompt, protection is disabled and the stack is deleted

```bash
delstack -f -s dev-goto-01-TestStack
```

### Large Template Handling

When using force mode with stacks that have large templates (exceeding **51,200 bytes**, CloudFormation's direct template size limit), `delstack` automatically:

1. Creates a temporary S3 bucket in your account
2. Uploads the modified template to the bucket
3. Updates the stack via S3 template URL
4. Deletes the temporary S3 bucket after the operation completes

## Parallel Stack Deletion with Automatic Dependency Resolution

When you specify multiple stacks (via the `-s` option or interactive mode `-i`), `delstack` automatically analyzes CloudFormation stack dependencies (via Outputs/Exports and Imports) and deletes stacks in parallel while respecting dependency constraints.

### How It Works

1. **Dependency Analysis**: Analyzes stack dependencies through CloudFormation Exports and Imports
2. **Circular Dependency Detection**: Detects and reports circular dependencies before deletion
3. **External Reference Protection**: Prevents deletion of stacks whose exports are used by non-target stacks
4. **Dynamic Parallel Deletion**: Deletes stacks in dependency order with maximum parallelism

### Deletion Algorithm

The deletion process uses a **reverse topological sort with dynamic scheduling**:

```text
Example: Stacks A, B, C, D, E, F with dependencies:
  C → A (C depends on A)
  D → A
  E → B
  F → C, D, E

Deletion order (reverse dependencies):
  Step 1: Delete F (no stacks depend on it)
  Step 2: Delete C, D, E in parallel (after F completes)
  Step 3: Delete B (after E completes)
  Step 4: Delete A (after both C and D complete)
```

**Key Features**:

- Stacks are deleted **as soon as all dependent stacks are deleted**
- Multiple independent stacks are deleted **in parallel**
- By default, there is **no limit** on the number of concurrent deletions
- The `-n` option allows you to limit the maximum number of concurrent deletions if needed

### Error Handling

- **Circular Dependencies**: Detected before deletion starts, with the dependency cycle path reported
- **External References**: If a target stack's export is imported by a non-target stack, deletion is prevented with a detailed error message
- **Partial Failures**: If any stack fails to delete, all remaining deletions are cancelled

## CDK Integration Details

The `delstack cdk` subcommand synthesizes a CDK app (or reads an existing `cdk.out`) and deletes all discovered stacks. It parses the Cloud Assembly manifest to determine stack names, regions, and dependencies.

![delstack for CDK](https://github.com/user-attachments/assets/0222969c-8c19-4700-80ce-096654ffba74)

### Cross-region deletion

When the CDK app deploys stacks to multiple regions (e.g., `us-east-1` for CloudFront + `ap-northeast-1` for the main app), `delstack cdk` automatically:

1. Detects each stack's region from the Cloud Assembly manifest (`environment` field)
2. Groups stacks by region and creates separate AWS sessions
3. Resolves cross-region dependencies and deletes in the correct order
4. Deletes independent regions in parallel

For environment-agnostic stacks (`unknown-region` in the manifest), the region from `-r` or the default AWS configuration is used.

## GitHub Actions

You can use delstack in GitHub Actions Workflow. To delete multiple stacks, specify stack names separated by commas.

> **Note**: The `yes` option defaults to `true` in GitHub Actions because CI/CD environments cannot handle interactive prompts. Set `yes: false` only if you want the action to abort when a confirmation prompt would appear (e.g., TerminationProtection disable confirmation).

```yaml
jobs:
  delstack:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v3
        with:
          role-to-assume: arn:aws:iam::123456789100:role/my-github-actions-role
          # Or specify access keys.
          # aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          # aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
      - name: Delete stack
        uses: go-to-k/delstack@main # Or specify the version instead of main
        with:
          stack-name: YourStack
          # stack-name: YourStack1, YourStack2, YourStack3 # To delete multiple stacks (deleted in parallel by default)
          force: true # Force Mode to delete stacks including resources with the deletion policy Retain or RetainExceptOnCreate (default: false)
          # yes: true # Skip confirmation prompts (default: true in GitHub Actions)
          concurrency-number: 4 # Number of parallel stack deletions (default: unlimited)
          region: us-east-1
```

### CDK mode

To delete stacks from a CDK app, set `cdk: true`. The action will run `cdk synth` and delete all discovered stacks. You can also pass CDK context values and specify an existing `cdk.out` directory.

```yaml
jobs:
  delstack:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v3
        with:
          role-to-assume: arn:aws:iam::123456789100:role/my-github-actions-role
          aws-region: us-east-1
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: "22"
      - name: Install CDK
        run: npm install -g aws-cdk
      - name: Delete CDK stacks
        uses: go-to-k/delstack@main
        with:
          cdk: true
          # cdk-app: ./cdk.out           # Use existing cdk.out (skip synthesis)
          # cdk-context: env=dev,foo=bar  # CDK context values (comma separated)
          force: true
          # yes: true  # default: true in GitHub Actions
```

### Raw commands

You can also run raw commands after installing the delstack binary.

```yaml
jobs:
  delstack:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v3
        with:
          role-to-assume: arn:aws:iam::123456789100:role/my-github-actions-role
          # Or specify access keys.
          # aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          # aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1
      - name: Install delstack
        uses: go-to-k/delstack@main # Or specify the version instead of main
      - name: Run delstack
        run: |
          echo "delstack"
          delstack -v
          delstack -s YourStack1 -s YourStack2 -r us-east-1
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, adding new resource support, and coding conventions.
