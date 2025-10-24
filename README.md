# Claude Code Template

Production-ready Claude Code project template with security hooks, workflow automation, and specialized agents for software development.

## Features

- **Security Hooks**: Prevent sensitive data leaks and enforce branch protection
- **Slash Commands**: Workflow automation for code quality, documentation, and project management
- **Specialized Agents**: Expert agents for Go development, code review, and documentation
- **Testing Infrastructure**: Comprehensive test suite with ~2000 lines of test coverage

## Security Hooks

| Hook | Description |
|------|-------------|
| [MCP Security Scanner](docs/mcp-security-scanner.md) | Scans MCP requests for API keys, tokens, database URLs, and PII |
| [Protected Branch Hook](docs/protect-main-branch-hook.md) | Enforces PR-based workflow by blocking direct edits to protected branches |

## Slash Commands

### Development
- `/dev:check` - Comprehensive code quality analysis with parallel auto-fixing
- `/dev:review` - Performs detailed code review on latest commit
- `/clean` - Removes redundant comments from codebase

### GitHub Workflow
- `/gh:commit` - Creates conventional commits with emoji
- `/gh:pr` - Creates GitHub pull requests with templates

### Project Management
- `/pm:create-prd` - Interactive PRD creation with clarifying questions
- `/pm:generate-tasks` - Generates task lists from PRDs with phased approach
- `/pm:process-tasks` - Task management protocols and workflows

### Documentation
- `/changelog` - Maintains CHANGELOG.md following Keep a Changelog format
- `/promptify` - Generates high-quality AI prompts with best practices
- `/prompt-reviewer` - Reviews and improves AI prompts

## Specialized Agents

- **Manager Agent**: Orchestrates multi-agent workflows and task delegation
- **Go Implementor**: Test-driven Go development with eevee framework patterns
- **Go Reviewer**: Expert Go code review focusing on idiomatic patterns
- **Document Agent**: Technical documentation and API specification generation
- **Staff Eyes**: Senior-level architectural review and design feedback
- **Specify Agent**: Requirements analysis and technical specification creation
- **Taskify Agent**: Breaks down epics into actionable development tasks

## Quick Start

```bash
make install        # Install hooks and make scripts executable
make test           # Run complete test suite
make lint           # Run ShellCheck and markdownlint
make check-tools    # Verify required and optional tools
make status         # Show configuration and tool status
```

## Project Structure

```
.claude/
├── agents/         # Specialized agent definitions
├── commands/       # Slash command implementations
│   ├── dev/       # Development workflow commands
│   ├── gh/        # GitHub automation commands
│   └── pm/        # Project management commands
├── hooks/         # Security and workflow hooks
└── settings.json  # Hook configuration

docs/              # Documentation
tests/             # Test suite
scripts/           # Utility scripts
Makefile           # Command interface
```

## Attribution

This project includes work from the following sources:

### Product Management Commands

The PM command suite (`/pm:create-prd`, `/pm:generate-tasks`, `/pm:process-tasks`) is based on patterns and workflows from:

- **Source**: [AI Dev Tasks](https://github.com/snarktank/ai-dev-tasks/tree/main)
- **License**: MIT License
- **Usage**: Adapted and extended for Claude Code project management workflows

Special thanks to the AI Dev Tasks project for providing the foundational patterns for structured AI-assisted development workflows.
