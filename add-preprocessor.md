Add a new preprocessor to run before CloudFormation stack deletion.

Preprocessor name: $ARGUMENTS

First determine which phase this preprocessor belongs to:
- **Checker**: Errors are fatal and abort deletion. Used for validation that must pass before proceeding (e.g., deletion protection check). All checkers run in parallel and all errors are collected before returning.
- **Modifier**: Errors are logged as warnings but do not stop deletion. Used for optimizations (e.g., Lambda VPC detachment).

Follow these steps:

1. **`internal/preprocessor/<name>.go`**: Implement the `IPreprocessor` interface.

2. **`internal/preprocessor/<name>_test.go`**: Write tests for the preprocessor.

3. **`internal/preprocessor/factory.go`**:
   - Add a `new<Name>FromConfig()` factory function.
   - Append the new preprocessor to the `checkers` or `modifiers` slice in `NewRecursivePreprocessorFromConfig()`.

4. **`README.md`**: Add a row to the "Pre-deletion Processing" table.
