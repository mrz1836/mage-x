package mage

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors (err113 compliance).
var (
	errPushNetwork     = errors.New("fatal: unable to access 'https://github.com/…': Could not read from remote repository")
	errPushRejected    = errors.New("fatal: failed to push some refs")
	errRevListNoTag    = errors.New("fatal: ambiguous argument: unknown revision")
	errLsRemoteNetwork = errors.New("fatal: unable to access remote: Could not resolve host: github.com")
)

// Two distinct 40-char commit SHAs used throughout the race tests.
const (
	testCommitA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testCommitB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testTagObjA = "1111111111111111111111111111111111111111"
	testTag     = "v1.27.0"
)

// scriptedResult is a single canned response for a command invocation.
type scriptedResult struct {
	output string
	err    error
}

// scriptedRunner is a concurrency-safe CommandRunner that returns a *sequence*
// of canned results per command key. Successive calls to the same command
// consume the next result; once the sequence is exhausted the final result
// repeats. This lets tests model state that changes between invocations — e.g.
// a tag that is absent on the pre-check but present after a failed push (the
// concurrent-push race).
type scriptedRunner struct {
	mu       sync.Mutex
	commands []string
	seq      map[string][]scriptedResult
	calls    map[string]int
}

func newScriptedRunner() *scriptedRunner {
	return &scriptedRunner{
		seq:   make(map[string][]scriptedResult),
		calls: make(map[string]int),
	}
}

func (r *scriptedRunner) set(key string, results ...scriptedResult) {
	r.seq[key] = results
}

// out registers a single successful output for key.
func (r *scriptedRunner) out(key, output string) {
	r.set(key, scriptedResult{output: output})
}

// fail registers a single failing result for key.
func (r *scriptedRunner) fail(key string, err error) {
	r.set(key, scriptedResult{err: err})
}

func (r *scriptedRunner) next(key string) scriptedResult {
	r.calls[key]++
	results := r.seq[key]
	if len(results) == 0 {
		return scriptedResult{}
	}
	idx := r.calls[key] - 1
	if idx >= len(results) {
		idx = len(results) - 1 // repeat the final scripted result
	}
	return results[idx]
}

func (r *scriptedRunner) RunCmd(name string, args ...string) error {
	key := strings.Join(append([]string{name}, args...), " ")
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, key)
	return r.next(key).err
}

func (r *scriptedRunner) RunCmdOutput(name string, args ...string) (string, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands = append(r.commands, key)
	res := r.next(key)
	return res.output, res.err
}

func (r *scriptedRunner) callCount(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[key]
}

// pushed reports whether the tag push was attempted.
func (r *scriptedRunner) pushed() bool {
	return r.callCount(keyPush) > 0
}

// lsRemoteLine renders a single `git ls-remote` line: "<sha>\t<ref>".
func lsRemoteLine(sha, ref string) string {
	return sha + "\t" + ref
}

// annotatedLsRemote renders `git ls-remote --tags origin <testTag>` output for
// an annotated tag pointing at commit: the tag-object line plus the peeled
// ("^{}") commit line. Scoped to the tag under test (testTag); the tag-object
// hash is irrelevant to the peel logic, so a fixed placeholder is used.
func annotatedLsRemote(commit string) string {
	ref := "refs/tags/" + testTag
	return lsRemoteLine(testTagObjA, ref) + "\n" +
		lsRemoteLine(commit, ref+"^{}")
}

// lightweightLsRemote renders `git ls-remote --tags origin <tag>` output for a
// lightweight tag: a single line pointing directly at the commit.
func lightweightLsRemote(tag, commit string) string {
	return lsRemoteLine(commit, "refs/tags/"+tag)
}

// command-key helpers for the tag under test.
const (
	keyRemoteV = "git remote -v"
	keyLsTags  = "git ls-remote --tags origin " + testTag
	keyRevList = "git rev-list -n 1 " + testTag
	keyPush    = "git push origin " + testTag
)

// withValidRemote wires up the commands validateGitRemote("origin") needs so a
// test can reach the tag/push logic in handlePushTag.
func withValidRemote(r *scriptedRunner) {
	r.out(keyRemoteV, "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
	// "git ls-remote --exit-code origin HEAD" (a RunCmd) is left unset → nil
	// error → the remote is considered reachable.
}

func TestHandlePushTag(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(originalRunner)) })

	t.Run("happy path pushes when remote has no tag", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		// keyLsTags unset → "" → remote has no such tag.
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag))
		assert.True(t, r.pushed(), "expected the tag to be pushed")
		assert.Zero(t, r.callCount(keyRevList), "local tag resolution is unnecessary when remote lacks the tag")
	})

	t.Run("idempotent when remote already has tag at same commit (annotated)", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		r.out(keyLsTags, annotatedLsRemote(testCommitA))
		r.out(keyRevList, testCommitA)
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag))
		assert.False(t, r.pushed(), "must NOT push when the exact tag is already on the remote")
	})

	t.Run("idempotent when remote already has tag at same commit (lightweight)", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		r.out(keyLsTags, lightweightLsRemote(testTag, testCommitA))
		r.out(keyRevList, testCommitA)
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag))
		assert.False(t, r.pushed(), "lightweight tag at same commit is still a no-op")
	})

	t.Run("conflict when remote tag points at a different commit", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		r.out(keyLsTags, annotatedLsRemote(testCommitB))
		r.out(keyRevList, testCommitA)
		require.NoError(t, SetRunner(r))

		err := handlePushTag(testTag)
		require.ErrorIs(t, err, errTagAlreadyExistsOnRemote)
		assert.False(t, r.pushed(), "must NOT push over a tag that points elsewhere")
	})

	t.Run("race: push fails but remote ends up at same commit -> success", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		// Pre-check sees no tag; after the (rejected) push the tag is present at
		// our commit — a concurrent run won the race and published the same tag.
		r.set(keyLsTags,
			scriptedResult{output: ""},                             // pre-check: absent
			scriptedResult{output: annotatedLsRemote(testCommitA)}, // post-failure: present
		)
		r.out(keyRevList, testCommitA)
		r.fail(keyPush, errPushRejected)
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag), "a concurrent push of the same release must not fail the bump")
		assert.True(t, r.pushed(), "the push should have been attempted")
		assert.GreaterOrEqual(t, r.callCount(keyLsTags), 2, "remote must be re-checked after a failed push")
	})

	t.Run("push fails and remote has a different commit -> error", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		r.set(keyLsTags,
			scriptedResult{output: ""},                             // pre-check: absent
			scriptedResult{output: annotatedLsRemote(testCommitB)}, // post-failure: someone else's commit
		)
		r.out(keyRevList, testCommitA)
		r.fail(keyPush, errPushRejected)
		require.NoError(t, SetRunner(r))

		err := handlePushTag(testTag)
		require.Error(t, err, "a real conflict after a failed push must surface")
		assert.ErrorIs(t, err, errPushRejected)
	})

	t.Run("genuine network failure on push surfaces the error", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		// keyLsTags stays empty on every call (remote never gets the tag).
		r.fail(keyPush, errPushNetwork)
		require.NoError(t, SetRunner(r))

		err := handlePushTag(testTag)
		require.Error(t, err)
		assert.ErrorIs(t, err, errPushNetwork, "network failures must fail loudly, never be masked")
	})

	t.Run("remote has tag but local tag cannot be resolved -> falls through to push", func(t *testing.T) {
		r := newScriptedRunner()
		withValidRemote(r)
		r.out(keyLsTags, annotatedLsRemote(testCommitA))
		r.fail(keyRevList, errRevListNoTag) // local resolution fails
		// push succeeds (unset → nil)
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag))
		assert.True(t, r.pushed(), "an unresolved local tag must not be treated as a match; push should proceed")
	})

	t.Run("remote validation failure skips push without erroring", func(t *testing.T) {
		r := newScriptedRunner()
		// keyRemoteV unset → validateGitRemote cannot find 'origin' → non-fatal skip.
		require.NoError(t, SetRunner(r))

		require.NoError(t, handlePushTag(testTag), "remote validation issues are non-fatal by design")
		assert.False(t, r.pushed(), "no push should be attempted when the remote is not usable")
	})
}

func TestRemoteTagCommit(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(originalRunner)) })

	t.Run("annotated tag returns peeled commit", func(t *testing.T) {
		r := newScriptedRunner()
		r.out(keyLsTags, annotatedLsRemote(testCommitA))
		require.NoError(t, SetRunner(r))

		commit, exists := remoteTagCommit(testTag)
		assert.True(t, exists)
		assert.Equal(t, testCommitA, commit, "annotated tags must peel to the underlying commit, not the tag object")
	})

	t.Run("lightweight tag returns direct commit", func(t *testing.T) {
		r := newScriptedRunner()
		r.out(keyLsTags, lightweightLsRemote(testTag, testCommitA))
		require.NoError(t, SetRunner(r))

		commit, exists := remoteTagCommit(testTag)
		assert.True(t, exists)
		assert.Equal(t, testCommitA, commit)
	})

	t.Run("no tag on remote", func(t *testing.T) {
		r := newScriptedRunner()
		r.out(keyLsTags, "")
		require.NoError(t, SetRunner(r))

		commit, exists := remoteTagCommit(testTag)
		assert.False(t, exists)
		assert.Empty(t, commit)
	})

	t.Run("ls-remote network error is treated as not found", func(t *testing.T) {
		r := newScriptedRunner()
		r.fail(keyLsTags, errLsRemoteNetwork)
		require.NoError(t, SetRunner(r))

		commit, exists := remoteTagCommit(testTag)
		assert.False(t, exists, "a lookup error must be reported as not-found so the push path decides")
		assert.Empty(t, commit)
	})

	t.Run("unrelated refs only", func(t *testing.T) {
		r := newScriptedRunner()
		r.out(keyLsTags, "deadbeef\trefs/tags/some-other-tag")
		require.NoError(t, SetRunner(r))

		_, exists := remoteTagCommit(testTag)
		assert.False(t, exists, "only an exact refs/tags/<tag> match counts")
	})

	t.Run("prefix collision does not match", func(t *testing.T) {
		r := newScriptedRunner()
		// Asking for v1.27.0 but the remote only has v1.27.00 — must not match.
		r.out(keyLsTags, lightweightLsRemote("v1.27.00", testCommitA))
		require.NoError(t, SetRunner(r))

		_, exists := remoteTagCommit(testTag)
		assert.False(t, exists)
	})

	t.Run("malformed lines are ignored, valid peeled line wins", func(t *testing.T) {
		r := newScriptedRunner()
		out := "garbage-without-tab\n" +
			"   \n" +
			annotatedLsRemote(testCommitA)
		r.out(keyLsTags, out)
		require.NoError(t, SetRunner(r))

		commit, exists := remoteTagCommit(testTag)
		assert.True(t, exists)
		assert.Equal(t, testCommitA, commit)
	})
}

func TestLocalTagCommit(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(originalRunner)) })

	t.Run("returns trimmed commit", func(t *testing.T) {
		r := newScriptedRunner()
		r.out(keyRevList, "  "+testCommitA+"\n")
		require.NoError(t, SetRunner(r))

		assert.Equal(t, testCommitA, localTagCommit(testTag))
	})

	t.Run("returns empty on error", func(t *testing.T) {
		r := newScriptedRunner()
		r.fail(keyRevList, errRevListNoTag)
		require.NoError(t, SetRunner(r))

		assert.Empty(t, localTagCommit(testTag))
	})
}

func TestShortSHA(t *testing.T) {
	assert.Equal(t, "aaaaaaaaaaaa", shortSHA(testCommitA), "long SHAs trim to 12 chars")
	assert.Equal(t, "abc", shortSHA("abc"), "short values are returned untouched")
	assert.Equal(t, "123456789012", shortSHA("123456789012"), "exactly 12 chars is untouched")
	assert.Empty(t, shortSHA(""))
}

// TestHandlePushTagConcurrent hammers handlePushTag from many goroutines with a
// shared, mutex-guarded runner where the tag is already published at our commit.
// Every caller must take the idempotent path and return nil, and the whole thing
// must be free of data races (run with -race).
func TestHandlePushTagConcurrent(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() { require.NoError(t, SetRunner(originalRunner)) })

	r := newScriptedRunner()
	withValidRemote(r)
	r.out(keyLsTags, annotatedLsRemote(testCommitA))
	r.out(keyRevList, testCommitA)
	require.NoError(t, SetRunner(r))

	const workers = 32
	var wg sync.WaitGroup
	errs := make([]error, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = handlePushTag(testTag)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		require.NoErrorf(t, err, "worker %d should have taken the idempotent path", i)
	}
	assert.False(t, r.pushed(), "no goroutine should push when the tag is already published at our commit")
}
