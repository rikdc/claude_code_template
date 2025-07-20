#!/usr/bin/env bash

# Claude Code Hook: Capture User Prompts
# Event: UserPromptSubmit
# Purpose: Ultra-fast capture of user prompts to file-based queue
# Performance: Sub-millisecond execution via atomic file operations

set -euo pipefail

# Configuration - Use absolute paths for reliability
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
readonly PROJECT_ROOT

QUEUE_DIR="${PROJECT_ROOT}/.claude/apps/prompt_manager/queue/incoming"
readonly QUEUE_DIR

# Generate unique filename with timestamp-session-event format
# Format: YYYYMMDD_HHMMSS_NANOS_SESSION_prompt.json
TIMESTAMP=$(date '+%Y%m%d_%H%M%S_%N')
SESSION_ID="${CLAUDE_SESSION_ID:-unknown}"
FILENAME="${TIMESTAMP}_${SESSION_ID}_prompt.json"
FILEPATH="${QUEUE_DIR}/${FILENAME}"

# Ultra-fast atomic file write
# Write to temp file first, then atomic move to prevent partial reads
TEMP_FILE="${FILEPATH}.tmp"

# Create JSON structure with all available hook data
cat > "$TEMP_FILE" << EOF
{
  "event": "UserPromptSubmit",
  "timestamp": "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')",
  "session_id": "${SESSION_ID}",
  "filename": "${FILENAME}",
  "data": ${CLAUDE_HOOK_DATA}
}
EOF

# Atomic move - ensures file appears complete or not at all
mv "$TEMP_FILE" "$FILEPATH"

# Exit with code 0 to allow the request to proceed
exit 0