# Makefile for Claude Code Template
# Provides a simple interface for common operations

# Configuration
SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -euo pipefail -c

# Directories
PROJECT_ROOT := $(shell pwd)
TESTS_DIR := $(PROJECT_ROOT)/tests
SCRIPTS_DIR := $(PROJECT_ROOT)/scripts
CLAUDE_DIR := $(PROJECT_ROOT)/.claude
HOOKS_DIR := $(CLAUDE_DIR)/hooks

# Scripts
SCANNER_SCRIPT := $(HOOKS_DIR)/mcp-security-scanner.sh
TEST_SCRIPT := $(TESTS_DIR)/test-scanner.sh
INSTALL_SCRIPT := $(SCRIPTS_DIR)/install-hooks.sh

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Default target
.DEFAULT_GOAL := help

##@ Testing

.PHONY: test
test: ## Run complete test suite
	@echo -e "$(BLUE)🧪 Running security scanner tests...$(NC)"
	@$(TEST_SCRIPT)

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo -e "$(BLUE)🧪 Running security scanner tests...$(NC)"
	@$(TEST_SCRIPT)

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo -e "$(BLUE)🧪 Running security scanner tests...$(NC)"
	@$(TEST_SCRIPT)

##@ Quality Assurance

.PHONY: lint
lint: ## Run ShellCheck on all scripts
	@echo -e "$(BLUE)🔍 Running ShellCheck on all scripts...$(NC)"
	@find . -name "*.sh" -type f -perm +111 | grep -v node_modules | xargs shellcheck --rcfile=.github/linters/.shellcheckrc

##@ Installation and Setup

.PHONY: install
install: ## Install hooks to current project
	@echo -e "$(BLUE)📦 Installing hooks...$(NC)"
	@chmod +x "$(SCANNER_SCRIPT)"
	@echo -e "$(GREEN)✅ Installation complete$(NC)"

##@ Maintenance

.PHONY: clean
clean: ## Remove test artifacts and logs
	@echo -e "$(BLUE)🧹 Cleaning up test artifacts...$(NC)"
	@find "$(CLAUDE_DIR)" -name "*.log" -type f -delete 2>/dev/null || true
	@find "$(TESTS_DIR)" -name "*.log" -type f -delete 2>/dev/null || true
	@echo -e "$(GREEN)✅ Cleanup complete$(NC)"

##@ Development

.PHONY: check-tools
check-tools: ## Check for required and optional tools
	@echo -e "$(BLUE)🔧 Checking tool availability...$(NC)"
	@echo
	@echo "Required tools:"
	@for tool in jq grep awk mktemp; do
		if command -v $$tool >/dev/null 2>&1; then
			echo -e "  ✅ $$tool"
		else
			echo -e "  ❌ $$tool (required)"
		fi
	done
	@echo
	@echo "Optional security tools:"
	@for tool in gitleaks trufflehog git-secrets; do
		if command -v $$tool >/dev/null 2>&1; then
			echo -e "  ✅ $$tool"
		else
			echo -e "  ⚪ $$tool (optional)"
		fi
	done

##@ Information

.PHONY: status
status: ## Show current status and configuration
	@echo -e "$(BLUE)📊 Claude Code Template Status$(NC)"
	@echo
	@echo "Configuration:"
	@echo "  Project root: $(PROJECT_ROOT)"
	@echo "  Claude directory: $(CLAUDE_DIR)"
	@echo "  Scanner script: $(SCANNER_SCRIPT)"
	@echo
	@echo "Files:"
	@ls -la "$(SCANNER_SCRIPT)" 2>/dev/null || echo -e "  $(RED)❌ Scanner script not found$(NC)"
	@ls -la "$(CLAUDE_DIR)/settings.json" 2>/dev/null || echo -e "  $(RED)❌ Settings file not found$(NC)"
	@$(MAKE) check-tools

.PHONY: help
help: ## Display this help
	@echo -e "$(BLUE)Claude Code Template - Make Targets$(NC)"
	@echo
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo
	@echo -e "$(BLUE)Examples:$(NC)"
	@echo "  make test              # Run all tests"
	@echo "  make lint              # Run code quality checks"
	@echo "  make install           # Install hooks"
	@echo "  make clean             # Clean up artifacts"
	@echo

# Ensure scripts directory exists for utility scripts
$(SCRIPTS_DIR):
	@mkdir -p "$(SCRIPTS_DIR)"