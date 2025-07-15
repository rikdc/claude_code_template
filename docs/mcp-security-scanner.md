# MCP Security Scanner Hook

This directory contains a security scanning hook for Claude Code that prevents sensitive data from being sent to external MCP services.

## Files

- **`hooks/mcp-security-scanner.sh`** - Main security scanner script
- **`settings.json`** - Claude Code hook configuration
- **`security-patterns.conf.example`** - Example patterns file
- **`security-patterns.conf`** - Your custom patterns (auto-created)
- **`security-scan.log`** - Security scan audit log

## How It Works

The security scanner runs automatically before any MCP tool call (tools starting with `mcp__*`) and:

1. **Extracts content** from the MCP request (code, prompts, queries, etc.)
2. **Scans for sensitive data** using regex patterns and external tools
3. **Blocks the request** if sensitive data is detected
4. **Logs all activity** for audit purposes

## Setup

The hook is already configured in this repository. When you use Claude Code in this directory, it will automatically:

- Load the hook configuration from `.claude/settings.json`
- Run the security scanner for all MCP requests
- Create `security-patterns.conf` with default patterns on first run

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

## Testing

Test the scanner with a sample MCP request containing sensitive data:

```bash
echo '{
  "hook_event_name": "PreToolUse",
  "tool_name": "mcp__context7__get-library-docs",
  "tool_input": {
    "prompt": "API key: sk-test123456789"
  }
}' | .claude/hooks/mcp-security-scanner.sh
```

Expected output: Security alert and blocked request (exit code 1).

## Security Patterns Detected

The scanner detects these types of sensitive data by default:

### Authentication & Keys

- API keys and tokens (AWS, GitHub, OpenAI, etc.)
- Database connection strings
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
```

## Troubleshooting

### Hook Not Running

- Verify `.claude/settings.json` is properly formatted
- Check that the script is executable: `chmod +x .claude/hooks/mcp-security-scanner.sh`
- Ensure you're running Claude Code from this repository directory

### False Positives

- Edit `.claude/security-patterns.conf` to refine patterns
- Use more specific regex patterns to reduce false matches
- Add comments to document pattern purposes

### False Negatives

- Add new patterns to `.claude/security-patterns.conf`
- Install external security tools for better detection
- Test patterns with sample data before deploying

## Customization

The security scanner can be customized by:

1. **Adding patterns** in `security-patterns.conf`
2. **Modifying the script** in `hooks/mcp-security-scanner.sh`
3. **Adjusting hook configuration** in `settings.json`

Example: Block requests to specific MCP tools only:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "mcp__context7__.*",
        "hooks": [
          {
            "type": "command",
            "command": ".claude/hooks/mcp-security-scanner.sh"
          }
        ]
      }
    ]
  }
}
```

## Security Considerations

- **Patterns are stored in plaintext** - don't put actual secrets in pattern files
- **Logs may contain partial sensitive data** - secure the log file appropriately
- **External tools send data** to their scanning engines - review their privacy policies
- **Regex patterns can have false positives** - test thoroughly before deploying

## Contributing

To improve the security scanner:

1. Test new patterns thoroughly
2. Consider performance impact of complex regex
3. Document pattern purposes and expected matches
4. Validate against common false positive scenarios
