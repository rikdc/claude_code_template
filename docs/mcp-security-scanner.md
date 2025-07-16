# MCP Security Scanner Hook

This directory contains a security scanning hook for Claude Code that monitors and audits MCP
requests for sensitive data before they are sent to external services.

## Files

- **`hooks/mcp-security-scanner.sh`** - Main security scanner script
- **`settings.json`** - Claude Code hook configuration
- **`security-patterns.conf.example`** - Example patterns file
- **`security-patterns.conf`** - Your custom patterns (auto-created)
- **`security-scan.log`** - Security scan audit log

## How It Works

The security scanner runs automatically before any MCP tool call (tools starting with `mcp__*`) and:

1. **Extracts content** from the MCP request (code, prompts, queries, thoughts, etc.)
2. **Scans for sensitive data** using regex patterns and external security tools
3. **Logs security violations** when sensitive data is detected
4. **Provides comprehensive audit trail** for all MCP requests

**Important**: The scanner currently provides **monitoring and auditing** rather than blocking.
All MCP requests are logged, and security violations are clearly flagged in the audit log for
review and compliance purposes.

## Setup

The hook is already configured in this repository. When you use Claude Code in this directory,
it will automatically:

- Load the hook configuration from `.claude/settings.json`
- Run the security scanner for all MCP requests
- Create `security-patterns.conf` with default patterns on first run
- Log all activity to `security-scan.log`

**Note**: After making changes to hook configuration, restart Claude Code for changes to take effect.

## Configuration

### Custom Patterns

Edit `.claude/security-patterns.conf` to customize security patterns:

```bash
# Add your own patterns
COMPANY_API_KEY=(?i)(mycompany[_-]?api[_-]?key)["\s]*[:=]["\s]*[a-zA-Z0-9_-]{32,}
INTERNAL_TOKEN=(?i)(internal[_-]?token)["\s]*[:=]["\s]*[^"\s]+
```

### External Security Tools

For enhanced scanning, install these optional tools:

```bash
# macOS
brew install trufflesecurity/trufflehog/trufflehog
brew install gitleaks
brew install git-secrets

# Linux (Ubuntu/Debian)
# Follow installation instructions from their GitHub repositories
```

The scanner automatically detects and uses available tools to enhance pattern detection.

## Testing

### Running the Test Suite

Run the test suite:

```bash
make test
```

### Manual Testing

Test the scanner with a sample MCP request containing sensitive data:

```bash
echo '{
  "hook_event_name": "PreToolUse",
  "tool_name": "mcp__sequential-thinking__sequentialthinking",
  "tool_input": {
    "thought": "API key: sk-test123456789012345678901234567890"
  }
}' | .claude/hooks/mcp-security-scanner.sh
```

Expected output: Security alert message and exit code 1 (indicating violation detected).

### Testing with Live MCP Calls

You can test the hook integration by making MCP calls through Claude Code:

1. **Clean request** (should pass): Normal MCP usage without sensitive data
2. **Sensitive request** (should be logged as violation): Include test API keys, database URLs, etc.

Check `.claude/security-scan.log` to verify the hook is running and detecting patterns correctly.

## Security Patterns Detected

The scanner detects these types of sensitive data by default:

### Authentication & Keys

- API keys and tokens (AWS, GitHub, OpenAI, Slack, Discord, etc.)
- Database connection strings (PostgreSQL, MySQL, MongoDB)
- JWT secrets and passwords
- Private keys and certificates
- SSH keys

### Personal Information

- Email addresses
- Credit card numbers
- Social Security Numbers
- Phone numbers

### Cloud Provider Credentials

- Azure, GCP, Digital Ocean tokens
- Docker and CI/CD credentials
- Payment processor keys (Stripe, PayPal)

## Logs and Monitoring

All security scan activity is logged to `.claude/security-scan.log`:

```bash
# View recent security events
tail -f .claude/security-scan.log

# Search for violations
grep "SECURITY VIOLATION" .claude/security-scan.log

# View hook execution
grep "DEBUG: Hook script started" .claude/security-scan.log

# Check clean requests
grep "Security scan passed" .claude/security-scan.log
```

### Log Entry Types

- **DEBUG entries**: Hook execution and request processing
- **Security scan passed**: Clean requests that contain no sensitive data
- **SECURITY VIOLATION**: Requests containing detected sensitive patterns
- **Pattern violations found**: Specific patterns that were matched

## Troubleshooting

### Hook Not Running

Check if the hook is executing:

```bash
# Look for debug entries in the log
grep "DEBUG: Hook script started" .claude/security-scan.log

# If no entries, check:
# 1. Restart Claude Code (hooks loaded at startup)
# 2. Verify .claude/settings.json uses absolute paths
# 3. Check script permissions: chmod +x .claude/hooks/mcp-security-scanner.sh
```

### Verifying Hook Integration

```bash
# Make a test MCP call with sensitive data and check log:
tail -f .claude/security-scan.log
# Then use Claude Code to make an MCP call
```

### False Positives

- Edit `.claude/security-patterns.conf` to refine patterns
- Use more specific regex patterns to reduce false matches
- Test patterns with sample data before deploying

### False Negatives

- Add new patterns to `.claude/security-patterns.conf`
- Install external security tools for better detection
- Review the pattern coverage for your specific use cases

## Current Behavior vs. Blocking

**Current Implementation**: The security scanner provides comprehensive **monitoring and auditing**:

- ✅ **Detects** sensitive data in all MCP requests
- ✅ **Logs** detailed security violations with timestamps
- ✅ **Audits** all MCP tool calls for compliance
- ✅ **Alerts** when violations are detected
- ⚠️  **Does not block** MCP requests (provides audit trail instead)

This approach ensures:

- **Complete visibility** into data being sent to external MCP services
- **Compliance auditing** with detailed logs
- **Security awareness** without interrupting workflows
- **Pattern refinement** through comprehensive logging

## Customization

The security scanner can be customized by:

1. **Adding patterns** in `security-patterns.conf`
2. **Modifying the script** in `hooks/mcp-security-scanner.sh`
3. **Adjusting hook configuration** in `settings.json`

Example: Monitor only specific MCP tools:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "mcp__context7__.*",
        "hooks": [
          {
            "type": "command",
            "command": "/absolute/path/to/.claude/hooks/mcp-security-scanner.sh"
          }
        ]
      }
    ]
  }
}
```

**Important**: Use absolute paths in `settings.json` for reliable hook execution.

## Security Considerations

- **Patterns are stored in plaintext** - don't put actual secrets in pattern files
- **Logs may contain partial sensitive data** - secure the log file appropriately
- **External tools may send data** to their scanning engines - review their privacy policies
- **Regex patterns can have false positives** - test thoroughly before deploying
- **Hook provides audit trail** - review logs regularly for security violations

## MCP Server Coverage

The scanner monitors all MCP servers including:

- **mcp__sequential-thinking__sequentialthinking**: Monitors `thought` parameter
- **mcp__context7__resolve-library-id**: Monitors `libraryName` parameter
- **mcp__context7__get-library-docs**: Monitors `prompt`, `topic` parameters
- **mcp__magic__***: Monitors component generation requests
- **mcp__playwright__***: Monitors automation and testing requests

The scanner automatically extracts content from various parameter fields to ensure comprehensive coverage.

## Contributing

To improve the security scanner:

1. Test new patterns thoroughly against false positives
2. Consider performance impact of complex regex patterns
3. Document pattern purposes and expected matches
4. Validate MCP parameter extraction for new servers
5. Test hook integration after configuration changes
