---
allowed-tools: Task(mage-x-linter), Task(mage-x-test-finder), Task(mage-x-security), Task(mage-x-analyzer), Task(mage-x-architect), Task(mage-x-benchmark), Bash(mage lint:*), Bash(mage test:*), Bash(mage metrics:*), Read, Grep, Glob, LS
argument-hint: [full|quick|security|performance]
description: Comprehensive code quality assessment with parallel analysis
model: claude-sonnet-4-20250514
---

## Context
- Project metrics: !`find . -name "*.go" -not -path "./vendor/*" | wc -l | xargs echo "Go files:"`
- Test coverage: !`go test -cover ./... 2>/dev/null | grep -E "coverage:" | tail -5`
- Build status: !`go build ./... 2>&1 | grep -E "(Error|error:|cannot)" | head -5 || echo "Build: OK"`

## Your Task

Perform a comprehensive quality assessment through coordinated parallel agent analysis. Mode: ${ARGUMENTS:-full}

### Phase 1: Parallel Quality Analysis
Execute all agents simultaneously for maximum efficiency:

1. **mage-x-linter** - Code Quality:
   - Run full linting suite (`mage lintAll`)
   - Analyze code style compliance
   - Check documentation quality
   - Identify technical debt

2. **mage-x-test-finder** - Test Coverage:
   - Analyze test coverage gaps
   - Identify untested critical paths
   - Assess test quality and patterns
   - Check for missing test types

3. **mage-x-security** - Security Assessment:
   - Scan for vulnerabilities
   - Validate secure coding practices
   - Check dependency security
   - Review authentication/authorization

4. **mage-x-analyzer** - Complexity & Metrics:
   - Calculate cyclomatic complexity
   - Analyze code duplication
   - Measure technical debt
   - Performance bottleneck detection

5. **mage-x-architect** - Architecture Compliance:
   - Validate 30+ namespace patterns
   - Check interface implementations
   - Review dependency structure
   - Assess modularity

6. **mage-x-benchmark** - Performance Profile:
   - Identify performance hotspots
   - Memory usage patterns
   - Concurrency analysis
   - Resource utilization

### Phase 2: Integrated Analysis
Synthesize results from all agents to provide:
- Cross-cutting concerns identification
- Risk assessment matrix
- Priority-ordered improvement plan
- Quick wins vs long-term improvements

### Quality Assessment Modes
- **full**: Complete analysis across all dimensions
- **quick**: Fast scan of critical issues only  
- **security**: Security-focused deep analysis
- **performance**: Performance and scalability focus

## Expected Output

### Executive Summary
- Overall quality score and trend
- Critical issues requiring immediate attention
- Key metrics and KPIs

### Detailed Analysis by Domain
1. **Code Quality** (from linter)
2. **Test Coverage** (from test-finder)
3. **Security Posture** (from security)
4. **Performance Profile** (from analyzer & benchmark)
5. **Architecture Health** (from architect)

### Actionable Recommendations
- Immediate fixes (< 1 day)
- Short-term improvements (< 1 week)
- Long-term refactoring needs
- Process improvements

### Quality Metrics Dashboard
```
Code Quality:    [████████░░] 80%
Test Coverage:   [██████░░░░] 60%
Security:        [█████████░] 90%
Performance:     [███████░░░] 70%
Architecture:    [████████░░] 85%
```
