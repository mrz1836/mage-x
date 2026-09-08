package mage

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for the release idempotency pre-check.
var (
	errOwnerRepoUnresolved     = errors.New("could not determine GitHub owner/repo")
	errReleaseCheckToolMissing = errors.New("required tool not installed for release pre-check")
	errReleaseCheckStatus      = errors.New("unexpected GitHub API status during release pre-check")
)

// ownerRepoPattern captures the "owner/repo" slug from either SSH or HTTPS
// GitHub remote URLs, tolerating an optional trailing ".git".
//   - git@github.com:owner/repo.git
//   - https://github.com/owner/repo.git
var ownerRepoPattern = regexp.MustCompile(`[:/]([^/:]+/[^/]+?)(?:\.git)?/?$`)

// releaseDraftPattern extracts the boolean "draft" field from a GitHub release
// API response without pulling in a JSON dependency (matching the lightweight
// approach already used for tag_name lookups in speckit_release.go).
var releaseDraftPattern = regexp.MustCompile(`"draft"\s*:\s*(true|false)`)

// currentOwnerRepo resolves the "owner/repo" slug for the repository being
// released. It prefers GITHUB_REPOSITORY (always set inside GitHub Actions) and
// falls back to parsing the origin remote URL so the check also works locally.
func currentOwnerRepo() (string, error) {
	if r := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY")); r != "" {
		return r, nil
	}

	out, err := GetRunner().RunCmdOutput("git", "config", "--get", "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("%w: %w", errOwnerRepoUnresolved, err)
	}

	matches := ownerRepoPattern.FindStringSubmatch(strings.TrimSpace(out))
	if len(matches) < 2 {
		return "", fmt.Errorf("%w: unrecognized remote %q", errOwnerRepoUnresolved, strings.TrimSpace(out))
	}
	return matches[1], nil
}

// releaseAlreadyPublished reports whether a published GitHub release already
// exists for tag, logging what it finds. It never returns true on uncertainty:
// if the owner/repo cannot be resolved or the API lookup fails, it warns and
// returns false so the caller proceeds to GoReleaser (surfacing a real problem
// rather than silently skipping the release).
func releaseAlreadyPublished(tag string) bool {
	ownerRepo, err := currentOwnerRepo()
	if err != nil {
		utils.Warn("Skipping release pre-check (continuing to GoReleaser): %v", err)
		return false
	}

	published, err := publishedReleaseExists(ownerRepo, tag)
	if err != nil {
		utils.Warn("Release pre-check failed (continuing to GoReleaser): %v", err)
		return false
	}
	if published {
		utils.Success("Release %s is already published for %s (likely a duplicate or concurrent run); skipping GoReleaser", tag, ownerRepo)
	}
	return published
}

// publishedReleaseExists reports whether a published (non-draft) GitHub release
// already exists for tag. It is used to make `magex release` idempotent: when a
// duplicate or concurrent workflow run has already published the tag, GitHub's
// immutable-releases feature makes a second `goreleaser release` fail with
// "already exists and is immutable". Detecting that up front lets the release
// skip GoReleaser and succeed instead of failing the job.
//
// A draft release (a prior run that created but never published the release)
// returns false so the real release can still proceed. Any lookup error is
// returned to the caller, which treats it as "unknown" and proceeds to
// GoReleaser rather than masking a genuine problem.
func publishedReleaseExists(ownerRepo, tag string) (bool, error) {
	if !utils.CommandExists("curl") {
		return false, fmt.Errorf("%w: curl", errReleaseCheckToolMissing)
	}

	url := fmt.Sprintf("%s/repos/%s/releases/tags/%s", gitHubAPIBaseURL(), ownerRepo, tag)
	args := []string{
		"-sSL",
		"-w", "\\n%{http_code}", // append the HTTP status on its own trailing line
		"-H", "Accept: application/vnd.github+json",
		"-H", "X-GitHub-Api-Version: 2022-11-28",
	}
	if token := getEnvGitHubToken(); token != "" {
		args = append(args, "-H", "Authorization: Bearer "+token)
	}
	args = append(args, url)

	out, err := GetRunner().RunCmdOutput("curl", args...)
	if err != nil {
		return false, fmt.Errorf("curl %s: %w", url, err)
	}

	body, status := splitBodyStatus(out)
	switch status {
	case "404":
		// No release for this tag yet — proceed with the release.
		return false, nil
	case "200":
		// Release exists; only treat it as "already done" if it is published.
		if m := releaseDraftPattern.FindStringSubmatch(body); len(m) == 2 && m[1] == "true" {
			return false, nil
		}
		return true, nil
	default:
		return false, fmt.Errorf("%w: %s for %s", errReleaseCheckStatus, status, url)
	}
}

// splitBodyStatus separates the response body from the trailing HTTP status code
// that curl's `-w "\n%{http_code}"` appends. If no trailing status is present the
// whole string is returned as the body with an empty status.
func splitBodyStatus(out string) (body, status string) {
	out = strings.TrimSpace(out)
	idx := strings.LastIndex(out, "\n")
	if idx == -1 {
		return out, ""
	}
	return strings.TrimSpace(out[:idx]), strings.TrimSpace(out[idx+1:])
}
