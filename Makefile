RED=\033[31m
GREEN=\033[32m
RESET=\033[0m

COLORIZE_PASS=sed ''/PASS/s//$$(printf "$(GREEN)PASS$(RESET)")/''
COLORIZE_FAIL=sed ''/FAIL/s//$$(printf "$(RED)FAIL$(RESET)")/''

VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -s -w \
			-X 'github.com/go-to-k/delstack/option.Version=$(VERSION)' \
			-X 'github.com/go-to-k/delstack/option.Revision=$(REVISION)'
GO_FILES:=$(shell find . -type f -name '*.go' -print)

test:
	go test -race -cover -v ./... -coverpkg=./... | $(COLORIZE_PASS) | $(COLORIZE_FAIL)
test_view:
	go test -race -cover -v ./... -coverprofile=cover_file.out -coverpkg=./...
	go tool cover -html=cover_file.out -o cover_file.html
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