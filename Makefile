.PHONY: build release test race lint vet clean

# Build the project
build:
	go build -o bin/etl-go .

# Build release binaries for all supported platforms
release:
	./build-release.sh

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
