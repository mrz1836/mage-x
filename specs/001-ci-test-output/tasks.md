# Tasks: CI Test Output Mode

**Input**: Design documents from `/specs/001-ci-test-output/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests will be included alongside implementation following Go testing conventions. Each new file gets a corresponding `_test.go` file.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Project Type**: Single binary (Go CLI)
- **Implementation**: `pkg/mage/` for all CI mode code
- **Tests**: `pkg/mage/*_test.go` alongside implementation
- **Output**: `.mage-x/ci-results.jsonl` for structured output

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Core types and constants that all components depend on

- [X] T001 Define FailureType constants and FuzzInfo struct in pkg/mage/ci_types.go
- [X] T002 [P] Define CIMode configuration struct with defaults in pkg/mage/ci_types.go
- [X] T003 [P] Define CIResult and CISummary structs in pkg/mage/ci_types.go
- [X] T004 [P] Define CIMetadata struct in pkg/mage/ci_types.go
- [X] T005 Extend TestFailure struct with CI fields (File, Line, Column, Type, etc.) in pkg/mage/interfaces.go
- [X] T006 Add CIMode field to TestConfig in pkg/mage/config.go
- [X] T007 [P] Create unit tests for type validation in pkg/mage/ci_types_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**Note**: These components are required by all user stories - CIDetector for auto-detection, StreamParser for output processing, base reporter interface for output

- [X] T008 Define CIDetector interface in pkg/mage/ci_detector.go
- [X] T009 Implement CIDetector with environment variable checking (CI, GITHUB_ACTIONS, GITLAB_CI, etc.) in pkg/mage/ci_detector.go
- [X] T010 Implement GetConfig with priority order (param > env > config > default) in pkg/mage/ci_detector.go
- [X] T011 [P] Create unit tests for CIDetector in pkg/mage/ci_detector_test.go
- [X] T012 Define CIReporter interface in pkg/mage/ci_reporter.go
- [X] T013 [P] Define RingBuffer for context line capture in pkg/mage/ci_stream_parser.go
- [X] T014 Define StreamParser interface and TestEventHandler in pkg/mage/ci_stream_parser.go
- [X] T015 Implement core StreamParser with line-by-line JSON parsing in pkg/mage/ci_stream_parser.go
- [X] T016 Implement failure location extraction regex patterns (test, build, panic, race) in pkg/mage/ci_stream_parser.go
- [X] T017 Implement signature generation for deduplication in pkg/mage/ci_stream_parser.go
- [X] T018 Implement context capture from source files in pkg/mage/ci_stream_parser.go
- [X] T019 Create unit tests for StreamParser in pkg/mage/ci_stream_parser_test.go
- [X] T020 [P] Add fuzz tests for StreamParser robustness in pkg/mage/ci_stream_parser_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Automatic CI Detection (Priority: P1) MVP

**Goal**: Test commands automatically detect CI environment and enable structured output without configuration changes

**Independent Test**: Run `magex test:unit` with CI=true set and verify structured output is produced; run without CI=true and verify standard terminal output

### Implementation for User Story 1

- [X] T021 [US1] Define CIRunner interface wrapping CommandRunner in pkg/mage/ci_runner.go
- [X] T022 [US1] Implement CIRunner that intercepts `go test -json` output in pkg/mage/ci_runner.go
- [X] T023 [US1] Implement output teeing (parser AND stdout) for backwards compatibility in pkg/mage/ci_runner.go
- [X] T024 [US1] Implement GetResults() to return collected CIResult in pkg/mage/ci_runner.go
- [X] T025 [US1] Create unit tests for CIRunner in pkg/mage/ci_runner_test.go
- [X] T026 [US1] Wire CIRunner into Test namespace Unit() function in pkg/mage/test.go
- [X] T027 [US1] Wire CIRunner into Test namespace Race() function in pkg/mage/test.go
- [X] T028 [US1] Wire CIRunner into Test namespace Cover() function in pkg/mage/test.go
- [X] T029 [US1] Add `ci` parameter handling to test commands in pkg/mage/test.go
- [X] T030 [US1] Create integration test verifying CI auto-detection in pkg/mage/ci_runner_test.go

**Checkpoint**: User Story 1 complete - CI mode auto-detects and captures failures

---

## Phase 4: User Story 2 - GitHub Actions Integration (Priority: P1)

**Goal**: Test failures produce GitHub annotations with precise file:line locations that appear in PR sidebar

**Independent Test**: Trigger failing test in GitHub Actions and verify annotations appear with clickable file:line links

### Implementation for User Story 2

- [X] T031 [P] [US2] Implement GitHubReporter interface in pkg/mage/ci_reporter_github.go
- [X] T032 [US2] Implement ReportFailure() generating ::error:: workflow commands in pkg/mage/ci_reporter_github.go
- [X] T033 [US2] Implement WriteStepSummary() writing markdown to GITHUB_STEP_SUMMARY in pkg/mage/ci_reporter_github.go
- [X] T034 [US2] Implement WriteOutputs() for GITHUB_OUTPUT in pkg/mage/ci_reporter_github.go
- [X] T035 [US2] Implement WriteSummary() with test result table in pkg/mage/ci_reporter_github.go
- [X] T036 [US2] Create unit tests for GitHubReporter in pkg/mage/ci_reporter_github_test.go
- [X] T037 [US2] Wire GitHubReporter into CIRunner when Platform() == "github" in pkg/mage/ci_runner.go

**Checkpoint**: User Story 2 complete - GitHub Actions shows clickable annotations

---

## Phase 5: User Story 3 - Structured Failure Report (Priority: P2)

**Goal**: Test failures captured in JSON Lines format for integration with dashboards and automation tools

**Independent Test**: Run tests with CI mode and verify `.mage-x/ci-results.jsonl` is created with valid JSON objects per line

### Implementation for User Story 3

- [X] T038 [P] [US3] Implement JSONReporter interface in pkg/mage/ci_reporter_json.go
- [X] T039 [US3] Implement Start() writing metadata line to JSONL file in pkg/mage/ci_reporter_json.go
- [X] T040 [US3] Implement ReportFailure() writing failure objects to JSONL in pkg/mage/ci_reporter_json.go
- [X] T041 [US3] Implement WriteSummary() writing final summary line in pkg/mage/ci_reporter_json.go
- [X] T042 [US3] Implement file creation with .mage-x directory handling in pkg/mage/ci_reporter_json.go
- [X] T043 [US3] Create unit tests for JSONReporter in pkg/mage/ci_reporter_json_test.go
- [X] T044 [US3] Wire JSONReporter to always run (in addition to GitHubReporter when applicable) in pkg/mage/ci_runner.go

**Checkpoint**: User Story 3 complete - JSONL file captures all failures

---

## Phase 6: User Story 4 - Failure Context Capture (Priority: P2)

**Goal**: Failure reports include surrounding lines of code for quick diagnosis without opening source files

**Independent Test**: Run failing test and verify output includes lines before and after failure location

### Implementation for User Story 4

- [X] T045 [US4] Implement captureContext() reading lines around failure in pkg/mage/ci_stream_parser.go
- [X] T046 [US4] Add boundary handling for files with fewer lines than context window in pkg/mage/ci_stream_parser.go
- [X] T047 [US4] Wire context capture into failure processing pipeline in pkg/mage/ci_stream_parser.go
- [X] T048 [US4] Add context_lines configuration option to CIMode in pkg/mage/ci_types.go
- [X] T049 [US4] Create unit tests for context capture in pkg/mage/ci_stream_parser_test.go

**Checkpoint**: User Story 4 complete - Failures include code context

---

## Phase 7: User Story 5 - Panic and Race Detection (Priority: P2)

**Goal**: Panics and race conditions captured with precise source locations, same as test failures

**Independent Test**: Run tests that trigger panic or race condition, verify source location is captured correctly

### Implementation for User Story 5

- [X] T050 [US5] Add panic detection regex pattern (panic: message + stack location) in pkg/mage/ci_stream_parser.go
- [X] T051 [US5] Add race condition detection pattern (WARNING: DATA RACE + location) in pkg/mage/ci_stream_parser.go
- [X] T052 [US5] Implement stack trace capture for panics in pkg/mage/ci_stream_parser.go
- [X] T053 [US5] Add RaceRelated flag handling when race triggers panic in pkg/mage/ci_stream_parser.go
- [X] T054 [US5] Add build error detection pattern in pkg/mage/ci_stream_parser.go
- [X] T055 [US5] Create unit tests for panic detection in pkg/mage/ci_stream_parser_test.go
- [X] T056 [US5] Create unit tests for race detection in pkg/mage/ci_stream_parser_test.go
- [X] T057 [US5] Create unit tests for build error detection in pkg/mage/ci_stream_parser_test.go

**Checkpoint**: User Story 5 complete - All failure types captured with locations

---

## Phase 8: User Story 6 - Large Test Suite Handling (Priority: P3)

**Goal**: CI mode handles 5,000+ tests efficiently with bounded memory

**Independent Test**: Run tests against large test suite and verify memory usage remains under configured limit

### Implementation for User Story 6

- [X] T058 [US6] Implement adaptive CaptureStrategy selection based on test count in pkg/mage/ci_stream_parser.go
- [X] T059 [US6] Implement StrategySmartCapture (full for failures only) in pkg/mage/ci_stream_parser.go
- [X] T060 [US6] Implement StrategyEfficientCapture (limited context) in pkg/mage/ci_stream_parser.go
- [X] T061 [US6] Implement StrategyStreamingCapture (disk buffer fallback) in pkg/mage/ci_stream_parser.go
- [X] T062 [US6] Implement 10MB per-test output cap with truncation marker in pkg/mage/ci_stream_parser.go
- [X] T063 [US6] Add max_memory_mb configuration to CIMode in pkg/mage/ci_types.go
- [X] T064 [US6] Create benchmark tests for memory usage in pkg/mage/ci_stream_parser_test.go
- [X] T065 [US6] Create unit tests for adaptive strategy selection in pkg/mage/ci_stream_parser_test.go

**Checkpoint**: User Story 6 complete - Large test suites handled efficiently

---

## Phase 9: User Story 7 - Local CI Mode Preview (Priority: P3)

**Goal**: Developers can run CI mode locally to preview what CI will see

**Independent Test**: Run `magex test:unit ci` locally and verify CI-style output is produced with both terminal summary and JSONL file

### Implementation for User Story 7

- [X] T066 [US7] Implement explicit `ci` parameter to force CI mode in pkg/mage/test.go
- [X] T067 [US7] Implement `ci=false` parameter to disable CI mode in CI in pkg/mage/test.go
- [X] T068 [US7] Add terminal-friendly summary output when CI mode enabled locally in pkg/mage/ci_reporter_terminal.go
- [X] T069 [US7] Create integration test for local CI mode preview in pkg/mage/ci_reporter_terminal_test.go

**Checkpoint**: User Story 7 complete - Local CI preview works

---

## Phase 10: Additional Failure Types

**Purpose**: Handle edge cases and additional failure types

- [X] T070 [US8] Add fuzz test failure detection pattern in pkg/mage/ci_stream_parser.go
- [X] T071 [US8] Implement FuzzInfo extraction (corpus path, input, seed indicator) in pkg/mage/ci_stream_parser.go
- [X] T072 Add fatal crash detection (exit code + incomplete JSON) in pkg/mage/ci_runner.go
- [X] T073 Implement crash dump parsing from stderr in pkg/mage/ci_runner.go
- [X] T074 Add timeout failure detection in pkg/mage/ci_stream_parser.go
- [X] T075 [P] [US8] Create unit tests for fuzz failure detection in pkg/mage/ci_stream_parser_test.go
- [X] T076 [P] Create unit tests for fatal crash detection in pkg/mage/ci_runner_test.go

**Checkpoint**: Phase 10 complete - All additional failure types handled

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements and validation

- [X] T077 [P] Add error handling for disk full conditions (skip JSONL, log warning) in pkg/mage/ci_reporter_json.go
- [X] T078 [P] Add fallback output path handling in pkg/mage/ci_reporter_json.go
- [X] T079 Add graceful degradation when CI mode processing fails in pkg/mage/ci_runner.go
- [X] T080 Validate all regex patterns against edge cases in pkg/mage/ci_stream_parser_test.go
- [X] T081 [P] Add benchmarks for 10,000 test parsing performance in pkg/mage/ci_stream_parser_test.go
- [X] T082 Run quickstart.md validation scenarios (validated via comprehensive test suite)
- [X] T083 Verify backwards compatibility (existing workflows unchanged - tests pass)

**Checkpoint**: Phase 11 complete - All polish items addressed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - US1 + US2: Can run in parallel (both P1 priority)
  - US3 + US4 + US5: Can run in parallel after US1/US2 (all P2 priority)
  - US6 + US7: Can run in parallel (both P3 priority)
- **Additional Failure Types (Phase 10)**: Depends on US5 (uses same parser infrastructure)
- **Polish (Phase 11)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No story dependencies
- **User Story 2 (P1)**: Can start after Foundational - No story dependencies
- **User Story 3 (P2)**: Can start after Foundational - No story dependencies
- **User Story 4 (P2)**: Can start after Foundational - Uses StreamParser from Phase 2
- **User Story 5 (P2)**: Can start after Foundational - Extends StreamParser patterns
- **User Story 6 (P3)**: Can start after Foundational - Extends StreamParser strategies
- **User Story 7 (P3)**: Can start after US1 - Uses CIRunner from US1

### Within Each User Story

- Define interfaces before implementation
- Core implementation before wiring into test.go
- Unit tests alongside each implementation file
- Integration tests after components wired together

### Parallel Opportunities

- T002, T003, T004 can run in parallel (independent struct definitions)
- T011, T013, T020 can run in parallel (different files)
- T031, T038 can run in parallel (different reporter implementations)
- T055, T056, T057 can run in parallel (independent test cases)
- T075, T076 can run in parallel (different test files)
- T077, T078, T081 can run in parallel (independent concerns)

---

## Parallel Example: Phase 2 (Foundational)

```bash
# After T008-T010 (CIDetector), these can run in parallel:
Task: "Create unit tests for CIDetector in pkg/mage/ci_detector_test.go" (T011)
Task: "Define RingBuffer for context line capture in pkg/mage/ci_stream_parser.go" (T013)

# After T015-T018, these can run in parallel:
Task: "Create unit tests for StreamParser in pkg/mage/ci_stream_parser_test.go" (T019)
Task: "Add fuzz tests for StreamParser robustness in pkg/mage/ci_stream_parser_test.go" (T020)
```

## Parallel Example: User Story 2 & 3

```bash
# After Foundational phase, these can start in parallel:
# US2 - GitHub Reporter
Task: "Implement GitHubReporter interface in pkg/mage/ci_reporter_github.go" (T031)

# US3 - JSON Reporter
Task: "Implement JSONReporter interface in pkg/mage/ci_reporter_json.go" (T038)
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T020)
3. Complete Phase 3: User Story 1 - CI Auto-Detection (T021-T030)
4. Complete Phase 4: User Story 2 - GitHub Annotations (T031-T037)
5. **STOP and VALIDATE**: Test in GitHub Actions, verify annotations appear
6. Deploy if ready - this is the MVP

### Incremental Delivery

1. Setup + Foundational → Core infrastructure ready
2. Add User Story 1 + 2 → GitHub annotations working (MVP!)
3. Add User Story 3 → JSONL output for automation
4. Add User Story 4 → Code context in failures
5. Add User Story 5 → Panic/race detection
6. Add User Story 6 → Large suite handling
7. Add User Story 7 → Local preview
8. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:
1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (CI Runner)
   - Developer B: User Story 2 (GitHub Reporter)
   - Developer C: User Story 3 (JSON Reporter)
3. After MVP validated:
   - Continue with User Stories 4-7 in priority order

---

## Notes

- All new files use `ci_` prefix to group CI mode functionality
- Each `.go` file gets corresponding `_test.go` file
- Existing TestFailure struct extended (not replaced) for backwards compatibility
- Standard library only - no new dependencies added
- Memory bounded regardless of test suite size
- Streaming-friendly design throughout
