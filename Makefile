# DB Backup Pro - Makefile
# Cross-platform database backup application

.PHONY: dev build build-macos build-windows build-linux clean help

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME = dbbackup
VERSION = 1.0.0
WAILS = $(shell which wails 2>/dev/null || echo ~/go/bin/wails)

## Development
dev: ## Start development server
	$(WAILS) dev

## Building
build: ## Build for current platform
	$(WAILS) build

build-macos: ## Build for macOS (arm64 + amd64)
	$(WAILS) build -platform darwin/arm64 -o $(APP_NAME)-macos-arm64
	$(WAILS) build -platform darwin/amd64 -o $(APP_NAME)-macos-amd64

build-windows: ## Build for Windows (amd64)
	$(WAILS) build -platform windows/amd64 -o $(APP_NAME).exe

build-linux: ## Build for Linux (amd64)
	$(WAILS) build -platform linux/amd64 -o $(APP_NAME)-linux-amd64

build-all: build-macos build-windows build-linux ## Build for all platforms

## Utility
clean: ## Clean build artifacts
	rm -rf build/bin/*
	rm -rf frontend/dist

deps: ## Install dependencies
	go mod download
	cd frontend && npm install

generate: ## Generate Wails bindings
	$(WAILS) generate module

lint: ## Run linters
	go vet ./...
	cd frontend && npm run lint

test: ## Run tests
	go test ./...

## Help
help: ## Show this help
	@echo "DB Backup Pro - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
