package mage

import (
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
)

// defaultGitHubAPIBaseURL is the public GitHub REST API endpoint.
const defaultGitHubAPIBaseURL = "https://api.github.com"

// githubAPIBaseURL is the base URL used for GitHub REST API calls. It is a
// package-level variable so tests can point release lookups at a local server
// instead of reaching api.github.com over the network.
//
//nolint:gochecknoglobals // seam required to keep release lookups off the network in tests
var githubAPIBaseURL = defaultGitHubAPIBaseURL

// gitHubAPIBaseURL returns the base URL for GitHub REST API requests, with no
// trailing slash. GITHUB_API_URL takes precedence so GitHub Enterprise hosts
// (and test harnesses driving the compiled binary) can redirect the calls.
func gitHubAPIBaseURL() string {
	if configured := env.Get("GITHUB_API_URL"); configured != "" {
		return strings.TrimSuffix(configured, "/")
	}
	return strings.TrimSuffix(githubAPIBaseURL, "/")
}

// useGitHubCLI reports whether release lookups may shell out to the gh CLI. It
// is skipped when the API endpoint has been redirected, because gh talks to its
// own configured host and would ignore the override.
func useGitHubCLI() bool {
	return gitHubAPIBaseURL() == defaultGitHubAPIBaseURL && commandExists("gh")
}
