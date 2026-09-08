package mage

import (
	"testing"

	"github.com/mrz1836/mage-x/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// releaseCheckCurlCmd builds the exact command string the mock runner keys on
// for a publishedReleaseExists lookup with no Authorization header.
func releaseCheckCurlCmd(ownerRepo, tag string) string {
	return "curl -sSL -w \\n%{http_code} " +
		"-H Accept: application/vnd.github+json " +
		"-H X-GitHub-Api-Version: 2022-11-28 " +
		"https://api.github.com/repos/" + ownerRepo + "/releases/tags/" + tag
}

func TestSplitBodyStatus(t *testing.T) {
	t.Run("body and status", func(t *testing.T) {
		body, status := splitBodyStatus(`{"draft":false}` + "\n200")
		assert.Equal(t, `{"draft":false}`, body)
		assert.Equal(t, "200", status)
	})

	t.Run("trailing whitespace trimmed", func(t *testing.T) {
		body, status := splitBodyStatus(`{"draft":false}` + "\n404\n")
		assert.Equal(t, `{"draft":false}`, body)
		assert.Equal(t, "404", status)
	})

	t.Run("no newline returns body only", func(t *testing.T) {
		body, status := splitBodyStatus("just-a-body")
		assert.Equal(t, "just-a-body", body)
		assert.Empty(t, status)
	})
}

func TestOwnerRepoPattern(t *testing.T) {
	cases := map[string]string{
		"git@github.com:mrz1836/mage-x.git":       "mrz1836/mage-x",
		"https://github.com/mrz1836/mage-x.git":   "mrz1836/mage-x",
		"https://github.com/mrz1836/mage-x":       "mrz1836/mage-x",
		"git@github.com:mrz1836/mage-x":           "mrz1836/mage-x",
		"ssh://git@github.com/mrz1836/mage-x.git": "mrz1836/mage-x",
	}
	for remote, want := range cases {
		t.Run(remote, func(t *testing.T) {
			m := ownerRepoPattern.FindStringSubmatch(remote)
			require.Len(t, m, 2, "expected a match for %q", remote)
			assert.Equal(t, want, m[1])
		})
	}
}

func TestCurrentOwnerRepo(t *testing.T) {
	t.Run("prefers GITHUB_REPOSITORY", func(t *testing.T) {
		t.Setenv("GITHUB_REPOSITORY", "mrz1836/mage-x")

		got, err := currentOwnerRepo()
		require.NoError(t, err)
		assert.Equal(t, "mrz1836/mage-x", got)
	})

	t.Run("falls back to origin remote", func(t *testing.T) {
		t.Setenv("GITHUB_REPOSITORY", "")

		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := NewBmadMockRunner()
		mock.SetOutput("git config --get remote.origin.url", "git@github.com:mrz1836/mage-x.git", nil)
		require.NoError(t, SetRunner(mock))

		got, err := currentOwnerRepo()
		require.NoError(t, err)
		assert.Equal(t, "mrz1836/mage-x", got)
	})

	t.Run("unrecognized remote errors", func(t *testing.T) {
		t.Setenv("GITHUB_REPOSITORY", "")

		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := NewBmadMockRunner()
		mock.SetOutput("git config --get remote.origin.url", "not-a-remote", nil)
		require.NoError(t, SetRunner(mock))

		_, err := currentOwnerRepo()
		require.ErrorIs(t, err, errOwnerRepoUnresolved)
	})
}

func TestPublishedReleaseExists(t *testing.T) {
	if !utils.CommandExists("curl") {
		t.Skip("curl not installed in test environment")
	}

	const ownerRepo = "mrz1836/mage-x"
	const tag = "v1.27.0"
	cmd := releaseCheckCurlCmd(ownerRepo, tag)

	setup := func(t *testing.T) *BmadMockRunner {
		t.Helper()
		// Force getEnvGitHubToken to yield no token so curl is invoked without
		// an Authorization header (keeping the command string deterministic).
		t.Setenv("GH_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("GITHUB_API_URL", "")

		mock := NewBmadMockRunner()
		mock.SetOutput("gh auth token", "", errReleaseCheckToolMissing)
		return mock
	}

	t.Run("published non-draft release returns true", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := setup(t)
		mock.SetOutput(cmd, `{"tag_name":"v1.27.0","draft":false}`+"\n200", nil)
		require.NoError(t, SetRunner(mock))

		got, err := publishedReleaseExists(ownerRepo, tag)
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("draft release returns false", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := setup(t)
		mock.SetOutput(cmd, `{"tag_name":"v1.27.0","draft":true}`+"\n200", nil)
		require.NoError(t, SetRunner(mock))

		got, err := publishedReleaseExists(ownerRepo, tag)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("missing release (404) returns false", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := setup(t)
		mock.SetOutput(cmd, `{"message":"Not Found"}`+"\n404", nil)
		require.NoError(t, SetRunner(mock))

		got, err := publishedReleaseExists(ownerRepo, tag)
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("unexpected status errors", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { require.NoError(t, SetRunner(originalRunner)) }()

		mock := setup(t)
		mock.SetOutput(cmd, `{"message":"boom"}`+"\n500", nil)
		require.NoError(t, SetRunner(mock))

		_, err := publishedReleaseExists(ownerRepo, tag)
		require.ErrorIs(t, err, errReleaseCheckStatus)
	})
}

func TestReleaseAlreadyPublished_UnresolvedOwnerRepoReturnsFalse(t *testing.T) {
	t.Setenv("GITHUB_REPOSITORY", "")

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	mock := NewBmadMockRunner()
	mock.SetOutput("git config --get remote.origin.url", "not-a-remote", nil)
	require.NoError(t, SetRunner(mock))

	// Unresolvable owner/repo must never skip the release.
	assert.False(t, releaseAlreadyPublished("v1.27.0"))
}
