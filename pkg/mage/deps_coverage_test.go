package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// DepsCoverageTestSuite provides comprehensive coverage for Deps methods
type DepsCoverageTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	deps Deps
}

func TestDepsCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(DepsCoverageTestSuite))
}

func (ts *DepsCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.deps = Deps{}

	// Create .mage.yaml config
	ts.env.CreateMageConfig(`
project:
  name: test
`)

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *DepsCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *DepsCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// TestDepsDefaultSuccess tests Deps.Default
func (ts *DepsCoverageTestSuite) TestDepsDefaultSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Default()
	})

	ts.Require().NoError(err)
}

// TestDepsDownloadSuccess tests Deps.Download
func (ts *DepsCoverageTestSuite) TestDepsDownloadSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Download()
	})

	ts.Require().NoError(err)
}

// TestDepsTidySuccess tests Deps.Tidy
func (ts *DepsCoverageTestSuite) TestDepsTidySuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Tidy()
	})

	ts.Require().NoError(err)
}

// TestDepsUpdateSuccess tests Deps.Update
func (ts *DepsCoverageTestSuite) TestDepsUpdateSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Update()
	})

	ts.Require().NoError(err)
}

// TestDepsUpdateWithArgsSuccess tests Deps.UpdateWithArgs
func (ts *DepsCoverageTestSuite) TestDepsUpdateWithArgsSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.UpdateWithArgs()
	})

	ts.Require().NoError(err)
}

// TestDepsUpdateWithArgsDryRun tests Deps.UpdateWithArgs with dry-run
func (ts *DepsCoverageTestSuite) TestDepsUpdateWithArgsDryRun() {
	err := ts.withMockRunner(func() error {
		return ts.deps.UpdateWithArgs("dry-run=true")
	})

	ts.Require().NoError(err)
}

// TestDepsUpdateWithArgsVerbose tests Deps.UpdateWithArgs with verbose
func (ts *DepsCoverageTestSuite) TestDepsUpdateWithArgsVerbose() {
	err := ts.withMockRunner(func() error {
		return ts.deps.UpdateWithArgs("verbose=true")
	})

	ts.Require().NoError(err)
}

// TestDepsCleanSuccess tests Deps.Clean
func (ts *DepsCoverageTestSuite) TestDepsCleanSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Clean()
	})

	ts.Require().NoError(err)
}

// TestDepsGraphSuccess tests Deps.Graph
func (ts *DepsCoverageTestSuite) TestDepsGraphSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Graph()
	})

	ts.Require().NoError(err)
}

// TestDepsVerifySuccess tests Deps.Verify
func (ts *DepsCoverageTestSuite) TestDepsVerifySuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Verify()
	})

	ts.Require().NoError(err)
}

// TestDepsListSuccess tests Deps.List
func (ts *DepsCoverageTestSuite) TestDepsListSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.List()
	})

	ts.Require().NoError(err)
}

// TestDepsVendorSuccess tests Deps.Vendor
func (ts *DepsCoverageTestSuite) TestDepsVendorSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Vendor()
	})

	ts.Require().NoError(err)
}

// TestDepsCheckSuccess tests Deps.Check
func (ts *DepsCoverageTestSuite) TestDepsCheckSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Check()
	})

	ts.Require().NoError(err)
}

// TestDepsLicensesSuccess tests Deps.Licenses
func (ts *DepsCoverageTestSuite) TestDepsLicensesSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.deps.Licenses()
	})

	ts.Require().NoError(err)
}

// ================== Standalone Tests for Helper Functions ==================

func TestIsMajorVersionUpdate(t *testing.T) {
	tests := []struct {
		name       string
		oldVersion string
		newVersion string
		expected   bool
	}{
		{
			name:       "major version update v1 to v2",
			oldVersion: "v1.0.0",
			newVersion: "v2.0.0",
			expected:   true,
		},
		{
			name:       "minor version update v1.0 to v1.1",
			oldVersion: "v1.0.0",
			newVersion: "v1.1.0",
			expected:   false,
		},
		{
			name:       "patch version update",
			oldVersion: "v1.0.0",
			newVersion: "v1.0.1",
			expected:   false,
		},
		{
			name:       "same major different minor",
			oldVersion: "v2.1.0",
			newVersion: "v2.5.0",
			expected:   false,
		},
		{
			name:       "major jump v1 to v3",
			oldVersion: "v1.2.3",
			newVersion: "v3.0.0",
			expected:   true,
		},
		{
			name:       "without v prefix",
			oldVersion: "1.0.0",
			newVersion: "2.0.0",
			expected:   true,
		},
		{
			name:       "pre-release versions",
			oldVersion: "v1.0.0-alpha",
			newVersion: "v1.0.0",
			expected:   false,
		},
		{
			name:       "empty old version treated as 0",
			oldVersion: "",
			newVersion: "v1.0.0",
			expected:   true, // empty is treated as major 0, so 0->1 is a major update
		},
		{
			name:       "empty new version",
			oldVersion: "v1.0.0",
			newVersion: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMajorVersionUpdate(tt.oldVersion, tt.newVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractMajorVersion(t *testing.T) {
	// Note: extractMajorVersion extracts digits from a version PART (like "1" or "2"),
	// not a full version string. It's called on parts after splitting by '.'
	tests := []struct {
		name     string
		version  string
		expected int
	}{
		{
			name:     "simple digit 1",
			version:  "1",
			expected: 1,
		},
		{
			name:     "simple digit 2",
			version:  "2",
			expected: 2,
		},
		{
			name:     "double digits 10",
			version:  "10",
			expected: 10,
		},
		{
			name:     "with suffix",
			version:  "1-alpha",
			expected: 1,
		},
		{
			name:     "empty string",
			version:  "",
			expected: 0,
		},
		{
			name:     "non-digit prefix returns 0",
			version:  "v1",
			expected: 0,
		},
		{
			name:     "letters only",
			version:  "abc",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMajorVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{
			name:     "equal versions",
			v1:       "v1.0.0",
			v2:       "v1.0.0",
			expected: 0,
		},
		{
			name:     "v1 less than v2 major",
			v1:       "v1.0.0",
			v2:       "v2.0.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 major",
			v1:       "v2.0.0",
			v2:       "v1.0.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2 minor",
			v1:       "v1.1.0",
			v2:       "v1.2.0",
			expected: -1,
		},
		{
			name:     "v1 greater than v2 minor",
			v1:       "v1.3.0",
			v2:       "v1.2.0",
			expected: 1,
		},
		{
			name:     "v1 less than v2 patch",
			v1:       "v1.0.1",
			v2:       "v1.0.2",
			expected: -1,
		},
		{
			name:     "without v prefix",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantPatch int
	}{
		{
			name:      "standard version",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
		},
		{
			name:      "major minor only",
			version:   "1.2",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 0,
		},
		{
			name:      "major only",
			version:   "3",
			wantMajor: 3,
			wantMinor: 0,
			wantPatch: 0,
		},
		{
			name:      "pre-release",
			version:   "1.0.0-alpha",
			wantMajor: 1,
			wantMinor: 0,
			wantPatch: 0,
		},
		{
			name:      "empty string",
			version:   "",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := parseVersion(tt.version)
			assert.Equal(t, tt.wantMajor, parts.numbers[0], "major")
			assert.Equal(t, tt.wantMinor, parts.numbers[1], "minor")
			assert.Equal(t, tt.wantPatch, parts.numbers[2], "patch")
		})
	}
}

func TestIsVersionNewer(t *testing.T) {
	// isVersionNewer(v1, v2) returns true if v1 is newer than v2
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{
			name:     "v2 is newer than v1",
			v1:       "v2.0.0",
			v2:       "v1.0.0",
			expected: true,
		},
		{
			name:     "v1.2 is newer than v1.1",
			v1:       "v1.2.0",
			v2:       "v1.1.0",
			expected: true,
		},
		{
			name:     "v1.0.2 is newer than v1.0.1",
			v1:       "v1.0.2",
			v2:       "v1.0.1",
			expected: true,
		},
		{
			name:     "same version is not newer",
			v1:       "v1.0.0",
			v2:       "v1.0.0",
			expected: false,
		},
		{
			name:     "v1 is not newer than v2",
			v1:       "v1.0.0",
			v2:       "v2.0.0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isVersionNewer(tt.v1, tt.v2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDepsWhyRequiresName(t *testing.T) {
	deps := Deps{}

	// Should fail with empty dependency name
	err := deps.Why("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dependency name required")
}

func TestDepsInitRequiresName(t *testing.T) {
	deps := Deps{}

	// Should fail with empty module name
	err := deps.Init("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module name required")
}

func TestDepsInitWithExistingGoMod(t *testing.T) {
	deps := Deps{}
	tmpDir := t.TempDir()

	// Create go.mod file
	goModPath := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module test\n"), 0o600)
	require.NoError(t, err)

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // cleanup in defer

	// Should fail because go.mod exists
	err = deps.Init("test/module")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDepsOutdatedSuccess(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")

	// Set up general mocks for all commands
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()

	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			deps := Deps{}
			return deps.Outdated()
		},
	)

	// Should succeed (may show no outdated deps)
	require.NoError(t, err)
}

func TestDepsVulnCheckWithoutGoMod(t *testing.T) {
	deps := Deps{}
	tmpDir := t.TempDir()

	// Change to temp dir (no go.mod)
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // cleanup in defer

	// Should fail because no go.mod (error message varies by environment)
	err = deps.VulnCheck()
	require.Error(t, err)
	// Just verify an error occurred - exact message depends on govulncheck availability
}

func TestDepsAuditWithoutGoMod(t *testing.T) {
	deps := Deps{}
	tmpDir := t.TempDir()

	// Change to temp dir (no go.mod)
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // cleanup in defer

	// Should fail because no go.mod (error message varies by environment)
	err = deps.Audit()
	require.Error(t, err)
	// Just verify an error occurred - exact message depends on govulncheck availability
}
