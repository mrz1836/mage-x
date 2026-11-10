package mage

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkglog "github.com/mrz1836/mage-x/pkg/log"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// bumpJSONFields are the fields of the line version:bump prints with json.
// Scripts read these names, so they must not change.
var bumpJSONFields = []string{
	"from", "to", "bump", "expect", "dryRun", "tag", "commit", "pushed",
	"startBranch", "branch", "warnings", "error",
}

// mockedTagCommit is the commit the mocked v1.3.28 tag points at.
const mockedTagCommit = "0123456789abcdef0123456789abcdef01234567"

// TestVersionBumpJSON runs the magex binary, as a script would, in clones of a
// local bare origin, and reads the outcome from the line json adds to stdout.
// Nothing here reaches the network.
func TestVersionBumpJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping magex binary test in short mode")
	}

	magex := buildMagexForJSONTest(t)
	env := jsonBumpEnv(t)

	t.Run("a patch release from another branch reaches origin", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push=true", "expect=1.0.1", "json=true")
		require.Equalf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Equal(t, "patch", got.Bump)
		assert.Equal(t, "v1.0.1", got.Expect, "expect is reported as vX.Y.Z")
		assert.False(t, got.DryRun)
		assert.Equal(t, "v1.0.1", got.Tag)
		assert.Equal(t, fx.master, got.Commit, "the tag is on master, where the bump ran")
		assert.NotEqual(t, fx.feature, got.Commit, "the tag is not on the branch the clone started on")
		assert.True(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Empty(t, got.Error)

		assert.Equal(t, fx.master, jsonBumpGit(t, env, fx.origin, "rev-parse", "v1.0.1^{commit}"))
		assert.Equal(t, "feature-x", jsonBumpGit(t, env, fx.work, "branch", "--show-current"))
	})

	t.Run("a refused push leaves the tag local and says so", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)
		hook := filepath.Join(fx.work, ".git", "hooks", "pre-push")
		//nolint:gosec // G306: a git hook must be executable
		require.NoError(t, os.WriteFile(hook, []byte("#!/bin/sh\necho 'pushes are refused here' >&2\nexit 1\n"), 0o700))

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push", "json")
		require.NotEqualf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Empty(t, got.Expect)
		assert.Equal(t, "v1.0.1", got.Tag)
		assert.Equal(t, fx.master, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Contains(t, got.Error, errTagNotPushed.Error())
		assert.Contains(t, got.Error, "git push origin v1.0.1", "the error says how to finish the push")

		assert.Equal(t, fx.master, jsonBumpGit(t, env, fx.work, "rev-parse", "v1.0.1^{commit}"), "the tag stays in the clone")
		assert.Empty(t, jsonBumpGit(t, env, fx.origin, "tag", "--list", "v1.0.1"), "the tag never reached origin")
	})

	t.Run("an unreachable origin leaves the tag local and says so", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)
		jsonBumpGit(t, env, fx.work, "checkout", "master")
		jsonBumpGit(t, env, fx.work, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "missing.git"))

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work, "bump=patch", "push", "json")
		require.NotEqualf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Equal(t, "v1.0.1", got.Tag, "the tag is still reported")
		assert.Equal(t, fx.master, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "master", got.StartBranch)
		assert.Equal(t, "master", got.Branch)
		assert.Contains(t, got.Error, errTagNotPushed.Error())
		assert.True(t, slices.ContainsFunc(got.Warnings, func(w string) bool {
			return strings.HasPrefix(w, "Git remote validation failed: ")
		}), "warnings: %q", got.Warnings)

		assert.Equal(t, fx.master, jsonBumpGit(t, env, fx.work, "rev-parse", "v1.0.1^{commit}"), "the tag stays in the clone")
	})

	t.Run("an expect mismatch tags nothing", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push", "expect=v1.0.5", "json")
		require.NotEqualf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Equal(t, "v1.0.5", got.Expect)
		assert.Empty(t, got.Tag)
		assert.Empty(t, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Contains(t, got.Error, "v1.0.5", "the error names the expected version")
		assert.Contains(t, got.Error, "v1.0.1", "the error names the version the bump worked out")
		assert.Contains(t, got.Error, errExpectedVersionMismatch.Error())

		assert.Empty(t, jsonBumpGit(t, env, fx.work, "tag", "--list", "v1.0.1", "v1.0.5"), "nothing was tagged")
		assert.Empty(t, jsonBumpGit(t, env, fx.origin, "tag", "--list", "v1.0.1", "v1.0.5"), "nothing was pushed")
	})

	t.Run("uncommitted changes tag nothing", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)
		require.NoError(t, os.WriteFile(filepath.Join(fx.work, "test.txt"), []byte("not committed\n"), 0o600))

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push", "json")
		require.NotEqualf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Empty(t, got.From)
		assert.Empty(t, got.To)
		assert.Equal(t, "patch", got.Bump)
		assert.Empty(t, got.Tag)
		assert.Empty(t, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Equal(t, errVersionUncommittedChanges.Error(), got.Error)

		assert.Empty(t, jsonBumpGit(t, env, fx.work, "tag", "--list", "v1.0.1"))
	})

	t.Run("dry-run reports the version it would tag", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work, "bump=patch", "dry-run", "json")
		require.Equalf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Equal(t, "patch", got.Bump)
		assert.Empty(t, got.Expect)
		assert.True(t, got.DryRun)
		assert.Empty(t, got.Tag)
		assert.Empty(t, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Empty(t, got.Error)
		assert.Contains(t, got.Warnings,
			"No branch parameter specified - performing version bump on current branch: 'feature-x'",
			"warnings are the printed text without its emoji")

		assert.Empty(t, jsonBumpGit(t, env, fx.work, "tag", "--list", "v1.0.1"))
	})

	t.Run("without json no JSON line is printed", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work, "bump=patch", "dry-run")
		require.Equalf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		for _, line := range strings.Split(stdout, "\n") {
			assert.Falsef(t, strings.HasPrefix(line, "{"), "unexpected JSON line: %s", line)
		}
	})
}

// TestVersionBumpJSONInProcess covers what the magex runs above can't set up,
// with a mock runner in place of git.
func TestVersionBumpJSONInProcess(t *testing.T) {
	const origin = "origin\t/srv/origin.git (fetch)\norigin\t/srv/origin.git (push)"
	pushTests := []struct {
		name       string
		remotes    string // what git remote -v lists
		pushErr    error  // what git push returns
		wantPushed bool
	}{
		{
			name:       "a push that reaches origin is reported as pushed",
			remotes:    origin,
			wantPushed: true,
		},
		{
			name:    "a refused push is reported with the tag it left behind",
			remotes: origin,
			pushErr: errRemoteRejected,
		},
		{
			name:    "a push skipped for want of an origin is reported with the tag it left behind",
			remotes: "",
		},
	}
	for _, tt := range pushTests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newJSONBumpMock()
			mock.SetOutput("git remote -v", tt.remotes)
			if tt.pushErr != nil {
				mock.SetError("git push origin v1.3.28", tt.pushErr)
			}
			installJSONBumpRunner(t, mock)

			out, err := runBumpInProcess(t, "bump=patch", "push", "json")

			got := requireBumpJSON(t, out, "")
			assert.Equal(t, "v1.3.28", got.Tag)
			assert.Equal(t, mockedTagCommit, got.Commit)
			assert.Equal(t, tt.wantPushed, got.Pushed)
			if tt.wantPushed {
				require.NoError(t, err)
				assert.Empty(t, got.Error)
				return
			}
			require.ErrorIs(t, err, errTagNotPushed)
			assert.Equal(t, err.Error(), got.Error)
		})
	}

	t.Run("an expect mismatch in dry-run is a warning", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())

		out, err := runBumpInProcess(t, "bump=patch", "dry-run", "expect=v1.3.30", "json")
		require.NoError(t, err)

		got := requireBumpJSON(t, out, "")
		assert.Equal(t, "v1.3.30", got.Expect)
		assert.Equal(t, "v1.3.28", got.To)
		assert.Empty(t, got.Tag)
		assert.Empty(t, got.Error)
		assert.True(t, slices.ContainsFunc(got.Warnings, func(w string) bool {
			return strings.HasPrefix(w, "expected v1.3.30, but the next patch is v1.3.28")
		}), "warnings: %q", got.Warnings)
	})

	t.Run("an invalid expect is reported", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())

		out, err := runBumpInProcess(t, "bump=patch", "expect=latest", "json")
		require.ErrorIs(t, err, errInvalidExpectedVersion)

		got := requireBumpJSON(t, out, "")
		assert.Empty(t, got.Expect)
		assert.Empty(t, got.Tag)
		assert.Equal(t, err.Error(), got.Error)
	})

	t.Run("a failed switch back is reported", func(t *testing.T) {
		mock := newJSONBumpMock()
		mock.SetOutput("git branch -a", "  master\n* feature-x\n  remotes/origin/master")
		mock.SetError("git checkout feature-x", errBranchNotFoundLocally)
		mock.SetError("git checkout -b feature-x origin/feature-x", errBranchNotFoundLocally)
		installJSONBumpRunner(t, &branchTrackingRunner{VersionBumpMockRunner: mock, branch: "feature-x"})

		out, err := runBumpInProcess(t, "bump=patch", "branch=master", "json")
		require.NoError(t, err)

		got := requireBumpJSON(t, out, "")
		assert.Equal(t, "v1.3.28", got.Tag)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "master", got.Branch)
		assert.Contains(t, got.Warnings, "The bump started on branch 'feature-x' but ended on branch 'master'")
		assert.Empty(t, got.Error)
	})

	t.Run("a run that starts on a detached HEAD says where it ended", func(t *testing.T) {
		mock := newJSONBumpMock()
		mock.SetOutput("git branch -a", "* (HEAD detached at 0123456)\n  master")
		mock.SetError("git checkout ", errBranchNotFoundLocally)
		mock.SetError("git checkout -b  origin/", errBranchNotFoundLocally)
		installJSONBumpRunner(t, &branchTrackingRunner{VersionBumpMockRunner: mock})

		out, err := runBumpInProcess(t, "bump=patch", "branch=master", "json")
		require.NoError(t, err)

		got := requireBumpJSON(t, out, "")
		assert.Empty(t, got.StartBranch)
		assert.Equal(t, "master", got.Branch)
		assert.Contains(t, got.Warnings, "The bump started on a detached HEAD but ended on branch 'master'")
	})

	t.Run("json can't be combined with module", func(t *testing.T) {
		mock := newJSONBumpMock()
		installJSONBumpRunner(t, mock)

		out, err := runBumpInProcess(t, "module=models", "bump=patch", "json")
		require.ErrorIs(t, err, errBumpJSONWithModule)

		got := requireBumpJSON(t, out, "")
		assert.Equal(t, err.Error(), got.Error)
		assert.Empty(t, got.Tag)
		assert.False(t, mock.HasCommand([]string{"find", ".", "-name", "go.mod", "-type", "f"}),
			"no sub-module was looked for")
	})

	t.Run("an invalid bump type is reported", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())

		out, err := runBumpInProcess(t, "bump=huge", "dry-run", "json")
		require.ErrorIs(t, err, errInvalidBumpType)

		got := requireBumpJSON(t, out, "")
		assert.Empty(t, got.Bump)
		assert.True(t, got.DryRun)
		assert.Equal(t, "master", got.StartBranch)
		assert.Equal(t, err.Error(), got.Error)
	})

	t.Run("json is read from the arguments magex passes a delegated target", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())
		t.Setenv(utils.EnvMageArgsJSON, `["bump=minor","dry-run","json"]`)

		out, err := runBumpInProcess(t)
		require.NoError(t, err)

		got := requireBumpJSON(t, out, "")
		assert.Equal(t, "minor", got.Bump)
		assert.True(t, got.DryRun)
		assert.Equal(t, "v1.3.27", got.From)
		assert.Equal(t, "v1.4.0", got.To)
	})

	t.Run("json=false prints no JSON line and leaves the logger alone", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())
		logger := pkglog.Default()

		out, err := runBumpInProcess(t, "bump=patch", "dry-run", "json=false")
		require.NoError(t, err)

		for _, line := range strings.Split(out, "\n") {
			assert.Falsef(t, strings.HasPrefix(line, "{"), "unexpected JSON line: %s", line)
		}
		assert.Same(t, logger, pkglog.Default())
	})

	t.Run("the logger is restored after a json run", func(t *testing.T) {
		installJSONBumpRunner(t, newJSONBumpMock())
		logger := pkglog.Default()

		_, err := runBumpInProcess(t, "bump=patch", "dry-run", "json")
		require.NoError(t, err)

		assert.Same(t, logger, pkglog.Default())
	})
}

// TestPlainWarning checks that a warning is reported as the text magex printed,
// without its colors or its leading emoji, bullet, or indentation.
func TestPlainWarning(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain text", in: "Multiple tags found on HEAD: v1.0.1, v1.0.0", want: "Multiple tags found on HEAD: v1.0.1, v1.0.0"},
		{name: "warning sign", in: "⚠️  Skipped versions (exist from other branches): v1.0.1", want: "Skipped versions (exist from other branches): v1.0.1"},
		{name: "light bulb", in: "💡 To push manually: git push origin v1.0.1", want: "To push manually: git push origin v1.0.1"},
		{name: "indented bullet", in: "   • Check repository access permissions", want: "Check repository access permissions"},
		{name: "indented detail", in: "   Current version: v1.0.0", want: "Current version: v1.0.0"},
		{name: "colors", in: "\x1b[33mremote\x1b[0m rejected", want: "remote rejected"},
		{name: "emoji then colors", in: "⚠️ \x1b[1;31mfailed\x1b[0m", want: "failed"},
		{name: "two-byte escape", in: "\x1bcreset", want: "reset"},
		{name: "unterminated escape", in: "text\x1b[31", want: "text"},
		{name: "trailing escape", in: "text\x1b", want: "text"},
		{name: "trailing newline", in: "Available remotes:\n", want: "Available remotes:"},
		{name: "quote kept", in: "'force' was given", want: "'force' was given"},
		{name: "number kept", in: "2 commits since v1.0.0", want: "2 commits since v1.0.0"},
		{name: "emoji inside kept", in: "Tag ✅ created", want: "Tag ✅ created"},
		{name: "empty", in: "", want: ""},
		{name: "only decoration", in: "⚠️   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, plainWarning(tt.in))
		})
	}
}

// buildMagexForJSONTest builds the magex binary the tests run.
func buildMagexForJSONTest(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "magex")
	//nolint:gosec // G204: the test chooses every argument
	cmd := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/magex")
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "go build: %s", out)
	return binary
}

// jsonBumpEnv returns an environment for git and magex that reads none of the
// developer's git configuration (a fresh HOME and global config, and no system
// config), names no repository, passes magex no arguments, and turns magex's
// update check off, so nothing reaches the network.
func jsonBumpEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	gitconfig := filepath.Join(home, ".gitconfig")
	require.NoError(t, os.WriteFile(gitconfig, []byte(
		"[user]\n\tname = Test User\n\temail = test@example.com\n"+
			"[init]\n\tdefaultBranch = master\n"+
			"[commit]\n\tgpgsign = false\n"+
			"[tag]\n\tgpgsign = false\n",
	), 0o600))

	env := make([]string, 0, len(os.Environ())+7)
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		switch {
		case strings.HasPrefix(name, "GIT_"),
			name == "HOME", name == "XDG_CONFIG_HOME",
			name == utils.EnvMageArgs, name == utils.EnvMageArgsJSON:
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"HOME="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_CONFIG_GLOBAL="+gitconfig,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
		legacyDisableUpdateCheckEnv+"=true",
	)
}

// jsonBumpFixture is a clone of a local bare origin. master has v1.0.0 and one
// commit after it, both on origin; the clone is on feature-x, one commit ahead
// of master, so a bump with branch=master tags a commit other than the one the
// clone started on.
type jsonBumpFixture struct {
	origin  string // the bare repository the clone's origin points at
	work    string // the clone version:bump runs in
	master  string // the commit master points at
	feature string // the commit feature-x points at
}

// newJSONBumpFixture creates a jsonBumpFixture in a new temporary directory.
func newJSONBumpFixture(t *testing.T, env []string) jsonBumpFixture {
	t.Helper()

	root := t.TempDir()
	fx := jsonBumpFixture{
		origin: filepath.Join(root, "origin.git"),
		work:   filepath.Join(root, "work"),
	}

	jsonBumpGit(t, env, root, "init", "--bare", fx.origin)
	jsonBumpGit(t, env, root, "clone", fx.origin, fx.work)

	commit := func(content, message string) {
		require.NoError(t, os.WriteFile(filepath.Join(fx.work, "test.txt"), []byte(content), 0o600))
		jsonBumpGit(t, env, fx.work, "add", "test.txt")
		jsonBumpGit(t, env, fx.work, "commit", "-m", message)
	}

	commit("first release\n", "first release")
	jsonBumpGit(t, env, fx.work, "tag", "-a", "v1.0.0", "-m", "Release v1.0.0")
	commit("a fix\n", "a fix after v1.0.0")
	jsonBumpGit(t, env, fx.work, "push", "--set-upstream", "origin", "master")
	jsonBumpGit(t, env, fx.work, "push", "origin", "v1.0.0")
	fx.master = jsonBumpGit(t, env, fx.work, "rev-parse", "HEAD")

	jsonBumpGit(t, env, fx.work, "checkout", "-b", "feature-x")
	commit("feature work\n", "feature work")
	fx.feature = jsonBumpGit(t, env, fx.work, "rev-parse", "HEAD")

	return fx
}

// jsonBumpGit runs git with args in dir under env and returns its trimmed
// output, failing the test if git fails.
func jsonBumpGit(t *testing.T, env []string, dir string, args ...string) string {
	t.Helper()

	//nolint:gosec // G204: the test chooses every argument
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

// runMagexBump runs magex version:bump with args in dir under env, and returns
// its stdout, its stderr, and its exit status.
func runMagexBump(t *testing.T, magex string, env []string, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	//nolint:gosec // G204: the test chooses every argument
	cmd := exec.CommandContext(t.Context(), magex, append([]string{"version:bump"}, args...)...)
	cmd.Dir = dir
	cmd.Env = env
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return out.String(), errOut.String(), exitErr.ExitCode()
	}
	require.NoError(t, err)
	return out.String(), errOut.String(), 0
}

// requireBumpJSON returns the outcome version:bump printed with json: one JSON
// object on the last line of stdout, with exactly the documented fields. It
// also checks that no line of stderr starts with "{", so the outcome is still
// the last line that does when the two streams are merged.
func requireBumpJSON(t *testing.T, stdout, stderr string) bumpResult {
	t.Helper()

	require.Truef(t, strings.HasSuffix(stdout, "\n"), "stdout does not end with a whole line:\n%s", stdout)
	lines := strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")
	last := lines[len(lines)-1]
	require.Truef(t, strings.HasPrefix(last, "{"), "the last line of stdout is not the JSON outcome:\n%s", stdout)

	var fields map[string]json.RawMessage
	require.NoErrorf(t, json.Unmarshal([]byte(last), &fields), "the last line of stdout is not JSON: %s", last)
	require.ElementsMatch(t, bumpJSONFields, slices.Collect(maps.Keys(fields)), "the JSON fields changed: %s", last)
	require.Truef(t, bytes.HasPrefix(fields["warnings"], []byte("[")), "warnings is not a list: %s", last)

	var result bumpResult
	require.NoError(t, json.Unmarshal([]byte(last), &result))
	for _, warning := range result.Warnings {
		require.NotEmpty(t, warning, "an empty warning")
		assert.NotContains(t, warning, "\x1b", "a warning keeps an ANSI code")
		first, _ := utf8.DecodeRuneInString(warning)
		assert.Falsef(t, unicode.IsSpace(first) || unicode.Is(unicode.So, first), "a warning keeps its decoration: %q", warning)
	}

	for _, line := range strings.Split(stderr, "\n") {
		assert.Falsef(t, strings.HasPrefix(line, "{"), "a line of stderr could be taken for the outcome: %s", line)
	}
	return result
}

// newJSONBumpMock returns a mock runner for a clean clone on master whose
// highest tag, v1.3.27, is three commits back, so a patch bump tags v1.3.28 on
// mockedTagCommit.
func newJSONBumpMock() *VersionBumpMockRunner {
	mock := NewVersionBumpMockRunner()
	mock.SetOutput("git branch --show-current", "master")
	mock.SetOutput("git tag --sort=-version:refname", "v1.3.27")
	mock.SetOutput("git describe --tags --abbrev=0", "v1.3.27")
	mock.SetOutput("git rev-list --count v1.3.27..HEAD", "3")
	mock.SetOutput("git rev-list -n 1 v1.3.28", mockedTagCommit)
	return mock
}

// installJSONBumpRunner makes runner the command runner for the rest of the test.
func installJSONBumpRunner(t *testing.T, runner CommandRunner) {
	t.Helper()

	original := GetRunner()
	require.NoError(t, SetRunner(runner))
	t.Cleanup(func() { require.NoError(t, SetRunner(original)) })
}

// branchTrackingRunner is a VersionBumpMockRunner that follows git checkout:
// git branch --show-current names the branch the last successful checkout of
// a branch moved to, or "" for a detached HEAD before any.
type branchTrackingRunner struct {
	*VersionBumpMockRunner

	branch string
}

// RunCmd runs the mock command and follows a successful git checkout <branch>.
func (r *branchTrackingRunner) RunCmd(name string, args ...string) error {
	err := r.VersionBumpMockRunner.RunCmd(name, args...)
	if err == nil && name == "git" && len(args) == 2 && args[0] == "checkout" {
		r.branch = args[1]
	}
	return err
}

// RunCmdOutput answers git branch --show-current with the tracked branch, and
// any other command from the mock.
func (r *branchTrackingRunner) RunCmdOutput(name string, args ...string) (string, error) {
	if name == "git" && strings.Join(args, " ") == "branch --show-current" {
		return r.branch, nil
	}
	return r.VersionBumpMockRunner.RunCmdOutput(name, args...)
}

// runBumpInProcess runs Version.Bump with args, and returns what it printed and
// the error it returned.
func runBumpInProcess(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var err error
	out := captureBumpOutput(t, func() { err = Version{}.Bump(args...) })
	return out, err
}

// captureBumpOutput runs fn with os.Stdout and the CLI logger writing to one
// pipe, as they share stdout in a magex run, and returns what fn wrote in the
// order it wrote it.
func captureBumpOutput(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = w
	logger := pkglog.Default()
	logger.SetOutput(w)
	defer func() {
		os.Stdout = stdout
		logger.SetOutput(stdout)
	}()

	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r) //nolint:errcheck // the test reads what it can
		close(done)
	}()

	fn()

	require.NoError(t, w.Close())
	<-done
	require.NoError(t, r.Close())
	return buf.String()
}
