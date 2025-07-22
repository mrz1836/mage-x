# File Operations Migration to fileops Package

This document lists all direct file operations in the pkg/mage directory that should be migrated to use the fileops package.

## Summary

Total files with direct file operations: **41 files**
Total operations to migrate: **~116 operations**

## Migration Categories

### 1. Simple File Read/Write Operations

These are straightforward migrations from `os.ReadFile/WriteFile` to `fileops.ReadFile/WriteFile`:

#### Files with os.WriteFile operations:
- **analytics.go** (5 occurrences) - Lines: 137, 293, 423, 539, 665
- **audit.go** (3 occurrences) - Lines: 189, 253, 300, 355
- **audit_v2.go** (1 occurrence) - Line: 1143
- **cli.go** (3 occurrences) - Lines: 523, 647, 763, 1001
- **config.go** (2 occurrences) - Lines: 332, 343
- **configure.go** (2 occurrences) - Lines: 171, 261
- **docs.go** (3 occurrences) - Lines: 50, 124, 280
- **enterprise.go** (6 occurrences) - Lines: 443, 745, 782, 804, 829, 844
- **enterprise_config.go** (6 occurrences) - Lines: 188, 1013, 1073, 1108, 1292, 1649
- **help.go** (3 occurrences) - Lines: 653, 723, 768
- **init.go** (7 occurrences) - Lines: 425, 443, 564, 593, 631, 789, 822, 865
- **integrations.go** (3 occurrences) - Lines: 213, 929, 1168
- **mod.go** (1 occurrence) - Line: 167
- **recipes.go** (6 occurrences) - Lines: 661, 707, 744, 780, 821, 854
- **security.go** (3 occurrences) - Lines: 395, 868, 963
- **update.go** (3 occurrences) - Lines: 140, 466
- **workflow.go** (3 occurrences) - Lines: 606, 664
- **yaml.go** (1 occurrence) - Line: 554
- **testutil/helpers.go** (4 occurrences) - Lines: 89, 97, 128, 144

#### Files with os.ReadFile operations:
- **audit.go** (2 occurrences) - Lines: 230, 277
- **cli.go** (2 occurrences) - Lines: 499, 923, 1108
- **common.go** (1 occurrence) - Line: 25
- **config.go** (2 occurrences) - Lines: 134, 143
- **configure.go** (1 occurrence) - Line: 195
- **docs.go** (2 occurrences) - Lines: 34, 260
- **enterprise.go** (3 occurrences) - Lines: 464, 724, 862
- **enterprise_config.go** (2 occurrences) - Lines: 1023, 1653
- **generate.go** (2 occurrences) - Lines: 298, 342
- **generate_v2.go** (2 occurrences) - Lines: 821, 1074
- **integrations.go** (2 occurrences) - Lines: 234, 901
- **security.go** (1 occurrence) - Line: 441
- **security_v2.go** (1 occurrence) - Line: 2394
- **update.go** (2 occurrences) - Lines: 166, 446
- **workflow.go** (3 occurrences) - Lines: 360, 779, 811, 993
- **yaml.go** (1 occurrence) - Line: 527
- **testutil/helpers.go** (1 occurrence) - Line: 158

#### Files with os.Create operations:
- **bench.go** (1 occurrence) - Line: 118
- **init.go** (1 occurrence) - Line: 500
- **update.go** (1 occurrence) - Line: 406

#### Files with os.Open operations:
- **metrics.go** (2 occurrences) - Lines: 272, 318

### 2. JSON Operations with File I/O

These files perform JSON marshal/unmarshal operations combined with file I/O and should use `fileops.WriteJSON/ReadJSON`:

- **analytics.go**:
  - Lines 135-137: `json.MarshalIndent` → `os.WriteFile`
  - Lines 291-293: `json.MarshalIndent` → `os.WriteFile`
  - Lines 421-423: `json.MarshalIndent` → `os.WriteFile`
  - Lines 537-539: `json.MarshalIndent` → `os.WriteFile`
  - Lines 663-665: `json.MarshalIndent` → `os.WriteFile`

- **audit.go**:
  - Lines 187-189: `json.MarshalIndent` → `os.WriteFile`
  - Lines 353-355: `json.MarshalIndent` → `os.WriteFile`

- **enterprise_config.go**:
  - Lines 186-188: `json.MarshalIndent` → `os.WriteFile` (when format is JSON)
  - Lines 1011-1013: `json.MarshalIndent` → `os.WriteFile`
  - Lines 1647-1649: `json.MarshalIndent` → `os.WriteFile`

- **integrations.go**:
  - Lines 211-213: `json.MarshalIndent` → `os.WriteFile`
  - Lines 1166-1168: `json.MarshalIndent` → `os.WriteFile`

- **security.go**:
  - Lines 393-395: `json.MarshalIndent` → `os.WriteFile`
  - Lines 866-868: `json.MarshalIndent` → `os.WriteFile`
  - Lines 961-963: `json.MarshalIndent` → `os.WriteFile`

### 3. YAML Operations with File I/O

These files perform YAML marshal/unmarshal operations combined with file I/O and should use `fileops.WriteYAML/ReadYAML`:

- **config.go**:
  - Lines 134-143: `os.ReadFile` → `yaml.Unmarshal`
  - Lines 330-332: `yaml.Marshal` → `os.WriteFile`
  - Lines 341-343: `yaml.Marshal` → `os.WriteFile`

- **configure.go**:
  - Lines 195-200: `os.ReadFile` → `yaml.Unmarshal`
  - Line 257: `yaml.Marshal` → `os.WriteFile`

- **configure_v2.go**:
  - Already using fileops but has some operations that could be optimized
  - Lines 252-254: `yaml.Unmarshal` (could use `fileops.ReadYAML`)
  - Lines 314-316: `yaml.Marshal` (could use `fileops.WriteYAML`)

- **enterprise_config.go**:
  - Lines 186-188: `yaml.Marshal` → `os.WriteFile` (when format is YAML)
  - Lines 1011-1013: `yaml.Marshal` → `os.WriteFile`
  - Lines 1023-1027: `os.ReadFile` → `yaml.Unmarshal`
  - Lines 1071-1073: `yaml.Marshal` → `os.WriteFile`
  - Lines 1653-1657: `os.ReadFile` → `yaml.Unmarshal`

- **workflow.go**:
  - Lines 360-364: `os.ReadFile` → `yaml.Unmarshal`
  - Lines 604-606: `yaml.Marshal` → `os.WriteFile`
  - Lines 662-664: `yaml.Marshal` → `os.WriteFile`

- **yaml.go**:
  - Lines 527-531: `os.ReadFile` → `yaml.Unmarshal`
  - Lines 552-554: `yaml.Marshal` → `os.WriteFile`

### 4. Special Cases

#### File Creation/Opening for Streaming:
- **bench.go** (Line 118): `os.Create` - Used for benchmark output, may need streaming support
- **init.go** (Line 500): `os.Create` - Used for README.md creation with writer
- **update.go** (Line 406): `os.Create` - Used for downloading files
- **metrics.go** (Lines 272, 318): `os.Open` - Used for reading metrics files

#### Already Using fileops:
These v2 implementations already use fileops package:
- analytics_v2.go
- build_v2.go
- cli_v2.go
- configure_v2.go
- deps_v2.go
- docs_v2.go
- enterprise_v2.go
- format_v2.go
- generate_v2.go
- git_v2.go
- integrations_v2.go
- lint_v2.go
- release_v2.go
- test_v2.go
- tools_v2.go
- workflow_v2.go

## Migration Priority

### High Priority (Core functionality):
1. **config.go** - Configuration loading/saving
2. **configure.go** - Configuration management
3. **build.go** - Build operations
4. **test.go** - Test operations
5. **init.go** - Project initialization

### Medium Priority (Feature functionality):
1. **analytics.go** - Analytics and reporting
2. **audit.go** - Audit functionality
3. **security.go** - Security operations
4. **workflow.go** - Workflow management
5. **enterprise.go** & **enterprise_config.go** - Enterprise features

### Low Priority (Supporting functionality):
1. **docs.go** - Documentation generation
2. **help.go** - Help system
3. **recipes.go** - Recipe management
4. **integrations.go** - Integration support
5. **testutil/helpers.go** - Test utilities

## Migration Benefits

1. **Consistency**: All file operations go through a single interface
2. **Testability**: Easy to mock file operations in tests
3. **Error Handling**: Centralized error handling and recovery
4. **Atomic Operations**: Support for atomic writes and backups
5. **Format-specific Operations**: Dedicated JSON/YAML helpers reduce boilerplate

## Next Steps

1. Start with high-priority files that aren't already using v2 implementations
2. For each file:
   - Replace `os.ReadFile` with `fileops.ReadFile`
   - Replace `os.WriteFile` with `fileops.WriteFile`
   - Replace JSON marshal+write with `fileops.WriteJSON`
   - Replace YAML marshal+write with `fileops.WriteYAML`
   - Replace read+unmarshal with `fileops.ReadJSON/ReadYAML`
3. Update tests to use mock file operators
4. Consider adding streaming support to fileops for large file operations