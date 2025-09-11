---
allowed-tools: Task(mage-x-linter), Task(mage-x-test-writer), Task(mage-x-security), Task(mage-x-refactor), Bash(magex lint:*), Bash(magex test:*), Bash(go test:*), Read, Write, MultiEdit, Grep, Glob, LS
argument-hint: [all|lint|test|security]
description: Fix linting issues and test errors with parallel agent execution
model: claude-sonnet-4-20250514
---

## Context
- Current errors: !`magex lint 2>&1 | grep -E "(Error|Warning|FAIL)" | head -20`
- Test failures: !`go test ./... -v 2>&1 | grep -E "FAIL|Error:|panic:" | head -20`
- Linter status: !`magex lint:version 2>/dev/null || echo "Linter not configured"`

## Your Task

I need to fix code quality issues and test failures through parallel agent coordination. Focus area: $ARGUMENTS

### Phase 1: Parallel Issue Detection
Execute these agents simultaneously for comprehensive analysis:

1. **mage-x-linter** (if lint issues exist):
   - Run comprehensive linting with `magex lint`
   - Identify auto-fixable issues vs manual fixes needed
   - Prioritize by severity (critical → high → medium → low)

2. **Identify test failures** (if test issues exist):
   - Analyze failing tests with error messages
   - Categorize failures (compilation, assertion, panic, timeout)
   - Identify root causes and fix strategies

3. **mage-x-security** (if security issues detected):
   - Scan for security vulnerabilities in fixes
   - Validate CommandExecutor usage patterns
   - Ensure secure coding practices

### Phase 2: Parallel Fix Implementation
Based on Phase 1 results, execute fixes in parallel:

1. **Auto-fix linting issues**:
   - Apply `magex lint:fix` for safe automated corrections
   - Fix import organization and formatting
   - Resolve golangci-lint violations

2. **mage-x-test-writer** (for test fixes):
   - Fix failing test assertions
   - Update outdated test expectations
   - Add missing test setup/teardown
   - Ensure proper t.Parallel() usage

3. **mage-x-refactor** (for complex fixes):
   - Refactor code to resolve architectural issues
   - Apply proper error handling patterns
   - Implement interface-based designs where needed

### Phase 3: Validation (Parallel)
1. **Re-run linters**: `magex lint`
2. **Re-run tests**: `magex test:unit` and `magex test:race`
3. **Security validation**: Verify no new vulnerabilities introduced

## Fix Priorities
1. **Critical**: Security vulnerabilities, build failures
2. **High**: Failing tests, race conditions, memory leaks
3. **Medium**: Code style, inefficient patterns, missing documentation
4. **Low**: Optional improvements, minor style issues

## Expected Output
- Summary of all issues found and fixed
- Detailed fix report by category
- Validation results showing all issues resolved
- Recommendations for preventing future issues
