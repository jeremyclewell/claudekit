# Makefile for claudekit

.PHONY: help build test test-unit test-vhs test-all clean install-vhs

help: ## Show this help message
	@echo "claudekit - Claude Code Project Setup Tool"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the claudekit binary
	@echo "🔨 Building claudekit..."
	@go build .
	@echo "✅ Build complete"

test: test-unit ## Run unit tests only (default)

test-unit: ## Run Go unit tests
	@echo "🧪 Running unit tests..."
	@go test -v ./... -run 'Test[^V]' 2>&1 | grep -v "^?"
	@echo "✅ Unit tests complete"

test-vhs: ## Run VHS visual tests (requires VHS installation)
	@echo "🎬 Running VHS visual tests..."
	@if ! command -v vhs > /dev/null; then \
		echo "❌ VHS not installed. Run 'make install-vhs' first"; \
		exit 1; \
	fi
	@go test -v -run TestVHSVisualScenarios
	@echo "✅ VHS tests complete"
	@echo ""
	@echo "📁 Screenshots saved to: specs/002-lets-make-the/vhs-tests/output/"
	@echo "📋 Review with: specs/002-lets-make-the/vhs-tests/VALIDATION-CHECKLIST.md"

test-all: test-unit test-vhs ## Run all tests (unit + VHS visual)

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete"

fmt: ## Format Go code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Format complete"

clean: ## Clean build artifacts and test outputs
	@echo "🧹 Cleaning..."
	@rm -f claudekit
	@rm -rf specs/002-lets-make-the/vhs-tests/output/
	@echo "✅ Clean complete"

install-vhs: ## Install VHS for visual testing
	@echo "📦 Installing VHS..."
	@if command -v brew > /dev/null; then \
		brew install vhs; \
	else \
		echo "Homebrew not found. Installing via Go..."; \
		go install github.com/charmbracelet/vhs@latest; \
	fi
	@echo "✅ VHS installed"

run: build ## Build and run claudekit
	@./claudekit

check: fmt vet test-unit ## Run all checks (fmt, vet, unit tests)
	@echo "✅ All checks passed"

ci: vet test-unit build ## CI/CD target (no VHS, no fmt changes)
	@echo "✅ CI checks complete"

.DEFAULT_GOAL := help
