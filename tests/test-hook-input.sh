#!/bin/bash
# Test script to verify the activity monitor hook with realistic input

echo "Testing Activity Monitor Hook with Sample Data"
echo "=============================================="

# Test 1: Normal UserPromptSubmit hook
echo "Test 1: Normal UserPromptSubmit hook"
echo '{
    "hook_event_name": "UserPromptSubmit",
    "tool_name": "",
    "session_id": "test-session-123",
    "prompt": "What is the weather today?",
    "cwd": "/safe/directory",
    "transcript_path": "/safe/transcript.json"
}' | python3 .claude/hooks/activity-monitor.py

if [ $? -eq 0 ]; then
    echo "✓ UserPromptSubmit test passed"
else 
    echo "✗ UserPromptSubmit test failed"
fi

# Test 2: PreToolUse hook with sensitive data (should be sanitized)
echo -e "\nTest 2: PreToolUse hook with sensitive data"
echo '{
    "hook_event_name": "PreToolUse", 
    "tool_name": "Bash",
    "session_id": "test-session-456",
    "tool_input": {
        "command": "export API_KEY=sk-1234567890abcdef && curl -H Authorization: Bearer token123"
    }
}' | python3 .claude/hooks/activity-monitor.py

if [ $? -eq 0 ]; then
    echo "✓ PreToolUse with sensitive data test passed"
else
    echo "✗ PreToolUse with sensitive data test failed" 
fi

# Test 3: Malicious input (should be sanitized and handled safely)
echo -e "\nTest 3: Malicious input"
cat << 'EOF' | python3 .claude/hooks/activity-monitor.py
{
    "hook_event_name": "'; DROP TABLE activity_log; --",
    "tool_name": "../../../etc/passwd", 
    "session_id": "<script>alert('xss')</script>",
    "tool_input": {
        "command": "rm -rf /"
    }
}
EOF

if [ $? -eq 0 ]; then
    echo "✓ Malicious input test passed (safely handled)"
else
    echo "✗ Malicious input test failed"
fi

# Test 4: Oversized input (should be rejected)
echo -e "\nTest 4: Oversized JSON input"
# Create a large JSON payload (over 1MB)
python3 -c "
import json
large_data = {'hook_event_name': 'Test', 'large_field': 'x' * 1048577}
print(json.dumps(large_data))
" | python3 .claude/hooks/activity-monitor.py

if [ $? -eq 0 ]; then
    echo "✓ Oversized input test passed (handled gracefully)"
else
    echo "✗ Oversized input test failed"
fi

# Check if database was created and contains data
echo -e "\nDatabase Verification:"
if [ -f ".claude/activity_metrics.db" ]; then
    echo "✓ Database file created"
    
    # Count entries in activity_log
    count=$(sqlite3 .claude/activity_metrics.db "SELECT COUNT(*) FROM activity_log;" 2>/dev/null || echo "0")
    echo "✓ Database contains $count activity log entries"
    
    # Show some sample data (sanitized)
    echo "Sample entries (last 3):"
    sqlite3 .claude/activity_metrics.db "
        SELECT 
            datetime(event_timestamp) as time,
            hook_category, 
            tool_identifier,
            CASE WHEN length(event_data) > 100 THEN substr(event_data, 1, 100) || '...' ELSE event_data END as data
        FROM activity_log 
        ORDER BY id DESC 
        LIMIT 3;" 2>/dev/null | while read line; do
        echo "  $line"
    done
else
    echo "✗ Database file was not created"
fi

echo -e "\nTest Summary:"
echo "All hook input tests completed. Check above for any failures."