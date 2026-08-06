package mage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkglog "github.com/mrz1836/mage-x/pkg/log"
)

// errUpdateTest is a sentinel used to drive the update error paths offline.
var errUpdateTest = errors.New("update test failure")

// ============================================================================
// Test seams / helpers — every one keeps the update path fully offline.
// ============================================================================

// saveUpdateSeams snapshots and restores the update seam variables so a test
// can stub the network/exec boundary without leaking into its neighbors.
func saveUpdateSeams(t *testing.T) {
	t.Helper()
	oc, oi, op, og, ocan := checkFn, installFn, preflightFn, goInstallFn, canGoInstallUpdate
	t.Cleanup(func() {
		checkFn, installFn, preflightFn, goInstallFn, canGoInstallUpdate = oc, oi, op, og, ocan
	})
}

// stubbedCurrentVersion is the running version reported by withStubbedUpdateCheck.
const stubbedCurrentVersion = "v1.0.0"

// withStubbedUpdateCheck stubs the checkFn seam so Update.Check runs fully
// offline, reporting stubbedCurrentVersion as the running build and the given
// latest/availability. It is shared with the namespace tests that exercise
// Update.Check through the namespace wrapper.
func withStubbedUpdateCheck(t *testing.T, latest string, available bool) {
	t.Helper()
	orig := checkFn
	checkFn = func(_ context.Context, _ selfupdate.Config) (*selfupdate.Info, error) {
		return &selfupdate.Info{
			CurrentVersion:  stubbedCurrentVersion,
			LatestVersion:   latest,
			UpdateAvailable: available,
		}, nil
	}
	t.Cleanup(func() { checkFn = orig })
}

// captureUpdateLog redirects the shared CLI logger into a buffer for the
// duration of fn and returns everything written. The update tests are not
// parallel, so the global redirect is safe.
func captureUpdateLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	logger := pkglog.Default()
	logger.SetOutput(&buf)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })
	fn()
	return buf.String()
}

// roundTripFunc adapts a function into an http.RoundTripper so a test can serve
// release archives and checksums from memory, with no sockets.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func memResponse(code int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(body)),
		Header:     make(http.Header),
	}
}

// fakeReleaseSource is an injected selfupdate.ReleaseSource that returns a canned
// release without touching the network.
type fakeReleaseSource struct{ release *selfupdate.Release }

func (f fakeReleaseSource) Latest(_ context.Context) (*selfupdate.Release, error) {
	return f.release, nil
}

// makeTarGz builds an in-memory gzip'd tar containing a single named file.
func makeTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     name,
		Mode:     0o755,
		Size:     int64(len(content)),
		Typeflag: tar.TypeReg,
	}))
	_, err := tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	return buf.Bytes()
}

// ============================================================================
// ResolveVersion
// ============================================================================

func TestResolveVersion(t *testing.T) {
	t.Run("ldflags value wins", func(t *testing.T) {
		assert.Equal(t, "v1.2.3", ResolveVersion("v1.2.3"))
	})

	t.Run("dev ldflags falls through to build info or dev", func(t *testing.T) {
		// Under `go test`, build info Main.Version is "(devel)"/empty, so the
		// resolver falls all the way through to "dev".
		assert.Equal(t, versionDev, ResolveVersion(versionDev))
	})

	t.Run("empty ldflags falls through to build info or dev", func(t *testing.T) {
		assert.Equal(t, versionDev, ResolveVersion(""))
	})
}

// ============================================================================
// parseUpdateArgs — both CLI syntaxes
// ============================================================================

func TestParseUpdateArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want updateArgs
	}{
		{"empty", nil, updateArgs{}},
		{"flag --check", []string{"--check"}, updateArgs{check: true}},
		{"flag -c", []string{"-c"}, updateArgs{check: true}},
		{"flag --force", []string{"--force"}, updateArgs{force: true}},
		{"flag -f", []string{"-f"}, updateArgs{force: true}},
		{"flag --verbose", []string{"--verbose"}, updateArgs{verbose: true}},
		{"flag -v", []string{"-v"}, updateArgs{verbose: true}},
		{"single-dash long forms", []string{"-check", "-force", "-verbose"}, updateArgs{check: true, force: true, verbose: true}},
		{"kv check=true", []string{"check=true"}, updateArgs{check: true}},
		{"kv force=true", []string{"force=true"}, updateArgs{force: true}},
		{"kv verbose=true", []string{"verbose=true"}, updateArgs{verbose: true}},
		{"bare kv token", []string{"check"}, updateArgs{check: true}},
		{"kv false is not set", []string{"check=false"}, updateArgs{}},
		{"mix flag and kv", []string{"--force", "verbose=true"}, updateArgs{force: true, verbose: true}},
		{"all short flags", []string{"-c", "-f", "-v"}, updateArgs{check: true, force: true, verbose: true}},
		{"unknown dash token ignored", []string{"--nope", "force=true"}, updateArgs{force: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseUpdateArgs(tt.args...))
		})
	}
}

// ============================================================================
// Config builders
// ============================================================================

func TestSelfUpdateConfigIdentity(t *testing.T) {
	cfg := selfUpdateConfig()
	assert.Equal(t, updateOwner, cfg.Owner)
	assert.Equal(t, updateRepo, cfg.Repo)
	assert.Equal(t, updateBinaryName, cfg.BinaryName)
	assert.Equal(t, updateTokenEnvVar, cfg.TokenEnvVar)
	assert.NotEmpty(t, cfg.CurrentVersion)
}

func TestNotifyConfigIdentity(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv("HOME", home)
	cfg := notifyConfig()
	assert.Equal(t, updateOwner, cfg.Owner)
	assert.Equal(t, updateRepo, cfg.Repo)
	assert.Equal(t, updateAppName, cfg.AppName)
	assert.Equal(t, updateUpgradeCommand, cfg.UpgradeCommand)
	assert.Equal(t, updateTokenEnvVar, cfg.TokenEnvVar)
	assert.Equal(t, filepath.Join(home, updateCacheDirName), cfg.CacheDir)
	assert.Equal(t, updateCacheFileName, cfg.CacheFileName)
}

// ============================================================================
// Update.Check — output for both branches
// ============================================================================

func TestUpdateCheck(t *testing.T) {
	t.Run("up to date", func(t *testing.T) {
		saveUpdateSeams(t)
		withStubbedUpdateCheck(t, "v1.0.0", false)
		out := captureUpdateLog(t, func() {
			require.NoError(t, Update{}.Check())
		})
		assert.Contains(t, out, "magex is up to date (v1.0.0)")
	})

	t.Run("update available", func(t *testing.T) {
		saveUpdateSeams(t)
		withStubbedUpdateCheck(t, "v2.0.0", true)
		out := captureUpdateLog(t, func() {
			require.NoError(t, Update{}.Check())
		})
		assert.Contains(t, out, "A new version of magex is available: v1.0.0 -> v2.0.0")
		assert.Contains(t, out, `Run "magex update" to install it.`)
	})

	t.Run("check error is surfaced", func(t *testing.T) {
		saveUpdateSeams(t)
		checkFn = func(_ context.Context, _ selfupdate.Config) (*selfupdate.Info, error) {
			return nil, errUpdateTest
		}
		require.Error(t, Update{}.Check())
	})
}

// ============================================================================
// Update.Install — branch routing (which seam ran + output)
// ============================================================================

func TestUpdateInstallBinaryInPlace(t *testing.T) {
	saveUpdateSeams(t)
	t.Setenv("HOME", t.TempDir()) // isolate the cache clear

	var installCalled, goCalled bool
	var gotOpts int
	preflightFn = func(_ selfupdate.Config) error { return nil }
	installFn = func(_ context.Context, _ selfupdate.Config, opts ...selfupdate.Option) (selfupdate.Result, error) {
		installCalled = true
		gotOpts = len(opts)
		return selfupdate.Result{Updated: true, PreviousVersion: "v1.0.0", LatestVersion: "v2.0.0"}, nil
	}
	goInstallFn = func(_ context.Context) error { goCalled = true; return nil }

	out := captureUpdateLog(t, func() {
		require.NoError(t, Update{}.Install())
	})

	assert.True(t, installCalled, "installFn should run for a writable, unmanaged binary")
	assert.False(t, goCalled, "goInstallFn must not run for an in-place update")
	assert.Equal(t, 0, gotOpts, "no flags means no selfupdate options")
	assert.Contains(t, out, "Updated magex to v2.0.0")
}

func TestUpdateInstallMapsForceAndVerbose(t *testing.T) {
	saveUpdateSeams(t)
	t.Setenv("HOME", t.TempDir())

	var gotOpts int
	preflightFn = func(_ selfupdate.Config) error { return nil }
	installFn = func(_ context.Context, _ selfupdate.Config, opts ...selfupdate.Option) (selfupdate.Result, error) {
		gotOpts = len(opts)
		return selfupdate.Result{Updated: true, LatestVersion: "v2.0.0"}, nil
	}

	require.NoError(t, Update{}.Install("--force", "--verbose"))
	assert.Equal(t, 2, gotOpts, "--force and --verbose should each map to one selfupdate option")
}

func TestUpdateInstallGoInstallBinary(t *testing.T) {
	saveUpdateSeams(t)
	t.Setenv("HOME", t.TempDir())

	var installCalled, goCalled bool
	preflightFn = func(_ selfupdate.Config) error {
		return fmt.Errorf("%w: go bin", selfupdate.ErrManagedInstall)
	}
	canGoInstallUpdate = func() bool { return true }
	installFn = func(_ context.Context, _ selfupdate.Config, _ ...selfupdate.Option) (selfupdate.Result, error) {
		installCalled = true
		return selfupdate.Result{}, nil
	}
	goInstallFn = func(_ context.Context) error { goCalled = true; return nil }

	out := captureUpdateLog(t, func() {
		require.NoError(t, Update{}.Install())
	})

	assert.True(t, goCalled, "goInstallFn should run for a go-install build")
	assert.False(t, installCalled, "installFn must not run for a managed binary")
	assert.Contains(t, out, "go install "+updateModulePath+"@latest")
}

func TestUpdateInstallManagedInstructs(t *testing.T) {
	saveUpdateSeams(t)

	var installCalled, goCalled bool
	preflightFn = func(_ selfupdate.Config) error {
		return fmt.Errorf("%w: homebrew", selfupdate.ErrManagedInstall)
	}
	canGoInstallUpdate = func() bool { return false } // e.g. Homebrew, or `go` missing
	installFn = func(_ context.Context, _ selfupdate.Config, _ ...selfupdate.Option) (selfupdate.Result, error) {
		installCalled = true
		return selfupdate.Result{}, nil
	}
	goInstallFn = func(_ context.Context) error { goCalled = true; return nil }

	out := captureUpdateLog(t, func() {
		require.NoError(t, Update{}.Install())
	})

	assert.False(t, installCalled, "no in-place install for a managed binary")
	assert.False(t, goCalled, "no auto go-install when it is not a go-install build")
	assert.Contains(t, out, "go install "+updateModulePath+"@latest")
	assert.Contains(t, out, "~/.local/bin")
}

func TestUpdateInstallNotWritable(t *testing.T) {
	saveUpdateSeams(t)

	var installCalled, goCalled bool
	preflightFn = func(_ selfupdate.Config) error {
		return fmt.Errorf("%w: /usr/local/bin", selfupdate.ErrInstallDirNotWritable)
	}
	installFn = func(_ context.Context, _ selfupdate.Config, _ ...selfupdate.Option) (selfupdate.Result, error) {
		installCalled = true
		return selfupdate.Result{}, nil
	}
	goInstallFn = func(_ context.Context) error { goCalled = true; return nil }

	var err error
	out := captureUpdateLog(t, func() {
		err = Update{}.Install()
	})

	require.Error(t, err)
	require.ErrorIs(t, err, selfupdate.ErrInstallDirNotWritable)
	assert.False(t, installCalled)
	assert.False(t, goCalled)
	assert.Contains(t, out, "~/.local/bin")
}

func TestUpdateInstallCheckOnlyDelegates(t *testing.T) {
	saveUpdateSeams(t)

	var checkCalled, preflightCalled bool
	checkFn = func(_ context.Context, _ selfupdate.Config) (*selfupdate.Info, error) {
		checkCalled = true
		return &selfupdate.Info{CurrentVersion: "v1.0.0", LatestVersion: "v1.0.0"}, nil
	}
	preflightFn = func(_ selfupdate.Config) error { preflightCalled = true; return nil }

	require.NoError(t, Update{}.Install("--check"))
	assert.True(t, checkCalled, "check-only should perform a read-only check")
	assert.False(t, preflightCalled, "check-only must not reach preflight/install")
}

func TestUpdateInstallPreflightErrorSurfaces(t *testing.T) {
	saveUpdateSeams(t)
	preflightFn = func(_ selfupdate.Config) error { return errUpdateTest }
	require.Error(t, Update{}.Install())
}

// ============================================================================
// isGoInstallBinary / isGoInstallPath
// ============================================================================

func TestIsGoInstallPath(t *testing.T) {
	tmp := t.TempDir()

	gopath := filepath.Join(tmp, "go")
	gopathBin := filepath.Join(gopath, "bin")
	require.NoError(t, os.MkdirAll(gopathBin, 0o750))
	gopathExe := filepath.Join(gopathBin, "magex")
	require.NoError(t, os.WriteFile(gopathExe, []byte("x"), 0o600))

	gobin := filepath.Join(tmp, "custombin")
	require.NoError(t, os.MkdirAll(gobin, 0o750))
	gobinExe := filepath.Join(gobin, "magex")
	require.NoError(t, os.WriteFile(gobinExe, []byte("x"), 0o600))

	local := filepath.Join(tmp, "local")
	require.NoError(t, os.MkdirAll(local, 0o750))
	localExe := filepath.Join(local, "magex")
	require.NoError(t, os.WriteFile(localExe, []byte("x"), 0o600))

	t.Run("binary under GOPATH/bin is a go-install build", func(t *testing.T) {
		getenv := func(k string) string {
			if k == "GOPATH" {
				return gopath
			}
			return ""
		}
		assert.True(t, isGoInstallPath(gopathExe, getenv))
	})

	t.Run("binary under GOBIN is a go-install build", func(t *testing.T) {
		getenv := func(k string) string {
			if k == "GOBIN" {
				return gobin
			}
			return ""
		}
		assert.True(t, isGoInstallPath(gobinExe, getenv))
	})

	t.Run("binary in an unrelated dir is not a go-install build", func(t *testing.T) {
		getenv := func(k string) string {
			if k == "GOPATH" {
				return gopath
			}
			return ""
		}
		assert.False(t, isGoInstallPath(localExe, getenv))
	})
}

// ============================================================================
// StartBackgroundUpdateCheck — legacy opt-out
// ============================================================================

func TestStartBackgroundUpdateCheckLegacyOptOut(t *testing.T) {
	t.Setenv(legacyDisableUpdateCheckEnv, "true")

	ch := StartBackgroundUpdateCheck(context.Background())
	select {
	case r, ok := <-ch:
		assert.False(t, ok, "channel must be closed with no result when opted out")
		assert.Nil(t, r)
	case <-time.After(2 * time.Second):
		t.Fatal("expected an immediately closed channel")
	}
}

// ============================================================================
// High-fidelity: real go-selfupdate Check+Install, offline via injected Source
// and an in-memory RoundTripper. No sockets.
// ============================================================================

func TestSelfUpdatePipelineOffline(t *testing.T) {
	if runtime.GOOS == OSWindows {
		t.Skip("go-selfupdate does not support in-place install on Windows")
	}

	const version = "v2.3.4"
	binContent := []byte("#!/bin/sh\necho new-magex\n")
	archive := makeTarGz(t, updateBinaryName, binContent)

	assetName := fmt.Sprintf("mage-x_2.3.4_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	checksumAsset := "mage-x_2.3.4_checksums.txt"
	digest := sha256.Sum256(archive)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(digest[:]), assetName)

	const (
		downloadURL = "https://example.invalid/dl/" + "mage-x.tar.gz"
		checksumURL = "https://example.invalid/dl/checksums.txt"
	)

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case downloadURL:
			return memResponse(http.StatusOK, archive), nil
		case checksumURL:
			return memResponse(http.StatusOK, []byte(checksums)), nil
		default:
			return memResponse(http.StatusNotFound, nil), nil
		}
	})

	release := &selfupdate.Release{
		TagName: version,
		Name:    version,
		Assets: []selfupdate.ReleaseAsset{
			{Name: assetName, BrowserDownloadURL: downloadURL},
			{Name: checksumAsset, BrowserDownloadURL: checksumURL},
		},
	}

	target := filepath.Join(t.TempDir(), updateBinaryName)
	require.NoError(t, os.WriteFile(target, []byte("old-magex"), 0o600))

	cfg := selfupdate.Config{
		Owner:          updateOwner,
		Repo:           updateRepo,
		BinaryName:     updateBinaryName,
		CurrentVersion: "v1.0.0",
		TargetPath:     target,
		Source:         fakeReleaseSource{release: release},
		Client:         &http.Client{Transport: transport},
		Stdout:         io.Discard,
	}

	t.Run("check reports the newer version", func(t *testing.T) {
		info, err := selfupdate.Check(context.Background(), cfg)
		require.NoError(t, err)
		assert.True(t, info.UpdateAvailable)
		assert.Equal(t, version, info.LatestVersion)
		assert.Equal(t, assetName, info.AssetName)
	})

	t.Run("install verifies the checksum and replaces the binary", func(t *testing.T) {
		res, err := selfupdate.Install(context.Background(), cfg)
		require.NoError(t, err)
		assert.True(t, res.Updated)
		assert.Equal(t, version, res.LatestVersion)

		got, err := os.ReadFile(target) //nolint:gosec // G304: target is a test-controlled temp path
		require.NoError(t, err)
		assert.Equal(t, binContent, got, "the target binary should be replaced with the archived one")
	})
}
