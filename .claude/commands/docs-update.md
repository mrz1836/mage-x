---
allowed-tools: Task(mage-x-docs), Task(mage-x-analyzer), Task(mage-x-architect), Task(mage-x-test-finder), Bash(mage docs:*), Read, Write, MultiEdit, Grep, Glob, LS
argument-hint: [feature|api|changelog|all]
description: Update documentation for new or modified features
model: sonnet
---

## Context
- Recent changes: !`git diff --name-only HEAD~10..HEAD | grep -E "\.go$" | head -10`
- Documentation status: !`find . -name "*.md" -mtime -7 | wc -l | xargs echo "Recently updated docs:"`
- API changes: !`git diff HEAD~10..HEAD | grep -E "^[+-]func|^[+-]type" | wc -l | xargs echo "API changes:"`

## Your Task

Update documentation to reflect new or modified features. Focus: ${ARGUMENTS:-all}

### Phase 1: Change Analysis (Parallel)

1. **mage-x-analyzer** - Code Changes:
   - Identify new functions and types
   - Detect API modifications
   - Find deprecated features
   - Analyze breaking changes

2. **mage-x-architect** - Architectural Changes:
   - New interfaces or patterns
   - Modified namespace structure
   - Updated dependencies
   - Architecture evolution

3. **mage-x-test-finder** - Example Discovery:
   - Find test examples
   - Identify usage patterns
   - Extract best practices
   - Document edge cases

### Phase 2: Documentation Generation

**mage-x-docs** coordinates updates for:

1. **API Documentation**:
   - Function signatures
   - Parameter descriptions
   - Return values
   - Usage examples
   - Error conditions

2. **Feature Documentation**:
   - Feature overview
   - Configuration options
   - Integration guides
   - Migration notes

3. **Code Examples**:
   - Working examples
   - Common patterns
   - Best practices
   - Troubleshooting

## Documentation Types

- **feature**: New feature documentation
- **api**: API reference updates
- **changelog**: Version changelog entries
- **all**: Comprehensive documentation update

## Expected Output

### Documentation Updates Summary
```
Files Updated: [count]
New Sections: [count]
API Changes Documented: [count]
Examples Added: [count]
```

### API Documentation
```go
// FunctionName performs specific operation
//
// Parameters:
//   - param1: Description of parameter
//   - param2: Description with constraints
//
// Returns:
//   - Type: Description of return value
//   - error: Possible error conditions
//
// Example:
//   result, err := FunctionName(value1, value2)
//   if err != nil {
//       // Handle error
//   }
```

### Feature Documentation
```markdown
## Feature Name

### Overview
Brief description of the feature and its purpose.

### Configuration
- `option1`: Description (default: value)
- `option2`: Description (required)

### Usage
Step-by-step guide for using the feature.

### Examples
Practical examples demonstrating the feature.

### Troubleshooting
Common issues and solutions.
```

### Migration Guide
For breaking changes:
- What changed
- Why it changed
- How to migrate
- Timeline for deprecation