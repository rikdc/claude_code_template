# Activity Monitor Testing

This directory contains comprehensive tests for the activity monitor security improvements.

## Test Files

### `test-activity-monitor.py`
Comprehensive Python test suite that validates:
- **Security Validation Functions**: Path traversal protection, data sanitization, input validation
- **Database Context Manager**: Proper connection handling and cleanup
- **Input Limits**: JSON size limits and execution duration caps
- **Integration**: End-to-end testing with mock hook data

### `test-hook-input.sh`
Shell script that tests the hook with realistic JSON input:
- Normal hook events (UserPromptSubmit, PreToolUse)
- Sensitive data sanitization
- Malicious input handling
- Oversized input rejection

## Running Tests

### Security Test Suite
```bash
python3 tests/test-activity-monitor.py
```

### Hook Input Tests
```bash
./tests/test-hook-input.sh
```

### Manual Testing

You can manually test the hook by piping JSON to it:

```bash
echo '{"hook_event_name": "UserPromptSubmit", "prompt": "test"}' | \
  python3 .claude/hooks/activity-monitor.py
```

## Security Improvements Tested

✅ **SQL Injection Prevention**: Parameterized queries prevent injection attacks  
✅ **Path Traversal Protection**: File paths are validated and sanitized  
✅ **Data Sanitization**: Sensitive information is masked in logs and database  
✅ **Input Validation**: All user inputs are validated and sanitized  
✅ **Size Limits**: JSON input is limited to 1MB to prevent DoS attacks  
✅ **Context Management**: Database connections use proper cleanup patterns  

## Expected Results

All tests should pass with output showing:
- Path traversal attempts are blocked
- Sensitive data (API keys, passwords, emails) is sanitized
- Malicious inputs are handled safely
- Database operations work correctly
- Integration tests complete successfully

The activity database (`.claude/activity_metrics.db`) should contain sanitized entries with no exposed sensitive information.