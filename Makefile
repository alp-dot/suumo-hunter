.PHONY: build build-lambda lint test clean deploy

# Build settings
BINARY_NAME=bootstrap
BUILD_DIR=build
LAMBDA_DIR=cmd/lambda

# Go settings
GOOS=linux
GOARCH=arm64
CGO_ENABLED=0

# Build the Lambda binary (ARM64 for Graviton2)
build: build-lambda

build-lambda:
	@echo "Building Lambda binary for ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./$(LAMBDA_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Create deployment package
package: build-lambda
	@echo "Creating deployment package..."
	cd $(BUILD_DIR) && zip -j lambda.zip $(BINARY_NAME)
	@echo "Package created: $(BUILD_DIR)/lambda.zip"

# Run linter
lint:
	@echo "Running golangci-lint..."
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" golangci-lint run ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./internal/...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Deploy to AWS (requires ENV variable)
# Usage: make deploy ENV=shibuya
deploy: package
ifndef ENV
	$(error ENV is required. Usage: make deploy ENV=shibuya)
endif
	@echo "Deploying $(ENV) to AWS..."
	cd terraform/$(ENV) && terraform apply -auto-approve

# Initialize terraform for an environment
# Usage: make terraform-init ENV=shibuya
terraform-init:
ifndef ENV
	$(error ENV is required. Usage: make terraform-init ENV=shibuya)
endif
	@echo "Initializing Terraform for $(ENV)..."
	cd terraform/$(ENV) && terraform init

# Plan terraform changes
# Usage: make terraform-plan ENV=shibuya
terraform-plan: package
ifndef ENV
	$(error ENV is required. Usage: make terraform-plan ENV=shibuya)
endif
	@echo "Planning Terraform changes for $(ENV)..."
	cd terraform/$(ENV) && terraform plan

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Check for security vulnerabilities
vuln:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...
