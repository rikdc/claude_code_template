# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Claude Code Template project that provides security hooks and testing infrastructure for Claude Code projects. The main purpose is to prevent sensitive data from being sent to external MCP services through a security scanning hook.

## Key Commands

All project operations are managed through the Makefile:

```bash
make test           # Run complete test suite (unit + integration)
make test-unit      # Run unit tests only  
make test-integration # Run integration tests only
make lint           # Run ShellCheck on all shell scripts
make install        # Install hooks (make scanner script executable)
make clean          # Remove test artifacts and logs
make status         # Show current status and configuration
make check-tools    # Check for required and optional security tools
make help           # Display all available commands
```

## Architecture

### Core Components

- **Security Scanner Hook** (`.claude/hooks/mcp-security-scanner.sh`): Main security scanner that intercepts MCP requests and scans for sensitive data using regex patterns and optional external tools
- **Hook Configuration** (`.claude/settings.json`): Claude Code hook configuration that automatically runs the security scanner for all MCP tool calls (`mcp__*`)
- **Security Patterns** (`.claude/security-patterns.conf`): Configurable regex patterns for detecting sensitive data (API keys, tokens, database URLs, etc.)

### Testing Infrastructure

The project uses a simple, focused testing approach:

- **Simple Test Script** (`tests/test-scanner.sh`): Single test script that validates core security scanner functionality without complex infrastructure
- **Core Functionality Tests**: Tests clean content allowance, sensitive data detection, MCP tool filtering, and hook integration
- **Fast Execution**: Tests run in seconds and focus on essential security functionality

### Security Pattern Detection

The scanner detects multiple categories of sensitive data:
- Authentication & API keys (AWS, GitHub, OpenAI, Slack, Discord, etc.)
- Database connection strings (PostgreSQL, MySQL, MongoDB)
- Personal information (emails, credit cards, SSNs, phone numbers)
- Cloud provider credentials and private keys

## Development Workflow

1. **Testing**: Always run `make test` before committing changes
2. **Linting**: Use `make lint` to ensure shell script quality
3. **Pattern Updates**: Modify `.claude/security-patterns.conf` to add custom security patterns
4. **Hook Testing**: Test security patterns with sample MCP requests as shown in the documentation

## External Dependencies

**Required tools**: `jq`, `grep`, `awk`, `mktemp`

**Optional security tools** (enhance detection capabilities):
- `trufflehog`: Advanced secret scanning
- `gitleaks`: Git secret detection  
- `git-secrets`: Prevent secrets in commits

## Key Files

- `.claude/settings.json`: Hook configuration for all MCP tools
- `.claude/security-patterns.conf`: Security detection patterns
- `.claude/security-scan.log`: Audit log of all security scan activity
- `tests/test-scanner.sh`: Simple test script for security scanner validation