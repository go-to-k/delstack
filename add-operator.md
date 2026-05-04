Add support for a new AWS resource type (Operator) that causes `DELETE_FAILED` during CloudFormation stack deletion.

Resource type to add: $ARGUMENTS

Follow these steps in order:

1. **`internal/resourcetype/resourcetype.go`**: Add a new resource type constant to the `const` block and append it to the `ResourceTypes` slice.

2. **`pkg/client/<service>.go`**: Create or extend the AWS SDK client wrapper.
   - If no client file exists for this AWS service yet: create a new file with a `//go:generate mockgen` comment on **line 1**, an `I<Service>` interface, and a concrete implementation struct.
   - If a client file for the same service already exists (e.g., adding `AWS::IAM::User` when `iam.go` already has `AWS::IAM::Group`): add the new methods to the existing interface and struct instead of creating a new file.

3. **`pkg/client/<service>_mock.go`**: Run `make mockgen` to regenerate this file automatically.

4. **`pkg/client/<service>_test.go`**: Write client tests using AWS SDK middleware mocks (inject a `middleware.FinalizeMiddlewareFunc` via `config.WithAPIOptions`). For paginated APIs, use an Initialize-stage middleware to capture `NextToken` via `middleware.WithStackValue` / `GetStackValue`. See `pkg/client/backup_test.go` or `pkg/client/s3_test.go` for examples.

5. **`internal/operation/<resource_type>.go`**: Implement the `IOperator` interface. Add `var _ IOperator = (*XxxOperator)(nil)` for a compile-time type check.

6. **`internal/operation/<resource_type>_test.go`**: Write operator tests using gomock with the generated mock interfaces from `pkg/client/`.

7. **`internal/operation/operator_factory.go`**: Add a `Create<Resource>Operator()` method that initializes the SDK client.

8. **`internal/operation/operator_collection.go`**: Update 4 places:
   - Instantiate the operator in `SetOperatorCollection()`
   - Add a `case` branch in the switch statement
   - Append to the `operators` slice
   - Add a row to `supportedStackResourcesData` in `RaiseUnsupportedResourceError()`

9. **`internal/operation/operator_collection_test.go`**: Update test cases to include the new resource type.

10. **`go.mod` / `go.sum`**: Add the AWS SDK service dependency:
    ```
    go get github.com/aws/aws-sdk-go-v2/service/<service>
    go mod tidy
    ```

11. **`README.md`**: Add a row to the "Resource Types that can be forced to delete" table.

12. **E2E tests**: Update `e2e/full/` following the `/project:add-e2e-resource` skill.

For a reference implementation, see [PR #569 (Athena WorkGroup)](https://github.com/go-to-k/delstack/pull/569).
