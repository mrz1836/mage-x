package mage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSpeckitConstitutionPathUnit tests getSpeckitConstitutionPath function
func TestGetSpeckitConstitutionPathUnit(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "default path when empty",
			config: &Config{
				Speckit: SpeckitConfig{},
			},
			expected: DefaultSpeckitConstitutionPath,
		},
		{
			name: "custom path",
			config: &Config{
				Speckit: SpeckitConfig{
					ConstitutionPath: "custom/path/constitution.md",
				},
			},
			expected: "custom/path/constitution.md",
		},
		{
			name: "relative path",
			config: &Config{
				Speckit: SpeckitConfig{
					ConstitutionPath: "./my-constitution.md",
				},
			},
			expected: "./my-constitution.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSpeckitConstitutionPath(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPrintSpeckitUpgradeSummary tests printSpeckitUpgradeSummary function
func TestPrintSpeckitUpgradeSummary(t *testing.T) {
	tests := []struct {
		name       string
		oldVersion string
		newVersion string
		backupPath string
	}{
		{
			name:       "with backup path",
			oldVersion: "v0.0.19",
			newVersion: "v0.0.20",
			backupPath: "/path/to/backup.md",
		},
		{
			name:       "without backup path",
			oldVersion: "v0.0.19",
			newVersion: "v0.0.20",
			backupPath: "",
		},
		{
			name:       "unknown versions",
			oldVersion: statusUnknown,
			newVersion: statusUnknown,
			backupPath: "",
		},
		{
			name:       "same version",
			oldVersion: "v0.0.20",
			newVersion: "v0.0.20",
			backupPath: "backup.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			assert.NotPanics(t, func() {
				printSpeckitUpgradeSummary(tt.oldVersion, tt.newVersion, tt.backupPath)
			})
		})
	}
}

// TestCheckSpeckitPrerequisites tests checkSpeckitPrerequisites function
func TestCheckSpeckitPrerequisites(t *testing.T) {
	// This test verifies the function runs without panicking
	// Actual results depend on whether uv/uvx/specify are installed
	err := checkSpeckitPrerequisites()
	// We don't assert on the error since it depends on the test environment
	// Just verify it returns one of the expected errors or nil
	if err != nil {
		validErrors := []error{errUVNotInstalled, errSpecifyNotInstalled}
		found := false
		for _, validErr := range validErrors {
			if errors.Is(err, validErr) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// TestGetSpeckitVersionFromUVListWithMock tests getSpeckitVersionFromUVList
// (the fallback path that does not depend on the specify CLI being on PATH).
func TestGetSpeckitVersionFromUVListWithMock(t *testing.T) {
	// Save original runner and restore after test
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	t.Run("successful parse", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("uv tool list", "specify-cli v0.0.20\nsome-other-tool v1.0.0", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{
				CLIName: "specify-cli",
			},
		}

		version, err := getSpeckitVersionFromUVList(config)
		require.NoError(t, err)
		assert.Equal(t, "v0.0.20", version)
	})

	t.Run("with default CLI name", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("uv tool list", DefaultSpeckitCLIName+" v0.1.0", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{}, // Empty, should use default
		}

		version, err := getSpeckitVersionFromUVList(config)
		require.NoError(t, err)
		assert.Equal(t, "v0.1.0", version)
	})

	t.Run("version not found", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("uv tool list", "other-tool v1.0.0", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{
				CLIName: "specify-cli",
			},
		}

		_, err = getSpeckitVersionFromUVList(config)
		require.ErrorIs(t, err, errVersionParseFailed)
	})

	t.Run("command error", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("uv tool list", "", errCommandFailed)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{},
		}

		_, err = getSpeckitVersionFromUVList(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to run uv tool list")
	})
}

// TestBackupSpeckitConstitutionUnit tests backupSpeckitConstitution function
func TestBackupSpeckitConstitutionUnit(t *testing.T) {
	t.Run("constitution not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := &Config{
			Speckit: SpeckitConfig{
				ConstitutionPath: filepath.Join(tmpDir, "nonexistent", "constitution.md"),
				BackupDir:        filepath.Join(tmpDir, "backups"),
			},
		}

		_, err := backupSpeckitConstitution(config)
		require.ErrorIs(t, err, errConstitutionNotFound)
	})

	t.Run("successful backup", func(t *testing.T) {
		tmpDir := t.TempDir()
		constitutionDir := filepath.Join(tmpDir, ".specify", "memory")
		require.NoError(t, os.MkdirAll(constitutionDir, 0o750))

		constitutionPath := filepath.Join(constitutionDir, "constitution.md")
		testContent := "# Test Constitution\n\nThis is a test."
		require.NoError(t, os.WriteFile(constitutionPath, []byte(testContent), 0o600))

		config := &Config{
			Speckit: SpeckitConfig{
				ConstitutionPath: constitutionPath,
				BackupDir:        filepath.Join(tmpDir, "backups"),
			},
		}

		backupPath, err := backupSpeckitConstitution(config)
		require.NoError(t, err)
		require.FileExists(t, backupPath)

		// Verify backup content
		content, err := os.ReadFile(backupPath) //nolint:gosec // test with controlled path
		require.NoError(t, err)
		assert.Equal(t, testContent, string(content))
	})

	t.Run("uses default backup dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(oldWd))
		})

		// Create constitution in default location
		constitutionDir := filepath.Join(tmpDir, ".specify", "memory")
		require.NoError(t, os.MkdirAll(constitutionDir, 0o750))
		constitutionPath := filepath.Join(constitutionDir, "constitution.md")
		require.NoError(t, os.WriteFile(constitutionPath, []byte("test"), 0o600))

		config := &Config{
			Speckit: SpeckitConfig{
				ConstitutionPath: constitutionPath,
				BackupDir:        "", // Should use default
			},
		}

		backupPath, err := backupSpeckitConstitution(config)
		require.NoError(t, err)
		assert.Contains(t, backupPath, DefaultSpeckitBackupDir)
	})
}

// TestRestoreSpeckitConstitutionUnit tests restoreSpeckitConstitution function
func TestRestoreSpeckitConstitutionUnit(t *testing.T) {
	t.Run("successful restore", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create backup file
		backupDir := filepath.Join(tmpDir, "backups")
		require.NoError(t, os.MkdirAll(backupDir, 0o750))
		backupPath := filepath.Join(backupDir, "constitution.backup.md")
		backupContent := "# Restored Constitution"
		require.NoError(t, os.WriteFile(backupPath, []byte(backupContent), 0o600))

		// Target constitution path
		constitutionPath := filepath.Join(tmpDir, ".specify", "constitution.md")

		config := &Config{
			Speckit: SpeckitConfig{
				ConstitutionPath: constitutionPath,
			},
		}

		err := restoreSpeckitConstitution(config, backupPath)
		require.NoError(t, err)

		// Verify content
		content, err := os.ReadFile(constitutionPath) //nolint:gosec // test with controlled path
		require.NoError(t, err)
		assert.Equal(t, backupContent, string(content))
	})

	t.Run("backup not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := &Config{
			Speckit: SpeckitConfig{
				ConstitutionPath: filepath.Join(tmpDir, "constitution.md"),
			},
		}

		err := restoreSpeckitConstitution(config, "/nonexistent/backup.md")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read backup")
	})
}

// TestUpdateSpeckitVersionFileUnit tests updateSpeckitVersionFile function
func TestUpdateSpeckitVersionFileUnit(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		tmpDir := t.TempDir()
		versionFile := filepath.Join(tmpDir, ".specify", "version.txt")

		config := &Config{
			Speckit: SpeckitConfig{
				VersionFile: versionFile,
			},
		}

		err := updateSpeckitVersionFile(config, "v0.0.19", "v0.0.20", "/path/to/backup.md")
		require.NoError(t, err)

		content, err := os.ReadFile(versionFile) //nolint:gosec // test with controlled path
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "version=v0.0.20")
		assert.Contains(t, contentStr, "previous_version=v0.0.19")
		assert.Contains(t, contentStr, "upgrade_method=automated")
		assert.Contains(t, contentStr, "constitution_backup=/path/to/backup.md")
		assert.Contains(t, contentStr, "last_upgrade=")
	})

	t.Run("uses default version file", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(oldWd))
		})

		config := &Config{
			Speckit: SpeckitConfig{
				VersionFile: "", // Should use default
			},
		}

		err = updateSpeckitVersionFile(config, "v1", "v2", "")
		require.NoError(t, err)

		_, err = os.Stat(DefaultSpeckitVersionFile)
		require.NoError(t, err)
	})
}

// TestCleanOldSpeckitBackupsUnit tests cleanOldSpeckitBackups function
func TestCleanOldSpeckitBackupsUnit(t *testing.T) {
	t.Run("no backup directory", func(t *testing.T) {
		config := &Config{
			Speckit: SpeckitConfig{
				BackupDir:     "/nonexistent/backup/dir",
				BackupsToKeep: 5,
			},
		}

		err := cleanOldSpeckitBackups(config)
		require.NoError(t, err) // Should not error if dir doesn't exist
	})

	t.Run("fewer backups than limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		backupDir := filepath.Join(tmpDir, "backups")
		require.NoError(t, os.MkdirAll(backupDir, 0o750))

		// Create 3 backups
		for i := 1; i <= 3; i++ {
			filename := filepath.Join(backupDir, "constitution.backup.20251201.10000"+string(rune('0'+i))+".md")
			require.NoError(t, os.WriteFile(filename, []byte("backup"), 0o600))
		}

		config := &Config{
			Speckit: SpeckitConfig{
				BackupDir:     backupDir,
				BackupsToKeep: 5, // Keep more than exist
			},
		}

		err := cleanOldSpeckitBackups(config)
		require.NoError(t, err)

		entries, err := os.ReadDir(backupDir)
		require.NoError(t, err)
		assert.Len(t, entries, 3) // All should remain
	})

	t.Run("deletes old backups", func(t *testing.T) {
		tmpDir := t.TempDir()
		backupDir := filepath.Join(tmpDir, "backups")
		require.NoError(t, os.MkdirAll(backupDir, 0o750))

		// Create 10 backups with unique timestamps
		for i := 1; i <= 10; i++ {
			filename := filepath.Join(backupDir, "constitution.backup.2025120"+string(rune('0'+i))+".100000.md")
			require.NoError(t, os.WriteFile(filename, []byte("backup"), 0o600))
		}

		config := &Config{
			Speckit: SpeckitConfig{
				BackupDir:     backupDir,
				BackupsToKeep: 5,
			},
		}

		err := cleanOldSpeckitBackups(config)
		require.NoError(t, err)

		entries, err := os.ReadDir(backupDir)
		require.NoError(t, err)
		assert.Len(t, entries, 5) // Only 5 should remain
	})

	t.Run("uses default backups to keep", func(t *testing.T) {
		tmpDir := t.TempDir()
		backupDir := filepath.Join(tmpDir, "backups")
		require.NoError(t, os.MkdirAll(backupDir, 0o750))

		config := &Config{
			Speckit: SpeckitConfig{
				BackupDir:     backupDir,
				BackupsToKeep: 0, // Should use default
			},
		}

		err := cleanOldSpeckitBackups(config)
		require.NoError(t, err)
	})
}

// TestInstallSpeckitCLIWithMock tests installSpeckitCLI with mock runner.
// The install path always pins to a tag and uses --force --from to switch
// the source from any unofficial PyPI install to the official git release.
func TestInstallSpeckitCLIWithMock(t *testing.T) {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	const tag = "v0.8.1"

	t.Run("successful install", func(t *testing.T) {
		mock := NewBmadMockRunner()
		expectedCmd := "uv tool install --force --from git+https://github.com/github/spec-kit.git@" + tag + " specify-cli"
		mock.SetOutput(expectedCmd, "", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{
				CLIName: "specify-cli",
				GitURL:  "https://github.com/github/spec-kit.git",
			},
		}

		err = installSpeckitCLI(config, tag)
		require.NoError(t, err)

		commands := mock.GetCommands()
		require.Len(t, commands, 1)
		assert.Equal(t, expectedCmd, commands[0])
	})

	t.Run("uses default CLI name and git URL", func(t *testing.T) {
		mock := NewBmadMockRunner()
		expectedCmd := "uv tool install --force --from git+" + DefaultSpeckitGitURL + "@" + tag + " " + DefaultSpeckitCLIName
		mock.SetOutput(expectedCmd, "", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{
			Speckit: SpeckitConfig{}, // Empty, should use defaults
		}

		err = installSpeckitCLI(config, tag)
		require.NoError(t, err)

		commands := mock.GetCommands()
		require.Len(t, commands, 1)
		assert.Equal(t, expectedCmd, commands[0])
	})

	t.Run("rejects invalid tag", func(t *testing.T) {
		mock := NewBmadMockRunner()
		err := SetRunner(mock)
		require.NoError(t, err)

		config := &Config{Speckit: SpeckitConfig{}}

		err = installSpeckitCLI(config, "main")
		require.Error(t, err)
		require.ErrorIs(t, err, errSpeckitTagInvalid)
		assert.Empty(t, mock.GetCommands(), "no command should run when tag is invalid")
	})
}

// TestUpgradeSpeckitProjectConfigPinsTagAndIntegration verifies that the
// project-init step pins the tag in the --from value and uses --integration
// (the modern flag), not the deprecated --ai flag.
func TestUpgradeSpeckitProjectConfigPinsTagAndIntegration(t *testing.T) {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	t.Setenv("GH_TOKEN", "set-by-test")
	t.Setenv("GITHUB_TOKEN", "")

	const tag = "v0.8.1"
	expectedCmd := "uvx --from git+https://github.com/github/spec-kit.git@" + tag + " specify init --here --force --integration claude"

	mock := NewBmadMockRunner()
	mock.SetOutput(expectedCmd, "", nil)
	err := SetRunner(mock)
	require.NoError(t, err)

	config := &Config{
		Speckit: SpeckitConfig{
			GitURL:      "https://github.com/github/spec-kit.git",
			Integration: "claude",
		},
	}

	err = upgradeSpeckitProjectConfig(config, tag)
	require.NoError(t, err)

	commands := mock.GetCommands()
	require.Len(t, commands, 1, "GH_TOKEN is set, so no gh auth token call should occur")
	cmd := commands[0]
	assert.Contains(t, cmd, "@"+tag, "command must pin the resolved release tag")
	assert.Contains(t, cmd, "--integration claude", "command must use the modern --integration flag")
	assert.NotContains(t, cmd, "--ai ", "command must not use the deprecated --ai flag")
}

// TestUpgradeSpeckitProjectConfigRejectsInvalidTag confirms the project-init
// step refuses to run with a non-release tag, so we never silently fall back
// to the default branch.
func TestUpgradeSpeckitProjectConfigRejectsInvalidTag(t *testing.T) {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	mock := NewBmadMockRunner()
	require.NoError(t, SetRunner(mock))

	err := upgradeSpeckitProjectConfig(&Config{Speckit: SpeckitConfig{}}, "main")
	require.Error(t, err)
	require.ErrorIs(t, err, errSpeckitTagInvalid)
	assert.Empty(t, mock.GetCommands(), "no command should run when the tag is invalid")
}

// TestVerifySpeckitInstallationWithMock tests verifySpeckitInstallation with mock
func TestVerifySpeckitInstallationWithMock(t *testing.T) {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	t.Run("successful verification", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("specify check", "", nil)

		err := SetRunner(mock)
		require.NoError(t, err)

		err = verifySpeckitInstallation()
		require.NoError(t, err)
	})

	t.Run("verification failure", func(t *testing.T) {
		mock := NewBmadMockRunner()
		mock.SetOutput("specify check", "", errCheckFailed)

		err := SetRunner(mock)
		require.NoError(t, err)

		err = verifySpeckitInstallation()
		require.Error(t, err)
	})
}

// TestUpgradeSpeckitUVToolWithMock tests upgradeSpeckitUVTool with mock.
// Upgrade emits the same command shape as install — `uv tool upgrade` cannot
// switch sources, so the implementation always reinstalls from git+...@<tag>
// with --force.
func TestUpgradeSpeckitUVToolWithMock(t *testing.T) {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		require.NoError(t, err)
	}()

	const tag = "v0.8.1"
	expectedCmd := "uv tool install --force --from git+https://github.com/github/spec-kit.git@" + tag + " specify-cli"

	mock := NewBmadMockRunner()
	mock.SetOutput(expectedCmd, "", nil)

	err := SetRunner(mock)
	require.NoError(t, err)

	config := &Config{
		Speckit: SpeckitConfig{
			CLIName: "specify-cli",
			GitURL:  "https://github.com/github/spec-kit.git",
		},
	}

	err = upgradeSpeckitUVTool(config, tag)
	require.NoError(t, err)

	commands := mock.GetCommands()
	require.Len(t, commands, 1)
	assert.Equal(t, expectedCmd, commands[0])
}

// TestSpeckitStaticErrors tests that static errors are properly defined
func TestSpeckitStaticErrors(t *testing.T) {
	errs := []error{
		errUVNotInstalled,
		errSpecifyNotInstalled,
		errConstitutionNotFound,
		errBackupFailed,
		errVersionParseFailed,
		errSpeckitInstallFailed,
		errSpeckitAlreadyInstalled,
		errSpeckitTagResolveFailed,
		errSpeckitTagInvalid,
		errSpeckitTagEmpty,
	}
	for _, e := range errs {
		require.Error(t, e)
		require.NotEmpty(t, e.Error())
	}
}

// TestSpeckitInstallBlocksOnExistingConstitution tests that Install returns an error
// when a constitution file already exists, forcing users to use Upgrade instead
func TestSpeckitInstallBlocksOnExistingConstitution(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	// Create the constitution file
	constitutionDir := filepath.Join(tmpDir, ".specify", "memory")
	require.NoError(t, os.MkdirAll(constitutionDir, 0o750))
	constitutionPath := filepath.Join(constitutionDir, "constitution.md")
	require.NoError(t, os.WriteFile(constitutionPath, []byte("# Existing Constitution"), 0o600))

	// Set up config with the test constitution path
	testConfig := &Config{
		Speckit: SpeckitConfig{
			ConstitutionPath: constitutionPath,
		},
	}
	TestSetConfig(testConfig)
	t.Cleanup(func() {
		TestResetConfig()
	})

	// Call Install and verify it returns the expected error
	speckit := Speckit{}
	err := speckit.Install()
	require.Error(t, err)
	require.ErrorIs(t, err, errSpeckitAlreadyInstalled)
}

// TestGetGitHubTokenFromGH tests getGitHubTokenFromGH function
func TestGetGitHubTokenFromGH(t *testing.T) {
	// This test verifies the function runs without panicking
	// Actual results depend on whether gh is installed and authenticated
	token := getGitHubTokenFromGH()
	// Token can be empty or a valid token - just verify it doesn't panic
	_ = token
}
