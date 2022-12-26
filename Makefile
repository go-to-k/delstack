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

TEST_RESULT := "$$(go test -race -cover -v ./... -coverpkg=./...)"
TEST_COV_RESULT := "$$(go test -race -cover -v ./... -coverpkg=./... -coverprofile=cover_file.out)"
FAIL_CHECK := "^[^\s\t]*FAIL[^\s\t]*$$"

test:
	@! echo $(TEST_RESULT) | $(COLORIZE_PASS) | $(COLORIZE_FAIL) | tee /dev/stderr | grep $(FAIL_CHECK) > /dev/null
test_view:
	@! echo $(TEST_COV_RESULT) | $(COLORIZE_PASS) | $(COLORIZE_FAIL) | tee /dev/stderr | grep $(FAIL_CHECK) > /dev/null
	go tool cover -html=cover_file.out -o cover_file.html
shadow:
	find . -type f -name '*.go' | sed -e "s/\/[^\.\/]*\.go//g" | uniq | xargs shadow
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