package mage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// isolateBumpGit points git at a fresh HOME with its own global config and no
// system config, and drops any repository the environment names (a git hook
// sets GIT_DIR and GIT_INDEX_FILE), so the test's git reads none of the
// developer's configuration and touches no other repository.
func isolateBumpGit(t *testing.T) {
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

// bumpGit runs git with args in dir and returns its trimmed output, failing
// the test if git fails.
func bumpGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	//nolint:gosec // G204: Subprocess launched with variable - safe in test context
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoErrorf(t, err, "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

// bumpRace is a release in progress: a local bare origin whose master carries
// v1.4.2 and one commit after it, the clone that runs version:bump, and a
// second clone that can release another version in the meantime.
type bumpRace struct {
	origin string // the bare repository both clones push to
	work   string // the clone version:bump runs in
	other  string // the clone that tags a release meanwhile
}

// newBumpRace sets up a bumpRace in a new temporary directory.
func newBumpRace(t *testing.T) bumpRace {
	t.Helper()

	root := t.TempDir()
	race := bumpRace{
		origin: filepath.Join(root, "origin.git"),
		work:   filepath.Join(root, "work"),
		other:  filepath.Join(root, "other"),
	}
	seed := filepath.Join(root, "seed")

	bumpGit(t, root, "init", "--bare", "--initial-branch=master", race.origin)
	bumpGit(t, root, "init", "--initial-branch=master", seed)
	bumpGit(t, seed, "commit", "--allow-empty", "-m", "first release")
	bumpGit(t, seed, "tag", "-a", "v1.4.2", "-m", "Release v1.4.2")
	bumpGit(t, seed, "commit", "--allow-empty", "-m", "change to release")
	bumpGit(t, seed, "remote", "add", "origin", race.origin)
	bumpGit(t, seed, "push", "origin", "master", "v1.4.2")

	bumpGit(t, root, "clone", race.origin, race.work)
	bumpGit(t, root, "clone", race.origin, race.other)
	return race
}

// releaseMeanwhile has the other clone tag a new commit as tag and push it to
// origin with one more commit on top, the way another release lands on master
// after the work clone last fetched.
func (r bumpRace) releaseMeanwhile(t *testing.T, tag string) {
	t.Helper()

	bumpGit(t, r.other, "commit", "--allow-empty", "-m", "change released as "+tag)
	bumpGit(t, r.other, "tag", "-a", tag, "-m", "Release "+tag)
	bumpGit(t, r.other, "commit", "--allow-empty", "-m", "change after "+tag)
	bumpGit(t, r.other, "push", "origin", "master", tag)
}

// hasTag reports whether the repository in dir has tag.
func hasTag(t *testing.T, dir, tag string) bool {
	t.Helper()
	return bumpGit(t, dir, "tag", "--list", tag) == tag
}

// useRealGitRunner makes version:bump run real git commands for the rest of
// the test.
func useRealGitRunner(t *testing.T) {
	t.Helper()

	original := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(original)) })
	require.NoError(t, SetRunner(NewSecureCommandRunner()))
}

// TestVersionBumpExpectRealRemote runs version:bump branch=master push in a
// clone whose bare origin gained a release after the clone last fetched. The
// bump fetches and pulls first, so without expect it releases the version
// after the one pushed meanwhile; with expect it creates the expected version
// or nothing at all. A returned error is what makes magex exit non-zero.
// Nothing here reaches the network.
func TestVersionBumpExpectRealRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	tests := []struct {
		name      string
		meanwhile string   // the release origin gained after the clone last fetched, "" for none
		args      []string // version:bump's parameters
		wantErr   string   // the whole error text, "" when the bump succeeds
		wantTag   string   // the tag the bump creates locally and pushes to origin
		absentTag string   // a tag that must exist neither locally nor on origin
	}{
		{
			name:      "without expect the bump follows the release made meanwhile",
			meanwhile: "v1.4.3",
			args:      []string{"bump=patch", "branch=master", "push=true"},
			wantTag:   "v1.4.4",
		},
		{
			name:      "patch bump stops when another version was tagged meanwhile",
			meanwhile: "v1.4.3",
			args:      []string{"bump=patch", "branch=master", "push=true", "expect=v1.4.3"},
			wantErr:   "expected v1.4.3, but the next patch is v1.4.4: another version was tagged meanwhile; nothing was tagged",
			absentTag: "v1.4.4",
		},
		{
			name:      "minor bump stops when another version was tagged meanwhile",
			meanwhile: "v1.5.0",
			args:      []string{"bump=minor", "branch=master", "push=true", "expect=v1.5.0"},
			wantErr:   "expected v1.5.0, but the next minor is v1.6.0: another version was tagged meanwhile; nothing was tagged",
			absentTag: "v1.6.0",
		},
		{
			name:      "confirmed major bump stops when another major version was tagged meanwhile",
			meanwhile: "v2.0.0",
			args:      []string{"bump=major", "major-confirm", "branch=master", "push", "expect=v2.0.0"},
			wantErr:   "expected v2.0.0, but the next major is v3.0.0: another version was tagged meanwhile; nothing was tagged",
			absentTag: "v3.0.0",
		},
		{
			name:      "bump stops when the next version is lower than expected",
			args:      []string{"bump=patch", "branch=master", "push=true", "expect=v1.4.5"},
			wantErr:   "expected v1.4.5, but the next patch is v1.4.3; nothing was tagged",
			absentTag: "v1.4.3",
		},
		{
			name:      "matching expect releases as before",
			meanwhile: "v1.4.3",
			args:      []string{"bump=patch", "branch=master", "push=true", "expect=v1.4.4"},
			wantTag:   "v1.4.4",
		},
		{
			name:    "matching expect without the leading v releases as before",
			args:    []string{"bump=patch", "branch=master", "push=true", "expect=1.4.3"},
			wantTag: "v1.4.3",
		},
		{
			name:    "matching expect on a confirmed major bump releases as before",
			args:    []string{"bump=major", "major-confirm", "branch=master", "push", "expect=v2.0.0"},
			wantTag: "v2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			race := newBumpRace(t)
			if tt.meanwhile != "" {
				race.releaseMeanwhile(t, tt.meanwhile)
			}
			t.Chdir(race.work)

			err := Version{}.Bump(tt.args...)

			if tt.wantErr != "" {
				require.ErrorIs(t, err, errExpectedVersionMismatch)
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				require.NoError(t, err)
			}

			if tt.wantTag != "" {
				require.True(t, hasTag(t, race.work, tt.wantTag), "the bump should create %s locally", tt.wantTag)
				assert.Equal(t, bumpGit(t, race.work, "rev-parse", "HEAD"),
					bumpGit(t, race.work, "rev-list", "-n", "1", tt.wantTag), "%s should tag the pulled HEAD", tt.wantTag)
				assert.True(t, hasTag(t, race.origin, tt.wantTag), "the bump should push %s to origin", tt.wantTag)
			}
			if tt.absentTag != "" {
				assert.False(t, hasTag(t, race.work, tt.absentTag), "nothing may tag %s locally", tt.absentTag)
				assert.False(t, hasTag(t, race.origin, tt.absentTag), "nothing may push %s to origin", tt.absentTag)
			}
		})
	}
}

// TestVersionBumpExpectMalformedChangesNothing gives version:bump a malformed
// expect in a clone whose origin gained a release meanwhile. The bump must
// fail before it fetches, pulls, switches branches, or tags anything.
func TestVersionBumpExpectMalformedChangesNothing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	race := newBumpRace(t)
	race.releaseMeanwhile(t, "v1.4.3")
	t.Chdir(race.work)

	head := bumpGit(t, race.work, "rev-parse", "HEAD")
	refs := bumpGit(t, race.work, "for-each-ref")
	originRefs := bumpGit(t, race.origin, "for-each-ref")

	err := Version{}.Bump("bump=patch", "branch=master", "push=true", "expect=v1.4")
	require.ErrorIs(t, err, errInvalidExpectedVersion)
	assert.Equal(t, `expect must be a release version such as v1.4.3 or 1.4.3, got "v1.4"`, err.Error())

	assert.Equal(t, head, bumpGit(t, race.work, "rev-parse", "HEAD"), "the bump must not pull")
	assert.Equal(t, refs, bumpGit(t, race.work, "for-each-ref"), "the bump must not fetch or tag")
	assert.False(t, hasTag(t, race.work, "v1.4.3"), "the bump must not fetch the release made meanwhile")
	assert.Equal(t, originRefs, bumpGit(t, race.origin, "for-each-ref"), "the bump must not push")
}

// TestVersionBumpExpectDryRun checks expect in dry-run, which does not fetch:
// it compares expect with the next version from the tags the clone already
// has, reports a mismatch as a warning, and still succeeds without tagging.
func TestVersionBumpExpectDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real git repository test in short mode")
	}

	useRealGitRunner(t)
	isolateBumpGit(t)

	tests := []struct {
		name         string
		args         []string
		wantLog      string // what the dry run reports about expect
		wantMismatch bool
	}{
		{
			name:    "matching expect previews the bump",
			args:    []string{"bump=patch", "branch=master", "push", "dry-run", "expect=v1.4.3"},
			wantLog: "v1.4.3 matches expect",
		},
		{
			name:         "mismatching expect is reported as a warning",
			args:         []string{"bump=patch", "branch=master", "push", "dry-run", "expect=v1.5.0"},
			wantLog:      "expected v1.5.0, but the next patch is v1.4.3 (would fail in normal mode)",
			wantMismatch: true,
		},
		{
			name:         "mismatching expect on a major bump is reported as a warning",
			args:         []string{"bump=major", "branch=master", "dry-run", "expect=1.5.0"},
			wantLog:      "expected v1.5.0, but the next major is v2.0.0 (would fail in normal mode)",
			wantMismatch: true,
		},
		{
			name:         "expect already tagged locally is reported as a warning",
			args:         []string{"bump=patch", "dry-run", "expect=v1.4.2"},
			wantLog:      "expected v1.4.2, but the next patch is v1.4.3: another version was tagged meanwhile (would fail in normal mode)",
			wantMismatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			race := newBumpRace(t)
			t.Chdir(race.work)
			refs := bumpGit(t, race.work, "for-each-ref")

			var err error
			log := captureCLILog(t, func() { err = Version{}.Bump(tt.args...) })

			require.NoError(t, err)
			assert.Contains(t, log, tt.wantLog)
			assert.Equal(t, tt.wantMismatch, strings.Contains(log, "would fail in normal mode"))
			assert.Contains(t, log, "DRY-RUN completed")
			assert.Equal(t, refs, bumpGit(t, race.work, "for-each-ref"), "dry-run must not change any ref")
		})
	}
}

// TestVersionBumpExpectMalformedRunsNoCommand checks that a malformed expect,
// or expect combined with module, fails before version:bump runs any command.
func TestVersionBumpExpectMalformedRunsNoCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantIs   error
		wantText string
	}{
		{
			name:     "two components",
			args:     []string{"bump=patch", "branch=master", "push", "expect=1.4"},
			wantIs:   errInvalidExpectedVersion,
			wantText: `expect must be a release version such as v1.4.3 or 1.4.3, got "1.4"`,
		},
		{
			name:     "empty value",
			args:     []string{"bump=patch", "branch=master", "push", "expect="},
			wantIs:   errInvalidExpectedVersion,
			wantText: `expect must be a release version such as v1.4.3 or 1.4.3, got ""`,
		},
		{
			name:     "bare flag",
			args:     []string{"bump=major", "major-confirm", "push", "expect"},
			wantIs:   errInvalidExpectedVersion,
			wantText: `expect must be a release version such as v1.4.3 or 1.4.3, got "true"`,
		},
		{
			name:     "with module",
			args:     []string{"module=*", "bump=patch", "push", "expect=v1.4.3"},
			wantIs:   errExpectWithModule,
			wantText: "expect can't be combined with module: it pins the root module's next version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := useRecordingRunner(t)

			err := Version{}.Bump(tt.args...)

			require.ErrorIs(t, err, tt.wantIs)
			assert.Equal(t, tt.wantText, err.Error())
			assert.Empty(t, recorder.calls, "no command may run before expect is checked")
		})
	}
}

// TestParseBumpConfigExpect checks which expect values version:bump accepts
// and the vX.Y.Z form it compares them in.
func TestParseBumpConfigExpect(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   string // the pinned version; "" when expect is not given
		wantIs error  // the error a rejected value returns
	}{
		{name: "absent leaves the version unpinned", args: []string{"bump=patch", "push"}},
		{name: "with the leading v", args: []string{"expect=v1.4.3"}, want: "v1.4.3"},
		{name: "without the leading v", args: []string{"expect=1.4.3"}, want: "v1.4.3"},
		{name: "zero components", args: []string{"expect=v0.0.0"}, want: "v0.0.0"},
		{name: "multi-digit components", args: []string{"expect=10.20.300"}, want: "v10.20.300"},
		{name: "surrounding spaces are trimmed", args: []string{"expect= v2.0.0 "}, want: "v2.0.0"},
		{name: "with the root module", args: []string{"module=", "expect=v1.4.3"}, want: "v1.4.3"},
		{name: "empty value", args: []string{"expect="}, wantIs: errInvalidExpectedVersion},
		{name: "bare flag", args: []string{"expect"}, wantIs: errInvalidExpectedVersion},
		{name: "two components", args: []string{"expect=v1.4"}, wantIs: errInvalidExpectedVersion},
		{name: "four components", args: []string{"expect=1.4.3.0"}, wantIs: errInvalidExpectedVersion},
		{name: "pre-release suffix", args: []string{"expect=v1.4.3-rc.1"}, wantIs: errInvalidExpectedVersion},
		{name: "build suffix", args: []string{"expect=v1.4.3+build.5"}, wantIs: errInvalidExpectedVersion},
		{name: "uppercase V", args: []string{"expect=V1.4.3"}, wantIs: errInvalidExpectedVersion},
		{name: "doubled v", args: []string{"expect=vv1.4.3"}, wantIs: errInvalidExpectedVersion},
		{name: "leading zero", args: []string{"expect=v1.04.3"}, wantIs: errInvalidExpectedVersion},
		{name: "negative component", args: []string{"expect=v1.-4.3"}, wantIs: errInvalidExpectedVersion},
		{name: "not a number", args: []string{"expect=v1.4.x"}, wantIs: errInvalidExpectedVersion},
		{name: "a sub-module tag", args: []string{"expect=models/v1.4.3"}, wantIs: errInvalidExpectedVersion},
		{name: "with a sub-module", args: []string{"module=models", "expect=v1.4.3"}, wantIs: errExpectWithModule},
		{name: "with all sub-modules", args: []string{"module=all", "expect=v1.4.3"}, wantIs: errExpectWithModule},
		{name: "with the root and all sub-modules", args: []string{"module=*", "expect=v1.4.3"}, wantIs: errExpectWithModule},
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
			assert.Equal(t, tt.want, cfg.expect)
		})
	}
}

// TestParseBumpConfigExpectFromMageArgs checks that expect is read from the
// arguments magex hands a delegated mage target, like the other parameters.
func TestParseBumpConfigExpectFromMageArgs(t *testing.T) {
	t.Setenv(utils.EnvMageArgs, "")
	t.Setenv(utils.EnvMageArgsJSON, `["bump=minor","branch=master","push","expect=1.5.0"]`)

	cfg, err := parseBumpConfig(nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "minor", cfg.bumpType)
	assert.Equal(t, "master", cfg.targetBranch)
	assert.Equal(t, "v1.5.0", cfg.expect)
}
