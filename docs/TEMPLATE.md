# [Hook Name] Hook

[Brief description of what this hook does and its primary purpose]

## Files

- **`hooks/[hook-script-name].sh`** - Main hook script
- **`settings.json`** - Claude Code hook configuration
- **`[config-file].conf`** - Configuration file (if applicable)
- **`[log-file].log`** - Hook execution log (if applicable)

## How It Works

[Detailed explanation of the hook's operation, including:]

The [hook name] runs automatically [when/trigger condition] and:

1. **[Primary Action]** - [Description of main functionality]
2. **[Secondary Action]** - [Description of additional functionality]
3. **[Logging/Output]** - [Description of logging or output behavior]

**Important**: [Key behavioral notes, limitations, or important considerations]

## Setup

[Instructions for setting up the hook, including:]

The hook is [already configured/needs configuration] in this repository. When you use Claude Code in this directory, it will automatically:

- [Automatic behavior 1]
- [Automatic behavior 2]
- [File creation/initialization behavior]

**Note**: [Important setup considerations, restart requirements, etc.]

## Configuration

### [Configuration Section 1]

[Description of primary configuration options]

```bash
# Example configuration
[SETTING_NAME]=[example_value]
[ANOTHER_SETTING]=[another_example]
```

### [Configuration Section 2] (Optional)

[Description of secondary configuration options or external dependencies]

```bash
# Installation example (macOS)
brew install [tool-name]

# Installation example (Linux)
# [Installation instructions]
```

[Description of how external tools enhance functionality]

## Testing

### Running the Test Suite

[Instructions for automated testing]

```bash
make test
# or
[test-command]
```

### Manual Testing

[Instructions for manual testing with examples]

```bash
# Test with sample input
echo '[sample-input]' | .claude/hooks/[hook-script-name].sh
```

Expected output: [Description of expected behavior]

### Testing with Live Integration

[Instructions for testing the hook in real Claude Code usage]

1. **[Test Scenario 1]**: [Description and expected outcome]
2. **[Test Scenario 2]**: [Description and expected outcome]

Check `[log-file]` to verify [what to verify].

## [Primary Feature] Detected/Handled

[List of what the hook detects, processes, or handles, organized by category]

### [Category 1]

- [Item 1 with description]
- [Item 2 with description]
- [Item 3 with description]

### [Category 2]

- [Item 1 with description]
- [Item 2 with description]

### [Category 3]

- [Item 1 with description]
- [Item 2 with description]

## Logs and Monitoring

[Description of logging behavior and how to monitor the hook]

All [hook name] activity is logged to `[log-file]`:

```bash
# View recent events
tail -f [log-file]

# Search for [specific events]
grep "[search-pattern]" [log-file]

# View hook execution
grep "[execution-pattern]" [log-file]

# Check [successful operations]
grep "[success-pattern]" [log-file]
```

### Log Entry Types

- **[Log Level 1]**: [Description of when this appears]
- **[Log Level 2]**: [Description of when this appears]
- **[Log Level 3]**: [Description of when this appears]
- **[Log Level 4]**: [Description of when this appears]

## Troubleshooting

### [Common Issue 1]

[Description of the issue and how to identify it]

```bash
# Diagnostic commands
[diagnostic-command-1]
[diagnostic-command-2]

# If no results, check:
# 1. [Check item 1]
# 2. [Check item 2]
# 3. [Check item 3]
```

### [Common Issue 2]

[Description and diagnostic steps]

```bash
# Verification commands
[verification-command]
```

### [Common Issue 3]

- [Solution approach 1]
- [Solution approach 2]
- [Solution approach 3]

### [Common Issue 4]

- [Solution approach 1]
- [Solution approach 2]
- [Solution approach 3]

## Current Behavior vs. [Alternative Behavior]

**Current Implementation**: [Description of current behavior and approach]

- ✅ **[Capability 1]** - [Description]
- ✅ **[Capability 2]** - [Description]
- ✅ **[Capability 3]** - [Description]
- ✅ **[Capability 4]** - [Description]
- ✅ **[Capability 5]** - [Description]

This approach ensures:

- **[Benefit 1]** - [Description]
- **[Benefit 2]** - [Description]
- **[Benefit 3]** - [Description]
- **[Benefit 4]** - [Description]

## Customization

The [hook name] can be customized by:

1. **[Customization Method 1]** in `[config-file]`
2. **[Customization Method 2]** in `[hook-script]`
3. **[Customization Method 3]** in `[settings-file]`

Example: [Specific customization scenario]:

```json
{
  "hooks": {
    "[HookType]": [
      {
        "matcher": "[pattern]",
        "hooks": [
          {
            "type": "command",
            "command": "/absolute/path/to/.claude/hooks/[hook-script-name].sh"
          }
        ]
      }
    ]
  }
}
```

**Important**: [Important notes about customization, such as absolute paths]

## [Domain-Specific Considerations]

[Section specific to the hook's domain, e.g., "Security Considerations", "Performance Considerations", etc.]

- **[Consideration 1]** - [description and implications]
- **[Consideration 2]** - [description and implications]
- **[Consideration 3]** - [description and implications]
- **[Consideration 4]** - [description and implications]
- **[Consideration 5]** - [description and implications]

## [Integration/Coverage Section]

[Description of what the hook integrates with or covers]

The [hook name] [works with/monitors] [list of systems/tools/components]:

- **[Component 1]**: [Description of interaction/coverage]
- **[Component 2]**: [Description of interaction/coverage]
- **[Component 3]**: [Description of interaction/coverage]
- **[Component 4]**: [Description of interaction/coverage]

[Additional details about coverage or integration scope]

## Contributing

To improve the [hook name]:

1. [Contribution guideline 1]
2. [Contribution guideline 2]
3. [Contribution guideline 3]
4. [Contribution guideline 4]
5. [Contribution guideline 5]

---

## Template Usage Notes

**When using this template:**

1. **Replace all bracketed placeholders** `[like this]` with actual content
2. **Remove unused sections** that don't apply to your hook
3. **Add domain-specific sections** as needed for your hook's purpose
4. **Customize the structure** while maintaining the core organization
5. **Include relevant code examples** specific to your hook's functionality
6. **Update file paths and commands** to match your actual implementation

**Section Guidelines:**

- **Files section**: List all files created or used by the hook
- **How It Works**: Explain the hook's operation and trigger conditions
- **Setup**: Provide clear installation and initialization instructions
- **Configuration**: Document all configuration options with examples
- **Testing**: Include both automated and manual testing procedures
- **Troubleshooting**: Address common issues with diagnostic steps
- **Customization**: Show how users can adapt the hook for their needs
- **Domain Considerations**: Address security, performance, or other relevant concerns