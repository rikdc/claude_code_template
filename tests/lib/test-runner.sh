#!/usr/bin/env bash

# Test Runner Framework
# Orchestrates test execution and provides common functionality

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
TESTS_ROOT="$(dirname "$SCRIPT_DIR")"
readonly TESTS_ROOT
PROJECT_ROOT="$(dirname "$TESTS_ROOT")"
readonly PROJECT_ROOT

# Load test helpers
# shellcheck source=test-helpers.sh
source "$SCRIPT_DIR/test-helpers.sh"

# Default configuration
DEFAULT_SCANNER_SCRIPT="$PROJECT_ROOT/.claude/hooks/mcp-security-scanner.sh"
DEFAULT_TEST_OUTPUT_DIR="$TESTS_ROOT/output"
DEFAULT_FIXTURES_DIR="$TESTS_ROOT/security-scanner/fixtures"

# Test configuration
export SCANNER_SCRIPT="${SCANNER_SCRIPT:-$DEFAULT_SCANNER_SCRIPT}"
export TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-$DEFAULT_TEST_OUTPUT_DIR}"
export FIXTURES_DIR="${FIXTURES_DIR:-$DEFAULT_FIXTURES_DIR}"

# Test discovery and execution
discover_tests() {
    local test_dir="$1"
    local pattern="${2:-*test*.sh}"
    
    find "$test_dir" -name "$pattern" -type f -executable | sort
}

run_test_file() {
    local test_file="$1"
    local test_name
    local test_name
    test_name=$(basename "$test_file" .sh)
    
    log_info "Running test file: $test_name"
    
    # Create isolated test environment
    local test_work_dir="$TEST_OUTPUT_DIR/$test_name"
    setup_test_env "$test_work_dir"
    
    # Export test environment variables
    export TEST_WORK_DIR="$test_work_dir"
    export TEST_NAME="$test_name"
    
    # Run the test file
    local test_result=0
    if ! "$test_file"; then
        test_result=1
        log_error "Test file failed: $test_name"
    fi
    
    # Cleanup if successful and not in debug mode
    if [[ $test_result -eq 0 && "${DEBUG:-}" != "1" ]]; then
        cleanup_test_env "$test_work_dir" false
    else
        cleanup_test_env "$test_work_dir" true
    fi
    
    return $test_result
}

run_tests_in_directory() {
    local test_dir="$1"
    local pattern="${2:-*test*.sh}"
    
    if [[ ! -d "$test_dir" ]]; then
        log_error "Test directory not found: $test_dir"
        return 1
    fi
    
    local test_files
    mapfile -t test_files < <(discover_tests "$test_dir" "$pattern")
    
    if [[ ${#test_files[@]} -eq 0 ]]; then
        log_warning "No test files found in: $test_dir"
        return 0
    fi
    
    local failed_tests=0
    for test_file in "${test_files[@]}"; do
        if ! run_test_file "$test_file"; then
            ((failed_tests++))
        fi
        echo
    done
    
    return $failed_tests
}

# Pre-flight checks
verify_test_environment() {
    local errors=0
    
    # Check if scanner script exists
    if [[ ! -f "$SCANNER_SCRIPT" ]]; then
        log_error "Scanner script not found: $SCANNER_SCRIPT"
        ((errors++))
    elif [[ ! -x "$SCANNER_SCRIPT" ]]; then
        log_error "Scanner script is not executable: $SCANNER_SCRIPT"
        ((errors++))
    fi
    
    # Check if test fixtures exist
    if [[ -d "$FIXTURES_DIR" ]]; then
        if ! validate_test_fixtures "$FIXTURES_DIR"; then
            ((errors++))
        fi
    else
        log_warning "Test fixtures directory not found: $FIXTURES_DIR"
    fi
    
    # Check required tools
    local required_tools=("jq" "grep" "awk")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            log_error "Required tool not found: $tool"
            ((errors++))
        fi
    done
    
    # Check optional tools
    local optional_tools=("gitleaks" "trufflehog" "git-secrets" "shellcheck")
    for tool in "${optional_tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            log_debug "Optional tool available: $tool"
        else
            log_debug "Optional tool not available: $tool"
        fi
    done
    
    return $errors
}

# Test result reporting
generate_test_report() {
    local output_dir="$1"
    local report_file="$output_dir/test-report.md"
    
    mkdir -p "$output_dir"
    
    cat > "$report_file" << EOF
# Test Report

Generated at: $(date)

## Environment
- Scanner Script: $SCANNER_SCRIPT
- Test Output Directory: $TEST_OUTPUT_DIR
- Fixtures Directory: $FIXTURES_DIR

## Test Results
EOF
    
    # Find all test log files
    local log_files
    if mapfile -t log_files < <(find "$output_dir" -name "test-results.log" 2>/dev/null); then
        for log_file in "${log_files[@]}"; do
            echo "### $(dirname "$log_file" | xargs basename)" >> "$report_file"
            echo '```' >> "$report_file"
            cat "$log_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo >> "$report_file"
        done
    fi
    
    log_info "Test report generated: $report_file"
}

# Cleanup functions
cleanup_all_test_artifacts() {
    if [[ -d "$TEST_OUTPUT_DIR" ]]; then
        rm -rf "$TEST_OUTPUT_DIR"
        log_info "Cleaned up all test artifacts"
    fi
}

# Main execution functions
main_test_runner() {
    local test_type="${1:-all}"
    local pattern="${2:-*test*.sh}"
    
    # Create output directory
    mkdir -p "$TEST_OUTPUT_DIR"
    
    # Verify environment
    if ! verify_test_environment; then
        log_error "Test environment verification failed"
        return 1
    fi
    
    local failed_suites=0
    
    case "$test_type" in
        "all")
            log_info "🚀 Running all tests"
            run_tests_in_directory "$TESTS_ROOT/security-scanner" "$pattern" || ((failed_suites++))
            run_tests_in_directory "$TESTS_ROOT/hooks" "$pattern" || ((failed_suites++))
            ;;
        "security")
            log_info "🔒 Running security scanner tests"
            run_tests_in_directory "$TESTS_ROOT/security-scanner" "$pattern" || ((failed_suites++))
            ;;
        "hooks")
            log_info "🪝 Running hook integration tests"
            run_tests_in_directory "$TESTS_ROOT/hooks" "$pattern" || ((failed_suites++))
            ;;
        "unit")
            log_info "🧪 Running unit tests"
            run_tests_in_directory "$TESTS_ROOT/security-scanner" "*unit*test*.sh" || ((failed_suites++))
            ;;
        "integration")
            log_info "🔗 Running integration tests"
            run_tests_in_directory "$TESTS_ROOT/security-scanner" "*integration*test*.sh" || ((failed_suites++))
            run_tests_in_directory "$TESTS_ROOT/hooks" "*integration*test*.sh" || ((failed_suites++))
            ;;
        *)
            log_error "Unknown test type: $test_type"
            log_info "Available types: all, security, hooks, unit, integration"
            return 1
            ;;
    esac
    
    # Generate report
    generate_test_report "$TEST_OUTPUT_DIR"
    
    if [[ $failed_suites -eq 0 ]]; then
        log_success "🎉 All test suites passed!"
        return 0
    else
        log_error "💥 $failed_suites test suite(s) failed!"
        return 1
    fi
}

# Help function
show_help() {
    cat << EOF
Test Runner Framework

Usage: $0 [test_type] [pattern]

Test Types:
  all           Run all tests (default)
  security      Run security scanner tests only
  hooks         Run hook integration tests only
  unit          Run unit tests only
  integration   Run integration tests only

Pattern:
  Optional file pattern to match test files (default: *test*.sh)

Environment Variables:
  SCANNER_SCRIPT    Path to the security scanner script
  TEST_OUTPUT_DIR   Directory for test output and logs
  FIXTURES_DIR      Directory containing test fixtures
  DEBUG             Set to 1 for debug output

Examples:
  $0                    # Run all tests
  $0 security           # Run security tests only
  $0 unit "*unit*"      # Run unit tests matching pattern
  DEBUG=1 $0 all        # Run all tests with debug output

EOF
}

# Script execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Handle help
    if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
        show_help
        exit 0
    fi
    
    # Run main function
    main_test_runner "${1:-all}" "${2:-*test*.sh}"
fi