#!/bin/bash
# Advanced markdown linting fixer with proper state tracking

set -euo pipefail

# Process a single markdown file
fix_markdown_file() {
    local file="$1"
    local temp_file
    temp_file=$(mktemp)

    awk '
    BEGIN {
        prev_line = ""
        prev_was_blank = 1
        prev_was_heading = 0
        prev_was_list = 0
        prev_was_fence = 0
        in_code_block = 0
        fence_count = 0
    }

    # Detect line types
    function is_blank(line) {
        return line ~ /^[[:space:]]*$/
    }

    function is_heading(line) {
        return line ~ /^#{1,6} /
    }

    function is_list(line) {
        return line ~ /^[[:space:]]*[-*+] / || line ~ /^[[:space:]]*[0-9]+\. /
    }

    function is_fence(line) {
        return line ~ /^```/
    }

    function is_emphasis_heading(line) {
        # Matches patterns like **1. Text** or **Example 1:**
        return line ~ /^\*\*[0-9]+\. [^*]+\*\*$/ || line ~ /^\*\*[^*]+:\*\*$/
    }

    function convert_emphasis_to_heading(line) {
        # Convert **1. Text** to #### 1. Text
        gsub(/^\*\*/, "#### ", line)
        gsub(/\*\*$/, "", line)
        return line
    }

    function needs_blank_before() {
        # Need blank line if transitioning from content to heading/list/fence
        if (is_heading($0) && !prev_was_blank && !prev_was_heading && prev_line != "") return 1
        if (is_list($0) && !prev_was_blank && !prev_was_list && prev_line != "") return 1
        if (is_fence($0) && !in_code_block && !prev_was_blank && prev_line != "") return 1
        return 0
    }

    # Convert emphasis-as-heading first
    {
        if (is_emphasis_heading($0)) {
            $0 = convert_emphasis_to_heading($0)
        }
    }

    # Track code block state
    {
        if (is_fence($0)) {
            if (in_code_block) {
                # Closing fence
                in_code_block = 0
                fence_count++
            } else {
                # Opening fence - add language if missing
                if ($0 == "```") {
                    $0 = "```text"
                }
                in_code_block = 1
                fence_count++
            }
        }
    }

    # Add blank line before if needed
    {
        if (needs_blank_before() && !prev_was_blank) {
            print ""
        }
    }

    # Print current line
    {
        print $0
    }

    # Add blank line after heading/list/fence if next line needs it
    {
        current_is_heading = is_heading($0)
        current_is_list = is_list($0)
        current_is_fence = is_fence($0)
        current_is_blank = is_blank($0)

        # Store for next iteration
        prev_line = $0
        prev_was_blank = current_is_blank
        prev_was_heading = current_is_heading
        prev_was_list = current_is_list
        prev_was_fence = current_is_fence
    }

    END {
        # Add blank line at end if last line was fence closing
        if (prev_was_fence && !in_code_block) {
            print ""
        }
    }
    ' "$file" > "$temp_file"

    # Second pass: ensure blank lines after headings and lists
    awk '
    BEGIN {
        prev_line = ""
        prev_was_heading = 0
        prev_was_list = 0
        prev_was_fence_close = 0
    }

    function is_blank(line) {
        return line ~ /^[[:space:]]*$/
    }

    function is_heading(line) {
        return line ~ /^#{1,6} /
    }

    function is_list(line) {
        return line ~ /^[[:space:]]*[-*+] / || line ~ /^[[:space:]]*[0-9]+\. /
    }

    function is_fence(line) {
        return line ~ /^```/
    }

    # Print previous line
    NR > 1 {
        print prev_line

        # Add blank after heading if next line is not blank/heading/list
        if (prev_was_heading && !is_blank($0) && !is_heading($0) && !is_list($0) && !is_fence($0)) {
            print ""
        }

        # Add blank after list if next line is not blank/list
        if (prev_was_list && !is_blank($0) && !is_list($0) && !is_fence($0)) {
            print ""
        }

        # Add blank after closing fence if next line is not blank
        if (prev_was_fence_close && !is_blank($0)) {
            print ""
        }
    }

    {
        prev_line = $0
        prev_was_heading = is_heading($0)
        prev_was_list = is_list($0)

        # Track fence closes (odd numbered fences are opens, even are closes)
        if (is_fence($0)) {
            fence_count++
            prev_was_fence_close = (fence_count % 2 == 0)
        } else {
            prev_was_fence_close = 0
        }
    }

    END {
        if (NR > 0) print prev_line
    }
    ' "$temp_file" > "$file"

    rm "$temp_file"
}

# Find all markdown files and process them
find .claude/agents .claude/commands -name "*.md" -type f | while read -r file; do
    echo "Processing: $file"
    fix_markdown_file "$file"
done

echo "Markdown linting fixes applied successfully"
