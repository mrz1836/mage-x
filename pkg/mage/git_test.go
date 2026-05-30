//go:build integration
// +build integration

package mage

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// GitTestSuite defines the test suite for Git functions
type GitTestSuite struct {
	suite.Suite

	env *testutil.TestEnvironment
	git Git
}

// SetupTest runs before each test
func (ts *GitTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.git = Git{}
}

// TearDownTest runs after each test
func (ts *GitTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGitDiff tests the Diff method
func (ts *GitTestSuite) TestGitDiff() {
	ts.Run("clean working directory", func() {
		// Mock clean git diff
		ts.env.Runner.On("RunCmdOutput", "git", []string{"diff", "--exit-code"}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Diff()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitTag tests the Tag and TagWithArgs methods
func (ts *GitTestSuite) TestGitTag() {
	ts.Run("bare Tag requires TagWithArgs", func() {
		// Tag() is a stub that directs callers to TagWithArgs and runs no git commands.
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Tag()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "use TagWithArgs instead")
	})

	ts.Run("successful tag creation", func() {
		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v1.2.3", "-m", "Pending full release..."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagWithArgs("version=1.2.3")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("missing version parameter", func() {
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagWithArgs()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "version parameter is required")
	})
}

// TestGitTagRemove tests the TagRemove and TagRemoveWithArgs methods
func (ts *GitTestSuite) TestGitTagRemove() {
	ts.Run("bare TagRemove requires TagRemoveWithArgs", func() {
		// TagRemove() is a stub that directs callers to TagRemoveWithArgs and runs no git commands.
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagRemove()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "use TagRemoveWithArgs instead")
	})

	ts.Run("successful tag removal", func() {
		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-d", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "--delete", "origin", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagRemoveWithArgs("version=1.2.3")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("missing version parameter", func() {
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagRemoveWithArgs()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "version parameter is required")
	})
}

// TestGitTagUpdate tests the TagUpdate method
func (ts *GitTestSuite) TestGitTagUpdate() {
	ts.Run("successful tag update", func() {
		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"push", "--force", "origin", "HEAD:refs/tags/v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagUpdate("version=1.2.3")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("missing version parameter", func() {
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagUpdate()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "version parameter is required")
	})
}

// TestGitStatus tests the Status method
func (ts *GitTestSuite) TestGitStatus() {
	ts.Run("successful status display", func() {
		// Mock successful git status
		ts.env.Runner.On("RunCmd", "git", []string{"status"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Status()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitLog tests the Log method
func (ts *GitTestSuite) TestGitLog() {
	ts.Run("successful log display", func() {
		// Mock successful git log
		ts.env.Runner.On("RunCmd", "git", []string{"log", "--oneline", "-10"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Log()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitBranch tests the Branch method
func (ts *GitTestSuite) TestGitBranch() {
	ts.Run("successful branch display", func() {
		// Mock successful git branch commands
		ts.env.Runner.On("RunCmdOutput", "git", []string{"branch", "--show-current"}).Return("main", nil)
		ts.env.Runner.On("RunCmd", "git", []string{"branch", "-a"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Branch()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitPull tests the Pull method
func (ts *GitTestSuite) TestGitPull() {
	ts.Run("successful pull", func() {
		// Mock successful git fetch and pull
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--all", "--prune"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"pull", "--rebase"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Pull()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitCommit tests the Commit method
func (ts *GitTestSuite) TestGitCommit() {
	ts.Run("commit with message parameter", func() {
		// Mock successful git add and commit
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "test commit"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Commit("message=test commit")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("commit with message containing spaces", func() {
		// Mock successful git add and commit
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "feature commit message"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Commit("message=feature commit message")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("missing commit message", func() {
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Commit()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "message parameter is required")
	})
}

// TestGitInit tests the Init method
func (ts *GitTestSuite) TestGitInit() {
	ts.Run("successful init", func() {
		// Mock successful git init
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Init()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitAdd tests the Add method
func (ts *GitTestSuite) TestGitAdd() {
	ts.Run("add all files (no arguments)", func() {
		// Mock successful git add .
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Add()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("add specific files", func() {
		// Mock successful git add with specific files
		ts.env.Runner.On("RunCmd", "git", []string{"add", "file1.go", "file2.go"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Add("file1.go", "file2.go")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitClone tests the Clone method
func (ts *GitTestSuite) TestGitClone() {
	ts.Run("successful clone", func() {
		// Mock successful git clone
		ts.env.Runner.On("RunCmd", "git", []string{"clone"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Clone()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitPush tests the Push method
func (ts *GitTestSuite) TestGitPush() {
	ts.Run("successful push", func() {
		// Mock successful git push
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "main"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.Push("origin", "main")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitPullWithRemote tests the PullWithRemote method
func (ts *GitTestSuite) TestGitPullWithRemote() {
	ts.Run("successful pull with remote", func() {
		// Mock successful git pull
		ts.env.Runner.On("RunCmd", "git", []string{"pull", "origin", "main"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.PullWithRemote("origin", "main")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitTagWithMessage tests the TagWithMessage method
func (ts *GitTestSuite) TestGitTagWithMessage() {
	ts.Run("successful tag with message", func() {
		// Mock successful git tag
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v1.0.0", "-m", "Release version 1.0.0"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.TagWithMessage("v1.0.0", "Release version 1.0.0")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitBranchWithName tests the BranchWithName method
func (ts *GitTestSuite) TestGitBranchWithName() {
	ts.Run("successful branch creation", func() {
		// Mock successful git branch
		ts.env.Runner.On("RunCmd", "git", []string{"branch", "feature-branch"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.BranchWithName("feature-branch")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitCloneRepo tests the CloneRepo method
func (ts *GitTestSuite) TestGitCloneRepo() {
	ts.Run("successful repository clone", func() {
		// Mock successful git clone
		ts.env.Runner.On("RunCmd", "git", []string{"clone", "https://github.com/user/repo.git", "local-dir"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.CloneRepo("https://github.com/user/repo.git", "local-dir")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitLogWithCount tests the LogWithCount method
func (ts *GitTestSuite) TestGitLogWithCount() {
	ts.Run("successful log with count", func() {
		// Mock successful git log
		ts.env.Runner.On("RunCmd", "git", []string{"log", "-5"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.git.LogWithCount(5)
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitIntegration tests integration scenarios
func (ts *GitTestSuite) TestGitIntegration() {
	ts.Run("git operations with parameters", func() {
		// Mock git operations driven by command-line parameters
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "Version 2.0.0 release"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v2.0.0", "-m", "Pending full release..."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "v2.0.0"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				if err := ts.git.Commit("message=Version 2.0.0 release"); err != nil {
					return err
				}
				return ts.git.TagWithArgs("version=2.0.0")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGitTestSuite runs the test suite
func TestGitTestSuite(t *testing.T) {
	suite.Run(t, new(GitTestSuite))
}
