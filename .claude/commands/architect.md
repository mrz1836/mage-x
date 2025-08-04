---
allowed-tools: Task(mage-x-architect), Task(mage-x-analyzer), Task(mage-x-refactor), Task(mage-x-linter), Read, Grep, Glob, LS
argument-hint: [review|validate|suggest]
description: Architecture compliance review and improvement suggestions
model: sonnet
---

## Context
- Namespace count: !`find pkg/mage -name "*.go" -exec grep -l "type.*Namespace" {} \; | wc -l | xargs echo "Namespaces:"`
- Interface definitions: !`find . -name "*.go" -exec grep -l "type.*interface" {} \; | wc -l | xargs echo "Interface files:"`

## Your Task

Perform architecture analysis and provide improvement recommendations. Mode: ${ARGUMENTS:-review}

### Phase 1: Architecture Analysis (Parallel)

1. **mage-x-architect** - Pattern Validation:
   - Validate 30+ namespace architecture
   - Check interface implementations
   - Review factory functions (New*Namespace)
   - Verify registry patterns
   - Analyze dependency structure

2. **mage-x-analyzer** - Structural Analysis:
   - Package coupling metrics
   - Cyclic dependency detection
   - Code organization assessment
   - Layer violation detection

### Phase 2: Compliance Verification

Based on mage-x architectural principles:

1. **Interface-Based Design**:
   - All major components use interfaces
   - Proper abstraction levels
   - Testable implementations
   - Loose coupling

2. **Namespace Pattern**:
   - Consistent namespace structure
   - Proper method organization
   - Clear responsibilities
   - No overlapping concerns

3. **Security Architecture**:
   - CommandExecutor usage
   - Input validation patterns
   - Error handling consistency
   - Security boundaries

### Phase 3: Improvement Generation

**mage-x-refactor** assists with:
- Architectural debt identification
- Refactoring recommendations
- Pattern implementation guides
- Migration strategies

## Analysis Modes

- **review**: Comprehensive architecture review
- **validate**: Compliance checking only
- **suggest**: Focus on improvements

## Expected Output

### Architecture Score
```
Overall Health: [A-F grade]
├── Pattern Compliance: 85%
├── Interface Coverage: 90%
├── Coupling Score: B+
└── Security Patterns: A
```

### Namespace Analysis
```
✅ Compliant Namespaces (28/30):
   - Build, Test, Lint, Security...

⚠️  Issues Found (2/30):
   - Legacy: Missing interface definition
   - Custom: Inconsistent method signatures
```

### Architectural Findings

#### High Priority Issues
```
🔴 Cyclic Dependency: 
   pkg/mage/builder → pkg/mage/test → pkg/mage/builder
   Impact: Tight coupling, difficult to test
   Fix: Extract shared interface to pkg/mage/common

🔴 Layer Violation:
   pkg/mage/ui directly accesses pkg/mage/internal
   Impact: Breaks encapsulation
   Fix: Use public API through proper interfaces
```

#### Pattern Violations
```
⚠️  Factory Function: NewCustomNamespace()
   Issue: Returns concrete type instead of interface
   Fix: Return CustomNamespace interface

⚠️  Registry Pattern: 
   Issue: Direct map access instead of methods
   Fix: Add Get/Set/List methods
```

### Improvement Recommendations

1. **Immediate Actions**:
   - Fix cyclic dependencies
   - Implement missing interfaces
   - Correct layer violations

2. **Short-term Improvements**:
   - Standardize error handling
   - Enhance interface documentation
   - Add architecture tests

3. **Long-term Evolution**:
   - Microservice extraction points
   - Plugin architecture readiness
   - Horizontal scaling preparation

### Architecture Diagrams
```
Current State:          Recommended:
┌─────────┐            ┌─────────┐
│   UI    │            │   UI    │
├─────────┤            ├─────────┤
│  Mage   │            │Interface│
├─────────┤            ├─────────┤
│Internal │            │  Core   │
└─────────┘            ├─────────┤
                       │ Plugins │
                       └─────────┘
```