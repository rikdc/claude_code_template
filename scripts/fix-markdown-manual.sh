#!/usr/bin/env bash
#
# fix-markdown-manual.sh - Fix remaining markdown issues that require manual attention
#
# Usage: fix-markdown-manual.sh [target_directory]
#        If no directory provided, uses current directory
#

set -euo pipefail

# Get the target directory (default to current directory)
TARGET_DIR="${1:-.}"

# Resolve to absolute path
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"

# Verify .claude directory exists
if [[ ! -d "$TARGET_DIR/.claude" ]]; then
    echo "Error: No .claude directory found in $TARGET_DIR" >&2
    exit 1
fi

# Fix go-review-agent.md - Add inline disable comments for duplicate headings in example template
if [[ -f "$TARGET_DIR/.claude/agents/go-review-agent.md" ]]; then
    sed -i '' 's/^### Major Issues ⚠️$/<!-- markdownlint-disable MD024 -->\n### Major Issues ⚠️/' \
        "$TARGET_DIR/.claude/agents/go-review-agent.md"

    sed -i '' 's/^### Minor Issues 💡$/### Minor Issues 💡\n<!-- markdownlint-enable MD024 -->/' \
        "$TARGET_DIR/.claude/agents/go-review-agent.md"

    # Fix heading increment issue - change h4 to h3 in go-review-agent.md line 538
    sed -i '' '540s/^#### 2\. Race Condition in Concurrent Updates$/### 2. Race Condition in Concurrent Updates/' \
        "$TARGET_DIR/.claude/agents/go-review-agent.md"
fi

# Fix emphasis as headings in agents and commands
# Change **Example 1:** to #### Example 1: (proper heading)
if [[ -d "$TARGET_DIR/.claude/agents" ]]; then
    find "$TARGET_DIR/.claude/agents" \
        -name "*.md" -exec sed -i '' 's/^\*\*Example \([0-9]\+\):/#### Example \1:/' {} \;
fi

if [[ -d "$TARGET_DIR/.claude/commands" ]]; then
    find "$TARGET_DIR/.claude/commands" \
        -name "*.md" -exec sed -i '' 's/^\*\*Example \([0-9]\+\):/#### Example \1:/' {} \;

    # Fix **1. Clarity** style headings to proper subheadings
    find "$TARGET_DIR/.claude/commands" \
        -name "*.md" -exec sed -i '' 's/^\*\*\([0-9]\+\)\. \(.*\)\*\*$/#### \1. \2/' {} \;
fi

echo "Manual markdown fixes applied to $TARGET_DIR"
