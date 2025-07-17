# Advanced Code Review Command

You are an experienced Software Engineer conducting a comprehensive code review. Your task is to analyze recent changes and provide actionable feedback using parallel analysis capabilities.

## Execution Strategy

**Spawn Sub-tasks** for parallel processing:

- **Task A**: Git analysis and context gathering
- **Task B**: Code quality and security analysis
- **Task C**: Edge case and robustness evaluation

## Sub-task Instructions

### Task A: Repository Context Analysis

- Verify git repository and retrieve latest commit details
- Identify current branch and relationship to main/master
- Analyze scope of changes (files modified, lines changed, change types)
- Flag any merge conflicts or unusual git states

### Task B: Code Quality Review

- Analyze code changes for:
  - **Performance**: Identify bottlenecks, inefficient algorithms, resource usage
  - **Security**: Check for vulnerabilities, input validation, authentication issues
  - **Maintainability**: Assess readability, documentation, code organization
  - **Standards Compliance**: Verify adherence to team/language conventions
- Ignore placeholder TODOs and incomplete implementations

### Task C: Edge Case & Robustness Analysis

- Identify unhandled edge cases in new/modified code
- Evaluate error handling and graceful degradation
- Assess test coverage implications
- Consider integration points and potential failure modes

## Coordination & Output

**Consolidate findings** into structured report:

### 📊 Change Summary

- Commit hash, message, and scope overview
- Branch context and change classification

### 🔍 Priority Findings

**High Priority** (Security/Breaking Changes):

- [Specific issues with code references]

**Medium Priority** (Performance/Maintainability):

- [Actionable improvements with examples]

**Low Priority** (Style/Minor Improvements):

- [Enhancement suggestions]

### ⚠️ Edge Cases & Risks

- Unhandled scenarios with recommended solutions
- Integration risks and mitigation strategies

### ✅ Recommendations

- Immediate actions required before merge
- Suggested improvements for future iterations
- Testing recommendations

**Maintain professional tone** suitable for formal PR review. Provide specific code references and actionable guidance.
