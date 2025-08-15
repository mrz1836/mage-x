---
allowed-tools: Task(mage-x-releaser), Task(mage-x-git), Task(mage-x-gh), Task(mage-x-security), Task(mage-x-docs), Task(mage-x-linter), Task(mage-x-test-writer), Bash(mage release:*), Bash(mage version:*), Bash(git:*), Bash(gh release:*), Read, Write, MultiEdit, Grep, Glob
argument-hint: [version|channel|dry-run]
description: Comprehensive release preparation with parallel validation
model: claude-sonnet-4-20250514
---

## Context
- Current version: !`mage version:show 2>/dev/null || git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"`
- Pending changes: !`git status --porcelain | wc -l | xargs echo "Modified files:"`
- Recent commits: !`git log --oneline -10`
- Release channels: !`grep -E "channels:|stable|beta|edge" .mage.yaml 2>/dev/null | head -5`

## Your Task

Prepare a comprehensive release through coordinated agent execution. Target: ${ARGUMENTS:-stable}

### Phase 1: Pre-Release Validation (Parallel)

Execute simultaneously for comprehensive validation:

1. **mage-x-security** - Security Audit:
   - Final vulnerability scan
   - Dependency security check
   - Secret/credential validation
   - Security compliance verification

2. **mage-x-linter** - Code Quality:
   - Full linting pass
   - Documentation completeness
   - Code standard compliance
   - Breaking change detection

3. **Test Validation** (if needed):
   - Run full test suite
   - Verify all tests pass
   - Check coverage thresholds
   - Validate benchmarks

4. **mage-x-docs** - Documentation:
   - API documentation updates
   - README version updates
   - Migration guide creation
   - Changelog generation

### Phase 2: Release Preparation

1. **mage-x-git** - Version Control:
   - Create version tag
   - Validate commit history
   - Ensure clean working tree
   - Branch protection checks

2. **mage-x-releaser** - Release Orchestration:
   - Version bump execution
   - Changelog generation
   - Asset preparation
   - Multi-channel configuration

3. **Build Validation** (Parallel):
   - Cross-platform builds
   - Binary generation
   - Docker image creation
   - Archive preparation

### Phase 3: Release Execution

1. **mage-x-gh** - GitHub Release:
   - Create GitHub release
   - Upload release assets
   - Set release notes
   - Configure pre-release flags

2. **Post-Release Tasks**:
   - Documentation deployment
   - Announcement preparation
   - Downstream notifications
   - Metrics collection

## Release Channels

- **stable**: Production-ready releases
- **beta**: Pre-release testing versions
- **edge**: Bleeding-edge development builds
- **dry-run**: Simulate release without publishing

## Expected Output

### Pre-Release Checklist
```
✅ Security scan passed
✅ All tests passing
✅ Documentation updated
✅ Changelog generated
✅ Version bumped
✅ Assets built
```

### Release Summary
- **Version**: [new version]
- **Channel**: [stable/beta/edge]
- **Changes**: [count] commits since last release
- **Breaking Changes**: [yes/no with details]

### Generated Assets
```
📦 Binaries:
  - linux/amd64: mage-x-linux-amd64
  - darwin/amd64: mage-x-darwin-amd64
  - darwin/arm64: mage-x-darwin-arm64
  - windows/amd64: mage-x-windows-amd64.exe

🐳 Docker Images:
  - mage-x:latest
  - mage-x:[version]
  - mage-x:[channel]
```

### Release Notes
[Automated changelog with categorized changes]

### Next Steps
1. Review release draft
2. Publish when ready
3. Monitor release metrics
4. Address any issues
