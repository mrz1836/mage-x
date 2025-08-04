---
allowed-tools: Task(mage-x-analyzer), Task(mage-x-architect), Task(mage-x-docs), Task(mage-x-test-finder), Read, Grep, Glob, LS
argument-hint: [function|file|architecture|feature]
description: Multi-agent code explanation and documentation
model: claude-sonnet-4-20250514
---

## Context
- Target: $ARGUMENTS
- Current location: !`pwd`

## Your Task

Provide comprehensive code explanation through parallel agent analysis of: $ARGUMENTS

### Phase 1: Parallel Analysis

Execute simultaneously for comprehensive understanding:

1. **mage-x-analyzer** - Code Structure:
   - Function signatures and parameters
   - Variable usage and data flow
   - Complexity analysis
   - Dependencies and imports
   - Performance characteristics

2. **mage-x-architect** - Design Patterns:
   - Architectural decisions
   - Interface implementations
   - Design pattern usage
   - Namespace integration
   - Registry patterns

3. **mage-x-docs** - Documentation:
   - Existing documentation
   - API specifications
   - Usage examples
   - Related documentation

4. **mage-x-test-finder** - Test Context:
   - How the code is tested
   - Test coverage
   - Example usage in tests
   - Edge cases handled

### Phase 2: Integrated Explanation

Synthesize findings into comprehensive explanation covering:
- **Purpose**: What the code does and why
- **Architecture**: How it fits into the system
- **Implementation**: Key algorithms and logic
- **Usage**: How to use it correctly
- **Testing**: How it's validated
- **Performance**: Efficiency considerations

## Explanation Modes

- **function**: Deep dive into specific function
- **file**: Explain entire file purpose and structure
- **architecture**: System-wide architectural explanation
- **feature**: End-to-end feature explanation

## Expected Output

### Executive Summary
Brief overview of the code's purpose and importance

### Detailed Explanation

#### Purpose & Context
- What problem it solves
- Why it exists
- Who uses it

#### Technical Implementation
```go
// Key code snippets with annotations
```

#### Architecture Integration
- How it connects to other components
- Dependencies and dependents
- Interface contracts

#### Usage Examples
```go
// Practical usage examples
```

#### Testing Strategy
- Test coverage
- Key test scenarios
- Validation approach

#### Performance Considerations
- Time complexity
- Memory usage
- Optimization opportunities

### Related Components
- Similar functionality
- Alternative approaches
- Future improvements
