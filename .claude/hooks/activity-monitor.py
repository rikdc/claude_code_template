#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.9"
# dependencies = [
#     "toml",
# ]
# ///

"""
Claude Code Activity Monitor Hook
=================================

Monitors and records Claude Code hook activity using a functional approach.
Captures tool usage, user interactions, and system metrics.

Hook Types: PreToolUse, PostToolUse, UserPromptSubmit, Stop
"""

import json
import os
import sys
import uuid
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple
from contextlib import contextmanager
import sqlite3

try:
    import toml
except ImportError:
    toml = None

# Security and Configuration Constants
MAX_JSON_SIZE = 1024 * 1024  # 1MB limit for JSON input
MAX_LOG_ENTRIES = 500
DEFAULT_CLEANUP_DAYS = 30
MAX_PROMPT_SNIPPET_LENGTH = 50
MAX_COMMAND_SNIPPET_LENGTH = 100

def validate_file_path(file_path: str, base_dir: str = ".") -> str:
    """Validate and sanitize file paths to prevent directory traversal."""
    try:
        # Resolve the path to an absolute path
        resolved_path = Path(file_path).resolve()
        base_path = Path(base_dir).resolve()
        
        # Check if the resolved path is within the allowed base directory
        try:
            resolved_path.relative_to(base_path)
        except ValueError:
            # Path is outside base directory - use safe default
            return str(base_path / "activity_metrics.db")
        
        # Additional security checks
        path_str = str(resolved_path)
        
        # Block suspicious patterns
        suspicious_patterns = ['..', '/etc/', '/proc/', '/sys/', '/dev/']
        if any(pattern in path_str for pattern in suspicious_patterns):
            return str(base_path / "activity_metrics.db")
        
        return path_str
        
    except Exception:
        # Fall back to safe default on any error
        return str(Path(base_dir).resolve() / "activity_metrics.db")

def sanitize_for_logging(data: Any, max_length: int = 100) -> str:
    """Sanitize data for safe logging by removing sensitive information."""
    if data is None:
        return "None"
    
    # Convert to string
    data_str = str(data)
    
    # Remove common sensitive patterns
    sensitive_patterns = [
        # API keys (like sk-1234567890abcdef)
        (r'(?i)\b(sk-[a-zA-Z0-9]{20,})', '***'),
        # Generic API keys, tokens, passwords, secrets
        (r'(?i)(api[_-]?key|token|password|secret|credential)["\']?\s*[:=]\s*["\']?([^\s"\',}\n\r]+)', r'\1=***'),
        # Bearer tokens
        (r'(?i)(bearer\s+)([a-zA-Z0-9\-_]{10,})', r'\1***'),
        # Database URLs
        (r'(?i)(postgresql|mysql|mongodb)://[^\s]+', 'DATABASE_URL=***'),
        # Email addresses
        (r'(?i)\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}\b', '***@***.***'),
        # Credit card numbers
        (r'(?i)\b[0-9]{4}\s*-?\s*[0-9]{4}\s*-?\s*[0-9]{4}\s*-?\s*[0-9]{4}\b', '****-****-****-****'),
    ]
    
    import re
    for pattern, replacement in sensitive_patterns:
        data_str = re.sub(pattern, replacement, data_str)
    
    # Truncate to max length
    if len(data_str) > max_length:
        data_str = data_str[:max_length] + "...[truncated]"
    
    return data_str

def validate_session_id(session_id: str) -> str:
    """Validate and sanitize session ID."""
    if not session_id or not isinstance(session_id, str):
        return str(uuid.uuid4())
    
    # Check for suspicious patterns first
    import re
    suspicious_patterns = ['./', '../', '\\', 'DROP', 'SELECT', 'INSERT', 'DELETE', '<script>', '&lt;']
    session_lower = session_id.lower()
    
    for pattern in suspicious_patterns:
        if pattern.lower() in session_lower:
            return str(uuid.uuid4())
    
    # Remove any non-alphanumeric characters except hyphens and underscores
    sanitized = re.sub(r'[^a-zA-Z0-9\-_]', '', session_id)
    
    # Ensure reasonable length
    if len(sanitized) < 5 or len(sanitized) > 100:
        return str(uuid.uuid4())
    
    return sanitized

def validate_tool_identifier(tool_id: str) -> str:
    """Validate and sanitize tool identifier."""
    if not tool_id or not isinstance(tool_id, str):
        return "unknown"
    
    # Check for suspicious patterns first  
    import re
    suspicious_patterns = ['./', '../', '\\', '<', '>', 'script', 'DROP', 'SELECT']
    tool_lower = tool_id.lower()
    
    for pattern in suspicious_patterns:
        if pattern.lower() in tool_lower:
            return "unknown"
    
    # Remove any non-alphanumeric characters except underscores and hyphens
    sanitized = re.sub(r'[^a-zA-Z0-9_\-]', '', tool_id)
    
    # Ensure reasonable length
    if len(sanitized) > 100:
        sanitized = sanitized[:100]
    
    return sanitized or "unknown"

def validate_hook_category(hook_category: str) -> str:
    """Validate hook category against allowed values."""
    allowed_categories = {
        "PreToolUse", "PostToolUse", "UserPromptSubmit", "Stop", 
        "FilterMatch", "Unknown"
    }
    
    if not hook_category or not isinstance(hook_category, str):
        return "Unknown"
    
    # Check if it's in allowed list
    if hook_category in allowed_categories:
        return hook_category
    
    # For backward compatibility, allow hook categories that start with known prefixes
    for allowed in allowed_categories:
        if hook_category.startswith(allowed):
            return allowed
    
    return "Unknown"

@contextmanager
def database_connection():
    """Context manager for database connections with proper cleanup."""
    conn = None
    try:
        config = load_monitor_config()
        db_path = config.get("storage", {}).get("database_path", ".claude/activity_metrics.db")
        
        # Validate and sanitize the database path
        safe_db_path = validate_file_path(db_path, ".claude")
        
        # Ensure directory exists with safe path
        safe_dir = os.path.dirname(safe_db_path)
        if safe_dir:
            os.makedirs(safe_dir, exist_ok=True)
        
        conn = sqlite3.connect(safe_db_path)
        initialize_database_schema(conn)
        yield conn
        
    except Exception as e:
        if conn:
            conn.rollback()
        raise e
    finally:
        if conn:
            conn.close()

# Configuration management
def load_monitor_config(config_file: str = ".claude/activity-monitor.toml") -> Dict[str, Any]:
    """Load monitoring configuration from TOML file."""
    try:
        config_path = Path(config_file)
        if not config_path.exists():
            return get_default_config()
        
        if toml is None:
            print("Warning: TOML library not available, using defaults", file=sys.stderr)
            return get_default_config()
            
        return toml.load(config_path)
    except Exception:
        return get_default_config()

def get_default_config() -> Dict[str, Any]:
    """Return default configuration."""
    return {
        "monitoring": {
            "enabled": True,
            "capture_tool_usage": True,
            "capture_user_prompts": True,
            "excluded_tools": [],
            "hook_types": []
        },
        "storage": {
            "database_path": ".claude/activity_metrics.db",
            "backup_enabled": False,
            "cleanup_days": 30
        },
        "analytics": {
            "enable_insights": True,
            "daily_summaries": True,
            "performance_tracking": True
        },
        "filters": [
            {
                "name": "large_operations", 
                "condition": "content_size > 50000",
                "description": "Track large content operations"
            },
            {
                "name": "slow_commands",
                "condition": "execution_time > 5000", 
                "description": "Track slow command executions"
            }
        ]
    }

# Database operations
def get_database_connection():
    """Get database connection with proper setup (legacy method).
    
    DEPRECATED: Use database_connection() context manager instead.
    """
    config = load_monitor_config()
    db_path = config.get("storage", {}).get("database_path", ".claude/activity_metrics.db")
    
    # Validate and sanitize the database path
    safe_db_path = validate_file_path(db_path, ".claude")
    
    # Ensure directory exists with safe path
    safe_dir = os.path.dirname(safe_db_path)
    if safe_dir:
        os.makedirs(safe_dir, exist_ok=True)
    
    conn = sqlite3.connect(safe_db_path)
    initialize_database_schema(conn)
    return conn

def initialize_database_schema(conn):
    """Initialize database with our unique schema."""
    cursor = conn.cursor()
    
    # Main activity log table
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS activity_log (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            event_timestamp TEXT NOT NULL,
            session_identifier TEXT NOT NULL,
            hook_category TEXT NOT NULL,
            tool_identifier TEXT,
            execution_duration INTEGER,
            event_data TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    
    # Session summary table  
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS session_summary (
            session_identifier TEXT PRIMARY KEY,
            start_timestamp TEXT NOT NULL,
            end_timestamp TEXT,
            total_events INTEGER DEFAULT 0,
            total_execution_time INTEGER DEFAULT 0,
            primary_tool TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    
    # Insights table for analytics
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS activity_insights (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            insight_date DATE NOT NULL,
            insight_type TEXT NOT NULL,
            insight_data TEXT NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    ''')
    
    # Create indexes
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_activity_session ON activity_log(session_identifier)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_activity_timestamp ON activity_log(event_timestamp)')
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_insights_date ON activity_insights(insight_date)')
    
    conn.commit()

# Hook data processing functions
def parse_hook_event(hook_data: Dict[str, Any]) -> Tuple[str, str, str]:
    """Parse hook event to extract key identifiers."""
    # Claude Code uses hook_event_name and tool_name
    hook_category = hook_data.get("hook_event_name") or hook_data.get("hookType", "Unknown")
    tool_identifier = hook_data.get("tool_name") or hook_data.get("toolName", "")
    session_id = hook_data.get("session_id", str(uuid.uuid4()))
    
    # Validate and sanitize all inputs
    hook_category = validate_hook_category(hook_category)
    tool_identifier = validate_tool_identifier(tool_identifier)
    session_id = validate_session_id(session_id)
    
    return hook_category, tool_identifier, session_id

def extract_event_metadata(hook_data: Dict[str, Any], hook_category: str, tool_identifier: str) -> Dict[str, Any]:
    """Extract and structure event metadata."""
    metadata = {
        "hook_category": hook_category,
        "tool_identifier": tool_identifier
    }
    
    # Handle UserPromptSubmit events
    if hook_category == "UserPromptSubmit":
        if "prompt" in hook_data:
            metadata["prompt_character_count"] = len(str(hook_data["prompt"]))
            prompt_str = str(hook_data["prompt"])
            # Sanitize prompt snippet for storage
            sanitized_prompt = sanitize_for_logging(prompt_str, MAX_PROMPT_SNIPPET_LENGTH)
            metadata["prompt_snippet"] = sanitized_prompt
        
        if "transcript_path" in hook_data:
            metadata["transcript_location"] = hook_data["transcript_path"]
        
        if "cwd" in hook_data:
            metadata["working_path"] = hook_data["cwd"]
    
    # Handle tool events
    elif hook_category in ["PreToolUse", "PostToolUse"]:
        if tool_identifier:
            metadata["tool_identifier"] = tool_identifier
        
        # Extract tool input details
        if "tool_input" in hook_data:
            input_data = hook_data["tool_input"]
            if isinstance(input_data, dict):
                metadata["input_size"] = len(json.dumps(input_data))
                
                # Tool-specific metadata extraction
                if "command" in input_data:
                    command_str = str(input_data["command"])
                    metadata["command_snippet"] = sanitize_for_logging(command_str, MAX_COMMAND_SNIPPET_LENGTH)
                    
                if "file_path" in input_data:
                    file_path = Path(input_data["file_path"])
                    metadata["file_extension"] = file_path.suffix
                    metadata["target_file"] = file_path.name
    
    # Common metadata
    metadata["event_timestamp"] = datetime.now().isoformat()
    
    return metadata

def calculate_execution_duration(hook_data: Dict[str, Any]) -> Optional[int]:
    """Calculate execution duration from hook data."""
    # Look for duration in various formats
    try:
        if "execution_time" in hook_data and hook_data["execution_time"] is not None:
            duration = int(float(hook_data["execution_time"]) * 1000)
            return max(0, min(duration, 3600000))  # Cap at 1 hour
        
        if "duration_ms" in hook_data and hook_data["duration_ms"] is not None:
            duration = int(hook_data["duration_ms"])
            return max(0, min(duration, 3600000))  # Cap at 1 hour
        
        if "duration" in hook_data and hook_data["duration"] is not None:
            duration = int(float(hook_data["duration"]) * 1000)
            return max(0, min(duration, 3600000))  # Cap at 1 hour
        
    except (ValueError, TypeError, OverflowError):
        pass
    
    return None

def should_monitor_event(hook_data: Dict[str, Any], config: Dict[str, Any]) -> bool:
    """Determine if event should be monitored based on configuration."""
    monitoring_config = config.get("monitoring", {})
    
    if not monitoring_config.get("enabled", True):
        return False
    
    hook_category, tool_identifier, _ = parse_hook_event(hook_data)
    
    # Check excluded tools
    excluded_tools = monitoring_config.get("excluded_tools", [])
    if tool_identifier in excluded_tools:
        return False
    
    # Check hook type filters
    allowed_hooks = monitoring_config.get("hook_types", [])
    if allowed_hooks and hook_category not in allowed_hooks:
        return False
    
    return True

# Core monitoring functions
def record_activity_event(hook_data: Dict[str, Any]) -> bool:
    """Record an activity event to the database."""
    try:
        config = load_monitor_config()
        
        if not should_monitor_event(hook_data, config):
            return True
        
        hook_category, tool_identifier, session_id = parse_hook_event(hook_data)
        metadata = extract_event_metadata(hook_data, hook_category, tool_identifier)
        duration = calculate_execution_duration(hook_data)
        
        with database_connection() as conn:
            cursor = conn.cursor()
            
            # Insert activity record
            cursor.execute('''
                INSERT INTO activity_log 
                (event_timestamp, session_identifier, hook_category, tool_identifier, execution_duration, event_data)
                VALUES (?, ?, ?, ?, ?, ?)
            ''', (
                datetime.now().isoformat(),
                session_id,
                hook_category, 
                tool_identifier,
                duration,
                json.dumps(metadata)
            ))
            
            # Update session summary
            cursor.execute('''
                INSERT OR REPLACE INTO session_summary 
                (session_identifier, start_timestamp, total_events, total_execution_time, primary_tool)
                VALUES (?, ?, 
                        COALESCE((SELECT total_events FROM session_summary WHERE session_identifier = ?), 0) + 1,
                        COALESCE((SELECT total_execution_time FROM session_summary WHERE session_identifier = ?), 0) + COALESCE(?, 0),
                        ?)
            ''', (session_id, datetime.now().isoformat(), session_id, session_id, duration or 0, tool_identifier))
            
            conn.commit()
        
        # Log success
        log_monitor_event("INFO", f"Recorded {hook_category} activity for {tool_identifier or 'system'}")
        
        # Check custom filters
        evaluate_activity_filters(metadata, config)
        
        return True
        
    except Exception as e:
        log_monitor_event("ERROR", f"Failed to record activity: {e}")
        return False

def evaluate_activity_filters(metadata: Dict[str, Any], config: Dict[str, Any]) -> None:
    """Evaluate custom activity filters."""
    filters = config.get("filters", [])
    
    for filter_config in filters:
        filter_name = filter_config.get("name", "")
        condition = filter_config.get("condition", "")
        description = filter_config.get("description", "")
        
        if not filter_name or not condition:
            continue
        
        try:
            if evaluate_filter_condition(condition, metadata):
                # Record filter match
                with database_connection() as conn:
                    cursor = conn.cursor()
                    
                    cursor.execute('''
                        INSERT INTO activity_log 
                        (event_timestamp, session_identifier, hook_category, tool_identifier, execution_duration, event_data)
                        VALUES (?, ?, ?, ?, ?, ?)
                    ''', (
                        datetime.now().isoformat(),
                        metadata.get("session_id", "unknown"),
                        "FilterMatch",
                        filter_name,
                        None,
                        json.dumps({
                            "filter_name": filter_name,
                            "description": description,
                            "condition": condition,
                            "original_event": metadata.get("hook_category", ""),
                            "triggered_by": metadata.get("tool_identifier", "")
                        })
                    ))
                    
                    conn.commit()
                
                log_monitor_event("INFO", f"Activity filter '{filter_name}' triggered")
                
        except Exception as e:
            log_monitor_event("WARNING", f"Filter '{filter_name}' evaluation failed: {e}")

def evaluate_filter_condition(condition: str, metadata: Dict[str, Any]) -> bool:
    """Safely evaluate filter conditions."""
    # Safe field mapping
    allowed_fields = {
        "content_size": metadata.get("input_size", 0),
        "execution_time": metadata.get("execution_duration", 0),  
        "prompt_character_count": metadata.get("prompt_character_count", 0),
        "tool_identifier": metadata.get("tool_identifier", ""),
        "hook_category": metadata.get("hook_category", "")
    }
    
    # Simple condition evaluation
    try:
        for operator in [">=", "<=", ">", "<", "==", "!="]:
            if operator in condition:
                parts = condition.split(operator, 1)
                if len(parts) != 2:
                    continue
                    
                field = parts[0].strip()
                value = parts[1].strip().strip('"\'')
                
                if field not in allowed_fields:
                    continue
                
                field_value = allowed_fields[field]
                
                # Numeric comparisons
                if operator in [">=", "<=", ">", "<"]:
                    try:
                        field_num = float(field_value) if field_value else 0
                        value_num = float(value)
                        
                        if operator == ">=":
                            return field_num >= value_num
                        elif operator == "<=":
                            return field_num <= value_num
                        elif operator == ">":
                            return field_num > value_num
                        elif operator == "<":
                            return field_num < value_num
                    except ValueError:
                        continue
                
                # String comparisons
                elif operator == "==":
                    return str(field_value).lower() == value.lower()
                elif operator == "!=":
                    return str(field_value).lower() != value.lower()
                    
                break
                
    except Exception:
        pass
    
    return False

def generate_activity_summary(session_id: Optional[str] = None) -> Dict[str, Any]:
    """Generate activity summary report."""
    try:
        with database_connection() as conn:
            cursor = conn.cursor()
            
            if session_id:
                # Session-specific summary
                cursor.execute('''
                    SELECT hook_category, tool_identifier, COUNT(*), AVG(COALESCE(execution_duration, 0))
                    FROM activity_log 
                    WHERE session_identifier = ?
                    GROUP BY hook_category, tool_identifier
                    ORDER BY COUNT(*) DESC
                ''', (session_id,))
            else:
                # Overall summary
                cursor.execute('''
                    SELECT hook_category, tool_identifier, COUNT(*), AVG(COALESCE(execution_duration, 0))
                    FROM activity_log 
                    WHERE DATE(event_timestamp) = DATE('now')
                    GROUP BY hook_category, tool_identifier
                    ORDER BY COUNT(*) DESC
                ''')
            
            results = cursor.fetchall()
        
        summary = {
            "total_events": sum(row[2] for row in results),
            "event_breakdown": {},
            "most_used_tool": results[0][1] if results else "None",
            "average_execution_time": sum(row[3] for row in results) / len(results) if results else 0
        }
        
        for hook_cat, tool_id, count, avg_time in results:
            key = f"{hook_cat}:{tool_id}" if tool_id else hook_cat
            summary["event_breakdown"][key] = {
                "count": count,
                "avg_execution_time": avg_time
            }
        
        return summary
        
    except Exception as e:
        log_monitor_event("ERROR", f"Failed to generate summary: {e}")
        return {"error": str(e)}

def log_monitor_event(level: str, message: str) -> None:
    """Log monitoring events."""
    try:
        log_dir = Path("logs")
        log_dir.mkdir(exist_ok=True)
        
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "level": level.upper(),
            "category": "activity_monitor",
            "message": message,
            "session_id": os.getenv("CLAUDE_SESSION_ID", "unknown")
        }
        
        log_file = log_dir / "activity_monitor.json"
        logs = []
        
        if log_file.exists():
            try:
                with open(log_file, 'r') as f:
                    logs = json.load(f)
            except (json.JSONDecodeError, ValueError):
                logs = []
        
        logs.append(log_entry)
        
        # Keep last MAX_LOG_ENTRIES entries
        if len(logs) > MAX_LOG_ENTRIES:
            logs = logs[-MAX_LOG_ENTRIES:]
        
        with open(log_file, 'w') as f:
            json.dump(logs, f, indent=2)
            
    except Exception:
        # Silent failure for logging
        pass

def cleanup_old_data() -> None:
    """Clean up old activity data based on retention policy."""
    try:
        config = load_monitor_config()
        cleanup_days = config.get("storage", {}).get("cleanup_days", DEFAULT_CLEANUP_DAYS)
        
        # Validate cleanup_days to prevent injection
        if not isinstance(cleanup_days, int) or cleanup_days < 1 or cleanup_days > 365:
            cleanup_days = DEFAULT_CLEANUP_DAYS
        
        with database_connection() as conn:
            cursor = conn.cursor()
            
            # Clean old activity logs - using parameterized query
            cursor.execute('''
                DELETE FROM activity_log 
                WHERE DATE(event_timestamp) < DATE('now', '-' || ? || ' days')
            ''', (str(cleanup_days),))
            
            # Clean old insights - using parameterized query
            cursor.execute('''
                DELETE FROM activity_insights 
                WHERE insight_date < DATE('now', '-' || ? || ' days')  
            ''', (str(cleanup_days),))
            
            rows_deleted = cursor.rowcount
            conn.commit()
        
        if rows_deleted > 0:
            log_monitor_event("INFO", f"Cleaned up {rows_deleted} old activity records")
            
    except Exception as e:
        log_monitor_event("ERROR", f"Data cleanup failed: {e}")

def main():
    """Main entry point for the activity monitor hook."""
    try:
        # Read hook data from stdin with size limit
        raw_input = sys.stdin.read()
        
        # Validate input size to prevent DoS attacks
        if len(raw_input) > MAX_JSON_SIZE:
            log_monitor_event("WARNING", f"Input size {len(raw_input)} exceeds maximum {MAX_JSON_SIZE}")
            sys.exit(0)
        
        # Validate JSON format
        if not raw_input.strip():
            log_monitor_event("WARNING", "Empty input received")
            sys.exit(0)
        
        hook_data = json.loads(raw_input)
        
        # Debug logging with sanitized data
        sanitized_data = sanitize_for_logging(hook_data, 200)
        log_monitor_event("DEBUG", f"Processing hook event: {sanitized_data}")
        
        hook_category, tool_identifier, session_id = parse_hook_event(hook_data)
        
        # Handle different hook categories
        if hook_category in ["PreToolUse", "PostToolUse", "UserPromptSubmit"]:
            success = record_activity_event(hook_data)
            
            if not success:
                log_monitor_event("WARNING", f"Failed to record {hook_category} event")
                
        elif hook_category == "Stop":
            # Generate session summary and cleanup
            summary = generate_activity_summary(session_id)
            log_monitor_event("INFO", f"Session ended: {summary.get('total_events', 0)} events recorded")
            
            # Periodic cleanup
            cleanup_old_data()
        
        # Always return success (non-blocking)
        sys.exit(0)
        
    except json.JSONDecodeError:
        log_monitor_event("ERROR", "Invalid JSON input received")
        sys.exit(0)  # Non-blocking
    except Exception as e:
        log_monitor_event("ERROR", f"Activity monitor error: {e}")
        sys.exit(0)  # Non-blocking

if __name__ == "__main__":
    main()