#!/usr/bin/env bash

# Configuration Validation Script for MCP Security Scanner
# Validates all configuration files and setup

set -euo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly CLAUDE_DIR="$PROJECT_ROOT/.claude"
readonly HOOKS_DIR="$CLAUDE_DIR/hooks"
readonly SCANNER_SCRIPT="$HOOKS_DIR/mcp-security-scanner.sh"
readonly SETTINGS_FILE="$CLAUDE_DIR/settings.json"
readonly PATTERNS_FILE="$CLAUDE_DIR/security-patterns.conf"
readonly PATTERNS_EXAMPLE="$CLAUDE_DIR/security-patterns.conf.example"

# Colors for output
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly YELLOW='\033[1;33m'
readonly RED='\033[0;31m'
readonly NC='\033[0m'

# Validation counters
CHECKS_RUN=0
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Logging functions
info() {
    echo -e "${BLUE}ℹ️  $*${NC}"
}

success() {
    echo -e "${GREEN}✅ $*${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $*${NC}"
    ((WARNINGS++))
}

error() {
    echo -e "${RED}❌ $*${NC}"
}

# Validation helper
validate_check() {
    local description="$1"
    local test_command="$2"
    local is_warning="${3:-false}"
    
    ((CHECKS_RUN++))
    
    if eval "$test_command" >/dev/null 2>&1; then
        success "$description"
        ((CHECKS_PASSED++))
        return 0
    else
        if [[ "$is_warning" == "true" ]]; then
            warning "$description"
        else
            error "$description"
            ((CHECKS_FAILED++))
        fi
        return 1
    fi
}

# Show validation banner
show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║              MCP Security Scanner Validator                  ║
║                                                              ║
║  Validating configuration and setup                          ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo
}

# Validate file structure
validate_file_structure() {
    info "Validating file structure..."
    
    # Required files
    validate_check "Claude directory exists" "[[ -d '$CLAUDE_DIR' ]]"
    validate_check "Hooks directory exists" "[[ -d '$HOOKS_DIR' ]]"
    validate_check "Security scanner script exists" "[[ -f '$SCANNER_SCRIPT' ]]"
    validate_check "Hook settings file exists" "[[ -f '$SETTINGS_FILE' ]]"
    validate_check "Security patterns example exists" "[[ -f '$PATTERNS_EXAMPLE' ]]" "true"
    
    # Optional files
    validate_check "Security patterns file exists" "[[ -f '$PATTERNS_FILE' ]]" "true"
    validate_check "README file exists" "[[ -f '$CLAUDE_DIR/README.md' ]]" "true"
    
    echo
}

# Validate file permissions
validate_permissions() {
    info "Validating file permissions..."
    
    validate_check "Security scanner script is executable" "[[ -x '$SCANNER_SCRIPT' ]]"
    validate_check "Security scanner script is readable" "[[ -r '$SCANNER_SCRIPT' ]]"
    validate_check "Hook settings file is readable" "[[ -r '$SETTINGS_FILE' ]]"
    
    if [[ -f "$PATTERNS_FILE" ]]; then
        validate_check "Security patterns file is readable" "[[ -r '$PATTERNS_FILE' ]]"
    fi
    
    echo
}

# Validate JSON configuration
validate_json_config() {
    info "Validating JSON configuration..."
    
    # Validate JSON syntax
    validate_check "Hook settings JSON is valid" "jq empty '$SETTINGS_FILE'"
    
    if [[ $? -eq 0 ]]; then
        # Validate structure
        validate_check "Hook configuration has 'hooks' section" "jq -e '.hooks' '$SETTINGS_FILE'"
        validate_check "Hook configuration has 'PreToolUse' section" "jq -e '.hooks.PreToolUse' '$SETTINGS_FILE'"
        validate_check "Hook configuration has matcher pattern" "jq -e '.hooks.PreToolUse[0].matcher' '$SETTINGS_FILE'"
        validate_check "Hook configuration has command path" "jq -e '.hooks.PreToolUse[0].hooks[0].command' '$SETTINGS_FILE'"
        
        # Validate matcher pattern
        local matcher_pattern
        matcher_pattern=$(jq -r '.hooks.PreToolUse[0].matcher' "$SETTINGS_FILE" 2>/dev/null || echo "")
        if [[ "$matcher_pattern" == "mcp__.*" ]]; then
            success "Hook matcher pattern is correct: $matcher_pattern"
            ((CHECKS_PASSED++))
        else
            error "Hook matcher pattern is incorrect: $matcher_pattern (expected: mcp__.*)"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))
        
        # Validate command path
        local command_path
        command_path=$(jq -r '.hooks.PreToolUse[0].hooks[0].command' "$SETTINGS_FILE" 2>/dev/null || echo "")
        if [[ "$command_path" == ".claude/hooks/mcp-security-scanner.sh" ]]; then
            success "Hook command path is correct: $command_path"
            ((CHECKS_PASSED++))
        else
            error "Hook command path is incorrect: $command_path (expected: .claude/hooks/mcp-security-scanner.sh)"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))
    fi
    
    echo
}

# Validate security patterns
validate_security_patterns() {
    info "Validating security patterns..."
    
    if [[ -f "$PATTERNS_FILE" ]]; then
        # Check if patterns file has content
        local pattern_count
        pattern_count=$(grep -c '^[A-Z_]*=' "$PATTERNS_FILE" 2>/dev/null || echo "0")
        
        if [[ $pattern_count -gt 0 ]]; then
            success "Security patterns file contains $pattern_count patterns"
            ((CHECKS_PASSED++))
        else
            warning "Security patterns file appears to be empty"
        fi
        ((CHECKS_RUN++))
        
        # Validate pattern syntax
        local invalid_patterns=0
        while IFS='=' read -r pattern_name pattern_regex; do
            [[ "$pattern_name" =~ ^#.*$ ]] || [[ -z "$pattern_name" ]] && continue
            [[ -z "$pattern_regex" ]] && continue
            
            # Test if pattern is a valid regex (basic test)
            if ! echo "test" | grep -q "$pattern_regex" 2>/dev/null && ! echo "test" | grep -qv "$pattern_regex" 2>/dev/null; then
                ((invalid_patterns++))
            fi
        done < "$PATTERNS_FILE"
        
        if [[ $invalid_patterns -eq 0 ]]; then
            success "All security patterns have valid syntax"
            ((CHECKS_PASSED++))
        else
            error "$invalid_patterns security patterns have invalid syntax"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))
    else
        warning "Security patterns file not found (will be auto-created on first run)"
    fi
    
    echo
}

# Validate script syntax
validate_script_syntax() {
    info "Validating script syntax..."
    
    # Check bash syntax
    validate_check "Security scanner script has valid bash syntax" "bash -n '$SCANNER_SCRIPT'"
    
    # Check shebang
    local shebang
    shebang=$(head -n1 "$SCANNER_SCRIPT" 2>/dev/null || echo "")
    if [[ "$shebang" =~ ^\#\!/usr/bin/env\ bash$ ]]; then
        success "Security scanner script has correct shebang: $shebang"
        ((CHECKS_PASSED++))
    else
        error "Security scanner script has incorrect shebang: $shebang (expected: #!/usr/bin/env bash)"
        ((CHECKS_FAILED++))
    fi
    ((CHECKS_RUN++))
    
    # Check for common issues
    if grep -q "set -euo pipefail" "$SCANNER_SCRIPT"; then
        success "Security scanner script uses strict error handling"
        ((CHECKS_PASSED++))
    else
        warning "Security scanner script should use 'set -euo pipefail' for strict error handling"
    fi
    ((CHECKS_RUN++))
    
    echo
}

# Validate dependencies
validate_dependencies() {
    info "Validating dependencies..."
    
    local required_tools=("jq" "grep" "awk" "mktemp")
    local optional_tools=("gitleaks" "trufflehog" "git-secrets" "shellcheck")
    
    # Check required tools
    for tool in "${required_tools[@]}"; do
        validate_check "Required tool '$tool' is available" "command -v '$tool'"
    done
    
    # Check optional tools
    local available_optional=0
    for tool in "${optional_tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            success "Optional tool '$tool' is available"
            ((available_optional++))
        fi
    done
    
    if [[ $available_optional -gt 0 ]]; then
        info "$available_optional optional security tools are available"
    else
        warning "No optional security tools found - consider installing gitleaks, trufflehog, or git-secrets"
    fi
    
    echo
}

# Validate functionality
validate_functionality() {
    info "Validating functionality..."
    
    # Test with clean content
    local clean_input='{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "How to use React hooks?"}}'
    
    if echo "$clean_input" | "$SCANNER_SCRIPT" >/dev/null 2>&1; then
        success "Clean content test passed"
        ((CHECKS_PASSED++))
    else
        error "Clean content test failed"
        ((CHECKS_FAILED++))
    fi
    ((CHECKS_RUN++))
    
    # Test with sensitive content
    local sensitive_input='{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "API key: sk-test123456789abcdef"}}'
    
    if echo "$sensitive_input" | "$SCANNER_SCRIPT" >/dev/null 2>&1; then
        error "Sensitive content test failed - scanner did not block sensitive data"
        ((CHECKS_FAILED++))
    else
        success "Sensitive content test passed"
        ((CHECKS_PASSED++))
    fi
    ((CHECKS_RUN++))
    
    # Test non-MCP tool (should be allowed)
    local non_mcp_input='{"hook_event_name": "PreToolUse", "tool_name": "Read", "tool_input": {"file_path": "/test"}}'
    
    if echo "$non_mcp_input" | "$SCANNER_SCRIPT" >/dev/null 2>&1; then
        success "Non-MCP tool test passed"
        ((CHECKS_PASSED++))
    else
        error "Non-MCP tool test failed"
        ((CHECKS_FAILED++))
    fi
    ((CHECKS_RUN++))
    
    echo
}

# Show validation results
show_results() {
    echo
    echo "════════════════════════════════════════════════════════════════"
    info "📊 Validation Results Summary"
    echo
    echo "Total checks: $CHECKS_RUN"
    success "Passed: $CHECKS_PASSED"
    
    if [[ $CHECKS_FAILED -gt 0 ]]; then
        error "Failed: $CHECKS_FAILED"
    else
        echo "Failed: $CHECKS_FAILED"
    fi
    
    if [[ $WARNINGS -gt 0 ]]; then
        warning "Warnings: $WARNINGS"
    else
        echo "Warnings: $WARNINGS"
    fi
    
    echo
    
    if [[ $CHECKS_FAILED -eq 0 ]]; then
        success "🎉 All critical validation checks passed!"
        echo
        echo "The MCP Security Scanner is properly configured and ready to use."
        
        if [[ $WARNINGS -gt 0 ]]; then
            echo
            warning "Note: There are $WARNINGS warnings that should be addressed for optimal functionality."
        fi
    else
        error "💥 $CHECKS_FAILED validation checks failed!"
        echo
        echo "Please address the failed checks before using the security scanner."
        echo "Run 'make install' to fix common issues automatically."
    fi
    
    echo "════════════════════════════════════════════════════════════════"
}

# Main validation function
main() {
    show_banner
    
    # Run all validation checks
    validate_file_structure
    validate_permissions
    validate_json_config
    validate_security_patterns
    validate_script_syntax
    validate_dependencies
    validate_functionality
    
    # Show results
    show_results
    
    # Return appropriate exit code
    return $CHECKS_FAILED
}

# Handle help
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat << EOF
MCP Security Scanner Configuration Validator

This script validates the configuration and setup of the MCP Security Scanner.

Usage: $0 [options]

Options:
  -h, --help    Show this help message

The validator checks:
1. File structure and existence
2. File permissions
3. JSON configuration syntax and structure
4. Security patterns file
5. Script syntax and bash compatibility
6. Required and optional dependencies
7. Functional testing with sample inputs

Exit codes:
  0 - All critical checks passed
  N - N critical checks failed (warnings don't affect exit code)

Use 'make validate' to run this script, or 'make install' to fix common issues.

EOF
    exit 0
fi

# Run main validation
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi