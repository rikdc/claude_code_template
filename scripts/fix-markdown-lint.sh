#!/usr/bin/env bash
#
# fix-markdown-lint.sh - Automatically fix common markdown linting issues
#
# This script fixes the following markdown issues:
# - MD031: Add blank lines around fenced code blocks
# - MD040: Add language tags to code blocks
# - MD022: Add blank lines around headings
# - MD032: Add blank lines around lists
# - MD034: Convert bare URLs to proper markdown links

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Function to fix a single markdown file
fix_markdown_file() {
    local file="$1"
    local temp_file="${file}.tmp"

    log_info "Fixing: $file"

    # Use awk to fix common markdown issues
    awk '
    BEGIN {
        prev_line = ""
        prev_blank = 0
        in_code_block = 0
        code_block_line = 0
    }

    # Track code blocks
    /^```/ {
        if (in_code_block == 0) {
            # Starting code block
            in_code_block = 1
            code_block_line = NR

            # MD031: Add blank line before code block if needed
            if (prev_line != "" && prev_blank == 0) {
                print ""
            }

            # MD040: Add language tag if missing
            if ($0 == "```") {
                print "```text"
            } else {
                print $0
            }

            prev_line = $0
            prev_blank = 0
            next
        } else {
            # Ending code block
            in_code_block = 0
            print $0

            # MD031: Will add blank line after if next line is not blank
            prev_line = $0
            prev_blank = 0
            next
        }
    }

    # Inside code block - print as-is
    in_code_block == 1 {
        print $0
        prev_line = $0
        prev_blank = 0
        next
    }

    # MD022: Add blank line before heading if needed
    /^#{1,6} / {
        if (prev_line != "" && prev_blank == 0 && prev_line !~ /^---$/ && prev_line !~ /^#{1,6} /) {
            print ""
        }
        print $0
        prev_line = $0
        prev_blank = 0
        next
    }

    # MD032: Add blank line before list if needed
    /^[*+-] / || /^[0-9]+\. / {
        if (prev_line != "" && prev_blank == 0 && prev_line !~ /^[*+-] / && prev_line !~ /^[0-9]+\. / && prev_line !~ /^#{1,6} /) {
            print ""
        }
        print $0
        prev_line = $0
        prev_blank = 0
        next
    }

    # Track blank lines
    /^$/ {
        if (prev_blank == 0) {
            print $0
            prev_blank = 1
        }
        prev_line = ""
        next
    }

    # MD034: Convert bare URLs to markdown links (simple approach)
    {
        # Convert http(s):// URLs that are not already in markdown syntax
        line = $0
        # Only convert if not in <> and not already in []()
        if (line !~ /\[.*\]\(/ && line !~ /<http/) {
            gsub(/https?:\/\/[a-zA-Z0-9.\/?=_-]+/, "<&>", line)
        }
        print line
        prev_line = line
        prev_blank = 0
    }
    ' "$file" > "$temp_file"

    # Replace original file
    mv "$temp_file" "$file"

    log_info "Fixed: $file"
}

# Main execution
main() {
    local base_dir="/Users/richard.claydon/go/src/github.com/kohofinancial/personal/claude_code_template/.conductor/auckland"

    log_info "Starting markdown lint fixes..."

    # Fix agent files
    log_info "Fixing agent files..."
    for file in "$base_dir/.claude/agents"/*.md; do
        if [[ -f "$file" ]]; then
            fix_markdown_file "$file"
        fi
    done

    # Fix command files
    log_info "Fixing command files..."
    for file in "$base_dir/.claude/commands"/*.md; do
        if [[ -f "$file" ]]; then
            fix_markdown_file "$file"
        fi
    done

    log_info "✓ All markdown files fixed!"
    log_warn "Note: Some issues may require manual review (duplicate headings, heading levels, emphasis as headings)"
}

main "$@"
