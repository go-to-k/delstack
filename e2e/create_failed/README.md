# create_failed E2E

Generic E2E for **create-failed phantom** resources (issue #647).

## What it covers

Some resources fail their CloudFormation **CREATE** and then also fail to **DELETE** on rollback,
leaving a *phantom*: a resource that does not actually exist in AWS, but that CloudFormation
cannot cleanly delete, so the stack is stuck in `ROLLBACK_FAILED` with the resource in
`DELETE_FAILED`. This happens for the class of resource types whose delete handlers are not
idempotent: Custom Resources, and Attachment/Association/Settings style types.

delstack must be able to force-delete such stacks (retain the phantom and remove it from the
stack) instead of erroring out. This scenario verifies that.

It is intentionally **generic**: any future create-failed phantom type is considered covered by
this scenario. The concrete vehicle is `AWS::Cognito::UserPoolUICustomizationAttachment`, which
is created **without** a `UserPoolDomain`. `SetUICustomization` requires a domain, so the
attachment's CREATE fails and its rollback DELETE fails too, producing the phantom reliably and
cheaply.

## Run

```sh
# Deploy the (intentionally failing) stack, then force-delete it with delstack:
make e2e_create_failed

# Or just leave the stuck stack in place:
make testgen_create_failed
```

`cdk deploy` is expected to exit non-zero here (the CREATE is designed to fail). `deploy.go`
then verifies the stack is `ROLLBACK_FAILED` with a `DELETE_FAILED` phantom before delstack runs.
