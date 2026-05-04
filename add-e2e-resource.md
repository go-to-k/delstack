Update E2E tests in `e2e/full/` when adding a new target resource type (Operator).

Resource type: $ARGUMENTS

**Important**: The goal is to reproduce the `DELETE_FAILED` state. Dependency resources that block deletion must be created **outside of CloudFormation** (via SDK calls in `deploy.go`), not inside the CDK stack — otherwise CloudFormation resolves dependency order and deletes successfully. CDK creates only the base resource; `deploy.go` attaches blocking dependencies via SDK after deployment.

If a blocking dependency cannot be reproduced outside CloudFormation (e.g., MFA devices, FIDO/Passkey), skip E2E and cover with unit tests only.

Follow these steps:

1. **`e2e/full/cdk/lib/resource/<resource>.go`**: Create a CDK constructor `New<Resource>(scope)` that creates the base resource in a non-deletable state (e.g., ECR with `EmptyOnDelete: false`).

2. **`e2e/full/cdk/cdk.go`**: Call `resource.New<Resource>(stack, ...)` in `NewTestStack()`.

3. **`e2e/full/deploy.go`**: After CDK deployment, use the AWS SDK to populate data or attach dependencies that block deletion (e.g., upload objects to S3, push images to ECR).

4. **`e2e/full/go.mod`**: Add required AWS SDK service dependencies.

5. **`e2e/full/cdk/go.mod`**: Add CDK construct dependencies if applicable.

6. **`e2e/full/README.md`**: Add the new resource to the resource list.
