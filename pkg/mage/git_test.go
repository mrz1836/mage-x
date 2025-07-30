package mage

import (
	"os"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Diff()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitTag tests the Tag method
func (ts *GitTestSuite) TestGitTag() {
	ts.Run("successful tag creation", func() {
		// Set version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			require.NoError(ts.T(), os.Setenv("version", originalVersion))
		}()
		require.NoError(ts.T(), os.Setenv("version", "1.2.3"))

		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v1.2.3", "-m", "Pending full release..."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Tag()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("missing version environment variable", func() {
		// Clear version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			if err := os.Setenv("version", originalVersion); err != nil {
				ts.T().Logf("Failed to restore version: %v", err)
			}
		}()
		if err := os.Unsetenv("version"); err != nil {
			ts.T().Logf("Failed to unset version: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Tag()
			},
		)

		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "version variable is required")
	})
}

// TestGitTagRemove tests the TagRemove method
func (ts *GitTestSuite) TestGitTagRemove() {
	ts.Run("successful tag removal", func() {
		// Set version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			require.NoError(ts.T(), os.Setenv("version", originalVersion))
		}()
		require.NoError(ts.T(), os.Setenv("version", "1.2.3"))

		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-d", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "--delete", "origin", "v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.TagRemove()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("missing version environment variable", func() {
		// Clear version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			if err := os.Setenv("version", originalVersion); err != nil {
				ts.T().Logf("Failed to restore version: %v", err)
			}
		}()
		if err := os.Unsetenv("version"); err != nil {
			ts.T().Logf("Failed to unset version: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.TagRemove()
			},
		)

		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "version variable is required")
	})
}

// TestGitTagUpdate tests the TagUpdate method
func (ts *GitTestSuite) TestGitTagUpdate() {
	ts.Run("successful tag update", func() {
		// Set version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			require.NoError(ts.T(), os.Setenv("version", originalVersion))
		}()
		require.NoError(ts.T(), os.Setenv("version", "1.2.3"))

		// Mock successful git commands
		ts.env.Runner.On("RunCmd", "git", []string{"push", "--force", "origin", "HEAD:refs/tags/v1.2.3"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.TagUpdate()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("missing version environment variable", func() {
		// Clear version environment variable
		originalVersion := os.Getenv("version")
		defer func() {
			if err := os.Setenv("version", originalVersion); err != nil {
				ts.T().Logf("Failed to restore version: %v", err)
			}
		}()
		if err := os.Unsetenv("version"); err != nil {
			ts.T().Logf("Failed to unset version: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.TagUpdate()
			},
		)

		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "version variable is required")
	})
}

// TestGitStatus tests the Status method
func (ts *GitTestSuite) TestGitStatus() {
	ts.Run("successful status display", func() {
		// Mock successful git status
		ts.env.Runner.On("RunCmd", "git", []string{"status"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Status()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitLog tests the Log method
func (ts *GitTestSuite) TestGitLog() {
	ts.Run("successful log display", func() {
		// Mock successful git log
		ts.env.Runner.On("RunCmd", "git", []string{"log", "--oneline", "-10"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Log()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitBranch tests the Branch method
func (ts *GitTestSuite) TestGitBranch() {
	ts.Run("successful branch display", func() {
		// Mock successful git branch commands
		ts.env.Runner.On("RunCmdOutput", "git", []string{"branch", "--show-current"}).Return("main", nil)
		ts.env.Runner.On("RunCmd", "git", []string{"branch", "-a"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Branch()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitPull tests the Pull method
func (ts *GitTestSuite) TestGitPull() {
	ts.Run("successful pull", func() {
		// Mock successful git fetch and pull
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--all", "--prune"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"pull", "--rebase"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Pull()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitCommit tests the Commit method
func (ts *GitTestSuite) TestGitCommit() {
	ts.Run("commit with parameter message", func() {
		// Mock successful git add and commit
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "test commit"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Commit("test commit")
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("commit with environment variable message", func() {
		// Set message environment variable
		originalMessage := os.Getenv("message")
		defer func() {
			require.NoError(ts.T(), os.Setenv("message", originalMessage))
		}()
		require.NoError(ts.T(), os.Setenv("message", "env commit message"))

		// Mock successful git add and commit
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "env commit message"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Commit()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("missing commit message", func() {
		// Clear message environment variable
		originalMessage := os.Getenv("message")
		defer func() {
			if err := os.Setenv("message", originalMessage); err != nil {
				ts.T().Logf("Failed to restore message: %v", err)
			}
		}()
		if err := os.Unsetenv("message"); err != nil {
			ts.T().Logf("Failed to unset message: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Commit()
			},
		)

		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "message parameter or environment variable is required")
	})
}

// TestGitInit tests the Init method
func (ts *GitTestSuite) TestGitInit() {
	ts.Run("successful init", func() {
		// Mock successful git init
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Init()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitAdd tests the Add method
func (ts *GitTestSuite) TestGitAdd() {
	ts.Run("add all files (no arguments)", func() {
		// Mock successful git add .
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Add()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("add specific files", func() {
		// Mock successful git add with specific files
		ts.env.Runner.On("RunCmd", "git", []string{"add", "file1.go", "file2.go"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Add("file1.go", "file2.go")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitClone tests the Clone method
func (ts *GitTestSuite) TestGitClone() {
	ts.Run("successful clone", func() {
		// Mock successful git clone
		ts.env.Runner.On("RunCmd", "git", []string{"clone"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Clone()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitPush tests the Push method
func (ts *GitTestSuite) TestGitPush() {
	ts.Run("successful push", func() {
		// Mock successful git push
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "main"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.Push("origin", "main")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitPullWithRemote tests the PullWithRemote method
func (ts *GitTestSuite) TestGitPullWithRemote() {
	ts.Run("successful pull with remote", func() {
		// Mock successful git pull
		ts.env.Runner.On("RunCmd", "git", []string{"pull", "origin", "main"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.PullWithRemote("origin", "main")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitTagWithMessage tests the TagWithMessage method
func (ts *GitTestSuite) TestGitTagWithMessage() {
	ts.Run("successful tag with message", func() {
		// Mock successful git tag
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v1.0.0", "-m", "Release version 1.0.0"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.TagWithMessage("v1.0.0", "Release version 1.0.0")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitBranchWithName tests the BranchWithName method
func (ts *GitTestSuite) TestGitBranchWithName() {
	ts.Run("successful branch creation", func() {
		// Mock successful git branch
		ts.env.Runner.On("RunCmd", "git", []string{"branch", "feature-branch"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.BranchWithName("feature-branch")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitCloneRepo tests the CloneRepo method
func (ts *GitTestSuite) TestGitCloneRepo() {
	ts.Run("successful repository clone", func() {
		// Mock successful git clone
		ts.env.Runner.On("RunCmd", "git", []string{"clone", "https://github.com/user/repo.git", "local-dir"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.CloneRepo("https://github.com/user/repo.git", "local-dir")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitLogWithCount tests the LogWithCount method
func (ts *GitTestSuite) TestGitLogWithCount() {
	ts.Run("successful log with count", func() {
		// Mock successful git log
		ts.env.Runner.On("RunCmd", "git", []string{"log", "-5"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.git.LogWithCount(5)
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitIntegration tests integration scenarios
func (ts *GitTestSuite) TestGitIntegration() {
	ts.Run("git operations with environment variables", func() {
		// Set environment variables
		originalVersion := os.Getenv("version")
		originalMessage := os.Getenv("message")
		defer func() {
			require.NoError(ts.T(), os.Setenv("version", originalVersion))
			require.NoError(ts.T(), os.Setenv("message", originalMessage))
		}()
		require.NoError(ts.T(), os.Setenv("version", "2.0.0"))
		require.NoError(ts.T(), os.Setenv("message", "Version 2.0.0 release"))

		// Mock git operations using environment variables
		ts.env.Runner.On("RunCmd", "git", []string{"add", "-A"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"commit", "-m", "Version 2.0.0 release"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"tag", "-a", "v2.0.0", "-m", "Pending full release..."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"push", "origin", "v2.0.0"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"fetch", "--tags", "-f"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				if err := ts.git.Commit(); err != nil {
					return err
				}
				return ts.git.Tag()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGitTestSuite runs the test suite
func TestGitTestSuite(t *testing.T) {
	suite.Run(t, new(GitTestSuite))
}
