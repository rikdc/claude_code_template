---
title: Clean comments
description: Removes useless comments from the code
---

- Look for comments that are obvious or redundant and remove them. Examples of
  comments that can be removed include:
  - Commented out code.
  - Comments that describe edits like "added", "removed", or "changed" something.
  - Explanations that are just obvious because they are close to method names.
- Do not delete all comments:
  - Don't remove comments that indicate code is a stub for future work.
  - Don't remove comments that start with TODO.
  - Don't remove comments if doing so would make a scope empty, like an empty catch
    block or an empty else block.
  - Don't remove comments that suppress linters or formatters, like
    `// prettier-ignore`
- If you find any end-of-line comments, move them above the code they describe.
  Comments should go on their own lines.

Look at the local code changes that are not yet committed to git
