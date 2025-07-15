# Claude Code Template

A template for Claude Code projects with utilities, hooks, and commands.

## Hooks

| Hook | Description |
|------|-------------|
| [MCP Security Scanner](docs/mcp-security-scanner.md) | Prevents sensitive data from being sent to MCP services |

## Commands

```bash
make install        # Install hooks to current project
make test           # Run complete test suite
make test-unit      # Run unit tests only
make test-integration # Run integration tests only
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
