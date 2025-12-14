# Feature Specification: CI Test Output Mode

**Feature Branch**: `001-ci-test-output`
**Created**: 2025-12-12
**Status**: Draft
**Input**: CI upgrade plan for native CI mode capabilities in MAGE-X test commands

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Automatic CI Detection (Priority: P1)

As a CI/CD pipeline maintainer, I want the test commands to automatically detect when running in a CI environment and produce structured output, so that I don't need to modify existing workflow configurations.

**Why this priority**: This is the core value proposition - zero-friction adoption. Existing workflows continue working unchanged while gaining improved output. Without auto-detection, users must manually update all their CI configurations.

**Independent Test**: Can be fully tested by running `magex test:unit` in a CI environment (with CI=true) and verifying structured output is produced, while running locally produces standard terminal output.

**Acceptance Scenarios**:

1. **Given** a CI environment (CI=true or GITHUB_ACTIONS=true), **When** user runs `magex test:unit`, **Then** the system automatically enables CI mode and produces structured output
2. **Given** a local development environment (no CI variables set), **When** user runs `magex test:unit`, **Then** the system uses standard terminal output (unchanged behavior)
3. **Given** an explicit `ci=false` parameter, **When** user runs `magex test:unit ci=false` in CI, **Then** CI mode is disabled regardless of environment

---

### User Story 2 - GitHub Actions Integration (Priority: P1)

As a developer reviewing a failed GitHub Actions run, I want to see precise file and line numbers in the GitHub annotations sidebar, so that I can click directly to the failing code without parsing log output.

**Why this priority**: GitHub Actions is the primary CI platform for MAGE-X projects. Direct annotation integration provides immediate value by reducing time to identify failure locations from minutes to seconds.

**Independent Test**: Can be fully tested by triggering a failing test in GitHub Actions and verifying annotations appear in the PR/commit view with clickable file:line links.

**Acceptance Scenarios**:

1. **Given** a test failure in GitHub Actions, **When** the test run completes, **Then** GitHub annotations appear with correct file paths and line numbers
2. **Given** multiple test failures across different files, **When** viewing the GitHub Actions summary, **Then** each failure has its own annotation with precise location
3. **Given** a panic or race condition during tests, **When** the test run completes, **Then** the source location of the panic/race is annotated

---

### User Story 3 - Structured Failure Report (Priority: P2)

As a CI/CD pipeline maintainer, I want test failures to be captured in a structured format, so that I can integrate test results with other tools (dashboards, notifications, issue trackers).

**Why this priority**: While GitHub annotations provide immediate visual feedback, structured output (JSON) enables automation and integration with broader tooling ecosystems.

**Independent Test**: Can be fully tested by running tests with CI mode and verifying a structured output file is created with all failure details.

**Acceptance Scenarios**:

1. **Given** test failures in CI mode, **When** tests complete, **Then** a structured results file is created containing failure details
2. **Given** a test failure, **When** examining the structured output, **Then** each failure includes: test name, package, file, line, error message, and context
3. **Given** tests pass, **When** examining the structured output, **Then** summary shows passed count and no failures array

---

### User Story 4 - Failure Context Capture (Priority: P2)

As a developer debugging a test failure, I want to see the surrounding code context in the failure report, so that I can understand the failure without immediately opening the source file.

**Why this priority**: Context reduces context-switching. Developers can often diagnose simple failures directly from CI output without checking out code locally.

**Independent Test**: Can be fully tested by running a failing test and verifying the output includes lines before and after the failure location.

**Acceptance Scenarios**:

1. **Given** a test failure with assertion error, **When** viewing failure details, **Then** surrounding lines of code from the test file are included
2. **Given** configurable context lines (default 20), **When** user specifies different context size, **Then** the output includes the specified number of context lines
3. **Given** a failure near the start of a file, **When** capturing context, **Then** only available lines are included (graceful handling of boundaries)

---

### User Story 5 - Panic and Race Detection (Priority: P2)

As a developer, I want panic errors and race conditions to be captured with the same precision as test failures, so that I can quickly locate and fix these critical issues.

**Why this priority**: Panics and races are more severe than test failures but harder to locate. The current bash parsing often fails to extract correct locations for these failure types.

**Independent Test**: Can be fully tested by running tests that trigger panics or race conditions and verifying precise source locations are captured.

**Acceptance Scenarios**:

1. **Given** a test that triggers a panic, **When** the test run completes, **Then** the panic location (file:line) is captured in structured output
2. **Given** a test that triggers a data race, **When** running with race detector, **Then** the race location and goroutine information is captured
3. **Given** a build error before tests run, **When** the build fails, **Then** the build error location is captured and reported

---

### User Story 6 - Large Test Suite Handling (Priority: P3)

As a maintainer of a project with thousands of tests, I want CI mode to handle large test suites efficiently, so that test runs don't fail due to memory issues.

**Why this priority**: While most projects have smaller test suites, some MAGE-X consumers run 5,000+ tests. Memory efficiency ensures reliability across the full range of use cases.

**Independent Test**: Can be fully tested by running tests against a project with thousands of tests and monitoring memory usage remains bounded.

**Acceptance Scenarios**:

1. **Given** a test suite with 5,000+ tests, **When** running in CI mode, **Then** memory usage remains bounded (under configurable limit)
2. **Given** memory pressure during test execution, **When** approaching limits, **Then** system switches to more memory-efficient capture strategy
3. **Given** a very large test suite, **When** processing output, **Then** only failures are fully captured; passing tests use minimal memory

---

### User Story 7 - Local CI Mode Preview (Priority: P3)

As a developer, I want to run CI mode locally to preview what CI will see, so that I can debug CI-specific issues without pushing commits.

**Why this priority**: Enables developers to reproduce CI behavior locally, reducing the feedback loop for CI-related issues.

**Independent Test**: Can be fully tested by running `magex test:unit ci` locally and verifying CI-style output is produced.

**Acceptance Scenarios**:

1. **Given** a local development environment, **When** user explicitly runs `magex test:unit ci`, **Then** CI mode is enabled and structured output is produced
2. **Given** CI mode enabled locally, **When** tests complete, **Then** a terminal-friendly summary is displayed alongside structured output file

---

### User Story 8 - Fuzz Test Failure Detection (Priority: P3)

As a developer using fuzz testing, I want fuzz failures to be captured with reproduction information (corpus path, failing input), so that I can quickly reproduce and fix the issue.

**Why this priority**: Fuzz testing is increasingly common in Go projects. Without proper capture, developers must manually parse fuzz output to find reproduction steps.

**Independent Test**: Can be fully tested by running a fuzz test that fails and verifying the corpus path and failing input are captured in structured output.

**Acceptance Scenarios**:

1. **Given** a fuzz test failure, **When** the test run completes, **Then** the corpus file path is captured in structured output
2. **Given** a fuzz test with printable failing input, **When** examining structured output, **Then** the raw input string is included
3. **Given** a fuzz test with binary failing input, **When** examining structured output, **Then** the input is base64 encoded

---

### Edge Cases

- What happens when test output format changes between Go versions? → **Resolved**: Rely on `go test -json` contract stability; document supported Go versions in release notes
- How does system handle tests that write directly to stdout/stderr (bypassing test framework)? → **Resolved**: Capture raw output but don't attempt location extraction; include in test's output blob
- What happens when disk is full and structured output cannot be written? → **Resolved**: Continue test execution; log warning and skip structured output file
- How does system handle nested test suites (subtests) with failures? → **Resolved**: Report full hierarchy as single test name (`Parent/Child/Grandchild`)
- What happens when a test produces millions of lines of output? → **Resolved**: 10MB cap per test; output truncated with warning marker indicating truncation occurred
- How does system handle concurrent test packages writing interleaved output? → **Resolved**: Rely on `go test -json` which handles interleaving natively with per-event package identifiers
- How does system handle when race detection triggers a panic? → **Resolved**: Report as `panic` type with `race_related: true` flag; the panic is the immediate failure while race is the root cause. Both locations captured.
- How does system handle fatal errors that crash the test binary (e.g., SIGSEGV, cgo crash)? → **Resolved**: Detect via non-zero exit code without proper JSON termination. Report as `fatal` type with available stderr output. File:line extracted from crash dump if present.
- How does system handle a test with multiple failure types (e.g., race detected then panic)? → **Resolved**: Report each failure as separate TestFailure entry with same test name. Deduplication uses full signature including type, so both are preserved.

## Requirements *(mandatory)*

### Functional Requirements

#### CI Detection
- **FR-001**: System MUST automatically detect CI environment via standard environment variables (CI, GITHUB_ACTIONS, GITLAB_CI, etc.)
- **FR-002**: System MUST allow explicit CI mode override via command parameter (`ci=true` or `ci=false`)
- **FR-003**: System MUST support configuration file setting for CI mode (`ci_mode: auto/on/off`)
- **FR-004**: Priority order for CI mode MUST be: explicit parameter > environment > config > default (auto)

#### Output Generation
- **FR-005**: System MUST generate GitHub Actions annotations for all failure types when running in GitHub Actions; the JSONL output (FR-006) serves as the generic parseable format for other CI platforms
- **FR-006**: System MUST produce structured output file in JSON Lines format (`.jsonl`) containing one JSON object per line for streaming-friendly processing
- **FR-007**: System MUST capture precise source location (file, line, column where available) for all failures
- **FR-008**: System MUST capture configurable lines of context around each failure
- **FR-009**: System MUST produce a summary suitable for display in GitHub Actions step summary

#### Failure Detection
- **FR-010**: System MUST detect and capture test assertion failures with source location
- **FR-011**: System MUST detect and capture panic errors with stack trace and source location
- **FR-012**: System MUST detect and capture data race warnings with goroutine info and location
- **FR-013**: System MUST detect and capture build errors (compilation failures) with location
- **FR-014**: System MUST detect and capture fuzz test failures with reproduction information including:
  - Corpus file path (e.g., `testdata/fuzz/FuzzFoo/corpus_hash`)
  - Failing input bytes (base64 encoded if binary, raw if printable)
  - Seed corpus vs generated input indicator

#### Memory and Scale
- **FR-015**: System MUST use bounded memory regardless of test suite size
- **FR-016**: System MUST support adaptive memory strategies based on test count
- **FR-017**: System MUST fall back to streaming/disk-based capture for very large suites
- **FR-018**: System MUST cap output capture at 10MB per individual test; when exceeded, truncate and insert warning marker

#### Backwards Compatibility
- **FR-019**: System MUST NOT change existing command signatures or behavior when CI mode is disabled
- **FR-020**: System MUST preserve all existing terminal output when CI mode is disabled
- **FR-021**: System MUST produce structured output even when tests fail (not silently skip on error)

#### Fatal Crash Handling
- **FR-025**: System MUST detect and capture fatal test binary crashes (segfaults, cgo crashes) with available diagnostic output

#### Error Handling
- **FR-022**: System MUST fall back to standard output if CI mode processing fails
- **FR-023**: System MUST log warnings when falling back, enabling debugging
- **FR-024**: System MUST write output to alternative location if primary location fails

### Out of Scope

- **Test execution control**: Parallelism settings, timeout configuration, test filtering, or any modification of how `go test` runs tests
- **Historical analysis**: Test result storage, trend analysis, flaky test detection, or dashboarding across multiple runs
- **Focus**: This feature is purely single-run output formatting and structured capture

### Key Entities

- **TestFailure**: Represents a single test failure with package, test name (full hierarchy for subtests, e.g., `Parent/Child/Grandchild`), error message, source location (file, line, column), failure type, stack trace, context lines, signature (for deduplication: composed of package + test name + failure type + file:line), and duration
- **CIResult**: Aggregated test results containing summary (status, total/passed/failed/skipped counts, duration), failure list, and timestamp
- **CIMode**: Configuration for CI behavior including enabled state, output format, context lines, and memory limits
- **StreamParser**: Real-time processor of test output that maintains state across output lines and captures failures as they occur

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can identify test failure source location within 5 seconds of opening CI results (vs minutes of log parsing)
- **SC-002**: 100% of test failure types (assertions, panics, races, build errors) are captured with source location
- **SC-003**: Zero false positives in failure detection (no incorrect file:line attributions)
- **SC-004**: Memory usage remains under 100MB for test suites up to 10,000 tests
- **SC-005**: Structured output parsing time under 5 seconds for 10,000-test results
- **SC-006**: All existing workflows continue functioning without modification (100% backwards compatible)
- **SC-007**: CI workflow complexity reduced by 90%+ (measured by lines of bash/jq parsing code eliminated)
- **SC-008**: Test failure location accuracy is 100% (every annotated file:line points to actual failure source)

## Clarifications

### Session 2025-12-12

- Q: What format should the structured output file use for failure details (FR-006)? → A: JSON Lines (`.jsonl`) - one JSON object per line, streaming-friendly
- Q: For non-GitHub CI platforms, should the system produce platform-specific annotation formats or only the JSONL file? → A: GitHub annotations plus a generic format others can parse
- Q: When a single test produces excessive output (millions of lines), how should the system handle it? → A: 10MB cap per test, truncate with warning marker
- Q: How should nested subtests (e.g., `TestFoo/SubCase/DeepCase`) be reported in failure output? → A: Full hierarchy - report `Parent/Child/Grandchild` as single test name
- Q: How should the system handle concurrent test packages writing interleaved output? → A: Rely on `go test -json` which handles interleaving natively
- Q: How should the system handle test output format changes between Go versions? → A: Rely on `go test -json` contract stability; document supported Go versions
- Q: How should the system handle tests that write directly to stdout/stderr (bypassing test framework)? → A: Capture raw but don't attempt location extraction; include in test's output blob
- Q: What happens when disk is full and structured output cannot be written? → A: Continue test execution; log warning and skip structured output file
- Q: What should be explicitly out of scope for this feature? → A: Test execution control (parallelism, timeouts, filtering) AND historical trend analysis/storage/dashboarding - focus purely on single-run output formatting
- Q: What fields should compose the TestFailure signature for deduplication? → A: Package + TestName + FailureType + File:Line (most granular)

## Assumptions

- Go test JSON output format (`go test -json`) provides sufficient information for location extraction
- GitHub Actions workflow commands (::error::, ::warning::) remain stable
- Projects using MAGE-X have Go 1.24+ (tested with 1.24; `go test -json` output format is stable across versions)
- Disk space is available for structured output files (typically <1MB per test run)
- Environment variables for CI detection (CI, GITHUB_ACTIONS, etc.) follow industry standard conventions
