---
allowed-tools: Task(mage-x-gh), Task(mage-x-security), Task(mage-x-tools), Bash(gh workflow:*), Bash(gh run:*), Bash(gh api:*), Read, Grep, Glob, LS
argument-hint: [workflow-name|run-id|all]
description: Diagnose CI/CD issues in GitHub workflows with intelligent analysis
model: claude-sonnet-4-20250514
---

## Context
- GitHub workflows: !`ls -la .github/workflows/*.yml 2>/dev/null | wc -l | xargs echo "Workflow files:"`
- Recent workflow runs: !`gh run list --limit 5 2>/dev/null || echo "GitHub CLI not configured"`
- Current branch: !`git branch --show-current`

## Your Task

Diagnose CI/CD issues in GitHub Actions workflows through parallel agent analysis. Target: ${ARGUMENTS:-all}

### Phase 1: Parallel CI/CD Analysis

1. **mage-x-gh** - GitHub Integration Analysis:
   - Fetch recent workflow runs and their status
   - Analyze failing jobs and steps
   - Review workflow run logs for errors
   - Check PR check status and requirements
   - Identify patterns in failures

2. **mage-x-workflow** - Workflow Configuration:
   - Validate workflow YAML syntax and structure
   - Check for common workflow anti-patterns
   - Analyze job dependencies and parallelization
   - Review caching and artifact usage
   - Identify optimization opportunities

3. **mage-x-tools** - Tool & Environment Issues:
   - Verify tool versions and compatibility
   - Check for missing dependencies
   - Analyze environment setup issues
   - Review cross-platform compatibility
   - Identify tool installation failures

4. **mage-x-security** - Security & Permissions:
   - Check workflow permissions and secrets
   - Validate token usage and scopes
   - Review security best practices
   - Identify potential security issues

### Phase 2: Root Cause Analysis

Based on Phase 1 results:
- Correlate failures across multiple runs
- Identify environmental vs code issues
- Detect flaky tests and race conditions
- Analyze resource constraints and timeouts
- Review matrix build failures

### Phase 3: Solution Generation

Provide fixes for identified issues:
- Workflow configuration corrections
- Environment setup improvements
- Test stability enhancements
- Performance optimizations
- Security remediations

## Diagnostic Modes

- **specific workflow**: Analyze a particular workflow file
- **run-id**: Deep dive into a specific failed run
- **all**: Comprehensive analysis of all workflows

## Expected Output

### Executive Summary
- CI/CD health status
- Critical failures and their impact
- Success rate trends

### Detailed Diagnostics

1. **Failed Workflows Analysis**
   - Workflow name and failure rate
   - Common failure patterns
   - Root cause identification

2. **Environmental Issues**
   - Tool version conflicts
   - Missing dependencies
   - Platform-specific failures

3. **Performance Bottlenecks**
   - Slow steps and jobs
   - Inefficient caching
   - Parallel execution opportunities

4. **Security Concerns**
   - Exposed secrets or tokens
   - Excessive permissions
   - Vulnerable dependencies

### Actionable Fixes

```yaml
# Example workflow fixes will be provided here
```

### Optimization Recommendations
- Quick fixes (< 10 min)
- Infrastructure improvements
- Long-term CI/CD enhancements
- Monitoring and alerting setup
