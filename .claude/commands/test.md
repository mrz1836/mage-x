---
allowed-tools: Task(mage-x-test-finder), Task(mage-x-test-writer), Task(mage-x-analyzer), Task(mage-x-security), Task(mage-x-linter), Task(mage-x-architect), Bash(mage test:*), Bash(go test:*), Read, Write, MultiEdit, Grep, Glob, LS
argument-hint: [specific-path-or-namespace]
description: Create comprehensive Go tests with parallel agent execution
model: claude-sonnet-4-20250514
---

## Context
- Project overview: !`mage -version 2>/dev/null || echo "mage-x build system"`
- Test execution modes: !`mage -l | grep -E "^test" | head -10`
- Recent test results: !`go test -v ./... 2>&1 | grep -E "(PASS|FAIL|ok|SKIP)" | tail -20`

## Your Task

I need comprehensive Go tests written for the mage-x project with maximum efficiency through parallel agent execution. $ARGUMENTS

Please coordinate the following agents in parallel for optimal test creation:

### Phase 1: Parallel Analysis (Execute Simultaneously)
1. **mage-x-test-finder**: Identify untested code, prioritize by criticality, analyze test gaps
2. **mage-x-analyzer**: Assess code complexity, identify performance-critical paths, recommend test strategies
3. **mage-x-security**: Identify security-sensitive code requiring specialized tests
4. **mage-x-architect**: Validate architectural patterns and identify interface testing needs

### Phase 2: Test Implementation (Based on Phase 1 Results)
1. **mage-x-test-writer**: Create comprehensive tests based on the combined analysis:
   - Unit tests with t.Parallel() for concurrent execution
   - Table-driven tests for comprehensive coverage
   - Benchmark tests for performance-critical code
   - Fuzz tests for input validation functions
   - Integration tests for namespace interactions
   - Security tests for sensitive operations

### Phase 3: Quality Validation (Parallel)
1. **mage-x-linter**: Validate test code quality and patterns
2. **Execute Tests**: Run `mage testUnit` and `mage testRace` to validate

## Requirements
- Focus on the 30+ namespace architecture and interface implementations
- Ensure all tests follow mage-x security-first principles
- Maximize parallel test execution with proper t.Parallel() usage
- Include comprehensive error path testing
- Create interface-based mocks for better maintainability
- Follow Go testing best practices and conventions

## Expected Output
Coordinate agents to deliver:
1. Executive summary of test gaps and implementation
2. Prioritized test creation across critical areas
3. Comprehensive test files with parallel optimization
4. Validation results from test execution
5. Recommendations for ongoing test maintenance
