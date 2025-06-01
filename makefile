.PHONY: build test clean run

# Build the brace compiler
build:
	go build -o bin/brace ./cmd/brace

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Run the compiler on the example file
run: build
	./bin/brace example.brace

# Install the compiler to GOPATH/bin
install:
	go install ./cmd/brace

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run