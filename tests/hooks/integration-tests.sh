#!/usr/bin/env bash

# Hook Integration Tests
# Tests the complete Claude Code hook system integration

set -euo pipefail

# Get script directory and load test framework
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly LIB_DIR="$(dirname "$SCRIPT_DIR")/lib"

# shellcheck source=../lib/test-helpers.sh
source "$LIB_DIR/test-helpers.sh"

# Test configuration
readonly TEST_SUITE_NAME="Hook System Integration Tests"
readonly PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
readonly CLAUDE_DIR="$PROJECT_ROOT/.claude"

# Test: Hook configuration validation
test_hook_configuration() {
    local settings_file="$CLAUDE_DIR/settings.json"
    
    # Verify settings.json exists and is valid
    if [[ ! -f "$settings_file" ]]; then
        log_error "Hook settings file not found: $settings_file"
        ((TESTS_FAILED++))
        ((TESTS_RUN++))
        return 1
    fi
    
    # Validate JSON syntax
    if jq empty "$settings_file" 2>/dev/null; then
        log_success "Hook settings JSON is valid"
        ((TESTS_PASSED++))
    else
        log_error "Hook settings JSON is invalid"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    # Verify hook configuration structure
    local has_hooks has_pretooluse has_matcher has_command
    has_hooks=$(jq 'has("hooks")' "$settings_file")
    has_pretooluse=$(jq '.hooks | has("PreToolUse")' "$settings_file")
    has_matcher=$(jq '.hooks.PreToolUse[0] | has("matcher")' "$settings_file")
    has_command=$(jq '.hooks.PreToolUse[0].hooks[0] | has("command")' "$settings_file")
    
    if [[ "$has_hooks" == "true" && "$has_pretooluse" == "true" && "$has_matcher" == "true" && "$has_command" == "true" ]]; then
        log_success "Hook configuration structure is correct"
        ((TESTS_PASSED++))
    else
        log_error "Hook configuration structure is incomplete"
        log_error "has_hooks: $has_hooks, has_pretooluse: $has_pretooluse, has_matcher: $has_matcher, has_command: $has_command"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    # Verify matcher pattern
    local matcher_pattern
    matcher_pattern=$(jq -r '.hooks.PreToolUse[0].matcher' "$settings_file")
    if [[ "$matcher_pattern" == "mcp__.*" ]]; then
        log_success "Hook matcher pattern is correct: $matcher_pattern"
        ((TESTS_PASSED++))
    else
        log_error "Hook matcher pattern is incorrect: $matcher_pattern (expected: mcp__.*)"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    # Verify command path
    local command_path
    command_path=$(jq -r '.hooks.PreToolUse[0].hooks[0].command' "$settings_file")
    local expected_path=".claude/hooks/mcp-security-scanner.sh"
    if [[ "$command_path" == "$expected_path" ]]; then
        log_success "Hook command path is correct: $command_path"
        ((TESTS_PASSED++))
    else
        log_error "Hook command path is incorrect: $command_path (expected: $expected_path)"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
}

# Test: Hook script permissions and execution
test_hook_script_executable() {
    local hook_script="$CLAUDE_DIR/hooks/mcp-security-scanner.sh"
    
    # Verify script exists
    if [[ ! -f "$hook_script" ]]; then
        log_error "Hook script not found: $hook_script"
        ((TESTS_FAILED++))
        ((TESTS_RUN++))
        return 1
    fi
    
    log_success "Hook script exists: $hook_script"
    ((TESTS_PASSED++))
    ((TESTS_RUN++))
    
    # Verify script is executable
    if [[ -x "$hook_script" ]]; then
        log_success "Hook script is executable"
        ((TESTS_PASSED++))
    else
        log_error "Hook script is not executable"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    # Verify script has proper shebang
    local shebang
    shebang=$(head -n1 "$hook_script")
    if [[ "$shebang" =~ ^\#\!/usr/bin/env\ bash$ ]]; then
        log_success "Hook script has correct shebang: $shebang"
        ((TESTS_PASSED++))
    else
        log_error "Hook script has incorrect shebang: $shebang"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
}

# Test: Hook environment and dependencies
test_hook_dependencies() {
    # Test required tools availability
    local required_tools=("jq" "grep" "awk" "mktemp")
    local missing_tools=()
    
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -eq 0 ]]; then
        log_success "All required tools are available: ${required_tools[*]}"
        ((TESTS_PASSED++))
    else
        log_error "Missing required tools: ${missing_tools[*]}"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    
    # Test optional tools availability
    local optional_tools=("gitleaks" "trufflehog" "git-secrets")
    local available_optional=()
    
    for tool in "${optional_tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            available_optional+=("$tool")
        fi
    done
    
    if [[ ${#available_optional[@]} -gt 0 ]]; then
        log_info "Available optional security tools: ${available_optional[*]}"
    else
        log_warning "No optional security tools available"
    fi
}

# Test: Configuration file structure
test_configuration_files() {
    # Test security patterns example file
    local patterns_example="$CLAUDE_DIR/security-patterns.conf.example"
    if [[ -f "$patterns_example" ]]; then
        log_success "Security patterns example file exists"
        ((TESTS_PASSED++))
        
        # Verify it has valid pattern format
        local pattern_count
        pattern_count=$(grep -c '^[A-Z_]*=' "$patterns_example" || true)
        if [[ $pattern_count -gt 0 ]]; then
            log_success "Security patterns example contains $pattern_count patterns"
            ((TESTS_PASSED++))
        else
            log_error "Security patterns example appears to be empty or invalid"
            ((TESTS_FAILED++))
        fi
    else
        log_error "Security patterns example file not found: $patterns_example"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN += 2))
    
    # Test README file
    local readme_file="$CLAUDE_DIR/README.md"
    if [[ -f "$readme_file" ]]; then
        log_success "Claude directory README exists"
        ((TESTS_PASSED++))
        
        # Check for key sections
        local required_sections=("How It Works" "Setup" "Configuration" "Testing")
        local missing_sections=()
        
        for section in "${required_sections[@]}"; do
            if ! grep -q "## $section" "$readme_file"; then
                missing_sections+=("$section")
            fi
        done
        
        if [[ ${#missing_sections[@]} -eq 0 ]]; then
            log_success "README contains all required sections"
            ((TESTS_PASSED++))
        else
            log_error "README missing sections: ${missing_sections[*]}"
            ((TESTS_FAILED++))
        fi
    else
        log_error "Claude directory README not found: $readme_file"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN += 2))
}

# Test: Hook system behavior simulation
test_hook_system_simulation() {
    # Simulate how Claude Code would call the hook
    local temp_dir
    temp_dir=$(mktemp -d)
    
    # Create test input that mimics Claude Code's hook invocation
    local test_inputs=(
        '{"session_id": "test-123", "transcript_path": "/tmp/test.jsonl", "hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "How to use React?"}}'
        '{"session_id": "test-456", "transcript_path": "/tmp/test.jsonl", "hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "API key: sk-test123456789abcdef"}}'
        '{"session_id": "test-789", "transcript_path": "/tmp/test.jsonl", "hook_event_name": "PreToolUse", "tool_name": "Read", "tool_input": {"file_path": "/etc/passwd"}}'
    )
    
    local expected_exits=(0 1 0)
    local test_names=("clean_content_full" "sensitive_content_full" "non_mcp_tool_full")
    
    for i in "${!test_inputs[@]}"; do
        local input="${test_inputs[$i]}"
        local expected_exit="${expected_exits[$i]}"
        local test_name="${test_names[$i]}"
        
        test_hook_response "hook_simulation_$test_name" "$input" "$expected_exit" "Full hook system simulation: $test_name"
    done
    
    rm -rf "$temp_dir"
}

# Test: Performance under hook system conditions
test_hook_performance() {
    # Test response time under typical hook conditions
    local input='{"session_id": "perf-test", "transcript_path": "/tmp/test.jsonl", "hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "How to optimize React performance?"}}'
    
    # Should complete within 2 seconds for typical clean content
    benchmark_test "hook_performance_clean" "echo '$input' | '$SCANNER_SCRIPT'" 2000 "Hook should respond quickly to clean content"
    
    # Test with sensitive content (may take longer due to external tools)
    input='{"session_id": "perf-test", "transcript_path": "/tmp/test.jsonl", "hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "API_KEY=sk-test123456789abcdefghijklmnopqrstuvwxyz123456"}}'
    
    # Should complete within 5 seconds even with external tool scanning
    benchmark_test "hook_performance_sensitive" "echo '$input' | '$SCANNER_SCRIPT'" 5000 "Hook should respond reasonably quickly to sensitive content"
}

# Test: Hook error handling and recovery
test_hook_error_handling() {
    # Test malformed input handling
    local malformed_inputs=(
        ""
        "invalid json"
        '{"incomplete": true'
        '{"hook_event_name": "PreToolUse"}'
    )
    
    for i in "${!malformed_inputs[@]}"; do
        local input="${malformed_inputs[$i]}"
        # Malformed inputs should not crash the hook (exit gracefully)
        run_test "hook_error_malformed_$i" "echo '$input' | '$SCANNER_SCRIPT' >/dev/null 2>&1" 1 "Hook should handle malformed input gracefully"
    done
}

# Main test execution
main() {
    test_suite_start "$TEST_SUITE_NAME"
    
    log_info "Testing hook configuration..."
    test_hook_configuration
    
    log_info "Testing hook script executable status..."
    test_hook_script_executable
    
    log_info "Testing hook dependencies..."
    test_hook_dependencies
    
    log_info "Testing configuration files..."
    test_configuration_files
    
    log_info "Testing hook system simulation..."
    test_hook_system_simulation
    
    log_info "Testing hook performance..."
    test_hook_performance
    
    log_info "Testing hook error handling..."
    test_hook_error_handling
    
    # End test suite
    test_suite_end "$TEST_SUITE_NAME"
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi