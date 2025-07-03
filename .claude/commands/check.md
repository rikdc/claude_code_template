---
allowed-tools: all
description: Fix all code quality issues using parallel sub-tasks
---

# Code Quality Check & Auto-Fix

Comprehensive analysis and fixing of all code quality issues. This command **fixes problems, not just reports them**.

## Workflow

1. **Analyze** project and identify all issues
2. **Spawn sub-tasks** for independent fixes (linting, formatting, simple tests)
3. **Handle dependencies** sequentially (complex tests, integration issues)
4. **Verify** all fixes work together
5. **Continue** until everything passes

## Sub-task Strategy

When I find multiple issues, I'll spawn parallel sub-tasks:

- Separate sub-tasks for different directories/modules
- Parallel execution of independent fixes (formatting, linting, simple tests)
- Sequential handling of dependent changes
- Coordination to prevent conflicts

## Success Criteria

- ✅ All linters pass (zero warnings)
- ✅ All tests pass
- ✅ All builds succeed
- ✅ No security vulnerabilities

## Execution Promise

I will work systematically through all issues, spawning sub-tasks where beneficial, and continue until every check shows ✅ **PASSING**.

Example: "Found 20 issues across 8 files. Spawning 3 sub-tasks: formatting fixes, linting violations, and test failures. Coordinating results and verifying integration..."
