---
name: pr
description: Creates a draft GitHub pull request with a plain-text conventional-commit title and a succinct description of the branch. Use when the user asks to open, raise, or create a pull request or PR.
user-invocable: true
argument-hint: "[--base <branch>]"
allowed-tools: Bash(git add:*), Bash(git status:*), Bash(git commit:*), Bash(git log:*), Bash(git diff:*), Bash(git push:*), Bash(gh *), Read, Write
---

# PR - GitHub Pull Request Creation

Creates a draft pull request with a conventional-commit title and a short
description of what the branch actually changes.

## Workflow

1. **Validate**: `gh auth status`, current branch, uncommitted changes.
2. **Read the branch**: `git log <base>..HEAD` and `git diff <base>...HEAD` —
   the description covers exactly this, nothing more.
3. **Write the body**: use `.github/pull_request_template.md` if present,
   otherwise the sections below.
4. **Create**: `gh pr create --draft --title "<type>(<scope>): <subject>" --body-file <file> --base main`

## Title Rules

- Conventional commit format: `<type>(<scope>): <subject>`.
- **No emoji.** The title is plain text.
- Under 70 characters, imperative mood, no trailing period.
- For a single-commit PR, reuse that commit's subject.

Examples:

- `feat(auth): add session refresh endpoint`
- `fix(scanner): match patterns case-insensitively`
- `docs: document the scanner exit codes`

## Body Rules

Keep it succinct. The PR body describes the commits on the branch and nothing
else.

- **Summary**: one to three sentences. What changed and why. No restating the
  title, no background essay.
- **Changes**: one bullet per meaningful change, one line each. Group trivial
  edits rather than listing every file. Under six bullets for a normal PR — if
  it needs more, the PR is probably too large.
- **Testing**: one line stating what was run (`make test`, `go test ./...`) and
  the result. Say so plainly if nothing was run.
- Drop any template section that has nothing real to say.
- No emoji, no headings beyond the template's, no "Notes for reviewers",
  "Future work", "Impact", or risk-assessment sections unless asked.
- Do not claim behaviour you have not verified.

Default body when there is no template:

```markdown
## Summary

<1-3 sentences>

## Changes

- <change>
- <change>

## Testing

<command and result>
```

Example body:

```markdown
## Summary

Pattern matching in the MCP scanner was case-sensitive, so uppercase AWS keys
passed the scan. Matching is now case-insensitive.

## Changes

- Add `-i` to the pattern grep in `mcp-security-scanner.sh`
- Cover uppercase keys in `tests/test-scanner.sh`

## Testing

`make test` — all suites pass.
```

## Common Actions

```bash
gh pr ready <PR-NUMBER>                                    # draft to ready
gh pr edit <PR-NUMBER> --add-reviewer user1,user2          # add reviewers
gh pr status                                               # check status
```

## Error Handling

- Missing `gh`: prompt to install.
- Not authenticated: run `gh auth status` and prompt to log in.
- Missing PR template: use the default body above; do not create the template
  file unless asked.
