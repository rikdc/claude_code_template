# Claude Code Template

Production-ready Claude Code project template with security hooks, workflow automation, and specialized agents for software development.

## Features

- **Security Hooks**: Prevent sensitive data leaks and enforce branch protection
- **Slash Commands**: Workflow automation for code quality, documentation, and project management
- **Specialized Agents**: Expert agents for Go development, code review, and documentation
- **Testing Infrastructure**: Comprehensive test suite with ~2000 lines of test coverage
- **Marketplace Plugins**: Modular plugin architecture with 6 installable plugins

## Installation

### Install Individual Plugins

The repository is organized as a marketplace with 6 independent plugins:

```bash
# Install specific plugins
claude code plugins install security-hooks
claude code plugins install dev-skills
claude code plugins install git-workflow
claude code plugins install pm-tools
claude code plugins install code-quality
claude code plugins install prompt-tools

# Or install from GitHub
claude code plugins install github:rikdc/claude_code_template/security-hooks
```

### Clone for Local Development

Clone the entire repository for local development with all components:

```bash
git clone https://github.com/rikdc/claude_code_template.git
cd claude_code_template
make install
make test
```

## Available Plugins

| Plugin | Category | Description |
|--------|----------|-------------|
| [security-hooks](plugins/security-hooks/) | Security | MCP security scanner and protected branch hooks |
| [dev-skills](plugins/dev-skills/) | Development | Expert skills for Go, documentation, and architecture |
| [git-workflow](plugins/git-workflow/) | Productivity | Git automation for commits, PRs, and changelog |
| [pm-tools](plugins/pm-tools/) | Productivity | Project management commands for PRDs and tasks |
| [code-quality](plugins/code-quality/) | Development | Code quality analysis, review, and cleanup |
| [prompt-tools](plugins/prompt-tools/) | AI | AI prompt generation and review tools |

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
make sync-plugins   # Sync .claude/ changes to plugins/ for distribution
```

## Project Structure

```text
plugins/                          # Marketplace plugin distribution
├── security-hooks/              # MCP security and branch protection
├── dev-skills/                  # Development skills (Go, docs, etc.)
├── git-workflow/                # Git automation commands
├── pm-tools/                    # Project management commands
├── code-quality/                # Code review and cleanup
└── prompt-tools/                # AI prompt generation

.claude/                         # Local development structure
├── skills/                      # Specialized skills
├── commands/                    # Slash command implementations
│   ├── dev/                    # Development workflow commands
│   ├── gh/                     # GitHub automation commands
│   └── pm/                     # Project management commands
├── hooks/                       # Security and workflow hooks
└── settings.json                # Hook configuration

docs/                            # Documentation
tests/                           # Test suite
Makefile                         # Command interface
```

**Note**: The `plugins/` directory is for marketplace distribution. The `.claude/` directory is used for local development. Use `make sync-plugins` to sync changes from `.claude/` to `plugins/`.

## Attribution

This project includes work from the following sources:

### Product Management Commands

The PM command suite (`/pm:create-prd`, `/pm:generate-tasks`, `/pm:process-tasks`) is based on patterns and workflows from:

- **Source**: [AI Dev Tasks](https://github.com/snarktank/ai-dev-tasks/tree/main)
- **License**: MIT License
- **Usage**: Adapted and extended for Claude Code project management workflows

Special thanks to the AI Dev Tasks project for providing the foundational patterns for structured AI-assisted development workflows.
