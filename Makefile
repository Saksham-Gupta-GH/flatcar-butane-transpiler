.PHONY: build test lint clean run-example

BINARY_NAME := transpiler
EXAMPLE_INPUT := examples/worker-node.cloud-config.yaml
EXAMPLE_OUTPUT := examples/worker-node.butane.yaml

## build: Compile the transpiler binary
build:
	go build -o $(BINARY_NAME) .

## test: Run all unit tests with race detection
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## run-example: Transpile the example cloud-config to Butane YAML
run-example: build
	./$(BINARY_NAME) -input $(EXAMPLE_INPUT) -output $(EXAMPLE_OUTPUT)
	@echo "Output written to $(EXAMPLE_OUTPUT)"
	@cat $(EXAMPLE_OUTPUT)

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out $(EXAMPLE_OUTPUT)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@grep -E '^## ' Makefile | sed 's/## /  /'
