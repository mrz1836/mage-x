package mage

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// withFakeGitHubAPI points GitHub release lookups at a local test server and
// hides the gh CLI for the duration of the test, so release code paths run
// against a known payload instead of reaching api.github.com.
//
// Both the package seam and GITHUB_API_URL are set: the env var wins in
// gitHubAPIBaseURL, and CI runners (GitHub Actions) define it, so leaving it
// alone would send the test back to the real API.
func withFakeGitHubAPI(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	originalBaseURL := githubAPIBaseURL
	githubAPIBaseURL = server.URL
	t.Setenv("GITHUB_API_URL", server.URL)

	originalCommandExists := commandExists
	commandExists = func(name string) bool {
		if name == "gh" {
			return false
		}
		return originalCommandExists(name)
	}

	t.Cleanup(func() {
		githubAPIBaseURL = originalBaseURL
		commandExists = originalCommandExists
	})
}

// writeFakeGitHubJSON writes a JSON body for a fake GitHub API response.
func writeFakeGitHubJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("failed to write fake GitHub API response: %v", err)
	}
}
