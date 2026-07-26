#!/usr/bin/env bash

set -euo pipefail

LOG_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}/.claude"
LOG_FILE="$LOG_DIR/protect-main-branch.log"

log() {
    local level="$1"
    shift
    [[ -d "$LOG_DIR" ]] || return 0
    echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] $level: $*" >> "$LOG_FILE" 2>/dev/null || true
}

log_debug() { log DEBUG "$@"; }
log_info()  { log INFO "$@"; }
log_warn()  { log WARN "$@"; }
log_error() { log ERROR "$@"; }

# PreToolUse hooks always exit 0. A deny is expressed through
# hookSpecificOutput.permissionDecision on stdout, not an exit code.
allow() {
    exit 0
}

deny() {
    local reason="$1"
    jq -n --arg reason "$reason" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: $reason
      }
    }'
    exit 0
}

log_debug "Hook script started"
log_debug "Working directory: $(pwd)"

STDIN_CONTENT=$(cat)

TOOL_NAME=$(echo "$STDIN_CONTENT" | jq -r '.tool_name // empty')
log_debug "Tool name: $TOOL_NAME"

if [[ -z "$TOOL_NAME" ]]; then
    log_error "No tool name found in stdin"
    allow
fi

PROTECTED_TOOLS_PATTERN="^(Edit|Write|Bash|Task)$"
if [[ ! "$TOOL_NAME" =~ $PROTECTED_TOOLS_PATTERN ]]; then
    log_debug "Tool '$TOOL_NAME' is not a protected tool, skipping"
    allow
fi

if [[ -n "${ALLOW_PROTECTED_BRANCH_EDIT:-}" ]]; then
    log_warn "ALLOW_PROTECTED_BRANCH_EDIT set, bypassing protection for '$TOOL_NAME'"
    allow
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
    allow
fi

if [[ ! "$CURRENT_BRANCH" =~ $PROTECTED_PATTERN ]]; then
    log_info "Branch '$CURRENT_BRANCH' is not protected, allowing tool execution"
    allow
fi

log_error "PROTECTED BRANCH VIOLATION: Attempt to use tool '$TOOL_NAME' on protected branch '$CURRENT_BRANCH'"

deny "Direct edits to protected branch '$CURRENT_BRANCH' are not allowed.

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
