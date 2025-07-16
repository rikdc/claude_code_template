# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Claude Code Template project that provides security hooks and testing infrastructure for Claude Code projects. The main purpose is to prevent sensitive data from being sent to external MCP services through a security scanning hook.

## Key Commands

All project operations are managed through the Makefile:

```bash
make test           # Run complete test suite
make lint           # Run ShellCheck on all shell scripts
make install        # Install hooks (make scanner script executable)
make clean          # Remove test artifacts and logs
make status         # Show current status and configuration
make check-tools    # Check for required and optional security tools
make help           # Display all available commands
```

## Claude Code Slash Commands

This project includes custom Claude Code slash commands in `.claude/commands/`:

- `/check` - Comprehensive code quality analysis and auto-fix with parallel sub-task strategy
- `/clean` - Remove redundant and obvious comments from codebase  
- `/changelog` - Create and maintain project changelog following Keep a Changelog format

## Architecture

### Core Components

- **Security Scanner Hook** (`.claude/hooks/mcp-security-scanner.sh`): Main security scanner that intercepts MCP requests and scans for sensitive data using regex patterns and optional external tools
- **Hook Configuration** (`.claude/settings.json`): Claude Code hook configuration that automatically runs the security scanner for all MCP tool calls (`mcp__*`)
- **Security Patterns** (`.claude/security-patterns.conf`): Configurable regex patterns for detecting sensitive data (API keys, tokens, database URLs, etc.)
- **Custom Commands** (`.claude/commands/`): Slash commands that provide specialized workflows for code quality, cleanup, and changelog maintenance

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
5. **Quality Checks**: Use `/check` command for comprehensive code quality analysis and auto-fixing
6. **Code Cleanup**: Use `/clean` command to remove redundant comments before commits
7. **Changelog**: Use `/changelog` command to maintain project changelog with proper versioning

## Hook System Architecture

The security system operates through Claude Code's PreToolUse hook mechanism:

- **Hook Trigger**: All MCP tool calls matching pattern `mcp__.*` 
- **Execution Flow**: Request → Security Scanner → Pattern Detection → Logging/Blocking
- **Scope**: Monitors all MCP servers (Sequential, Context7, Magic, Playwright, etc.)
- **Response**: Currently provides monitoring/auditing with violation logging
- **Restart Requirement**: Hook configuration changes require Claude Code restart

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
- `.claude/commands/`: Custom slash commands (check, clean, changelog)
- `tests/test-scanner.sh`: Simple test script for security scanner validation
- `Makefile`: Primary interface for all project operations
- `docs/mcp-security-scanner.md`: Comprehensive security scanner documentation

## Monitoring and Debugging

**Security Scanning**: Check `.claude/security-scan.log` for:

- Hook execution: `grep "DEBUG: Hook script started"`
- Clean requests: `grep "Security scan passed"`  
- Violations: `grep "SECURITY VIOLATION"`

**MCP Coverage**: The scanner monitors all MCP servers and extracts content from various parameter fields (`thought`, `prompt`, `topic`, `libraryName`, etc.) for comprehensive security coverage.
