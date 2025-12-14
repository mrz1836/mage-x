# Quickstart: CI Test Output Mode

**Feature Branch**: `001-ci-test-output`
**Date**: 2025-12-12

## Overview

CI Test Output Mode adds automatic detection of CI environments and produces structured test output with precise file:line locations. This eliminates complex bash/jq parsing in CI workflows while maintaining 100% backwards compatibility.

## Basic Usage

### Automatic CI Detection (Default)

No configuration needed! When running in CI, structured output is automatically enabled:

```bash
# In GitHub Actions (or any CI with CI=true)
magex test:unit  # Automatically produces GitHub annotations + JSONL

# Locally (no CI variables)
magex test:unit  # Standard terminal output (unchanged)
```

### Explicit CI Mode

Force CI mode on or off regardless of environment:

```bash
# Enable CI mode locally (preview CI output)
magex test:unit ci

# Disable CI mode in CI (use standard output)
magex test:unit ci=false
```

### All Test Commands Support CI Mode

```bash
magex test:unit ci      # Unit tests with CI output
magex test:race ci      # Race detection with CI output
magex test:cover ci     # Coverage with CI output
magex test:fuzz ci      # Fuzz tests with CI output
```

## Configuration

### Via .mage.yaml

```yaml
# .mage.yaml
test:
  ci_mode:
    enabled: auto          # auto (default), on, or off
    format: github         # auto, github, or json
    context_lines: 20      # Lines of code context around failures
    max_memory_mb: 100     # Memory limit for large test suites
    dedup: true            # Deduplicate identical failures
    output_path: ".mage-x/ci-results.jsonl"
```

### Via Environment Variables

```bash
export MAGE_X_CI_MODE=auto
export MAGE_X_CI_FORMAT=github
export MAGE_X_CI_CONTEXT=50
export MAGE_X_CI_MAX_MEMORY=200MB
```

## Output Formats

### GitHub Actions Annotations

When running in GitHub Actions, failures appear in the PR annotations sidebar:

```
::error file=pkg/mage/build_test.go,line=42,title=TestBuildDefault::Expected 'linux', got 'darwin'
```

### JSON Lines Output

Structured output written to `.mage-x/ci-results.jsonl`:

```jsonl
{"type":"start","timestamp":"2025-12-12T10:30:00Z"}
{"type":"failure","failure":{"package":"pkg/mage","test":"TestBuild","file":"build_test.go","line":42}}
{"type":"summary","summary":{"status":"failed","total":100,"passed":99,"failed":1}}
```

### GitHub Step Summary

Markdown table written to `$GITHUB_STEP_SUMMARY`:

| Status | Count |
|--------|-------|
| ✅ Passed | 99 |
| ❌ Failed | 1 |

## GitHub Actions Workflow

### Before (Complex Parsing)

```yaml
- name: Run tests
  run: magex test:cover -json | tee test-output.log

- name: Extract failures
  if: failure()
  run: |
    # 100+ lines of complex bash/jq parsing
    grep '^{' test-output.log | \
      jq -r 'select(.Action=="fail")' | \
      # ... extensive processing ...
```

### After (Native CI Mode)

```yaml
- name: Run tests
  run: magex test:cover
  # CI mode auto-enabled:
  # - GitHub annotations with file:line
  # - .mage-x/ci-results.jsonl created
  # - $GITHUB_STEP_SUMMARY populated
```

## Troubleshooting

### CI Mode Not Activating

Check environment variables:
```bash
echo $CI $GITHUB_ACTIONS $GITLAB_CI
```

Force CI mode explicitly:
```bash
magex test:unit ci
```

### Memory Issues with Large Test Suites

Increase memory limit:
```yaml
test:
  ci_mode:
    max_memory_mb: 200
```

Or reduce context lines:
```yaml
test:
  ci_mode:
    context_lines: 5
```

### Output File Not Created

Check permissions and path:
```bash
mkdir -p .mage-x
magex test:unit ci
cat .mage-x/ci-results.jsonl
```

## API Reference

### CIMode Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool/string | auto | Activation mode |
| `format` | string | auto | Output format |
| `context_lines` | int | 20 | Code context lines |
| `max_memory_mb` | int | 100 | Memory limit |
| `dedup` | bool | true | Deduplicate failures |
| `output_path` | string | .mage-x/ci-results.jsonl | Output file |

### TestFailure Fields

| Field | Type | Description |
|-------|------|-------------|
| `package` | string | Go package path |
| `test` | string | Test name (with subtest path) |
| `type` | string | test, build, panic, race, fuzz, timeout |
| `file` | string | Source file (relative) |
| `line` | int | Line number |
| `error` | string | Error message |
| `context` | []string | Surrounding code lines |
| `signature` | string | Deduplication key |

## Next Steps

1. **Implement Session 1**: CI Detection & Config (`ci_detector.go`)
2. **Implement Session 2**: Extended TestFailure
3. **Follow implementation plan**: See `plan.md` for full session breakdown
