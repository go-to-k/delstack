# delstack

The description in **Japanese** is as follows -> [Blog](https://go-to-k.hatenablog.com/entry/delstack)

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
  delstack -s <stackName> [-p <profile>] [-r <region>]
  ```

- -s, --stackName: **required**
  - CloudFormation stack name
- -p, --profile: optional
  - AWS profile name
- -r, --region: optional(default: `ap-northeast-1`)
  - AWS Region
