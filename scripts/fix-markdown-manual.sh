#!/usr/bin/env bash
#
# fix-markdown-manual.sh - Fix remaining markdown issues that require manual attention
#

set -euo pipefail

# Fix go-review-agent.md - Add inline disable comments for duplicate headings in example template
sed -i '' 's/^### Major Issues ⚠️$/<!-- markdownlint-disable MD024 -->\n### Major Issues ⚠️/' \
    /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/agents/go-review-agent.md

sed -i '' 's/^### Minor Issues 💡$/### Minor Issues 💡\n<!-- markdownlint-enable MD024 -->/' \
    /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/agents/go-review-agent.md

# Fix heading increment issue - change h4 to h3 in go-review-agent.md line 538
sed -i '' '540s/^#### 2\. Race Condition in Concurrent Updates$/### 2. Race Condition in Concurrent Updates/' \
    /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/agents/go-review-agent.md

# Fix emphasis as headings in staff-eyes-agent.md and others
# Change **Example 1:** to #### Example 1: (proper heading)
find /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/agents \
    -name "*.md" -exec sed -i '' 's/^\*\*Example \([0-9]\+\):/#### Example \1:/' {} \;

find /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/commands \
    -name "*.md" -exec sed -i '' 's/^\*\*Example \([0-9]\+\):/#### Example \1:/' {} \;

# Fix **1. Clarity** style headings to proper subheadings
find /Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland/.claude/commands \
    -name "*.md" -exec sed -i '' 's/^\*\*\([0-9]\+\)\. \(.*\)\*\*$/#### \1. \2/' {} \;

echo "✓ Manual markdown fixes applied"
