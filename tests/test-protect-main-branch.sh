#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOK_SCRIPT="$PROJECT_ROOT/plugins/security-hooks/hooks/protect-main-branch.sh"

TESTS_PASSED=0
TESTS_FAILED=0

test_passed() {
    echo "✓ $1"
    ((TESTS_PASSED++)) || true
}

test_failed() {
    echo "✗ $1"
    ((TESTS_FAILED++)) || true
}

run_test() {
    local test_name="$1"
    local branch="$2"
    local tool="$3"
    local should_block="$4"

    local tmpfile
    tmpfile=$(mktemp)

    cat > "$tmpfile" <<EOF
{
  "hook_event_name": "PreToolUse",
  "tool_name": "$tool",
  "tool_input": {}
}
EOF

    local output
    local exit_code=0

    if output=$(TEST_BRANCH_NAME="$branch" CLAUDE_PROJECT_DIR="$PROJECT_ROOT" \
        bash "$HOOK_SCRIPT" < "$tmpfile" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi

    rm -f "$tmpfile"

    # PreToolUse always exits 0; the decision lives in the JSON on stdout.
    if [[ $exit_code -ne 0 ]]; then
        test_failed "$test_name (hook exited $exit_code, expected 0)"
        return
    fi

    local decision=""
    if [[ -n "$output" ]]; then
        decision=$(echo "$output" | jq -r '.hookSpecificOutput.permissionDecision // empty' 2>/dev/null || echo "")
    fi

    if [[ "$should_block" == "true" ]]; then
        if [[ "$decision" == "deny" ]]; then
            test_passed "$test_name"
        else
            test_failed "$test_name (expected deny, got '${decision:-no decision}')"
        fi
    else
        if [[ -z "$decision" ]]; then
            test_passed "$test_name"
        else
            test_failed "$test_name (expected allow, got '$decision')"
        fi
    fi
}

echo "Running Protected Branch Hook Tests..."
echo "========================================"

run_test "Block Edit on main branch" "main" "Edit" "true"
run_test "Block Write on main branch" "main" "Write" "true"
run_test "Block Bash on main branch" "main" "Bash" "true"
run_test "Block Task on main branch" "main" "Task" "true"

run_test "Block Edit on master branch" "master" "Edit" "true"
run_test "Block Edit on production branch" "production" "Edit" "true"
run_test "Block Edit on release branch" "release" "Edit" "true"

run_test "Allow Edit on feature branch" "feature/test" "Edit" "false"
run_test "Allow Write on feature branch" "alice/PROJ-123-feature" "Write" "false"
run_test "Allow Bash on feature branch" "bob/fix-bug" "Bash" "false"
run_test "Allow Task on feature branch" "carol/test-workflow" "Task" "false"

run_test "Allow Read on main branch" "main" "Read" "false"
run_test "Allow Grep on main branch" "main" "Grep" "false"

# The deny payload must be parseable JSON: the previous heredoc embedded raw
# newlines in a string literal, which no JSON parser accepts.
deny_output=$(echo '{"tool_name":"Edit","tool_input":{}}' |
    TEST_BRANCH_NAME="main" CLAUDE_PROJECT_DIR="$PROJECT_ROOT" bash "$HOOK_SCRIPT")
if echo "$deny_output" | jq -e '.hookSpecificOutput.permissionDecisionReason' >/dev/null 2>&1; then
    test_passed "Deny payload is valid JSON"
else
    test_failed "Deny payload is valid JSON"
fi

override_output=$(echo '{"tool_name":"Edit","tool_input":{}}' |
    TEST_BRANCH_NAME="main" CLAUDE_PROJECT_DIR="$PROJECT_ROOT" \
    ALLOW_PROTECTED_BRANCH_EDIT=1 bash "$HOOK_SCRIPT")
if [[ -z "$override_output" ]]; then
    test_passed "ALLOW_PROTECTED_BRANCH_EDIT bypasses protection"
else
    test_failed "ALLOW_PROTECTED_BRANCH_EDIT bypasses protection"
fi

echo ""
echo "========================================"
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $TESTS_FAILED"

if [[ $TESTS_FAILED -gt 0 ]]; then
    exit 1
fi

exit 0
