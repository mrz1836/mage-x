package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/cache"
	"github.com/mrz1836/mage-x/pkg/common/providers"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Build namespace for build-related tasks
type Build mg.Namespace

// Static errors for err113 compliance
var (
	ErrBuildFailedError   = errors.New("build failed")
	ErrBuildErrors        = errors.New("build errors")
	ErrDockerNotFound     = errors.New("docker command not found")
	ErrDockerfileNotFound = errors.New("dockerfile not found")
	ErrNoMainPackage      = errors.New("no main package found for binary build")
)

// CacheManagerProvider defines the interface for providing cache manager instances
type CacheManagerProvider interface {
	GetCacheManager() *cache.Manager
}

// DefaultCacheManagerProvider provides a thread-safe singleton cache manager using generic provider
type DefaultCacheManagerProvider struct {
	*providers.Provider[*cache.Manager]
}

// NewDefaultCacheManagerProvider creates a new default cache manager provider using generic framework
func NewDefaultCacheManagerProvider() *DefaultCacheManagerProvider {
	factory := func() *cache.Manager {
		config := cache.DefaultConfig()
		// Cache configuration uses default settings for now.
		// Future enhancement: integrate with main Config struct for customization.

		// Check if cache is disabled via environment
		if os.Getenv("MAGE_X_CACHE_DISABLED") == "true" {
			config.Enabled = false
		}

		manager := cache.NewManager(config)
		if manager != nil {
			if err := manager.Init(); err != nil {
				utils.Warn("Failed to initialize cache: %v", err)
				// Continue without caching
				config.Enabled = false
				manager = cache.NewManager(config)
			}
		}
		return manager
	}

	return &DefaultCacheManagerProvider{
		Provider: providers.NewProvider(factory),
	}
}

// GetCacheManager returns a cache manager instance using thread-safe singleton pattern
func (p *DefaultCacheManagerProvider) GetCacheManager() *cache.Manager {
	return p.Get()
}

// packageCacheManagerProvider provides a generic package-level cache manager provider using the generic framework
//
//nolint:gochecknoglobals // Required for package-level singleton access pattern
var packageCacheManagerProvider = providers.NewPackageProvider(func() CacheManagerProvider {
	return NewDefaultCacheManagerProvider()
})

// getCacheManagerInstance returns the cache manager using the generic package provider
func getCacheManagerInstance() *cache.Manager {
	return packageCacheManagerProvider.Get().GetCacheManager()
}

// setCacheManagerProvider sets a custom cache manager provider using the generic package provider
func setCacheManagerProvider(provider CacheManagerProvider) {
	packageCacheManagerProvider.Set(provider)
}

// SetCacheManagerProvider sets a custom cache manager provider for the Build namespace
// This allows for dependency injection and testing with mock providers
func SetCacheManagerProvider(provider CacheManagerProvider) {
	setCacheManagerProvider(provider)
}

// getCacheManager returns the cache manager for backward compatibility
func getCacheManager() *cache.Manager {
	return getCacheManagerInstance()
}

// Default builds the application for the current platform
func (b Build) Default() error {
	utils.Header("Building Application")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	buildCtx, err := b.createBuildContext(config)
	if err != nil {
		return err
	}

	buildHash := b.generateBuildHash(buildCtx)

	// Try to use cached build result
	if b.checkCachedBuild(buildHash, buildCtx.outputPath, buildCtx.cacheManager) {
		return nil
	}

	// Perform actual build
	return b.executeBuild(buildCtx, buildHash)
}

// buildContext holds all build-related configuration and state
type buildContext struct {
	cfg          *Config
	cacheManager *cache.Manager
	outputPath   string
	packagePath  string
	buildArgs    []string
	ldflags      string
	sourceFiles  []string
	configFiles  []string
	platform     string
}

// createBuildContext creates a build context with all configuration
func (b Build) createBuildContext(cfg *Config) (*buildContext, error) {
	cm := getCacheManager()
	outputPath := b.determineBuildOutput(cfg)

	// For binary builds, require a main package
	requireMain := outputPath != "" && !strings.HasSuffix(outputPath, "/...")
	packagePath, err := b.determinePackagePath(outputPath, requireMain)
	if err != nil {
		return nil, err
	}

	buildArgs := b.createBuildArgs(cfg, outputPath, packagePath)
	ldflags := b.determineLDFlags(cfg)
	sourceFiles := b.findSourceFiles()
	configFiles := b.getConfigFiles()
	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

	return &buildContext{
		cfg:          cfg,
		cacheManager: cm,
		outputPath:   outputPath,
		packagePath:  packagePath,
		buildArgs:    buildArgs,
		ldflags:      ldflags,
		sourceFiles:  sourceFiles,
		configFiles:  configFiles,
		platform:     platform,
	}, nil
}

// determineBuildOutput determines the output path for the binary
func (b Build) determineBuildOutput(cfg *Config) string {
	binary := cfg.Project.Binary
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	return filepath.Join(cfg.Build.Output, binary)
}

// determinePackagePath determines the package path to build
func (b Build) determinePackagePath(outputPath string, requireMain bool) (string, error) {
	// Check for magex first (primary binary)
	if utils.DirExists("cmd/magex") {
		return "./cmd/magex", nil
	}
	if utils.DirExists("cmd/mage-init") {
		return "./cmd/mage-init", nil
	}
	if utils.DirExists("cmd/example") {
		return "./cmd/example", nil
	}
	if utils.FileExists("main.go") {
		return ".", nil
	}

	// Check for any main package in cmd/ subdirectories
	if cmdPath := findMainInCmdDir(); cmdPath != "" {
		return cmdPath, nil
	}

	// If we're building to a specific binary output and require main, error
	if requireMain && outputPath != "" && !strings.Contains(outputPath, "...") {
		return "", ErrNoMainPackage
	}

	// This is a library project, just verify compilation
	utils.Info("No main package found, building library packages for verification")
	return "./...", nil
}

// createBuildArgs creates the build arguments
func (b Build) createBuildArgs(cfg *Config, outputPath, packagePath string) []string {
	args := []string{"build"}
	args = append(args, buildFlags(cfg)...)
	args = append(args, "-o", outputPath)

	if cfg.Build.Verbose {
		args = append(args, "-v")
	}

	args = append(args, packagePath)
	return args
}

// determineLDFlags determines the linker flags to use
func (b Build) determineLDFlags(cfg *Config) string {
	ldflags := strings.Join(cfg.Build.LDFlags, " ")
	if ldflags != "" {
		return ldflags
	}

	// Use default ldflags
	defaultLDFlags := []string{
		fmt.Sprintf("-X main.version=%s", getVersion()),
		fmt.Sprintf("-X main.commit=%s", getCommit()),
		fmt.Sprintf("-X main.buildDate=%s", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("-X main.buildTime=%s", time.Now().Format(time.RFC3339)),
	}
	if !utils.GetEnvBool("DEBUG", false) {
		defaultLDFlags = append(defaultLDFlags, "-s", "-w")
	}
	return strings.Join(defaultLDFlags, " ")
}

// findSourceFiles finds source files for cache hash
func (b Build) findSourceFiles() []string {
	sourceFiles, err := findSourceFiles()
	if err != nil {
		utils.Warn("Failed to find source files for caching: %v", err)
		return []string{"go.mod", "go.sum"}
	}
	return sourceFiles
}

// getConfigFiles gets configuration files for cache hash
func (b Build) getConfigFiles() []string {
	configFiles := []string{"go.mod", "go.sum"}
	if utils.FileExists(".mage.yaml") {
		configFiles = append(configFiles, ".mage.yaml")
	}
	return configFiles
}

// generateBuildHash generates a build hash for caching
func (b Build) generateBuildHash(ctx *buildContext) string {
	if ctx.cacheManager == nil {
		return ""
	}

	buildHash, err := ctx.cacheManager.GenerateBuildHash(
		ctx.platform, ctx.ldflags, ctx.sourceFiles, ctx.configFiles)
	if err != nil {
		utils.Warn("Failed to generate build hash: %v", err)
		return ""
	}
	return buildHash
}

// checkCachedBuild attempts to use a cached build result
func (b Build) checkCachedBuild(buildHash, outputPath string, cm *cache.Manager) bool {
	if buildHash == "" || cm == nil {
		return false
	}
	return b.tryUseCachedBuild(buildHash, outputPath, *cm)
}

// executeBuild performs the actual build and handles caching
func (b Build) executeBuild(ctx *buildContext, buildHash string) error {
	start := time.Now()
	buildSuccess := true
	buildError := ""

	if err := GetRunner().RunCmd("go", ctx.buildArgs...); err != nil {
		buildSuccess = false
		buildError = err.Error()
	}

	buildDuration := time.Since(start)

	// Store result in cache if caching is enabled
	b.storeBuildResult(ctx, buildHash, buildSuccess, buildError, buildDuration)

	if !buildSuccess {
		return fmt.Errorf("%w: %s", ErrBuildFailedError, buildError)
	}

	utils.Success("Built %s in %s", ctx.outputPath, utils.FormatDuration(buildDuration))
	return nil
}

// storeBuildResult stores the build result in cache if enabled
func (b Build) storeBuildResult(ctx *buildContext, buildHash string, success bool, buildError string, duration time.Duration) {
	if buildHash == "" || ctx.cacheManager == nil || !ctx.cacheManager.IsEnabled() {
		return
	}

	buildResult := &cache.BuildResult{
		Binary:     ctx.outputPath,
		Platform:   ctx.platform,
		BuildFlags: ctx.buildArgs,
		Success:    success,
		Error:      buildError,
		Metrics: cache.BuildMetrics{
			CompileTime: duration,
			BinarySize:  getBinarySize(ctx.outputPath),
		},
	}

	if err := ctx.cacheManager.GetBuildCache().StoreBuildResult(buildHash, buildResult); err != nil {
		utils.Warn("Failed to store build result in cache: %v", err)
	}
}

// tryUseCachedBuild attempts to use a cached build result and returns whether cache was used
func (b Build) tryUseCachedBuild(buildHash, outputPath string, cm cache.Manager) bool {
	buildResult, found := cm.GetBuildCache().GetBuildResult(buildHash)
	if !found {
		return false // No cached result
	}

	if !buildResult.Success || !utils.FileExists(buildResult.Binary) {
		return false // Invalid cached result
	}

	utils.Success("Build cache hit! Using cached binary %s (built in %s)",
		buildResult.Binary,
		utils.FormatDuration(buildResult.Metrics.CompileTime))

	// Copy cached binary to expected output location if different
	if buildResult.Binary == outputPath {
		return true // Cache used successfully
	}

	if err := utils.CopyFile(buildResult.Binary, outputPath); err != nil {
		utils.Warn("Failed to copy cached binary: %v", err)
		return false // Fall through to rebuild
	}

	return true // Cache used successfully
}

// All builds for all configured platforms
func (b Build) All() error {
	utils.Header("Building for All Platforms")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	start := time.Now()
	buildErrs := make(chan error, len(config.Build.Platforms))

	for _, platform := range config.Build.Platforms {
		go func(p string) {
			buildErrs <- b.Platform(p)
		}(platform)
	}

	var buildErrors []string
	for range config.Build.Platforms {
		if err := <-buildErrs; err != nil {
			buildErrors = append(buildErrors, err.Error())
		}
	}

	if len(buildErrors) > 0 {
		return fmt.Errorf("%w:\n%s", ErrBuildErrors, strings.Join(buildErrors, "\n"))
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	binary := fmt.Sprintf("%s-%s-%s%s",
		config.Project.Binary,
		p.OS,
		p.Arch,
		utils.GetBinaryExt(p))

	outputPath := filepath.Join(config.Build.Output, binary)

	args := []string{"build"}
	args = append(args, buildFlags(config)...)
	args = append(args, "-o", outputPath)

	// Add the package path to build
	if utils.DirExists("cmd/mage-init") {
		args = append(args, "./cmd/mage-init")
	} else if utils.DirExists("cmd/example") {
		args = append(args, "./cmd/example")
	} else if utils.FileExists("main.go") {
		args = append(args, ".")
	} else if cmdPath := findMainInCmdDir(); cmdPath != "" {
		args = append(args, cmdPath)
	} else {
		// No main package found for platform-specific binary build
		return ErrNoMainPackage
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	if !utils.CommandExists("docker") {
		return ErrDockerNotFound
	}

	// Check if Dockerfile exists
	if _, err := os.Stat(config.Docker.Dockerfile); os.IsNotExist(err) {
		return ErrDockerfileNotFound
	}

	tag := fmt.Sprintf("%s/%s:%s",
		config.Docker.Registry,
		config.Docker.Repository,
		getVersion())

	args := []string{"build", "-t", tag}

	// Add build args
	for k, v := range config.Docker.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, "-f", config.Docker.Dockerfile, ".")

	return GetRunner().RunCmd("docker", args...)
}

// Clean removes build artifacts
func (Build) Clean() error {
	utils.Header("Cleaning Build Artifacts")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Clean output directory
	if config.Build.Output != "" {
		if err := utils.CleanDir(config.Build.Output); err != nil {
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"install"}
	args = append(args, buildFlags(config)...)

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	// Add the package path to install
	if utils.DirExists("cmd/mage-init") {
		args = append(args, "./cmd/mage-init")
	} else if utils.DirExists("cmd/example") {
		args = append(args, "./cmd/example")
	} else if utils.FileExists("main.go") {
		args = append(args, ".")
	} else if cmdPath := findMainInCmdDir(); cmdPath != "" {
		args = append(args, cmdPath)
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

	utils.Success("Installed %s to %s", config.Project.Binary, filepath.Join(gopath, "bin"))
	return nil
}

// Generate runs go generate
func (Build) Generate() error {
	utils.Header("Running Code Generation")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"generate"}

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	if len(config.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(config.Build.Tags, ","))
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"build"}

	if config.Build.Verbose {
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
			fmt.Sprintf("-X main.buildDate=%s", time.Now().Format(time.RFC3339)),
			fmt.Sprintf("-X main.buildTime=%s", time.Now().Format(time.RFC3339)),
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

// findMainInCmdDir searches for any main package in cmd/ subdirectories
func findMainInCmdDir() string {
	if !utils.DirExists("cmd") {
		return ""
	}

	// Read the cmd directory
	entries, err := os.ReadDir("cmd")
	if err != nil {
		return ""
	}

	// Check each subdirectory for a main.go file
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subdir := filepath.Join("cmd", entry.Name())
		mainFile := filepath.Join(subdir, "main.go")

		if utils.FileExists(mainFile) {
			// Verify it's actually a main package by checking the package declaration
			if isMainPackage(mainFile) {
				return "./" + subdir
			}
		}
	}

	return ""
}

// isMainPackage checks if a Go file contains a main package declaration
func isMainPackage(filePath string) bool {
	// Validate the file path to satisfy gosec G304
	// The path is constructed by findMainInCmdDir and should be safe
	// but we add validation to ensure it's a .go file in the expected location
	if !strings.HasSuffix(filePath, ".go") || !strings.Contains(filePath, "cmd/") {
		return false
	}

	content, err := os.ReadFile(filepath.Clean(filePath)) // #nosec G304 - path is validated above
	if err != nil {
		return false
	}

	// Simple check for "package main" at the beginning of the file
	// This is a basic implementation - could be enhanced with proper Go parsing
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") {
			continue // Skip empty lines and comments
		}
		return strings.HasPrefix(line, "package main")
	}
	return false
}
