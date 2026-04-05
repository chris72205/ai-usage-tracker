.PHONY: build test lint vet tidy check helm-lint

# Build the server binary
build:
	go build ./...

# Run tests with race detector (mirrors CI)
test:
	go test -race -count=1 ./...

# Run go vet
vet:
	go vet ./...

# Run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint: vet
	golangci-lint run ./...

# Tidy and verify module dependencies
tidy:
	go mod tidy
	go mod verify

# Lint the Helm chart (requires helm: https://helm.sh/docs/intro/install/)
helm-lint:
	helm lint charts/ai-usage-svc

# Full check: lint + test (mirrors CI)
check: lint test
