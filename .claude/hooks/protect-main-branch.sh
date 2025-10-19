#!/usr/bin/env bash

set -euo pipefail

LOG_FILE="${WORKSPACE:-.}/.claude/protect-main-branch.log"

log_debug() {
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] DEBUG: $*" >> "$LOG_FILE"
}

log_info() {
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] INFO: $*" >> "$LOG_FILE"
}

log_warn() {
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] WARN: $*" >> "$LOG_FILE"
}

log_error() {
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] ERROR: $*" >> "$LOG_FILE"
}

log_debug "Hook script started"
log_debug "Working directory: $(pwd)"

STDIN_CONTENT=$(cat)
log_debug "Received stdin: $STDIN_CONTENT"

TOOL_NAME=$(echo "$STDIN_CONTENT" | jq -r '.tool // empty')
log_debug "Tool name: $TOOL_NAME"

if [[ -z "$TOOL_NAME" ]]; then
    log_error "No tool name found in stdin"
    exit 0
fi

PROTECTED_TOOLS_PATTERN="^(Edit|Write|Bash|Task)$"
if [[ ! "$TOOL_NAME" =~ $PROTECTED_TOOLS_PATTERN ]]; then
    log_debug "Tool '$TOOL_NAME' is not a protected tool, skipping"
    exit 0
fi

PROTECTED_BRANCHES=("main" "master" "production" "release")
PROTECTED_PATTERN="^(main|master|production|release)$"

if [[ -n "${TEST_BRANCH_NAME:-}" ]]; then
    CURRENT_BRANCH="$TEST_BRANCH_NAME"
    log_debug "Using test branch name: $CURRENT_BRANCH"
else
    CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "")
    log_debug "Current branch: $CURRENT_BRANCH"
fi

if [[ -z "$CURRENT_BRANCH" ]]; then
    log_warn "Could not determine current branch"
    exit 0
fi

is_protected_branch() {
    local branch="$1"
    if [[ "$branch" =~ $PROTECTED_PATTERN ]]; then
        return 0
    fi
    return 1
}

if is_protected_branch "$CURRENT_BRANCH"; then
    log_error "PROTECTED BRANCH VIOLATION: Attempt to use tool '$TOOL_NAME' on protected branch '$CURRENT_BRANCH'"

    cat <<EOF
{
  "blocked": true,
  "reason": "Direct edits to protected branch '$CURRENT_BRANCH' are not allowed.

Protected branches enforce PR-based workflows to ensure:
- Code review and quality standards
- CI/CD pipeline validation
- Collaboration and knowledge sharing
- Audit trail for all changes

To proceed:
1. Create a feature branch:
   git checkout -b your-name/feature-description

2. Make your changes on the feature branch

3. Push and create a Pull Request:
   git push -u origin your-name/feature-description

Protected branches: ${PROTECTED_BRANCHES[*]}

For emergency changes, use the override environment variable:
ALLOW_PROTECTED_BRANCH_EDIT=1

This hook ensures repository safety and collaboration best practices."
}
EOF
    exit 1
fi

log_info "Branch '$CURRENT_BRANCH' is not protected, allowing tool execution"
exit 0
