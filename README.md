# delstack

[![Go Report Card](https://goreportcard.com/badge/github.com/go-to-k/delstack)](https://goreportcard.com/report/github.com/go-to-k/delstack) ![GitHub](https://img.shields.io/github/license/go-to-k/delstack) ![GitHub](https://img.shields.io/github/v/release/go-to-k/delstack) [![ci](https://github.com/go-to-k/delstack/actions/workflows/ci.yml/badge.svg)](https://github.com/go-to-k/delstack/actions/workflows/ci.yml)

The description in **Japanese** is available on the following blog page. -> [Blog](https://go-to-k.hatenablog.com/entry/delstack)

The description in **English** is available on the following blog page. -> [Blog](https://dev.to/aws-builders/a-cli-tool-to-force-delete-cloudformation-stacks-3808)

## What is

Tool to force delete the **entire** AWS CloudFormation stack, **even if it contains resources that fail to delete by the CloudFormation delete operation**.

## Resource Types that can be forced to delete

Among the resources that **fail in the normal CloudFormation stack deletion**, this tool supports the following resources.

All resources that do not fail normal deletion can be deleted as is.

|  RESOURCE TYPE  |  DETAILS  |
| ---- | ---- |
|  AWS::S3::Bucket  |  S3 Buckets, including buckets with **Non-empty or Versioning enabled** and DeletionPolicy **not Retain**.(Because "Retain" buckets should not be deleted.)  |
|  AWS::IAM::Role  |  IAM Roles, including roles **with policies from outside the stack**.  |
|  AWS::ECR::Repository  |  ECR Repositories, including repositories **containing images**.  |
|  AWS::Backup::BackupVault  |  Backup Vaults, including vaults **containing recovery points**.  |
|  AWS::CloudFormation::Stack  |  **Nested Child Stacks** that failed to delete. If any of the other resources are included in the child stack, **they too will be deleted**.  |
|  Custom::Xxx  |  Custom Resources, but they will be deleted on its own.  |

- This tool can be used **even for stacks that do not contain any of the above targets** for forced deletion.
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
  # e.g. version 0.12.1
  curl -fsSL https://raw.githubusercontent.com/go-to-k/delstack/main/install.sh | sh -s "v1.2.0"
  delstack -h
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
  delstack [-s <stackName>] [-p <profile>] [-r <region>] [-i]
  ```

- -s, --stackName: optional
  - CloudFormation stack name
    - Must be specified in **not** interactive mode
    - Otherwise you can specify it in the interactive mode
- -p, --profile: optional
  - AWS profile name
- -r, --region: optional(default: `us-east-1`)
  - AWS Region
- -i, --interactive: optional
  - Interactive Mode

## Interactive Mode

### ResourceTypes

The `-i, --interactive` option allows you to select the ResourceTypes **you wish to force delete even if DELETE_FAILED**. This feature allows you to **protect resources** you really do not want to delete by **"do not select the ResourceTypes"**!

However, if a resource can be deleted **without becoming DELETE_FAILED** by the **normal** CloudFormation stack deletion feature, the resource will be deleted **even if you do not select that resource type**. This tool is not intended to protect specific resources from the normal CloudFormation stack deletion feature, so I implemented this feature with the nuance that **only those resources that really should not be deleted will not be forced to be deleted**.

If the stack contains resources that will be DELETE_FAILED but is not selected, **all DELETE_FAILED resources including the selected or not selected resources and the stack will remain undeleted**.

```bash
❯ delstack -s YourStack -i
? Select ResourceTypes you wish to delete even if DELETE_FAILED.
However, if a resource can be deleted without becoming DELETE_FAILED by the normal CloudFormation stack deletion feature, the resource will be deleted even if you do not select that resource type.
  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [ ]  AWS::S3::Bucket
  [x]  AWS::IAM::Role
> [x]  AWS::ECR::Repository
  [ ]  AWS::Backup::BackupVault
  [x]  AWS::CloudFormation::Stack
  [ ]  Custom::
```

### StackName Selection

If you do not specify a stack name in command options in the interactive mode, you can search stack names in a **case-insensitive** and select a stack.

It can be **empty**.

```bash
❯ delstack -i
Filter a keyword of stack names(case-insensitive): test-goto
```

Then you select stack names in the UI.

```bash
? Select StackName.
Nested child stacks and XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks are not displayed.
  [Use arrows to move, type to filter]
> test-goto-stack-1
  test-goto-stack-2
  test-goto-stack-3
  TEST-GOTO-stack-4
  Test-Goto-stack-5
  TEST-goto-stack-6
```

In addition, **child stacks of nested stacks are not displayed**. This is because it is unlikely that there are cases where only child stacks of nested stacks are deleted without deleting the parent stack, and also because it is possible that the parent stack may be buried in the stack list if there are child stacks, or that the child stacks may be accidentally deleted.

However, the `-s` command option allows deletion of child stacks by specifying their names, so please use this option if you want.

Also stacks with **the XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) CloudFormation status** are not displayed, because multiple CloudFormation operations should not be duplicated at the same time.

## GitHub Actions

You can use delstack with parameters **"stack-name" and "region"** in GitHub Actions Workflow.

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
          delstack -s YourStack1 -r us-east-1
          delstack -s YourStack2 -r us-east-1
```
