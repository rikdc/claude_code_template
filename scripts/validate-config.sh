#!/usr/bin/env bash

# Configuration Validation Script for the security-hooks plugin
# Validates the plugin manifest, hook scripts, and their behaviour

# Deliberately no -e: this script's contract is to run every check and exit
# with the number of failures. Under -e the first failing validate_check
# aborted the run, skipping every later check and the summary.
set -uo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly PROJECT_ROOT
readonly PLUGIN_DIR="$PROJECT_ROOT/plugins/security-hooks"
readonly HOOKS_DIR="$PLUGIN_DIR/hooks"
readonly SCANNER_SCRIPT="$HOOKS_DIR/mcp-security-scanner.sh"
readonly PROTECT_BRANCH_SCRIPT="$HOOKS_DIR/protect-main-branch.sh"
readonly MANIFEST_FILE="$PLUGIN_DIR/.claude-plugin/plugin.json"
readonly PATTERNS_FILE="$PLUGIN_DIR/security-patterns.conf"

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
║              security-hooks Plugin Validator                 ║
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
    validate_check "Plugin directory exists" "[[ -d '$PLUGIN_DIR' ]]"
    validate_check "Hooks directory exists" "[[ -d '$HOOKS_DIR' ]]"
    validate_check "Security scanner script exists" "[[ -f '$SCANNER_SCRIPT' ]]"
    validate_check "Protected branch script exists" "[[ -f '$PROTECT_BRANCH_SCRIPT' ]]"
    validate_check "Plugin manifest exists" "[[ -f '$MANIFEST_FILE' ]]"

    # Optional files
    validate_check "Security patterns file exists" "[[ -f '$PATTERNS_FILE' ]]" "true"
    validate_check "README file exists" "[[ -f '$PLUGIN_DIR/README.md' ]]" "true"
    
    echo
}

# Validate file permissions
validate_permissions() {
    info "Validating file permissions..."
    
    validate_check "Security scanner script is executable" "[[ -x '$SCANNER_SCRIPT' ]]"
    validate_check "Security scanner script is readable" "[[ -r '$SCANNER_SCRIPT' ]]"
    validate_check "Protected branch script is executable" "[[ -x '$PROTECT_BRANCH_SCRIPT' ]]"
    validate_check "Plugin manifest is readable" "[[ -r '$MANIFEST_FILE' ]]"
    
    if [[ -f "$PATTERNS_FILE" ]]; then
        validate_check "Security patterns file is readable" "[[ -r '$PATTERNS_FILE' ]]"
    fi
    
    echo
}

# Validate JSON configuration
validate_json_config() {
    info "Validating JSON configuration..."
    
    # Validate JSON syntax
    if validate_check "Plugin manifest JSON is valid" "jq empty '$MANIFEST_FILE'"; then
        # Validate structure
        validate_check "Manifest declares a plugin name" "jq -e '.name' '$MANIFEST_FILE'"
        validate_check "Manifest has 'hooks' section" "jq -e '.hooks' '$MANIFEST_FILE'"
        validate_check "Manifest has 'PreToolUse' section" "jq -e '.hooks.PreToolUse' '$MANIFEST_FILE'"
        validate_check "Manifest registers both hooks" \
            "[[ \$(jq '.hooks.PreToolUse | length' '$MANIFEST_FILE') -eq 2 ]]"

        # Validate matcher pattern for the MCP scanner
        local matcher_pattern
        matcher_pattern=$(jq -r '.hooks.PreToolUse[0].matcher' "$MANIFEST_FILE" 2>/dev/null || echo "")
        if [[ "$matcher_pattern" == "mcp__.*" ]]; then
            success "Scanner matcher pattern is correct: $matcher_pattern"
            ((CHECKS_PASSED++))
        else
            error "Scanner matcher pattern is incorrect: $matcher_pattern (expected: mcp__.*)"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))

        # Every hook command must resolve through ${CLAUDE_PLUGIN_ROOT}. Earlier
        # revisions used ${PLUGIN_DIR} and ${WORKSPACE}, neither of which Claude
        # Code defines, so they expanded to empty and the paths never resolved.
        local bad_vars
        bad_vars=$(jq -r '[.hooks[][] | .hooks[].command
            | select(contains("${CLAUDE_PLUGIN_ROOT}") | not)] | length' \
            "$MANIFEST_FILE" 2>/dev/null || echo "1")
        if [[ "$bad_vars" == "0" ]]; then
            success "All hook commands resolve through \${CLAUDE_PLUGIN_ROOT}"
            ((CHECKS_PASSED++))
        else
            error "$bad_vars hook command(s) do not use \${CLAUDE_PLUGIN_ROOT}"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))

        # Each referenced script must actually exist on disk
        local missing_scripts=0
        while IFS= read -r cmd; do
            [[ -z "$cmd" ]] && continue
            local resolved="${cmd//\"/}"
            resolved="${resolved/\$\{CLAUDE_PLUGIN_ROOT\}/$PLUGIN_DIR}"
            [[ -f "$resolved" ]] || ((missing_scripts++))
        done < <(jq -r '.hooks[][] | .hooks[].command' "$MANIFEST_FILE" 2>/dev/null)

        if [[ $missing_scripts -eq 0 ]]; then
            success "All hook commands point at existing scripts"
            ((CHECKS_PASSED++))
        else
            error "$missing_scripts hook command(s) reference missing scripts"
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
            
            # Test if pattern is a valid regex (basic test). -e keeps grep from
            # parsing dash-prefixed patterns as options.
            if ! echo "test" | grep -q -e "$pattern_regex" 2>/dev/null && ! echo "test" | grep -qv -e "$pattern_regex" 2>/dev/null; then
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
    
    local script name shebang
    for script in "$SCANNER_SCRIPT" "$PROTECT_BRANCH_SCRIPT"; do
        name=$(basename "$script")

        validate_check "$name has valid bash syntax" "bash -n '$script'"

        shebang=$(head -n1 "$script" 2>/dev/null || echo "")
        if [[ "$shebang" =~ ^\#\!/usr/bin/env\ bash$ ]]; then
            success "$name has correct shebang: $shebang"
            ((CHECKS_PASSED++))
        else
            error "$name has incorrect shebang: $shebang (expected: #!/usr/bin/env bash)"
            ((CHECKS_FAILED++))
        fi
        ((CHECKS_RUN++))

        if grep -q "set -euo pipefail" "$script"; then
            success "$name uses strict error handling"
            ((CHECKS_PASSED++))
        else
            warning "$name should use 'set -euo pipefail' for strict error handling"
        fi
        ((CHECKS_RUN++))
    done
    
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

    # Protected branch hook. A PreToolUse hook signals a block through
    # hookSpecificOutput.permissionDecision, not an exit code, so parse stdout.
    local branch_input='{"hook_event_name": "PreToolUse", "tool_name": "Edit", "tool_input": {"file_path": "/test"}}'
    local decision

    decision=$(echo "$branch_input" | TEST_BRANCH_NAME="main" "$PROTECT_BRANCH_SCRIPT" 2>/dev/null |
        jq -r '.hookSpecificOutput.permissionDecision // empty' 2>/dev/null || echo "")
    if [[ "$decision" == "deny" ]]; then
        success "Protected branch test passed (Edit on main denied)"
        ((CHECKS_PASSED++))
    else
        error "Protected branch test failed - expected deny, got '${decision:-no decision}'"
        ((CHECKS_FAILED++))
    fi
    ((CHECKS_RUN++))

    decision=$(echo "$branch_input" | TEST_BRANCH_NAME="feature/test" "$PROTECT_BRANCH_SCRIPT" 2>/dev/null |
        jq -r '.hookSpecificOutput.permissionDecision // empty' 2>/dev/null || echo "")
    if [[ -z "$decision" ]]; then
        success "Feature branch test passed (Edit on feature branch allowed)"
        ((CHECKS_PASSED++))
    else
        error "Feature branch test failed - expected allow, got '$decision'"
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
    
    # Report the tally directly rather than through warning(), which would
    # increment the counter it is reporting.
    if [[ $WARNINGS -gt 0 ]]; then
        echo -e "${YELLOW}⚠️  Warnings: $WARNINGS${NC}"
    else
        echo "Warnings: $WARNINGS"
    fi
    
    echo
    
    if [[ $CHECKS_FAILED -eq 0 ]]; then
        success "🎉 All critical validation checks passed!"
        echo
        echo "The security-hooks plugin is properly configured and ready to use."
        
        if [[ $WARNINGS -gt 0 ]]; then
            echo
            echo -e "${YELLOW}⚠️  Note: There are $WARNINGS warnings that should be addressed for optimal functionality.${NC}"
        fi
    else
        error "💥 $CHECKS_FAILED validation checks failed!"
        echo
        echo "Please address the failed checks before using the security scanner."
        echo "Run 'make install' to restore executable permissions on the hook scripts."
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
    return "$CHECKS_FAILED"
}

# Handle help
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat << EOF
security-hooks Plugin Configuration Validator

This script validates the configuration and setup of the security-hooks plugin.

Usage: $0 [options]

Options:
  -h, --help    Show this help message

The validator checks:
1. File structure and existence
2. File permissions
3. Plugin manifest syntax, structure, and command paths
4. Security patterns file
5. Script syntax and bash compatibility
6. Required and optional dependencies
7. Functional testing with sample inputs

Exit codes:
  0 - All critical checks passed
  N - N critical checks failed (warnings don't affect exit code)

Use 'make validate' to run this script.

EOF
    exit 0
fi

# Run main validation
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi