# Changelog

## [v1.16.1](https://github.com/go-to-k/delstack/compare/v1.16.0...v1.16.1) - 2025-05-20
- test: add option for testgen_retain command by @go-to-k in https://github.com/go-to-k/delstack/pull/489
- docs: force mode in README by @go-to-k in https://github.com/go-to-k/delstack/pull/491
- chore: remove deletion policy for nested stacks in parallel by @go-to-k in https://github.com/go-to-k/delstack/pull/490

## [v1.16.0](https://github.com/go-to-k/delstack/compare/v1.15.0...v1.16.0) - 2025-05-19
- feat: support force (`-f`) option for deletion of Retain resources by @go-to-k in https://github.com/go-to-k/delstack/pull/486

## [v1.15.0](https://github.com/go-to-k/delstack/compare/v1.14.0...v1.15.0) - 2025-04-14
- ci: change PR label names for release by @go-to-k in https://github.com/go-to-k/delstack/pull/456
- docs: improve style for README by @go-to-k in https://github.com/go-to-k/delstack/pull/455
- docs: improve description for AWS::IAM::Group in README by @go-to-k in https://github.com/go-to-k/delstack/pull/473
- docs: add gif in README by @go-to-k in https://github.com/go-to-k/delstack/pull/476
- chore: migrate golangci to v2 by @go-to-k in https://github.com/go-to-k/delstack/pull/478
- fix(client): add condition for retry by @go-to-k in https://github.com/go-to-k/delstack/pull/480
- test: create deploy script for test with golang instead of shell by @go-to-k in https://github.com/go-to-k/delstack/pull/479
- test: use AWS CDK for Go in test scripts by @go-to-k in https://github.com/go-to-k/delstack/pull/482
- feat: support S3 Tables by @go-to-k in https://github.com/go-to-k/delstack/pull/483

## [v1.14.0](https://github.com/go-to-k/delstack/compare/v1.13.3...v1.14.0) - 2024-08-26
- ci: tweak for pr-lint by @go-to-k in https://github.com/go-to-k/delstack/pull/387
- ci: Manage labels in PR lint by @go-to-k in https://github.com/go-to-k/delstack/pull/389
- ci: tweak for semantic-pull-request workflow by @go-to-k in https://github.com/go-to-k/delstack/pull/390
- ci: fix bug that labels are not created by @go-to-k in https://github.com/go-to-k/delstack/pull/391
- ci: ignore lint on tagpr PR by @go-to-k in https://github.com/go-to-k/delstack/pull/392
- ci: add revert type in prlint by @go-to-k in https://github.com/go-to-k/delstack/pull/394
- ci: change token for tagpr by @go-to-k in https://github.com/go-to-k/delstack/pull/397
- ci: don't run CI in PR actions by @go-to-k in https://github.com/go-to-k/delstack/pull/398
- ci: add error linters by @go-to-k in https://github.com/go-to-k/delstack/pull/395
- ci: change token for tagpr by @go-to-k in https://github.com/go-to-k/delstack/pull/400
- feat(io): redesign UI implementation with a new library by @go-to-k in https://github.com/go-to-k/delstack/pull/393

## [v1.13.3](https://github.com/go-to-k/delstack/compare/v1.13.2...v1.13.3) - 2024-08-16
- ci(deps): upgrade to goreleaser-action@v6 by @go-to-k in https://github.com/go-to-k/delstack/pull/384
- ci: PR-Lint for PR titles by @go-to-k in https://github.com/go-to-k/delstack/pull/386

## [v1.13.1](https://github.com/go-to-k/delstack/compare/v1.13.0...v1.13.1) - 2024-08-16
- chore: use math/rand/v2 by @go-to-k in https://github.com/go-to-k/delstack/pull/377
- chore: use new gomock by @go-to-k in https://github.com/go-to-k/delstack/pull/378
- ci: add linter by @go-to-k in https://github.com/go-to-k/delstack/pull/379
- ci: use tagpr by @go-to-k in https://github.com/go-to-k/delstack/pull/380

## [v1.13.0](https://github.com/go-to-k/delstack/compare/v1.12.0...v1.13.0) - 2024-08-15
- feat(operation): unsupport IAM role because it can be deleted by normal deletion now by @go-to-k in https://github.com/go-to-k/delstack/pull/369
- chore(deps): bump github.com/urfave/cli/v2 from 2.25.0 to 2.27.4 by @dependabot in https://github.com/go-to-k/delstack/pull/367
- test: improve testdata by @go-to-k in https://github.com/go-to-k/delstack/pull/372
- fix: BackVault deletion fails by @go-to-k in https://github.com/go-to-k/delstack/pull/371
- fix: DeleteBucket error on Directory Buckets by @go-to-k in https://github.com/go-to-k/delstack/pull/374

## [v1.12.0](https://github.com/go-to-k/delstack/compare/v1.11.0...v1.12.0) - 2024-08-13
- chore(deps): bump goreleaser/goreleaser-action from 5 to 6 by @dependabot in https://github.com/go-to-k/delstack/pull/348
- chore(deps): bump github.com/rs/zerolog from 1.32.0 to 1.33.0 by @dependabot in https://github.com/go-to-k/delstack/pull/359
- chore(deps): bump github.com/aws/aws-sdk-go-v2/service/ecr from 1.31.0 to 1.32.0 by @dependabot in https://github.com/go-to-k/delstack/pull/364
- chore(deps): bump golang.org/x/sync from 0.5.0 to 0.8.0 by @dependabot in https://github.com/go-to-k/delstack/pull/365
- feat(operator): support IAM Groups by @go-to-k in https://github.com/go-to-k/delstack/pull/368

## [v1.11.0](https://github.com/go-to-k/delstack/compare/v1.10.0...v1.11.0) - 2024-08-07
- chore(client): remove ListObjectVersions method by @go-to-k in https://github.com/go-to-k/delstack/pull/360
- test(client): remove unused functions by @go-to-k in https://github.com/go-to-k/delstack/pull/361
- test: fix region for ecr in deploy.sh by @go-to-k in https://github.com/go-to-k/delstack/pull/362
- refactor(client): change unused methods from operator to private by @go-to-k in https://github.com/go-to-k/delstack/pull/366
- feat(operator): support Directory Buckets for S3 Express One Zone by @go-to-k in https://github.com/go-to-k/delstack/pull/363

## [v1.10.0](https://github.com/go-to-k/delstack/compare/v1.9.0...v1.10.0) - 2024-08-01
- chore(deps): update AWS SDK versions by @go-to-k in https://github.com/go-to-k/delstack/pull/355
- feat(client): retry on APIs other than DeleteObjects in S3 client by @go-to-k in https://github.com/go-to-k/delstack/pull/357
- refactor(client): retryer in IAM client by @go-to-k in https://github.com/go-to-k/delstack/pull/356

## [v1.9.0](https://github.com/go-to-k/delstack/compare/v1.8.0...v1.9.0) - 2024-07-21
- chore: change config of brews in .goreleaser.yaml by @go-to-k in https://github.com/go-to-k/delstack/pull/346
- chore(deps): bump github.com/aws/aws-sdk-go-v2/service/cloudformation from 1.39.1 to 1.47.0 by @dependabot in https://github.com/go-to-k/delstack/pull/337
- chore(deps): bump github.com/aws/aws-sdk-go-v2/service/iam from 1.27.2 to 1.31.1 by @dependabot in https://github.com/go-to-k/delstack/pull/336
- chore(deps): bump actions/cache from 3 to 4 by @dependabot in https://github.com/go-to-k/delstack/pull/315
- chore(deps): bump actions/upload-artifact from 3 to 4 by @dependabot in https://github.com/go-to-k/delstack/pull/303
- chore(deps): bump go.uber.org/goleak from 1.2.1 to 1.3.0 by @dependabot in https://github.com/go-to-k/delstack/pull/264
- chore(deps): bump github.com/rs/zerolog from 1.30.0 to 1.32.0 by @dependabot in https://github.com/go-to-k/delstack/pull/317
- feat(client): improve deletion logic for S3 by @go-to-k in https://github.com/go-to-k/delstack/pull/347

## [v1.8.0](https://github.com/go-to-k/delstack/compare/v1.7.1...v1.8.0) - 2024-03-27
- feat(operation): check all stacks for TerminationProtection before starting deletion when multiple stacks by @go-to-k in https://github.com/go-to-k/delstack/pull/345

## [v1.7.1](https://github.com/go-to-k/delstack/compare/v1.7.0...v1.7.1) - 2024-03-25
- chore(action): tweak for handling stack names with spaces by @go-to-k in https://github.com/go-to-k/delstack/pull/343

## [v1.7.0](https://github.com/go-to-k/delstack/compare/v1.6.0...v1.7.0) - 2024-03-24
- feat(action): can specify multiple stacks in GitHub Actions by @go-to-k in https://github.com/go-to-k/delstack/pull/342

## [v1.6.0](https://github.com/go-to-k/delstack/compare/v1.5.0...v1.6.0) - 2024-03-23
- chore: add PR template by @go-to-k in https://github.com/go-to-k/delstack/pull/309
- docs: aqua install in README by @go-to-k in https://github.com/go-to-k/delstack/pull/314
- chore(operation): add an issue link in the unsupported error message by @go-to-k in https://github.com/go-to-k/delstack/pull/338
- feat(app): multiple stacks deletion by @go-to-k in https://github.com/go-to-k/delstack/pull/341

## [v1.5.0](https://github.com/go-to-k/delstack/compare/v1.4.2...v1.5.0) - 2023-12-22
- feat(io): keep filter for resource types selection active in interactive mode by @go-to-k in https://github.com/go-to-k/delstack/pull/308

## [v1.4.2](https://github.com/go-to-k/delstack/compare/v1.4.1...v1.4.2) - 2023-12-22
- fix(io): trim spaces from stack name selection filter in interactive mode by @go-to-k in https://github.com/go-to-k/delstack/pull/307

## [v1.4.1](https://github.com/go-to-k/delstack/compare/v1.4.0...v1.4.1) - 2023-12-08
- chore(app): change message for stack name selection when interactive mode by @go-to-k in https://github.com/go-to-k/delstack/pull/299

## [v1.4.0](https://github.com/go-to-k/delstack/compare/v1.3.0...v1.4.0) - 2023-12-08
- feat(operation): do not display stacks with delete protection in the interactive mode by @go-to-k in https://github.com/go-to-k/delstack/pull/296
- feat(operation): throws error if a stack with XxxInProgress is specified by @go-to-k in https://github.com/go-to-k/delstack/pull/298

## [v1.3.0](https://github.com/go-to-k/delstack/compare/v1.2.0...v1.3.0) - 2023-12-07
- feat(install): Use Script Install by @go-to-k in https://github.com/go-to-k/delstack/pull/292

## [v1.2.0](https://github.com/go-to-k/delstack/compare/v1.1.2...v1.2.0) - 2023-12-07
- chore(client): upgrade aws-sdk-go-v2/service/s3 to v1.47.3 and fix a breaking change by the version by @go-to-k in https://github.com/go-to-k/delstack/pull/291
- chore(deps): bump actions/setup-go from 4 to 5 by @dependabot in https://github.com/go-to-k/delstack/pull/289

## [v1.1.2](https://github.com/go-to-k/delstack/compare/v1.1.1...v1.1.2) - 2023-11-17
- docs: fix a default region in README by @go-to-k in https://github.com/go-to-k/delstack/pull/258
- chore(deps): aws sdk version up to 1.23.0 by @go-to-k in https://github.com/go-to-k/delstack/pull/263
- chore(deps): bump golang.org/x/sync from 0.3.0 to 0.5.0 by @dependabot in https://github.com/go-to-k/delstack/pull/254
- chore(deps): bump goreleaser/goreleaser-action from 4 to 5 by @dependabot in https://github.com/go-to-k/delstack/pull/234
- chore(deps): bump actions/checkout from 3 to 4 by @dependabot in https://github.com/go-to-k/delstack/pull/233

## [v1.1.1](https://github.com/go-to-k/delstack/compare/v1.1.0...v1.1.1) - 2023-10-29
- chore: minor improvement for keyword search by @go-to-k in https://github.com/go-to-k/delstack/pull/247

## [v1.1.0](https://github.com/go-to-k/delstack/compare/v1.0.4...v1.1.0) - 2023-10-08
- docs: change version sample for github actions by @go-to-k in https://github.com/go-to-k/delstack/pull/227
- docs: change GitHub Actions sample by @go-to-k in https://github.com/go-to-k/delstack/pull/228
- chore(deps): bump github.com/aws/aws-sdk-go-v2 from 1.20.3 to 1.21.0 by @dependabot in https://github.com/go-to-k/delstack/pull/213
- docs: Resource Types description in README by @go-to-k in https://github.com/go-to-k/delstack/pull/229
- test: add goleak by @go-to-k in https://github.com/go-to-k/delstack/pull/231
- ci: fix coverage report path by @go-to-k in https://github.com/go-to-k/delstack/pull/232
- chore: go version to 1.21 by @go-to-k in https://github.com/go-to-k/delstack/pull/243

## [v1.0.4](https://github.com/go-to-k/delstack/compare/v1.0.3...v1.0.4) - 2023-08-27
- fix: action shell by @go-to-k in https://github.com/go-to-k/delstack/pull/226

## [v1.0.3](https://github.com/go-to-k/delstack/compare/v1.0.2...v1.0.3) - 2023-08-27
- chore: add with in custom action by @go-to-k in https://github.com/go-to-k/delstack/pull/225

## [v1.0.2](https://github.com/go-to-k/delstack/compare/v1.0.1...v1.0.2) - 2023-08-26
- docs: change github actions sample by @go-to-k in https://github.com/go-to-k/delstack/pull/223
- chore: github action run commands by @go-to-k in https://github.com/go-to-k/delstack/pull/224
