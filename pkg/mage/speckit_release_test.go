package mage

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/utils"
)

var errReleaseTestFailure = errors.New("release lookup failed")

// findCommandWithPrefix returns the first command in the recorded list that
// starts with prefix, failing the test if none is found. Used by mock-based
// tests when the code path also shells out to incidental commands like
// `gh auth token` that we don't want to entangle with the assertion.
func findCommandWithPrefix(t *testing.T, commands []string, prefix string) string {
	t.Helper()
	for _, c := range commands {
		if strings.HasPrefix(c, prefix) {
			return c
		}
	}
	t.Fatalf("no command found with prefix %q in %v", prefix, commands)
	return ""
}

func TestValidateSpeckitTag(t *testing.T) {
	cases := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{"plain release", "v0.8.1", false},
		{"two-digit minor", "v0.10.0", false},
		{"prerelease rc", "v1.0.0-rc.1", false},
		{"prerelease beta", "v2.3.4-beta.5", false},
		{"build metadata", "v1.0.0+build.7", false},
		{"missing v prefix", "0.8.1", true},
		{"only major minor", "v1.0", true},
		{"branch name main", "main", true},
		{"branch name latest", "latest", true},
		{"empty", "", true},
		{"garbage", "not-a-tag", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSpeckitTag(tc.tag)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSpeckitFromSpec(t *testing.T) {
	got := speckitFromSpec("https://github.com/github/spec-kit.git", "v0.8.1")
	assert.Equal(t, "git+https://github.com/github/spec-kit.git@v0.8.1", got)
}

func TestResolveSpeckitIntegrationFallback(t *testing.T) {
	cases := []struct {
		name string
		cfg  SpeckitConfig
		want string
	}{
		{"prefers integration", SpeckitConfig{Integration: "copilot", AIProvider: "claude"}, "copilot"},
		{"falls back to ai_provider", SpeckitConfig{AIProvider: "gemini"}, "gemini"},
		{"defaults when both empty", SpeckitConfig{}, DefaultSpeckitIntegration},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveSpeckitIntegration(&Config{Speckit: tc.cfg})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveSpeckitOwnerRepo(t *testing.T) {
	assert.Equal(t, DefaultSpeckitOwnerRepo, resolveSpeckitOwnerRepo(&Config{Speckit: SpeckitConfig{}}))
	assert.Equal(t, "fork/spec-kit", resolveSpeckitOwnerRepo(&Config{Speckit: SpeckitConfig{OwnerRepo: "fork/spec-kit"}}))
}

func TestResolveSpeckitGitURL(t *testing.T) {
	assert.Equal(t, DefaultSpeckitGitURL, resolveSpeckitGitURL(&Config{Speckit: SpeckitConfig{}}))
	assert.Equal(t, "https://example.test/spec-kit.git", resolveSpeckitGitURL(&Config{Speckit: SpeckitConfig{GitURL: "https://example.test/spec-kit.git"}}))
}

func TestResolveSpeckitCLIName(t *testing.T) {
	assert.Equal(t, DefaultSpeckitCLIName, resolveSpeckitCLIName(&Config{Speckit: SpeckitConfig{}}))
	assert.Equal(t, "custom-cli", resolveSpeckitCLIName(&Config{Speckit: SpeckitConfig{CLIName: "custom-cli"}}))
}

func TestResolveTagViaGH_HappyPath(t *testing.T) {
	if !utils.CommandExists("gh") {
		t.Skip("gh CLI not installed in test environment")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	mock := NewBmadMockRunner()
	mock.SetOutput("gh api repos/github/spec-kit/releases/latest --jq .tag_name", "v0.8.1\n", nil)
	require.NoError(t, SetRunner(mock))

	tag, err := resolveTagViaGH("github/spec-kit")
	require.NoError(t, err)
	assert.Equal(t, "v0.8.1", tag)
}

func TestResolveTagViaGH_EmptyResponse(t *testing.T) {
	if !utils.CommandExists("gh") {
		t.Skip("gh CLI not installed in test environment")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	mock := NewBmadMockRunner()
	mock.SetOutput("gh api repos/github/spec-kit/releases/latest --jq .tag_name", "null\n", nil)
	require.NoError(t, SetRunner(mock))

	_, err := resolveTagViaGH("github/spec-kit")
	require.ErrorIs(t, err, errSpeckitTagEmpty)
}

func TestResolveTagViaCurl_ParsesTagFromJSON(t *testing.T) {
	if !utils.CommandExists("curl") {
		t.Skip("curl not installed in test environment")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	jsonBody := `{"url":"https://api.github.com/repos/github/spec-kit/releases/12345","tag_name":"v0.8.1","name":"v0.8.1","draft":false}`

	t.Run("without token", func(t *testing.T) {
		t.Setenv("GH_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "")

		mock := NewBmadMockRunner()
		// Force the gh-auth fallback path to return no token so curl is
		// invoked without an Authorization header.
		mock.SetOutput("gh auth token", "", errReleaseTestFailure)
		expectedCmd := "curl -fsSL -H Accept: application/vnd.github+json -H X-GitHub-Api-Version: 2022-11-28 https://api.github.com/repos/github/spec-kit/releases/latest"
		mock.SetOutput(expectedCmd, jsonBody, nil)
		require.NoError(t, SetRunner(mock))

		tag, err := resolveTagViaCurl("github/spec-kit")
		require.NoError(t, err)
		assert.Equal(t, "v0.8.1", tag)

		curlCmd := findCommandWithPrefix(t, mock.GetCommands(), "curl ")
		assert.NotContains(t, curlCmd, "Authorization", "no Authorization header should be sent without a token")
	})

	t.Run("with GITHUB_TOKEN sends Authorization header", func(t *testing.T) {
		t.Setenv("GH_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "ghp_testtoken123")

		mock := NewBmadMockRunner()
		expectedCmd := "curl -fsSL -H Accept: application/vnd.github+json -H X-GitHub-Api-Version: 2022-11-28 -H Authorization: Bearer ghp_testtoken123 https://api.github.com/repos/github/spec-kit/releases/latest"
		mock.SetOutput(expectedCmd, jsonBody, nil)
		require.NoError(t, SetRunner(mock))

		tag, err := resolveTagViaCurl("github/spec-kit")
		require.NoError(t, err)
		assert.Equal(t, "v0.8.1", tag)

		curlCmd := findCommandWithPrefix(t, mock.GetCommands(), "curl ")
		assert.Contains(t, curlCmd, "Authorization: Bearer ghp_testtoken123")
	})

	t.Run("missing tag_name", func(t *testing.T) {
		t.Setenv("GH_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "")

		mock := NewBmadMockRunner()
		mock.SetOutput("gh auth token", "", errReleaseTestFailure)
		expectedCmd := "curl -fsSL -H Accept: application/vnd.github+json -H X-GitHub-Api-Version: 2022-11-28 https://api.github.com/repos/github/spec-kit/releases/latest"
		mock.SetOutput(expectedCmd, `{"message":"Not Found"}`, nil)
		require.NoError(t, SetRunner(mock))

		_, err := resolveTagViaCurl("github/spec-kit")
		require.ErrorIs(t, err, errSpeckitTagResolveFailed)
	})
}

func TestResolveTagViaGitLsRemote_PicksFirstSemverTag(t *testing.T) {
	if !utils.CommandExists("git") {
		t.Skip("git not installed in test environment")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	// Real git ls-remote --sort=-v:refname output format: "<sha>\trefs/tags/<tag>".
	// First semver-matching line wins; non-conforming refs (e.g. release-candidate
	// branches accidentally tagged) are skipped by the regex.
	output := "abc123\trefs/tags/some-feature-branch\n" +
		"def456\trefs/tags/v0.8.1\n" +
		"789abc\trefs/tags/v0.8.0\n"

	mock := NewBmadMockRunner()
	expectedCmd := "git ls-remote --tags --refs --sort=-v:refname https://github.com/github/spec-kit.git"
	mock.SetOutput(expectedCmd, output, nil)
	require.NoError(t, SetRunner(mock))

	tag, err := resolveTagViaGitLsRemote("https://github.com/github/spec-kit.git")
	require.NoError(t, err)
	assert.Equal(t, "v0.8.1", tag)
}

func TestResolveTagViaGitLsRemote_NoTagsFound(t *testing.T) {
	if !utils.CommandExists("git") {
		t.Skip("git not installed in test environment")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	mock := NewBmadMockRunner()
	expectedCmd := "git ls-remote --tags --refs --sort=-v:refname https://github.com/github/spec-kit.git"
	mock.SetOutput(expectedCmd, "abc\trefs/tags/not-a-version\n", nil)
	require.NoError(t, SetRunner(mock))

	_, err := resolveTagViaGitLsRemote("https://github.com/github/spec-kit.git")
	require.ErrorIs(t, err, errSpeckitTagResolveFailed)
}

func TestResolveLatestSpeckitTag_FallsThroughOnError(t *testing.T) {
	if !utils.CommandExists("gh") || !utils.CommandExists("curl") {
		t.Skip("requires gh and curl to exercise fallthrough")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	mock := NewBmadMockRunner()
	// Make gh fail so the resolver falls through to curl.
	mock.SetOutput("gh api repos/github/spec-kit/releases/latest --jq .tag_name", "", errReleaseTestFailure)
	curlCmd := "curl -fsSL -H Accept: application/vnd.github+json -H X-GitHub-Api-Version: 2022-11-28 https://api.github.com/repos/github/spec-kit/releases/latest"
	mock.SetOutput(curlCmd, `{"tag_name":"v0.8.1"}`, nil)
	require.NoError(t, SetRunner(mock))

	tag, source, err := resolveLatestSpeckitTag("github/spec-kit", "https://github.com/github/spec-kit.git")
	require.NoError(t, err)
	assert.Equal(t, "v0.8.1", tag)
	assert.Equal(t, speckitTagSourceCurl, source)
}

func TestResolveLatestSpeckitTag_AllFail(t *testing.T) {
	if !utils.CommandExists("gh") || !utils.CommandExists("curl") || !utils.CommandExists("git") {
		t.Skip("requires gh, curl, and git to exercise total failure")
	}

	originalRunner := GetRunner()
	defer func() { require.NoError(t, SetRunner(originalRunner)) }()

	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	mock := NewBmadMockRunner()
	mock.SetOutput("gh api repos/github/spec-kit/releases/latest --jq .tag_name", "", errReleaseTestFailure)
	curlCmd := "curl -fsSL -H Accept: application/vnd.github+json -H X-GitHub-Api-Version: 2022-11-28 https://api.github.com/repos/github/spec-kit/releases/latest"
	mock.SetOutput(curlCmd, "", errReleaseTestFailure)
	mock.SetOutput("git ls-remote --tags --refs --sort=-v:refname https://github.com/github/spec-kit.git", "", errReleaseTestFailure)
	require.NoError(t, SetRunner(mock))

	_, _, err := resolveLatestSpeckitTag("github/spec-kit", "https://github.com/github/spec-kit.git")
	require.Error(t, err)
	require.ErrorIs(t, err, errSpeckitTagResolveFailed)
}
