package mage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// VersionProtectionTestSuite provides comprehensive tests to prevent unexpected version bumps
// This test suite ensures that parameters are handled correctly and prevents issues like
// accidentally jumping from v1.0.6 to v2.0.0 instead of v1.0.7 due to parameter contamination
type VersionProtectionTestSuite struct {
	suite.Suite

	originalRunner interface{}
}

// SetupSuite runs once before all tests
func (vpts *VersionProtectionTestSuite) SetupSuite() {
	// Save original runner
	vpts.originalRunner = GetRunner()
}

// TearDownSuite runs once after all tests
func (vpts *VersionProtectionTestSuite) TearDownSuite() {
	// Restore original runner
	_ = SetRunner(vpts.originalRunner.(CommandRunner)) //nolint:errcheck // Cleanup operation
}

// SetupTest runs before each test
func (vpts *VersionProtectionTestSuite) SetupTest() {
	// No cleanup needed since we use parameters instead of environment variables
}

// TestVersionProtectionSuite runs the protection test suite
func TestVersionProtectionSuite(t *testing.T) {
	suite.Run(t, new(VersionProtectionTestSuite))
}

// TestParameterHandling tests that parameters are handled correctly without contamination
func (vpts *VersionProtectionTestSuite) TestParameterHandling() {
	vpts.Run("DefaultBumpWhenNoParameter", func() {
		// Test parsing empty parameters
		params := utils.ParseParams([]string{})
		bumpType := utils.GetParam(params, "bump", "patch")
		vpts.Equal("patch", bumpType, "Default bump should be 'patch' when no parameter is provided")
	})

	vpts.Run("ExplicitBumpParameter", func() {
		testCases := []string{"major", "minor", "patch"}

		for _, expectedBump := range testCases {
			vpts.Run(fmt.Sprintf("Bump_%s", expectedBump), func() {
				params := utils.ParseParams([]string{"bump=" + expectedBump})
				bumpType := utils.GetParam(params, "bump", "patch")
				vpts.Equal(expectedBump, bumpType, "Explicit bump parameter should override default")
			})
		}
	})

	vpts.Run("ParameterIsolation", func() {
		// Test that different parameter sets don't interfere with each other
		params1 := utils.ParseParams([]string{"bump=major", "push"})
		params2 := utils.ParseParams([]string{"bump=patch", "dry-run"})

		// First parameter set
		bumpType1 := utils.GetParam(params1, "bump", "minor")
		vpts.Equal("major", bumpType1)
		vpts.True(utils.HasParam(params1, "push"))
		vpts.False(utils.HasParam(params1, "dry-run"))

		// Second parameter set (independent)
		bumpType2 := utils.GetParam(params2, "bump", "minor")
		vpts.Equal("patch", bumpType2)
		vpts.False(utils.HasParam(params2, "push"))
		vpts.True(utils.HasParam(params2, "dry-run"))
	})

	vpts.Run("EmptyParameterValue", func() {
		// Test handling of empty parameter values
		params := utils.ParseParams([]string{"bump="})
		bumpType := utils.GetParam(params, "bump", "patch")
		vpts.Empty(bumpType, "Empty parameter value should be preserved")
	})
}

// TestCommandSimulation tests the exact command scenarios that could cause issues
func (vpts *VersionProtectionTestSuite) TestCommandSimulation() {
	vpts.Run("PushWithDefaultBump", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Set up mock for successful patch bump
		mock.SetOutput("git status --porcelain", "", nil)
		mock.SetOutput("git tag --points-at HEAD", "", nil)
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git tag --sort=-version:refname", "v1.0.6", nil) // Highest tag in repo
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)  // Reachable tag
		mock.SetOutput("git tag -a v1.0.7 -m GitHubRelease v1.0.7", "", nil)
		// Mock git remote validation
		mock.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)", nil)
		mock.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main", nil)
		mock.SetOutput("git push origin v1.0.7", "", nil)

		// Run version bump with push parameter (default bump type)
		version := Version{}
		err := version.Bump("push")
		vpts.Require().NoError(err, "push with default bump should succeed")

		// Verify it was a patch bump (v1.0.6 -> v1.0.7)
		vpts.Contains(mock.commands, "git tag -a v1.0.7 -m GitHubRelease v1.0.7",
			"Should create v1.0.7 tag (patch bump from v1.0.6)")
		vpts.Contains(mock.commands, "git push origin v1.0.7",
			"Should push the v1.0.7 tag")
	})

	vpts.Run("PushWithExplicitPatch", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Set up mock for successful patch bump
		mock.SetOutput("git status --porcelain", "", nil)
		mock.SetOutput("git tag --points-at HEAD", "", nil)
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git tag --sort=-version:refname", "v1.0.6", nil) // Highest tag in repo
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)  // Reachable tag
		mock.SetOutput("git tag -a v1.0.7 -m GitHubRelease v1.0.7", "", nil)
		// Mock git remote validation
		mock.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)", nil)
		mock.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main", nil)
		mock.SetOutput("git push origin v1.0.7", "", nil)

		version := Version{}
		err := version.Bump("bump=patch", "push")
		vpts.Require().NoError(err)

		vpts.Contains(mock.commands, "git tag -a v1.0.7 -m GitHubRelease v1.0.7")
	})

	vpts.Run("AccidentalMajorBumpPrevented", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Set up mock
		mock.SetOutput("git status --porcelain", "", nil)
		mock.SetOutput("git tag --points-at HEAD", "", nil)
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git tag --sort=-version:refname", "v1.0.6", nil) // Highest tag in repo
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)  // Reachable tag

		version := Version{}
		err := version.Bump("bump=major", "push") // Deliberately NOT passing "confirm" to test protection
		vpts.Require().Error(err, "Should prevent major bump without confirmation")
		vpts.Contains(err.Error(), "major version bump requires explicit confirmation")

		// Verify no major version tag was created
		vpts.NotContains(mock.commands, "git tag -a v2.0.0",
			"Enhanced protection: major bump blocked without confirmation")
	})
}

// TestVersionJumpProtection tests validation of version progression
func (vpts *VersionProtectionTestSuite) TestVersionJumpProtection() {
	vpts.Run("DetectUnexpectedMajorJump", func() {
		// Test the validateVersionProgression function directly
		err := validateVersionProgression("v1.0.6", "v2.0.0", "patch")
		vpts.Require().Error(err, "Should detect illogical version jump")
		vpts.Require().ErrorIs(err, errIllogicalVersionJump)
		vpts.Contains(err.Error(), "expected v1.0.6 → v1.0.7",
			"Error should specify expected patch increment")
	})

	vpts.Run("DetectUnexpectedMinorJump", func() {
		err := validateVersionProgression("v1.0.6", "v1.1.0", "patch")
		vpts.Require().Error(err)
		vpts.Require().ErrorIs(err, errIllogicalVersionJump)
	})

	vpts.Run("AllowValidPatchProgression", func() {
		err := validateVersionProgression("v1.0.6", "v1.0.7", "patch")
		vpts.NoError(err, "Valid patch progression should be allowed")
	})

	vpts.Run("AllowValidMinorProgression", func() {
		err := validateVersionProgression("v1.0.6", "v1.1.0", "minor")
		vpts.NoError(err, "Valid minor progression should be allowed")
	})

	vpts.Run("AllowValidMajorProgression", func() {
		err := validateVersionProgression("v1.0.6", "v2.0.0", "major")
		vpts.NoError(err, "Valid major progression should be allowed")
	})
}

// TestBumpVersionEdgeCases tests the bumpVersion function with edge cases
func (vpts *VersionProtectionTestSuite) TestBumpVersionEdgeCases() {
	vpts.Run("BumpFromActualTags", func() {
		testCases := []struct {
			current  string
			bumpType string
			expected string
		}{
			{"v1.0.6", "patch", "v1.0.7"},
			{"v1.0.6", "minor", "v1.1.0"},
			{"v1.0.6", "major", "v2.0.0"},
			{"v0.0.1", "patch", "v0.0.2"},
			{"v0.1.0", "minor", "v0.2.0"},
			{"v1.0.0", "major", "v2.0.0"},
		}

		for _, tc := range testCases {
			vpts.Run(fmt.Sprintf("%s_%s_to_%s", tc.current, tc.bumpType, tc.expected), func() {
				result, err := bumpVersion(tc.current, tc.bumpType)
				vpts.Require().NoError(err)
				vpts.Equal(tc.expected, result,
					"Bump %s with %s should produce %s", tc.current, tc.bumpType, tc.expected)
			})
		}
	})

	vpts.Run("BumpWithoutVPrefix", func() {
		result, err := bumpVersion("1.0.6", "patch")
		vpts.Require().NoError(err)
		vpts.Equal("v1.0.7", result, "Should add v prefix to result")
	})

	vpts.Run("PreventInvalidBumpTypes", func() {
		invalidTypes := []string{"hotfix", "beta", "alpha", "rc", "invalid"}

		for _, invalidType := range invalidTypes {
			vpts.Run(fmt.Sprintf("InvalidType_%s", invalidType), func() {
				// This test validates the Version.Bump method validation
				mock := NewVersionMockRunner()
				vpts.Require().NoError(SetRunner(mock))

				version := Version{}
				err := version.Bump("bump=" + invalidType)

				vpts.Require().Error(err, "Invalid bump type should be rejected")
				vpts.ErrorIs(err, errInvalidBumpType)
			})
		}

		// Test default bump type (when no bump parameter is provided)
		vpts.Run("DefaultBumpType", func() {
			mock := NewVersionMockRunner()
			vpts.Require().NoError(SetRunner(mock))

			// Set up successful mock for default patch bump
			mock.SetOutput("git status --porcelain", "", nil)
			mock.SetOutput("git tag --points-at HEAD", "", nil)
			mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
			mock.SetOutput("git describe --tags --abbrev=0", "v1.0.0", nil)
			mock.SetOutput("git tag -a v1.0.1 -m GitHubRelease v1.0.1", "", nil)

			version := Version{}
			err := version.Bump() // No parameters - should default to patch
			vpts.NoError(err)
		})
	})
}

// TestGitTagScenarios tests getCurrentGitTag behavior with complex tag scenarios
func (vpts *VersionProtectionTestSuite) TestGitTagScenarios() {
	vpts.Run("MultipleTagsPreferHighestVersion", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Simulate multiple tags on HEAD - should prefer highest version
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD",
			"v2.0.0\nv1.0.6\nv1.0.5\nv0.9.0", nil)

		tag := getCurrentGitTag()
		vpts.Equal("v2.0.0", tag, "Should prefer highest version when multiple tags exist")
	})

	vpts.Run("FallbackToDescribeWhenNoTagsOnHEAD", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// No tags on HEAD
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		// But tags exist in history
		mock.SetOutput("git tag --sort=-version:refname", "v1.0.6", nil) // Highest tag in repo
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)  // Reachable tag

		tag := getCurrentGitTag()
		vpts.Equal("v1.0.6", tag, "Should fall back to git describe when no tags on HEAD")
	})

	vpts.Run("HandleNoTagsAnywhere", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// No tags anywhere
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git tag --sort=-version:refname", "", errNoTags) // No tags in repo
		mock.SetOutput("git describe --tags --abbrev=0", "", errNoTags)  // Fallback also fails

		tag := getCurrentGitTag()
		vpts.Empty(tag, "Should return empty string when no tags exist")
	})
}

// TestDryRunProtection tests that dry-run mode prevents actual version bumps
func (vpts *VersionProtectionTestSuite) TestDryRunProtection() {
	vpts.Run("DryRunWithMajorBump", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Set up mock for dry run
		mock.SetOutput("git status --porcelain", "", nil)
		mock.SetOutput("git tag --points-at HEAD", "", nil)
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)

		version := Version{}
		err := version.Bump("bump=major", "dry-run", "push")
		vpts.Require().NoError(err, "Dry run should always succeed")

		// Verify no actual git commands were executed
		vpts.NotContains(mock.commands, "git tag -a v2.0.0",
			"Dry run should not create actual tags")
		vpts.NotContains(mock.commands, "git push",
			"Dry run should not push anything")
	})

	vpts.Run("DryRunWithContaminatedEnvironment", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Even dirty repo should work in dry run
		mock.SetOutput("git status --porcelain", "M some-file.go", nil)
		mock.SetOutput("git tag --points-at HEAD", "", nil)
		mock.SetOutput("git tag --sort=-version:refname --points-at HEAD", "", errNoTags)
		mock.SetOutput("git describe --tags --abbrev=0", "v1.0.6", nil)

		version := Version{}
		err := version.Bump("bump=major", "dry-run")
		vpts.NoError(err, "Dry run should work even with dirty repo")
	})
}

// TestErrorConditionPrevention tests scenarios that should prevent version bumping
func (vpts *VersionProtectionTestSuite) TestErrorConditionPrevention() {
	vpts.Run("PreventBumpWithUncommittedChanges", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Simulate dirty working directory
		mock.SetOutput("git status --porcelain", "M pkg/mage/version.go\n?? test-file.txt", nil)

		version := Version{}
		err := version.Bump()
		vpts.Require().Error(err, "Should prevent bump with uncommitted changes")
		vpts.ErrorIs(err, errVersionUncommittedChanges)
	})

	vpts.Run("PreventBumpWhenTagsExistOnCommit", func() {
		mock := NewVersionMockRunner()
		vpts.Require().NoError(SetRunner(mock))

		// Clean working directory
		mock.SetOutput("git status --porcelain", "", nil)
		// But tags already exist on current commit
		mock.SetOutput("git tag --points-at HEAD", "v1.0.6\nv1.0.7", nil)

		version := Version{}
		err := version.Bump()
		vpts.Require().Error(err, "Should prevent bump when tags already exist on commit")
		vpts.ErrorIs(err, errMultipleTagsOnCommit)
	})
}

// TestVersionStringParsing tests version string parsing edge cases
func (vpts *VersionProtectionTestSuite) TestVersionStringParsing() {
	vpts.Run("ParseVersionsWithDifferentFormats", func() {
		validVersions := []string{
			"v1.0.6",
			"1.0.6",
			"v0.0.1",
			"v10.20.30",
		}

		for _, version := range validVersions {
			vpts.Run(fmt.Sprintf("Parse_%s", version), func() {
				result, err := bumpVersion(version, "patch")
				vpts.Require().NoError(err, "Should parse valid version: %s", version)
				vpts.NotEmpty(result, "Should return non-empty result")
				vpts.Equal(byte('v'), result[0], "Result should start with 'v'")
			})
		}
	})

	vpts.Run("RejectInvalidVersionFormats", func() {
		invalidVersions := []string{
			"1.0",     // Missing patch
			"1.0.0.0", // Too many parts
			"v1.0.x",  // Non-numeric patch
			"v1.x.0",  // Non-numeric minor
			"vx.0.0",  // Non-numeric major
			"",        // Empty
			"invalid", // Not a version
		}

		for _, version := range invalidVersions {
			vpts.Run(fmt.Sprintf("Reject_%s", version), func() {
				_, err := bumpVersion(version, "patch")
				vpts.Require().Error(err, "Should reject invalid version: %s", version)
			})
		}
	})

	vpts.Run("AcceptPrereleaseVersions", func() {
		// Pre-release and build metadata are now supported
		validPrereleaseVersions := []struct {
			input    string
			expected string
		}{
			{"1.0.0-alpha", "v1.0.1"},
			{"v1.0.0+build", "v1.0.1"},
			{"v1.0.0-beta.1", "v1.0.1"},
			{"v1.0.0-rc.2", "v1.0.1"},
		}

		for _, tc := range validPrereleaseVersions {
			vpts.Run(fmt.Sprintf("Accept_%s", tc.input), func() {
				result, err := bumpVersion(tc.input, "patch")
				vpts.Require().NoError(err, "Should accept pre-release version: %s", tc.input)
				vpts.Equal(tc.expected, result, "Bumping %s should give %s", tc.input, tc.expected)
			})
		}
	})
}

// Benchmark tests to ensure version operations are performant
func BenchmarkVersionProtection(b *testing.B) {
	b.Run("ParseParams", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = utils.ParseParams([]string{"bump=patch", "push", "dry-run"})
		}
	})

	b.Run("GetParam", func(b *testing.B) {
		params := utils.ParseParams([]string{"bump=patch", "push", "dry-run"})
		for i := 0; i < b.N; i++ {
			_ = utils.GetParam(params, "bump", "patch")
		}
	})

	b.Run("BumpVersionCalculation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = bumpVersion("v1.0.6", "patch") //nolint:errcheck // Benchmark intentionally ignores errors
		}
	})

	b.Run("ValidateVersionProgression", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = validateVersionProgression("v1.0.6", "v1.0.7", "patch") //nolint:errcheck // Benchmark intentionally ignores errors
		}
	})
}
