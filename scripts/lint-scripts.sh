#!/usr/bin/env bash

# Linting Script for MCP Security Scanner
# Runs ShellCheck and other code quality checks on all shell scripts

set -euo pipefail

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
readonly GREEN='\033[0;32m'
readonly BLUE='\033[0;34m'
readonly YELLOW='\033[1;33m'
readonly RED='\033[0;31m'
readonly NC='\033[0m'

# Counters
SCRIPTS_CHECKED=0
SCRIPTS_PASSED=0
SCRIPTS_FAILED=0
ISSUES_FOUND=0

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

# Show linting banner
show_banner() {
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════╗
║                 MCP Security Scanner Linter                 ║
║                                                              ║
║  Running code quality checks on all shell scripts          ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo
}

# Check if ShellCheck is available
check_shellcheck() {
    if ! command -v shellcheck >/dev/null 2>&1; then
        error "ShellCheck is not installed"
        echo
        echo "Install ShellCheck:"
        echo "  macOS: brew install shellcheck"
        echo "  Ubuntu/Debian: apt-get install shellcheck"
        echo "  Other: https://github.com/koalaman/shellcheck#installing"
        echo
        return 1
    fi
    
    local version
    version=$(shellcheck --version | head -3 | tail -1 | awk '{print $2}')
    success "ShellCheck $version is available"
    return 0
}

# Find all shell scripts
find_shell_scripts() {
    local scripts=()
    
    # Find shell scripts by extension and shebang
    while IFS= read -r -d '' script; do
        scripts+=("$script")
    done < <(find "$PROJECT_ROOT" -type f \( -name "*.sh" -o -name "*.bash" \) -print0)
    
    # Find files with bash shebang but no extension
    while IFS= read -r -d '' script; do
        if [[ ! "$script" =~ \.(sh|bash)$ ]]; then
            scripts+=("$script")
        fi
    done < <(find "$PROJECT_ROOT" -type f -executable -exec grep -l '^#!/.*bash' {} \; -print0 2>/dev/null)
    
    # Remove duplicates and sort
    printf '%s\n' "${scripts[@]}" | sort -u
}

# Run ShellCheck on a single script
lint_script() {
    local script="$1"
    local relative_path
    relative_path=$(realpath --relative-to="$PROJECT_ROOT" "$script" 2>/dev/null || basename "$script")
    
    ((SCRIPTS_CHECKED++))
    
    info "Checking $relative_path..."
    
    # Create temporary output file
    local output_file
    output_file=$(mktemp)
    
    # Run ShellCheck with specific configuration
    local shellcheck_exit=0
    shellcheck \
        --external-sources \
        --source-path="$PROJECT_ROOT" \
        --format=gcc \
        "$script" > "$output_file" 2>&1 || shellcheck_exit=$?
    
    # Process results
    local issue_count
    issue_count=$(wc -l < "$output_file")
    
    if [[ $shellcheck_exit -eq 0 ]]; then
        success "  No issues found"
        ((SCRIPTS_PASSED++))
    else
        error "  $issue_count issues found"
        ((SCRIPTS_FAILED++))
        ((ISSUES_FOUND += issue_count))
        
        # Show issues with proper formatting
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Format: file:line:column: level: message [SC####]
                if [[ "$line" =~ ^([^:]+):([0-9]+):([0-9]+):\ (warning|error|info|note):\ (.+)$ ]]; then
                    local file="${BASH_REMATCH[1]}"
                    local line_num="${BASH_REMATCH[2]}"
                    local col_num="${BASH_REMATCH[3]}"
                    local level="${BASH_REMATCH[4]}"
                    local message="${BASH_REMATCH[5]}"
                    
                    case "$level" in
                        "error")
                            echo -e "    ${RED}[ERROR]${NC} Line $line_num:$col_num - $message"
                            ;;
                        "warning")
                            echo -e "    ${YELLOW}[WARN]${NC}  Line $line_num:$col_num - $message"
                            ;;
                        "info"|"note")
                            echo -e "    ${BLUE}[INFO]${NC}  Line $line_num:$col_num - $message"
                            ;;
                    esac
                else
                    echo "    $line"
                fi
            fi
        done < "$output_file"
        echo
    fi
    
    rm -f "$output_file"
}

# Run additional checks
additional_checks() {
    info "Running additional code quality checks..."
    
    local additional_issues=0
    
    # Check for common issues
    info "Checking for common shell scripting issues..."
    
    # Check for use of `set -e` or `set -euo pipefail`
    local scripts_without_strict_mode=()
    while IFS= read -r script; do
        if ! grep -q "set -e" "$script"; then
            scripts_without_strict_mode+=("$(basename "$script")")
        fi
    done < <(find_shell_scripts)
    
    if [[ ${#scripts_without_strict_mode[@]} -gt 0 ]]; then
        warning "Scripts without strict error handling (set -e): ${scripts_without_strict_mode[*]}"
        ((additional_issues++))
    else
        success "All scripts use strict error handling"
    fi
    
    # Check for executable permissions
    local non_executable_scripts=()
    while IFS= read -r script; do
        if [[ ! -x "$script" ]]; then
            non_executable_scripts+=("$(basename "$script")")
        fi
    done < <(find_shell_scripts)
    
    if [[ ${#non_executable_scripts[@]} -gt 0 ]]; then
        warning "Scripts without executable permissions: ${non_executable_scripts[*]}"
        ((additional_issues++))
    else
        success "All scripts have executable permissions"
    fi
    
    # Check for proper shebang
    local scripts_with_bad_shebang=()
    while IFS= read -r script; do
        local shebang
        shebang=$(head -n1 "$script")
        if [[ ! "$shebang" =~ ^\#\!/usr/bin/env\ bash$ ]] && [[ ! "$shebang" =~ ^\#\!/bin/bash$ ]]; then
            scripts_with_bad_shebang+=("$(basename "$script")")
        fi
    done < <(find_shell_scripts)
    
    if [[ ${#scripts_with_bad_shebang[@]} -gt 0 ]]; then
        warning "Scripts with non-standard shebang: ${scripts_with_bad_shebang[*]}"
        warning "Recommended: #!/usr/bin/env bash"
        ((additional_issues++))
    else
        success "All scripts have proper shebang"
    fi
    
    # Check for TODO/FIXME comments
    local todo_count
    todo_count=$(grep -r "TODO\|FIXME\|XXX" --include="*.sh" "$PROJECT_ROOT" | wc -l)
    
    if [[ $todo_count -gt 0 ]]; then
        info "$todo_count TODO/FIXME comments found in shell scripts"
    fi
    
    return $additional_issues
}

# Generate lint report
generate_report() {
    local report_file="$PROJECT_ROOT/lint-report.txt"
    
    cat > "$report_file" << EOF
MCP Security Scanner - Lint Report
Generated: $(date)

Summary:
  Scripts checked: $SCRIPTS_CHECKED
  Scripts passed: $SCRIPTS_PASSED
  Scripts failed: $SCRIPTS_FAILED
  Total issues: $ISSUES_FOUND

Scripts analyzed:
EOF
    
    find_shell_scripts | while IFS= read -r script; do
        local relative_path
        relative_path=$(realpath --relative-to="$PROJECT_ROOT" "$script" 2>/dev/null || basename "$script")
        echo "  - $relative_path" >> "$report_file"
    done
    
    echo >> "$report_file"
    echo "Run 'shellcheck <script>' for detailed analysis of specific issues." >> "$report_file"
    
    info "Lint report saved to: $report_file"
}

# Show results summary
show_results() {
    echo
    echo "════════════════════════════════════════════════════════════════"
    info "📊 Linting Results Summary"
    echo
    echo "Scripts checked: $SCRIPTS_CHECKED"
    success "Scripts passed: $SCRIPTS_PASSED"
    
    if [[ $SCRIPTS_FAILED -gt 0 ]]; then
        error "Scripts failed: $SCRIPTS_FAILED"
        error "Total issues: $ISSUES_FOUND"
    else
        echo "Scripts failed: $SCRIPTS_FAILED"
        echo "Total issues: $ISSUES_FOUND"
    fi
    
    echo
    
    if [[ $SCRIPTS_FAILED -eq 0 ]]; then
        success "🎉 All shell scripts passed linting checks!"
        echo
        echo "Your code follows shell scripting best practices."
    else
        error "💥 $SCRIPTS_FAILED shell scripts have linting issues!"
        echo
        echo "Please address the issues shown above."
        echo "Run 'shellcheck <script>' for detailed analysis."
    fi
    
    echo "════════════════════════════════════════════════════════════════"
}

# Main linting function
main() {
    local generate_report_flag=false
    local verbose=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --report)
                generate_report_flag=true
                shift
                ;;
            --verbose|-v)
                verbose=true
                shift
                ;;
            *)
                warning "Unknown option: $1"
                shift
                ;;
        esac
    done
    
    show_banner
    
    # Check prerequisites
    check_shellcheck || return 1
    
    # Find scripts to check
    local scripts
    mapfile -t scripts < <(find_shell_scripts)
    
    if [[ ${#scripts[@]} -eq 0 ]]; then
        warning "No shell scripts found to check"
        return 0
    fi
    
    info "Found ${#scripts[@]} shell scripts to check"
    echo
    
    # Run ShellCheck on each script
    for script in "${scripts[@]}"; do
        lint_script "$script"
    done
    
    # Run additional checks
    additional_checks
    
    # Generate report if requested
    if [[ "$generate_report_flag" == "true" ]]; then
        generate_report
    fi
    
    # Show results
    show_results
    
    # Return appropriate exit code
    return $SCRIPTS_FAILED
}

# Handle help
if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    cat << EOF
MCP Security Scanner Linter

Runs ShellCheck and additional code quality checks on all shell scripts.

Usage: $0 [options]

Options:
  --report        Generate a detailed lint report
  --verbose, -v   Show verbose output
  -h, --help      Show this help message

The linter checks:
1. ShellCheck analysis (syntax, best practices, common bugs)
2. Proper error handling (set -e usage)
3. Executable permissions
4. Shebang consistency
5. TODO/FIXME comment counts

Exit codes:
  0 - All scripts passed linting
  N - N scripts failed linting

Use 'make lint' to run this script.

EOF
    exit 0
fi

# Run main linting
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi