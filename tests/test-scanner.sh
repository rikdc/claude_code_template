#!/usr/bin/env bash

# Simple Functional Tests for MCP Security Scanner

set -euo pipefail

SCANNER="./.claude/hooks/mcp-security-scanner.sh"
PASSED=0
FAILED=0

echo "🧪 MCP Security Scanner - Simple Functional Tests"
echo "=================================================="

# Test function using portable arithmetic
test_basic() {
    local name="$1"
    local input="$2"
    local expected="$3"

    echo -n "$name... "

    # Hide external tools to avoid hanging, keep essential tools
    local temp_path="/tmp/test-path-$$"
    mkdir -p "$temp_path"

    # Copy essential tools to temp path
    for tool in jq grep awk mktemp; do
        if command -v "$tool" >/dev/null 2>&1; then
            ln -s "$(command -v "$tool")" "$temp_path/$tool" 2>/dev/null || true
        fi
    done

    set +e
    PATH="$temp_path" echo "$input" | "$SCANNER" >/dev/null 2>&1
    local result=$?
    set -e

    rm -rf "$temp_path"

    if [[ "$result" -eq "$expected" ]]; then
        echo "PASS"
        PASSED=$((PASSED + 1))
    else
        echo "FAIL (got $result, expected $expected)"
        FAILED=$((FAILED + 1))
    fi
}

echo
echo "🚀 Running core tests (patterns only)..."

test_basic "Clean content" \
    '{"hook_event_name":"PreToolUse","tool_name":"mcp__context7__get-library-docs","tool_input":{"prompt":"How to use React?"}}' \
    0

test_basic "API key detection" \
    '{"hook_event_name":"PreToolUse","tool_name":"mcp__context7__get-library-docs","tool_input":{"prompt":"sk-1234567890123456789012345678901234567890"}}' \
    2

test_basic "Database URL" \
    '{"hook_event_name":"PreToolUse","tool_name":"mcp__context7__get-library-docs","tool_input":{"prompt":"postgresql://user:pass@host/db"}}' \
    2

test_basic "Non-MCP tool" \
    '{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{"prompt":"sk-1234567890123456789012345678901234567890"}}' \
    0

test_basic "Non-PreToolUse" \
    '{"hook_event_name":"PostToolUse","tool_name":"mcp__context7__get-library-docs","tool_input":{"prompt":"sk-test123"}}' \
    0

echo
total=$((PASSED + FAILED))
echo "📊 Tests: $total, Passed: $PASSED, Failed: $FAILED"

if [[ "$FAILED" -eq 0 ]]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ $FAILED test(s) failed!"
    exit 1
fi
