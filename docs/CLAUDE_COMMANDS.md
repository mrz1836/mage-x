# 🤖 Claude Code Commands for MAGE-X

> Optimized slash commands leveraging the 19-agent ecosystem for maximum development efficiency

## Overview

MAGE-X provides 13 specialized Claude Code slash commands that orchestrate intelligent agent collaboration for common development workflows. These commands are designed to be:

- **Easy to remember**: Short, intuitive names
- **Performant**: Parallel agent execution where possible
- **Comprehensive**: Cover the full development lifecycle
- **Intelligent**: Agents coordinate automatically for optimal results

## Command Directory

### 🧪 Testing & Quality Commands

#### `/test [specific-path-or-namespace]`
**Purpose**: Create comprehensive Go tests with parallel agent execution
**Agents**: test-finder, test-writer, analyzer, security, linter, architect
**Example**: `/test pkg/mage/build`

Orchestrates multiple agents to:
- Identify untested code and prioritize by criticality
- Analyze complexity and security requirements
- Generate comprehensive test suites with parallel execution
- Validate test quality and patterns

#### `/fix [all|lint|test|security]`
**Purpose**: Fix linting issues and test errors with parallel remediation
**Agents**: linter, test-writer, security, refactor
**Example**: `/fix lint`

Coordinates fixes across:
- Auto-fixable linting issues
- Failing test assertions
- Security vulnerabilities
- Code refactoring needs

#### `/quality [full|quick|security|performance]`
**Purpose**: Comprehensive code quality assessment
**Agents**: linter, test-finder, security, analyzer, architect, benchmark
**Example**: `/quality full`

Provides parallel analysis of:
- Code quality metrics
- Test coverage gaps
- Security posture
- Performance profile
- Architecture compliance

### 🔧 Code Improvement Commands

#### `/dedupe [aggressive|conservative]`
**Purpose**: Remove duplicate code through intelligent refactoring
**Agents**: analyzer, refactor, architect, linter
**Example**: `/dedupe conservative`

Identifies and eliminates:
- Duplicate functions and logic
- Similar code patterns
- Redundant implementations
- Copy-paste violations

#### `/explain [function|file|architecture]`
**Purpose**: Multi-agent code explanation and documentation
**Agents**: analyzer, architect, docs, test-finder
**Example**: `/explain architecture`

Provides comprehensive explanations:
- Code functionality and purpose
- Architectural decisions
- Dependencies and interactions
- Usage examples

#### `/optimize [performance|memory|build]`
**Purpose**: Performance analysis and optimization
**Agents**: benchmark, analyzer, refactor, builder
**Example**: `/optimize performance`

Delivers optimization through:
- Performance bottleneck identification
- Memory usage analysis
- Build time improvements
- Algorithmic enhancements

### 📚 Documentation Commands

#### `/docs-update [feature|api|all]`
**Purpose**: Update documentation for new or modified features
**Agents**: docs, analyzer, architect, test-finder
**Example**: `/docs-update api`

Automatically updates:
- API documentation
- Feature descriptions
- Usage examples
- Migration guides

#### `/docs-review`
**Purpose**: Review if documented features still exist
**Agents**: docs, analyzer, refactor
**Example**: `/docs-review`

Validates documentation:
- Feature existence verification
- API accuracy
- Example validity
- Dead link detection

### 🚀 Development Workflow Commands

#### `/ci-diagnose [workflow-name|run-id|all]`
**Purpose**: Diagnose CI/CD issues in GitHub workflows
**Agents**: gh, security, tools
**Example**: `/ci-diagnose fortress.yml`

Analyzes and fixes:
- Workflow failures
- Environment issues
- Tool compatibility
- Security problems

#### `/release [version|channel|dry-run]`
**Purpose**: Comprehensive release preparation
**Agents**: releaser, git, gh, security, docs, linter
**Example**: `/release v1.2.0`

Orchestrates complete release:
- Pre-release validation
- Version management
- Asset generation
- Documentation updates
- Security scanning

### 🏗️ Architecture Commands

#### `/architect [review|validate|suggest]`
**Purpose**: Architecture compliance and improvement
**Agents**: architect, analyzer, refactor, linter
**Example**: `/architect review`

Provides architectural:
- Pattern validation
- Interface compliance
- Dependency analysis
- Improvement suggestions

#### `/security [full|deps|code|secrets|fix]`
**Purpose**: Comprehensive security audit and fixes
**Agents**: security, deps, refactor
**Example**: `/security full`

Performs security:
- Vulnerability scanning
- Dependency audits
- Secret detection
- Automated remediation

### 🎯 Planning Commands

#### `/prd [feature-name]`
**Purpose**: Generate product requirement documents
**Agents**: architect, analyzer, docs
**Example**: `/prd authentication-system`

Creates comprehensive PRDs:
- Feature specifications
- Technical requirements
- Implementation strategy
- Success criteria

## Usage Patterns

### Basic Usage
```bash
# Simple command
/test

# With arguments
/fix lint

# With specific targets
/optimize performance
```

### Advanced Usage
```bash
# Specific namespace testing
/test pkg/mage/security

# Comprehensive quality check
/quality full

# Dry-run release
/release dry-run
```

### Parallel Execution

Commands automatically leverage parallel agent execution where beneficial:

```
/quality full
├── mage-x-linter     ─┐
├── mage-x-test-finder ├─ Parallel Phase 1
├── mage-x-security    │
├── mage-x-analyzer    │
├── mage-x-architect   │
└── mage-x-benchmark   ─┘
    │
    └── Integrated Analysis → Results
```

## Best Practices

### 1. Start with Quality
Begin new features with `/quality quick` to understand the codebase state.

### 2. Test-Driven Development
Use `/test` before implementing features to establish test patterns.

### 3. Regular Security Scans
Run `/security deps` regularly to catch vulnerabilities early.

### 4. Pre-Release Validation
Always use `/release dry-run` before actual releases.

### 5. Fix Issues Promptly
Use `/fix all` after major changes to maintain code quality.

## Command Efficiency Tips

### Parallel Optimization
Commands are designed for parallel execution. Avoid sequential thinking:

```bash
# ✅ Good: Enables parallel analysis
/quality full

# ❌ Less optimal: Forces sequential execution
/test then /lint then /security
```

### Targeted Execution
Use arguments to focus agent efforts:

```bash
# Specific path testing
/test pkg/critical/auth

# Focused security scan
/security deps

# Quick quality check
/quality quick
```

### Agent Coordination
Let agents coordinate automatically rather than running them separately:

```bash
# ✅ Good: Coordinated analysis
/explain architecture

# ❌ Less optimal: Manual coordination
Run analyzer, then architect, then docs
```

## Integration with MAGE-X

All commands integrate seamlessly with MAGE-X build system:

- Respect `.mage.yaml` configuration
- Use mage commands where available
- Follow project conventions
- Maintain compatibility

## Performance Characteristics

| Command     | Agents Used | Parallel Phases | Typical Duration |
|-------------|-------------|-----------------|------------------|
| `/test`     | 6           | 3               | 30-60s           |
| `/quality`  | 6           | 2               | 20-40s           |
| `/fix`      | 4           | 3               | 15-30s           |
| `/release`  | 6           | 3               | 45-90s           |
| `/security` | 4           | 3               | 25-45s           |

## Troubleshooting

### Command Not Found
Ensure commands are in `.claude/commands/` directory.

### Agent Timeout
Some operations may take time. Agents will provide progress updates.

### Insufficient Permissions
Some commands require specific tool access. Check error messages for required permissions.

## Future Enhancements

Planned improvements:
- Custom command creation tools
- Command composition and pipelines
- Performance metrics tracking
- Team collaboration features

---

For more information about the MAGE-X Agent Ecosystem, see [SUB_AGENTS.md](SUB_AGENTS.md).
