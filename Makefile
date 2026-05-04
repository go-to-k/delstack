RED=\033[31m
GREEN=\033[32m
RESET=\033[0m

COLORIZE_PASS = sed "s/^\([- ]*\)\(PASS\)/\1$$(printf "$(GREEN)")\2$$(printf "$(RESET)")/g"
COLORIZE_FAIL = sed "s/^\([- ]*\)\(FAIL\)/\1$$(printf "$(RED)")\2$$(printf "$(RESET)")/g"

VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -s -w \
			-X 'github.com/go-to-k/delstack/internal/version.Version=$(VERSION)' \
			-X 'github.com/go-to-k/delstack/internal/version.Revision=$(REVISION)'
GO_FILES := $(shell find . -type f -name '*.go' -print)

DIFF_FILE := "$$(git diff --name-only --diff-filter=ACMRT | grep .go$ | xargs -I{} dirname {} | sort | uniq | xargs -I{} echo ./{})"

TEST_DIFF_RESULT := "$$(go test -race -cover -v $$(echo $(DIFF_FILE)) -coverpkg=./...)"
TEST_FULL_RESULT := "$$(go test -race -cover -v ./... -coverpkg=./...)"
TEST_COV_RESULT := "$$(go test -race -cover -v ./... -coverpkg=./... -coverprofile=cover.out.tmp)"

FAIL_CHECK := "^[^\s\t]*FAIL[^\s\t]*$$"

.PHONY: test_diff test test_view lint lint_diff mockgen shadow cognit deadcode run build install clean testgen_full testgen_full_retain testgen_large_template testgen_dependency testgen_dependency_retain testgen_preprocessor testgen_vpc_lambda testgen_lambda_edge testgen_deletion_protection testgen_deletion_protection_no_tp testgen_cdk_integration testgen_help e2e_full e2e_full_retain e2e_large_template e2e_dependency e2e_dependency_retain e2e_preprocessor e2e_vpc_lambda e2e_lambda_edge e2e_deletion_protection e2e_deletion_protection_no_tp e2e_cdk_integration e2e_help

test_diff:
	@! echo $(TEST_DIFF_RESULT) | $(COLORIZE_PASS) | $(COLORIZE_FAIL) | tee /dev/stderr | grep $(FAIL_CHECK) > /dev/null
test:
	@! echo $(TEST_FULL_RESULT) | $(COLORIZE_PASS) | $(COLORIZE_FAIL) | tee /dev/stderr | grep $(FAIL_CHECK) > /dev/null
test_view:
	@! echo $(TEST_COV_RESULT) | $(COLORIZE_PASS) | $(COLORIZE_FAIL) | tee /dev/stderr | grep $(FAIL_CHECK) > /dev/null
	cat cover.out.tmp | grep -v "**_mock.go" > cover.out
	rm cover.out.tmp
	go tool cover -func=cover.out
	go tool cover -html=cover.out -o cover.html
lint:
	golangci-lint run
lint_diff:
	golangci-lint run $$(echo $(DIFF_FILE))
mockgen:
	go generate ./...
shadow:
	find . -type f -name '*.go' | sed -e "s/\/[^\.\/]*\.go//g" | uniq | xargs shadow
cognit:
	gocognit -top 10 ./ | grep -v "_test.go"
deadcode:
	deadcode ./...
run:
	go mod tidy
	go run -ldflags "$(LDFLAGS)" cmd/delstack/main.go $${OPT}
build: $(GO_FILES)
	go mod tidy
	go build -ldflags "$(LDFLAGS)" -o delstack cmd/delstack/main.go
install:
	go install -ldflags "$(LDFLAGS)" github.com/go-to-k/delstack/cmd/delstack
clean:
	go clean
	rm -f delstack

# Test stack generation commands
# ==================================

# Run test stack generator with full resources
testgen_full:
	@echo "Running test stack generator with full resources..."
	@cd e2e/full && go mod tidy && go run deploy.go $(OPT)

# Run test stack generator for all RETAIN resources to test \`-f\` option
testgen_full_retain:
	@echo "Running test stack generator for all RETAIN resources..."
	@cd e2e/full && go mod tidy && go run deploy.go -r $(OPT)

# Generate and deploy large CloudFormation template for testing S3 upload functionality (>51200 bytes)
# S3 bucket is automatically deleted after stack creation
testgen_large_template:
	@echo "Setting up large CloudFormation template test stack..."
	@cd e2e/s3_template_cfn && go mod tidy && go run main.go $(OPT)

# Generate and deploy CDK dependency test stacks for testing complex dependency graphs
testgen_dependency:
	@echo "Setting up CDK dependency test stacks..."
	@cd e2e/dependency && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK dependency test stacks with RETAIN resources
testgen_dependency_retain:
	@echo "Setting up CDK dependency test stacks with RETAIN resources..."
	@cd e2e/dependency && go mod tidy && go run deploy.go -r $(OPT)

# Generate and deploy Lambda@Edge test stacks
testgen_lambda_edge:
	@echo "Setting up Lambda@Edge test stacks..."
	@cd e2e/lambda_edge && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy preprocessor test stacks for Lambda VPC detachment
testgen_preprocessor:
	@echo "Setting up preprocessor test stacks for Lambda VPC detachment..."
	@cd e2e/preprocessor && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy VPC Lambda orphan-ENI test stack (issue #637)
testgen_vpc_lambda:
	@echo "Setting up VPC Lambda orphan-ENI test stack..."
	@cd e2e/vpc_lambda && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy deletion protection test stacks
testgen_deletion_protection:
	@echo "Setting up deletion protection test stacks..."
	@cd e2e/deletion_protection && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy deletion protection test stacks without stack TerminationProtection
testgen_deletion_protection_no_tp:
	@echo "Setting up deletion protection test stacks without stack TerminationProtection..."
	@cd e2e/deletion_protection && go mod tidy && go run deploy.go -t $(OPT)

# Generate and deploy CDK integration test stacks for `delstack cdk` subcommand
testgen_cdk_integration:
	@echo "Setting up CDK integration test stacks..."
	@cd e2e/cdk_integration && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK integration test stacks with RETAIN resources
testgen_cdk_integration_retain:
	@echo "Setting up CDK integration test stacks with RETAIN resources..."
	@cd e2e/cdk_integration && go mod tidy && go run deploy.go -r $(OPT)

# Generate and deploy CDK cross-region test stacks for `delstack cdk` cross-region deletion
testgen_cdk_cross_region:
	@echo "Setting up CDK cross-region test stacks..."
	@cd e2e/cdk_cross_region && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK cross-region test stacks with RETAIN resources
testgen_cdk_cross_region_retain:
	@echo "Setting up CDK cross-region test stacks with RETAIN resources..."
	@cd e2e/cdk_cross_region && go mod tidy && go run deploy.go -r $(OPT)

# Generate and deploy CDK app option test stack for `delstack cdk -a` testing
testgen_cdk_app_option:
	@echo "Setting up CDK app option test stack..."
	@cd e2e/cdk_app_option && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK glob pattern test stacks for `delstack cdk -s` with glob patterns
testgen_cdk_glob:
	@echo "Setting up CDK glob pattern test stacks..."
	@cd e2e/cdk_glob && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK Stage test stacks for `delstack cdk` with CDK Stages
testgen_cdk_stage:
	@echo "Setting up CDK Stage test stacks..."
	@cd e2e/cdk_stage && go mod tidy && go run deploy.go $(OPT)

# Generate and deploy CDK TerminationProtection test stacks for `delstack cdk -f`
testgen_cdk_termination_protection:
	@echo "Setting up CDK TerminationProtection test stacks..."
	@cd e2e/cdk_termination_protection && go mod tidy && go run deploy.go $(OPT)

# Help for test stack generation
testgen_help:
	@echo "Test stack generation targets:"
	@echo "  testgen_full                - Run the test stack generator with full resources"
	@echo "  testgen_full_retain         - Run the test stack generator for all RETAIN resources to test \`-f\` option"
	@echo "  testgen_large_template      - Generate and deploy large CFn template (>51KB), S3 bucket auto-deleted"
	@echo "  testgen_dependency          - Generate and deploy CDK dependency test stacks for complex dependency graphs"
	@echo "  testgen_dependency_retain   - Generate and deploy CDK dependency test stacks with RETAIN resources"
	@echo "  testgen_lambda_edge         - Generate and deploy Lambda@Edge test stacks"
	@echo "  testgen_preprocessor        - Generate and deploy preprocessor test stacks for Lambda VPC detachment"
	@echo "  testgen_vpc_lambda          - Generate and deploy VPC Lambda orphan-ENI test stack (issue #637)"
	@echo "  testgen_deletion_protection       - Generate and deploy deletion protection test stacks"
	@echo "  testgen_deletion_protection_no_tp - Generate and deploy deletion protection test stacks without stack TP"
	@echo "  testgen_cdk_integration           - Generate and deploy CDK integration test stacks"
	@echo "  testgen_cdk_integration_retain    - Generate and deploy CDK integration test stacks with RETAIN"
	@echo "  testgen_cdk_cross_region          - Generate and deploy CDK cross-region test stacks"
	@echo "  testgen_cdk_cross_region_retain   - Generate and deploy CDK cross-region test stacks with RETAIN"
	@echo "  testgen_cdk_app_option            - Generate and deploy CDK app option test stack"
	@echo "  testgen_cdk_glob                  - Generate and deploy CDK glob pattern test stacks"
	@echo "  testgen_cdk_stage                 - Generate and deploy CDK Stage test stacks"
	@echo "  testgen_cdk_termination_protection - Generate and deploy CDK TerminationProtection test stacks"
	@echo ""
	@echo "Example usage:"
	@echo "  make testgen_full"
	@echo "  make testgen_full OPT=\"-s my-stage\""
	@echo "  make testgen_full OPT=\"-p my-profile\""
	@echo "  make testgen_full_retain"
	@echo "  make testgen_full_retain OPT=\"-s my-stage\""
	@echo "  make testgen_full_retain OPT=\"-p my-profile\""
	@echo "  make testgen_large_template"
	@echo "  make testgen_large_template OPT=\"-p my-profile\""
	@echo "  make testgen_dependency"
	@echo "  make testgen_dependency OPT=\"-s my-stage\""
	@echo "  make testgen_dependency OPT=\"-p my-profile\""
	@echo "  make testgen_dependency_retain"
	@echo "  make testgen_dependency_retain OPT=\"-s my-stage\""
	@echo "  make testgen_dependency_retain OPT=\"-p my-profile\""
	@echo "  make testgen_lambda_edge"
	@echo "  make testgen_lambda_edge OPT=\"-s my-stage\""
	@echo "  make testgen_preprocessor"
	@echo "  make testgen_preprocessor OPT=\"-s my-stage\""
	@echo "  make testgen_preprocessor OPT=\"-p my-profile\""
	@echo "  make testgen_vpc_lambda"
	@echo "  make testgen_vpc_lambda OPT=\"-s my-stage\""
	@echo "  make testgen_vpc_lambda OPT=\"-p my-profile\""
	@echo "  make testgen_deletion_protection"
	@echo "  make testgen_deletion_protection OPT=\"-s my-stage\""
	@echo "  make testgen_deletion_protection OPT=\"-p my-profile\""
	@echo "  make testgen_deletion_protection_no_tp"
	@echo "  make testgen_deletion_protection_no_tp OPT=\"-s my-stage\""
	@echo "  make testgen_deletion_protection_no_tp OPT=\"-p my-profile\""
	@echo "  make testgen_cdk_integration"
	@echo "  make testgen_cdk_integration OPT=\"-s my-stage\""
	@echo "  make testgen_cdk_integration OPT=\"-p my-profile\""
	@echo "  make testgen_cdk_integration_retain"
	@echo "  make testgen_cdk_cross_region"
	@echo "  make testgen_cdk_cross_region OPT=\"-s my-stage\""
	@echo "  make testgen_cdk_cross_region_retain"
	@echo "  make testgen_cdk_stage"
	@echo "  make testgen_cdk_stage OPT=\"-s my-stage\""
	@echo "  make testgen_cdk_termination_protection"
	@echo "  make testgen_cdk_termination_protection OPT=\"-s my-stage\""

# E2E test commands (testgen + delstack run)
# ==================================

E2E_RANDOM = $(shell awk 'BEGIN{srand(); printf "%04d", int(rand()*10000)}')

# Run full resource E2E test (deploy + delete)
e2e_full: STAGE = e2e-full-$(E2E_RANDOM)
e2e_full:
	@$(MAKE) testgen_full OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) $(OPT)"

# Run full resource E2E test with RETAIN resources (deploy + force delete)
e2e_full_retain: STAGE = e2e-full-$(E2E_RANDOM)
e2e_full_retain:
	@$(MAKE) testgen_full_retain OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) -f $(OPT)"

# Run large template E2E test (deploy + force delete)
e2e_large_template: STAGE = e2e-large-tpl-$(E2E_RANDOM)
e2e_large_template:
	@$(MAKE) testgen_large_template OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) -f $(OPT)"

# Run dependency graph E2E test (deploy 6 stacks + delete all)
e2e_dependency: STAGE = e2e-dependency-$(E2E_RANDOM)
e2e_dependency:
	@$(MAKE) testgen_dependency OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE)-Stack-A -s $(STAGE)-Stack-B -s $(STAGE)-Stack-C -s $(STAGE)-Stack-D -s $(STAGE)-Stack-E -s $(STAGE)-Stack-F $(OPT)"

# Run dependency graph E2E test with RETAIN resources (deploy + force delete)
e2e_dependency_retain: STAGE = e2e-dependency-$(E2E_RANDOM)
e2e_dependency_retain:
	@$(MAKE) testgen_dependency_retain OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE)-Stack-A -s $(STAGE)-Stack-B -s $(STAGE)-Stack-C -s $(STAGE)-Stack-D -s $(STAGE)-Stack-E -s $(STAGE)-Stack-F -f $(OPT)"

# Run Lambda@Edge E2E test (deploy + delete)
e2e_lambda_edge: STAGE = e2e-lambda-edge-$(E2E_RANDOM)
e2e_lambda_edge:
	@$(MAKE) testgen_lambda_edge OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) $(OPT)"

# Run preprocessor E2E test (deploy + delete)
e2e_preprocessor: STAGE = e2e-preprocessor-$(E2E_RANDOM)
e2e_preprocessor:
	@$(MAKE) testgen_preprocessor OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) $(OPT)"

# Run VPC Lambda orphan-ENI E2E test (deploy + force delete)
e2e_vpc_lambda: STAGE = e2e-vpc-lambda-$(E2E_RANDOM)
e2e_vpc_lambda:
	@$(MAKE) testgen_vpc_lambda OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) -y $(OPT)"

# Run deletion protection E2E test (deploy + force delete)
e2e_deletion_protection: STAGE = e2e-dp-$(E2E_RANDOM)
e2e_deletion_protection:
	@$(MAKE) testgen_deletion_protection OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) -f -y $(OPT)"

# Run deletion protection E2E test without TerminationProtection (deploy + force delete)
e2e_deletion_protection_no_tp: STAGE = e2e-dp-$(E2E_RANDOM)
e2e_deletion_protection_no_tp:
	@$(MAKE) testgen_deletion_protection_no_tp OPT="-s $(STAGE) $(OPT)"
	@$(MAKE) run OPT="-s $(STAGE) -f $(OPT)"

# Run CDK integration E2E test (deploy + delstack cdk)
e2e_cdk_integration: STAGE = e2e-cdk-$(E2E_RANDOM)
e2e_cdk_integration: build
	@$(MAKE) testgen_cdk_integration OPT="-s $(STAGE) $(OPT)"
	@cd e2e/cdk_integration/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -c RETAIN_MODE=false -f -y $(OPT)

# Run CDK integration E2E test with RETAIN resources (deploy + force delete)
e2e_cdk_integration_retain: STAGE = e2e-cdk-$(E2E_RANDOM)
e2e_cdk_integration_retain: build
	@$(MAKE) testgen_cdk_integration OPT="-s $(STAGE) -r $(OPT)"
	@cd e2e/cdk_integration/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -c RETAIN_MODE=true -f -y $(OPT)

# Run CDK --app option E2E test (both directory and command)
# Deploys 1 stack, deletes via -a (cdk.out directory), redeploys, deletes via -a (app command)
e2e_cdk_app_option: STAGE = e2e-cdk-appopt-$(E2E_RANDOM)
e2e_cdk_app_option: build
	@$(MAKE) testgen_cdk_app_option OPT="-s $(STAGE) $(OPT)"
	@echo "=== Test 1: -a with cdk.out directory ==="
	@cd e2e/cdk_app_option/cdk && ../../../delstack cdk -a cdk.out -s $(STAGE)-AppOptStack -f -y $(OPT)
	@echo "=== Test 2: -a with app command (redeploy first) ==="
	@cd e2e/cdk_app_option && go run deploy.go -s $(STAGE) $(OPT)
	@cd e2e/cdk_app_option/cdk && rm -rf cdk.out
	@cd e2e/cdk_app_option/cdk && ../../../delstack cdk -a "go mod download && go run cdk.go" -c PJ_PREFIX=$(STAGE) -s $(STAGE)-AppOptStack -f -y $(OPT)

# Run CDK cross-region E2E test (deploy + delstack cdk)
e2e_cdk_cross_region: STAGE = e2e-cdk-xr-$(E2E_RANDOM)
e2e_cdk_cross_region: build
	@$(MAKE) testgen_cdk_cross_region OPT="-s $(STAGE) $(OPT)"
	@cd e2e/cdk_cross_region/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -c RETAIN_MODE=false -f -y $(OPT)

# Run CDK cross-region E2E test with RETAIN resources (deploy + force delete)
e2e_cdk_cross_region_retain: STAGE = e2e-cdk-xr-$(E2E_RANDOM)
e2e_cdk_cross_region_retain: build
	@$(MAKE) testgen_cdk_cross_region_retain OPT="-s $(STAGE) $(OPT)"
	@cd e2e/cdk_cross_region/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -c RETAIN_MODE=true -f -y $(OPT)

# Run CDK glob pattern E2E test (deploy 3 stacks, delete 2 with glob, then delete remaining)
e2e_cdk_glob: STAGE = e2e-cdk-glob-$(E2E_RANDOM)
e2e_cdk_glob: build
	@$(MAKE) testgen_cdk_glob OPT="-s $(STAGE) $(OPT)"
	@echo "=== Test 1: Delete Api* top-level stacks with glob pattern ==="
	@cd e2e/cdk_glob/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -s "$(STAGE)-Api*" -f -y $(OPT)
	@echo "=== Test 2: Delete Staged* stacks (inside Stage) with glob pattern ==="
	@cd e2e/cdk_glob/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -s "$(STAGE)-Staged*" -f -y $(OPT)
	@echo "=== Test 3: Delete remaining WebStack by exact name ==="
	@cd e2e/cdk_glob/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -s $(STAGE)-WebStack -f -y $(OPT)

# Run CDK Stage E2E test (deploy + delstack cdk)
e2e_cdk_stage: STAGE = e2e-cdk-stg-$(E2E_RANDOM)
e2e_cdk_stage: build
	@$(MAKE) testgen_cdk_stage OPT="-s $(STAGE) $(OPT)"
	@cd e2e/cdk_stage/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -f -y $(OPT)

# Run CDK TerminationProtection E2E test (deploy + delstack cdk -f)
e2e_cdk_termination_protection: STAGE = e2e-cdk-tp-$(E2E_RANDOM)
e2e_cdk_termination_protection: build
	@$(MAKE) testgen_cdk_termination_protection OPT="-s $(STAGE) $(OPT)"
	@cd e2e/cdk_termination_protection/cdk && ../../../delstack cdk -c PJ_PREFIX=$(STAGE) -f -y $(OPT)

# Help for E2E test targets
e2e_help:
	@echo "E2E test targets (testgen + delstack run):"
	@echo "  e2e_full                    - Deploy full resources and delete with delstack"
	@echo "  e2e_full_retain             - Deploy full RETAIN resources and force delete"
	@echo "  e2e_large_template          - Deploy large CFn template and force delete"
	@echo "  e2e_dependency              - Deploy 6 dependency stacks and delete all"
	@echo "  e2e_dependency_retain       - Deploy 6 dependency stacks with RETAIN and force delete"
	@echo "  e2e_lambda_edge             - Deploy Lambda@Edge stacks and delete (takes ~20 min)"
	@echo "  e2e_preprocessor            - Deploy preprocessor stacks and delete"
	@echo "  e2e_vpc_lambda              - Deploy VPC Lambda + setup orphan ENI and force delete (issue #637)"
	@echo "  e2e_deletion_protection     - Deploy deletion protection stacks and force delete"
	@echo "  e2e_deletion_protection_no_tp - Deploy deletion protection stacks (no TP) and force delete"
	@echo "  e2e_cdk_integration          - Deploy CDK stacks and delete with 'delstack cdk'"
	@echo "  e2e_cdk_integration_retain   - Deploy CDK stacks with RETAIN and force delete"
	@echo "  e2e_cdk_cross_region         - Deploy CDK cross-region stacks and delete"
	@echo "  e2e_cdk_cross_region_retain  - Deploy CDK cross-region stacks with RETAIN and force delete"
	@echo "  e2e_cdk_app_option           - Deploy CDK stack and test --app with directory and command"
	@echo "  e2e_cdk_glob                 - Deploy CDK stacks and test -s with glob patterns"
	@echo "  e2e_cdk_stage                - Deploy CDK Stage stacks and delete"
	@echo "  e2e_cdk_termination_protection - Deploy CDK TP stacks and force delete"
	@echo ""
	@echo "Options:"
	@echo "  STAGE=<name>  - Override default stage name (default: auto-generated with random suffix)"
	@echo "  OPT=\"-p <profile>\"  - Pass additional options (e.g., AWS profile)"
	@echo ""
	@echo "Example usage:"
	@echo "  make e2e_full"
	@echo "  make e2e_full STAGE=my-stage"
	@echo "  make e2e_full STAGE=my-stage OPT=\"-p my-profile\""
	@echo "  make e2e_cdk_integration"
	@echo "  make e2e_cdk_cross_region"
	@echo "  make e2e_cdk_stage"
	@echo "  make e2e_cdk_termination_protection"