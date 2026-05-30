package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	cacheRoot, err := os.MkdirTemp("", "mage-package-test-cache-*")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create test cache root: %v\n", err)
		return 1
	}
	defer func() {
		if removeErr := os.RemoveAll(cacheRoot); removeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to remove test cache root: %v\n", removeErr)
		}
	}()

	cacheDirs := map[string]string{
		"GOCACHE":        filepath.Join(cacheRoot, "go-build"),
		"MAGEFILE_CACHE": filepath.Join(cacheRoot, "magefile"),
	}
	for key, dir := range cacheDirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to create %s directory: %v\n", key, err)
			return 1
		}
		if err := os.Setenv(key, dir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to set %s: %v\n", key, err)
			return 1
		}
	}
	if err := os.Setenv("GOTELEMETRY", "off"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to disable Go telemetry: %v\n", err)
		return 1
	}

	return m.Run()
}
