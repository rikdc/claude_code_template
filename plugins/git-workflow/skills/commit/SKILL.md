---
name: commit
description: Writes brief conventional-commit messages and commits the staged changes. Use when the user asks to commit, stage and commit, or write a commit message.
user-invocable: true
argument-hint: "[--no-verify] [--help]"
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(git diff:*), Bash(git log:*)
---

# Commit - Conventional Commit Messages

Creates commits with short conventional commit messages. Runs pre-commit checks
and suggests splitting when the staged changes cover more than one concern.

## Usage

```bash
/git-workflow:commit [--no-verify] [--help]
```

## Workflow

1. **Pre-commit checks** (unless `--no-verify`): run `make checks` if available.
2. **Git analysis**: inspect staged files and `git diff` to identify the change.
3. **Commit**: write the message, then commit.

If nothing is staged, stage all modified files. If the diff covers distinct
concerns, propose atomic commits before committing.

## Message Rules

```
<type>(<scope>): <subject>

[body — only when required]
```

- **No emoji.** Not in the subject, not in the body.
- Subject: imperative mood, lower case, no trailing period, under 60 characters.
- `scope` is optional; include it only when it disambiguates.
- **Default to a subject line alone.** Most commits need nothing more.
- Add a body only for information the diff cannot show: why the change was
  made, a breaking change, or an issue reference. Cap it at two short sentences
  or three bullets.
- Never restate the diff, list touched files, summarise your process, or add
  closing commentary.

## Types

- `feat`: new feature
- `fix`: bug fix, security fix, warning fix
- `docs`: documentation
- `style`: formatting
- `refactor`: restructuring without behaviour change
- `perf`: performance
- `test`: tests
- `chore`: tooling, dependencies, configuration
- `ci`: CI/CD

## Splitting Criteria

Propose separate commits when the staged changes mix concerns, mix types
(feature + fix + docs), or are large enough that one message cannot describe
them accurately.

## Examples

Good:

- `feat(auth): add session refresh endpoint`
- `fix(render): release texture handles on unmount`
- `docs: document the scanner exit codes`
- `refactor: collapse duplicate error wrapping`

With a body, where the reason is not visible in the diff:

```
fix(scanner): match patterns case-insensitively

Uppercase AWS keys were slipping through, reported in #212.
```

Too verbose — do not do this:

```
✨ feat(auth): add a new session refresh endpoint to the auth package

This commit adds a new endpoint. It modifies auth/handler.go to add the
handler, auth/routes.go to register the route, and auth/handler_test.go to
cover it. This improves the developer experience and makes the codebase
more maintainable going forward.
```

## Options

- `--no-verify`: skip pre-commit checks
- `--help`: show this reference

## Notes

- **Pre-commit failures**: ask whether to fix or proceed.
- **No staged files**: auto-stage modified and new files.
- **Large changes**: suggest atomic commits with guided staging.
