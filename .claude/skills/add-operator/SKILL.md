---
name: add-operator
description: Add a new Operator (force-delete support for an AWS resource type that causes DELETE_FAILED, e.g. AWS::Athena::WorkGroup, AWS::IAM::User). Use when adding a new entry to the "Resource Types that can be forced to delete" table. Not for Preprocessors (deletion-protection checks, Lambda VPC detach, etc.) — use `add-preprocessor` for those.
---

# Add a new Operator

You are adding force-delete support for a CloudFormation resource type that causes `DELETE_FAILED`. The Operator force-deletes the blocking sub-resources (e.g. emptying an S3 bucket, deleting ECR images, detaching IAM user dependencies) so CloudFormation can finish the stack delete.

Reference PR for the full diff shape: [PR #569 (Athena WorkGroup)](https://github.com/go-to-k/delstack/pull/569).

## Input

The user provides an AWS resource type, e.g. `AWS::Athena::WorkGroup`. Derive:

- Service name (lowercase, e.g. `athena`)
- Resource name (e.g. `WorkGroup`)
- AWS SDK v2 package: `github.com/aws/aws-sdk-go-v2/service/<service>`
- Client filename: `pkg/client/<service>.go` — **if it already exists, extend the existing interface and struct rather than creating a new file** (e.g. when adding `AWS::IAM::User` and `iam.go` is already there)
- Operator filename: `internal/operation/<resource_type_snake>.go` (e.g. `athena_workgroup.go`)
- Resource type constant name: PascalCase (e.g. `AthenaWorkGroup`)

## Steps

1. **`internal/resourcetype/resourcetype.go`** — Add the constant to the `const` block and append it to the `ResourceTypes` slice. Keep it in the existing grouping (most go in the "For Force Deletion" group).

2. **`pkg/client/<service>.go`** — AWS SDK client wrapper.
   - **Line 1 must be** the `//go:generate mockgen` directive (mockgen requires this exact placement).
   - Define `I<Service>` interface and a `<Service>` struct that wraps the SDK client.
   - Public methods must wrap errors with `ClientError` (see how an existing client like `pkg/client/ecr.go` or `pkg/client/athena.go` does it).
   - If the file already exists, just add new methods to the existing interface + struct.
   - Methods can be either thin SDK 1:1 wrappers OR higher-level helpers that combine multiple SDK calls (and may add parallelism with `errgroup` + `semaphore.NewWeighted(runtime.NumCPU())`) when the abstraction belongs at the client layer. Example: `pkg/client/ec2.go` `DeleteOrphanLambdaENIsByFilter` discovers ENIs via `DescribeNetworkInterfaces` then deletes them in parallel via `DeleteNetworkInterface`.

3. **`pkg/client/<service>_mock.go`** — Auto-generated. Do not edit by hand. Run `make mockgen` after step 2.

4. **`pkg/client/<service>_test.go`** — Tests use AWS SDK middleware mocks (NOT gomock). Inject a `middleware.FinalizeMiddlewareFunc` via `config.WithAPIOptions`. For paginated APIs, capture `NextToken` in an Initialize-stage middleware via `middleware.WithStackValue` and read it back via `middleware.GetStackValue`. See `pkg/client/s3_test.go` and `pkg/client/backup_test.go`.

5. **`internal/operation/<resource>.go`** — The Operator.
   - Implement `IOperator`. Add a compile-time assertion: `var _ IOperator = (*<Resource>Operator)(nil)`.
   - Operations must be **idempotent** (check existence before deleting, or rely on a NotFound-tolerant client method).
   - Concurrency uses `errgroup` + `semaphore.NewWeighted(runtime.NumCPU())`.
   - Errors: return client errors as-is (already wrapped). Only wrap with `ClientError` for errors generated locally in the operation layer (e.g. `ctx.Done()`, validation).
   - **Deletion model**: the operator deletes the AWS resource itself directly via the SDK (and any blocking dependencies it owns). The CFN deletion loop in `internal/operation/cloudformation_stack.go` then passes every DELETE_FAILED `LogicalResourceId` as `RetainResources` on the next `DeleteStack` call, so CFN skips that resource on retry. Do **not** rely on CloudFormation to redelete the resource for you.
   - **Cross-resource cleanup (advanced)**: an operator may also remove out-of-stack dependencies that block deletion (example: `EC2SubnetOperator` / `EC2SecurityGroupOperator` first remove orphan AWS Lambda VPC ENIs that AWS Lambda has not yet released, then delete the Subnet / SecurityGroup). When you do this, scope the dependency filter very tightly (e.g. `status=available` AND a description prefix) so you cannot accidentally touch unrelated resources.

6. **`internal/operation/<resource>_test.go`** — Tests use **gomock** (not middleware). Use the `I<Service>` mock generated in step 3.

7. **`internal/operation/operator_factory.go`** — Add `Create<Resource>Operator()` returning the new operator with its SDK client(s) initialised (use `SDKRetryMaxAttempts` and `aws.RetryModeStandard` like the existing factories).

8. **`internal/operation/operator_collection.go`** — Update **4 sites**:
   1. Instantiate the operator in `SetOperatorCollection()` (`<resource>Operator := c.operatorFactory.Create<Resource>Operator()`).
   2. Add a `case resourcetype.<Resource>:` branch inside the switch over `*resource.ResourceType` that calls `<resource>Operator.AddResource(&resource)`.
   3. Append the operator to the `c.operators` slice at the bottom of `SetOperatorCollection()`.
   4. Add a row to `supportedStackResourcesData` in `RaiseUnsupportedResourceError()` describing what blocks deletion for this resource.

9. **`internal/operation/operator_collection_test.go`** — Update test cases to cover the new resource type.

10. **`go.mod` / `go.sum`** — `go get github.com/aws/aws-sdk-go-v2/service/<service>`, then `go mod tidy` so the dep ends up classified correctly (direct vs indirect).

11. **`README.md`** — Add a row to the "Resource Types that can be forced to delete" table (see the existing rows for tone and format).

12. **E2E test** — Two layouts are valid:

    **(a) Extend `e2e/full/`** — preferred when the resource fits the existing combined test stack:
    1. `e2e/full/cdk/lib/resource/<resource>.go` — `New<Resource>(scope, ...)` constructor that creates the resource in a state that itself does **not** block deletion (e.g. IAM User with no attachments).
    2. `e2e/full/cdk/cdk.go` — Call the constructor in `NewTestStack()`.
    3. `e2e/full/deploy.go` — After CDK deployment, attach the dependencies that **do** block deletion via SDK calls (e.g. upload S3 objects, push ECR images, attach IAM user policies). This is essential: if dependencies are inside the same stack, CloudFormation resolves the dependency order and the stack would delete cleanly, defeating the purpose of the test.
    4. `e2e/full/go.mod` — Add SDK service deps used by `deploy.go`.
    5. `e2e/full/cdk/go.mod` — Add CDK construct deps if used.
    6. `e2e/full/README.md` — Add the resource to the list.

    **(b) Dedicated `e2e/<scenario>/` directory** — preferred when the failure mode requires special topology (VPC, Lambda@Edge) or out-of-band setup that would interfere with `e2e/full/`. Mirror an existing scenario such as `e2e/preprocessor/` or `e2e/vpc_lambda/`. Required files:
    - `cdk/cdk.go`, `cdk/cdk.json`, `cdk/lib/resource/<thing>.go`, `cdk/go.mod`, `cdk/go.sum`
    - `cdk/.gitignore` (excluding at minimum `cdk.context.json` and `cdk.out`)
    - `deploy.go`, `go.mod`, `go.sum`, `README.md`
    - `.gitignore` at the e2e module root with `/<basename>` to ignore the binary `go build ./...` produces from `deploy.go` (the binary is named after the module's basename, e.g. `e2e/vpc_lambda/vpc_lambda`).
    - Add `testgen_<scenario>` and `e2e_<scenario>` targets to the root `Makefile`, plus matching lines under the `testgen_help` / `e2e_help` listings.

    If a dependency cannot feasibly be reproduced outside CloudFormation (e.g. MFA TOTP, FIDO/Passkey hardware), skip that dimension in E2E and rely on unit tests instead. Note this in the PR description.

    The maintainer runs E2E tests; you do not need to deploy them yourself, but the code must compile and the CDK must synth.

13. **Verify locally**:
    - `make mockgen` (regenerates the mock from step 3)
    - `make test`
    - `make lint`

## Conventions reminder

- Code comments: minimal, in English.
- Test naming: `Test[ReceiverType]_[MethodName]` (e.g. `TestEcr_DeleteRepository`, `TestS3BucketOperator_DeleteS3Bucket`).
- All public-facing text (README rows, error messages, PR/commit) is English.
- Do not introduce a `pkg/client/<service>.go` if one already exists for that AWS service — extend it.
