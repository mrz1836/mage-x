---
allowed-tools: Task(mage-x-security), Task(mage-x-deps), Task(mage-x-workflow), Task(mage-x-refactor), Bash(mage tools:vulncheck), Bash(go list:*), Bash(govulncheck:*), Read, Write, MultiEdit, Grep, Glob, LS
argument-hint: [full|deps|code|secrets|fix]
description: Comprehensive security audit with automated remediation
model: sonnet
---

## Context
- Dependency count: !`go list -m all 2>/dev/null | wc -l | xargs echo "Total dependencies:"`
- Security tools: !`which govulncheck gosec 2>/dev/null | wc -l | xargs echo "Security tools available:"`
- Last scan: !`ls -la .security-scan-* 2>/dev/null | tail -1 || echo "No recent scans"`

## Your Task

Perform a comprehensive security audit with parallel analysis and remediation. Mode: ${ARGUMENTS:-full}

### Phase 1: Parallel Security Analysis

1. **mage-x-security** - Code Security:
   - Command injection vulnerabilities
   - SQL injection risks
   - Path traversal issues
   - Insecure random usage
   - Hardcoded credentials
   - CommandExecutor validation

2. **mage-x-deps** - Dependency Security:
   - Known vulnerabilities (CVEs)
   - Outdated dependencies
   - License compliance
   - Supply chain risks
   - Transitive dependencies

3. **Secret Scanning**:
   - API keys and tokens
   - Passwords and credentials
   - Private keys
   - Connection strings
   - Environment variables

4. **mage-x-workflow** - CI/CD Security:
   - Workflow permissions
   - Secret exposure
   - Third-party actions
   - Token scopes
   - Build security

### Phase 2: Risk Assessment

Analyze findings to determine:
- **Critical**: Immediate exploitation possible
- **High**: Significant security risk
- **Medium**: Potential future risk
- **Low**: Best practice violations

### Phase 3: Automated Remediation

1. **Dependency Updates** (mage-x-deps):
   - Update vulnerable dependencies
   - Pin versions securely
   - Remove unused dependencies

2. **Code Fixes** (mage-x-refactor):
   - Fix injection vulnerabilities
   - Implement proper validation
   - Add security headers
   - Enhance error handling

3. **Configuration Hardening**:
   - Tighten permissions
   - Rotate exposed secrets
   - Update security policies

### Security Scan Modes

- **full**: Complete security audit
- **deps**: Dependency vulnerabilities only
- **code**: Source code security issues
- **secrets**: Secret and credential scanning
- **fix**: Apply automated fixes

## Expected Output

### Executive Summary
```
Security Score: [A-F rating]
Critical Issues: [count]
High Risk: [count]
Medium Risk: [count]
Low Risk: [count]
```

### Critical Vulnerabilities
```
🚨 CRITICAL: [CVE-ID or Issue Type]
   File: [path:line]
   Description: [details]
   Fix: [remediation steps]
```

### Dependency Vulnerabilities
```
📦 Package: [name@version]
   Severity: [critical/high/medium/low]
   CVE: [CVE-ID]
   Fixed in: [version]
   Action: Update to [version]
```

### Code Security Issues
```
🔍 Issue: [vulnerability type]
   Location: [file:line]
   Code: [snippet]
   Risk: [description]
   Remediation: [fix code]
```

### Applied Fixes
```
✅ Fixed: [count] issues automatically
📝 Manual fixes required: [count]
⏳ Pending review: [count]
```

### Compliance Report
- OWASP Top 10 compliance
- CWE coverage analysis
- Security best practices
- Industry standard alignment

### Next Steps
1. Review and apply remaining fixes
2. Update security policies
3. Schedule regular scans
4. Security training needs