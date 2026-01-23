.PHONY: test test-cover lint fmt vet clean check-agents

# Go parameters
GO_CMD := go
GO_TEST := $(GO_CMD) test
GO_COVER := $(GO_CMD) tool cover

test:
	$(GO_TEST) ./...

test-cover:
	$(GO_TEST) ./... -coverprofile=coverage.out
	$(GO_COVER) -html=coverage.out -o coverage.html

test-cover-html:
	$(GO_TEST) ./... -covermode=atomic -coverprofile=coverage.out
	$(GO_COVER) -html=coverage.out

lint:
	@which golangci-lint >/dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	$(GO_CMD) fmt ./...

vet:
	$(GO_CMD) vet ./...

clean:
	rm -f coverage.out coverage.html

check-agents:
	./scripts/check-agents-md.sh