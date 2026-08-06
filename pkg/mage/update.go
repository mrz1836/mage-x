// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/magefile/mage/mg"
	selfupdate "github.com/mrz1836/go-selfupdate"
	"github.com/mrz1836/go-selfupdate/notify"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/mage/runtimectx"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Self-update identity. mage-x updates itself from its own GitHub releases;
// these values pin the owner/repo/binary that go-selfupdate resolves against.
const (
	// updateOwner is the GitHub account hosting mage-x releases.
	updateOwner = "mrz1836"
	// updateRepo is the mage-x repository name.
	updateRepo = "mage-x"
	// updateBinaryName is the executable name inside a mage-x release archive.
	updateBinaryName = "magex"
	// updateAppName identifies mage-x to the passive notifier (banner text,
	// cache slug, and environment-variable prefix).
	updateAppName = "magex"
	// updateModulePath is the go-installable package path for the magex binary.
	// A `go install` build is refreshed with "<updateModulePath>@latest".
	updateModulePath = "github.com/mrz1836/mage-x/cmd/magex"
	// updateTokenEnvVar names the mage-x-specific GitHub token, consulted before
	// GITHUB_TOKEN and GH_TOKEN to lift the anonymous rate limit.
	updateTokenEnvVar = "MAGE_X_GITHUB_TOKEN" //nolint:gosec // G101: environment variable name, not a hardcoded credential
	// updateUpgradeCommand is the line the passive banner tells users to run.
	updateUpgradeCommand = "magex update"

	// legacyDisableUpdateCheckEnv is the pre-go-selfupdate opt-out variable. It
	// is still honored alongside go-selfupdate's MAGEX_NO_UPDATE_CHECK (and the
	// shared NO_UPDATE_CHECK / CI gates) so existing setups keep working.
	legacyDisableUpdateCheckEnv = "MAGEX_DISABLE_UPDATE_CHECK"

	// updateCacheDirName and updateCacheFileName preserve the historical passive
	// check cache location (~/.magex/update-check.json) so the notifier keeps
	// the state it already maintains.
	updateCacheDirName  = ".magex"
	updateCacheFileName = "update-check.json"
)

// Network timeouts bounding the explicit update commands. The passive check
// keeps its own (shorter) timeout inside go-selfupdate.
const (
	// updateCheckTimeout bounds an explicit `update:check` release lookup.
	updateCheckTimeout = 30 * time.Second
	// updateInstallTimeout bounds a full download+install or `go install`.
	updateInstallTimeout = 6 * time.Minute
)

// Update namespace for self-update functionality.
type Update mg.Namespace

// Test seams. These package-level function variables are the entire network
// and exec boundary of the update surface, so the test suite stubs them and
// stays fully offline. Production wires them to go-selfupdate and `go install`.
//
//nolint:gochecknoglobals // seams required to keep the update tests network-free
var (
	checkFn     = selfupdate.Check
	installFn   = selfupdate.Install
	preflightFn = selfupdate.InstallPreflight
	goInstallFn = runGoInstallLatest

	// canGoInstallUpdate reports whether the running binary is a `go install`
	// build AND the go toolchain is available, i.e. an ErrManagedInstall binary
	// can be auto-refreshed with `go install …@latest`. It is a seam so the
	// managed-install branch is testable without a real go-install layout.
	canGoInstallUpdate = func() bool {
		return isGoInstallBinary() && commandExists("go")
	}
)

// registeredBinaryVersion holds the version main resolved at startup. It is
// read from the background update-check goroutine and written by
// RegisterBinaryVersion, so it must be accessed atomically.
//
//nolint:gochecknoglobals // required for version registration from main
var registeredBinaryVersion atomic.Value

// RegisterBinaryVersion records the running binary's resolved version so both
// the update commands and the passive banner report it correctly. main calls
// this once at startup with the value from ResolveVersion; when it is not
// called (library use, tests) the version falls back to "dev".
func RegisterBinaryVersion(version string) {
	registeredBinaryVersion.Store(version)
}

// resolvedBinaryVersion returns the version registered by main, or "dev" when
// none was registered.
func resolvedBinaryVersion() string {
	if v, ok := registeredBinaryVersion.Load().(string); ok && v != "" {
		return v
	}
	return versionDev
}

// ResolveVersion picks the most trustworthy version string for the running
// binary, so both supported install methods report a correct version:
//
//   - a goreleaser binary carries its version in ldflags (main.version);
//   - a `go install …@vX` build carries the module version in build info;
//   - a local `go build` has neither and is reported as "dev".
//
// The ldflags value wins when present because it is the release's own stamp;
// build info is the fallback that makes `go install` builds self-aware.
func ResolveVersion(ldflags string) string {
	if ldflags != "" && ldflags != versionDev {
		return ldflags
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return versionDev
}

// selfUpdateConfig builds the go-selfupdate configuration for mage-x's own
// binary. TargetPath is left empty so go-selfupdate resolves the running
// executable (symlinks followed) itself.
func selfUpdateConfig() selfupdate.Config {
	return selfupdate.Config{
		Owner:          updateOwner,
		Repo:           updateRepo,
		BinaryName:     updateBinaryName,
		CurrentVersion: resolvedBinaryVersion(),
		TokenEnvVar:    updateTokenEnvVar,
	}
}

// notifyConfig builds the passive-notifier configuration. It pins the cache to
// the historical ~/.magex/update-check.json location and derives the mage-x
// opt-out variable (MAGEX_NO_UPDATE_CHECK) and token from the app identity.
func notifyConfig() notify.Config {
	return notify.Config{
		Owner:          updateOwner,
		Repo:           updateRepo,
		AppName:        updateAppName,
		BinaryName:     updateBinaryName,
		CurrentVersion: resolvedBinaryVersion(),
		UpgradeCommand: updateUpgradeCommand,
		TokenEnvVar:    updateTokenEnvVar,
		CacheDir:       filepath.Join(env.Home(), updateCacheDirName),
		CacheFileName:  updateCacheFileName,
	}
}

// updateArgs is the parsed set of update command switches.
type updateArgs struct {
	check   bool
	force   bool
	verbose bool
}

// parseUpdateArgs accepts BOTH CLI syntaxes so the same handler serves the new
// verbs and the historical namespaced commands:
//
//   - flag form:      --check/-c, --force/-f, --verbose/-v
//   - key=value form: check=true, force=true, verbose=true (or bare check/force/verbose)
//
// Dash tokens are matched as flags; any other unknown dash token is ignored so
// it never pollutes the key=value parse. Remaining tokens are parsed as
// parameters, so `magex update:install force=true` and `magex update --force`
// are equivalent.
func parseUpdateArgs(args ...string) updateArgs {
	var ua updateArgs
	kv := make([]string, 0, len(args))

	for _, a := range args {
		switch a {
		case "--check", "-c", "-check":
			ua.check = true
		case "--force", "-f", "-force":
			ua.force = true
		case "--verbose", "-v", "-verbose":
			ua.verbose = true
		default:
			// An unrecognized dash token is not a key=value parameter; drop it
			// rather than let it become params["--whatever"]=true.
			if strings.HasPrefix(a, "-") {
				continue
			}
			kv = append(kv, a)
		}
	}

	params := utils.ParseParams(kv)
	if utils.IsParamTrue(params, "check") {
		ua.check = true
	}
	if utils.IsParamTrue(params, "force") {
		ua.force = true
	}
	if utils.IsParamTrue(params, "verbose") {
		ua.verbose = true
	}
	return ua
}

// Check reports whether a newer magex release is available. It performs no
// writes. Arguments are accepted for CLI symmetry (so `magex update:check
// verbose=true` is valid) but do not change the read-only behavior.
func (Update) Check(_ ...string) error {
	utils.Header("Checking for Updates")

	cfg := selfUpdateConfig()
	ctx, cancel := context.WithTimeout(runtimectx.Context(), updateCheckTimeout)
	defer cancel()

	info, err := checkFn(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if info.UpdateAvailable {
		utils.Info("A new version of magex is available: %s -> %s", info.CurrentVersion, info.LatestVersion)
		utils.Info(`Run "magex update" to install it.`)
	} else {
		utils.Success("magex is up to date (%s)", info.CurrentVersion)
	}
	return nil
}

// Install updates magex to the latest release. It supports both install
// methods:
//
//   - A binary install in a writable, unmanaged directory (e.g. ~/.local/bin)
//     is replaced in place by go-selfupdate.
//   - A `go install` build (in GOBIN/GOPATH/bin) cannot be overwritten by
//     go-selfupdate, so it is refreshed by auto-running
//     `go install <module>@latest`.
//   - Any other managed case (Homebrew, a root-owned system dir, or `go`
//     missing) falls back to clear, copy-pasteable guidance.
func (Update) Install(args ...string) error {
	ua := parseUpdateArgs(args...)

	// check-only short-circuits to the read-only report.
	if ua.check {
		return Update{}.Check(args...)
	}

	utils.Header("Installing Update")

	cfg := selfUpdateConfig()

	// Preflight resolves the running binary and applies the same location
	// guards Install would, without any network access, so we can pick the
	// right route before downloading anything.
	preErr := preflightFn(cfg)
	switch {
	case preErr == nil:
		return installInPlace(cfg, ua)
	case errors.Is(preErr, selfupdate.ErrManagedInstall):
		return installManaged()
	case errors.Is(preErr, selfupdate.ErrInstallDirNotWritable):
		utils.Error("Cannot update magex in place: %v", preErr)
		utils.Info("Move the magex binary to a user-writable directory on your PATH")
		utils.Info("(for example ~/.local/bin) to enable in-place self-update, or reinstall with:")
		utils.Info("  go install %s@latest", updateModulePath)
		return preErr
	default:
		return fmt.Errorf("update preflight failed: %w", preErr)
	}
}

// installInPlace performs a go-selfupdate in-place binary replacement. It maps
// the parsed switches onto go-selfupdate options and, on success, prints a
// closing line and clears the passive-check cache so the next run does not show
// a stale "update available" banner.
func installInPlace(cfg selfupdate.Config, ua updateArgs) error {
	ctx, cancel := context.WithTimeout(runtimectx.Context(), updateInstallTimeout)
	defer cancel()

	var opts []selfupdate.Option
	if ua.force {
		opts = append(opts, selfupdate.WithForce())
	}
	if ua.verbose {
		opts = append(opts, selfupdate.WithVerbose())
	}

	result, err := installFn(ctx, cfg, opts...)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if result.Updated {
		utils.Success("Updated magex to %s", result.LatestVersion)
		clearUpdateCache()
	}
	// When not updated, go-selfupdate has already explained why on stdout
	// ("already up to date" or "development build; run with --force"), so we
	// add nothing here.
	return nil
}

// installManaged handles the ErrManagedInstall case. A `go install` build is
// updated by auto-running `go install <module>@latest`; every other managed
// case (Homebrew, root-owned system dir, or `go` missing) is instructed
// instead, since overwriting a file another installer owns would break both.
func installManaged() error {
	if canGoInstallUpdate() {
		utils.Info("Detected a `go install` build of magex; updating with:")
		utils.Info("  go install %s@latest", updateModulePath)

		ctx, cancel := context.WithTimeout(runtimectx.Context(), updateInstallTimeout)
		defer cancel()

		if err := goInstallFn(ctx); err != nil {
			return fmt.Errorf("go install update failed: %w", err)
		}

		utils.Success("Updated magex via `go install %s@latest`", updateModulePath)
		clearUpdateCache()
		return nil
	}

	utils.Warn("magex was installed by another tool and cannot self-update in place.")
	utils.Info("To update, run:")
	utils.Info("  go install %s@latest", updateModulePath)
	utils.Info("Or drop a prebuilt binary into a user-writable directory on your PATH")
	utils.Info("(for example ~/.local/bin) to enable in-place self-update.")
	return nil
}

// isGoInstallBinary reports whether the running binary lives in GOBIN or a
// GOPATH/bin directory — i.e. it was produced by `go install` rather than
// dropped in as a release binary. It resolves symlinks on both sides so a
// GOPATH under a symlinked home still matches, and is self-contained so it does
// not depend on go-selfupdate internals. This distinguishes the go-install case
// (auto-updatable via `go install`) from other managed cases (Homebrew, system
// package managers) that share the ErrManagedInstall signal.
func isGoInstallBinary() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return isGoInstallPath(exe, os.Getenv)
}

// isGoInstallPath is isGoInstallBinary with the executable path and environment
// injected, so it is testable without touching process state.
func isGoInstallPath(exePath string, getenv func(string) string) bool {
	if resolved, rerr := filepath.EvalSymlinks(exePath); rerr == nil {
		exePath = resolved
	}
	dir := realDir(filepath.Dir(exePath))
	if dir == "" {
		return false
	}

	var candidates []string
	if gobin := strings.TrimSpace(getenv("GOBIN")); gobin != "" {
		candidates = append(candidates, gobin)
	}

	gopath := strings.TrimSpace(getenv("GOPATH"))
	if gopath == "" {
		if home, herr := os.UserHomeDir(); herr == nil {
			gopath = filepath.Join(home, "go")
		}
	}
	for _, entry := range filepath.SplitList(gopath) {
		if entry != "" {
			candidates = append(candidates, filepath.Join(entry, "bin"))
		}
	}

	for _, candidate := range candidates {
		if resolved := realDir(candidate); resolved != "" && resolved == dir {
			return true
		}
	}
	return false
}

// realDir returns dir as an absolute path with symlinks resolved, falling back
// to the cleaned absolute form when the path does not exist on disk.
func realDir(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}
	if resolved, rerr := filepath.EvalSymlinks(abs); rerr == nil {
		return resolved
	}
	return filepath.Clean(abs)
}

// runGoInstallLatest refreshes a go-install build of magex by running
// `go install <module>@latest`. It is the production implementation behind the
// goInstallFn seam.
func runGoInstallLatest(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return GetRunner().RunCmd("go", "install", updateModulePath+"@latest")
}

// clearUpdateCache removes the passive-check cache so the next invocation
// performs a fresh check and does not show a stale "update available" banner
// right after an update.
func clearUpdateCache() {
	if err := notify.ClearCache(notifyConfig()); err != nil {
		utils.Warn("failed to clear update cache: %v", err)
	}
}

// RegisterBinaryVersion, StartBackgroundUpdateCheck, and ShowUpdateBanner are
// the three thin wrappers main wires up; keeping them here keeps main's call
// sites unchanged across the refactor.

// StartBackgroundUpdateCheck kicks off the passive update check in the
// background and returns a channel that yields at most one result. It honors the
// legacy MAGEX_DISABLE_UPDATE_CHECK opt-out in addition to go-selfupdate's own
// gates (MAGEX_NO_UPDATE_CHECK, the shared NO_UPDATE_CHECK, CI, and dev builds).
func StartBackgroundUpdateCheck(ctx context.Context) <-chan *notify.Result {
	if env.GetBool(legacyDisableUpdateCheckEnv, false) {
		ch := make(chan *notify.Result)
		close(ch)
		return ch
	}
	return notify.StartBackgroundCheck(ctx, notifyConfig())
}

// ShowUpdateBanner prints the passive "a new version is available" banner when
// result reports an available update, and stays silent otherwise.
func ShowUpdateBanner(result *notify.Result) {
	notify.ShowBanner(notifyConfig(), result)
}
