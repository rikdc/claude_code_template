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

try:
    import toml
except ImportError:
    toml = None

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
    """Get database connection with proper setup."""
    import sqlite3
    
    config = load_monitor_config()
    db_path = config.get("storage", {}).get("database_path", ".claude/activity_metrics.db")
    
    # Ensure directory exists
    os.makedirs(os.path.dirname(db_path), exist_ok=True)
    
    conn = sqlite3.connect(db_path)
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
            metadata["prompt_snippet"] = str(hook_data["prompt"])[:50] + "..." if len(str(hook_data["prompt"])) > 50 else str(hook_data["prompt"])
        
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
                    metadata["command_snippet"] = str(input_data["command"])[:100]
                    
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
    if "execution_time" in hook_data and hook_data["execution_time"] is not None:
        return int(float(hook_data["execution_time"]) * 1000)
    
    if "duration_ms" in hook_data and hook_data["duration_ms"] is not None:
        return int(hook_data["duration_ms"])
    
    if "duration" in hook_data and hook_data["duration"] is not None:
        return int(float(hook_data["duration"]) * 1000)
    
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
        
        conn = get_database_connection()
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
        conn.close()
        
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
                conn = get_database_connection()
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
                conn.close()
                
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
        conn = get_database_connection()
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
        conn.close()
        
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
        
        # Keep last 500 entries
        if len(logs) > 500:
            logs = logs[-500:]
        
        with open(log_file, 'w') as f:
            json.dump(logs, f, indent=2)
            
    except Exception:
        # Silent failure for logging
        pass

def cleanup_old_data() -> None:
    """Clean up old activity data based on retention policy."""
    try:
        config = load_monitor_config()
        cleanup_days = config.get("storage", {}).get("cleanup_days", 30)
        
        conn = get_database_connection()
        cursor = conn.cursor()
        
        # Clean old activity logs
        cursor.execute('''
            DELETE FROM activity_log 
            WHERE DATE(event_timestamp) < DATE('now', '-{} days')
        '''.format(cleanup_days))
        
        # Clean old insights
        cursor.execute('''
            DELETE FROM activity_insights 
            WHERE insight_date < DATE('now', '-{} days')  
        '''.format(cleanup_days))
        
        rows_deleted = cursor.rowcount
        conn.commit()
        conn.close()
        
        if rows_deleted > 0:
            log_monitor_event("INFO", f"Cleaned up {rows_deleted} old activity records")
            
    except Exception as e:
        log_monitor_event("ERROR", f"Data cleanup failed: {e}")

def main():
    """Main entry point for the activity monitor hook."""
    try:
        # Read hook data from stdin
        hook_data = json.loads(sys.stdin.read())
        
        # Debug logging
        log_monitor_event("DEBUG", f"Processing hook event: {json.dumps(hook_data)}")
        
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