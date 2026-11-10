package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errGitRefused stands in for a git command that fails.
var errGitRefused = errors.New("git refused")

// The version the ff-only tests' clone releases next: newBumpRace tags v1.4.2
// on origin's first commit.
const ffOnlyNextTag = "v1.4.3"

// The errors ff-only fails with when master has commits origin/master doesn't
// have: two (ahead), one, or two while origin/master has one master doesn't
// (diverged).
const (
	ffOnlyAheadErr = "ff-only: branch 'master' has 2 commits that origin/master doesn't have, so it can't be fast-forwarded; nothing was tagged\n" +
		"to release them, push them first (git push origin master); " +
		"to drop them, reset the branch to origin in a terminal (git checkout master && git reset --hard origin/master)"
	ffOnlyAheadOneErr = "ff-only: branch 'master' has 1 commit that origin/master doesn't have, so it can't be fast-forwarded; nothing was tagged\n" +
		"to release it, push it first (git push origin master); " +
		"to drop it, reset the branch to origin in a terminal (git checkout master && git reset --hard origin/master)"
	ffOnlyDivergedErr = "ff-only: branch 'master' has diverged from origin/master: it has 2 commits that origin/master doesn't have, " +
		"and origin/master has 1 commit it doesn't; nothing was tagged\n" +
		"to release them, rebase them onto origin/master and push them first " +
		"(git checkout master && git pull --rebase origin master && git push origin master); " +
		"to drop them, reset the branch to origin in a terminal (git checkout master && git reset --hard origin/master)"
)

// commitNewFile commits a new file named name in dir, so commits made in
// different clones never conflict when a rebase replays them.
func commitNewFile(t *testing.T, dir, name string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o600))
	bumpGit(t, dir, "add", name)
	bumpGit(t, dir, "commit", "-m", "add "+name)
}

// ffOnlyScenario is how a clone's master stands against origin's when
// version:bump runs.
type ffOnlyScenario struct {
	start      string // the branch the clone is on: "master", or "feature" to make branch=master switch
	local      int    // commits on the clone's master that origin doesn't have
	meanwhile  int    // commits origin's master gained that the clone hasn't fetched
	noUpstream bool   // the clone's master tracks no upstream
}

// setUp builds the scenario in a new bumpRace and changes into its clone.
func (s ffOnlyScenario) setUp(t *testing.T) bumpRace {
	t.Helper()

	race := newBumpRace(t)
	for i := range s.local {
		commitNewFile(t, race.work, fmt.Sprintf("local-%d", i))
	}
	for i := range s.meanwhile {
		commitNewFile(t, race.other, fmt.Sprintf("merged-%d", i))
	}
	if s.meanwhile > 0 {
		bumpGit(t, race.other, "push", "origin", "master")
	}
	if s.noUpstream {
		bumpGit(t, race.work, "branch", "--unset-upstream", "master")
	}
	if s.start != "master" {
		bumpGit(t, race.work, "checkout", "-b", s.start)
	}
	t.Chdir(race.work)
	return race
}

// TestVersionBumpFFOnlyRealRemote runs version:bump with ff-only in real
// clones of a local bare origin, with an isolated HOME and git config, and
// nothing reaches the network. A branch behind origin is fast-forwarded and
// released; a level one is released as is; a branch with commits origin
// doesn't have (ahead or diverged) fails before tagging, with no tag locally
// or on origin and its commits left as they were. The clone always ends on
// the branch it started on.
func TestVersionBumpFFOnlyRealRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	tests := []struct {
		name     string
		scenario ffOnlyScenario
		args     []string
		wantErr  string // the whole error text, "" when the bump releases
	}{
		{
			name:     "ahead on the target branch fails before tagging",
			scenario: ffOnlyScenario{start: "master", local: 2},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyAheadErr,
		},
		{
			name:     "ahead by one commit fails before tagging",
			scenario: ffOnlyScenario{start: "master", local: 1},
			args:     []string{"bump=patch", "branch=master", "push", "ff-only=true"},
			wantErr:  ffOnlyAheadOneErr,
		},
		{
			name:     "ahead after switching branches fails and switches back",
			scenario: ffOnlyScenario{start: "feature", local: 2},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyAheadErr,
		},
		{
			name:     "diverged fails before tagging without rebasing",
			scenario: ffOnlyScenario{start: "master", local: 2, meanwhile: 1},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyDivergedErr,
		},
		{
			name:     "diverged after switching branches fails and switches back",
			scenario: ffOnlyScenario{start: "feature", local: 2, meanwhile: 1},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyDivergedErr,
		},
		{
			name:     "ahead with the root and all sub-modules fails before tagging",
			scenario: ffOnlyScenario{start: "master", local: 2},
			args:     []string{"module=*", "bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyAheadErr,
		},
		{
			name:     "ahead without upstream tracking fails before tagging",
			scenario: ffOnlyScenario{start: "master", local: 2, noUpstream: true},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
			wantErr:  ffOnlyAheadErr,
		},
		{
			name:     "behind on the target branch fast-forwards and releases",
			scenario: ffOnlyScenario{start: "master", meanwhile: 1},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "behind after switching branches fast-forwards, releases, and switches back",
			scenario: ffOnlyScenario{start: "feature", meanwhile: 1},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "behind without upstream tracking fast-forwards and releases",
			scenario: ffOnlyScenario{start: "master", meanwhile: 1, noUpstream: true},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "behind without upstream tracking after switching branches fast-forwards and releases",
			scenario: ffOnlyScenario{start: "feature", meanwhile: 1, noUpstream: true},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "level releases",
			scenario: ffOnlyScenario{start: "master"},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "level without upstream tracking releases",
			scenario: ffOnlyScenario{start: "master", noUpstream: true},
			args:     []string{"bump=patch", "branch=master", "push=true", "ff-only"},
		},
		{
			name:     "behind with the root and all sub-modules fast-forwards and releases",
			scenario: ffOnlyScenario{start: "master", meanwhile: 1},
			args:     []string{"module=*", "bump=patch", "branch=master", "push=true", "ff-only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			race := tt.scenario.setUp(t)
			masterBefore := bumpGit(t, race.work, "rev-parse", "refs/heads/master")
			originMaster := bumpGit(t, race.origin, "rev-parse", "refs/heads/master")

			err := Version{}.Bump(tt.args...)

			assert.Equal(t, tt.scenario.start, bumpGit(t, race.work, "branch", "--show-current"),
				"the bump must end on the branch it started on")
			assert.Equal(t, originMaster, bumpGit(t, race.origin, "rev-parse", "refs/heads/master"),
				"the bump must never change origin's branch")

			if tt.wantErr != "" {
				require.ErrorIs(t, err, errBranchHasUnpushedCommits)
				assert.Equal(t, tt.wantErr, err.Error())
				assert.Equal(t, masterBefore, bumpGit(t, race.work, "rev-parse", "refs/heads/master"),
					"ff-only must leave the branch's commits as they were, with no rebase")
				assert.NoDirExists(t, filepath.Join(race.work, ".git", "rebase-merge"), "no rebase may be left in progress")
				assert.NoDirExists(t, filepath.Join(race.work, ".git", "rebase-apply"), "no rebase may be left in progress")
				assert.False(t, hasTag(t, race.work, ffOnlyNextTag), "nothing may tag %s locally", ffOnlyNextTag)
				assert.False(t, hasTag(t, race.origin, ffOnlyNextTag), "nothing may push %s to origin", ffOnlyNextTag)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, originMaster, bumpGit(t, race.work, "rev-parse", "refs/heads/master"),
				"ff-only must bring the branch level with origin's")
			require.True(t, hasTag(t, race.work, ffOnlyNextTag), "the bump should create %s locally", ffOnlyNextTag)
			assert.Equal(t, originMaster, bumpGit(t, race.work, "rev-list", "-n", "1", ffOnlyNextTag),
				"%s must tag the commit origin's branch has", ffOnlyNextTag)
			assert.True(t, hasTag(t, race.origin, ffOnlyNextTag), "the bump should push %s to origin", ffOnlyNextTag)
		})
	}
}

// TestVersionBumpFFOnlyJSON runs the magex binary with ff-only and json, as a
// script would, in clones of a local bare origin (see TestVersionBumpJSON),
// and reads the outcome from the line json adds to stdout. A refusal tags
// nothing and reports the ff-only error; a fast-forwarded release reports as
// any other. Nothing here reaches the network.
func TestVersionBumpFFOnlyJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping magex binary test in short mode")
	}

	magex := buildMagexForJSONTest(t)
	env := jsonBumpEnv(t)

	// pushFromOtherClone commits a new file to origin's master from another
	// clone, which the fixture's clone hasn't fetched.
	pushFromOtherClone := func(t *testing.T, fx jsonBumpFixture) string {
		t.Helper()
		other := filepath.Join(t.TempDir(), "other")
		jsonBumpGit(t, env, filepath.Dir(other), "clone", fx.origin, other)
		require.NoError(t, os.WriteFile(filepath.Join(other, "merged.txt"), []byte("merged\n"), 0o600))
		jsonBumpGit(t, env, other, "add", "merged.txt")
		jsonBumpGit(t, env, other, "commit", "-m", "merged meanwhile")
		jsonBumpGit(t, env, other, "push", "origin", "master")
		return jsonBumpGit(t, env, other, "rev-parse", "HEAD")
	}

	t.Run("a refusal tags nothing and reports the error", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)
		jsonBumpGit(t, env, fx.work, "checkout", "master")
		for _, name := range []string{"local-0.txt", "local-1.txt"} {
			require.NoError(t, os.WriteFile(filepath.Join(fx.work, name), []byte(name+"\n"), 0o600))
			jsonBumpGit(t, env, fx.work, "add", name)
			jsonBumpGit(t, env, fx.work, "commit", "-m", "add "+name)
		}
		localMaster := jsonBumpGit(t, env, fx.work, "rev-parse", "HEAD")
		jsonBumpGit(t, env, fx.work, "checkout", "feature-x")
		originMaster := pushFromOtherClone(t, fx)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push=true", "ff-only", "json=true")
		require.NotEqualf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "patch", got.Bump)
		assert.Empty(t, got.Tag)
		assert.Empty(t, got.Commit)
		assert.False(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Equal(t, ffOnlyDivergedErr, got.Error)

		assert.Equal(t, localMaster, jsonBumpGit(t, env, fx.work, "rev-parse", "refs/heads/master"), "master is not rebased")
		assert.Equal(t, originMaster, jsonBumpGit(t, env, fx.origin, "rev-parse", "refs/heads/master"))
		assert.Empty(t, jsonBumpGit(t, env, fx.work, "tag", "--list", "v1.0.1"), "nothing was tagged")
		assert.Empty(t, jsonBumpGit(t, env, fx.origin, "tag", "--list", "v1.0.1"), "nothing was pushed")
	})

	t.Run("a fast-forwarded release reports as any other", func(t *testing.T) {
		fx := newJSONBumpFixture(t, env)
		originMaster := pushFromOtherClone(t, fx)

		stdout, stderr, code := runMagexBump(t, magex, env, fx.work,
			"bump=patch", "branch=master", "push", "ff-only", "json")
		require.Equalf(t, 0, code, "stdout:\n%s\nstderr:\n%s", stdout, stderr)

		got := requireBumpJSON(t, stdout, stderr)
		assert.Equal(t, "v1.0.0", got.From)
		assert.Equal(t, "v1.0.1", got.To)
		assert.Equal(t, "patch", got.Bump)
		assert.False(t, got.DryRun)
		assert.Equal(t, "v1.0.1", got.Tag)
		assert.Equal(t, originMaster, got.Commit, "the tag is on origin's master, fast-forwarded to")
		assert.True(t, got.Pushed)
		assert.Equal(t, "feature-x", got.StartBranch)
		assert.Equal(t, "feature-x", got.Branch)
		assert.Empty(t, got.Error)

		assert.Equal(t, originMaster, jsonBumpGit(t, env, fx.work, "rev-parse", "refs/heads/master"), "master was fast-forwarded")
		assert.Equal(t, originMaster, jsonBumpGit(t, env, fx.origin, "rev-parse", "v1.0.1^{commit}"))
	})
}

// TestVersionBumpWithoutFFOnlyUnchanged checks that without ff-only, branch=
// still fetches and runs git pull --rebase origin as before: commits origin
// doesn't have are rebased onto origin's branch and the rebased HEAD is
// tagged, and a branch that tracks no upstream fails the pull.
func TestVersionBumpWithoutFFOnlyUnchanged(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	t.Run("diverged is rebased and tagged", func(t *testing.T) {
		race := ffOnlyScenario{start: "master", local: 2, meanwhile: 1}.setUp(t)
		masterBefore := bumpGit(t, race.work, "rev-parse", "refs/heads/master")
		originMaster := bumpGit(t, race.origin, "rev-parse", "refs/heads/master")

		require.NoError(t, Version{}.Bump("bump=patch", "branch=master", "push=true"))

		head := bumpGit(t, race.work, "rev-parse", "HEAD")
		assert.NotEqual(t, masterBefore, head, "without ff-only the local commits are rebased")
		bumpGit(t, race.work, "merge-base", "--is-ancestor", originMaster, head)
		assert.Equal(t, head, bumpGit(t, race.work, "rev-list", "-n", "1", ffOnlyNextTag), "the rebased HEAD is tagged")
		assert.True(t, hasTag(t, race.origin, ffOnlyNextTag), "the tag is pushed")
	})

	t.Run("no upstream tracking fails the pull", func(t *testing.T) {
		race := ffOnlyScenario{start: "master", meanwhile: 1, noUpstream: true}.setUp(t)

		err := Version{}.Bump("bump=patch", "branch=master", "push=true")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to pull latest changes")
		assert.False(t, hasTag(t, race.work, ffOnlyNextTag), "nothing is tagged")
	})
}

// TestVersionBumpFFOnlyDryRun checks that dry-run with ff-only says what it
// would do and changes nothing, even with commits origin doesn't have.
func TestVersionBumpFFOnlyDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	for _, start := range []string{"master", "feature"} {
		t.Run("starting on "+start, func(t *testing.T) {
			race := ffOnlyScenario{start: start, local: 2, meanwhile: 1}.setUp(t)
			refs := bumpGit(t, race.work, "for-each-ref")

			var err error
			log := captureCLILog(t, func() { err = Version{}.Bump("bump=patch", "branch=master", "push", "ff-only", "dry-run") })

			require.NoError(t, err)
			assert.Contains(t, log, "Would fast-forward 'master' to origin/master, failing if it has commits origin/master doesn't have")
			assert.NotContains(t, log, "Would pull latest changes")
			assert.Contains(t, log, "DRY-RUN completed")
			assert.Equal(t, refs, bumpGit(t, race.work, "for-each-ref"), "dry-run must not change any ref")
			assert.Equal(t, start, bumpGit(t, race.work, "branch", "--show-current"))
		})
	}
}

// TestVersionBumpFFOnlyInvalidRunsNoCommand checks that ff-only is refused
// before any command runs where it can't be honored, so a caller who asked
// for it never gets a release without the check.
func TestVersionBumpFFOnlyInvalidRunsNoCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantIs   error
		wantText string
	}{
		{
			name:     "without branch",
			args:     []string{"bump=patch", "push", "ff-only"},
			wantIs:   errFFOnlyWithoutBranch,
			wantText: "ff-only needs branch=: without it, version:bump doesn't fetch or update any branch",
		},
		{
			name:     "with an empty branch",
			args:     []string{"bump=patch", "branch=", "push", "ff-only=true"},
			wantIs:   errFFOnlyWithoutBranch,
			wantText: "ff-only needs branch=: without it, version:bump doesn't fetch or update any branch",
		},
		{
			name:     "with a sub-module",
			args:     []string{"module=models", "bump=patch", "branch=master", "push", "ff-only"},
			wantIs:   errFFOnlyWithModule,
			wantText: "ff-only can't be combined with module=<name> or module=all: they don't switch to or update a branch",
		},
		{
			name:     "with all sub-modules",
			args:     []string{"module=all", "bump=patch", "branch=master", "push", "ff-only"},
			wantIs:   errFFOnlyWithModule,
			wantText: "ff-only can't be combined with module=<name> or module=all: they don't switch to or update a branch",
		},
		{
			name:     "with a value that is neither true nor false",
			args:     []string{"bump=patch", "branch=master", "push", "ff-only=maybe"},
			wantIs:   errInvalidFFOnly,
			wantText: `ff-only must be true or false, got "maybe"`,
		},
		{
			name:     "with an empty value",
			args:     []string{"bump=patch", "branch=master", "push", "ff-only="},
			wantIs:   errInvalidFFOnly,
			wantText: `ff-only must be true or false, got ""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := useRecordingRunner(t)

			err := Version{}.Bump(tt.args...)

			require.ErrorIs(t, err, tt.wantIs)
			assert.Equal(t, tt.wantText, err.Error())
			assert.Empty(t, recorder.calls, "no command may run before ff-only is checked")
		})
	}
}

// TestParseBumpConfigFFOnly checks which ff-only values version:bump accepts.
func TestParseBumpConfigFFOnly(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   bool
		wantIs error // the error a rejected value returns
	}{
		{name: "absent keeps pull --rebase", args: []string{"branch=master"}},
		{name: "bare flag", args: []string{"branch=master", "ff-only"}, want: true},
		{name: "true", args: []string{"branch=master", "ff-only=true"}, want: true},
		{name: "yes", args: []string{"branch=master", "ff-only=yes"}, want: true},
		{name: "1", args: []string{"branch=master", "ff-only=1"}, want: true},
		{name: "false", args: []string{"branch=master", "ff-only=false"}},
		{name: "false without branch", args: []string{"ff-only=false"}},
		{name: "0 without branch", args: []string{"ff-only=0"}},
		{name: "with the root and all sub-modules", args: []string{"module=*", "branch=master", "ff-only"}, want: true},
		{name: "with the root module", args: []string{"module=", "branch=master", "ff-only"}, want: true},
		{name: "without branch", args: []string{"ff-only"}, wantIs: errFFOnlyWithoutBranch},
		{name: "with an empty branch", args: []string{"branch=", "ff-only"}, wantIs: errFFOnlyWithoutBranch},
		{name: "with a sub-module", args: []string{"module=models", "branch=master", "ff-only"}, wantIs: errFFOnlyWithModule},
		{name: "with all sub-modules", args: []string{"module=all", "branch=master", "ff-only"}, wantIs: errFFOnlyWithModule},
		{name: "neither true nor false", args: []string{"branch=master", "ff-only=maybe"}, wantIs: errInvalidFFOnly},
		{name: "empty value", args: []string{"branch=master", "ff-only="}, wantIs: errInvalidFFOnly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseBumpConfig(tt.args)

			if tt.wantIs != nil {
				require.ErrorIs(t, err, tt.wantIs)
				assert.Nil(t, cfg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, tt.want, cfg.ffOnly)
		})
	}
}

// TestFastForwardBranchCommands checks the git commands ff-only runs and how
// it reports a git command that fails or answers unexpectedly. None of them
// is git pull, and nothing runs after a failure.
func TestFastForwardBranchCommands(t *testing.T) {
	const (
		fetch   = "git fetch --tags --force origin +refs/heads/master:refs/remotes/origin/master"
		compare = "git rev-list --left-right --count HEAD...refs/remotes/origin/master"
		merge   = "git merge --ff-only refs/remotes/origin/master"
	)

	tests := []struct {
		name         string
		outputs      map[string]string
		errors       map[string]error
		wantErr      string // a fragment of the error, "" for success
		wantCommands []string
	}{
		{
			name:         "behind fast-forwards",
			outputs:      map[string]string{compare: "0\t3"},
			wantCommands: []string{fetch, compare, merge},
		},
		{
			name:         "level does nothing more",
			outputs:      map[string]string{compare: "0\t0"},
			wantCommands: []string{fetch, compare},
		},
		{
			name:         "ahead fails before merging",
			outputs:      map[string]string{compare: "1\t0"},
			wantErr:      "has 1 commit that origin/master doesn't have",
			wantCommands: []string{fetch, compare},
		},
		{
			name:         "a failed fetch stops",
			errors:       map[string]error{fetch: errGitRefused},
			wantErr:      "failed to fetch branch 'master' from origin: git refused",
			wantCommands: []string{fetch},
		},
		{
			name:         "a failed comparison stops",
			errors:       map[string]error{compare: errGitRefused},
			wantErr:      "failed to compare HEAD with origin/master: git refused",
			wantCommands: []string{fetch, compare},
		},
		{
			name:         "an unexpected comparison stops",
			outputs:      map[string]string{compare: "warning: something"},
			wantErr:      `failed to compare HEAD with origin/master: unexpected git rev-list output "warning: something"`,
			wantCommands: []string{fetch, compare},
		},
		{
			name:         "a non-numeric comparison stops",
			outputs:      map[string]string{compare: "x\t1"},
			wantErr:      `failed to compare HEAD with origin/master: unexpected git rev-list output "x\t1"`,
			wantCommands: []string{fetch, compare},
		},
		{
			name:         "a failed fast-forward stops",
			outputs:      map[string]string{compare: "0\t2"},
			errors:       map[string]error{merge: errGitRefused},
			wantErr:      "failed to fast-forward 'master' to origin/master: git refused",
			wantCommands: []string{fetch, compare, merge},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := GetRunner()
			t.Cleanup(func() { require.NoError(t, SetRunner(original)) })
			runner := NewVersionBumpMockRunner()
			for cmd, out := range tt.outputs {
				runner.SetOutput(cmd, out)
			}
			for cmd, err := range tt.errors {
				runner.SetError(cmd, err)
			}
			require.NoError(t, SetRunner(runner))

			err := fastForwardBranch("master")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			commands := make([]string, 0, len(runner.GetCommands()))
			for _, cmd := range runner.GetCommands() {
				commands = append(commands, strings.Join(cmd, " "))
			}
			assert.Equal(t, tt.wantCommands, commands)
		})
	}
}
