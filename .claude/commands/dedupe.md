---
allowed-tools: Task(mage-x-analyzer), Task(mage-x-refactor), Task(mage-x-architect), Task(mage-x-linter), Read, Write, MultiEdit, Grep, Glob, LS
argument-hint: [aggressive|conservative]
description: Remove duplicate code through intelligent refactoring
model: claude-sonnet-4-20250514
---

## Context
- Project size: !`find . -name "*.go" -not -path "./vendor/*" | wc -l | xargs echo "Go files:"`
- Code complexity: !`find . -name "*.go" -exec wc -l {} + | awk '{total += $1} END {print "Total LOC:", total}'`

## Your Task

Identify and remove duplicate code through coordinated agent analysis. Mode: ${ARGUMENTS:-conservative}

### Phase 1: Parallel Duplicate Detection

1. **mage-x-analyzer** - Code Analysis:
   - Identify duplicate functions and methods
   - Find similar code patterns
   - Detect copy-paste violations
   - Analyze structural duplicates
   - Calculate duplication metrics

2. **mage-x-architect** - Pattern Analysis:
   - Identify common patterns that could be abstracted
   - Find repeated interface implementations
   - Detect similar namespace structures
   - Analyze factory function duplicates

### Phase 2: Refactoring Strategy

Based on Phase 1 findings:

1. **mage-x-refactor** - Code Consolidation:
   - Extract common functions to shared utilities
   - Create generic implementations
   - Implement interface-based abstractions
   - Consolidate similar logic
   - Apply DRY principles

2. **mage-x-linter** - Quality Validation:
   - Ensure refactoring maintains quality
   - Validate no functionality broken
   - Check for proper documentation
   - Verify test coverage maintained

### Deduplication Modes

- **conservative**: Only remove exact duplicates with high confidence
- **aggressive**: Also refactor similar patterns and near-duplicates

## Refactoring Priorities

1. **Exact Duplicates**: Identical code blocks
2. **Functional Duplicates**: Same logic, different names
3. **Structural Duplicates**: Similar patterns
4. **Partial Duplicates**: Overlapping functionality

## Expected Output

### Duplication Analysis
```
Total Duplicate Lines: [count]
Duplicate Blocks: [count]
Affected Files: [count]
Potential Savings: [percentage]%
```

### Refactoring Actions
1. **Functions Consolidated**: List of merged functions
2. **Utilities Created**: New shared utilities
3. **Patterns Abstracted**: Interface implementations
4. **Code Removed**: Lines eliminated

### Quality Impact
- Test coverage maintained: ✅
- No breaking changes: ✅
- Documentation updated: ✅
- Performance impact: Neutral/Positive
