#!/usr/bin/env bash

# Integration Tests for MCP Security Scanner
# Tests complete hook workflow and external tool integration

set -euo pipefail

# Get script directory and load test framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"
readonly LIB_DIR

# shellcheck source=../lib/test-helpers.sh
source "$LIB_DIR/test-helpers.sh"

# Test configuration
TEST_SUITE_NAME="Security Scanner Integration Tests"
readonly TEST_SUITE_NAME
FIXTURES_DIR="$SCRIPT_DIR/fixtures"
readonly FIXTURES_DIR

# Test: External security tools integration
test_external_tools_integration() {
    local available_tools=()
    
    # Check which tools are available
    command -v gitleaks >/dev/null 2>&1 && available_tools+=("gitleaks")
    command -v trufflehog >/dev/null 2>&1 && available_tools+=("trufflehog")
    command -v git-secrets >/dev/null 2>&1 && available_tools+=("git-secrets")
    
    if [[ ${#available_tools[@]} -eq 0 ]]; then
        log_warning "No external security tools available for integration testing"
        return 0
    fi
    
    log_info "Testing integration with external tools: ${available_tools[*]}"
    
    # Test that external tools are properly integrated
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "API_KEY=sk-test123456789abcdefghijklmnopqrstuvwxyz123456")
    
    # The scanner should detect secrets using external tools
    test_hook_response "external_tools_detection" "$input" 1 "External tools should detect secrets"
    
    # Test with clean content
    input=$(create_mcp_request "mcp__context7__get-library-docs" "How to use React hooks?")
    test_hook_response "external_tools_clean" "$input" 0 "External tools should allow clean content"
}

# Test: Logging and audit trail
test_logging_functionality() {
    local temp_log_dir
    temp_log_dir=$(mktemp -d)
    local test_log_file="$temp_log_dir/security-scan.log"
    
    # Test logging with sensitive content
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "API_KEY=sk-test123456789abcdef")
    
    LOG_FILE="$test_log_file" test_hook_response "logging_sensitive" "$input" 1 "Sensitive content should be logged"
    
    # Verify log file was created and contains expected entries
    if [[ -f "$test_log_file" ]]; then
        if grep -q "SECURITY VIOLATION" "$test_log_file"; then
            log_success "Security violation properly logged"
            ((TESTS_PASSED++))
        else
            log_error "Security violation not found in log"
            ((TESTS_FAILED++))
        fi
        ((TESTS_RUN++))
    else
        log_error "Log file was not created"
        ((TESTS_FAILED++))
        ((TESTS_RUN++))
    fi
    
    # Test logging with clean content
    input=$(create_mcp_request "mcp__context7__get-library-docs" "How to use React hooks?")
    LOG_FILE="$test_log_file" test_hook_response "logging_clean" "$input" 0 "Clean content should be logged as passed"
    
    # Verify success is logged
    if grep -q "Security scan passed" "$test_log_file"; then
        log_success "Clean content success properly logged"
        ((TESTS_PASSED++))
    else
        log_error "Clean content success not found in log"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    rm -rf "$temp_log_dir"
}

# Test: Custom patterns configuration
test_custom_patterns() {
    local temp_config_dir
    temp_config_dir=$(mktemp -d)
    local custom_patterns_file="$temp_config_dir/custom-patterns.conf"
    
    # Create custom patterns file
    cat > "$custom_patterns_file" << 'EOF'
# Custom test patterns
CUSTOM_SECRET=custom.*secret.*[=:]
COMPANY_API_KEY=mycompany.*api.*key
TEST_TOKEN=test.*token.*[0-9]\{6,\}
EOF
    
    # Test with custom pattern
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "CUSTOM_SECRET=mysecretvalue")
    
    CONFIG_FILE="$custom_patterns_file" test_hook_response "custom_pattern_match" "$input" 1 "Custom patterns should be detected"
    
    # Test with non-matching content
    input=$(create_mcp_request "mcp__context7__get-library-docs" "normal content")
    CONFIG_FILE="$custom_patterns_file" test_hook_response "custom_pattern_nomatch" "$input" 0 "Non-matching content should be allowed"
    
    rm -rf "$temp_config_dir"
}

# Test: Performance with large inputs
test_performance_scenarios() {
    local large_sizes=(1000 5000 10000)  # Characters
    
    for size in "${large_sizes[@]}"; do
        local large_content
        large_content=$(printf 'a%.0s' $(seq 1 "$size"))
        
        local input
        input=$(create_mcp_request "mcp__context7__get-library-docs" "$large_content")
        
        # Should complete within 5 seconds
        benchmark_test "performance_${size}_chars" "echo '$input' | '$SCANNER_SCRIPT'" 5000 "Processing ${size} characters should be fast"
    done
}

# Test: Concurrent execution
test_concurrent_execution() {
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # Create multiple test inputs
    local clean_input sensitive_input
    clean_input=$(create_mcp_request "mcp__context7__get-library-docs" "How to use React hooks?")
    sensitive_input=$(create_mcp_request "mcp__context7__get-library-docs" "API_KEY=sk-test123456789abcdef")
    
    # Run multiple scanner instances concurrently
    local pids=()
    for i in {1..5}; do
        (
            echo "$clean_input" | "$SCANNER_SCRIPT" > "$temp_dir/clean_$i.out" 2>&1
            echo $? > "$temp_dir/clean_$i.exit"
        ) &
        pids+=($!)
        
        (
            echo "$sensitive_input" | "$SCANNER_SCRIPT" > "$temp_dir/sensitive_$i.out" 2>&1
            echo $? > "$temp_dir/sensitive_$i.exit"
        ) &
        pids+=($!)
    done
    
    # Wait for all background processes
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    # Verify results
    local concurrent_failures=0
    
    for i in {1..5}; do
        # Check clean content results
        if [[ -f "$temp_dir/clean_$i.exit" ]]; then
            local exit_code
            exit_code=$(cat "$temp_dir/clean_$i.exit")
            if [[ $exit_code -ne 0 ]]; then
                log_error "Concurrent clean test $i failed with exit code $exit_code"
                ((concurrent_failures++))
            fi
        else
            log_error "Concurrent clean test $i result file missing"
            ((concurrent_failures++))
        fi
        
        # Check sensitive content results
        if [[ -f "$temp_dir/sensitive_$i.exit" ]]; then
            local exit_code
            exit_code=$(cat "$temp_dir/sensitive_$i.exit")
            if [[ $exit_code -ne 1 ]]; then
                log_error "Concurrent sensitive test $i failed with exit code $exit_code (expected 1)"
                ((concurrent_failures++))
            fi
        else
            log_error "Concurrent sensitive test $i result file missing"
            ((concurrent_failures++))
        fi
    done
    
    if [[ $concurrent_failures -eq 0 ]]; then
        log_success "Concurrent execution test passed"
        ((TESTS_PASSED++))
    else
        log_error "Concurrent execution test failed ($concurrent_failures failures)"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    rm -rf "$temp_dir"
}

# Test: Hook system integration
test_hook_system_integration() {
    # Test that the hook properly integrates with Claude Code's hook system
    # This tests the JSON parsing and response format
    
    local test_cases=(
        # Standard MCP request
        '{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "clean content"}}'
        # Missing tool_input
        '{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs"}'
        # Empty tool_input
        '{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {}}'
        # Multiple input fields
        '{"hook_event_name": "PreToolUse", "tool_name": "mcp__magic__generate-component", "tool_input": {"prompt": "create button", "code": "const clean = true;", "query": "button component"}}'
    )
    
    local test_names=(
        "standard_request"
        "missing_tool_input"
        "empty_tool_input"
        "multiple_input_fields"
    )
    
    for i in "${!test_cases[@]}"; do
        local input="${test_cases[$i]}"
        local name="${test_names[$i]}"
        
        test_hook_response "hook_integration_$name" "$input" 0 "Hook system integration: $name"
    done
}

# Test: Error recovery and resilience
test_error_recovery() {
    # Test behavior when external tools fail or are unavailable
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # Create a mock "gitleaks" that always fails
    cat > "$temp_dir/gitleaks" << 'EOF'
#!/bin/bash
exit 2  # Simulate tool failure
EOF
    chmod +x "$temp_dir/gitleaks"
    
    # Test with failed external tool
    local input
    input=$(create_mcp_request "mcp__context7__get-library-docs" "API_KEY=sk-test123456789abcdef")
    
    PATH="$temp_dir:$PATH" test_hook_response "external_tool_failure" "$input" 1 "Should handle external tool failures gracefully"
    
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
    
    # Run integration test categories
    log_info "Testing external tools integration..."
    test_external_tools_integration
    
    log_info "Testing logging functionality..."
    test_logging_functionality
    
    log_info "Testing custom patterns..."
    test_custom_patterns
    
    log_info "Testing performance scenarios..."
    test_performance_scenarios
    
    log_info "Testing concurrent execution..."
    test_concurrent_execution
    
    log_info "Testing hook system integration..."
    test_hook_system_integration
    
    log_info "Testing error recovery..."
    test_error_recovery
    
    # End test suite
    test_suite_end "$TEST_SUITE_NAME"
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi