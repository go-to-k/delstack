# Delstack Test Environment - S3 Files

This directory reproduces the real user-facing scenario of issue #634 and verifies that `delstack` can recover from it.

## Why the deletion fails

When an S3 Files file system is deleted, S3 Files first checks that every change made through the file system has been exported to the linked S3 bucket. If anything is still pending, the delete is rejected:

> If there are changes that have not yet been synchronized, S3 Files returns an error and the deletion does not proceed.
> — [Deleting file systems](https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-files-file-systems-deleting.html)

CloudFormation does not pass the force delete option, so the stack ends in `DELETE_FAILED`:

```
File system 'fs-...' has data pending export to S3. Use forceDelete=true to force delete without exporting pending data. (Service: S3Files, Status Code: 409, ...)
```

The export lags writes by about a minute, so this happens whenever the stack is deleted shortly after something wrote to the file system. The `S3FilesFileSystemOperator` deletes the file system with `ForceDelete=true`. The data that was not exported yet is lost.

## What `deploy.go` does

1. `cdk deploy` a **client stack** (`<stage>-client`): a VPC with a public subnet and an EC2 instance with the S3 Files mount helper installed.
2. `cdk deploy` the **target stack** (`<stage>`): a versioned bucket, an S3 Files file system on it, and a mount target in the client stack's subnet. The VPC/Subnet IDs are passed with `-c` context, not a CloudFormation Export, because an Export imported by a non-target stack makes `delstack` refuse the deletion.
3. Through SSM, mount the file system on the instance and start a background loop that keeps writing files, so there is always data that has not been exported yet.

The writer has to live **outside** the target stack. Inside it, CloudFormation would terminate the instance before deleting the mount target and the file system, and the pending data would be exported before the file system delete, so the stack would delete cleanly and the run would test nothing.

`delstack` itself drives `DeleteStack`. CloudFormation deletes the mount target and then fails on the file system, the `S3FilesFileSystemOperator` force-deletes it, and the next passes remove the rest (the bucket is emptied by the existing `S3BucketOperator`).

## Verifying the scenario and cleaning up

Run the script again with `-d` after `delstack` finishes:

```bash
go run e2e/s3files/deploy.go -s <stage> -d
```

It checks the events of the deleted target stack for the `FileSystem` `DELETE_FAILED` with `has data pending export to S3`, then deletes the client stack. It exits non-zero if that event is missing, since the run then passed without exercising the operator.

`make e2e_s3files` does this automatically after `delstack` finishes. If `delstack` itself fails, make stops before that step and the client instance keeps running, so run the command above by hand.

## Test Stack Deployment

You need AWS CDK installed:

```bash
npm install -g aws-cdk@latest
```

```bash
go run e2e/s3files/deploy.go -s <stage> [-p <profile>]
```

Delete the target stack right after the script finishes, while the writes are still running.

### Options

- `-s <stage>` : Stage name, used as the target stack name (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)
- `-d` : Verify the deleted target stack and delete the client stack instead of deploying

### Using the Makefile

```bash
# Deploy both stacks and start writing with an auto-generated stage name
make testgen_s3files

# Deploy, delete with delstack while writing, verify, then remove the client stack
make e2e_s3files

# Custom stage / profile
make e2e_s3files STAGE=my-stage OPT="-p my-profile"
```
