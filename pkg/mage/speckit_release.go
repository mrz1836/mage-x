package mage

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	errSpeckitTagResolveFailed = errors.New("failed to resolve latest spec-kit release tag")
	errSpeckitTagInvalid       = errors.New("invalid spec-kit release tag")
	errSpeckitTagEmpty         = errors.New("empty spec-kit release tag")
	errSpeckitToolMissing      = errors.New("required tool not installed")
)

// speckitTagSourceGH, speckitTagSourceCurl, and speckitTagSourceGit name the
// origin of a resolved tag for log/error messaging.
const (
	speckitTagSourceGH   = "gh"
	speckitTagSourceCurl = "curl"
	speckitTagSourceGit  = "git"
)

// speckitTagPattern matches the official spec-kit release tag form (e.g., v0.8.1, v1.0.0-rc.1).
// Refuses bare branch names like "main" or "latest" so a fallback never silently
// installs from a moving target.
var speckitTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+(?:[-+.][\w.]+)?$`)

// validateSpeckitTag returns nil if tag matches the official release form.
func validateSpeckitTag(tag string) error {
	if tag == "" {
		return errSpeckitTagEmpty
	}
	if !speckitTagPattern.MatchString(tag) {
		return fmt.Errorf("%w: %q", errSpeckitTagInvalid, tag)
	}
	return nil
}

// resolveLatestSpeckitTag returns the newest official release tag for the
// given GitHub repo, trying gh first, then authenticated curl, then git
// ls-remote. The second return value names which path succeeded so callers
// can surface it in user-facing logs.
func resolveLatestSpeckitTag(ownerRepo, gitURL string) (string, string, error) {
	type attempt struct {
		source string
		fn     func() (string, error)
	}

	attempts := []attempt{
		{speckitTagSourceGH, func() (string, error) { return resolveTagViaGH(ownerRepo) }},
		{speckitTagSourceCurl, func() (string, error) { return resolveTagViaCurl(ownerRepo) }},
		{speckitTagSourceGit, func() (string, error) { return resolveTagViaGitLsRemote(gitURL) }},
	}

	var failures []string
	for _, a := range attempts {
		tag, err := a.fn()
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", a.source, err))
			continue
		}
		if vErr := validateSpeckitTag(tag); vErr != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", a.source, vErr))
			continue
		}
		return tag, a.source, nil
	}

	return "", "", fmt.Errorf("%w: %s", errSpeckitTagResolveFailed, strings.Join(failures, "; "))
}

// resolveTagViaGH queries the GitHub API via the gh CLI. Returns an error if
// gh is not on PATH so the caller can move on to the next fallback.
func resolveTagViaGH(ownerRepo string) (string, error) {
	if !utils.CommandExists("gh") {
		return "", fmt.Errorf("%w: gh", errSpeckitToolMissing)
	}

	endpoint := fmt.Sprintf("repos/%s/releases/latest", ownerRepo)
	out, err := GetRunner().RunCmdOutput("gh", "api", endpoint, "--jq", ".tag_name")
	if err != nil {
		return "", fmt.Errorf("gh api %s: %w", endpoint, err)
	}

	tag := strings.TrimSpace(out)
	if tag == "" || tag == "null" {
		return "", errSpeckitTagEmpty
	}
	return tag, nil
}

// speckitTagJSONPattern extracts the tag_name field from a GitHub releases
// API response without requiring a JSON dependency.
var speckitTagJSONPattern = regexp.MustCompile(`"tag_name"\s*:\s*"([^"]+)"`)

// resolveTagViaCurl hits the GitHub REST API directly with curl. Reuses any
// available GH/GITHUB token (including one from gh auth) so the fallback
// shares the higher 5000-req/hr authenticated limit instead of the 60-req/hr
// anonymous limit.
func resolveTagViaCurl(ownerRepo string) (string, error) {
	if !utils.CommandExists("curl") {
		return "", fmt.Errorf("%w: curl", errSpeckitToolMissing)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", ownerRepo)
	args := []string{
		"-fsSL",
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
	}

	if token := getEnvGitHubToken(); token != "" {
		args = append(args, "-H", "Authorization: Bearer "+token)
	}
	args = append(args, url)

	out, err := GetRunner().RunCmdOutput("curl", args...)
	if err != nil {
		return "", fmt.Errorf("curl %s: %w", url, err)
	}

	matches := speckitTagJSONPattern.FindStringSubmatch(out)
	if len(matches) < 2 {
		return "", fmt.Errorf("%w: tag_name not found in response", errSpeckitTagResolveFailed)
	}
	return strings.TrimSpace(matches[1]), nil
}

// speckitLsRemoteTagPattern matches a single git ls-remote line ending in a
// release tag, capturing the tag itself.
var speckitLsRemoteTagPattern = regexp.MustCompile(`refs/tags/(v\d+\.\d+\.\d+(?:[-+.][\w.]+)?)$`)

// resolveTagViaGitLsRemote queries the remote git repo for tags. Last-resort
// fallback that works even when the GitHub REST API is unreachable, as long
// as plain HTTPS to github.com is available.
func resolveTagViaGitLsRemote(gitURL string) (string, error) {
	if !utils.CommandExists("git") {
		return "", fmt.Errorf("%w: git", errSpeckitToolMissing)
	}

	out, err := GetRunner().RunCmdOutput("git", "ls-remote", "--tags", "--refs", "--sort=-v:refname", gitURL)
	if err != nil {
		return "", fmt.Errorf("git ls-remote %s: %w", gitURL, err)
	}

	for _, line := range strings.Split(out, "\n") {
		if matches := speckitLsRemoteTagPattern.FindStringSubmatch(line); len(matches) == 2 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("%w: no release tags found at %s", errSpeckitTagResolveFailed, gitURL)
}

// getEnvGitHubToken returns the first GitHub token available, preferring
// existing environment variables over a freshly-fetched gh auth token.
func getEnvGitHubToken() string {
	if t := strings.TrimSpace(os.Getenv("GH_TOKEN")); t != "" {
		return t
	}
	if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
		return t
	}
	return getGitHubTokenFromGH()
}

// resolveSpeckitOwnerRepo returns the configured owner/repo, falling back to
// the default. Centralized so install/upgrade/check share the same source.
func resolveSpeckitOwnerRepo(config *Config) string {
	if config.Speckit.OwnerRepo != "" {
		return config.Speckit.OwnerRepo
	}
	return DefaultSpeckitOwnerRepo
}

// resolveSpeckitGitURL returns the configured bare git URL. Falls back to the
// default. Always returned without the `git+` prefix so callers can compose
// `git+<url>@<tag>` strings cleanly.
func resolveSpeckitGitURL(config *Config) string {
	if config.Speckit.GitURL != "" {
		return config.Speckit.GitURL
	}
	return DefaultSpeckitGitURL
}

// resolveSpeckitIntegration returns the spec-kit integration target,
// preferring the modern Integration field but accepting the deprecated
// AIProvider for back-compat with existing .mage.yaml files.
func resolveSpeckitIntegration(config *Config) string {
	if config.Speckit.Integration != "" {
		return config.Speckit.Integration
	}
	if config.Speckit.AIProvider != "" {
		return config.Speckit.AIProvider
	}
	return DefaultSpeckitIntegration
}

// resolveSpeckitCLIName returns the spec-kit CLI package name with the
// default fallback.
func resolveSpeckitCLIName(config *Config) string {
	if config.Speckit.CLIName != "" {
		return config.Speckit.CLIName
	}
	return DefaultSpeckitCLIName
}

// speckitFromSpec composes the canonical `git+<url>@<tag>` reference uv and
// uvx accept for tag-pinned installs.
func speckitFromSpec(gitURL, tag string) string {
	return fmt.Sprintf("git+%s@%s", gitURL, tag)
}
