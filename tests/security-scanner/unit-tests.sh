#!/usr/bin/env bash

# Unit Tests for MCP Security Scanner
# Tests individual functions and pattern matching in isolation

set -euo pipefail

# Get script directory and load test framework
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/test-helpers.sh
source "$LIB_DIR/test-helpers.sh"

# Test configuration
readonly TEST_SUITE_NAME="Security Scanner Unit Tests"
readonly FIXTURES_DIR="$SCRIPT_DIR/fixtures"

# Test: Non-MCP tools should be allowed
test_non_mcp_tools() {
    local test_cases=(
        "Read"
        "Write" 
        "Edit"
        "Bash"
        "Grep"
        "Glob"
    )
    
    for tool in "${test_cases[@]}"; do
        local input
        input=$(create_mcp_request "$tool" "test content")
        test_hook_response "non_mcp_tool_$tool" "$input" 0 "Non-MCP tool $tool should be allowed"
    done
}

# Test: Non-PreToolUse events should be allowed
test_non_pretooluse_events() {
    local events=("PostToolUse" "Notification" "Stop")
    
    for event in "${events[@]}"; do
        local input='{"hook_event_name": "'$event'", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "API key: sk-test123456"}}'
        test_hook_response "non_pretooluse_$event" "$input" 0 "Non-PreToolUse event $event should be allowed"
    done
}

# Test: Clean content should be allowed
test_clean_content() {
    if [[ ! -f "$FIXTURES_DIR/clean-inputs.json" ]]; then
        log_error "Clean inputs fixture not found: $FIXTURES_DIR/clean-inputs.json"
        return 1
    fi
    
    # Use jq to parse test cases from fixture
    local test_cases
    test_cases=$(jq -r '.test_cases[] | @base64' "$FIXTURES_DIR/clean-inputs.json")
    
    while IFS= read -r case_data; do
        local case_json
        case_json=$(echo "$case_data" | base64 --decode)
        
        local name description hook_event tool_name tool_input
        name=$(echo "$case_json" | jq -r '.name')
        description=$(echo "$case_json" | jq -r '.description')
        hook_event=$(echo "$case_json" | jq -r '.hook_event_name')
        tool_name=$(echo "$case_json" | jq -r '.tool_name')
        tool_input=$(echo "$case_json" | jq -c '.tool_input')
        
        local full_input="{\"hook_event_name\": \"$hook_event\", \"tool_name\": \"$tool_name\", \"tool_input\": $tool_input}"
        
        test_hook_response "clean_content_$name" "$full_input" 0 "$description"
    done <<< "$test_cases"
}

# Test: Sensitive content should be blocked
test_sensitive_content() {
    if [[ ! -f "$FIXTURES_DIR/sensitive-inputs.json" ]]; then
        log_error "Sensitive inputs fixture not found: $FIXTURES_DIR/sensitive-inputs.json"
        return 1
    fi
    
    # Use jq to parse test cases from fixture
    local test_cases
    test_cases=$(jq -r '.test_cases[] | @base64' "$FIXTURES_DIR/sensitive-inputs.json")
    
    while IFS= read -r case_data; do
        local case_json
        case_json=$(echo "$case_data" | base64 --decode)
        
        local name description hook_event tool_name tool_input
        name=$(echo "$case_json" | jq -r '.name')
        description=$(echo "$case_json" | jq -r '.description')
        hook_event=$(echo "$case_json" | jq -r '.hook_event_name')
        tool_name=$(echo "$case_json" | jq -r '.tool_name')
        tool_input=$(echo "$case_json" | jq -c '.tool_input')
        
        local full_input="{\"hook_event_name\": \"$hook_event\", \"tool_name\": \"$tool_name\", \"tool_input\": $tool_input}"
        
        test_hook_response "sensitive_content_$name" "$full_input" 1 "$description"
    done <<< "$test_cases"
}

# Test: Pattern matching functionality
test_pattern_matching() {
    # Test individual patterns against known good/bad content
    local test_patterns=(
        "api.*key.*[=:]|API key should match configuration strings|API_KEY=secret123|true"
        "sk-[A-Za-z0-9]\{40,\}|OpenAI key should match|sk-test123456789abcdefghijklmnopqrstuvwxyz123456|true"
        "[0-9]\{3\}-[0-9]\{2\}-[0-9]\{4\}|SSN should match|123-45-6789|true"
        "postgres.*://|PostgreSQL URL should match|postgresql://user:pass@host/db|true"
        "-----BEGIN.*PRIVATE KEY-----|Private key should match|-----BEGIN RSA PRIVATE KEY-----|true"
        "api.*key.*[=:]|Clean text should not match|How to use React hooks?|false"
        "sk-[A-Za-z0-9]\{40,\}|Short string should not match|sk-short|false"
        "[0-9]\{3\}-[0-9]\{2\}-[0-9]\{4\}|Invalid SSN should not match|12-345-6789|false"
    )
    
    for pattern_test in "${test_patterns[@]}"; do
        IFS='|' read -r pattern description content should_match <<< "$pattern_test"
        test_pattern_match "pattern_${pattern//[^a-zA-Z0-9]/_}" "$content" "$pattern" "$should_match" "$description"
    done
}

# Test: Empty and malformed input handling
test_edge_cases() {
    # Empty input
    test_hook_response "empty_input" "" 1 "Empty input should be handled gracefully"
    
    # Malformed JSON
    test_hook_response "malformed_json" "invalid json" 1 "Malformed JSON should be handled gracefully"
    
    # Missing tool name
    local input='{"hook_event_name": "PreToolUse", "tool_input": {"prompt": "test"}}'
    test_hook_response "missing_tool_name" "$input" 0 "Missing tool name should be allowed"
    
    # Empty tool input
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "")
    test_hook_response "empty_tool_input" "$input" 0 "Empty tool input should be allowed"
    
    # Large content (performance test)
    local large_content
    large_content=$(printf 'a%.0s' {1..1000})  # 1KB of 'a' characters
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "$large_content")
    benchmark_test "large_content_performance" "echo '$input' | '$SCANNER_SCRIPT'" 3000 "Large content should process quickly"
}

# Test: Configuration and environment
test_configuration() {
    # Test with missing patterns file
    local temp_dir
    temp_dir=$(mktemp -d)
    
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "clean content")
    
    CONFIG_FILE="$temp_dir/nonexistent.conf" test_hook_response "missing_config" "$input" 0 "Missing config should use defaults"
    
    rm -rf "$temp_dir"
}

# Main test execution
main() {
    test_suite_start "$TEST_SUITE_NAME"
    
    # Verify test environment
    if [[ ! -f "$SCANNER_SCRIPT" ]]; then
        log_error "Scanner script not found: $SCANNER_SCRIPT"
        return 1
    fi
    
    if ! validate_test_fixtures "$FIXTURES_DIR"; then
        log_error "Test fixtures validation failed"
        return 1
    fi
    
    # Run test categories
    log_info "Testing non-MCP tools..."
    test_non_mcp_tools
    
    log_info "Testing non-PreToolUse events..."
    test_non_pretooluse_events
    
    log_info "Testing clean content..."
    test_clean_content
    
    log_info "Testing sensitive content detection..."
    test_sensitive_content
    
    log_info "Testing pattern matching..."
    test_pattern_matching
    
    log_info "Testing edge cases..."
    test_edge_cases
    
    log_info "Testing configuration handling..."
    test_configuration
    
    # End test suite
    test_suite_end "$TEST_SUITE_NAME"
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi