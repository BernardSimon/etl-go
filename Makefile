.PHONY: build test race lint vet clean

# Build the project
build:
	go build -o bin/etl-go .

# Run all tests
test:
	go test ./...

# Run tests with race detector
race:
	go test -race ./...

# Run go vet
vet:
	go vet ./...

# Run lint (requires golangci-lint installed)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/
