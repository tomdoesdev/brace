.PHONY: build test clean run install fmt lint example-json example-yaml

# Build the brace compiler
build:
	go build -o bin/brace ./cmd/brace

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Run the compiler on the example file (JSON output)
run: build
	./bin/brace example.brace

# Run the compiler on the example file (YAML output)
run-yaml: build
	./bin/brace -format=yaml example.brace

# Install the compiler to GOPATH/bin
install:
	go install ./cmd/brace

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Generate example outputs
example-json: build
	./bin/brace -output=example.json example.brace
	@echo "JSON output generated: example.json"

example-yaml: build
	./bin/brace -format=yaml -output=example.yaml example.brace
	@echo "YAML output generated: example.yaml"

# Generate both example formats
examples: example-json example-yaml