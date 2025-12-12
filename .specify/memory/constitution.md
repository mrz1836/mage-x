<!--
Sync Impact Report
==================
Version change: 1.0.0 (initial)
Modified principles: N/A (first version)
Added sections: Core Principles (5), Development Workflow, Quality Gates, Governance
Removed sections: N/A
Templates requiring updates:
  - .specify/templates/plan-template.md: ✅ Compatible (Constitution Check section aligns)
  - .specify/templates/spec-template.md: ✅ Compatible (requirements structure aligns)
  - .specify/templates/tasks-template.md: ✅ Compatible (phase structure aligns)
Follow-up TODOs: None
-->

# MAGE-X Constitution

## Core Principles

### I. Zero-Configuration First

Every capability MUST work immediately without requiring configuration files, wrapper functions, or boilerplate code. Users install `magex` and all 400+ commands work instantly in any Go project.

**Requirements:**
- Built-in commands MUST execute without any project-specific setup
- Sensible defaults MUST be provided for all configurable options
- Optional `.mage.yaml` configuration extends but never gates functionality
- Auto-detection MUST identify project type (single binary, multi-binary, library) and adapt behavior accordingly

**Rationale:** Build tools that require extensive setup create friction and reduce adoption. True zero-config means developers can be productive within seconds of installation.

### II. Security-First Architecture

All code execution, input handling, and external interactions MUST follow security-first design patterns. Security is not a feature; it is a foundational constraint.

**Requirements:**
- All user inputs MUST be validated before processing
- Command execution MUST use the `pkg/security/` validation layer
- Secrets and credentials MUST never be logged, exposed, or committed
- Dependencies MUST be audited regularly via `govulncheck`
- OWASP Top 10 vulnerabilities (command injection, XSS, SQL injection, etc.) MUST be actively prevented

**Rationale:** Build automation tools have elevated privileges. A compromised build tool compromises every project that uses it. Security failures are existential risks.

### III. Interface-Based Architecture

All namespaces MUST be defined through Go interfaces with factory functions. This enables testability, extensibility, and clear contracts.

**Requirements:**
- Every namespace (Build, Test, Lint, etc.) MUST have a corresponding interface (e.g., `BuildNamespace`)
- Every namespace MUST provide a factory function (e.g., `NewBuildNamespace()`)
- Implementations MUST satisfy their interface contracts completely
- New functionality MUST extend existing interfaces or create new ones rather than adding standalone functions

**Rationale:** Interface-based design enables mocking for tests, allows multiple implementations, and creates explicit contracts that documentation and tooling can verify.

### IV. Tests Required

All code changes MUST include appropriate test coverage. Tests validate behavior, prevent regressions, and serve as executable documentation.

**Requirements:**
- New features MUST include unit tests demonstrating core functionality
- Bug fixes MUST include regression tests proving the fix works
- Public APIs MUST have integration tests verifying contract compliance
- Test coverage SHOULD remain above 80% for the `pkg/` directory
- Tests MUST pass locally before commits are pushed

**Rationale:** Tests are the safety net that enables confident refactoring and rapid iteration. Untested code is a liability that compounds over time.

### V. Simplicity Over Sophistication

Code MUST be as simple as possible while meeting requirements. Complexity is a cost that must be justified.

**Requirements:**
- YAGNI (You Aren't Gonna Need It): Do not implement speculative features
- Prefer standard library over external dependencies when functionality is equivalent
- Avoid premature abstraction; three similar implementations may be better than one premature generalization
- Configuration options MUST solve real user problems, not hypothetical ones
- Documentation MUST explain the "why" when complexity is unavoidable

**Rationale:** Every line of code is a liability. Simpler code is easier to understand, test, maintain, and debug. Complexity should be earned, not inherited.

## Development Workflow

### Code Organization

- **pkg/mage/**: Core namespace implementations (30+ namespaces)
- **pkg/common/**: Shared utilities (env, fileops, paths)
- **pkg/security/**: Command validation and security infrastructure
- **pkg/utils/**: General-purpose utilities
- **cmd/magex/**: CLI entry point
- **examples/**: Usage demonstrations
- **docs/**: User and developer documentation

### Command Format

- Use kebab-case for commands: `magex test:unit` NOT `magex testUnit`
- Simple aliases available: `test`, `build`, `lint`, `clean`
- Full format follows `namespace:action` pattern

### Contribution Flow

1. Read existing code before proposing changes
2. Create tests for new functionality
3. Implement the minimal solution that passes tests
4. Run `magex lint:fix` and `magex test` before committing
5. Follow conventional commit format for messages

## Quality Gates

### Pre-Commit Requirements

- [ ] `go build ./...` compiles without errors
- [ ] `magex lint` passes without new warnings
- [ ] `magex test:unit` passes all tests
- [ ] `go-pre-commit run --all-files` passes
- [ ] No secrets or credentials in diff

### Pre-Merge Requirements

- [ ] All CI checks pass
- [ ] Code review approved
- [ ] Documentation updated if public API changed
- [ ] CHANGELOG updated for user-facing changes

### Release Requirements

- [ ] All quality gates pass
- [ ] Version follows semantic versioning
- [ ] Release notes document breaking changes
- [ ] `govulncheck` shows no new vulnerabilities

## Governance

This constitution establishes the non-negotiable principles for MAGE-X development. All contributors, whether human or AI, MUST adhere to these rules.

**Amendment Process:**
1. Propose change via GitHub issue with rationale
2. Discuss with maintainers and community
3. Update constitution document with version bump
4. Document migration path if principle changes affect existing code

**Version Policy:**
- MAJOR: Removing or fundamentally redefining a principle
- MINOR: Adding a new principle or section
- PATCH: Clarifying existing principles without changing meaning

**Compliance:**
- All pull requests MUST be reviewed for constitution compliance
- Violations require justification and explicit approval from maintainers
- Repeated violations may result in reverted changes

**Reference Documents:**
- `.github/AGENTS.md` - Detailed technical conventions
- `.github/CLAUDE.md` - AI assistant integration guidelines
- `docs/NAMESPACE_INTERFACES.md` - Architecture reference

**Version**: 1.0.0 | **Ratified**: 2025-07-22 | **Last Amended**: 2025-12-12
