---
description: Triage PR review comments — classify, accept/reject, and create follow-up tasks
allowed-tools: Read, Glob, Grep, Bash(gh pr:*), Bash(gh api:*), Bash(git diff:*), Bash(git log:*), Bash(git status:*), Task
---

# /triage — PR Review Comment Triage

Triage all unresolved review comments on a pull request. Classifies each comment, decides whether to accept or reject, and takes action.

## Usage

```bash
/triage                           # Triage current branch's PR
/triage 123                       # Triage PR #123
/triage --dry-run                 # Preview without acting
/triage --no-resolve              # Skip resolving handled threads
```

## Workflow

Execute the triage-reviews skill with the provided arguments.

Use the `triage-reviews` skill to process the PR review comments. Pass through all arguments from `$ARGUMENTS`.

### Steps

1. **Identify the target PR**
   - If `$ARGUMENTS` contains a number or URL, use that PR
   - Otherwise, run `gh pr view --json number,url,title` for the current branch

2. **Fetch review comments**
   - Get all review comments via `gh api repos/{owner}/{repo}/pulls/{number}/comments`
   - Get review summaries via `gh api repos/{owner}/{repo}/pulls/{number}/reviews`
   - Get the PR diff for code context via `gh pr diff`

3. **Classify each comment** into: security, correctness, error-handling, performance, design, testing, style, nit, praise, or question

4. **Evaluate each comment** against the actual code and project conventions:
   - **Accept**: Real issues, valid improvements aligned with project standards
   - **Reject**: Contradicts conventions, subjective preference, misunderstands intent
   - **Acknowledge**: Praise (thank), questions (answer)

5. **Present summary table** with proposed actions grouped by decision

6. **Wait for user confirmation** (unless `--dry-run`)

7. **Execute actions**:
   - Accepted: thumbsup reaction + create beads issue with context and PR comment link
   - Rejected: post respectful reply explaining why, citing conventions, then resolve the thread
   - Praise: post thank-you reply, then resolve the thread
   - Questions: post answer, then resolve the thread (unless uncertain)

8. **Report completion** with counts of actions taken

### Idempotency

Before any action, check if it was already performed (existing reactions, replies, or linked beads issues). Skip duplicates silently.
