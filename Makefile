.PHONY: test test-cover lint fmt vet clean check-agents

# Go parameters
GO_CMD := go
GO_TEST := $(GO_CMD) test
GO_COVER := $(GO_CMD) tool cover

.PHONY: help
help: ## Lists the available commands. Add a comment with '##' to describe a command.
	@grep -E '^[a-zA-Z_-].+:.*?## .*$$' $(MAKEFILE_LIST)\
		| sort\
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Runs the unit tests.
	$(GO_TEST) ./...

test-cover:
	$(GO_TEST) ./... -coverprofile=coverage.out
	$(GO_COVER) -html=coverage.out -o coverage.html

test-cover-html:
	$(GO_TEST) ./... -covermode=atomic -coverprofile=coverage.out
	$(GO_COVER) -html=coverage.out

lint:
	@which golangci-lint >/dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.8.0)
	golangci-lint run ./...

fmt:
	$(GO_CMD) fmt ./...

vet:
	$(GO_CMD) vet ./...

clean:
	rm -f coverage.out coverage.html

check-agents:
	./scripts/check-agents-md.sh

.PHONY: install-hooks ensure-lefthook
install-hooks: ensure-lefthook
	lefthook install

ensure-lefthook:
	@which lefthook >/dev/null || (echo "Installing lefthook..." && go install github.com/evilmartians/lefthook@latest)
