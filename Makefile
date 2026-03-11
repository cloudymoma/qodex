.PHONY: help build build-all run stop test clean deps frontend frontend-dev

APP_NAME := qodex
BINARY := bin/$(APP_NAME)
CONFIG := conf.yaml
PORT := 1983
PID_FILE := .qodex.pid

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	go mod download
	go mod tidy

build: ## Build the Go backend
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/server

build-all: frontend build ## Build frontend and Go backend
	@echo "Full build complete."

run: build ## Build and run the application
	@echo "Starting $(APP_NAME) on port $(PORT)..."
	./$(BINARY)

run-accesscode: build ## Build and run with access code protection
	@echo "Starting $(APP_NAME) on port $(PORT) with access code..."
	./$(BINARY) --accesscode

start: build ## Build and start in background
	@echo "Starting $(APP_NAME) in background on port $(PORT)..."
	@./$(BINARY) & echo $$! > $(PID_FILE)
	@echo "PID: $$(cat $(PID_FILE))"

stop: ## Stop the running application
	@echo "Stopping $(APP_NAME)..."
	@if [ -f $(PID_FILE) ]; then \
		kill $$(cat $(PID_FILE)) 2>/dev/null && rm -f $(PID_FILE) && echo "Stopped."; \
	else \
		ps aux | grep '$(BINARY)' | grep -v grep | awk '{print $$2}' | xargs -r kill && echo "Stopped." || echo "No running process found."; \
	fi

test: ## Run Go tests
	go test -v -race -cover ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format Go code
	go fmt ./...
	goimports -w .

lint: ## Run linter
	golangci-lint run ./...

frontend: ## Build frontend for production
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@mkdir -p web/static
	@rm -rf web/static/*
	@cp -r frontend/dist/* web/static/
	@echo "Frontend built and copied to web/static/"

frontend-dev: ## Start frontend dev server
	cd frontend && npm install && npm run dev

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f $(PID_FILE)

clean-all: clean ## Clean everything including cloned repos and indexes
	rm -rf $(HOME)/.qodex/*

.DEFAULT_GOAL := help
