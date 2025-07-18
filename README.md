# Claude Code Template

A template for Claude Code projects with utilities, hooks, and commands.

## Hooks

| Hook | Description |
|------|-------------|
| [MCP Security Scanner](docs/mcp-security-scanner.md) | Prevents sensitive data from being sent to MCP services |

## Slash Commands

| Command | Description |
|---------|-------------|
| `/changelog` | Maintains a CHANGELOG.md document in this repository |
| `/check` | Fix all code quality issues using parallel sub-tasks |
| `/clean` | Removes useless comments from the code |
| `/gh:commit` | Creates well-formatted commits with conventional commit messages and emoji |
| `/gh:pr` | Creates a PR on GitHub |
| `/pm:create-prd` | Interactive PRD creation workflow with clarifying questions |
| `/pm:generate-tasks` | Automatic task list generation from PRDs with phased approach |
| `/pm:process-tasks` | Task management protocols and completion workflows |
| `/review-code` | Performs comprehensive code review on the latest commit |

## Commands

```bash
make install        # Install hooks to current project
make test           # Run complete test suite
make lint           # Run ShellCheck on all scripts
make clean          # Remove test artifacts
make status         # Show current status
make check-tools    # Check tool availability
make help           # Show all commands
```

## Project Structure

```bash
/
├── .claude/                    # Claude Code configuration
│   ├── hooks/                  # Claude Code hooks
│   │   └── mcp-security-scanner.sh
│   ├── settings.json           # Hook configuration
│   └── security-patterns.conf.example
├── docs/                       # Documentation
│   └── mcp-security-scanner.md
├── tests/                      # Test suite
├── scripts/                    # Utility scripts
└── Makefile                    # Command interface
```

## Attribution

This project includes work from the following sources:

### Product Management Commands
The PM command suite (`/pm:create-prd`, `/pm:generate-tasks`, `/pm:process-tasks`) is based on patterns and workflows from:
- **Source**: [AI Dev Tasks](https://github.com/snarktank/ai-dev-tasks/tree/main)
- **License**: MIT License
- **Usage**: Adapted and extended for Claude Code project management workflows

Special thanks to the AI Dev Tasks project for providing the foundational patterns for structured AI-assisted development workflows.
