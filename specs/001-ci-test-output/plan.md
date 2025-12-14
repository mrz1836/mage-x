# Implementation Plan: CI Test Output Mode

**Branch**: `001-ci-test-output` | **Date**: 2025-12-12 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-ci-test-output/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Add native CI mode capabilities to MAGE-X test commands that automatically detect CI environments and produce structured output (GitHub annotations, JSON Lines) with precise file:line locations for failures. This eliminates 1,700+ lines of complex bash/jq parsing in CI workflows while maintaining 100% backwards compatibility.

## Technical Context

**Language/Version**: Go 1.24+ (tested with 1.24)
**Primary Dependencies**: magefile/mage (build system), standard library (encoding/json, io, sync)
**Storage**: N/A (file-based output: `.mage-x/ci-results.jsonl`)
**Testing**: `go test` with race detector, fuzz tests, benchmarks
**Target Platform**: Linux (CI), macOS, Windows (cross-platform CLI)
**Project Type**: Single binary (CLI tool)
**Performance Goals**: <5 seconds parsing for 10,000 tests, <100MB memory for 10k+ test suites
**Constraints**: <100MB memory for 10k tests, streaming output for large suites, 10MB cap per individual test output
**Scale/Scope**: 30+ projects use mage-x as build/dev tool, test suites ranging from 100 to 5,000+ tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Zero-Configuration First ✅
- **Requirement**: CI mode auto-detects from environment (CI=true, GITHUB_ACTIONS=true, GITLAB_CI=true)
- **Status**: PASS - No configuration required; existing workflows continue unchanged
- **Evidence**: FR-001 specifies auto-detection, FR-019 ensures backwards compatibility

### II. Security-First Architecture ✅
- **Requirement**: Validate inputs, use pkg/security/ validation layer
- **Status**: PASS - Output paths are controlled, no user-provided execution paths
- **Evidence**: Output limited to `.mage-x/ci-results.jsonl` (controlled path)

### III. Interface-Based Architecture ✅
- **Requirement**: Namespaces via Go interfaces with factory functions
- **Status**: PASS - Will extend existing Test namespace, CommandRunner interface
- **Evidence**: Plan uses existing `CommandRunner` interface pattern from interfaces.go:177

### IV. Tests Required ✅
- **Requirement**: Unit tests for new features, 80%+ coverage
- **Status**: PASS - Original plan specifies 90%+ coverage, tests at each session
- **Evidence**: Each implementation session ends with `magex lint` and `magex test:unit`

### V. Simplicity Over Sophistication ✅
- **Requirement**: YAGNI, prefer standard library, avoid premature abstraction
- **Status**: PASS - Uses standard library (encoding/json, io, sync), no new dependencies
- **Evidence**: No external dependencies added; extends existing patterns

## Project Structure

### Documentation (this feature)

```text
specs/001-ci-test-output/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
pkg/mage/
├── interfaces.go        # MODIFY: Extend TestFailure struct with CI fields
├── config.go            # MODIFY: Add CIConfig to TestConfig
├── test.go              # MODIFY: Wire CI mode into test commands
├── ci_detector.go       # NEW: CI environment detection
├── ci_runner.go         # NEW: CommandRunner wrapper for CI mode
├── ci_stream_parser.go  # NEW: Real-time Go test output parser
├── ci_reporter_github.go # NEW: GitHub Actions annotation output
├── ci_reporter_json.go  # NEW: JSON Lines structured output
├── ci_detector_test.go  # NEW: Tests for CI detection
├── ci_runner_test.go    # NEW: Tests for CI runner
├── ci_stream_parser_test.go # NEW: Tests for stream parser
├── ci_reporter_github_test.go # NEW: Tests for GitHub reporter
└── ci_reporter_json_test.go # NEW: Tests for JSON reporter
```

**Structure Decision**: Single project (Go CLI). All CI mode implementation resides in `pkg/mage/` following existing namespace patterns. New files prefixed with `ci_` to group related functionality.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. All constitution principles are satisfied by this design.

---

## Constitution Check: Post-Design Re-evaluation ✅

*Performed after Phase 1 design artifacts generated.*

### I. Zero-Configuration First ✅ (Confirmed)
- **Design evidence**: `CIDetector` interface auto-detects via environment variables
- **Default config**: `DefaultCIMode()` provides sensible defaults without any YAML required
- **Explicit override**: Only when user wants non-default behavior

### II. Security-First Architecture ✅ (Confirmed)
- **File paths**: Output path validated, defaults to `.mage-x/ci-results.jsonl`
- **No command injection**: Uses `CommandRunner` interface, not shell execution
- **Input validation**: Regex patterns are compile-time constants

### III. Interface-Based Architecture ✅ (Confirmed)
- **New interfaces**: `CIDetector`, `StreamParser`, `CIReporter`, `CIRunner`
- **Factory functions**: `NewCIDetector()`, `NewStreamParser()`, `NewGitHubReporter()`, `NewJSONReporter()`, `NewCIRunner()`
- **Contracts documented**: See `contracts/interfaces.go.txt`

### IV. Tests Required ✅ (Confirmed)
- **Test files planned**: One test file per implementation file
- **Test patterns**: Unit tests, fuzz tests for parser, benchmarks for memory
- **Coverage target**: 90%+ maintained from original plan

### V. Simplicity Over Sophistication ✅ (Confirmed)
- **Standard library only**: `encoding/json`, `io`, `sync`, `regexp`, `bufio`
- **No external dependencies**: Doesn't add any new go.mod requires
- **Ring buffer**: Simple fixed-size slice, not imported library
- **Streaming**: Uses `bufio.Scanner` from standard library
