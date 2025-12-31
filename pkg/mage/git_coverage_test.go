package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// GitCoverageTestSuite provides comprehensive coverage for Git methods
type GitCoverageTestSuite struct {
	suite.Suite

	env *testutil.TestEnvironment
	git Git
}

func TestGitCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(GitCoverageTestSuite))
}

func (ts *GitCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.git = Git{}

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *GitCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *GitCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// TestGitDiffSuccess tests Git.Diff when working directory is clean
func (ts *GitCoverageTestSuite) TestGitDiffSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Diff()
	})

	ts.Require().NoError(err)
}

// TestGitStatusSuccess tests Git.Status
func (ts *GitCoverageTestSuite) TestGitStatusSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Status()
	})

	ts.Require().NoError(err)
}

// TestGitLogSuccess tests Git.Log
func (ts *GitCoverageTestSuite) TestGitLogSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Log()
	})

	ts.Require().NoError(err)
}

// TestGitBranchSuccess tests Git.Branch
func (ts *GitCoverageTestSuite) TestGitBranchSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Branch()
	})

	ts.Require().NoError(err)
}

// TestGitPullSuccess tests Git.Pull
func (ts *GitCoverageTestSuite) TestGitPullSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Pull()
	})

	ts.Require().NoError(err)
}

// TestGitInitSuccess tests Git.Init
func (ts *GitCoverageTestSuite) TestGitInitSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Init()
	})

	ts.Require().NoError(err)
}

// TestGitAddSuccess tests Git.Add
func (ts *GitCoverageTestSuite) TestGitAddSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Add(".")
	})

	ts.Require().NoError(err)
}

// TestGitCloneSuccess tests Git.Clone
func (ts *GitCoverageTestSuite) TestGitCloneSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Clone()
	})

	ts.Require().NoError(err)
}

// TestGitPushSuccess tests Git.Push
func (ts *GitCoverageTestSuite) TestGitPushSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Push("origin", "main")
	})

	ts.Require().NoError(err)
}

// TestGitPullWithRemoteSuccess tests Git.PullWithRemote
func (ts *GitCoverageTestSuite) TestGitPullWithRemoteSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.PullWithRemote("origin", "main")
	})

	ts.Require().NoError(err)
}

// TestGitLogWithCountSuccess tests Git.LogWithCount
func (ts *GitCoverageTestSuite) TestGitLogWithCountSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.LogWithCount(10)
	})

	ts.Require().NoError(err)
}

// TestGitBranchWithNameSuccess tests Git.BranchWithName
func (ts *GitCoverageTestSuite) TestGitBranchWithNameSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.BranchWithName("feature-branch")
	})

	ts.Require().NoError(err)
}

// TestGitCloneRepoSuccess tests Git.CloneRepo
func (ts *GitCoverageTestSuite) TestGitCloneRepoSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.CloneRepo("https://github.com/example/repo.git", "repo")
	})

	ts.Require().NoError(err)
}

// TestGitTagWithMessageSuccess tests Git.TagWithMessage
func (ts *GitCoverageTestSuite) TestGitTagWithMessageSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.TagWithMessage("v1.0.0", "Release v1.0.0")
	})

	ts.Require().NoError(err)
}

// TestGitTagWithArgsSuccess tests Git.TagWithArgs with valid version
func (ts *GitCoverageTestSuite) TestGitTagWithArgsSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.TagWithArgs("version=1.0.0")
	})

	ts.Require().NoError(err)
}

// TestGitTagRemoveWithArgsSuccess tests Git.TagRemoveWithArgs with valid version
func (ts *GitCoverageTestSuite) TestGitTagRemoveWithArgsSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.TagRemoveWithArgs("version=1.0.0")
	})

	ts.Require().NoError(err)
}

// TestGitTagUpdateSuccess tests Git.TagUpdate
func (ts *GitCoverageTestSuite) TestGitTagUpdateSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.TagUpdate("version=1.0.0")
	})

	ts.Require().NoError(err)
}

// TestGitCommitSuccess tests Git.Commit with valid message
func (ts *GitCoverageTestSuite) TestGitCommitSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.git.Commit("message=test commit")
	})

	ts.Require().NoError(err)
}

// ================== Standalone Tests ==================

func TestGitTagRequiresArgs(t *testing.T) {
	git := Git{}

	// Tag() should return error indicating args are required
	err := git.Tag()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version parameter")
}

func TestGitTagRemoveRequiresArgs(t *testing.T) {
	git := Git{}

	// TagRemove() should return error indicating args are required
	err := git.TagRemove()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version parameter")
}

func TestGitTagWithArgsRequiresVersion(t *testing.T) {
	git := Git{}

	// TagWithArgs without version should return error
	err := git.TagWithArgs()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestGitTagWithArgsInvalidVersion(t *testing.T) {
	git := Git{}

	// TagWithArgs with invalid version should return error
	err := git.TagWithArgs("version=invalid-version")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGitTagRemoveWithArgsRequiresVersion(t *testing.T) {
	git := Git{}

	// TagRemoveWithArgs without version should return error
	err := git.TagRemoveWithArgs()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestGitTagRemoveWithArgsInvalidVersion(t *testing.T) {
	git := Git{}

	// TagRemoveWithArgs with invalid version should return error
	err := git.TagRemoveWithArgs("version=invalid-version")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGitTagUpdateRequiresVersion(t *testing.T) {
	git := Git{}

	// TagUpdate without version should return error
	err := git.TagUpdate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
}

func TestGitTagUpdateInvalidVersion(t *testing.T) {
	git := Git{}

	// TagUpdate with invalid version should return error
	err := git.TagUpdate("version=invalid-version")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestGitCommitRequiresMessage(t *testing.T) {
	git := Git{}

	// Commit without message should return error
	err := git.Commit()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "message")
}

func TestGitAddWithEmptyPatternFails(t *testing.T) {
	git := Git{}

	// Add with empty string pattern will fail (git add "" fails)
	err := git.Add("")
	// Error occurs because git add "" fails, just verify an error occurred
	require.Error(t, err)
}
