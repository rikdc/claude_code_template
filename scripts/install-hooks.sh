#!/usr/bin/env bash

# Install Script for MCP Security Scanner
# Sets up the security scanner in the current project

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
readonly PROJECT_ROOT
readonly CLAUDE_DIR="$PROJECT_ROOT/.claude"
readonly HOOKS_DIR="$CLAUDE_DIR/hooks"
readonly SCANNER_SCRIPT="$HOOKS_DIR/mcp-security-scanner.sh"
readonly SETTINGS_FILE="$CLAUDE_DIR/settings.json"

# Colors for output
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly YELLOW='\033[1;33m'
readonly RED='\033[0;31m'
readonly NC='\033[0m'

# Logging functions
info() {
    echo -e "${BLUE}ℹ️  $*${NC}"
}

success() {
    echo -e "${GREEN}✅ $*${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $*${NC}"
}

error() {
    echo -e "${RED}❌ $*${NC}"
}

# Show installation banner
show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                MCP Security Scanner Installer                ║
║                                                              ║
║  Installing Claude Code security hooks for this project     ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo
}

# Check if Claude Code is installed
check_claude_code() {
    info "Checking Claude Code installation..."
    
    if command -v claude >/dev/null 2>&1; then
        local version
        version=$(claude --version 2>/dev/null | head -1 || echo "unknown")
        success "Claude Code is installed: $version"
        return 0
    else
        warning "Claude Code CLI not found in PATH"
        warning "This is optional - the hooks will still work when Claude Code is installed"
        return 0
    fi
}

# Verify project structure
verify_project_structure() {
    info "Verifying project structure..."
    
    local missing_files=()
    
    # Check for required files
    [[ ! -f "$SCANNER_SCRIPT" ]] && missing_files+=("Security scanner script")
    [[ ! -f "$SETTINGS_FILE" ]] && missing_files+=("Hook settings file")
    
    if [[ ${#missing_files[@]} -gt 0 ]]; then
        error "Missing required files:"
        for file in "${missing_files[@]}"; do
            echo "  - $file"
        done
        return 1
    fi
    
    success "Project structure is valid"
    return 0
}

# Set up hook permissions
setup_permissions() {
    info "Setting up file permissions..."
    
    # Make scanner script executable
    if [[ -f "$SCANNER_SCRIPT" ]]; then
        chmod +x "$SCANNER_SCRIPT"
        success "Security scanner script is executable"
    else
        error "Security scanner script not found: $SCANNER_SCRIPT"
        return 1
    fi
    
    # Make test scripts executable
    if [[ -d "$PROJECT_ROOT/tests" ]]; then
        find "$PROJECT_ROOT/tests" -name "*.sh" -type f -exec chmod +x {} \;
        success "Test scripts are executable"
    fi
    
    return 0
}

# Validate configuration files
validate_configuration() {
    info "Validating configuration files..."
    
    # Validate settings.json
    if ! jq empty "$SETTINGS_FILE" 2>/dev/null; then
        error "Invalid JSON in settings file: $SETTINGS_FILE"
        return 1
    fi
    
    # Check hook configuration structure
    local has_hooks has_pretooluse has_matcher has_command
    has_hooks=$(jq 'has("hooks")' "$SETTINGS_FILE")
    has_pretooluse=$(jq '.hooks | has("PreToolUse")' "$SETTINGS_FILE")
    has_matcher=$(jq '.hooks.PreToolUse[0] | has("matcher")' "$SETTINGS_FILE")
    has_command=$(jq '.hooks.PreToolUse[0].hooks[0] | has("command")' "$SETTINGS_FILE")
    
    if [[ "$has_hooks" != "true" || "$has_pretooluse" != "true" || "$has_matcher" != "true" || "$has_command" != "true" ]]; then
        error "Invalid hook configuration structure in: $SETTINGS_FILE"
        return 1
    fi
    
    success "Configuration files are valid"
    return 0
}

# Check dependencies
check_dependencies() {
    info "Checking dependencies..."
    
    local required_tools=("jq" "grep" "awk")
    local missing_tools=()
    local optional_tools=("gitleaks" "trufflehog" "git-secrets" "shellcheck")
    local available_optional=()
    
    # Check required tools
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error "Missing required tools: ${missing_tools[*]}"
        info "Install missing tools and try again"
        return 1
    fi
    
    success "All required tools are available"
    
    # Check optional tools
    for tool in "${optional_tools[@]}"; do
        if command -v "$tool" >/dev/null 2>&1; then
            available_optional+=("$tool")
        fi
    done
    
    if [[ ${#available_optional[@]} -gt 0 ]]; then
        success "Available optional tools: ${available_optional[*]}"
    else
        warning "No optional security tools found"
        warning "Consider installing gitleaks, trufflehog, or git-secrets for enhanced scanning"
    fi
    
    return 0
}

# Test installation
test_installation() {
    info "Testing installation..."
    
    # Test with clean content (should pass)
    local clean_test='{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "How to use React hooks?"}}'
    
    if echo "$clean_test" | "$SCANNER_SCRIPT" >/dev/null 2>&1; then
        success "Clean content test passed"
    else
        error "Clean content test failed"
        return 1
    fi
    
    # Test with sensitive content (should fail/block)
    local sensitive_test='{"hook_event_name": "PreToolUse", "tool_name": "mcp__context7__get-library-docs", "tool_input": {"prompt": "API key: sk-test123456789abcdef"}}'
    
    if echo "$sensitive_test" | "$SCANNER_SCRIPT" >/dev/null 2>&1; then
        error "Sensitive content test failed - scanner did not block sensitive data"
        return 1
    else
        success "Sensitive content test passed"
    fi
    
    return 0
}

# Show post-installation information
show_post_install_info() {
    success "🎉 MCP Security Scanner installation completed successfully!"
    echo
    echo "📁 Installation Details:"
    echo "  • Security scanner: $SCANNER_SCRIPT"
    echo "  • Hook configuration: $SETTINGS_FILE"
    echo "  • Patterns file: $CLAUDE_DIR/security-patterns.conf (auto-created)"
    echo "  • Logs location: $CLAUDE_DIR/security-scan.log"
    echo
    echo "🛡️  The security scanner will now:"
    echo "  • Automatically scan all MCP requests for sensitive data"
    echo "  • Block requests containing API keys, passwords, private keys, etc."
    echo "  • Log all security scan activity for audit purposes"
    echo "  • Allow clean content to pass through normally"
    echo
    echo "🧪 Testing:"
    echo "  • Run 'make test' to verify everything is working"
    echo "  • Run 'make demo' to see the scanner in action"
    echo "  • Run 'make status' to check current configuration"
    echo
    echo "⚙️  Customization:"
    echo "  • Edit $CLAUDE_DIR/security-patterns.conf to customize detection patterns"
    echo "  • View logs with 'make logs' or check $CLAUDE_DIR/security-scan.log"
    echo "  • Run 'make help' to see all available commands"
    echo
    info "The security scanner is now active and protecting your MCP requests! 🛡️"
}

# Main installation function
main() {
    show_banner
    
    # Run installation steps
    check_claude_code
    verify_project_structure || return 1
    setup_permissions || return 1
    validate_configuration || return 1
    check_dependencies || return 1
    test_installation || return 1
    
    # Show success information
    show_post_install_info
    
    return 0
}

# Handle help
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat << EOF
MCP Security Scanner Installer

This script installs and configures the MCP Security Scanner for Claude Code.

Usage: $0 [options]

Options:
  -h, --help    Show this help message

The installer will:
1. Check for Claude Code installation (optional)
2. Verify project structure and required files
3. Set up proper file permissions
4. Validate configuration files
5. Check for required dependencies
6. Test the installation
7. Display setup information

After installation, use 'make test' to run the test suite and 'make demo' 
to see the security scanner in action.

EOF
    exit 0
fi

# Run main installation
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi