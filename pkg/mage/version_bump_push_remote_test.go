package mage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Origins the real-remote version:bump tests set up.
const (
	originBare        = "bare"         // a local bare repository holding the branch
	originMissingPath = "missing-path" // a remote whose path does not exist
	originNone        = "none"         // no origin remote at all
)

// isolateGit points git at a fresh HOME with its own global config and no
// system config, and drops any repository the environment names (a git hook
// sets GIT_DIR and GIT_INDEX_FILE), so the test's git reads none of the
// developer's configuration and touches no other repository.
func isolateGit(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	globalConfig := filepath.Join(home, ".gitconfig")
	require.NoError(t, os.WriteFile(globalConfig, []byte(
		"[user]\n\tname = Test User\n\temail = test@example.com\n"+
			"[init]\n\tdefaultBranch = master\n"+
			"[commit]\n\tgpgsign = false\n"+
			"[tag]\n\tgpgsign = false\n",
	), 0o600))

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	for _, name := range []string{
		"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_COMMON_DIR",
		"GIT_CONFIG", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_COUNT",
	} {
		t.Setenv(name, "") // restores the original value when the test ends
		require.NoError(t, os.Unsetenv(name))
	}
}

// gitIn runs git with args in dir and returns its trimmed output, failing the
// test if git fails.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()

	//nolint:gosec // G204: Subprocess launched with variable - safe in test context
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

// TestVersionBumpPushRealRemote runs version:bump in real git repositories
// whose origin is a local bare repository, a path that does not exist, or
// absent. A returned error is what makes magex exit non-zero, so a push that
// was requested but never reached origin must return one. Nothing here reaches
// the network.
func TestVersionBumpPushRealRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	originalRunner := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(originalRunner)) })
	require.NoError(t, SetRunner(NewSecureCommandRunner()))

	isolateGit(t)

	const newTag = "v1.0.1"

	tests := []struct {
		name       string
		origin     string
		args       []string
		wantErr    bool
		wantLocal  bool // the new tag exists in the local repository afterwards
		wantRemote bool // the new tag reached the bare origin
	}{
		{
			name:       "push reaches a local bare origin",
			origin:     originBare,
			args:       []string{"bump=patch", "push=true"},
			wantLocal:  true,
			wantRemote: true,
		},
		{
			name:      "push fails when the origin path does not exist",
			origin:    originMissingPath,
			args:      []string{"bump=patch", "push=true"},
			wantErr:   true,
			wantLocal: true,
		},
		{
			name:      "push fails when there is no origin remote",
			origin:    originNone,
			args:      []string{"bump=patch", "push"},
			wantErr:   true,
			wantLocal: true,
		},
		{
			name:      "without push the tag stays local and the bump succeeds",
			origin:    originMissingPath,
			args:      []string{"bump=patch"},
			wantLocal: true,
		},
		{
			name:   "dry-run with push changes nothing",
			origin: originMissingPath,
			args:   []string{"bump=patch", "push=true", "dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := t.TempDir()
			initGit(t, repo)
			createInitialCommitAndTag(t, repo, "v1.0.0")
			createTestCommit(t, repo, "change to release")

			var bare string
			switch tt.origin {
			case originBare:
				bare = filepath.Join(t.TempDir(), "origin.git")
				gitIn(t, repo, "init", "--bare", "--initial-branch=master", bare)
				gitIn(t, repo, "remote", "add", "origin", bare)
				gitIn(t, repo, "push", "origin", "master")
			case originMissingPath:
				gitIn(t, repo, "remote", "add", "origin", filepath.Join(t.TempDir(), "missing.git"))
			case originNone:
			}

			t.Chdir(repo)
			err := Version{}.Bump(tt.args...)

			if tt.wantErr {
				require.Error(t, err, "a requested push that never reached origin must fail the bump")
				require.ErrorIs(t, err, errTagNotPushed)
				assert.Contains(t, err.Error(), "git push origin "+newTag, "the error must say how to finish the push")
				assert.Contains(t, err.Error(), "git tag -d "+newTag, "the error must say how to undo the tag")
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantLocal, gitIn(t, repo, "tag", "--list", newTag) == newTag, "local tag %s", newTag)
			if bare != "" {
				assert.Equal(t, tt.wantRemote, gitIn(t, bare, "tag", "--list", newTag) == newTag, "tag %s on origin", newTag)
			}
		})
	}
}
