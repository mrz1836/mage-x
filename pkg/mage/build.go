package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/cache"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Build namespace for build-related tasks
type Build mg.Namespace

var (
	cacheManager *cache.Manager //nolint:gochecknoglobals // Required for cache singleton
	cacheOnce    sync.Once      //nolint:gochecknoglobals // Required for singleton initialization
)

// initCacheManager initializes the cache manager if not already done
func initCacheManager() *cache.Manager {
	cacheOnce.Do(func() {
		config := cache.DefaultConfig()
		// TODO: Add cache configuration to Config struct if needed
		// For now, use default cache configuration

		// Check if cache is disabled via environment
		if os.Getenv("MAGE_CACHE_DISABLED") == "true" {
			config.Enabled = false
		}

		cacheManager = cache.NewManager(config)
		if cacheManager != nil {
			if err := cacheManager.Init(); err != nil {
				utils.Warn("Failed to initialize cache: %v", err)
				// Continue without caching
				config.Enabled = false
				cacheManager = cache.NewManager(config)
			}
		}
	})
	return cacheManager
}

// Default builds the application for the current platform
func (Build) Default() error {
	utils.Header("Building Application")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	cm := initCacheManager()

	binary := cfg.Project.Binary
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	outputPath := filepath.Join(cfg.Build.Output, binary)

	// Check if we have a cmd directory with main packages
	packagePath := ""
	if utils.DirExists("cmd/mage-init") {
		packagePath = "./cmd/mage-init"
	} else if utils.DirExists("cmd/example") {
		packagePath = "./cmd/example"
	} else if utils.FileExists("main.go") {
		packagePath = "."
	} else {
		// This is a library project, just verify compilation
		utils.Info("No main package found, building library packages for verification")
		outputPath = "/dev/null"
		packagePath = "./..."
	}

	// Generate cache hash for this build
	args := []string{"build"}
	args = append(args, buildFlags(cfg)...)
	args = append(args, "-o", outputPath)

	if cfg.Build.Verbose {
		args = append(args, "-v")
	}

	args = append(args, packagePath)

	ldflags := strings.Join(cfg.Build.LDFlags, " ")
	if ldflags == "" {
		// Use default ldflags for hash
		defaultLDFlags := []string{
			fmt.Sprintf("-X main.version=%s", getVersion()),
			fmt.Sprintf("-X main.commit=%s", getCommit()),
			fmt.Sprintf("-X main.date=%s", time.Now().Format(time.RFC3339)),
		}
		if !utils.GetEnvBool("DEBUG", false) {
			defaultLDFlags = append(defaultLDFlags, "-s", "-w")
		}
		ldflags = strings.Join(defaultLDFlags, " ")
	}

	// Find source files for cache hash
	sourceFiles, err := findSourceFiles()
	if err != nil {
		utils.Warn("Failed to find source files for caching: %v", err)
		sourceFiles = []string{"go.mod", "go.sum"}
	}

	configFiles := []string{"go.mod", "go.sum"}
	if utils.FileExists(".mage.yaml") {
		configFiles = append(configFiles, ".mage.yaml")
	}

	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	var buildHash string
	if cm != nil {
		var hashErr error
		buildHash, hashErr = cm.GenerateBuildHash(platform, ldflags, sourceFiles, configFiles)
		if hashErr != nil {
			utils.Warn("Failed to generate build hash: %v", hashErr)
			// Continue without caching
		}
	}

	if buildHash != "" && cm != nil {
		// Check cache first
		if buildResult, found := cm.GetBuildCache().GetBuildResult(buildHash); found {
			if buildResult.Success && utils.FileExists(buildResult.Binary) {
				utils.Success("Build cache hit! Using cached binary %s (built in %s)",
					buildResult.Binary,
					utils.FormatDuration(buildResult.Metrics.CompileTime))

				// Copy cached binary to expected output location if different
				if buildResult.Binary != outputPath {
					if err := utils.CopyFile(buildResult.Binary, outputPath); err != nil {
						utils.Warn("Failed to copy cached binary: %v", err)
						// Fall through to rebuild
					} else {
						return nil
					}
				} else {
					return nil
				}
			}
		}
	}

	// Cache miss or disabled - perform actual build
	start := time.Now()
	buildSuccess := true
	buildError := ""

	if err := GetRunner().RunCmd("go", args...); err != nil {
		buildSuccess = false
		buildError = err.Error()
	}

	buildDuration := time.Since(start)

	// Store result in cache if caching is enabled
	if buildHash != "" && cm != nil && cm.IsEnabled() {
		buildResult := &cache.BuildResult{
			Binary:     outputPath,
			Platform:   platform,
			BuildFlags: args,
			Success:    buildSuccess,
			Error:      buildError,
			Metrics: cache.BuildMetrics{
				CompileTime: buildDuration,
				BinarySize:  getBinarySize(outputPath),
			},
		}

		if err := cm.GetBuildCache().StoreBuildResult(buildHash, buildResult); err != nil {
			utils.Warn("Failed to store build result in cache: %v", err)
		}
	}

	if !buildSuccess {
		return fmt.Errorf("build failed: %s", buildError)
	}

	utils.Success("Built %s in %s", outputPath, utils.FormatDuration(buildDuration))
	return nil
}

// All builds for all configured platforms
func (b Build) All() error {
	utils.Header("Building for All Platforms")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	start := time.Now()
	errors := make(chan error, len(cfg.Build.Platforms))

	for _, platform := range cfg.Build.Platforms {
		go func(p string) {
			errors <- b.Platform(p)
		}(platform)
	}

	var buildErrors []string
	for range cfg.Build.Platforms {
		if err := <-errors; err != nil {
			buildErrors = append(buildErrors, err.Error())
		}
	}

	if len(buildErrors) > 0 {
		return fmt.Errorf("build errors:\n%s", strings.Join(buildErrors, "\n"))
	}

	utils.Success("Built all platforms in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Platform builds for a specific platform (e.g., "linux/amd64")
func (b Build) Platform(platform string) error {
	p, err := utils.ParsePlatform(platform)
	if err != nil {
		return err
	}

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	binary := fmt.Sprintf("%s-%s-%s%s",
		cfg.Project.Binary,
		p.OS,
		p.Arch,
		utils.GetBinaryExt(p))

	outputPath := filepath.Join(cfg.Build.Output, binary)

	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", p.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", p.Arch))

	args := []string{"build"}
	args = append(args, buildFlags(cfg)...)
	args = append(args, "-o", outputPath)

	// Add the package path to build
	if utils.DirExists("cmd/mage-init") {
		args = append(args, "./cmd/mage-init")
	} else if utils.DirExists("cmd/example") {
		args = append(args, "./cmd/example")
	} else if utils.FileExists("main.go") {
		args = append(args, ".")
	} else {
		// This is a library project, skip platform builds
		utils.Info("Skipping %s build for library project", platform)
		return nil
	}

	utils.Info("Building %s", platform)

	// Set environment for cross-compilation
	if err := os.Setenv("GOOS", p.OS); err != nil {
		return fmt.Errorf("failed to set GOOS: %w", err)
	}
	if err := os.Setenv("GOARCH", p.Arch); err != nil {
		return fmt.Errorf("failed to set GOARCH: %w", err)
	}
	defer func() {
		if err := os.Unsetenv("GOOS"); err != nil {
			utils.Warn("Failed to unset GOOS: %v", err)
		}
		if err := os.Unsetenv("GOARCH"); err != nil {
			utils.Warn("Failed to unset GOARCH: %v", err)
		}
	}()

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("build %s failed: %w", platform, err)
	}

	utils.Success("Built %s", outputPath)
	return nil
}

// Linux builds for Linux (amd64)
func (b Build) Linux() error {
	return b.Platform("linux/amd64")
}

// Darwin builds for macOS (amd64 and arm64)
func (b Build) Darwin() error {
	if err := b.Platform("darwin/amd64"); err != nil {
		return err
	}
	return b.Platform("darwin/arm64")
}

// Windows builds for Windows (amd64)
func (b Build) Windows() error {
	return b.Platform("windows/amd64")
}

// Docker builds a Docker image
func (Build) Docker() error {
	utils.Header("Building Docker Image")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	if !utils.CommandExists("docker") {
		return fmt.Errorf("docker command not found")
	}

	// Check if Dockerfile exists
	if _, err := os.Stat(cfg.Docker.Dockerfile); os.IsNotExist(err) {
		return fmt.Errorf("Dockerfile not found")
	}

	tag := fmt.Sprintf("%s/%s:%s",
		cfg.Docker.Registry,
		cfg.Docker.Repository,
		getVersion())

	args := []string{"build", "-t", tag}

	// Add build args
	for k, v := range cfg.Docker.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, "-f", cfg.Docker.Dockerfile, ".")

	return GetRunner().RunCmd("docker", args...)
}

// Clean removes build artifacts
func (Build) Clean() error {
	utils.Header("Cleaning Build Artifacts")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	// Clean output directory
	if cfg.Build.Output != "" {
		if err := utils.CleanDir(cfg.Build.Output); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Clean test cache
	if err := GetRunner().RunCmd("go", "clean", "-testcache"); err != nil {
		return fmt.Errorf("failed to clean test cache: %w", err)
	}

	// Clean build cache if requested
	if utils.GetEnvBool("CLEAN_CACHE", false) {
		utils.Info("Cleaning build cache")
		if err := GetRunner().RunCmd("go", "clean", "-cache"); err != nil {
			return fmt.Errorf("failed to clean build cache: %w", err)
		}
	}

	utils.Success("Build artifacts cleaned")
	return nil
}

// Install installs the binary to $GOPATH/bin
func (Build) Install() error {
	utils.Header("Installing Binary")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"install"}
	args = append(args, buildFlags(cfg)...)

	if cfg.Build.Verbose {
		args = append(args, "-v")
	}

	// Add the package path to install
	if utils.DirExists("cmd/mage-init") {
		args = append(args, "./cmd/mage-init")
	} else if utils.DirExists("cmd/example") {
		args = append(args, "./cmd/example")
	} else if utils.FileExists("main.go") {
		args = append(args, ".")
	} else {
		// This is a library project, install all packages
		args = append(args, "./...")
	}

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	utils.Success("Installed %s to %s", cfg.Project.Binary, filepath.Join(gopath, "bin"))
	return nil
}

// Generate runs go generate
func (Build) Generate() error {
	utils.Header("Running Code Generation")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"generate"}

	if cfg.Build.Verbose {
		args = append(args, "-v")
	}

	if len(cfg.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(cfg.Build.Tags, ","))
	}

	args = append(args, "./...")

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("generate failed: %w", err)
	}

	utils.Success("Code generation completed")
	return nil
}

// PreBuild pre-builds all packages to warm cache
func (Build) PreBuild() error {
	utils.Header("Pre-building Packages")

	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"build"}

	if cfg.Build.Verbose {
		args = append(args, "-v")
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("pre-build failed: %w", err)
	}

	utils.Success("Pre-build completed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Helper functions

// buildFlags returns common build flags
func buildFlags(cfg *Config) []string {
	var flags []string

	// Add tags
	if len(cfg.Build.Tags) > 0 {
		flags = append(flags, "-tags", strings.Join(cfg.Build.Tags, ","))
	}

	// Add ldflags
	if len(cfg.Build.LDFlags) > 0 {
		flags = append(flags, "-ldflags", strings.Join(cfg.Build.LDFlags, " "))
	} else {
		// Default ldflags
		ldflags := []string{
			fmt.Sprintf("-X main.version=%s", getVersion()),
			fmt.Sprintf("-X main.commit=%s", getCommit()),
			fmt.Sprintf("-X main.date=%s", time.Now().Format(time.RFC3339)),
		}

		// Add stripping flags for release builds
		if !utils.GetEnvBool("DEBUG", false) {
			ldflags = append(ldflags, "-s", "-w")
		}

		flags = append(flags, "-ldflags", strings.Join(ldflags, " "))
	}

	// Add trimpath
	if cfg.Build.TrimPath {
		flags = append(flags, "-trimpath")
	}

	// Add verbose flag
	if cfg.Build.Verbose {
		flags = append(flags, "-v")
	}

	// Add custom go flags
	if len(cfg.Build.GoFlags) > 0 {
		flags = append(flags, cfg.Build.GoFlags...)
	}

	return flags
}

// getVersion returns the current version

// getCommit returns the current git commit
func getCommit() string {
	if commit, err := GetRunner().RunCmdOutput("git", "rev-parse", "--short", "HEAD"); err == nil {
		return strings.TrimSpace(commit)
	}
	return "unknown"
}

// findSourceFiles finds Go source files for cache hashing
func findSourceFiles() ([]string, error) {
	files := []string{}

	// Add go.mod and go.sum
	files = append(files, "go.mod")
	if utils.FileExists("go.sum") {
		files = append(files, "go.sum")
	}

	// Find .go files in current directory and subdirectories
	goFiles, err := utils.FindFiles(".", "*.go")
	if err != nil {
		return files, err
	}

	files = append(files, goFiles...)
	return files, nil
}

// getBinarySize returns the size of a binary file
func getBinarySize(path string) int64 {
	if info, err := os.Stat(path); err == nil {
		return info.Size()
	}
	return 0
}
