package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
	ErrBuildFailedError      = errors.New("build failed")
	ErrBuildErrors           = errors.New("build errors")
	ErrNoMainPackage         = errors.New("no main package found for binary build")
	ErrInvalidMainPath       = errors.New("configured main path is invalid")
	ErrNoPackagesInWorkspace = errors.New("no packages found in workspace modules")
	ErrNotInWorkspaceMode    = errors.New("buildWorkspaceModules called but not in workspace mode")
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
		if os.Getenv("MAGE_X_CACHE_DISABLED") == trueValue {
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
	packagePath, err := b.determinePackagePath(cfg, outputPath, requireMain)
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
	if runtime.GOOS == OSWindows {
		binary += ".exe"
	}
	return filepath.Join(cfg.Build.Output, binary)
}

// determinePackagePath determines the package path to build
func (b Build) determinePackagePath(cfg *Config, outputPath string, requireMain bool) (string, error) {
	// Check for configured main path first
	if cfg.Project.Main != "" {
		return b.validateConfiguredMainPath(cfg.Project.Main)
	}

	return b.findPackagePath(outputPath, requireMain)
}

// validateConfiguredMainPath validates and returns the configured main path
func (b Build) validateConfiguredMainPath(configuredMain string) (string, error) {
	mainPath := b.normalizeMainPath(configuredMain)

	if b.isValidMainPath(mainPath) {
		return mainPath, nil
	}

	return "", fmt.Errorf("%w: %s does not exist or does not contain a valid main package", ErrInvalidMainPath, configuredMain)
}

// normalizeMainPath ensures the main path is properly formatted
func (b Build) normalizeMainPath(mainPath string) string {
	// Ensure the path is relative and clean
	if !strings.HasPrefix(mainPath, "./") && !strings.HasPrefix(mainPath, "/") {
		mainPath = "./" + mainPath
	}

	// Extract the directory path from the main.go file path
	if strings.HasSuffix(mainPath, "/main.go") {
		return strings.TrimSuffix(mainPath, "/main.go")
	}
	if strings.HasSuffix(mainPath, "main.go") {
		return filepath.Dir(mainPath)
	}

	return mainPath
}

// isValidMainPath checks if the main path exists and contains a valid main package
func (b Build) isValidMainPath(mainPath string) bool {
	dirPath := strings.TrimPrefix(mainPath, "./")
	if !utils.DirExists(dirPath) {
		return false
	}

	mainFile := filepath.Join(dirPath, "main.go")
	return utils.FileExists(mainFile) && isMainPackage(mainFile)
}

// findPackagePath finds the package path using fallback logic
func (b Build) findPackagePath(outputPath string, requireMain bool) (string, error) {
	// Check for magex first (primary binary)
	if utils.DirExists("cmd/magex") {
		return "./cmd/magex", nil
	}
	if utils.DirExists("cmd/mage-init") {
		return "./cmd/mage-init", nil
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

	// Determine the package path using the same logic as the main build
	packagePath, err := b.determinePackagePath(config, outputPath, true)
	if err != nil {
		return err
	}
	args = append(args, packagePath)

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
func (b Build) Install() error {
	utils.Header("Installing Binary")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Show version information
	version := getVersion()
	if version == versionDev {
		utils.Warn("⚠️  Repository has uncommitted changes - building with version 'dev'")
	} else {
		utils.Info("Building with version: %s", version)
	}

	args := []string{"install"}
	args = append(args, buildFlags(config)...)

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	// Determine the package path using the same logic as the main build
	packagePath, err := b.determinePackagePath(config, "", false)
	if err != nil {
		return err
	}
	args = append(args, packagePath)

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

// Dev builds and installs the binary with forced dev version for local development
func (b Build) Dev() error {
	utils.Header("Building Development Version")
	utils.Info("🔧 Building development version (forced 'dev')")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Force dev version by setting environment variable
	if setErr := os.Setenv("MAGE_X_VERSION", versionDev); setErr != nil {
		return fmt.Errorf("failed to set dev version: %w", setErr)
	}
	defer func() {
		if unsetErr := os.Unsetenv("MAGE_X_VERSION"); unsetErr != nil {
			utils.Error("Failed to unset MAGE_X_VERSION: %v", unsetErr)
		}
	}()

	args := []string{"install"}
	args = append(args, buildFlags(config)...)

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	// Determine the package path using the same logic as the main build
	packagePath, err := b.determinePackagePath(config, "", false)
	if err != nil {
		return err
	}
	args = append(args, packagePath)

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("dev build failed: %w", err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	utils.Success("Installed development build of %s to %s", config.Project.Binary, filepath.Join(gopath, "bin"))
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
	return Build{}.PreBuildWithArgs()
}

// PreBuildWithArgs pre-builds all packages to warm cache with configurable strategies and parallel execution
func (b Build) PreBuildWithArgs(argsList ...string) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Parse command-line parameters from os.Args
	// Find arguments after the target name
	var targetArgs []string
	for i, arg := range os.Args {
		if strings.Contains(arg, "build:prebuild") {
			targetArgs = os.Args[i+1:]
			break
		}
	}
	params := utils.ParseParams(targetArgs)

	// Parse strategy parameters with environment variable fallback
	strategy := utils.GetParam(params, "strategy", os.Getenv("MAGE_X_BUILD_STRATEGY"))
	if strategy == "" {
		strategy = config.Build.PreBuild.Strategy
	}
	if strategy == "" {
		strategy = "smart" // Default to smart strategy
	}

	// Parse batch parameters
	batchSizeStr := utils.GetParam(params, "batch", os.Getenv("MAGE_X_BUILD_BATCH_SIZE"))
	if batchSizeStr == "" && config.Build.PreBuild.BatchSize > 0 {
		batchSizeStr = fmt.Sprintf("%d", config.Build.PreBuild.BatchSize)
	}
	batchSize := 10 // Default
	if batchSizeStr != "" {
		if bs, err := strconv.Atoi(batchSizeStr); err == nil && bs > 0 {
			batchSize = bs
		}
	}

	// Parse delay parameter
	delayStr := utils.GetParam(params, "delay", os.Getenv("MAGE_X_BUILD_BATCH_DELAY_MS"))
	if delayStr == "" && config.Build.PreBuild.BatchDelay > 0 {
		delayStr = fmt.Sprintf("%d", config.Build.PreBuild.BatchDelay)
	}
	delayMs := 0 // Default: no delay
	if delayStr != "" {
		if d, err := strconv.Atoi(delayStr); err == nil && d >= 0 {
			delayMs = d
		}
	}

	// Parse memory limit
	memoryLimit := utils.GetParam(params, "memory_limit", os.Getenv("MAGE_X_BUILD_MEMORY_LIMIT"))
	if memoryLimit == "" {
		memoryLimit = config.Build.PreBuild.MemoryLimit
	}

	// Parse exclusion pattern
	exclude := utils.GetParam(params, "exclude", os.Getenv("MAGE_X_BUILD_EXCLUDE_PATTERN"))
	if exclude == "" {
		exclude = config.Build.PreBuild.Exclude
	}

	// Parse priority pattern (currently unused but preserved for future use)
	_ = utils.GetParam(params, "priority", os.Getenv("MAGE_X_BUILD_PRIORITY_PATTERN"))

	// Parse verbose flag
	verbose := utils.IsParamTrue(params, "verbose") ||
		config.Build.Verbose ||
		config.Build.PreBuild.Verbose ||
		os.Getenv("MAGE_X_BUILD_VERBOSE") == trueValue

	// Parse parallelism (maintain backward compatibility)
	parallelism := utils.GetParam(params, "parallel", "")
	if parallelism == "" {
		parallelism = utils.GetParam(params, "p", "")
	}
	if parallelism == "" && config.Build.Parallel > 0 {
		parallelism = fmt.Sprintf("%d", config.Build.Parallel)
	}

	// Parse mains-only flag for mains-first strategy
	mainsOnly := utils.IsParamTrue(params, "mains-only") ||
		utils.IsParamTrue(params, "mains_only")

	// Apply memory limit if specified
	if memoryLimit != "" && memoryLimit != string(CIFormatAuto) {
		strategy = b.applyMemoryLimit(memoryLimit, strategy)
	}

	// Execute the selected strategy
	start := time.Now()
	var buildErr error

	switch strings.ToLower(strategy) {
	case "incremental":
		buildErr = b.buildIncremental(batchSize, delayMs, exclude, verbose, parallelism)

	case "mains-first", "mains_first", "mainsfirst":
		buildErr = b.buildMainsFirst(batchSize, mainsOnly, exclude, verbose, parallelism)

	case "smart", "auto":
		buildErr = b.buildSmart(exclude, verbose, parallelism)

	case "full", "traditional", "legacy":
		buildErr = b.buildFull(parallelism, config.Build.Verbose, exclude)

	default:
		utils.Warn("Unknown strategy %q, using smart strategy", strategy)
		buildErr = b.buildSmart(exclude, verbose, parallelism)
	}

	if buildErr != nil {
		return buildErr
	}

	utils.Success("Pre-build completed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Package discovery and batching utilities

// getWorkspaceModuleDirs returns workspace module directories if go.work exists.
// Returns nil and false if not in workspace mode or if discovery fails.
func (b Build) getWorkspaceModuleDirs() ([]string, bool) {
	if _, err := os.Stat("go.work"); os.IsNotExist(err) {
		return nil, false // No workspace file
	}

	// Get workspace module directories using go list -m
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{.Dir}}")
	if err != nil {
		utils.Warn("Failed to list workspace modules: %v", err)
		return nil, false
	}

	var modules []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			modules = append(modules, line)
		}
	}

	if len(modules) > 0 {
		utils.Debug("Workspace mode detected with %d modules", len(modules))
	}

	return modules, len(modules) > 0
}

// shouldExcludeModuleByName checks if a module name is in the exclusion list.
// This allows modules like "magefiles" to be excluded from prebuild using the
// same MAGE_X_TEST_EXCLUDE_MODULES configuration used for tests.
func shouldExcludeModuleByName(moduleName string, config *Config) bool {
	if config == nil || len(config.Test.ExcludeModules) == 0 {
		return false
	}
	for _, excludedName := range config.Test.ExcludeModules {
		if moduleName == excludedName {
			return true
		}
	}
	return false
}

// buildWorkspaceModules builds all packages in each workspace module by running
// go build ./... from within each module directory. This avoids workspace validation
// errors that occur when running from the repository root with modules that exist
// but aren't listed in go.work (e.g., magefiles with build constraints).
// When exclude is provided, it filters packages within each module using the exclude pattern.
func (b Build) buildWorkspaceModules(verbose bool, parallelism, exclude string) error {
	moduleDirs, isWorkspace := b.getWorkspaceModuleDirs()
	if !isWorkspace {
		return ErrNotInWorkspaceMode
	}

	// Load config to check exclusion list
	config, err := GetConfig()
	if err != nil {
		utils.Warn("Failed to load config for exclusion check: %v", err)
		// Continue without filtering - don't fail the build
		config = nil
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	for _, modDir := range moduleDirs {
		modName := filepath.Base(modDir)

		// Check if module should be excluded by name (using config.Test.ExcludeModules)
		if shouldExcludeModuleByName(modName, config) {
			utils.Info("Skipping module %s (excluded from prebuild)", modName)
			continue
		}

		// Also check if module matches the exclude pattern
		if exclude != "" && strings.Contains(modName, exclude) {
			utils.Info("Skipping module %s (matches exclude pattern: %s)", modName, exclude)
			continue
		}

		utils.Info("Building packages in module: %s", modName)

		if err := os.Chdir(modDir); err != nil {
			utils.Warn("Failed to change to module directory %s: %v", modDir, err)
			continue
		}

		args := []string{"build"}
		if parallelism != "" {
			args = append(args, "-p", parallelism)
		}
		if verbose {
			args = append(args, "-v")
		}

		// If exclude pattern provided, discover and filter packages within this module
		if exclude != "" {
			packages, discoverErr := b.discoverPackages(exclude)
			if discoverErr != nil {
				utils.Warn("Failed to discover packages in %s: %v", modDir, discoverErr)
				// Fall back to ./... without filtering
				args = append(args, "./...")
			} else if len(packages) == 0 {
				utils.Info("No packages to build in %s after filtering", modName)
				continue
			} else {
				args = append(args, packages...)
			}
		} else {
			args = append(args, "./...")
		}

		if err := GetRunner().RunCmd("go", args...); err != nil {
			// Change back to original directory before returning error
			if chErr := os.Chdir(originalDir); chErr != nil {
				utils.Warn("Failed to restore original directory: %v", chErr)
			}
			return fmt.Errorf("failed to build module %s: %w", modName, err)
		}
	}

	// Restore original directory
	if err := os.Chdir(originalDir); err != nil {
		return fmt.Errorf("failed to restore original directory: %w", err)
	}

	return nil
}

// discoverPackages returns a list of all Go packages in the project.
// In workspace mode (go.work exists), it discovers packages from each workspace
// module to avoid errors from modules that exist but aren't in the workspace.
func (b Build) discoverPackages(exclude string) ([]string, error) {
	var packages []string
	var err error

	// In workspace mode, list packages per workspace module to avoid errors
	// from modules that exist but aren't listed in go.work (e.g., magefiles)
	if moduleDirs, isWorkspace := b.getWorkspaceModuleDirs(); isWorkspace {
		utils.Debug("Discovering packages from %d workspace modules", len(moduleDirs))
		for _, modDir := range moduleDirs {
			// Build pattern relative to module directory
			modPattern := filepath.Join(modDir, "...")
			pkgs, listErr := utils.GoList(modPattern)
			if listErr != nil {
				// Skip modules that fail (e.g., modules with build constraints like magefiles)
				utils.Debug("Skipping module %s: %v", modDir, listErr)
				continue
			}
			packages = append(packages, pkgs...)
		}
		if len(packages) == 0 {
			return nil, ErrNoPackagesInWorkspace
		}
	} else {
		// Standard mode: list all packages
		packages, err = utils.GoList("./...")
		if err != nil {
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}
	}

	// Filter out packages
	filtered := make([]string, 0, len(packages))
	for _, pkg := range packages {
		// Skip the root package if it's just a magefile (has build constraints that exclude it)
		if pkg == "github.com/mrz1836/mage-x" {
			continue
		}

		// Skip test packages that typically only have test files
		if strings.Contains(pkg, "/tests/") || strings.Contains(pkg, "/test/") {
			continue
		}

		// Apply exclusion filter if provided
		if exclude != "" {
			matched, err := filepath.Match(exclude, pkg)
			// Exclude if glob matched, or always try substring match as fallback
			if (err == nil && matched) || strings.Contains(pkg, exclude) {
				continue
			}
		}

		filtered = append(filtered, pkg)
	}

	return filtered, nil
}

// findMainPackages identifies packages containing main functions.
// In workspace mode, it searches each workspace module separately.
func (b Build) findMainPackages() ([]string, error) {
	var mainPackages []string

	// Determine patterns to search based on workspace mode
	patterns := []string{"./..."}
	if moduleDirs, isWorkspace := b.getWorkspaceModuleDirs(); isWorkspace {
		patterns = make([]string, 0, len(moduleDirs))
		for _, modDir := range moduleDirs {
			patterns = append(patterns, filepath.Join(modDir, "..."))
		}
		utils.Debug("Searching for main packages in %d workspace modules", len(patterns))
	}

	// Search each pattern for main packages
	for _, pattern := range patterns {
		output, err := GetRunner().RunCmdOutput("go", "list", "-json", pattern)
		if err != nil {
			// In workspace mode, skip modules that fail (e.g., build constraints)
			if len(patterns) > 1 {
				utils.Debug("Skipping pattern %s: %v", pattern, err)
				continue
			}
			return nil, fmt.Errorf("failed to list packages: %w", err)
		}

		lines := strings.Split(output, "\n{")
		for _, line := range lines {
			if !strings.HasPrefix(line, "{") {
				line = "{" + line
			}

			// Check for main package indicator and extract import path
			if pkg := extractMainPackageFromLine(line); pkg != "" {
				mainPackages = append(mainPackages, pkg)
			}
		}
	}

	return mainPackages, nil
}

// splitIntoBatches divides a slice of packages into smaller batches
func (b Build) splitIntoBatches(packages []string, batchSize int) [][]string {
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	var batches [][]string
	for i := 0; i < len(packages); i += batchSize {
		end := i + batchSize
		if end > len(packages) {
			end = len(packages)
		}
		batches = append(batches, packages[i:end])
	}

	return batches
}

// buildPackageBatch builds a single batch of packages (used in non-workspace mode)
func (b Build) buildPackageBatch(packages []string, parallelism string, verbose bool) error {
	args := []string{"build"}

	if parallelism != "" {
		args = append(args, "-p", parallelism)
	}

	if verbose {
		args = append(args, "-v")
	}

	args = append(args, packages...)

	return GetRunner().RunCmd("go", args...)
}

// Build strategies

// buildIncremental builds packages in small batches to manage memory usage
func (b Build) buildIncremental(batchSize, delayMs int, exclude string, verbose bool, parallelism string) error {
	utils.Header("Pre-building Packages (Incremental Strategy)")

	// In workspace mode, use per-module building to avoid validation issues
	if _, isWorkspace := b.getWorkspaceModuleDirs(); isWorkspace {
		utils.Info("Workspace mode detected - building each module separately")
		return b.buildWorkspaceModules(verbose, parallelism, exclude)
	}

	// Discover packages
	packages, err := b.discoverPackages(exclude)
	if err != nil {
		return err
	}

	if len(packages) == 0 {
		utils.Info("No packages to build")
		return nil
	}

	// Set default batch size if not specified
	if batchSize <= 0 {
		batchSize = 10
	}

	// Split into batches
	batches := b.splitIntoBatches(packages, batchSize)
	totalPackages := len(packages)

	utils.Info("Building %d packages in %d batches (batch size: %d)", totalPackages, len(batches), batchSize)

	// Build each batch
	start := time.Now()
	for i, batch := range batches {
		batchStart := time.Now()

		if verbose {
			utils.Info("Building batch %d/%d (%d packages)", i+1, len(batches), len(batch))
		}

		if err := b.buildPackageBatch(batch, parallelism, false); err != nil {
			return fmt.Errorf("batch %d failed: %w", i+1, err)
		}

		if verbose {
			utils.Success("Batch %d completed in %s", i+1, utils.FormatDuration(time.Since(batchStart)))
		}

		// Add delay between batches if specified (except after last batch)
		if delayMs > 0 && i < len(batches)-1 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	utils.Success("Incremental build completed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// buildMainsFirst builds main packages first, then other packages
func (b Build) buildMainsFirst(batchSize int, mainsOnly bool, exclude string, verbose bool, parallelism string) error {
	utils.Header("Pre-building Packages (Mains-First Strategy)")

	// Phase 1: Build main packages
	utils.Info("Phase 1: Building main packages...")
	mainPackages, err := b.findMainPackages()
	if err != nil {
		utils.Warn("Failed to identify main packages: %v", err)
		mainPackages = []string{}
	}

	if len(mainPackages) > 0 {
		utils.Info("Found %d main packages", len(mainPackages))

		// Build main packages (these typically pull in most dependencies)
		if buildErr := b.buildPackageBatch(mainPackages, parallelism, verbose); buildErr != nil {
			return fmt.Errorf("failed to build main packages: %w", buildErr)
		}

		utils.Success("Main packages built successfully")
	} else {
		utils.Info("No main packages found")
	}

	// If mains-only flag is set, stop here
	if mainsOnly {
		utils.Success("Mains-only build completed")
		return nil
	}

	// Phase 2: Build remaining packages
	utils.Info("Phase 2: Building remaining packages...")

	// Get all packages
	allPackages, err := b.discoverPackages(exclude)
	if err != nil {
		return err
	}

	// Filter out main packages that were already built
	mainSet := make(map[string]bool)
	for _, pkg := range mainPackages {
		mainSet[pkg] = true
	}

	var remainingPackages []string
	for _, pkg := range allPackages {
		if !mainSet[pkg] {
			remainingPackages = append(remainingPackages, pkg)
		}
	}

	if len(remainingPackages) == 0 {
		utils.Success("All packages already built with main packages")
		return nil
	}

	// Build remaining packages in batches
	if batchSize <= 0 {
		batchSize = 10
	}

	batches := b.splitIntoBatches(remainingPackages, batchSize)
	utils.Info("Building %d remaining packages in %d batches", len(remainingPackages), len(batches))

	for i, batch := range batches {
		if verbose {
			utils.Info("Building batch %d/%d (%d packages)", i+1, len(batches), len(batch))
		}

		if err := b.buildPackageBatch(batch, parallelism, false); err != nil {
			return fmt.Errorf("batch %d failed: %w", i+1, err)
		}
	}

	utils.Success("Mains-first build completed")
	return nil
}

// buildSmart automatically selects the best strategy based on available resources
func (b Build) buildSmart(exclude string, verbose bool, parallelism string) error {
	utils.Header("Pre-building Packages (Smart Strategy)")

	// Get system memory information
	memInfo, err := utils.GetSystemMemoryInfo()
	if err != nil {
		utils.Warn("Failed to detect system memory: %v", err)
		utils.Info("Falling back to incremental strategy with small batches")
		return b.buildIncremental(5, 500, exclude, verbose, parallelism)
	}

	availableMemoryMB := memInfo.AvailableMemory / (1024 * 1024)
	utils.Info("Available memory: %s", utils.FormatMemory(memInfo.AvailableMemory))

	// Check if we're in workspace mode - if so, use per-module building
	// to avoid issues with modules that exist but aren't in go.work
	if _, isWorkspace := b.getWorkspaceModuleDirs(); isWorkspace {
		utils.Info("Workspace mode detected - building each module separately")
		return b.buildWorkspaceModules(verbose, parallelism, exclude)
	}

	// Count packages to determine best strategy
	packages, err := b.discoverPackages(exclude)
	if err != nil {
		return err
	}
	packageCount := len(packages)
	utils.Info("Found %d packages to build", packageCount)

	// Estimate memory requirements
	estimatedMemory := utils.EstimatePackageBuildMemory(packageCount)
	estimatedMemoryMB := estimatedMemory / (1024 * 1024)
	utils.Info("Estimated memory requirement: %s", utils.FormatMemory(estimatedMemory))

	// Select strategy based on heuristics
	var strategy string
	var batchSize int
	var delayMs int

	switch {
	case availableMemoryMB < 4000 || packageCount > 500:
		// Low memory or many packages: Use small batches
		strategy = "incremental (low memory/high package count)"
		batchSize = 5
		delayMs = 500

	case availableMemoryMB < 8000 || packageCount > 200:
		// Medium resources: Use medium batches
		strategy = "incremental (medium resources)"
		batchSize = 20
		delayMs = 200

	case packageCount > 50 && estimatedMemoryMB > availableMemoryMB*80/100:
		// Memory usage would be close to limit: Use mains-first
		strategy = "mains-first (optimize memory usage)"
		utils.Info("Selected strategy: %s", strategy)
		return b.buildMainsFirst(10, false, exclude, verbose, parallelism)

	default:
		// Sufficient resources: Use full build
		strategy = "full (sufficient resources)"
		utils.Info("Selected strategy: %s", strategy)
		return b.buildFull(parallelism, verbose, exclude)
	}

	utils.Info("Selected strategy: %s", strategy)
	return b.buildIncremental(batchSize, delayMs, exclude, verbose, parallelism)
}

// buildFull performs a traditional full build (backward compatibility)
// When exclude is provided, it uses package discovery to filter excluded packages
func (b Build) buildFull(parallelism string, verbose bool, exclude string) error {
	// If exclude pattern provided, use filtered package list instead of ./...
	if exclude != "" {
		packages, err := b.discoverPackages(exclude)
		if err != nil {
			return fmt.Errorf("failed to discover packages: %w", err)
		}

		if len(packages) == 0 {
			utils.Info("No packages to build after filtering")
			return nil
		}

		utils.Info("Building %d packages (excluded: %s)", len(packages), exclude)

		args := []string{"build"}
		if parallelism != "" {
			args = append(args, "-p", parallelism)
		}
		if verbose {
			args = append(args, "-v")
		}
		args = append(args, packages...)

		utils.Info("Running: go %s", strings.Join(args, " "))

		if err := GetRunner().RunCmd("go", args...); err != nil {
			return fmt.Errorf("full pre-build failed: %w", err)
		}
		return nil
	}

	// Original behavior when no exclusion
	args := []string{"build"}

	if parallelism != "" {
		args = append(args, "-p", parallelism)
	}

	if verbose {
		args = append(args, "-v")
	}

	args = append(args, "./...")

	utils.Info("Running: go %s", strings.Join(args, " "))

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("full pre-build failed: %w", err)
	}

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
		// Expand template variables in ldflags
		expandedLDFlags := expandLDFlagsTemplates(cfg.Build.LDFlags)
		flags = append(flags, "-ldflags", strings.Join(expandedLDFlags, " "))
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

// expandLDFlagsTemplates expands template variables in ldflags
func expandLDFlagsTemplates(ldflags []string) []string {
	expanded := make([]string, len(ldflags))

	// Get template values
	version := getVersion()
	commit := getCommit()
	date := time.Now().Format(time.RFC3339)

	// Expand each ldflag
	for i, flag := range ldflags {
		expanded[i] = flag
		expanded[i] = strings.ReplaceAll(expanded[i], "{{.Version}}", version)
		expanded[i] = strings.ReplaceAll(expanded[i], "{{.Commit}}", commit)
		expanded[i] = strings.ReplaceAll(expanded[i], "{{.Date}}", date)
		expanded[i] = strings.ReplaceAll(expanded[i], "{{.BuildDate}}", date)
		expanded[i] = strings.ReplaceAll(expanded[i], "{{.BuildTime}}", date)
	}

	return expanded
}

// getVersion returns the current version

// getCommit returns the current git commit
func getCommit() string {
	if commit, err := GetRunner().RunCmdOutput("git", "rev-parse", "--short", "HEAD"); err == nil {
		return strings.TrimSpace(commit)
	}
	return statusUnknown
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

// extractMainPackageFromLine extracts the import path from a line containing main package info
func extractMainPackageFromLine(line string) string {
	// Check for main package indicator (handle both formats with/without spaces)
	if !strings.Contains(line, `"Name": "main"`) && !strings.Contains(line, `"Name":"main"`) {
		return ""
	}

	// Extract ImportPath (handle both formats)
	var pkg string
	if idx := strings.Index(line, `"ImportPath": "`); idx != -1 {
		start := idx + len(`"ImportPath": "`)
		end := strings.Index(line[start:], `"`)
		if end != -1 {
			pkg = line[start : start+end]
		}
	} else if idx := strings.Index(line, `"ImportPath":"`); idx != -1 {
		start := idx + len(`"ImportPath":"`)
		end := strings.Index(line[start:], `"`)
		if end != -1 {
			pkg = line[start : start+end]
		}
	}

	// Skip the root package if it's just a magefile
	if pkg != "" && pkg != "github.com/mrz1836/mage-x" {
		return pkg
	}
	return ""
}

// applyMemoryLimit parses memory limit and adjusts strategy if needed
func (b Build) applyMemoryLimit(memoryLimit, strategy string) string {
	// Parse memory limit (e.g., "4G", "4096M")
	var limitBytes uint64
	if strings.HasSuffix(memoryLimit, "G") {
		if gb, err := strconv.ParseUint(strings.TrimSuffix(memoryLimit, "G"), 10, 64); err == nil {
			limitBytes = gb * 1024 * 1024 * 1024
		}
	} else if strings.HasSuffix(memoryLimit, "M") {
		if mb, err := strconv.ParseUint(strings.TrimSuffix(memoryLimit, "M"), 10, 64); err == nil {
			limitBytes = mb * 1024 * 1024
		}
	}

	if limitBytes > 0 {
		// Check if we have enough memory for the selected strategy
		memInfo, err := utils.GetSystemMemoryInfo()
		if err == nil && memInfo.AvailableMemory < limitBytes {
			utils.Warn("Available memory (%s) is less than limit (%s), using incremental strategy",
				utils.FormatMemory(memInfo.AvailableMemory), memoryLimit)
			return "incremental"
		}
	}
	return strategy
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
