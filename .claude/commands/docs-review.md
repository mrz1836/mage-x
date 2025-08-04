---
allowed-tools: Task(mage-x-docs), Task(mage-x-analyzer), Task(mage-x-refactor), Read, Grep, Glob, LS
argument-hint: [strict|permissive]
description: Review if documented features still exist in codebase
model: claude-sonnet-4-20250514
---

## Context
- Documentation files: !`find . -name "*.md" -not -path "./vendor/*" | wc -l | xargs echo "Markdown files:"`
- Code references: !`grep -r "func\|type\|interface" docs/ 2>/dev/null | wc -l | xargs echo "Code references in docs:"`

## Your Task

Review documentation accuracy and identify outdated content. Mode: ${ARGUMENTS:-strict}

### Phase 1: Documentation Analysis

**mage-x-docs** performs comprehensive review:

1. **Code Reference Validation**:
   - Verify function signatures exist
   - Check type definitions
   - Validate interface contracts
   - Confirm constant values

2. **Example Verification**:
   - Test code examples compile
   - Verify import paths
   - Check API usage correctness
   - Validate output claims

3. **Link Validation**:
   - Check internal links
   - Verify file references
   - Validate external URLs
   - Find broken anchors

### Phase 2: Cross-Reference Analysis

**mage-x-analyzer** assists with:

1. **Feature Existence**:
   - Confirm documented features exist
   - Find renamed functionality
   - Identify deprecated features
   - Locate moved code

2. **API Accuracy**:
   - Compare documented vs actual APIs
   - Check parameter types
   - Verify return values
   - Validate error messages

### Phase 3: Remediation

**mage-x-refactor** helps fix issues:

1. **Documentation Updates**:
   - Update outdated references
   - Fix incorrect examples
   - Remove deprecated content
   - Add missing documentation

2. **Code Alignment**:
   - Add missing features
   - Restore removed functionality
   - Update deprecation notices

## Review Modes

- **strict**: Flag any discrepancy
- **permissive**: Only major issues

## Expected Output

### Review Summary
```
Documentation Accuracy: [percentage]%
Outdated References: [count]
Broken Examples: [count]
Dead Links: [count]
```

### Critical Issues
```
🚨 REMOVED FEATURE: FunctionName
   File: docs/api.md:45
   Status: Function no longer exists
   Action: Remove documentation or restore feature

🚨 BROKEN EXAMPLE: 
   File: docs/examples.md:120
   Error: undefined: OldTypeName
   Fix: Update to NewTypeName
```

### Outdated References
```
📝 File: README.md
   Line 234: References "mage buildAll" (now "mage build:all")
   Line 567: Links to deleted file "old-guide.md"

📝 File: docs/configuration.md
   Line 89: Describes removed config option "legacyMode"
   Line 123: Incorrect default value for "timeout"
```

### Validation Results
```
✅ Valid Code Examples: 45/50
❌ Invalid Examples: 5/50
✅ Working Links: 78/80  
❌ Broken Links: 2/80
```

### Recommended Actions
1. **Immediate**: Fix broken examples
2. **High Priority**: Update removed features
3. **Medium Priority**: Correct parameter docs
4. **Low Priority**: Update minor details
