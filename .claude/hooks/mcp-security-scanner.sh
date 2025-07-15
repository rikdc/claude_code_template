#!/usr/bin/env bash

# MCP Security Scanner Hook
# Scans MCP requests for sensitive data before sending to external services
# Compatible with Linux and macOS

set -euo pipefail

# Configuration
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_CLAUDE_DIR="$(dirname "$SCRIPT_DIR")"
readonly LOG_FILE="${PROJECT_CLAUDE_DIR}/security-scan.log"
readonly CONFIG_FILE="${PROJECT_CLAUDE_DIR}/security-patterns.conf"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE"
}

# Initialize configuration if it doesn't exist
init_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" << 'EOF'
# Security patterns configuration
# Format: PATTERN_NAME=regex_pattern

# API Keys and Tokens
API_KEY=(?i)(api[_-]?key|apikey|token|secret)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{16,}
AWS_ACCESS_KEY=AKIA[0-9A-Z]{16}
GITHUB_TOKEN=gh[pousr]_[A-Za-z0-9_]{36,255}
SLACK_TOKEN=xox[baprs]-[0-9]{12}-[0-9]{12}-[a-zA-Z0-9]{24}
DISCORD_TOKEN=[A-Za-z0-9_-]{23,28}\.[A-Za-z0-9_-]{6,7}\.[A-Za-z0-9_-]{27}

# Database Connections
DATABASE_URL=(?i)(database_url|db_url|connection_string)["\s]*[:=]["\s]*[^"\s]+
POSTGRES_URL=postgres(ql)?://[^"\s]+
MYSQL_URL=mysql://[^"\s]+
MONGODB_URL=mongodb(\+srv)?://[^"\s]+

# Passwords and Authentication
PASSWORD=(?i)(password|passwd|pwd)["\s]*[:=]["\s]*[^"\s]{8,}
JWT_SECRET=(?i)(jwt[_-]?secret|secret[_-]?key)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{32,}

# Private Keys
PRIVATE_KEY=-----BEGIN [A-Z]+ PRIVATE KEY-----
SSH_KEY=ssh-(rsa|dsa|ed25519) [A-Za-z0-9+/=]+

# Email and PII (common patterns)
EMAIL=[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}
CREDIT_CARD=\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|3[0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b
SSN=\b(?!000|666|9\d{2})\d{3}-(?!00)\d{2}-(?!0000)\d{4}\b

# Cloud Provider Keys
AZURE_KEY=(?i)(azure|az)[_-]?(key|secret|token|password)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{32,}
GCP_KEY=(?i)(gcp|google)[_-]?(key|secret|token|password)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{32,}

# Custom patterns (user can add more)
CUSTOM_PATTERN=
EOF
        log "Created default security patterns configuration at $CONFIG_FILE"
    fi
}

# Check if external security tools are available
check_security_tools() {
    local tools=()
    
    # Check for truffleHog
    if command -v trufflehog >/dev/null 2>&1; then
        tools+=("trufflehog")
    fi
    
    # Check for gitleaks
    if command -v gitleaks >/dev/null 2>&1; then
        tools+=("gitleaks")
    fi
    
    # Check for git-secrets
    if command -v git-secrets >/dev/null 2>&1; then
        tools+=("git-secrets")
    fi
    
    echo "${tools[@]:-}"
}

# Scan content using external tools
scan_with_tools() {
    local content="$1"
    local temp_file
    temp_file=$(mktemp)
    echo "$content" > "$temp_file"
    
    local tools
    tools=($(check_security_tools))
    
    for tool in "${tools[@]}"; do
        case "$tool" in
            "trufflehog")
                if trufflehog filesystem --no-update --fail "$temp_file" 2>/dev/null; then
                    rm -f "$temp_file"
                    return 1  # Found secrets
                fi
                ;;
            "gitleaks")
                if ! gitleaks detect --source "$temp_file" --no-git 2>/dev/null; then
                    rm -f "$temp_file"
                    return 1  # Found secrets
                fi
                ;;
            "git-secrets")
                if ! git-secrets --scan "$temp_file" 2>/dev/null; then
                    rm -f "$temp_file"
                    return 1  # Found secrets
                fi
                ;;
        esac
    done
    
    rm -f "$temp_file"
    return 0  # No secrets found
}

# Scan content using regex patterns
scan_with_patterns() {
    local content="$1"
    local violations=()
    
    # Load patterns from config
    while IFS='=' read -r pattern_name pattern_regex; do
        # Skip comments and empty lines
        [[ "$pattern_name" =~ ^#.*$ ]] || [[ -z "$pattern_name" ]] && continue
        [[ -z "$pattern_regex" ]] && continue
        
        # Use case-insensitive grep with basic patterns (macOS compatible)
        if echo "$content" | grep -qi "$pattern_regex" 2>/dev/null; then
            violations+=("$pattern_name")
        fi
    done < "$CONFIG_FILE"
    
    if [[ ${#violations[@]} -gt 0 ]]; then
        log "Pattern violations found: ${violations[*]}"
        return 1
    fi
    
    return 0
}

# Main scanning function
scan_content() {
    local content="$1"
    local scan_method="$2"
    
    case "$scan_method" in
        "tools")
            scan_with_tools "$content"
            ;;
        "patterns")
            scan_with_patterns "$content"
            ;;
        "both")
            if ! scan_with_tools "$content"; then
                return 1
            fi
            scan_with_patterns "$content"
            ;;
    esac
}

# Extract content from MCP tool input
extract_mcp_content() {
    local tool_input="$1"
    local content=""
    
    # Extract various fields that might contain sensitive data
    content+=$(echo "$tool_input" | jq -r '.code // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.prompt // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.query // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.content // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.libraryName // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.context7CompatibleLibraryID // empty' 2>/dev/null || echo "")
    content+=$(echo "$tool_input" | jq -r '.topic // empty' 2>/dev/null || echo "")
    
    echo "$content"
}

# Check if tool is MCP-related
is_mcp_tool() {
    local tool_name="$1"
    case "$tool_name" in
        mcp__*) return 0 ;;
        *) return 1 ;;
    esac
}

# Main hook logic
main() {
    local input
    input=$(cat)
    
    # Initialize configuration
    init_config
    
    # Parse JSON input
    local hook_event tool_name tool_input
    hook_event=$(echo "$input" | jq -r '.hook_event_name // empty')
    tool_name=$(echo "$input" | jq -r '.tool_name // empty')
    tool_input=$(echo "$input" | jq -r '.tool_input // empty')
    
    log "Hook event: $hook_event, Tool: $tool_name"
    
    # Only process PreToolUse events for MCP tools
    if [[ "$hook_event" != "PreToolUse" ]] || ! is_mcp_tool "$tool_name"; then
        exit 0  # Allow non-MCP tools to proceed
    fi
    
    # Extract content to scan
    local content
    content=$(extract_mcp_content "$tool_input")
    
    if [[ -z "$content" ]]; then
        log "No content to scan for tool: $tool_name"
        exit 0  # No content to scan
    fi
    
    log "Scanning content for tool: $tool_name (${#content} characters)"
    
    # Determine scan method based on available tools
    local available_tools scan_method
    available_tools=($(check_security_tools))
    
    if [[ ${#available_tools[@]} -gt 0 ]]; then
        scan_method="both"  # Use both tools and patterns
        log "Using security tools: ${available_tools[*]} + patterns"
    else
        scan_method="patterns"  # Fall back to patterns only
        log "Using pattern-based scanning only"
    fi
    
    # Perform security scan
    if ! scan_content "$content" "$scan_method"; then
        log "SECURITY VIOLATION: Sensitive data detected in MCP request to $tool_name"
        echo "🚨 Security Alert: Sensitive data detected in MCP request"
        echo "Tool: $tool_name"
        echo ""
        echo "The request contains potentially sensitive information that should not"
        echo "be sent to external services. Please review the content and remove any:"
        echo "• API keys, tokens, or secrets"
        echo "• Database connection strings"
        echo "• Passwords or authentication credentials"
        echo "• Private keys or certificates"
        echo "• Personal identifiable information (PII)"
        echo ""
        echo "Configure patterns in: $CONFIG_FILE"
        echo "View scan logs in: $LOG_FILE"
        exit 1  # Block the request
    fi
    
    log "Security scan passed for tool: $tool_name"
    exit 0  # Allow the request to proceed
}

# Handle script errors gracefully
trap 'log "Script error on line $LINENO"' ERR

# Run main function
main "$@"