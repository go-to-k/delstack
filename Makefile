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

.PHONY: test_diff test test_view lint lint_diff mockgen shadow cognit deadcode run build install clean testgen testgen_help

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

# Run test stack generator
testgen:
	@echo "Running test stack generator..."
	@cd testdata && go mod tidy && go run deploy.go $(OPT)

# Run test stack generator for all RETAIN resources to test \`-f\` option 
testgen_retain:
	@echo "Running test stack generator for all RETAIN resources..."
	@cd testdata && go mod tidy && go run deploy.go -r $(OPT)

# Help for test stack generation
testgen_help:
	@echo "Test stack generation targets:"
	@echo "  testgen         - Run the test stack generator"
	@echo "  testgen_retain  - Run the test stack generator for all RETAIN resources to test \`-f\` option"
	@echo ""
	@echo "Example usage:"
	@echo "  make testgen"
	@echo "  make testgen OPT=\"-s my-stage\""
	@echo "  make testgen OPT=\"-p my-profile\""
	@echo "  make testgen_retain"
	@echo "  make testgen_retain OPT=\"-s my-stage\""
	@echo "  make testgen_retain OPT=\"-p my-profile\""