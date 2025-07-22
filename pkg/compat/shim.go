package compat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Shim provides a compatibility shim for legacy magefiles
type Shim struct{}

// NewShim creates a new compatibility shim
func NewShim() *Shim {
	return &Shim{}
}

// Build provides legacy build targets
type Build mg.Namespace

func (Build) Default() error {
	return Global.adapter.WrapLegacyTarget("build:default", func() error {
		return Global.RunCmd("go", "build", "-o", getBinaryName(), ".")
	})()
}

func (Build) Linux() error {
	return Global.adapter.WrapLegacyTarget("build:linux", func() error {
		os.Setenv("GOOS", "linux")
		os.Setenv("GOARCH", "amd64")
		defer os.Unsetenv("GOOS")
		defer os.Unsetenv("GOARCH")
		return Global.RunCmd("go", "build", "-o", getBinaryName()+"-linux-amd64", ".")
	})()
}

func (Build) Windows() error {
	return Global.adapter.WrapLegacyTarget("build:windows", func() error {
		os.Setenv("GOOS", "windows")
		os.Setenv("GOARCH", "amd64")
		defer os.Unsetenv("GOOS")
		defer os.Unsetenv("GOARCH")
		return Global.RunCmd("go", "build", "-o", getBinaryName()+"-windows-amd64.exe", ".")
	})()
}

func (Build) Darwin() error {
	return Global.adapter.WrapLegacyTarget("build:darwin", func() error {
		os.Setenv("GOOS", "darwin")
		os.Setenv("GOARCH", "amd64")
		defer os.Unsetenv("GOOS")
		defer os.Unsetenv("GOARCH")
		return Global.RunCmd("go", "build", "-o", getBinaryName()+"-darwin-amd64", ".")
	})()
}

// Test provides legacy test targets
type Test mg.Namespace

func (Test) Default() error {
	return Global.adapter.WrapLegacyTarget("test:default", func() error {
		return Global.RunCmd("go", "test", "./...")
	})()
}

func (Test) Unit() error {
	return Global.adapter.WrapLegacyTarget("test:unit", func() error {
		return Global.RunCmd("go", "test", "-short", "./...")
	})()
}

func (Test) Integration() error {
	return Global.adapter.WrapLegacyTarget("test:integration", func() error {
		return Global.RunCmd("go", "test", "-run", "Integration", "./...")
	})()
}

func (Test) Coverage() error {
	return Global.adapter.WrapLegacyTarget("test:coverage", func() error {
		if err := Global.RunCmd("go", "test", "-coverprofile=coverage.out", "./..."); err != nil {
			return err
		}
		return Global.RunCmd("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	})()
}

// Lint provides legacy lint targets
type Lint mg.Namespace

func (Lint) Default() error {
	return Global.adapter.WrapLegacyTarget("lint:default", func() error {
		// Try golangci-lint first
		if commandExists("golangci-lint") {
			return Global.RunCmd("golangci-lint", "run")
		}
		// Fall back to go vet
		return Global.RunCmd("go", "vet", "./...")
	})()
}

func (Lint) Vet() error {
	return Global.adapter.WrapLegacyTarget("lint:vet", func() error {
		return Global.RunCmd("go", "vet", "./...")
	})()
}

func (Lint) Fmt() error {
	return Global.adapter.WrapLegacyTarget("lint:fmt", func() error {
		return Global.RunCmd("go", "fmt", "./...")
	})()
}

// Clean provides legacy clean targets
type Clean mg.Namespace

func (Clean) Default() error {
	return Global.adapter.WrapLegacyTarget("clean:default", func() error {
		patterns := []string{
			getBinaryName(),
			getBinaryName() + "-*",
			"coverage.out",
			"coverage.html",
			"dist/",
			"*.test",
			"*.out",
		}

		for _, pattern := range patterns {
			matches, _ := filepath.Glob(pattern)
			for _, match := range matches {
				utils.Info("Removing %s", match)
				os.RemoveAll(match)
			}
		}

		return nil
	})()
}

func (Clean) Cache() error {
	return Global.adapter.WrapLegacyTarget("clean:cache", func() error {
		return Global.RunCmd("go", "clean", "-cache")
	})()
}

func (Clean) Mod() error {
	return Global.adapter.WrapLegacyTarget("clean:mod", func() error {
		return Global.RunCmd("go", "clean", "-modcache")
	})()
}

// Docker provides legacy docker targets
type Docker mg.Namespace

func (Docker) Build() error {
	return Global.adapter.WrapLegacyTarget("docker:build", func() error {
		return Global.RunCmd("docker", "build", "-t", getDockerImage(), ".")
	})()
}

func (Docker) Push() error {
	return Global.adapter.WrapLegacyTarget("docker:push", func() error {
		return Global.RunCmd("docker", "push", getDockerImage())
	})()
}

func (Docker) Run() error {
	return Global.adapter.WrapLegacyTarget("docker:run", func() error {
		return Global.RunCmd("docker", "run", "--rm", getDockerImage())
	})()
}

// Deps provides legacy dependency targets
type Deps mg.Namespace

func (Deps) Download() error {
	return Global.adapter.WrapLegacyTarget("deps:download", func() error {
		return Global.RunCmd("go", "mod", "download")
	})()
}

func (Deps) Tidy() error {
	return Global.adapter.WrapLegacyTarget("deps:tidy", func() error {
		return Global.RunCmd("go", "mod", "tidy")
	})()
}

func (Deps) Update() error {
	return Global.adapter.WrapLegacyTarget("deps:update", func() error {
		return Global.RunCmd("go", "get", "-u", "./...")
	})()
}

func (Deps) Vendor() error {
	return Global.adapter.WrapLegacyTarget("deps:vendor", func() error {
		return Global.RunCmd("go", "mod", "vendor")
	})()
}

// Helper functions for legacy support

func getBinaryName() string {
	// Try to get from environment
	if name := os.Getenv("BINARY_NAME"); name != "" {
		return name
	}

	// Try to get from go.mod
	if data, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return filepath.Base(parts[1])
				}
			}
		}
	}

	// Default to current directory name
	dir, _ := os.Getwd()
	return filepath.Base(dir)
}

func getDockerImage() string {
	// Try to get from environment
	if image := os.Getenv("DOCKER_IMAGE"); image != "" {
		return image
	}

	// Default to binary name
	return getBinaryName() + ":latest"
}

func commandExists(cmd string) bool {
	_, err := utils.RunCmdOutput(cmd, "--version")
	return err == nil
}

// LegacyHelpers provides additional legacy helper functions
var LegacyHelpers = struct {
	// Sh runs a shell command (legacy sh function)
	Sh func(cmd string, args ...string) error

	// Must panics if error is not nil (legacy must function)
	Must func(err error)

	// MustRun runs a command and panics on error
	MustRun func(cmd string, args ...string)

	// Exists checks if file exists
	Exists func(path string) bool

	// Glob returns files matching pattern
	Glob func(pattern string) []string
}{
	Sh: func(cmd string, args ...string) error {
		return Global.RunCmd(cmd, args...)
	},

	Must: func(err error) {
		if err != nil {
			panic(err)
		}
	},

	MustRun: func(cmd string, args ...string) {
		if err := Global.RunCmd(cmd, args...); err != nil {
			panic(err)
		}
	},

	Exists: func(path string) bool {
		return Global.FileExists(path)
	},

	Glob: func(pattern string) []string {
		matches, _ := filepath.Glob(pattern)
		return matches
	},
}

// RegisterLegacyTarget allows manual registration of legacy targets
func RegisterLegacyTarget(name string, fn func() error) {
	legacyTargets[name] = Global.adapter.WrapLegacyTarget(name, fn)
}

// RegisterLegacyNamespace allows registration of legacy namespaces
func RegisterLegacyNamespace(name string, ns interface{}) {
	// This would use reflection to register all methods as targets
	// For now, it's a placeholder for future implementation
	utils.Info("Registered legacy namespace: %s", name)
}

// AutoMigrate attempts to automatically migrate legacy magefiles
func AutoMigrate() error {
	utils.Header("Auto-Migration")

	// Check if already migrated
	if !Global.adapter.legacyMode {
		utils.Success("No migration needed - already using new format!")
		return nil
	}

	// Create backup
	backupDir := fmt.Sprintf(".mage-backup-%d", os.Getpid())
	utils.Info("Creating backup in %s", backupDir)

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}

	// Backup legacy files
	legacyFiles := []string{
		"magefile.go",
		"mage.go",
		".mage.yml",
		"mage.yaml",
	}

	for _, file := range legacyFiles {
		if Global.FileExists(file) {
			data, err := Global.ReadFile(file)
			if err != nil {
				return err
			}

			backupPath := filepath.Join(backupDir, file)
			if err := Global.WriteFile(backupPath, data); err != nil {
				return err
			}

			utils.Info("Backed up %s", file)
		}
	}

	// Run migration
	return MigrateCommand()
}
