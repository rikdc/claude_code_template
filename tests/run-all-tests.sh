#!/usr/bin/env bash

# Master Test Runner
# Executes all test suites for the MCP Security Scanner

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
LIB_DIR="$SCRIPT_DIR/lib"
readonly LIB_DIR

# shellcheck source=lib/test-runner.sh
source "$LIB_DIR/test-runner.sh"

# Test runner configuration
# This variable is currently unused, but kept for future use
MASTER_SUITE_NAME="MCP Security Scanner - Complete Test Suite"
readonly MASTER_SUITE_NAME

# Display banner
show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                  MCP Security Scanner Test Suite             ║
║                                                              ║
║  Comprehensive testing for Claude Code security hooks       ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo
}

# Show test summary
show_test_summary() {
    local total_suites="$1"
    local failed_suites="$2"
    local duration="$3"
    
    echo
    echo "════════════════════════════════════════════════════════════════"
    log_info "📊 Master Test Suite Summary"
    echo
    echo "Total test suites: $total_suites"
    
    if [[ $failed_suites -eq 0 ]]; then
        log_success "Failed test suites: $failed_suites"
    else
        log_error "Failed test suites: $failed_suites"
    fi
    
    echo "Total duration: ${duration}s"
    echo
    
    if [[ $failed_suites -eq 0 ]]; then
        log_success "🎉 All test suites passed!"
        echo
        echo "The MCP Security Scanner is working correctly and ready for use."
    else
        log_error "💥 Some test suites failed!"
        echo
        echo "Please check the test output above for details."
        echo "Test logs and reports are available in: $TEST_OUTPUT_DIR"
    fi
    
    echo "════════════════════════════════════════════════════════════════"
}

# Run specific test category
run_test_category() {
    local category="$1"
    local description="$2"
    
    log_info "🚀 Starting $description"
    echo
    
    local start_time end_time duration
    start_time=$(date +%s)
    
    local result=0
    main_test_runner "$category" || result=$?
    
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    echo
    if [[ $result -eq 0 ]]; then
        log_success "✅ $description completed successfully (${duration}s)"
    else
        log_error "❌ $description failed (${duration}s)"
    fi
    echo
    
    return "$result"
}

# Main execution function
main() {
    local test_type="${1:-all}"
    # This variable is defined but not used in this function
    local pattern="${2:-*test*.sh}"
    export TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR:-$SCRIPT_DIR/output}"
    
    show_banner
    
    # Pre-flight checks
    log_info "🔍 Running pre-flight checks..."
    if ! verify_test_environment; then
        log_error "Pre-flight checks failed. Cannot continue."
        return 1
    fi
    log_success "Pre-flight checks passed"
    echo
    
    # Test execution tracking
    local master_start_time failed_suites total_suites
    master_start_time=$(date +%s)
    failed_suites=0
    total_suites=0
    
    case "$test_type" in
        "all")
            log_info "🧪 Running complete test suite"
            echo
            
            # Unit tests
            ((total_suites++))
            run_test_category "unit" "Unit Tests" || ((failed_suites++))
            
            # Integration tests  
            ((total_suites++))
            run_test_category "integration" "Integration Tests" || ((failed_suites++))
            
            # Hook system tests
            ((total_suites++))
            run_test_category "hooks" "Hook System Tests" || ((failed_suites++))
            ;;
            
        "unit")
            ((total_suites++))
            run_test_category "unit" "Unit Tests" || ((failed_suites++))
            ;;
            
        "integration")
            ((total_suites++))
            run_test_category "integration" "Integration Tests" || ((failed_suites++))
            ;;
            
        "security")
            log_info "🔒 Running all security scanner tests"
            echo
            
            # Security unit tests
            ((total_suites++))
            run_test_category "unit" "Security Unit Tests" || ((failed_suites++))
            
            # Security integration tests
            ((total_suites++))
            if ! run_tests_in_directory "$SCRIPT_DIR/security-scanner" "*integration*test*.sh"; then
                ((failed_suites++))
            fi
            ;;
            
        "hooks")
            ((total_suites++))
            run_test_category "hooks" "Hook System Tests" || ((failed_suites++))
            ;;
            
        "performance")
            log_info "⚡ Running performance tests"
            echo
            
            ((total_suites++))
            # Run tests with performance focus
            DEBUG=0 run_test_category "all" "Performance Tests" || ((failed_suites++))
            ;;
            
        "quick")
            log_info "🏃 Running quick test suite (unit tests only)"
            echo
            
            ((total_suites++))
            run_test_category "unit" "Quick Unit Tests" || ((failed_suites++))
            ;;
            
        *)
            log_error "Unknown test type: $test_type"
            log_info "Available types: all, unit, integration, security, hooks, performance, quick"
            return 1
            ;;
    esac
    
    # Calculate total duration
    local master_end_time total_duration
    master_end_time=$(date +%s)
    total_duration=$((master_end_time - master_start_time))
    
    # Show final summary
    show_test_summary "$total_suites" "$failed_suites" "$total_duration"
    
    # Generate consolidated report
    generate_test_report "$TEST_OUTPUT_DIR"
    
    return $failed_suites
}

# Help function
show_help() {
    cat << EOF
MCP Security Scanner - Master Test Runner

Usage: $0 [test_type] [pattern]

Test Types:
  all           Run complete test suite (default)
  unit          Run unit tests only
  integration   Run integration tests only
  security      Run all security scanner tests
  hooks         Run hook system tests only
  performance   Run performance-focused tests
  quick         Run quick test suite (unit tests)

Pattern:
  Optional file pattern to match test files (default: *test*.sh)

Environment Variables:
  SCANNER_SCRIPT    Path to the security scanner script
  TEST_OUTPUT_DIR   Directory for test output and logs  
  DEBUG             Set to 1 for debug output
  KEEP_LOGS         Set to 1 to preserve test logs

Examples:
  $0                        # Run all tests
  $0 unit                   # Run unit tests only
  $0 security               # Run all security tests
  $0 performance            # Run performance tests
  DEBUG=1 $0 all            # Run all tests with debug output
  KEEP_LOGS=1 $0 integration # Run integration tests and keep logs

Test Output:
  Test results and logs are saved to: $TEST_OUTPUT_DIR
  A consolidated report is generated after test completion.

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
    main "${1:-all}" "${2:-*test*.sh}"
fi