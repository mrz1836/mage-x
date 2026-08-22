package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/common/env"
)

func Test_resolveInstallDir(t *testing.T) {
	home := env.Home()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "whitespace only", in: "   ", want: ""},
		{name: "absolute path unchanged", in: "/opt/tools/bin", want: "/opt/tools/bin"},
		{name: "tilde only", in: "~", want: home},
		{name: "tilde slash expands to home", in: "~/.local/bin", want: filepath.Join(home, ".local", "bin")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveInstallDir(tt.in))
		})
	}

	t.Run("expands environment variables", func(t *testing.T) {
		t.Setenv("MAGEX_UT_INSTALL_ROOT", "/custom/root")
		assert.Equal(t, "/custom/root/bin", resolveInstallDir("$MAGEX_UT_INSTALL_ROOT/bin"))
	})
}

func Test_devInstallLocation(t *testing.T) {
	t.Run("explicit install dir wins", func(t *testing.T) {
		assert.Equal(t, "/x/bin", devInstallLocation("/x/bin"))
	})
	t.Run("falls back to GOBIN", func(t *testing.T) {
		t.Setenv("GOBIN", "/gobin")
		assert.Equal(t, "/gobin", devInstallLocation(""))
	})
	t.Run("falls back to GOPATH/bin", func(t *testing.T) {
		t.Setenv("GOBIN", "")
		t.Setenv("GOPATH", "/gopath")
		assert.Equal(t, filepath.Join("/gopath", "bin"), devInstallLocation(""))
	})
}

func Test_applyEnvOverrides_InstallDir(t *testing.T) {
	t.Setenv("MAGE_X_INSTALL_DIR", "/opt/tools/bin")
	c := &Config{}
	applyEnvOverrides(c)
	assert.Equal(t, "/opt/tools/bin", c.Build.InstallDir)
}

// installDirMockRunner records the value of GOBIN at the moment `go install` runs,
// so tests can verify Build.Dev pointed the install at the configured directory.
type installDirMockRunner struct {
	installCalled  bool
	gobinAtInstall string
}

func (m *installDirMockRunner) RunCmd(name string, args ...string) error {
	if name == "go" && len(args) > 0 && args[0] == "install" {
		m.installCalled = true
		m.gobinAtInstall = os.Getenv("GOBIN")
	}
	return nil
}

func (m *installDirMockRunner) RunCmdOutput(_ string, _ ...string) (string, error) {
	return "", nil
}

func TestBuild_Dev_InstallDir(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), "localbin") // deliberately not yet created

	TestSetConfig(&Config{
		Project: ProjectConfig{Name: "testbin", Binary: "testbin"}, // empty Main -> package path falls back to ./...
		Build:   BuildConfig{InstallDir: installDir},
	})
	defer TestResetConfig()

	orig := GetRunner()
	runner := &installDirMockRunner{}
	require.NoError(t, SetRunner(runner))
	defer func() {
		if err := SetRunner(orig); err != nil {
			t.Logf("failed to restore runner: %v", err)
		}
	}()

	gobinBefore, hadBefore := os.LookupEnv("GOBIN")

	require.NoError(t, Build{}.Dev())

	require.True(t, runner.installCalled, "expected `go install` to be invoked")
	assert.Equal(t, installDir, runner.gobinAtInstall, "GOBIN must point at InstallDir during install")
	assert.DirExists(t, installDir, "install dir should be created")

	// GOBIN must be restored to its prior state after Dev returns.
	gobinAfter, hadAfter := os.LookupEnv("GOBIN")
	assert.Equal(t, hadBefore, hadAfter, "GOBIN presence should be restored")
	assert.Equal(t, gobinBefore, gobinAfter, "GOBIN value should be restored")
}

func TestBuild_Dev_NoInstallDir_LeavesGOBIN(t *testing.T) {
	TestSetConfig(&Config{
		Project: ProjectConfig{Name: "testbin", Binary: "testbin"}, // empty Main -> package path falls back to ./...
		Build:   BuildConfig{},                                     // no InstallDir
	})
	defer TestResetConfig()

	orig := GetRunner()
	runner := &installDirMockRunner{}
	require.NoError(t, SetRunner(runner))
	defer func() {
		if err := SetRunner(orig); err != nil {
			t.Logf("failed to restore runner: %v", err)
		}
	}()

	t.Setenv("GOBIN", "/pre/existing/gobin")

	require.NoError(t, Build{}.Dev())

	require.True(t, runner.installCalled)
	// Without an install dir override, GOBIN is left exactly as the environment had it.
	assert.Equal(t, "/pre/existing/gobin", runner.gobinAtInstall)
	assert.Equal(t, "/pre/existing/gobin", os.Getenv("GOBIN"))
}
