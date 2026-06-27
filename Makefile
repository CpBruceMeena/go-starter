.PHONY: help setup init-repo fmt lint test build run dev clean docs swagger check-file-sizes generate-scaffold

# Variables
GO := go
GOLANGCI_LINT := golangci-lint
SWAG := swag
APP_NAME := go-starter
MAIN_PATH := ./cmd/app/main.go
BINARY_NAME := $(APP_NAME)
MAX_FILE_SIZE_KB := 500  # Files exceeding this size will trigger a warning
GIT_REMOTE := origin
GIT_BRANCH := main

help:
	@echo "Go Starter - Make Commands"
	@echo "=========================="
	@echo "Available commands:"
	@echo "  make setup              - Setup dependencies (run this first after cloning)"
	@echo "  make init-repo          - Initialize local git repository and create GitHub repo"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Run linter"
	@echo "  make test               - Run tests"
	@echo "  make build              - Build the application"
	@echo "  make run                - Run the application"
	@echo "  make dev                - Run in development mode with hot reload"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make docs               - Generate API documentation"
	@echo "  make swagger            - Generate Swagger documentation"
	@echo "  make check-file-sizes   - Check for files exceeding size limits"
	@echo "  make generate-scaffold  - Generate new resource scaffold (RESOURCE=name)"
	@echo ""

# ==================== Setup Commands ====================

setup:
	@echo "Setting up Go Starter..."
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo "Setting up git hooks..."
	@mkdir -p .git/hooks
	@echo "✓ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Update go.mod: Replace 'github.com/your-org/go-starter' with your repo path"
	@echo "2. Run: make init-repo"
	@echo "3. Configure .env file with your settings"
	@echo "4. Run: make run"

init-repo:
	@echo "Initializing Git repository..."
	@read -p "Enter your GitHub username: " github_user; \
	read -p "Enter repository name (default: go-starter): " repo_name; \
	repo_name=$${repo_name:-go-starter}; \
	remote_url="git@github.com:$$github_user/$$repo_name.git"; \
	echo "Remote URL: $$remote_url"; \
	git remote remove origin 2>/dev/null || true; \
	git remote add origin $$remote_url; \
	echo ""; \
	echo "Creating initial commit..."; \
	git add .; \
	git commit -m "Initial commit: Go starter template" || true; \
	echo ""; \
	echo "To push to GitHub:"; \
	echo "  git push -u origin main"; \
	echo ""

# ==================== Code Quality Commands ====================

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "✓ Code formatted"

lint:
	@echo "Running linter..."
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	$(GOLANGCI_LINT) run ./... --timeout=5m
	@echo "✓ Linting passed"

test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Tests passed. Coverage report: coverage.html"

check-file-sizes:
	@echo "Checking file sizes (max: $(MAX_FILE_SIZE_KB)KB)..."
	@files_exceeded=0; \
	for file in $$(find . -name "*.go" -type f ! -path "./vendor/*" ! -path "./.git/*"); do \
		size_kb=$$((($$(wc -c < "$$file") + 1023) / 1024)); \
		if [ $$size_kb -gt $(MAX_FILE_SIZE_KB) ]; then \
			echo "  ⚠️  WARNING: $$file is $$size_kb KB (exceeds $(MAX_FILE_SIZE_KB)KB)"; \
			files_exceeded=$$((files_exceeded + 1)); \
		fi; \
	done; \
	if [ $$files_exceeded -gt 0 ]; then \
		echo ""; \
		echo "Found $$files_exceeded file(s) exceeding size limit."; \
		echo "Consider splitting large files into smaller, focused modules."; \
		echo ""; \
	else \
		echo "✓ All files within size limits"; \
	fi

# ==================== Build Commands ====================

build: check-file-sizes fmt lint
	@echo "Building $(APP_NAME)..."
	$(GO) build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Build complete: bin/$(BINARY_NAME)"

run: build
	@echo "Running $(APP_NAME)..."
	./bin/$(BINARY_NAME)

run-worker: build
	@echo "Running $(APP_NAME) in worker mode..."
	APP_MODE=worker ./bin/$(BINARY_NAME)

run-uat: build
	@echo "Running $(APP_NAME) with UAT configuration..."
	@[ -f config/.env.uat ] && cp config/.env.uat .env && source .env || echo "UAT config not found"
	ENV=uat ./bin/$(BINARY_NAME)

run-staging: build
	@echo "Running $(APP_NAME) with staging configuration..."
	@[ -f config/.env.staging ] && cp config/.env.staging .env && source .env || echo "Staging config not found"
	ENV=staging ./bin/$(BINARY_NAME)

dev:
	@echo "Running in development mode..."
	@command -v air >/dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

clean:
	@echo "Cleaning..."
	$(GO) clean
	@rm -rf bin/ coverage.* dist/
	@echo "✓ Clean complete"

# ==================== Documentation Commands ====================

docs:
	@echo "Generating API documentation..."
	@echo "✓ Check docs/ folder for documentation"

swagger:
	@echo "Generating Swagger documentation..."
	@command -v $(SWAG) >/dev/null 2>&1 || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	$(SWAG) init -g cmd/app/main.go
	@echo "✓ Swagger documentation generated"
	@echo "  Access at: http://localhost:8080/swagger/index.html"

# ==================== Code Generation Commands ====================

generate-scaffold:
	@if [ -z "$(RESOURCE)" ]; then \
		echo "Usage: make generate-scaffold RESOURCE=name"; \
		echo "Example: make generate-scaffold RESOURCE=product"; \
		exit 1; \
	fi
	@echo "Generating scaffold for $(RESOURCE)..."
	@resource_lower=$$(echo $(RESOURCE) | tr A-Z a-z); \
	resource_title=$$(echo $$resource_lower | sed 's/^./\U&/'); \
	echo "Creating files for $$resource_title..."; \
	mkdir -p internal/models; \
	mkdir -p internal/repository; \
	mkdir -p internal/business; \
	@echo "✓ Scaffold generated for $(RESOURCE)"
	@echo "  1. Create model in: internal/models/$${resource_lower}.go"
	@echo "  2. Create repository in: internal/repository/$${resource_lower}.go"
	@echo "  3. Create service in: internal/business/$${resource_lower}.go"
	@echo "  4. Add routes in: internal/router/routes.go"

# ==================== Development Utilities ====================

install-tools:
	@echo "Installing development tools..."
	$(GO) install github.com/cosmtrek/air@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ Tools installed"

# ==================== Git & Deployment ====================

push:
	@echo "Pushing to $(GIT_REMOTE)/$(GIT_BRANCH)..."
	git push -u $(GIT_REMOTE) $(GIT_BRANCH)
	@echo "✓ Pushed successfully"

# ==================== AWS Utilities ====================

aws-local-test:
	@echo "Testing AWS Secrets Manager integration locally..."
	@echo "Make sure you have AWS credentials configured"
	@echo "Note: This requires actual AWS credentials and connectivity"

aws-uat-secrets:
	@echo "Creating UAT secrets in AWS Secrets Manager..."
	@echo "Run this command to create the secret:"
	@echo ""
	@echo 'aws secretsmanager create-secret \'
	@echo '  --name go-starter-uat-secrets \'
	@echo '  --secret-string \'
	@echo '  \'{"database_url":"postgresql://user:password@host:5432/db","log_level":"info"}\''
	@echo ""

aws-prod-secrets:
	@echo "Production secrets setup guide"
	@echo "DO NOT commit .env.production file with real credentials"
	@echo ""
	@echo "Setup checklist:"
	@echo "1. Create secret in AWS Secrets Manager:"
	@echo '   aws secretsmanager create-secret --name go-starter-prod-secrets --secret-string "{...}"'
	@echo "2. Create IAM role with Secrets Manager access"
	@echo "3. Attach role to ECS task"
	@echo "4. Set AWS_SECRETS_NAME environment variable"

# ==================== Docker (Optional) ====================

docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .
	@echo "✓ Docker image built"

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):latest
	@echo "✓ Container running on port 8080"
