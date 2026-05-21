# JobLinker Makefile
# Targets: up, down, restart, build, test, lint, clean, logs

SHELL := /bin/bash
ROOT := $(CURDIR)
BACKEND := $(ROOT)/joblinker/backend
FRONTEND := $(ROOT)/joblinker/frontend
DOCKER := docker compose -f $(ROOT)/joblinker/docker-compose.yml

.PHONY: help up down restart build-backend build-frontend \
        test-backend test-frontend lint-backend lint-frontend \
        logs-backend logs-frontend logs-db clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Infrastructure ──────────────────────────────────────────────

up: ## Start all Docker services (PostgreSQL, Redis, RabbitMQ)
	$(DOCKER) up -d

down: ## Stop all Docker services
	$(DOCKER) down

restart: down up ## Restart all Docker services

logs: ## Tail all Docker logs
	$(DOCKER) logs -f

# ─── Build ───────────────────────────────────────────────────────

build-backend: ## Build the Go backend binary
	cd $(BACKEND)/cmd/server && go build -o $(BACKEND)/joblinker-server .

build-frontend: ## Build the Next.js frontend
	cd $(FRONTEND) && npm run build

build: build-backend build-frontend ## Build everything

# ─── Test ────────────────────────────────────────────────────────

test-backend: ## Run all Go tests
	cd $(BACKEND) && go test ./...

test-frontend: ## Run all frontend (Jest) tests
	cd $(FRONTEND) && npx jest

test: test-backend test-frontend ## Run all tests

# ─── Lint ────────────────────────────────────────────────────────

lint-backend: ## Run Go vet
	cd $(BACKEND) && go vet ./...

lint-frontend: ## Run Next.js lint
	cd $(FRONTEND) && npm run lint

lint: lint-backend lint-frontend ## Run all linters

# ─── Dev Servers ─────────────────────────────────────────────────

run-backend: ## Run the backend in dev mode
	cd $(BACKEND)/cmd/server && go run . 2>&1

run-frontend: ## Run the frontend in dev mode
	cd $(FRONTEND) && npm run dev

# ─── Clean ───────────────────────────────────────────────────────

clean: ## Clean build artifacts
	rm -f $(BACKEND)/joblinker-server
	rm -rf $(FRONTEND)/.next
