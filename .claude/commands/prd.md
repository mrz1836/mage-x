---
allowed-tools: Task(mage-x-wizard), Task(mage-x-architect), Task(mage-x-analyzer), Task(mage-x-docs), Read, Write, Grep, Glob, LS
argument-hint: [feature-name]
description: Generate comprehensive product requirement documents
model: claude-sonnet-4-20250514
---

## Context
- Project: MAGE-X Build Automation System
- Feature request: $ARGUMENTS

## Your Task

Create a comprehensive Product Requirements Document (PRD) for: $ARGUMENTS

### Phase 1: Requirements Gathering (Parallel)

1. **mage-x-wizard** - User Story Development:
   - Gather user needs and pain points
   - Define user personas
   - Create user journey maps
   - Identify success criteria

2. **mage-x-architect** - Technical Feasibility:
   - Assess architectural impact
   - Identify integration points
   - Evaluate technical constraints
   - Design high-level solution

3. **mage-x-analyzer** - Impact Analysis:
   - Code complexity assessment
   - Performance implications
   - Resource requirements
   - Risk analysis

### Phase 2: PRD Generation

**mage-x-docs** synthesizes findings into comprehensive PRD covering:

1. **Executive Summary**
2. **Problem Statement**
3. **Solution Overview**
4. **User Stories & Requirements**
5. **Technical Specification**
6. **Success Metrics**
7. **Implementation Plan**
8. **Risk Assessment**

## Expected Output

### Product Requirements Document: $ARGUMENTS

---

## Executive Summary
Brief overview of the feature, its purpose, and expected impact on MAGE-X users.

## Problem Statement

### Current State
- What users currently experience
- Pain points and limitations
- Workarounds being used

### Desired State
- Ideal user experience
- Solved problems
- New capabilities enabled

## User Stories

### Primary User Story
**As a** [user type]  
**I want to** [action]  
**So that** [benefit]

### Acceptance Criteria
- [ ] Criteria 1
- [ ] Criteria 2
- [ ] Criteria 3

## Functional Requirements

### Core Features
1. **Feature 1**: Description
   - Requirement 1.1
   - Requirement 1.2

2. **Feature 2**: Description
   - Requirement 2.1
   - Requirement 2.2

### Non-Functional Requirements
- **Performance**: Response time < 100ms
- **Security**: Follow mage-x security patterns
- **Scalability**: Support 30+ repositories
- **Compatibility**: Cross-platform support

## Technical Specification

### Architecture Changes
```
New Components:
├── pkg/mage/[feature]/
│   ├── interface.go
│   ├── implementation.go
│   └── factory.go
```

### API Design
```go
type FeatureNamespace interface {
    // Core methods
    Execute(ctx context.Context) error
    Validate() error
}
```

### Integration Points
- Existing namespaces affected
- New dependencies required
- Configuration changes

## Success Metrics

### Quantitative Metrics
- Adoption rate: >60% of users within 3 months
- Performance: No regression in build times
- Quality: Zero critical bugs in first release

### Qualitative Metrics
- User satisfaction score
- Developer experience feedback
- Community engagement

## Implementation Plan

### Phase 1: Foundation (Week 1-2)
- Core implementation
- Basic testing
- Documentation

### Phase 2: Integration (Week 3-4)
- Namespace integration
- Agent coordination
- Advanced features

### Phase 3: Polish (Week 5-6)
- Performance optimization
- Comprehensive testing
- Documentation completion

## Risk Assessment

### Technical Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking changes | High | Medium | Feature flags |
| Performance regression | Medium | Low | Benchmarking |

### Timeline Risks
- Dependencies on external teams
- Scope creep potential
- Resource availability

## Open Questions
1. Question requiring stakeholder input
2. Technical decision pending research
3. User experience consideration

## Appendix
- Related documents
- Technical references
- Prior art analysis
