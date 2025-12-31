package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDefaultPathResolver_EnsureDir_Errors tests error handling in EnsureDir.
func TestDefaultPathResolver_EnsureDir_Errors(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("creates_nested_directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3")

		err := resolver.EnsureDir(nestedPath)
		require.NoError(t, err)

		info, err := os.Stat(nestedPath)
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("file_exists_at_path", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file_not_dir")

		// Create a file at the path
		err := os.WriteFile(filePath, []byte("content"), 0o600)
		require.NoError(t, err)

		// EnsureDir should not error if path already exists (even if it's a file)
		// The function only creates if not exists
		err = resolver.EnsureDir(filePath)
		require.NoError(t, err) // Does not error because Stat succeeds

		// Verify it's still a file
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		require.False(t, info.IsDir(), "path should still be a file")
	})

	t.Run("already_exists_as_directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// tmpDir already exists
		err := resolver.EnsureDir(tmpDir)
		require.NoError(t, err)

		info, err := os.Stat(tmpDir)
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("custom_mode", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Windows does not support Unix file modes")
		}

		tmpDir := t.TempDir()
		customPath := filepath.Join(tmpDir, "custom_mode_dir")

		err := resolver.EnsureDirWithMode(customPath, 0o700)
		require.NoError(t, err)

		info, err := os.Stat(customPath)
		require.NoError(t, err)
		require.True(t, info.IsDir())
		// Check mode (mask with 0777 to ignore extra bits)
		actualMode := info.Mode().Perm() & 0o777
		require.Equal(t, os.FileMode(0o700), actualMode)
	})
}

// TestDefaultPathResolver_Symlinks tests symlink resolution.
func TestDefaultPathResolver_Symlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests require Unix-like system or elevated privileges on Windows")
	}

	t.Run("resolve_follows_symlink", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create target directory
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0o750)
		require.NoError(t, err)

		// Create symlink to target
		linkPath := filepath.Join(tmpDir, "link")
		err = os.Symlink(targetDir, linkPath)
		require.NoError(t, err)

		// Resolver with FollowSymlinks=true (default)
		resolver := NewDefaultPathResolver()
		resolved, err := resolver.Resolve(linkPath)
		require.NoError(t, err)

		// The resolved path should be the target, not the link
		expectedTarget, err := filepath.EvalSymlinks(targetDir)
		require.NoError(t, err)
		require.Equal(t, expectedTarget, resolved)
	})

	t.Run("resolve_without_follow_symlinks", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create target directory
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0o750)
		require.NoError(t, err)

		// Create symlink to target
		linkPath := filepath.Join(tmpDir, "link")
		err = os.Symlink(targetDir, linkPath)
		require.NoError(t, err)

		// Resolver with FollowSymlinks=false
		resolver := NewDefaultPathResolverWithOptions(PathOptions{
			FollowSymlinks: false,
			ResolveEnvVars: true,
		})
		resolved, err := resolver.Resolve(linkPath)
		require.NoError(t, err)

		// The resolved path should be the link path itself (absolute)
		absLinkPath, err := filepath.Abs(linkPath)
		require.NoError(t, err)
		require.Equal(t, absLinkPath, resolved)
	})

	t.Run("broken_symlink_follows", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create symlink to non-existent target
		linkPath := filepath.Join(tmpDir, "broken_link")
		nonExistentTarget := filepath.Join(tmpDir, "does_not_exist")
		err := os.Symlink(nonExistentTarget, linkPath)
		require.NoError(t, err)

		// Resolver with FollowSymlinks=true
		resolver := NewDefaultPathResolver()
		resolved, err := resolver.Resolve(linkPath)
		// EvalSymlinks fails for broken symlinks, so it falls back
		// to the original path and then makes it absolute
		require.NoError(t, err)
		// Should still return a valid absolute path
		require.True(t, filepath.IsAbs(resolved))
	})

	t.Run("nested_symlinks", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create target
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0o750)
		require.NoError(t, err)

		// Create first symlink
		link1 := filepath.Join(tmpDir, "link1")
		err = os.Symlink(targetDir, link1)
		require.NoError(t, err)

		// Create second symlink pointing to first
		link2 := filepath.Join(tmpDir, "link2")
		err = os.Symlink(link1, link2)
		require.NoError(t, err)

		resolver := NewDefaultPathResolver()
		resolved, err := resolver.Resolve(link2)
		require.NoError(t, err)

		// Should resolve all the way to the target
		expectedTarget, err := filepath.EvalSymlinks(targetDir)
		require.NoError(t, err)
		require.Equal(t, expectedTarget, resolved)
	})
}

// TestDefaultPathResolver_GOROOT_Fallbacks tests GOROOT resolution.
func TestDefaultPathResolver_GOROOT_Fallbacks(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("uses_env_var_when_set", func(t *testing.T) {
		t.Setenv("GOROOT", "/custom/goroot")
		got := resolver.GOROOT()
		require.Equal(t, "/custom/goroot", got)
	})

	t.Run("falls_back_to_go_env_command", func(t *testing.T) {
		// Save and unset GOROOT to force fallback
		originalGoroot := os.Getenv("GOROOT")
		require.NoError(t, os.Unsetenv("GOROOT"))
		t.Cleanup(func() {
			if originalGoroot != "" {
				require.NoError(t, os.Setenv("GOROOT", originalGoroot))
			}
		})

		got := resolver.GOROOT()
		// Should get a valid path from 'go env GOROOT'
		// If go is not installed, it will return empty string
		if got != "" {
			require.True(t, filepath.IsAbs(got), "GOROOT should be absolute path")
		}
	})
}

// TestDefaultPathResolver_Expand tests path expansion.
func TestDefaultPathResolver_Expand(t *testing.T) {
	t.Run("tilde_expansion", func(t *testing.T) {
		resolver := NewDefaultPathResolver()
		t.Setenv("HOME", "/home/testuser")

		expanded := resolver.Expand("~/documents")
		expected := filepath.Join("/home/testuser", "documents")
		require.Equal(t, expected, expanded)
	})

	t.Run("env_var_expansion", func(t *testing.T) {
		resolver := NewDefaultPathResolver()
		t.Setenv("TEST_EXPAND_PATH", "/test/path")

		expanded := resolver.Expand("${TEST_EXPAND_PATH}/subdir")
		require.Equal(t, "/test/path/subdir", expanded)
	})

	t.Run("no_expansion_when_disabled", func(t *testing.T) {
		resolver := NewDefaultPathResolverWithOptions(PathOptions{
			ResolveEnvVars: false,
		})
		t.Setenv("TEST_NO_EXPAND", "value")

		// Tilde and env vars should not be expanded
		input := "~/path/${TEST_NO_EXPAND}"
		expanded := resolver.Expand(input)
		// Note: tilde expansion is tied to ResolveEnvVars option
		require.Equal(t, input, expanded)
	})
}

// TestDefaultPathResolver_MakeAbsolute tests absolute path conversion.
func TestDefaultPathResolver_MakeAbsolute(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("already_absolute", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			absPath := "C:\\absolute\\path"
			got, err := resolver.MakeAbsolute(absPath)
			require.NoError(t, err)
			require.Equal(t, absPath, got)
		} else {
			absPath := "/absolute/path"
			got, err := resolver.MakeAbsolute(absPath)
			require.NoError(t, err)
			require.Equal(t, absPath, got)
		}
	})

	t.Run("relative_path", func(t *testing.T) {
		relativePath := "relative/path"
		got, err := resolver.MakeAbsolute(relativePath)
		require.NoError(t, err)
		require.True(t, filepath.IsAbs(got))
		require.Contains(t, got, "relative/path")
	})
}

// TestDefaultPathResolver_Home tests home directory resolution.
func TestDefaultPathResolver_Home(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("returns_HOME_when_set", func(t *testing.T) {
		t.Setenv("HOME", "/custom/home")
		got := resolver.Home()
		require.Equal(t, "/custom/home", got)
	})

	t.Run("windows_fallbacks", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows-specific test")
		}

		// Test USERPROFILE fallback - need to unset HOME first
		originalHome := os.Getenv("HOME")
		require.NoError(t, os.Unsetenv("HOME"))
		t.Setenv("USERPROFILE", "C:\\Users\\TestUser")
		t.Cleanup(func() {
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			}
		})

		got := resolver.Home()
		require.Equal(t, "C:\\Users\\TestUser", got)
	})
}

// TestDefaultPathResolverHomeFallbacks tests all Home() fallback paths.
func TestDefaultPathResolverHomeFallbacks(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("userprofile_fallback_when_home_unset", func(t *testing.T) {
		// Save original values
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")

		// Unset HOME, set USERPROFILE
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Setenv("USERPROFILE", "/mock/userprofile"))

		t.Cleanup(func() {
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			} else {
				require.NoError(t, os.Unsetenv("HOME"))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			} else {
				require.NoError(t, os.Unsetenv("USERPROFILE"))
			}
		})

		got := resolver.Home()
		require.Equal(t, "/mock/userprofile", got)
	})

	t.Run("homedrive_homepath_fallback", func(t *testing.T) {
		// Save original values
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")

		// Unset HOME and USERPROFILE, set HOMEDRIVE and HOMEPATH
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Setenv("HOMEDRIVE", "C:"))
		require.NoError(t, os.Setenv("HOMEPATH", "\\Users\\TestUser"))

		t.Cleanup(func() {
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			} else {
				require.NoError(t, os.Unsetenv("HOME"))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			} else {
				require.NoError(t, os.Unsetenv("USERPROFILE"))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			} else {
				require.NoError(t, os.Unsetenv("HOMEDRIVE"))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			} else {
				require.NoError(t, os.Unsetenv("HOMEPATH"))
			}
		})

		got := resolver.Home()
		require.Equal(t, "C:\\Users\\TestUser", got)
	})

	t.Run("homedrive_without_homepath_returns_empty", func(t *testing.T) {
		// Save original values
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")

		// Unset all home-related vars except HOMEDRIVE
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Setenv("HOMEDRIVE", "C:"))
		require.NoError(t, os.Unsetenv("HOMEPATH"))

		t.Cleanup(func() {
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			} else {
				require.NoError(t, os.Unsetenv("HOME"))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			} else {
				require.NoError(t, os.Unsetenv("USERPROFILE"))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			} else {
				require.NoError(t, os.Unsetenv("HOMEDRIVE"))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			} else {
				require.NoError(t, os.Unsetenv("HOMEPATH"))
			}
		})

		got := resolver.Home()
		require.Empty(t, got)
	})

	t.Run("all_home_vars_unset_returns_empty", func(t *testing.T) {
		// Save original values
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")

		// Unset all home-related vars
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Unsetenv("HOMEDRIVE"))
		require.NoError(t, os.Unsetenv("HOMEPATH"))

		t.Cleanup(func() {
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			}
		})

		got := resolver.Home()
		require.Empty(t, got)
	})
}

// TestDefaultPathResolver_XDG tests XDG base directory support.
func TestDefaultPathResolver_XDG(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("xdg_config_home", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		got := resolver.ConfigDir("myapp")
		require.Equal(t, "/custom/config/myapp", got)
	})

	t.Run("xdg_data_home", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		got := resolver.DataDir("myapp")
		require.Equal(t, "/custom/data/myapp", got)
	})

	t.Run("xdg_cache_home", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/custom/cache")
		got := resolver.CacheDir("myapp")
		require.Equal(t, "/custom/cache/myapp", got)
	})
}

// TestDefaultPathResolver_GOPATH tests GOPATH resolution.
func TestDefaultPathResolver_GOPATH(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("uses_env_var_when_set", func(t *testing.T) {
		t.Setenv("GOPATH", "/custom/gopath")
		got := resolver.GOPATH()
		require.Equal(t, "/custom/gopath", got)
	})

	t.Run("defaults_to_home_go", func(t *testing.T) {
		// Unset GOPATH to force fallback
		originalGopath := os.Getenv("GOPATH")
		require.NoError(t, os.Unsetenv("GOPATH"))
		t.Setenv("HOME", "/home/testuser")
		t.Cleanup(func() {
			if originalGopath != "" {
				require.NoError(t, os.Setenv("GOPATH", originalGopath))
			}
		})

		got := resolver.GOPATH()
		require.Equal(t, "/home/testuser/go", got)
	})
}

// TestDefaultPathResolver_GOCACHE tests GOCACHE resolution.
func TestDefaultPathResolver_GOCACHE(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("uses_env_var_when_set", func(t *testing.T) {
		t.Setenv("GOCACHE", "/custom/go-build")
		got := resolver.GOCACHE()
		require.Equal(t, "/custom/go-build", got)
	})
}

// TestDefaultPathResolverGOCACHEPlatformDefaults tests GOCACHE platform-specific defaults.
func TestDefaultPathResolverGOCACHEPlatformDefaults(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("darwin_default_path", func(t *testing.T) {
		if runtime.GOOS != goosDarwin {
			t.Skip("Darwin-specific test")
		}

		// Unset GOCACHE to force fallback
		originalGOCACHE := os.Getenv("GOCACHE")
		require.NoError(t, os.Unsetenv("GOCACHE"))
		t.Setenv("HOME", "/Users/testuser")

		t.Cleanup(func() {
			if originalGOCACHE != "" {
				require.NoError(t, os.Setenv("GOCACHE", originalGOCACHE))
			}
		})

		got := resolver.GOCACHE()
		require.Equal(t, "/Users/testuser/Library/Caches/go-build", got)
	})

	t.Run("linux_default_path", func(t *testing.T) {
		if runtime.GOOS != goosLinux {
			t.Skip("Linux-specific test")
		}

		// Unset GOCACHE to force fallback
		originalGOCACHE := os.Getenv("GOCACHE")
		require.NoError(t, os.Unsetenv("GOCACHE"))
		t.Setenv("HOME", "/home/testuser")

		t.Cleanup(func() {
			if originalGOCACHE != "" {
				require.NoError(t, os.Setenv("GOCACHE", originalGOCACHE))
			}
		})

		got := resolver.GOCACHE()
		require.Equal(t, "/home/testuser/.cache/go-build", got)
	})

	t.Run("no_home_returns_empty", func(t *testing.T) {
		// Save all home-related env vars
		originalGOCACHE := os.Getenv("GOCACHE")
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")
		originalLocalAppData := os.Getenv("LOCALAPPDATA")

		// Unset all to force empty home
		require.NoError(t, os.Unsetenv("GOCACHE"))
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Unsetenv("HOMEDRIVE"))
		require.NoError(t, os.Unsetenv("HOMEPATH"))
		require.NoError(t, os.Unsetenv("LOCALAPPDATA"))

		t.Cleanup(func() {
			if originalGOCACHE != "" {
				require.NoError(t, os.Setenv("GOCACHE", originalGOCACHE))
			}
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			}
			if originalLocalAppData != "" {
				require.NoError(t, os.Setenv("LOCALAPPDATA", originalLocalAppData))
			}
		})

		got := resolver.GOCACHE()
		require.Empty(t, got)
	})
}

// TestDefaultPathResolverConfigDirPlatformDefaults tests ConfigDir platform-specific defaults.
func TestDefaultPathResolverConfigDirPlatformDefaults(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("darwin_without_xdg", func(t *testing.T) {
		if runtime.GOOS != goosDarwin {
			t.Skip("Darwin-specific test")
		}

		// Unset XDG_CONFIG_HOME to force platform default
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
		t.Setenv("HOME", "/Users/testuser")

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		})

		got := resolver.ConfigDir("myapp")
		require.Equal(t, "/Users/testuser/Library/Application Support/myapp", got)
	})

	t.Run("linux_without_xdg", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Linux-specific test")
		}

		// Unset XDG_CONFIG_HOME to force platform default
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
		t.Setenv("HOME", "/home/testuser")

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
		})

		got := resolver.ConfigDir("myapp")
		require.Equal(t, "/home/testuser/.config/myapp", got)
	})

	t.Run("empty_home_returns_empty", func(t *testing.T) {
		// Save all home-related env vars
		originalXDG := os.Getenv("XDG_CONFIG_HOME")
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")

		// Unset all to force empty home
		require.NoError(t, os.Unsetenv("XDG_CONFIG_HOME"))
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Unsetenv("HOMEDRIVE"))
		require.NoError(t, os.Unsetenv("HOMEPATH"))

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CONFIG_HOME", originalXDG))
			}
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			}
		})

		got := resolver.ConfigDir("myapp")
		require.Empty(t, got)
	})
}

// TestDefaultPathResolverCacheDirPlatformDefaults tests CacheDir platform-specific defaults.
func TestDefaultPathResolverCacheDirPlatformDefaults(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("darwin_without_xdg", func(t *testing.T) {
		if runtime.GOOS != goosDarwin {
			t.Skip("Darwin-specific test")
		}

		// Unset XDG_CACHE_HOME to force platform default
		originalXDG := os.Getenv("XDG_CACHE_HOME")
		require.NoError(t, os.Unsetenv("XDG_CACHE_HOME"))
		t.Setenv("HOME", "/Users/testuser")

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CACHE_HOME", originalXDG))
			}
		})

		got := resolver.CacheDir("myapp")
		require.Equal(t, "/Users/testuser/Library/Caches/myapp", got)
	})

	t.Run("linux_without_xdg", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Linux-specific test")
		}

		// Unset XDG_CACHE_HOME to force platform default
		originalXDG := os.Getenv("XDG_CACHE_HOME")
		require.NoError(t, os.Unsetenv("XDG_CACHE_HOME"))
		t.Setenv("HOME", "/home/testuser")

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CACHE_HOME", originalXDG))
			}
		})

		got := resolver.CacheDir("myapp")
		require.Equal(t, "/home/testuser/.cache/myapp", got)
	})

	t.Run("empty_home_returns_empty", func(t *testing.T) {
		// Save all home-related env vars
		originalXDG := os.Getenv("XDG_CACHE_HOME")
		originalHome := os.Getenv("HOME")
		originalUserProfile := os.Getenv("USERPROFILE")
		originalHomeDrive := os.Getenv("HOMEDRIVE")
		originalHomePath := os.Getenv("HOMEPATH")

		// Unset all to force empty home
		require.NoError(t, os.Unsetenv("XDG_CACHE_HOME"))
		require.NoError(t, os.Unsetenv("HOME"))
		require.NoError(t, os.Unsetenv("USERPROFILE"))
		require.NoError(t, os.Unsetenv("HOMEDRIVE"))
		require.NoError(t, os.Unsetenv("HOMEPATH"))

		t.Cleanup(func() {
			if originalXDG != "" {
				require.NoError(t, os.Setenv("XDG_CACHE_HOME", originalXDG))
			}
			if originalHome != "" {
				require.NoError(t, os.Setenv("HOME", originalHome))
			}
			if originalUserProfile != "" {
				require.NoError(t, os.Setenv("USERPROFILE", originalUserProfile))
			}
			if originalHomeDrive != "" {
				require.NoError(t, os.Setenv("HOMEDRIVE", originalHomeDrive))
			}
			if originalHomePath != "" {
				require.NoError(t, os.Setenv("HOMEPATH", originalHomePath))
			}
		})

		got := resolver.CacheDir("myapp")
		require.Empty(t, got)
	})
}

// TestDefaultEnvironmentClear tests the Clear() method.
func TestDefaultEnvironmentClear(t *testing.T) {
	t.Run("clear_controlled_variables", func(t *testing.T) {
		// IMPORTANT: Save entire environment first since Clear() is destructive
		originalEnv := os.Environ()
		t.Cleanup(func() {
			// Restore all original environment variables
			for _, envVar := range originalEnv {
				parts := splitEnvVar(envVar)
				if len(parts) == 2 {
					if err := os.Setenv(parts[0], parts[1]); err != nil {
						t.Logf("cleanup: failed to restore env var %s: %v", parts[0], err)
					}
				}
			}
		})

		// Create a fresh environment with test variables
		env := NewDefaultEnvironment()

		// Set some test variables with unique prefix to avoid system conflict
		testVars := []string{
			"MAGE_CLEAR_TEST_VAR1",
			"MAGE_CLEAR_TEST_VAR2",
			"MAGE_CLEAR_TEST_VAR3",
		}

		for i, v := range testVars {
			require.NoError(t, env.Set(v, fmt.Sprintf("value%d", i)))
		}

		// Verify they exist
		for _, v := range testVars {
			require.True(t, env.Exists(v), "variable %s should exist before clear", v)
		}

		// Clear all env vars
		err := env.Clear()
		require.NoError(t, err)

		// Verify test vars are gone
		for _, v := range testVars {
			require.False(t, env.Exists(v), "variable %s should not exist after clear", v)
		}
	})
}

// splitEnvVar splits an environment variable string into key and value.
func splitEnvVar(envVar string) []string {
	for i := 0; i < len(envVar); i++ {
		if envVar[i] == '=' {
			return []string{envVar[:i], envVar[i+1:]}
		}
	}
	return []string{envVar}
}

// TestDefaultEnvironmentSetMultipleErrors tests SetMultiple error handling.
func TestDefaultEnvironmentSetMultipleErrors(t *testing.T) {
	t.Run("empty_map_succeeds", func(t *testing.T) {
		env := NewDefaultEnvironment()
		err := env.SetMultiple(map[string]string{})
		require.NoError(t, err)
	})

	t.Run("overwrite_false_fails_on_existing", func(t *testing.T) {
		testKey := "SET_MULTIPLE_TEST_EXISTING"
		require.NoError(t, os.Unsetenv(testKey))

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		// Create env that doesn't allow overwrite
		env := NewDefaultEnvironmentWithOptions(Options{
			AllowOverwrite: false,
		})

		// Set the first variable
		require.NoError(t, env.Set(testKey, "original"))

		// Try to set multiple including the existing one
		err := env.SetMultiple(map[string]string{
			testKey: "new_value",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to set")
		require.Contains(t, err.Error(), testKey)

		// Original value should be preserved
		require.Equal(t, "original", env.Get(testKey))
	})
}

// TestDefaultPathResolver_GOMODCACHE tests GOMODCACHE resolution.
func TestDefaultPathResolver_GOMODCACHE(t *testing.T) {
	resolver := NewDefaultPathResolver()

	t.Run("uses_env_var_when_set", func(t *testing.T) {
		t.Setenv("GOMODCACHE", "/custom/mod")
		got := resolver.GOMODCACHE()
		require.Equal(t, "/custom/mod", got)
	})

	t.Run("defaults_to_gopath_pkg_mod", func(t *testing.T) {
		// Unset GOMODCACHE to force fallback
		originalGomodcache := os.Getenv("GOMODCACHE")
		require.NoError(t, os.Unsetenv("GOMODCACHE"))
		t.Setenv("GOPATH", "/custom/gopath")
		t.Cleanup(func() {
			if originalGomodcache != "" {
				require.NoError(t, os.Setenv("GOMODCACHE", originalGomodcache))
			}
		})

		got := resolver.GOMODCACHE()
		require.Equal(t, "/custom/gopath/pkg/mod", got)
	})
}
