package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_applyEnvOverrides_TestCoverPkg(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		yamlPkgs []string
		want     []string
	}{
		{
			name:     "env sets coverpkg when yaml is empty",
			envValue: "./...",
			want:     []string{"./..."},
		},
		{
			name:     "env wins over yaml",
			envValue: "./...",
			yamlPkgs: []string{"./internal/..."},
			want:     []string{"./..."},
		},
		{
			name:     "comma-separated list is split and trimmed",
			envValue: " ./internal/... , github.com/example/project/pkg/... ",
			want:     []string{"./internal/...", "github.com/example/project/pkg/..."},
		},
		{
			name:     "empty env keeps yaml value",
			envValue: "",
			yamlPkgs: []string{"./internal/..."},
			want:     []string{"./internal/..."},
		},
		{
			name:     "only separators keeps yaml value",
			envValue: " , , ",
			yamlPkgs: []string{"./internal/..."},
			want:     []string{"./internal/..."},
		},
		{
			name:     "unset env and empty yaml stays empty",
			envValue: "",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MAGE_X_TEST_COVERPKG", tt.envValue)
			c := &Config{}
			c.Test.CoverPkg = tt.yamlPkgs
			applyEnvOverrides(c)
			assert.Equal(t, tt.want, c.Test.CoverPkg)
		})
	}
}

// coverPkgRecordingRunner records the args of every `go test` call.
type coverPkgRecordingRunner struct {
	testArgs [][]string
}

func (r *coverPkgRecordingRunner) RunCmd(name string, args ...string) error {
	if name == "go" && len(args) > 0 && args[0] == "test" {
		r.testArgs = append(r.testArgs, args)
	}
	return nil
}

func (r *coverPkgRecordingRunner) RunCmdOutput(name string, args ...string) (string, error) {
	return "", r.RunCmd(name, args...)
}

func Test_runCoverageTests_UsesEnvCoverPkg(t *testing.T) {
	t.Setenv("MAGE_X_TEST_COVERPKG", "./...")
	c := &Config{}
	c.Test.CoverMode = "atomic"
	applyEnvOverrides(c)

	module := ModuleInfo{Path: t.TempDir(), Relative: ".", IsRoot: true, Name: "example"}
	runner := &coverPkgRecordingRunner{}

	err := runCoverageTestsForModulesWithRunner(c, []ModuleInfo{module}, true, nil, "", "coverage.txt", runner)
	require.NoError(t, err)
	require.Len(t, runner.testArgs, 1)
	assert.Contains(t, runner.testArgs[0], "-coverpkg=./...")
}
