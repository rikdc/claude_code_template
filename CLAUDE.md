# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Claude Code Template project that provides security hooks and testing infrastructure for Claude Code projects. The main purposes are:
1. Prevent sensitive data from being sent to external MCP services through security scanning
2. Enforce PR-based workflows by preventing direct edits to protected branches (main, master, production, release)

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
- **Protected Branch Hook** (`.claude/hooks/protect-main-branch.sh`): Prevents direct edits to protected branches (main, master, production, release) by blocking Edit, Write, Bash, and Task tools, enforcing PR-based workflows
- **Hook Configuration** (`.claude/settings.json`): Claude Code hook configuration that automatically runs hooks for MCP tools and code-modifying operations
- **Security Patterns** (`.claude/security-patterns.conf`): Configurable regex patterns for detecting sensitive data (API keys, tokens, database URLs, etc.)
- **Custom Commands** (`.claude/commands/`): Slash commands that provide specialized workflows for code quality, cleanup, and changelog maintenance

### Testing Infrastructure

The project uses a simple, focused testing approach:

- **Security Scanner Tests** (`tests/test-scanner.sh`): Validates core security scanner functionality including sensitive data detection and MCP tool filtering
- **Protected Branch Tests** (`tests/test-protect-main-branch.sh`): Validates branch protection logic, tool blocking, and feature branch allowance
- **Fast Execution**: Tests run in seconds and focus on essential functionality
- **Test Coverage**: Both hooks have comprehensive test suites ensuring correct behavior

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

The project uses Claude Code's PreToolUse hook mechanism with two complementary hooks:

### Security Scanner Hook
- **Trigger**: All MCP tool calls matching pattern `mcp__.*`
- **Flow**: Request → Security Scanner → Pattern Detection → Logging/Blocking
- **Scope**: Monitors all MCP servers (Sequential, Context7, Magic, Playwright, etc.)
- **Response**: Provides monitoring/auditing with violation logging

### Protected Branch Hook
- **Trigger**: Edit, Write, Bash, and Task tool calls matching pattern `(Edit|Write|Bash|Task).*`
- **Flow**: Tool Call → Branch Detection → Protection Check → Allow/Block
- **Protected Branches**: main, master, production, release
- **Response**: Blocks operations with guidance to create feature branches

**Restart Requirement**: Hook configuration changes require Claude Code restart

## External Dependencies

**Required tools**: `jq`, `grep`, `awk`, `mktemp`

**Optional security tools** (enhance detection capabilities):

- `trufflehog`: Advanced secret scanning
- `gitleaks`: Git secret detection  
- `git-secrets`: Prevent secrets in commits

## Key Files

- `.claude/settings.json`: Hook configuration for all hooks (MCP scanner and branch protection)
- `.claude/security-patterns.conf`: Security detection patterns
- `.claude/security-scan.log`: Audit log of all security scan activity
- `.claude/protect-main-branch.log`: Audit log of branch protection activity
- `.claude/commands/`: Custom slash commands (check, clean, changelog)
- `tests/test-scanner.sh`: Security scanner test suite
- `tests/test-protect-main-branch.sh`: Protected branch hook test suite
- `Makefile`: Primary interface for all project operations
- `docs/mcp-security-scanner.md`: Comprehensive security scanner documentation
- `docs/protect-main-branch-hook.md`: Comprehensive branch protection documentation

## Monitoring and Debugging

**Security Scanning**: Check `.claude/security-scan.log` for:
- Hook execution: `grep "DEBUG: Hook script started"`
- Clean requests: `grep "Security scan passed"`
- Violations: `grep "SECURITY VIOLATION"`

**Branch Protection**: Check `.claude/protect-main-branch.log` for:
- Hook execution: `grep "DEBUG: Hook script started"`
- Allowed operations: `grep "INFO: Branch.*is not protected"`
- Violations: `grep "ERROR: PROTECTED BRANCH VIOLATION"`
- Blocked tools: `grep "blocked" .claude/protect-main-branch.log`

**MCP Coverage**: The scanner monitors all MCP servers and extracts content from various parameter fields (`thought`, `prompt`, `topic`, `libraryName`, etc.) for comprehensive security coverage.

**Branch Protection Coverage**: The hook monitors Edit, Write, Bash, and Task tools on protected branches (main, master, production, release) while allowing Read-only operations.
