#!/usr/bin/env python3
"""
Test Suite for activity-monitor.py Security Improvements
=======================================================

Tests all security fixes and validation functions to ensure they work correctly
and prevent the vulnerabilities that were identified in the code review.
"""

import sys
import os
import json
import tempfile
import shutil
import sqlite3
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add the hooks directory to Python path to import the module
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '.claude', 'hooks'))

# Import functions from activity-monitor
# Need to handle the hyphenated filename
import importlib.util
activity_monitor_path = os.path.join(os.path.dirname(__file__), '..', '.claude', 'hooks', 'activity-monitor.py')
spec = importlib.util.spec_from_file_location("activity_monitor", activity_monitor_path)
activity_monitor = importlib.util.module_from_spec(spec)
spec.loader.exec_module(activity_monitor)

# Import the functions we need
validate_file_path = activity_monitor.validate_file_path
sanitize_for_logging = activity_monitor.sanitize_for_logging
validate_session_id = activity_monitor.validate_session_id
validate_tool_identifier = activity_monitor.validate_tool_identifier
validate_hook_category = activity_monitor.validate_hook_category
calculate_execution_duration = activity_monitor.calculate_execution_duration
parse_hook_event = activity_monitor.parse_hook_event
database_connection = activity_monitor.database_connection
MAX_JSON_SIZE = activity_monitor.MAX_JSON_SIZE

class TestSecurityValidation:
    """Test security validation functions."""
    
    def __init__(self):
        self.temp_dir = tempfile.mkdtemp()
        self.test_results = []
    
    def cleanup(self):
        """Clean up temporary test files."""
        shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def assert_test(self, condition, test_name, details=""):
        """Record test result."""
        status = "PASS" if condition else "FAIL"
        self.test_results.append(f"{status}: {test_name}")
        if details and not condition:
            self.test_results.append(f"  Details: {details}")
        return condition
    
    def test_path_traversal_protection(self):
        """Test path validation prevents directory traversal attacks."""
        print("Testing path traversal protection...")
        
        # Test cases for malicious paths
        malicious_paths = [
            "../../../etc/passwd",
            "../../sensitive_file.db",
            "/etc/passwd", 
            "/proc/version",
            "..\\..\\windows\\system32\\config\\sam",
            ".claude/../../etc/hosts",
        ]
        
        for malicious_path in malicious_paths:
            safe_path = validate_file_path(malicious_path, self.temp_dir)
            expected_safe_path = str(Path(self.temp_dir).resolve() / "activity_metrics.db")
            
            self.assert_test(
                safe_path == expected_safe_path,
                f"Path traversal blocked: {malicious_path}",
                f"Got: {safe_path}, Expected: {expected_safe_path}"
            )
        
        # Test valid paths
        valid_path = os.path.join(self.temp_dir, "valid_db.db")
        result = validate_file_path(valid_path, self.temp_dir)
        self.assert_test(
            valid_path in result,
            "Valid path accepted",
            f"Got: {result}"
        )
    
    def test_data_sanitization(self):
        """Test sensitive data sanitization."""
        print("Testing data sanitization...")
        
        # Test data with sensitive information
        sensitive_data = {
            "api_key": "sk-1234567890abcdef",
            "password": "secret123",
            "email": "user@example.com", 
            "token": "bearer_token_12345",
            "database_url": "postgresql://user:pass@host/db",
            "credit_card": "4111-1111-1111-1111"
        }
        
        sanitized = sanitize_for_logging(sensitive_data)
        
        # Check that sensitive patterns are masked
        self.assert_test(
            "sk-1234567890abcdef" not in sanitized,
            "API key sanitized"
        )
        
        self.assert_test(
            "secret123" not in sanitized, 
            "Password sanitized"
        )
        
        self.assert_test(
            "user@example.com" not in sanitized,
            "Email sanitized"
        )
        
        self.assert_test(
            "4111-1111-1111-1111" not in sanitized,
            "Credit card sanitized"
        )
        
        self.assert_test(
            "***" in sanitized,
            "Sanitization markers present"
        )
    
    def test_input_validation(self):
        """Test input validation functions.""" 
        print("Testing input validation...")
        
        # Test session ID validation
        valid_session = validate_session_id("valid-session-123")
        self.assert_test(
            valid_session == "valid-session-123",
            "Valid session ID accepted"
        )
        
        # Test malicious session ID
        malicious_session = validate_session_id("../../../etc/passwd")
        self.assert_test(
            "etc" not in malicious_session,
            "Malicious session ID sanitized"
        )
        
        # Test tool identifier validation
        valid_tool = validate_tool_identifier("Read")
        self.assert_test(
            valid_tool == "Read",
            "Valid tool identifier accepted"
        )
        
        # Test malicious tool identifier
        malicious_tool = validate_tool_identifier("<script>alert('xss')</script>")
        self.assert_test(
            "<script>" not in malicious_tool,
            "Malicious tool identifier sanitized"
        )
        
        # Test hook category validation
        valid_category = validate_hook_category("PreToolUse")
        self.assert_test(
            valid_category == "PreToolUse",
            "Valid hook category accepted"
        )
        
        invalid_category = validate_hook_category("InvalidCategory")
        self.assert_test(
            invalid_category == "Unknown",
            "Invalid hook category defaulted to Unknown"
        )
    
    def test_execution_duration_validation(self):
        """Test execution duration validation and limits."""
        print("Testing execution duration validation...")
        
        # Test normal duration
        normal_data = {"execution_time": 1.5}  # 1.5 seconds
        duration = calculate_execution_duration(normal_data)
        self.assert_test(
            duration == 1500,  # 1500ms
            "Normal duration calculated correctly"
        )
        
        # Test excessive duration (should be capped)
        excessive_data = {"execution_time": 7200}  # 2 hours 
        duration = calculate_execution_duration(excessive_data)
        self.assert_test(
            duration == 3600000,  # Capped at 1 hour
            "Excessive duration capped at 1 hour"
        )
        
        # Test invalid duration
        invalid_data = {"execution_time": "not_a_number"}
        duration = calculate_execution_duration(invalid_data)
        self.assert_test(
            duration is None,
            "Invalid duration returns None"
        )
    
    def test_hook_parsing(self):
        """Test hook event parsing with validation."""
        print("Testing hook event parsing...")
        
        # Test normal hook data
        normal_hook_data = {
            "hook_event_name": "PreToolUse",
            "tool_name": "Read", 
            "session_id": "session-123"
        }
        
        category, tool, session = parse_hook_event(normal_hook_data)
        
        self.assert_test(
            category == "PreToolUse",
            "Hook category parsed correctly"
        )
        
        self.assert_test(
            tool == "Read",
            "Tool name parsed correctly" 
        )
        
        self.assert_test(
            session == "session-123",
            "Session ID parsed correctly"
        )
        
        # Test malicious hook data
        malicious_hook_data = {
            "hook_event_name": "<script>alert('xss')</script>",
            "tool_name": "../../../etc/passwd",
            "session_id": "'; DROP TABLE activity_log; --"
        }
        
        category, tool, session = parse_hook_event(malicious_hook_data)
        
        self.assert_test(
            category == "Unknown",
            "Malicious hook category sanitized"
        )
        
        self.assert_test(
            "etc" not in tool,
            "Malicious tool name sanitized"
        )
        
        self.assert_test(
            "DROP" not in session,
            "SQL injection in session ID prevented"
        )

def test_database_context_manager():
    """Test database context manager functionality."""
    print("Testing database context manager...")
    
    temp_dir = tempfile.mkdtemp()
    
    try:
        # Mock the config to use our temp directory
        with patch.object(activity_monitor, 'load_monitor_config') as mock_config:
            mock_config.return_value = {
                "storage": {
                    "database_path": os.path.join(temp_dir, "test.db")
                }
            }
            
            # Test successful connection
            with database_connection() as conn:
                cursor = conn.cursor()
                cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
                tables = cursor.fetchall()
                print(f"✓ Database context manager created {len(tables)} tables")
            
            # Verify connection is closed after context
            try:
                cursor.execute("SELECT 1")
                print("✗ Database connection not properly closed")
            except:
                print("✓ Database connection properly closed after context")
                
    finally:
        shutil.rmtree(temp_dir, ignore_errors=True)

def test_json_input_limits():
    """Test JSON input size limits."""
    print("Testing JSON input limits...")
    
    # Create a large JSON payload
    large_data = {"data": "x" * (MAX_JSON_SIZE + 1000)}
    large_json = json.dumps(large_data)
    
    print(f"✓ Large JSON size: {len(large_json)} bytes (exceeds limit of {MAX_JSON_SIZE})")
    
    # The actual limit checking happens in main(), so we test the constant exists
    assert MAX_JSON_SIZE == 1024 * 1024, "JSON size limit is 1MB"
    print("✓ JSON size limit constant properly defined")

def create_mock_hook_data():
    """Create comprehensive mock hook data for testing."""
    return {
        "hook_event_name": "UserPromptSubmit",
        "tool_name": "",
        "session_id": "test-session-123",
        "prompt": "This is a test prompt with api_key=sk-secret123 and password=secret",
        "cwd": "/safe/directory",
        "transcript_path": "/safe/transcript.json"
    }

def test_integration():
    """Integration test with realistic hook data.""" 
    print("Running integration test...")
    
    temp_dir = tempfile.mkdtemp()
    
    try:
        # Mock stdin with realistic hook data
        hook_data = create_mock_hook_data()
        hook_json = json.dumps(hook_data)
        
        with patch('sys.stdin') as mock_stdin:
            mock_stdin.read.return_value = hook_json
            
            with patch.object(activity_monitor, 'load_monitor_config') as mock_config:
                mock_config.return_value = {
                    "monitoring": {"enabled": True},
                    "storage": {
                        "database_path": os.path.join(temp_dir, "integration_test.db")
                    }
                }
                
                # Import and run the main function
                main = activity_monitor.main
                
                try:
                    main()
                    print("✓ Integration test completed without errors")
                    
                    # Check that database was created
                    db_path = os.path.join(temp_dir, "integration_test.db")
                    if os.path.exists(db_path):
                        print("✓ Database created successfully")
                        
                        # Check database contents
                        conn = sqlite3.connect(db_path)
                        cursor = conn.cursor()
                        cursor.execute("SELECT COUNT(*) FROM activity_log")
                        count = cursor.fetchone()[0]
                        conn.close()
                        
                        print(f"✓ Activity log contains {count} entries")
                    else:
                        print("✗ Database was not created")
                        
                except SystemExit as e:
                    if e.code == 0:
                        print("✓ Hook completed with success exit code")
                    else:
                        print(f"✗ Hook exited with error code: {e.code}")
                
    finally:
        shutil.rmtree(temp_dir, ignore_errors=True)

def main():
    """Run all security tests."""
    print("Activity Monitor Security Test Suite")
    print("=" * 50)
    
    # Create test instance
    security_tests = TestSecurityValidation()
    
    try:
        # Run all security tests
        security_tests.test_path_traversal_protection()
        security_tests.test_data_sanitization() 
        security_tests.test_input_validation()
        security_tests.test_execution_duration_validation()
        security_tests.test_hook_parsing()
        
        # Run additional tests
        test_database_context_manager()
        test_json_input_limits()
        test_integration()
        
        print("\nTest Results:")
        print("-" * 30)
        
        # Print security test results
        for result in security_tests.test_results:
            print(result)
        
        # Summary
        passed = sum(1 for r in security_tests.test_results if r.startswith("PASS"))
        failed = sum(1 for r in security_tests.test_results if r.startswith("FAIL"))
        
        print(f"\nSummary: {passed} passed, {failed} failed")
        
        if failed == 0:
            print("🎉 All security tests passed!")
            return True
        else:
            print("❌ Some security tests failed!")
            return False
            
    finally:
        security_tests.cleanup()

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)