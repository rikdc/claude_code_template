# Git Workflow Plugin

Git workflow automation commands for commits, PRs, and changelog management.

## Commands Included

### `/commit`

Creates well-formatted commits with conventional commit messages and emoji.

**Features**:

- Conventional commit format (feat, fix, docs, etc.)
- Optional emoji prefixes
- Automatic commit message generation from staged changes
- Follows repository commit message style

### `/pr`

Creates a PR on GitHub with proper title and description.

**Features**:

- Analyzes full commit history for the branch
- Generates concise PR title (under 70 characters)
- Creates detailed PR description with summary and test plan
- Automatically pushes branch if needed

### `/changelog`

Maintains a CHANGELOG.md document following Keep a Changelog format.

**Features**:

- Semantic versioning
- Categorized changes (Added, Changed, Fixed, etc.)
- Automatic date formatting
- Preserves existing changelog structure

## Installation

Install via Claude Code marketplace:

```bash
claude code plugins install git-workflow
```

Or install from this repository:

```bash
claude code plugins install github:rikdc/claude_code_template/git-workflow
```

## Usage Examples

```bash
# Create a commit
/commit

# Create a commit with specific message
/commit -m "feat: add user authentication"

# Create a pull request
/pr

# Update changelog
/changelog Add new authentication feature to version 1.2.0
```

## License

MIT
