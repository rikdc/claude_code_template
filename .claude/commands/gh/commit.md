---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(git diff:*), Bash(git log:*)
description: Creates well-formatted commits with conventional commit messages and emoji
---

# Claude Command: Commit

Creates well-formatted commits with conventional commit messages and emoji.
Runs pre-commit checks and suggests commit splitting when appropriate.

## Usage

To create a commit, just type:

```bash
/commit [--no-verify] [--no-emoji] [--help]
```

## Workflow

This command spawns parallel sub-tasks for efficiency:

1. **Pre-commit checks** (unless `--no-verify`): Runs `make checks` if available.
2. **Git analysis**: Checks staged files, runs `git diff`, analyzes change patterns
3. **Commit preparation**: Generates conventional commit messages with appropriate emoji

If no files are staged, automatically stages all modified files.
If multiple distinct changes are detected, suggests splitting into atomic commits.

## Core Commit Types & Emoji

- ✨ `feat`: New feature
- 🐛 `fix`: Bug fix
- 📝 `docs`: Documentation
- 💄 `style`: Formatting/style
- ♻️ `refactor`: Code refactoring
- ⚡️ `perf`: Performance improvements
- ✅ `test`: Tests
- 🔧 `chore`: Tooling, configuration
- 🚀 `ci`: CI/CD improvements
- 🚨 `fix`: Fix compiler/linter warnings
- 🔒️ `fix`: Security fixes
- 🩹 `fix`: Simple non-critical fixes

*Type `/commit --help` for extended emoji reference*

## Commit Message Format

```markdown
<emoji> <type>: <description>
[optional body]
```

- Use present tense, imperative mood ("add feature" not "added feature")
- Keep first line under 72 characters
- Each commit should contain related changes serving a single purpose

## Splitting Criteria

Suggests multiple commits when changes involve:

- Different concerns or file types
- Mixed change types (feature + fix + docs)
- Large changes that would be clearer when separated

## Examples

Good commit messages:

- ✨ feat: add user authentication system
- 🐛 fix: resolve memory leak in rendering process
- 📝 docs: update API documentation with new endpoints
- ♻️ refactor: simplify error handling logic

Example split:

- ✨ feat: add new API endpoints
- 📝 docs: update API documentation
- ✅ test: add unit tests for new endpoints

## Options

- `--no-verify`: Skip pre-commit checks
- `--no-emoji`: Omit emoji from commit messages
- `--help`: Show extended emoji reference and examples

## Notes

- **Pre-commit failures**: Prompts to fix issues or proceed anyway
- **No staged files**: Auto-stages all modified/new files
- **Large changes**: Suggests atomic commits with guided staging
