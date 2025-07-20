# Prompt Manager Data Ingestion Strategy

## Problem Statement

The Prompt Manager needs to reliably capture conversation data from Claude Code hooks while maintaining system responsiveness. The primary challenge is that Claude Code hooks have strict timing constraints - they must execute quickly to avoid blocking user interactions or causing timeouts.

Initial consideration of a REST API approach raised concerns about network latency and HTTP request timeouts potentially impacting Claude Code performance.

## Hook Timing Constraints

Based on analysis of Claude Code's hook system:

- **Performance Requirements**: Hooks must execute in milliseconds to avoid noticeable delays
- **Blocking Risk**: Slow hooks can impact Claude Code's responsiveness
- **Network Dependencies**: HTTP calls introduce unpredictable latency and failure modes
- **Data Volume**: Conversation data can be substantial (multi-KB prompts, full transcripts)

## Solution Analysis

### Option 1: Direct REST API (NOT RECOMMENDED)

**Architecture**: Hook makes HTTP POST to Go backend API endpoint

**Pros**:
- Simple, straightforward implementation
- Real-time data availability in web interface
- Standard REST patterns

**Cons**:
- Network latency can cause hook timeouts
- HTTP failures could block Claude Code operations
- Dependency on web server availability during hooks
- Single point of failure

**Risk Assessment**: HIGH - Could impact Claude Code performance

### Option 2: Direct SQLite Write (RISKY)

**Architecture**: Hook writes directly to SQLite database

**Pros**:
- Fast local writes
- No network dependencies
- Immediate data availability

**Cons**:
- Database locking issues with concurrent access
- Risk of database corruption from failed writes
- Difficult to handle processing errors gracefully
- Mixing hook execution with database transactions

**Risk Assessment**: MEDIUM - Database locking could cause delays

### Option 3: File-Based Queue with Background Processing (RECOMMENDED)

**Architecture**: Hook writes to file queue, background daemon processes files

**Pros**:
- Ultra-fast hook execution (<1ms file writes)
- No network dependencies during hooks
- Graceful failure handling
- Atomic file operations prevent corruption
- Easy debugging and monitoring
- Scalable to high-volume conversations

**Cons**:
- Slightly more complex architecture
- Small delay between capture and web interface availability
- Requires background daemon management

**Risk Assessment**: LOW - Minimal impact on Claude Code performance

## Recommended Architecture

### File-Based Queue System

```
Claude Code Hook → Queue File → Background Processor → SQLite → REST API → Web UI
    (<1ms)         (instant)      (asynchronous)      (fast)     (standard)
```

### Directory Structure

```
apps/prompt_manager/
├── queue/
│   ├── incoming/     # New files from hooks
│   ├── processing/   # Files being processed
│   ├── failed/       # Failed processing attempts
│   └── archive/      # Successfully processed (optional)
├── database/
│   └── prompts.db    # SQLite database
└── logs/
    └── processor.log # Background processing logs
```

### File Format

Queue files use JSON format with consistent naming:
- **Filename**: `{timestamp}-{session-id}-{hook-type}.json`
- **Content**: Complete hook data as received from Claude Code

Example filename: `1642789234567890123-abc123-UserPromptSubmit.json`

### Hook Implementation

```bash
#!/bin/bash
# Ultra-fast hook execution
QUEUE_DIR="$WORKSPACE/apps/prompt_manager/queue/incoming"
TIMESTAMP=$(date +%s%N)  # Nanosecond precision
FILENAME="${TIMESTAMP}-${CLAUDE_SESSION_ID}-${HOOK_EVENT_NAME}.json"

# Atomic write operation
echo "$HOOK_DATA" > "$QUEUE_DIR/$FILENAME"
exit 0
```

### Background Processor Features

**File System Watcher**:
- Uses inotify (Linux) or fsevents (macOS) for instant processing
- Processes files as they arrive in queue directory

**Processing Pipeline**:
1. Move file from `incoming/` to `processing/`
2. Parse and validate JSON data
3. Update SQLite database with conversation data
4. Move file to `archive/` or delete if configured
5. Handle errors by moving to `failed/` directory

**Error Handling**:
- Retry logic for transient failures
- Dead letter queue for permanently failed files
- Comprehensive logging for debugging

**Data Processing**:
- Extract conversation context from transcript files
- Correlate multiple hook events for complete conversations
- Handle session splitting and organization logic

## Implementation Details

### Hook Configuration

Update `.claude/settings.json`:
```json
{
  "hooks": {
    "UserPromptSubmit": [{
      "matcher": ".*",
      "hooks": [{
        "type": "command",
        "command": "${WORKSPACE}/apps/prompt_manager/hooks/capture-prompt.sh"
      }]
    }],
    "Stop": [{
      "matcher": ".*", 
      "hooks": [{
        "type": "command",
        "command": "${WORKSPACE}/apps/prompt_manager/hooks/capture-completion.sh"
      }]
    }]
  }
}
```

### Database Schema Design

```sql
-- Conversations table
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY,
    session_id TEXT NOT NULL,
    user_session_id TEXT, -- User-defined session for splitting
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    prompt_text TEXT,
    response_text TEXT,
    tool_calls JSON,
    metadata JSON
);

-- Ratings and tags
CREATE TABLE ratings (
    conversation_id INTEGER,
    rating INTEGER CHECK(rating >= 1 AND rating <= 5),
    tags TEXT, -- JSON array of tags
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(conversation_id) REFERENCES conversations(id)
);
```

### Background Daemon

**Go Implementation**:
- Lightweight daemon using `fsnotify` library
- Configurable processing intervals and batch sizes
- Health check endpoints for monitoring
- Graceful shutdown handling

**Service Management**:
- systemd service file for Linux
- launchd plist for macOS  
- Process monitoring and auto-restart

### REST API Design

**Endpoints**:
- `GET /api/conversations` - List conversations with filtering
- `GET /api/conversations/{id}` - Get conversation details
- `POST /api/conversations/{id}/rating` - Add/update rating
- `POST /api/sessions/split` - Split session into multiple conversations
- `GET /api/health` - System health and queue status

## Benefits of This Approach

### Performance
- **Hook Execution**: Sub-millisecond file writes
- **No Blocking**: Zero impact on Claude Code responsiveness
- **Scalable**: Handles high-volume conversations efficiently

### Reliability
- **No Network Dependencies**: File operations during hooks
- **Atomic Operations**: Prevents data corruption
- **Failure Isolation**: Processing failures don't affect hooks
- **Retry Mechanisms**: Built-in error recovery

### Maintainability
- **Clear Separation**: Distinct responsibilities for each component
- **Easy Debugging**: Queue files provide audit trail
- **Monitoring**: File system metrics and processing logs
- **Testing**: Easy to simulate different scenarios

### Future Scalability
- **Multi-User Ready**: Can support centralized database migration
- **Horizontal Scaling**: Background processors can be distributed
- **Data Export**: Queue files provide portable data format

## Monitoring and Operations

### Health Checks
- Queue directory file counts
- Background processor status
- Database connection health
- Processing latency metrics

### Logging Strategy
- Hook execution logs (minimal, performance-focused)
- Background processing logs (detailed, error-focused)  
- API access logs (standard web server logs)

### Error Scenarios
- Disk space exhaustion: Alert and cleanup policies
- Database corruption: Backup and recovery procedures
- Processor crashes: Auto-restart and notification
- Hook failures: Fallback mechanisms and logging

## Implementation Priority

1. **Phase 1**: Basic file queue and simple processor
2. **Phase 2**: Database integration and REST API
3. **Phase 3**: Advanced features (session splitting, AI tagging)
4. **Phase 4**: Monitoring, health checks, and operations tools

This architecture provides the optimal balance of performance, reliability, and maintainability for the Prompt Manager's data ingestion requirements.