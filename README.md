# delstack

[![Go Report Card](https://goreportcard.com/badge/github.com/go-to-k/delstack)](https://goreportcard.com/report/github.com/go-to-k/delstack) ![GitHub](https://img.shields.io/github/license/go-to-k/delstack) ![GitHub](https://img.shields.io/github/v/release/go-to-k/delstack)

The description in **Japanese** is available on the following blog page. -> [Blog](https://go-to-k.hatenablog.com/entry/delstack)

## What is

Tool to force delete the **entire** CloudFormation stack, **even if it contains resources that fail to delete by the CloudFormation delete operation**.

## Resource Types that can be forced to delete

Among the resources that **fail in the normal CloudFormation stack deletion**, this tool supports the following resources.

|  RESOURCE TYPE  |  DETAILS  |
| ---- | ---- |
|  AWS::S3::Bucket  |  S3 Buckets, including buckets with **Non-empty or Versioning enabled** and DeletionPolicy **not Retain**.(Because "Retain" buckets should not be deleted.)  |
|  AWS::IAM::Role  |  IAM Roles, including roles **with policies from outside the stack**.  |
|  AWS::ECR::Repository  |  ECR Repositories, including repositories **containing images**.  |
|  AWS::Backup::BackupVault  |  Backup Vaults, including vaults **containing recovery points**.  |
|  AWS::CloudFormation::Stack  |  **Nested Child Stacks** that failed to delete. If any of the other resources are included in the child stack, **they too will be deleted**.  |
|  Custom::Xxx  |  Custom Resources, but they will be deleted on its own.  |

<br>

- This tool can be used **even for stacks that do not contain any of the above targets** for forced deletion.
  - So **all stack deletions can basically be done with this tool!!**
- If there are resources other than those listed above that result in DELETE_FAILED, the deletion will fail.
- **"Termination Protection" stacks will not be deleted.** Because it probably really should not want to delete it.
- Deletion of resources that fail to be deleted because they are used by other stack resources, i.e., **resources that are referenced (depended on) from outside the stack, is not supported**. Only forced deletion of resources that can be completed only within the stack is supported.

## Install

- Homebrew
  ```
  brew tap go-to-k/tap
  brew install go-to-k/tap/delstack
  ```
- Binary
  - [Releases](https://github.com/go-to-k/delstack/releases)
- Git Clone and install(for developers)
  ```
  git clone https://github.com/go-to-k/delstack.git
  cd delstack
  make install
  ```

## How to use
  ```
  delstack -s <stackName> [-p <profile>] [-r <region>] [-i]
  ```

- -s, --stackName: **required**
  - CloudFormation stack name
- -p, --profile: optional
  - AWS profile name
- -r, --region: optional(default: `ap-northeast-1`)
  - AWS Region
- -i, --interactive: optional
  - Interactive Mode

## Interactive Mode

If you selected `-i, --interactive` option, **you can select** ResourceTypes **you wish to delete even if DELETE_FAILED** in a prompt. This allows you to protect resources you do not want to delete!!

However, if resources of the selected ResourceTypes **will not be DELETE_FAILED when the stack is deleted**, the resources will be deleted **even if you selected**. The purpose of this tool is not to protect specific resources from CloudFormation's stack deletion feature, but simply to avoid forcing the deletion of something that really should not be deleted.

If the stack contains resources that will be DELETE_FAILED but is not selected, **all DELETE_FAILED resources including the selected or not selected resources and the stack will remain undeleted**.

```sh
❯ delstack -s YourStack -i
? Select ResourceTypes you wish to delete even if DELETE_FAILED.  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
  [ ]  AWS::S3::Bucket
  [x]  AWS::IAM::Role
> [x]  AWS::ECR::Repository
  [ ]  AWS::Backup::BackupVault
  [x]  AWS::CloudFormation::Stack
  [ ]  Custom::
```
