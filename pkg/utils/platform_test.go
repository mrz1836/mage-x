package utils

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWindows(t *testing.T) {
	t.Parallel()

	result := IsWindows()
	expected := runtime.GOOS == "windows"
	assert.Equal(t, expected, result)
}

func TestIsMac(t *testing.T) {
	t.Parallel()

	result := IsMac()
	expected := runtime.GOOS == "darwin"
	assert.Equal(t, expected, result)
}

func TestIsLinux(t *testing.T) {
	t.Parallel()

	result := IsLinux()
	expected := runtime.GOOS == "linux"
	assert.Equal(t, expected, result)
}

func TestGetShell(t *testing.T) {
	t.Parallel()

	shell, args := GetShell()

	if IsWindows() {
		assert.Equal(t, "cmd", shell)
		assert.Equal(t, []string{"/c"}, args)
	} else {
		assert.Equal(t, "sh", shell)
		assert.Equal(t, []string{"-c"}, args)
	}
}

func TestPlatformString(t *testing.T) {
	t.Parallel()

	t.Run("formats platform correctly", func(t *testing.T) {
		t.Parallel()
		p := Platform{OS: "linux", Arch: "amd64"}
		assert.Equal(t, "linux/amd64", p.String())
	})

	t.Run("handles darwin arm64", func(t *testing.T) {
		t.Parallel()
		p := Platform{OS: "darwin", Arch: "arm64"}
		assert.Equal(t, "darwin/arm64", p.String())
	})
}

func TestGetCurrentPlatform(t *testing.T) {
	t.Parallel()

	p := GetCurrentPlatform()
	assert.Equal(t, runtime.GOOS, p.OS)
	assert.Equal(t, runtime.GOARCH, p.Arch)
}

func TestParsePlatform(t *testing.T) {
	t.Parallel()

	t.Run("parses valid platform", func(t *testing.T) {
		t.Parallel()
		p, err := ParsePlatform("linux/amd64")
		require.NoError(t, err)
		assert.Equal(t, "linux", p.OS)
		assert.Equal(t, "amd64", p.Arch)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		t.Parallel()
		_, err := ParsePlatform("invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid platform format")
	})

	t.Run("returns error for empty parts", func(t *testing.T) {
		t.Parallel()
		_, err := ParsePlatform("/amd64")
		assert.Error(t, err)
	})

	t.Run("returns error for missing arch", func(t *testing.T) {
		t.Parallel()
		_, err := ParsePlatform("linux/")
		assert.Error(t, err)
	})
}

func TestGetBinaryExt(t *testing.T) {
	t.Parallel()

	t.Run("returns .exe for windows", func(t *testing.T) {
		t.Parallel()
		p := Platform{OS: "windows", Arch: "amd64"}
		assert.Equal(t, ".exe", GetBinaryExt(p))
	})

	t.Run("returns empty for linux", func(t *testing.T) {
		t.Parallel()
		p := Platform{OS: "linux", Arch: "amd64"}
		assert.Empty(t, GetBinaryExt(p))
	})

	t.Run("returns empty for darwin", func(t *testing.T) {
		t.Parallel()
		p := Platform{OS: "darwin", Arch: "arm64"}
		assert.Empty(t, GetBinaryExt(p))
	})
}
