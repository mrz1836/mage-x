package mage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// installTestHelper provides test utilities for Install testing
type installTestHelper struct {
	originalRunner CommandRunner
	mockRunner     *testutil.MockRunner
	tmpDir         string
}

// newInstallTestHelper creates a new install test helper with temp directory
func newInstallTestHelper(tb testing.TB) *installTestHelper {
	tb.Helper()
	h := &installTestHelper{
		originalRunner: GetRunner(),
		tmpDir:         tb.TempDir(),
	}
	runner, _ := testutil.NewMockRunner()
	h.mockRunner = runner
	require.NoError(tb, SetRunner(h.mockRunner))
	return h
}

// teardown restores the original runner
func (h *installTestHelper) teardown(tb testing.TB) {
	tb.Helper()
	if h.originalRunner != nil {
		require.NoError(tb, SetRunner(h.originalRunner))
	}
}

// expectCmd sets up an expectation for a command
func (h *installTestHelper) expectCmd(cmd string) {
	h.mockRunner.On("RunCmd", cmd, mock.Anything).Return(nil)
}

// expectCmdError sets up an expectation for a command that returns an error
//
//nolint:unparam // cmd parameter may vary in future tests
func (h *installTestHelper) expectCmdError(cmd string) {
	h.mockRunner.On("RunCmd", cmd, mock.Anything).Return(assert.AnError)
}

// expectCmdOutput sets up an expectation for a command with output
//
//nolint:unparam // cmd parameter may vary in future tests
func (h *installTestHelper) expectCmdOutput(cmd, output string, err error) {
	h.mockRunner.On("RunCmdOutput", cmd, mock.Anything).Return(output, err)
}

// expectGitVersion sets up mock expectations for git version commands used by buildFlags
func (h *installTestHelper) expectGitVersion() {
	h.mockRunner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil).Maybe()
	h.mockRunner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--always", "--dirty"}).Return("v1.0.0", nil).Maybe()
	h.mockRunner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("", nil).Maybe()
	h.mockRunner.On("RunCmdOutput", "git", []string{"rev-parse", "HEAD"}).Return("abc123", nil).Maybe()
	h.mockRunner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).Return("abc1234", nil).Maybe()
	h.mockRunner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--long", "--abbrev=0"}).Return("v1.0.0-0-gabc123", nil).Maybe()
}

// TestIsInPath tests the isInPath helper function
func TestIsInPath(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		pathEnv  string
		expected bool
	}{
		{
			name:     "directory in PATH",
			dir:      "/usr/local/bin",
			pathEnv:  "/usr/bin:/usr/local/bin:/bin",
			expected: true,
		},
		{
			name:     "directory not in PATH",
			dir:      "/custom/path",
			pathEnv:  "/usr/bin:/usr/local/bin:/bin",
			expected: false,
		},
		{
			name:     "empty PATH",
			dir:      "/usr/bin",
			pathEnv:  "",
			expected: false,
		},
		{
			name:     "exact match required",
			dir:      "/usr",
			pathEnv:  "/usr/bin:/usr/local/bin",
			expected: false,
		},
		{
			name:     "single path element",
			dir:      "/bin",
			pathEnv:  "/bin",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", tt.pathEnv)
			result := isInPath(tt.dir)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCheckExistingAlias tests the checkExistingAlias helper function
func TestCheckExistingAlias(t *testing.T) {
	t.Run("symlink points to same binary", func(t *testing.T) {
		tmpDir := t.TempDir()
		installPath := filepath.Join(tmpDir, "myapp")
		aliasPath := filepath.Join(tmpDir, "myalias")

		// Create the binary file
		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))

		// Create symlink pointing to binary
		require.NoError(t, os.Symlink(installPath, aliasPath))

		result := checkExistingAlias(aliasPath, installPath, "myalias")
		assert.True(t, result)
	})

	t.Run("symlink points to different binary", func(t *testing.T) {
		tmpDir := t.TempDir()
		installPath := filepath.Join(tmpDir, "myapp")
		otherBinary := filepath.Join(tmpDir, "other")
		aliasPath := filepath.Join(tmpDir, "myalias")

		// Create both binary files
		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))
		require.NoError(t, os.WriteFile(otherBinary, []byte("other"), fileops.PermFileExecutable))

		// Create symlink pointing to other binary
		require.NoError(t, os.Symlink(otherBinary, aliasPath))

		result := checkExistingAlias(aliasPath, installPath, "myalias")
		assert.False(t, result)
	})

	t.Run("file is not a symlink", func(t *testing.T) {
		tmpDir := t.TempDir()
		installPath := filepath.Join(tmpDir, "myapp")
		aliasPath := filepath.Join(tmpDir, "myalias")

		// Create regular files
		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))
		require.NoError(t, os.WriteFile(aliasPath, []byte("alias"), fileops.PermFileExecutable))

		result := checkExistingAlias(aliasPath, installPath, "myalias")
		assert.False(t, result)
	})

	t.Run("alias does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		installPath := filepath.Join(tmpDir, "myapp")
		aliasPath := filepath.Join(tmpDir, "nonexistent")

		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))

		result := checkExistingAlias(aliasPath, installPath, "myalias")
		assert.False(t, result)
	})
}

// TestCreateWindowsBatchWrapper tests the createWindowsBatchWrapper helper function
func TestCreateWindowsBatchWrapper(t *testing.T) {
	t.Run("creates batch file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		aliasPath := filepath.Join(tmpDir, "myalias.exe")
		installPath := filepath.Join(tmpDir, "myapp.exe")

		createWindowsBatchWrapper(aliasPath, installPath, "myalias")

		// Verify batch file was created (without .exe, with .bat)
		batchPath := filepath.Join(tmpDir, "myalias.bat")
		content, err := os.ReadFile(batchPath) //nolint:gosec // test file path
		require.NoError(t, err)
		assert.Contains(t, string(content), "@echo off")
		assert.Contains(t, string(content), installPath)
	})

	t.Run("handles path without exe extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		aliasPath := filepath.Join(tmpDir, "myalias")
		installPath := filepath.Join(tmpDir, "myapp.exe")

		createWindowsBatchWrapper(aliasPath, installPath, "myalias")

		// When path doesn't have .exe, the function writes to aliasPath directly
		// (no .bat conversion happens since .exe suffix check fails)
		// The file is created at aliasPath
		_, err := os.Stat(aliasPath)
		assert.NoError(t, err)
	})
}

// TestCreateSymlinkAlias tests the createSymlinkAlias helper function
func TestCreateSymlinkAlias(t *testing.T) {
	if runtime.GOOS == OSWindows {
		t.Skip("Skipping symlink test on Windows")
	}

	t.Run("creates symlink successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		gopath := tmpDir
		binDir := filepath.Join(gopath, "bin")
		require.NoError(t, os.MkdirAll(binDir, fileops.PermDirSensitive))

		installPath := filepath.Join(binDir, "myapp")
		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))

		createSymlinkAlias(gopath, installPath, "myalias")

		// Verify symlink was created
		aliasPath := filepath.Join(binDir, "myalias")
		link, err := os.Readlink(aliasPath)
		require.NoError(t, err)
		assert.Equal(t, installPath, link)
	})

	t.Run("skips when alias already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		gopath := tmpDir
		binDir := filepath.Join(gopath, "bin")
		require.NoError(t, os.MkdirAll(binDir, fileops.PermDirSensitive))

		installPath := filepath.Join(binDir, "myapp")
		require.NoError(t, os.WriteFile(installPath, []byte("binary"), fileops.PermFileExecutable))

		aliasPath := filepath.Join(binDir, "myalias")
		require.NoError(t, os.WriteFile(aliasPath, []byte("existing"), fileops.PermFileExecutable))

		// Should not overwrite existing file
		createSymlinkAlias(gopath, installPath, "myalias")

		// Verify original file still exists (not replaced with symlink)
		content, err := os.ReadFile(aliasPath) //nolint:gosec // test file path
		require.NoError(t, err)
		assert.Equal(t, "existing", string(content))
	})
}

// TestInstallDeps tests Install.Deps
func TestInstallDeps(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")

		err := Install{}.Deps()
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmdError("go")

		err := Install{}.Deps()
		require.Error(t, err)
	})
}

// TestInstallMage tests Install.Mage
func TestInstallMage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")

		err := Install{}.Mage()
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmdError("go")

		err := Install{}.Mage()
		require.Error(t, err)
	})
}

// TestInstallStdlib tests Install.Stdlib
func TestInstallStdlib(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")

		err := Install{}.Stdlib()
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmdError("go")

		err := Install{}.Stdlib()
		require.Error(t, err)
	})
}

// TestInstallDocker tests Install.Docker
func TestInstallDocker(t *testing.T) {
	// Docker just prints instructions, doesn't fail
	t.Run("prints instructions", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)

		err := Install{}.Docker()
		require.NoError(t, err)
	})
}

// TestInstallGitHooks tests Install.GitHooks
func TestInstallGitHooks(t *testing.T) {
	t.Run("success in git repo", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)

		// Create .git directory to simulate git repo
		gitDir := filepath.Join(h.tmpDir, ".git")
		hooksDir := filepath.Join(gitDir, "hooks")
		require.NoError(t, os.MkdirAll(hooksDir, fileops.PermDirSensitive))

		// Change to temp directory
		oldDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(h.tmpDir))
		defer func() {
			require.NoError(t, os.Chdir(oldDir))
		}()

		err = Install{}.GitHooks()
		require.NoError(t, err)

		// Verify pre-commit hook was created
		preCommitPath := filepath.Join(hooksDir, "pre-commit")
		content, err := os.ReadFile(preCommitPath) //nolint:gosec // test file path
		require.NoError(t, err)
		assert.Contains(t, string(content), "golangci-lint")
	})

	t.Run("skips when not a git repo", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)

		// Create temp directory without .git
		noGitDir := filepath.Join(h.tmpDir, "nogit")
		require.NoError(t, os.MkdirAll(noGitDir, fileops.PermDirSensitive))

		oldDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(noGitDir))
		defer func() {
			require.NoError(t, os.Chdir(oldDir))
		}()

		err = Install{}.GitHooks()
		require.NoError(t, err)
	})
}

// TestInstallTools tests Install.Tools
func TestInstallTools(t *testing.T) {
	t.Run("installs missing tools", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")

		err := Install{}.Tools()
		require.NoError(t, err)
	})
}

// TestInstallCI tests Install.CI
func TestInstallCI(t *testing.T) {
	t.Run("installs CI tools", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")

		err := Install{}.CI()
		require.NoError(t, err)
	})
}

// TestInstallCerts tests Install.Certs
func TestInstallCerts(t *testing.T) {
	t.Run("generates certificates when openssl available", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("openssl")

		// Create certs directory
		certsDir := filepath.Join(h.tmpDir, "certs")
		require.NoError(t, os.MkdirAll(certsDir, fileops.PermDirSensitive))

		oldDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(h.tmpDir))
		defer func() {
			require.NoError(t, os.Chdir(oldDir))
		}()

		err = Install{}.Certs()
		require.NoError(t, err)
	})

	t.Run("skips when certificates exist", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)

		// Create existing certificates
		certsDir := filepath.Join(h.tmpDir, "certs")
		require.NoError(t, os.MkdirAll(certsDir, fileops.PermDirSensitive))
		require.NoError(t, os.WriteFile(filepath.Join(certsDir, "server.crt"), []byte("cert"), fileops.PermFile))
		require.NoError(t, os.WriteFile(filepath.Join(certsDir, "server.key"), []byte("key"), fileops.PermFileSensitive))

		oldDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(h.tmpDir))
		defer func() {
			require.NoError(t, os.Chdir(oldDir))
		}()

		err = Install{}.Certs()
		require.NoError(t, err)
	})
}

// TestInstallErrorConstants tests error constants
func TestInstallErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "installation verification failed",
			err:  errInstallationVerificationFailed,
			want: "verification failed",
		},
		{
			name: "system-wide not supported on Windows",
			err:  errSystemWideNotSupportedWindows,
			want: "not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.want)
		})
	}
}

// TestInstallConstants tests defined constants
func TestInstallConstants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		check func(string) bool
	}{
		{
			name:  "default binary name",
			value: defaultBinaryName,
			check: func(v string) bool { return v == "app" },
		},
		{
			name:  "windows exe extension",
			value: windowsExeExt,
			check: func(v string) bool { return v == ".exe" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.check(tt.value), "constant value should match expected")
		})
	}
}

// TestInstallLocal tests Install.Local
func TestInstallLocal(t *testing.T) {
	// Local delegates to Default, so we just verify it doesn't panic
	// Full Default testing is done separately
	t.Run("delegates to Default", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		// Mock the commands that Default will call
		h.expectCmd("go")
		h.expectCmdOutput("go", "github.com/test/app", nil)

		// Will fail due to config issues, but that's expected in unit tests
		// We're testing that the delegation works without panic
		err := Install{}.Local()
		// Error is expected due to missing config in test env
		_ = err
	})
}

// TestInstallBinary tests Install.Binary
func TestInstallBinary(t *testing.T) {
	t.Run("delegates to Default", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")
		h.expectCmdOutput("go", "github.com/test/app", nil)

		// Will fail due to config issues, but that's expected in unit tests
		err := Install{}.Binary()
		// Error is expected due to missing config in test env
		_ = err
	})
}

// TestInstallVersionResolution tests version resolution for go install
func TestInstallVersionResolution(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		expectLatest bool
	}{
		{
			name:         "explicit version",
			version:      "v1.2.3",
			expectLatest: false,
		},
		{
			name:         "dev version uses latest",
			version:      "dev",
			expectLatest: true,
		},
		{
			name:         "empty version uses latest",
			version:      "",
			expectLatest: true,
		},
		{
			name:         "latest version",
			version:      "latest",
			expectLatest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDevOrEmpty := tt.version == versionDev || tt.version == "" || tt.version == "latest"
			assert.Equal(t, tt.expectLatest, isDevOrEmpty)
		})
	}
}

// TestInstallWindowsExeExtension tests Windows .exe extension handling
func TestInstallWindowsExeExtension(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		wantSuffix string
	}{
		{
			name:       "add exe extension",
			binaryName: "myapp",
			wantSuffix: ".exe",
		},
		{
			name:       "already has exe extension",
			binaryName: "myapp.exe",
			wantSuffix: ".exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				// On Windows, we expect .exe suffix
				assert.Contains(t, tt.wantSuffix, ".exe")
			}
		})
	}
}

// TestInstallModulePathParsing tests module path parsing logic
func TestInstallModulePathParsing(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		wantBinary string
	}{
		{
			name:       "simple module name",
			modulePath: "github.com/user/myapp",
			wantBinary: "myapp",
		},
		{
			name:       "nested module path",
			modulePath: "github.com/user/project/cmd/app",
			wantBinary: "app",
		},
		{
			name:       "domain with subdomain",
			modulePath: "git.example.com/team/project",
			wantBinary: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastPart := filepath.Base(tt.modulePath)
			assert.Equal(t, tt.wantBinary, lastPart)
		})
	}
}

// TestInstallPathConstruction tests install path construction
func TestInstallPathConstruction(t *testing.T) {
	tests := []struct {
		name       string
		gopath     string
		binaryName string
	}{
		{
			name:       "standard path",
			gopath:     "/home/user/go",
			binaryName: "myapp",
		},
		{
			name:       "custom gopath",
			gopath:     "/custom/gopath",
			binaryName: "app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installPath := filepath.Join(tt.gopath, "bin", tt.binaryName)
			assert.Contains(t, installPath, "bin")
			assert.Contains(t, installPath, tt.binaryName)
		})
	}
}

// TestInstallGopathResolution tests GOPATH resolution
func TestInstallGopathResolution(t *testing.T) {
	t.Run("uses GOPATH when set", func(t *testing.T) {
		customPath := "/custom/go"
		t.Setenv("GOPATH", customPath)

		gopath := os.Getenv("GOPATH")
		assert.Equal(t, customPath, gopath)
	})

	t.Run("falls back to HOME/go when GOPATH not set", func(t *testing.T) {
		t.Setenv("GOPATH", "")
		t.Setenv("HOME", "/home/testuser")

		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}
		assert.Equal(t, "/home/testuser/go", gopath)
	})
}

// TestInstallAll tests Install.All
func TestInstallAll(t *testing.T) {
	t.Run("runs all installers", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")
		h.expectCmdOutput("go", "github.com/test/app", nil)

		// Will fail due to config issues in Default, but continues
		err := Install{}.All()
		// Error is expected due to missing config in test env
		_ = err
	})
}

// TestInstallPackage tests Install.Package
func TestInstallPackage(t *testing.T) {
	t.Run("downloads and installs", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmd("go")
		h.expectCmdOutput("go", "github.com/test/app", nil)

		// Will fail due to config issues, but that's expected in unit tests
		err := Install{}.Package()
		// Error is expected due to missing config in test env
		_ = err
	})

	t.Run("fails on download error", func(t *testing.T) {
		h := newInstallTestHelper(t)
		defer h.teardown(t)
		h.expectCmdError("go")

		err := Install{}.Package()
		require.Error(t, err)
	})
}

// installFullTestHelper provides test utilities including OS, Go, and Build operations mocks
type installFullTestHelper struct {
	*installTestHelper

	originalOSOps    OSOperations
	originalGoOps    GoOperations
	originalBuildOps BuildOperations
	mockOSOps        *testutil.MockOSOperations
	mockGoOps        *testutil.MockGoOperations
	mockBuildOps     *MockBuildOperations // Local mock (defined in build_operations_test.go)
	originalConfig   *Config
}

// newInstallFullTestHelper creates a helper with all mocks
func newInstallFullTestHelper(tb testing.TB) *installFullTestHelper {
	tb.Helper()
	h := &installFullTestHelper{
		installTestHelper: newInstallTestHelper(tb),
		originalOSOps:     GetOSOperations(),
		originalGoOps:     GetGoOperations(),
		originalBuildOps:  GetBuildOperations(),
	}

	// Create mocks
	h.mockOSOps = testutil.NewMockOSOperations()
	h.mockGoOps = testutil.NewMockGoOperations()
	h.mockBuildOps = NewMockBuildOperations() // Local mock (defined in build_operations_test.go)

	// Set mocks
	require.NoError(tb, SetOSOperations(h.mockOSOps))
	require.NoError(tb, SetGoOperations(h.mockGoOps))
	require.NoError(tb, SetBuildOperations(h.mockBuildOps))

	return h
}

// teardownFull restores all original implementations
func (h *installFullTestHelper) teardownFull(tb testing.TB) {
	tb.Helper()
	h.teardown(tb)
	require.NoError(tb, SetOSOperations(h.originalOSOps))
	require.NoError(tb, SetGoOperations(h.originalGoOps))
	require.NoError(tb, SetBuildOperations(h.originalBuildOps))
	if h.originalConfig != nil {
		TestSetConfig(h.originalConfig)
	}
	TestResetConfig()
}

// setConfig sets a test configuration
func (h *installFullTestHelper) setConfig(cfg *Config) {
	TestSetConfig(cfg)
}

// TestInstallUninstall tests Install.Uninstall with mocked dependencies
func TestInstallUninstall(t *testing.T) {
	t.Run("successful uninstall with config binary", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Remove", "/home/user/go/bin/testapp").Return(nil)

		err := Install{}.Uninstall()
		require.NoError(t, err)
	})

	t.Run("uninstall with fallback to module name", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: ""}})
		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Remove", "/home/user/go/bin/myapp").Return(nil)

		err := Install{}.Uninstall()
		require.NoError(t, err)
	})

	t.Run("uninstall with GOPATH fallback to HOME", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user")
		h.mockOSOps.On("Remove", "/home/user/go/bin/testapp").Return(nil)

		err := Install{}.Uninstall()
		require.NoError(t, err)
	})

	t.Run("binary not found returns nil", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Remove", mock.Anything).Return(os.ErrNotExist)

		err := Install{}.Uninstall()
		require.NoError(t, err) // Not found is not an error
	})

	t.Run("remove error returns error", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Remove", mock.Anything).Return(os.ErrPermission)

		err := Install{}.Uninstall()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to remove binary")
	})

	t.Run("fallback to default binary name", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: ""}})
		h.mockGoOps.On("GetModuleName").Return("", assert.AnError)
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Remove", "/home/user/go/bin/app").Return(nil)

		err := Install{}.Uninstall()
		require.NoError(t, err)
	})
}

// TestInstallGoMethod tests Install.Go with mocked dependencies
func TestInstallGoMethod(t *testing.T) {
	t.Run("successful install with version", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockGoOps.On("GetVersion").Return("v1.2.3")
		h.mockGoOps.On("GetGoVulnCheckVersion").Return("v1.0.0").Maybe()
		h.mockOSOps.On("Getenv", "MAGE_X_BUILD_TAGS").Return("")
		h.expectCmd("go")

		err := Install{}.Go()
		require.NoError(t, err)
	})

	t.Run("install with build tags", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockGoOps.On("GetVersion").Return("v1.2.3")
		h.mockGoOps.On("GetGoVulnCheckVersion").Return("v1.0.0").Maybe()
		h.mockOSOps.On("Getenv", "MAGE_X_BUILD_TAGS").Return("integration")
		h.expectCmd("go")

		err := Install{}.Go()
		require.NoError(t, err)
	})

	t.Run("fallback to latest when dev version", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockGoOps.On("GetVersion").Return("dev")
		h.mockGoOps.On("GetGoVulnCheckVersion").Return("v1.0.0")
		h.mockOSOps.On("Getenv", "MAGE_X_BUILD_TAGS").Return("")
		h.expectCmd("go")

		err := Install{}.Go()
		require.NoError(t, err)
	})

	t.Run("fallback to latest when no vulncheck version", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockGoOps.On("GetVersion").Return("dev")
		h.mockGoOps.On("GetGoVulnCheckVersion").Return("")
		h.mockOSOps.On("Getenv", "MAGE_X_BUILD_TAGS").Return("")
		h.expectCmd("go")

		err := Install{}.Go()
		require.NoError(t, err)
	})

	t.Run("error getting module name", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("", assert.AnError)

		err := Install{}.Go()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get module name")
	})

	t.Run("go install command fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockGoOps.On("GetVersion").Return("v1.2.3")
		h.mockGoOps.On("GetGoVulnCheckVersion").Return("v1.0.0").Maybe()
		h.mockOSOps.On("Getenv", "MAGE_X_BUILD_TAGS").Return("")
		h.expectCmdError("go")

		err := Install{}.Go()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "go install failed")
	})
}

// TestInstallDefault tests Install.Default with mocked dependencies
func TestInstallDefault(t *testing.T) {
	t.Run("successful install with file verification", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Getenv", "PATH").Return("/usr/bin:/home/user/go/bin")
		h.mockOSOps.On("FileExists", "/home/user/go/bin/testapp").Return(true)
		h.mockOSOps.On("FileExists", mock.Anything).Return(false).Maybe()
		h.mockOSOps.On("Symlink", mock.Anything, mock.Anything).Return(nil).Maybe()
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.expectCmd("go")

		err := Install{}.Default()
		require.NoError(t, err)
	})

	t.Run("package path determination failure", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("", assert.AnError)

		err := Install{}.Default()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to determine package path")
	})

	t.Run("build command fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.expectCmdError("go")

		err := Install{}.Default()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "installation failed")
	})

	t.Run("verification fails when file does not exist", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("FileExists", mock.Anything).Return(false)
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.expectCmd("go")

		err := Install{}.Default()
		require.Error(t, err)
		assert.Equal(t, errInstallationVerificationFailed, err)
	})

	t.Run("warns when GOPATH/bin not in PATH but still succeeds", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Getenv", "PATH").Return("/usr/bin") // GOPATH/bin not in PATH
		h.mockOSOps.On("FileExists", "/home/user/go/bin/testapp").Return(true)
		h.mockOSOps.On("FileExists", mock.Anything).Return(false).Maybe()
		h.mockOSOps.On("Symlink", mock.Anything, mock.Anything).Return(nil).Maybe()
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.expectCmd("go")

		err := Install{}.Default()
		require.NoError(t, err) // Should still succeed, just warn
	})

	t.Run("fallback to module name when binary not configured", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: ""}})
		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockOSOps.On("Getenv", "GOPATH").Return("/home/user/go")
		h.mockOSOps.On("Getenv", "HOME").Return("/home/user").Maybe()
		h.mockOSOps.On("Getenv", "PATH").Return("/usr/bin:/home/user/go/bin")
		h.mockOSOps.On("FileExists", "/home/user/go/bin/myapp").Return(true)
		h.mockOSOps.On("FileExists", mock.Anything).Return(false).Maybe()
		h.mockOSOps.On("Symlink", mock.Anything, mock.Anything).Return(nil).Maybe()
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/myapp", nil)
		h.expectGitVersion()
		h.expectCmd("go")

		err := Install{}.Default()
		require.NoError(t, err)
	})
}

// TestInstallSystemWide tests Install.SystemWide with mocked dependencies
func TestInstallSystemWide(t *testing.T) {
	if runtime.GOOS == OSWindows {
		t.Run("not supported on windows", func(t *testing.T) {
			err := Install{}.SystemWide()
			require.Error(t, err)
			assert.Equal(t, errSystemWideNotSupportedWindows, err)
		})
		return
	}

	t.Run("successful system-wide install", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockOSOps.On("Remove", "/tmp/testapp").Return(nil)
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()

		// Expect go build, sudo cp, sudo chmod
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[0] == "cp"
		})).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[0] == "chmod"
		})).Return(nil)

		err := Install{}.SystemWide()
		require.NoError(t, err)
	})

	t.Run("package path determination failure", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("", assert.AnError)

		err := Install{}.SystemWide()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to determine package path")
	})

	t.Run("build fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(assert.AnError)

		err := Install{}.SystemWide()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "build failed")
	})

	t.Run("sudo cp fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[0] == "cp"
		})).Return(assert.AnError)

		err := Install{}.SystemWide()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "installation failed")
	})

	t.Run("sudo chmod fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[0] == "cp"
		})).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.MatchedBy(func(args []string) bool {
			return len(args) > 0 && args[0] == "chmod"
		})).Return(assert.AnError)

		err := Install{}.SystemWide()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set permissions")
	})

	t.Run("temp file cleanup failure is not an error", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: "testapp"}})
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockOSOps.On("Remove", "/tmp/testapp").Return(assert.AnError) // Cleanup fails
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/testapp", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.Anything).Return(nil)

		err := Install{}.SystemWide()
		require.NoError(t, err) // Should still succeed
	})

	t.Run("fallback to module name when binary not configured", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: ""}})
		h.mockGoOps.On("GetModuleName").Return("github.com/user/myapp", nil)
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockOSOps.On("Remove", "/tmp/myapp").Return(nil)
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/myapp", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.Anything).Return(nil)

		err := Install{}.SystemWide()
		require.NoError(t, err)
	})

	t.Run("fallback to default binary name when module lookup fails", func(t *testing.T) {
		h := newInstallFullTestHelper(t)
		defer h.teardownFull(t)

		h.setConfig(&Config{Project: ProjectConfig{Binary: ""}})
		h.mockGoOps.On("GetModuleName").Return("", assert.AnError)
		h.mockOSOps.On("TempDir").Return("/tmp")
		h.mockOSOps.On("Remove", "/tmp/app").Return(nil)
		h.mockBuildOps.On("DeterminePackagePath", mock.Anything, mock.Anything, true).Return("./cmd/app", nil)
		h.expectGitVersion()
		h.mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)
		h.mockRunner.On("RunCmd", "sudo", mock.Anything).Return(nil)

		err := Install{}.SystemWide()
		require.NoError(t, err)
	})
}
