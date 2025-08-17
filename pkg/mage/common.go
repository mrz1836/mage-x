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
	// Try to get from git
	if tag, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--abbrev=0"); err == nil {
		return strings.TrimSpace(tag)
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

	return "dev"
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

	if b == "dev" {
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
