RED=\033[31m
GREEN=\033[32m
RESET=\033[0m

COLORIZE_PASS=sed ''/PASS/s//$$(printf "$(GREEN)PASS$(RESET)")/''
COLORIZE_FAIL=sed ''/FAIL/s//$$(printf "$(RED)FAIL$(RESET)")/''

test:
	go test -v -cover ./... | $(COLORIZE_PASS) | $(COLORIZE_FAIL)
build: delstack
delstack: *.go cmd/delstack/main.go
	go build -o $@ cmd/delstack/main.go
install:
	go install github.com/go-to-k/delstack/cmd/delstack
