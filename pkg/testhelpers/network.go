package testhelpers

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
)

// ErrExternalNetworkBlocked is returned by LoopbackOnlyTransport when a test
// attempts an HTTP request to anything other than loopback.
var ErrExternalNetworkBlocked = errors.New("external network access is blocked in tests")

// blockedProxyURL points every proxy-aware client (curl, git, gh, npm, the Go
// toolchain, govulncheck) at a closed loopback port. Requests fail instantly
// with "connection refused" instead of leaving the machine.
const blockedProxyURL = "http://127.0.0.1:1"

// loopbackExceptions keeps httptest servers reachable for proxy-aware clients.
const loopbackExceptions = "localhost,127.0.0.1,::1"

// blockedNetworkEnv is the environment applied to the test process and, by
// inheritance, to every child process it spawns.
//
// GOPROXY/GOSUMDB/GOTOOLCHAIN are pinned separately from the proxy variables
// because the Go toolchain resolves modules and toolchains through its own
// endpoints rather than the proxy settings.
func blockedNetworkEnv() map[string]string {
	return map[string]string{
		"HTTP_PROXY":  blockedProxyURL,
		"HTTPS_PROXY": blockedProxyURL,
		"http_proxy":  blockedProxyURL,
		"https_proxy": blockedProxyURL,
		"ALL_PROXY":   blockedProxyURL,
		"all_proxy":   blockedProxyURL,
		"NO_PROXY":    loopbackExceptions,
		"no_proxy":    loopbackExceptions,
		"GOPROXY":     "off",
		"GOSUMDB":     "off",
		"GOTOOLCHAIN": "local",
		"GOTELEMETRY": "off",
	}
}

// BlockExternalNetwork makes outbound network access fail fast on loopback
// instead of reaching the internet. Call it from TestMain before m.Run() so the
// whole package - including any command it shells out to - stays hermetic.
//
// Tests that need HTTP should serve it from httptest; loopback is still allowed
// so those keep working. Anything else fails immediately, which keeps test runs
// off the network (and off the user's firewall prompts) and makes an accidental
// live call an obvious error rather than a silent dependency.
func BlockExternalNetwork() error {
	for key, value := range blockedNetworkEnv() {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	return nil
}

// LoopbackOnlyTransport returns an http.RoundTripper that serves loopback
// requests normally and rejects everything else. Proxy environment variables do
// not cover clients that use a custom transport, so packages holding such a
// client (for example utils.DefaultHTTPClient) should install this in TestMain:
//
//	utils.DefaultHTTPClient().Transport = testhelpers.LoopbackOnlyTransport()
func LoopbackOnlyTransport() http.RoundTripper {
	return &loopbackOnlyTransport{base: http.DefaultTransport}
}

// loopbackOnlyTransport rejects any request that is not aimed at loopback.
type loopbackOnlyTransport struct {
	base http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *loopbackOnlyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !isLoopbackHost(req.URL.Hostname()) {
		return nil, fmt.Errorf("%w: %s %s", ErrExternalNetworkBlocked, req.Method, req.URL.Redacted())
	}
	return t.base.RoundTrip(req)
}

// isLoopbackHost reports whether a request host refers to the local machine.
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
