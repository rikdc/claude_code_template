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
PROTECT_BRANCH_SCRIPT := $(HOOKS_DIR)/protect-main-branch.sh
TEST_SCRIPT := $(TESTS_DIR)/test-scanner.sh
TEST_PROTECT_BRANCH_SCRIPT := $(TESTS_DIR)/test-protect-main-branch.sh
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
	@echo
	@echo -e "$(BLUE)🧪 Running protected branch hook tests...$(NC)"
	@$(TEST_PROTECT_BRANCH_SCRIPT)

##@ Quality Assurance

.PHONY: lint
lint: ## Run ShellCheck on shell scripts and markdownlint on Markdown files
	@echo -e "$(BLUE)🔍 Running linting checks...$(NC)"
	@echo
	@echo "Shell scripts:"
	@if command -v shellcheck >/dev/null 2>&1; then \
		find . -name "*.sh" -type f -perm +111 | grep -v node_modules | xargs shellcheck --rcfile=.github/linters/.shellcheckrc && \
		echo -e "  $(GREEN)✅ ShellCheck passed$(NC)"; \
	else \
		echo -e "  $(YELLOW)⚠️  ShellCheck not installed - skipping shell script linting$(NC)"; \
	fi
	@echo
	@echo "Markdown files:"
	@if command -v markdownlint >/dev/null 2>&1; then \
		markdownlint --config .github/linters/.markdown-lint.yml **/*.md && \
		echo -e "  $(GREEN)✅ markdownlint passed$(NC)"; \
	else \
		echo -e "  $(YELLOW)⚠️  markdownlint not installed - skipping Markdown linting$(NC)"; \
	fi

##@ Installation and Setup

.PHONY: install
install: ## Install hooks to current project
	@echo -e "$(BLUE)📦 Installing hooks...$(NC)"
	@chmod +x "$(SCANNER_SCRIPT)"
	@chmod +x "$(PROTECT_BRANCH_SCRIPT)"
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
	@for tool in jq grep awk mktemp; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo -e "  ✅ $$tool"; \
		else \
			echo -e "  ❌ $$tool (required)"; \
		fi; \
	done
	@echo
	@echo "Development tools:"
	@for tool in shellcheck markdownlint; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo -e "  ✅ $$tool"; \
		else \
			echo -e "  ⚪ $$tool (recommended for linting)"; \
		fi; \
	done
	@echo
	@echo "Optional security tools:"
	@for tool in gitleaks trufflehog git-secrets; do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo -e "  ✅ $$tool"; \
		else \
			echo -e "  ⚪ $$tool (optional)"; \
		fi; \
	done

.PHONY: sync-plugins
sync-plugins: ## Sync files from .claude/ to plugins/ directories
	@echo -e "$(BLUE)🔄 Syncing files to plugins/...$(NC)"
	@echo
	@echo "Syncing skills to plugins/dev-skills/..."
	@mkdir -p plugins/dev-skills/skills
	@cp -r .claude/skills/* plugins/dev-skills/skills/
	@echo
	@echo "Syncing commands to plugin directories..."
	@mkdir -p plugins/git-workflow/commands
	@cp .claude/commands/gh/commit.md plugins/git-workflow/commands/commit.md
	@cp .claude/commands/gh/pr.md plugins/git-workflow/commands/pr.md
	@cp .claude/commands/changelog.md plugins/git-workflow/commands/changelog.md
	@mkdir -p plugins/pm-tools/commands
	@cp .claude/commands/pm/*.md plugins/pm-tools/commands/
	@mkdir -p plugins/code-quality/commands
	@cp .claude/commands/dev/*.md plugins/code-quality/commands/
	@cp .claude/commands/clean.md plugins/code-quality/commands/clean.md
	@mkdir -p plugins/prompt-tools/commands
	@cp .claude/commands/prompt-reviewer.md plugins/prompt-tools/commands/prompt-reviewer.md
	@cp .claude/commands/promptify.md plugins/prompt-tools/commands/promptify.md
	@echo
	@echo "Syncing hooks to plugins/security-hooks/..."
	@mkdir -p plugins/security-hooks/hooks
	@cp .claude/hooks/*.sh plugins/security-hooks/hooks/
	@cp .claude/security-patterns.conf plugins/security-hooks/
	@chmod +x plugins/security-hooks/hooks/*.sh
	@echo
	@echo -e "$(GREEN)✅ Plugin sync complete!$(NC)"

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
	@ls -la "$(PROTECT_BRANCH_SCRIPT)" 2>/dev/null || echo -e "  $(RED)❌ Protected branch script not found$(NC)"
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
