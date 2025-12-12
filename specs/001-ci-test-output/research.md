# Research: CI Test Output Mode

**Feature Branch**: `001-ci-test-output`
**Date**: 2025-12-12

## Research Tasks

### 1. Go Test JSON Output Format

**Question**: What is the exact structure of `go test -json` output and what information can be extracted?

**Finding**: Go test JSON output produces one JSON object per line (JSON Lines format) with the following structure:

```go
type TestEvent struct {
    Time    time.Time // encodes as an RFC3339-format string
    Action  string    // run, pause, cont, pass, bench, fail, output, skip
    Package string    // package path
    Test    string    // test name (empty for package-level events)
    Elapsed float64   // seconds (for pass/fail events)
    Output  string    // output text (for output events)
}
```

**Actions**:
- `run`: Test has started
- `pause`: Test has been paused (subtests)
- `cont`: Test has continued
- `pass`: Test passed
- `fail`: Test failed
- `skip`: Test was skipped
- `bench`: Benchmark completed
- `output`: Test output line

**Key insight**: File:line information is embedded in the `Output` field, not as structured fields. Must parse patterns like:
- Test failures: `filename_test.go:42: message`
- Panics: `/full/path/file.go:15 +0x39`
- Build errors: `./file.go:10:5: undefined: foo`

**Decision**: Parse `Output` field with regex patterns to extract file:line locations.
**Rationale**: This is the standard approach and matches how other tools (gotestsum, etc.) work.
**Alternatives considered**:
- Use gotestsum: Rejected (adds external dependency, violates constitution principle V)
- Custom test runner: Rejected (too invasive, violates backwards compatibility)

---

### 2. GitHub Actions Workflow Commands

**Question**: What is the format for GitHub Actions annotations?

**Finding**: GitHub Actions supports workflow commands via stdout:

```text
::error file={name},line={line},col={col},endColumn={endColumn},title={title}::{message}
::warning file={name},line={line},col={col},title={title}::{message}
::notice file={name},line={line},col={col},title={title}::{message}
```

For step summary (GITHUB_STEP_SUMMARY):
- Write markdown content to the file path in `$GITHUB_STEP_SUMMARY` environment variable
- Supports full markdown including tables, code blocks, collapsible sections

For outputs (GITHUB_OUTPUT):
- Write `name=value` pairs to the file path in `$GITHUB_OUTPUT` environment variable
- Multiline values use delimiter syntax

**Decision**: Use `::error::` format for individual failures, markdown summary for `GITHUB_STEP_SUMMARY`.
**Rationale**: Native GitHub integration without external dependencies.
**Alternatives considered**:
- GitHub API for annotations: Rejected (requires auth token, network calls)
- Problem matchers: Considered but workflow commands are simpler and more direct

---

### 3. CI Environment Detection

**Question**: What environment variables indicate CI environments?

**Finding**: Standard CI environment variables:

| CI Platform | Primary Variable | Secondary Variables |
|-------------|------------------|---------------------|
| Generic | `CI=true` | - |
| GitHub Actions | `GITHUB_ACTIONS=true` | `GITHUB_RUN_ID`, `GITHUB_WORKFLOW` |
| GitLab CI | `GITLAB_CI=true` | `CI_JOB_ID`, `CI_PIPELINE_ID` |
| CircleCI | `CIRCLECI=true` | `CIRCLE_BUILD_NUM` |
| Travis CI | `TRAVIS=true` | `TRAVIS_BUILD_NUMBER` |
| Jenkins | `JENKINS_URL` set | `BUILD_NUMBER`, `JOB_NAME` |
| Azure Pipelines | `TF_BUILD=True` | `BUILD_BUILDNUMBER` |
| Bitbucket Pipelines | `BITBUCKET_BUILD_NUMBER` set | `BITBUCKET_PIPELINE_UUID` |

**Decision**: Check `CI=true` first (covers most), then platform-specific variables for GitHub Actions features.
**Rationale**: Covers 99%+ of CI environments with simple environment checks.
**Alternatives considered**:
- TTY detection: Rejected (unreliable in containerized environments)
- Config file only: Rejected (violates zero-config principle)

---

### 4. Memory Management for Large Test Suites

**Question**: How to handle test suites with 5,000+ tests without running out of memory?

**Finding**: Strategies for bounded memory:

1. **Streaming processing**: Parse line-by-line, don't buffer entire output
2. **Ring buffer for context**: Keep last N lines for failure context, overwrite oldest
3. **Selective capture**: Full details for failures only, minimal for passes
4. **Disk fallback**: Write to temp file if memory threshold exceeded

Go's `bufio.Scanner` is ideal for line-by-line streaming processing.

Memory estimates:
- Each `TestFailure` struct: ~500 bytes (with 20 lines context)
- 100 failures: ~50KB
- 1000 failures: ~500KB
- Even with 10% failure rate on 10k tests: ~500KB

**Decision**: Streaming parser with ring buffer (20 lines), full capture for failures only.
**Rationale**: Bounded memory regardless of test count; most test runs have <100 failures.
**Alternatives considered**:
- Full output buffering: Rejected (unbounded memory)
- Database storage: Rejected (over-engineering, adds complexity)

---

### 5. Existing Codebase Integration Points

**Question**: How does CI mode integrate with existing mage-x test commands?

**Finding**: Key integration points in current codebase:

1. **CommandRunner interface** (`interfaces.go:177-180`):
   ```go
   type CommandRunner interface {
       RunCmd(name string, args ...string) error
       RunCmdOutput(name string, args ...string) (string, error)
   }
   ```
   Can wrap with CIRunner that intercepts `go test` commands.

2. **GetRunner/SetRunner pattern** (`test.go` uses `GetRunner()`):
   Allows swapping runner implementation for CI mode.

3. **TestConfig** (`config.go:69-91`):
   Add `CIMode` field to existing config structure.

4. **Test namespace** (`test.go`):
   Commands like `Unit()`, `Race()`, `Cover()` can check CI mode and wrap runner.

5. **Existing TestFailure** (`interfaces.go:263-268`):
   ```go
   type TestFailure struct {
       Package string
       Test    string
       Error   string
       Output  string
   }
   ```
   Extend with CI-specific fields (File, Line, Column, etc.).

**Decision**: Wrap CommandRunner with CIRunner, extend TestFailure struct, add CIConfig to TestConfig.
**Rationale**: Minimal changes to existing code, follows established patterns.
**Alternatives considered**:
- New Test namespace methods: Rejected (duplicates functionality)
- Environment variable activation only: Rejected (no config file option)

---

### 6. Test Failure Type Detection Patterns

**Question**: What regex patterns reliably detect different failure types?

**Finding**: Validated patterns from Go test output:

| Failure Type | Pattern | Example |
|--------------|---------|---------|
| Test assertion | `^(\S+\.go):(\d+):(.*)$` | `foo_test.go:42: expected 1, got 2` |
| Build error | `^\.?/?(\S+\.go):(\d+):(\d+): (.*)$` | `./pkg/foo.go:10:5: undefined: bar` |
| Panic | `panic: (.*)` | `panic: runtime error: invalid memory address` |
| Panic location | `^\t(\S+\.go):(\d+) \+0x` | `	/path/foo.go:15 +0x39` |
| Race condition | `WARNING: DATA RACE` | (literal match) |
| Race location | `^\s+(\S+\.go):(\d+) \+0x` | Similar to panic |
| Fuzz failure | `Failing input written to` | (literal match) |

**Decision**: Use compiled regex patterns with named groups for clarity.
**Rationale**: Covers all failure types identified in spec with validated patterns.
**Alternatives considered**:
- String parsing: Rejected (less maintainable, error-prone)
- External parser: Rejected (adds dependency)

---

### 7. Output Format Selection

**Question**: What structured output format should be used?

**Finding**: Options considered:

| Format | Pros | Cons |
|--------|------|------|
| JSON Lines | Streaming-friendly, line-by-line processing | Less human-readable |
| JSON | Universal, well-tooled | Must buffer entire output |
| JUnit XML | CI platform support | Verbose, outdated |
| TAP | Simple, test-focused | Limited metadata support |

**Decision**: JSON Lines (`.jsonl`) as primary format.
**Rationale**:
- Matches `go test -json` output format
- Streaming-friendly (each line is complete JSON)
- Can be processed with `jq` line-by-line
- Failure entries can be written immediately without buffering

**Alternatives considered**:
- JUnit XML: Rejected (verbose, requires buffering)
- TAP: Rejected (limited metadata support)

---

### 8. Backwards Compatibility Approach

**Question**: How to ensure 100% backwards compatibility?

**Finding**: Compatibility requirements:

1. **No signature changes**: `Test{}.Unit(args...)` signature unchanged
2. **Default behavior unchanged**: Without CI detection, output identical to current
3. **Explicit override**: `ci=false` disables CI mode even in CI environment
4. **Additive output**: CI mode adds structured files, doesn't replace stdout

**Decision**:
- CI mode only activates on auto-detection OR explicit `ci` parameter
- Terminal output preserved (tee to parser AND stdout)
- Structured output is additional, written to `.mage-x/ci-results.jsonl`

**Rationale**: Existing scripts and workflows continue working unchanged.
**Alternatives considered**:
- New command names (`test:ci`): Rejected (fragments namespace, confusing)
- Replace stdout entirely: Rejected (breaks existing workflows)

---

### 9. Fuzz Test Output Patterns

**Question**: What patterns does `go test -fuzz` output for failures?

**Finding**: Fuzz test failure output includes:

```text
--- FAIL: FuzzFoo (0.15s)
    --- FAIL: FuzzFoo/seed#0 (0.00s)
        foo_test.go:42: failed assertion

    Failing input written to testdata/fuzz/FuzzFoo/582528ddfad69eb5
    To re-run:
    go test -run=FuzzFoo/582528ddfad69eb5
```

**Patterns**:
- Corpus path: `Failing input written to (\S+)`
- Re-run command: `go test -run=(\S+)`
- Seed vs generated: `/seed#` in test name indicates seed corpus

**Decision**: Capture corpus path, parse failing input file, detect seed vs generated.
**Rationale**: Full reproduction info enables developers to re-run exact failing case.
**Alternatives considered**:
- Capture only corpus path: Rejected (input bytes help immediate diagnosis)
- Always store raw bytes: Rejected (binary data needs base64 encoding)

---

### 10. Fatal Crash Detection

**Question**: How to detect when the test binary crashes (SIGSEGV, cgo crash)?

**Finding**: Fatal crashes have distinct characteristics:
- Non-zero exit code from `go test`
- Incomplete JSON output (no final summary event)
- Crash dump on stderr (not stdout)

Crash dump pattern (runtime/debug.Stack output):
```text
fatal error: unexpected signal during runtime execution
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x...]

goroutine 1 [running]:
runtime/debug.Stack()
        /usr/local/go/src/runtime/debug/stack.go:24 +0x65
```

**Decision**: Detect via exit code + incomplete JSON; parse stderr for crash details.
**Rationale**: Handles all crash types including cgo, memory corruption, and signal-based crashes.
**Alternatives considered**:
- Rely on JSON only: Rejected (crashes may not produce valid JSON)
- Ignore crashes: Rejected (these are critical failures)

---

## Summary of Decisions

| Topic | Decision | Primary Rationale |
|-------|----------|-------------------|
| Test output parsing | Parse `go test -json` Output field | Standard format, no dependencies |
| Fuzz failures | Capture corpus path, input, seed indicator | Full reproduction info |
| Fatal crashes | Detect via exit code + incomplete JSON | Handles all crash types |
| GitHub integration | Workflow commands (`::error::`) | Native, no auth required |
| CI detection | Environment variables (CI, GITHUB_ACTIONS, etc.) | Universal, simple |
| Memory management | Streaming parser with ring buffer | Bounded memory, handles scale |
| Integration | Wrap CommandRunner, extend TestFailure | Follows existing patterns |
| Failure patterns | Compiled regex | Maintainable, validated |
| Output format | JSON Lines (`.jsonl`) | Streaming-friendly, matches Go |
| Backwards compat | Additive output, explicit override | Zero breaking changes |
