---
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(gh *)
description: Creates a PR on GitHub
---

# GitHub PR Creation Assistant

Creates a new GitHub pull request using our project template and conventions.

## Usage

This command will:

1. Check git status and GitHub CLI authentication
2. Retrieve and populate the PR template
3. Format title using conventional commits
4. Create draft PR with proper structure

## Command Structure

```bash
gh pr create --draft --title "✨(scope): Your descriptive title" --body-file .github/pull_request_template.md --base main
```

### Title Format

Use conventional commit format with emojis:

✨(feature): Add new functionality
🐛(bugfix): Fix critical issue
📝(docs): Update documentation
🔧(config): Modify configuration

### Workflow Steps

Sub-task 1: Validation

- Verify GitHub CLI installation and authentication
- Check current branch status
- Validate uncommitted changes

### Sub-task 2: Template Processing

- Read .github/pull_request_template.md
- Pre-populate known sections
- Validate template structure

### Sub-task 3: PR Creation

- Format title with appropriate emoji
- Create draft PR with populated template
- Set proper base branch and reviewers

### Template Sections

Ensure all sections are included:

```markdown
# Summary

## pr_agent:summary

## Changes

## pr_agent:walkthrough

## Testing

## Checklist
```

Common Actions

```bash
# Convert draft to ready
gh pr ready <PR-NUMBER>

# Add reviewers
gh pr edit <PR-NUMBER> --add-reviewer username1,username2

# Check PR status
gh pr status
```

### Error Handling

The command will automatically:

- Install GitHub CLI if missing (with permission)
- Prompt for authentication if needed
- Create template if .github/pull_request_template.md doesn't exist
- Validate title format and suggest corrections

**Prerequisites Check:***

Before execution, validates:

- GitHub CLI installation
- Repository authentication
- Clean working directory
- Valid branch for PR creation

Use `gh auth status` to verify authentication if issues occur.
