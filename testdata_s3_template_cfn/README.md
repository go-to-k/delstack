# Large CloudFormation Template Test Environment

Test environment for CloudFormation templates larger than 51,200 bytes with `-f` option.

## Setup Test Stack

Generate and deploy a large template (>51KB) with `DeletionPolicy: Retain`:

```bash
make testgen_large_template
# Or with a specific AWS profile
make testgen_large_template OPT="-p my-profile"
```

This will:

1. Generate a large CloudFormation template (>51KB)
2. Create a temporary S3 bucket: `delstack-test-cfn-templates-{account-id}-{region}-{random}`
3. Upload the template to S3
4. Create the CloudFormation stack: `delstack-test-large-template-XXXX` (random suffix)
5. Wait for stack creation to complete
6. **Automatically delete the temporary S3 bucket** (no longer needed)

Note the stack name from the output for testing.

## Test

Use delstack with the `-f` option on the deployed stack:

```bash
make run OPT="-f delstack-test-large-template-XXXX"
# Or with profile
make run OPT="-f delstack-test-large-template-XXXX --profile my-profile"
```

Expected behavior:

1. Detects template size exceeds 51,200 bytes
2. Creates temporary S3 bucket: `delstack-templates-{account}-{region}-{timestamp}`
3. Uploads modified template to S3
4. Updates stack using `UpdateStackWithTemplateURL`
5. Automatically deletes the temporary S3 bucket and template
