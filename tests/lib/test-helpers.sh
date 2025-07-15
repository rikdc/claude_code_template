#!/usr/bin/env bash

# Test Helpers Library
# Shared utilities for MCP Security Scanner tests

set -euo pipefail

# Colors for output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_PURPLE='\033[0;35m'
readonly COLOR_CYAN='\033[0;36m'
readonly COLOR_NC='\033[0m' # No Color

# Test result tracking
declare -g TESTS_RUN=0
declare -g TESTS_PASSED=0
declare -g TESTS_FAILED=0
declare -g TEST_START_TIME
declare -g TEST_LOG_FILE

# Logging functions
log_info() {
    echo -e "${COLOR_BLUE}ℹ️  $*${COLOR_NC}"
}

log_success() {
    echo -e "${COLOR_GREEN}✅ $*${COLOR_NC}"
}

log_warning() {
    echo -e "${COLOR_YELLOW}⚠️  $*${COLOR_NC}"
}

log_error() {
    echo -e "${COLOR_RED}❌ $*${COLOR_NC}"
}

log_debug() {
    if [[ "${DEBUG:-}" == "1" ]]; then
        echo -e "${COLOR_PURPLE}🐛 $*${COLOR_NC}"
    fi
}

# Test framework functions
test_suite_start() {
    local suite_name="$1"
    TEST_START_TIME=$(date +%s)
    readonly TEST_START_TIME
    TESTS_RUN=0
    TESTS_PASSED=0
    TESTS_FAILED=0
    
    log_info "🧪 Starting test suite: $suite_name"
    echo
}

test_suite_end() {
    local suite_name="$1"
    local end_time
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - TEST_START_TIME))
    
    echo
    log_info "📊 Test Suite Results: $suite_name"
    echo "Tests run: $TESTS_RUN"
    log_success "Tests passed: $TESTS_PASSED"
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        log_error "Tests failed: $TESTS_FAILED"
    else
        echo "Tests failed: $TESTS_FAILED"
    fi
    
    echo "Duration: ${duration}s"
    echo
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_success "🎉 All tests passed!"
        return 0
    else
        log_error "💥 Some tests failed!"
        return 1
    fi
}

# Core test assertion function
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    local description="${4:-}"
    
    ((TESTS_RUN++))
    
    if [[ -n "$description" ]]; then
        log_debug "Running: $test_name - $description"
    else
        log_debug "Running: $test_name"
    fi
    
    local actual_exit_code=0
    local test_output
    
    # Capture both stdout and stderr
    if test_output=$(eval "$test_command" 2>&1); then
        actual_exit_code=0
    else
        actual_exit_code=$?
    fi
    
    if [[ $actual_exit_code -eq $expected_exit_code ]]; then
        echo -e "${test_name}... ${COLOR_GREEN}PASS${COLOR_NC}"
        ((TESTS_PASSED++))
        
        if [[ -n "${TEST_LOG_FILE:-}" ]]; then
            echo "PASS: $test_name ($description)" >> "$TEST_LOG_FILE"
        fi
        
        return 0
    else
        echo -e "${test_name}... ${COLOR_RED}FAIL${COLOR_NC}"
        log_error "Expected exit code $expected_exit_code, got $actual_exit_code"
        
        if [[ -n "$test_output" ]]; then
            log_error "Output: $test_output"
        fi
        
        ((TESTS_FAILED++))
        
        if [[ -n "${TEST_LOG_FILE:-}" ]]; then
            echo "FAIL: $test_name ($description)" >> "$TEST_LOG_FILE"
            echo "Expected: $expected_exit_code, Got: $actual_exit_code" >> "$TEST_LOG_FILE"
            echo "Output: $test_output" >> "$TEST_LOG_FILE"
            echo "---" >> "$TEST_LOG_FILE"
        fi
        
        return 1
    fi
}

# Specialized test functions
test_hook_response() {
    local test_name="$1"
    local hook_input="$2"
    local expected_exit_code="$3"
    local description="$4"
    local scanner_script="${SCANNER_SCRIPT:-}"
    
    if [[ -z "$scanner_script" || ! -f "$scanner_script" ]]; then
        log_error "Scanner script not found: $scanner_script"
        return 1
    fi
    
    local test_command="echo '$hook_input' | '$scanner_script'"
    run_test "$test_name" "$test_command" "$expected_exit_code" "$description"
}

test_pattern_match() {
    local test_name="$1"
    local content="$2"
    local pattern="$3"
    local should_match="$4"  # true/false
    local description="$5"
    
    local test_command="echo '$content' | grep -qi '$pattern'"
    local expected_exit_code
    
    if [[ "$should_match" == "true" ]]; then
        expected_exit_code=0
    else
        expected_exit_code=1
    fi
    
    run_test "$test_name" "$test_command" "$expected_exit_code" "$description"
}

# Test environment setup/teardown
setup_test_env() {
    local test_dir="$1"
    
    if [[ -d "$test_dir" ]]; then
        rm -rf "$test_dir"
    fi
    
    mkdir -p "$test_dir"
    TEST_LOG_FILE="$test_dir/test-results.log"
    
    echo "Test run started at $(date)" > "$TEST_LOG_FILE"
    log_debug "Test environment created: $test_dir"
}

cleanup_test_env() {
    local test_dir="$1"
    local keep_logs="${2:-false}"
    
    if [[ "$keep_logs" == "false" && -d "$test_dir" ]]; then
        rm -rf "$test_dir"
        log_debug "Test environment cleaned up: $test_dir"
    elif [[ "$keep_logs" == "true" ]]; then
        log_info "Test logs preserved in: $test_dir"
    fi
}

# Utility functions
create_temp_file() {
    local content="$1"
    local temp_file
    local temp_file
    temp_file=$(mktemp)
    echo "$content" > "$temp_file"
    echo "$temp_file"
}

wait_for_file() {
    local file_path="$1"
    local timeout="${2:-10}"
    local counter=0
    
    while [[ ! -f "$file_path" && $counter -lt $timeout ]]; do
        sleep 1
        ((counter++))
    done
    
    [[ -f "$file_path" ]]
}

# Performance testing utilities
benchmark_test() {
    local test_name="$1"
    local test_command="$2"
    local max_duration_ms="${3:-5000}"  # 5 seconds default
    local description="$4"
    
    local start_time end_time duration_ms
    local start_time
    start_time=$(date +%s%N)
    
    local test_result=0
    eval "$test_command" || test_result=$?
    
    local end_time
    end_time=$(date +%s%N)
    duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    if [[ $duration_ms -le $max_duration_ms ]]; then
        echo -e "${test_name}... ${COLOR_GREEN}PASS${COLOR_NC} (${duration_ms}ms)"
        ((TESTS_PASSED++))
    else
        echo -e "${test_name}... ${COLOR_RED}SLOW${COLOR_NC} (${duration_ms}ms > ${max_duration_ms}ms)"
        ((TESTS_FAILED++))
    fi
    
    ((TESTS_RUN++))
    return $test_result
}

# JSON test utilities
create_mcp_request() {
    local tool_name="$1"
    local prompt="${2:-}"
    local code="${3:-}"
    local query="${4:-}"
    
    local json_content='{"hook_event_name": "PreToolUse", "tool_name": "'"$tool_name"'"'
    
    if [[ -n "$prompt" || -n "$code" || -n "$query" ]]; then
        json_content+=', "tool_input": {'
        
        local input_parts=()
        [[ -n "$prompt" ]] && input_parts+=("\"prompt\": \"$prompt\"")
        [[ -n "$code" ]] && input_parts+=("\"code\": \"$code\"")
        [[ -n "$query" ]] && input_parts+=("\"query\": \"$query\"")
        
        # Join array elements with commas
        local IFS=','
        json_content+="${input_parts[*]}"
        json_content+='}'
    fi
    
    json_content+='}'
    echo "$json_content"
}

# Test data validation
validate_test_fixtures() {
    local fixtures_dir="$1"
    local required_files=("clean-inputs.json" "sensitive-inputs.json" "test-patterns.conf")
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$fixtures_dir/$file" ]]; then
            log_error "Missing test fixture: $fixtures_dir/$file"
            return 1
        fi
    done
    
    log_debug "All test fixtures validated"
    return 0
}

# Export functions for use in test scripts
export -f log_info log_success log_warning log_error log_debug
export -f test_suite_start test_suite_end run_test test_hook_response test_pattern_match
export -f setup_test_env cleanup_test_env create_temp_file wait_for_file benchmark_test
export -f create_mcp_request validate_test_fixtures