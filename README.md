# delstack

[![Go Report Card](https://goreportcard.com/badge/github.com/go-to-k/delstack)](https://goreportcard.com/report/github.com/go-to-k/delstack) ![GitHub](https://img.shields.io/github/license/go-to-k/delstack) ![GitHub](https://img.shields.io/github/v/release/go-to-k/delstack) [![ci](https://github.com/go-to-k/delstack/actions/workflows/ci.yml/badge.svg)](https://github.com/go-to-k/delstack/actions/workflows/ci.yml)

The description in **Japanese** is available on the following blog page. -> [Blog](https://go-to-k.hatenablog.com/entry/delstack)

The description in **English** is available on the following blog page. -> [Blog](https://dev.to/aws-builders/a-cli-tool-to-force-delete-cloudformation-stacks-3808)

## What is

Tool to force delete the **entire** AWS CloudFormation stack, even if it contains resources that **fail to delete** by the CloudFormation delete operation.

**Works with stacks created by any tool**: Not just raw CloudFormation, but also stacks deployed via **AWS CDK**, **AWS SAM**, **Serverless Framework**, and other Infrastructure as Code tools that use CloudFormation under the hood.

![delstack](https://github.com/user-attachments/assets/5be877d5-5482-44f9-96aa-defc43fffb5b)

## Resource Types that can be forced to delete

Among the resources that **fail in the normal CloudFormation stack deletion**, this tool supports the following resources.

If you want to delete the unsupported resources, please create an issue at [GitHub](https://github.com/go-to-k/delstack/issues).

All resources that do not fail normal deletion can be deleted as is.

|  RESOURCE TYPE  |  DETAILS  |
| ---- | ---- |
|  AWS::S3::Bucket  |  S3 Buckets, including buckets with **Non-empty or Versioning enabled**.  |
|  AWS::S3Express::DirectoryBucket  |  S3 Directory Buckets for S3 Express One Zone, including buckets with Non-empty.  |
|  AWS::S3Tables::TableBucket  |  S3 Table Buckets, including buckets with any namespaces or tables.  |
|  AWS::S3Tables::Namespace  |  S3 Table Namespaces, including namespaces with any tables.  |
|  AWS::S3Vectors::VectorBucket  |  S3 Vector Buckets, including buckets with any indexes.  |
|  AWS::IAM::Group  |  IAM Groups, including groups **with IAM users from outside the stack.** In that case, this tool detaches the IAM users and then deletes the IAM group (but not the IAM users themselves).  |
|  AWS::ECR::Repository  |  ECR Repositories, including repositories that contain images and where **the `EmptyOnDelete` is not true.**  |
|  AWS::Backup::BackupVault  |  Backup Vaults, including vaults **containing recovery points**.  |
|  AWS::CloudFormation::Stack  |  **Nested Child Stacks** that failed to delete. If any of the other resources are included in the child stack, **they too will be deleted**.  |
|  Custom::Xxx  |  Custom Resources, including resources that **do not return a SUCCESS status.**  |

- This tool can be used even for stacks that do not contain any of the above targets for forced deletion.
  - So **all stack deletions can basically be done with this tool!!**
- If there are resources other than those listed above that result in DELETE_FAILED, the deletion will fail.
- **"Termination Protection" stacks will not be deleted.** Because it probably really should not want to delete it.
- Deletion of resources that fail to be deleted because they are used by other stack resources, i.e., **resources that are referenced (depended on) from outside the stack, is not supported**. Only forced deletion of resources that can be completed only within the stack is supported.

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
- Git Clone and install(for developers)

  ```bash
  git clone https://github.com/go-to-k/delstack.git
  cd delstack
  make install
  ```

## How to use

  ```bash
  delstack [-s <stackName>] [-p <profile>] [-r <region>] [-i|--interactive] [-f|--force] [-n <concurrencyNumber>]
  ```

- -s, --stackName: optional
  - CloudFormation stack name
    - Must be specified in **not** interactive mode
    - Otherwise you can specify it in the interactive mode
  - **Multiple specifications are possible.**
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
  - Force Mode to delete stacks including resources with **the deletion policy `Retain` or `RetainExceptOnCreate`**
- -n, --concurrencyNumber: optional(default: unlimited)
  - Specify the number of parallel stack deletions. Default is unlimited (delete all stacks in parallel).

## Interactive Mode

### StackName Selection

If you do not specify a stack name in command options in the interactive mode (`-i, --interactive`), you can search stack names in a **case-insensitive** and select a stack.

It can be **empty**.

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

In addition, **child stacks of nested stacks are not displayed**. This is because it is unlikely that there are cases where only child stacks of nested stacks are deleted without deleting the parent stack, and also because it is possible that the parent stack may be buried in the stack list if there are child stacks, or that the child stacks may be accidentally deleted.

However, **the `-s` command option allows deletion of CHILD stacks by specifying their names**, so please use this option if you want.

And stacks with **the XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) CloudFormation status** are not displayed, because multiple CloudFormation operations should not be duplicated at the same time.

Also, **"Termination Protection"** stacks will not be displayed, because it probably really should not want to delete it.

## Force Mode

If you specify the `-f, --force` option, stacks including resources with **the deletion policy `Retain` or `RetainExceptOnCreate`** will be deleted.

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

## GitHub Actions

You can use delstack with parameters **"stack-name", "region", and "concurrency-number"** in GitHub Actions Workflow.
To delete multiple stacks, specify stack names separated by commas.

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
          concurrency-number: 4 # Number of parallel stack deletions (default: unlimited)
          region: us-east-1
```

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
