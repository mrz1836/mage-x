package mage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBenchPassOutput(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		tags      string
		multiPass bool
		want      string
	}{
		{name: "SinglePassKeepsName", file: "cpu.prof", tags: "unit", want: "cpu.prof"},
		{name: "NoOutput", file: "", tags: "unit", multiPass: true, want: ""},
		{name: "BasePass", file: "cpu.prof", multiPass: true, want: "cpu.base.prof"},
		{name: "TagPass", file: "cpu.prof", tags: "unit", multiPass: true, want: "cpu.unit.prof"},
		{name: "CombinedTagSet", file: "cpu.prof", tags: "unit,e2e", multiPass: true, want: "cpu.unit_e2e.prof"},
		{name: "SpaceSeparatedTagSet", file: "cpu.prof", tags: "unit e2e", multiPass: true, want: "cpu.unit_e2e.prof"},
		{name: "NoExtension", file: "out/trace", tags: "unit", multiPass: true, want: "out/trace.unit"},
		{name: "Directory", file: "out/cpu.prof", tags: "unit", multiPass: true, want: "out/cpu.unit.prof"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, benchPassOutput(tt.file, tt.tags, tt.multiPass))
		})
	}
}

func TestBenchNamespaceAppliesBuildTags(t *testing.T) {
	perTag := &Config{Test: TestConfig{AutoDiscoverBuildTags: true, CombineBuildTags: false}}

	t.Run("TestTagsWithoutDiscovery", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, &Config{Test: TestConfig{Tags: "integration"}})

			require.NoError(t, Bench{}.DefaultWithArgs())

			require.Len(t, recorder.calls, 1)
			assert.Equal(t,
				[]string{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "integration", "-benchtime", "3s", "./..."},
				recorder.calls[0])
		})
	})

	t.Run("CombinedModeKeepsFileNames", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, &Config{Test: TestConfig{AutoDiscoverBuildTags: true, CombineBuildTags: true}})

			require.NoError(t, Bench{}.CPUWithArgs())

			assert.Equal(t, [][]string{
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "mage", "-benchtime", "3s", "-cpuprofile", "cpu.prof", "./..."},
				{"go", "tool", "pprof", "-top", "cpu.prof"},
			}, recorder.calls)
		})
	})

	t.Run("CPUPerTagModeSuffixesProfiles", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, perTag)

			require.NoError(t, Bench{}.CPUWithArgs())

			assert.Equal(t, [][]string{
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "3s", "-cpuprofile", "cpu.base.prof", "./..."},
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "mage", "-benchtime", "3s", "-cpuprofile", "cpu.mage.prof", "./..."},
				{"go", "tool", "pprof", "-top", "cpu.base.prof"},
				{"go", "tool", "pprof", "-top", "cpu.mage.prof"},
			}, recorder.calls)
		})
	})

	t.Run("MemPerTagModeSuffixesProfiles", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, perTag)

			require.NoError(t, Bench{}.MemWithArgs("profile=heap.out"))

			assert.Equal(t, [][]string{
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "3s", "-memprofile", "heap.base.out", "./..."},
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "mage", "-benchtime", "3s", "-memprofile", "heap.mage.out", "./..."},
				{"go", "tool", "pprof", "-top", "-alloc_space", "heap.base.out"},
				{"go", "tool", "pprof", "-top", "-alloc_space", "heap.mage.out"},
			}, recorder.calls)
		})
	})

	t.Run("SavePerTagModeAppendsEachPass", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, perTag)

			require.NoError(t, Bench{}.SaveWithArgs("output=results.txt"))

			assert.Equal(t, [][]string{
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "3s", "./..."},
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "mage", "-benchtime", "3s", "./..."},
			}, recorder.calls)

			saved, err := os.ReadFile("results.txt")
			require.NoError(t, err)
			assert.Equal(t, "# Build tags: none\n# Module: main\n\n# Build tags: mage\n# Module: main\n", string(saved))
		})
	})

	t.Run("TracePerTagModeSuffixesTraces", func(t *testing.T) {
		withCleanBenchEnv(t, func() {
			recorder := setupBuildTagModule(t, perTag)

			require.NoError(t, Bench{}.TraceWithArgs())

			assert.Equal(t, [][]string{
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-trace", "trace.base.out", "-benchtime", "3s", "./..."},
				{"go", "test", "-bench=.", "-benchmem", "-run=^$", "-tags", "mage", "-trace", "trace.mage.out", "-benchtime", "3s", "./..."},
			}, recorder.calls)
		})
	})
}
