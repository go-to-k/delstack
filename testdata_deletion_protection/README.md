# Delstack Test Environment - Deletion Protection

This directory contains a tool for creating a test environment specifically for `delstack`'s deletion protection handling feature. This tool (`deploy.go`) deploys AWS CloudFormation stacks with resources that have deletion protection enabled, to test the automatic disabling and deletion functionality.

## Test Stack Deployment

You can deploy test CloudFormation stacks using the included `deploy.go` script **with AWS CDK for Go**. So you need to install AWS CDK.

```bash
npm install -g aws-cdk@latest
```

This script creates a CloudFormation stack (with **TerminationProtection enabled** by default) containing:

- EC2 Instance with API termination protection enabled (`DisableApiTermination: true`)
- RDS DBInstance with deletion protection enabled (`DeletionProtection: true`)
- RDS Aurora DBCluster (MySQL) with deletion protection enabled (`DeletionProtection: true`)
- Cognito UserPool with deletion protection active (`DeletionProtection: ACTIVE`)
- ELBv2 Application Load Balancer with deletion protection enabled
- CloudWatch LogGroup with deletion protection enabled
- VPC with public subnet (shared by EC2, RDS, and ELBv2)

```bash
go run testdata_deletion_protection/deploy.go -s <stage> [-p <profile>] [-t]
```

### Options

- `-s <stage>` : Stage name, used as part of stack naming (optional, auto-generated if not specified)
- `-p <profile>` : AWS CLI profile name to use (optional)
- `-t` (`--no-tp`) : Disable stack TerminationProtection (default: enabled). Use this to test resource-level deletion protection only.

### Using the Makefile

For convenience, you can also use the Makefile target:

```bash
# Deploy with auto-generated stage name
make testgen_deletion_protection

# Deploy with custom stage and profile
make testgen_deletion_protection OPT="-s my-stage -p my-profile"

# Deploy without stack TerminationProtection (resource-level protection only)
make testgen_deletion_protection OPT="-t"
```

### Testing Deletion Protection

After deployment, you can test the deletion protection handling by running:

```bash
# Without -f flag: should show TerminationProtection error
delstack -s <stack-name>

# With -f flag: should disable TerminationProtection and resource deletion protection, then delete the stack
delstack -s <stack-name> -f

# Interactive mode without -f: TP-enabled stacks are hidden from the list
delstack -i

# Interactive mode with -f: TP-enabled stacks are shown with "* " prefix marker
delstack -i -f
```

The `-f` (force) flag will automatically:
1. Disable stack TerminationProtection (with confirmation prompt)
2. Detect resources with deletion protection enabled
3. Disable deletion protection on each resource
4. Proceed with stack deletion

### Notes

- This test environment focuses on deletion protection configurations only
- The stack includes resources from multiple AWS services to test comprehensive protection handling
- VPC uses a single public subnet with 1 AZ to minimize costs
- RDS DBInstance uses the smallest available instance type (db.t3.micro) to minimize costs
- RDS Aurora DBCluster uses db.t3.medium (smallest Aurora-compatible instance)
- EC2 uses t3.micro to minimize costs
