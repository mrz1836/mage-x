// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Common helper functions used across multiple files

// getVersion returns the current version
func getVersion() string {
	// Check for release version override (used during release process)
	if releaseVersion := utils.GetEnvClean("MAGE_X_RELEASE_VERSION"); releaseVersion != "" {
		return releaseVersion
	}

	// Check for environment variable override
	if envVersion := utils.GetEnvClean("MAGE_X_VERSION"); envVersion != "" {
		return envVersion
	}

	// Try to get from git
	if gitVersion := getVersionFromGit(); gitVersion != "" {
		return gitVersion
	}

	// Try reading from version file
	fileOps := fileops.New()
	if content, err := fileOps.File.ReadFile("VERSION"); err == nil {
		return strings.TrimSpace(string(content))
	}

	// Try to get from config
	if config, err := GetConfig(); err == nil && config.Project.Version != "" {
		return config.Project.Version
	}

	return versionDev
}

// getModuleName returns the current module name
func getModuleName() (string, error) {
	return utils.GetModuleName()
}

// getDirSize calculates directory size
func getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// getCPUCount returns the number of CPU cores
func getCPUCount() int {
	return runtime.NumCPU()
}

// isNewer checks if version a is newer than version b
func isNewer(a, b string) bool {
	// Simple version comparison - in production use semver library
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	if b == versionDev {
		return true
	}

	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		var aNum, bNum int
		if _, err := fmt.Sscanf(aParts[i], "%d", &aNum); err != nil {
			// Non-numeric parts are treated as 0
			aNum = 0
		}
		if _, err := fmt.Sscanf(bParts[i], "%d", &bNum); err != nil {
			// Non-numeric parts are treated as 0
			bNum = 0
		}

		if aNum > bNum {
			return true
		} else if aNum < bNum {
			return false
		}
	}

	return len(aParts) > len(bParts)
}

// formatReleaseNotes formats release notes for display
func formatReleaseNotes(body string) string {
	lines := strings.Split(body, "\n")
	var formatted []string

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			formatted = append(formatted, "  "+line)
		}
	}

	return strings.Join(formatted, "\n")
}

// getVersionFromGit gets version from git, returning versionDev for development builds
func getVersionFromGit() string {
	tag, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return ""
	}
	tag = strings.TrimSpace(tag)

	// Check if we're exactly on the tag or if there are commits after
	if fullDesc, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--always", "--dirty"); err == nil {
		fullDesc = strings.TrimSpace(fullDesc)
		// If the full description is different from just the tag, we're in development
		if fullDesc != tag {
			utils.Debug("Git describe shows '%s' (not exactly on tag '%s') - using dev version", fullDesc, tag)
			return versionDev
		}
	}

	// Check if working directory is dirty
	if status, err := GetRunner().RunCmdOutput("git", "status", "--porcelain"); err == nil {
		if strings.TrimSpace(status) != "" {
			utils.Debug("Working directory has uncommitted changes - using dev version")
			return versionDev
		}
	}

	return tag
}
