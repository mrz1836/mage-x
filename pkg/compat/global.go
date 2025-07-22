package compat

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/env"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/mage"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Global provides global backward compatibility functions
var Global = &GlobalCompat{
	adapter:     NewV1Adapter(),
	initialized: false,
}

// GlobalCompat provides global compatibility layer
type GlobalCompat struct {
	adapter     *V1Adapter
	initialized bool
	config      *mage.Config
	runner      mage.CommandRunner
}

// Init initializes the compatibility layer
func (g *GlobalCompat) Init() error {
	if g.initialized {
		return nil
	}

	// Check for legacy mode
	if g.adapter.legacyMode {
		utils.Info("Running in legacy compatibility mode")

		// Load legacy config
		cfg, err := g.adapter.LoadLegacyConfig()
		if err != nil {
			return err
		}

		if cfg != nil {
			g.config = cfg
		}
	}

	// Initialize runner
	g.runner = mage.GetRunner()

	g.initialized = true
	return nil
}

// RunCmd provides backward compatible command execution
func (g *GlobalCompat) RunCmd(name string, args ...string) error {
	if err := g.Init(); err != nil {
		return err
	}

	// Legacy behavior: print command before running
	if os.Getenv("MAGE_VERBOSE") != "" || os.Getenv("MAGEFILE_VERBOSE") != "" {
		fmt.Printf("exec: %s %s\n", name, strings.Join(args, " "))
	}

	return g.runner.RunCmd(name, args...)
}

// RunCmdV runs a command with verbose output (legacy)
func (g *GlobalCompat) RunCmdV(name string, args ...string) error {
	// Always verbose in legacy
	fmt.Printf("exec: %s %s\n", name, strings.Join(args, " "))
	return g.RunCmd(name, args...)
}

// Deps provides legacy mg.Deps compatibility
func (g *GlobalCompat) Deps(fns ...interface{}) {
	// Convert to mg.Deps call
	mg.Deps(fns...)
}

// SerialDeps provides legacy mg.SerialDeps compatibility
func (g *GlobalCompat) SerialDeps(fns ...interface{}) {
	// Convert to mg.SerialDeps call
	mg.SerialDeps(fns...)
}

// CtxDeps provides legacy mg.CtxDeps compatibility
func (g *GlobalCompat) CtxDeps(fns ...interface{}) {
	// Convert to mg.CtxDeps call with context
	ctx := context.Background()
	mg.CtxDeps(ctx, fns...)
}

// Target provides legacy target registration
func (g *GlobalCompat) Target(name string, fn interface{}) {
	// Register as wrapped legacy target
	wrappedFn := g.wrapTarget(name, fn)

	// Store for later execution
	legacyTargets[name] = wrappedFn
}

// wrapTarget wraps a target function for compatibility
func (g *GlobalCompat) wrapTarget(name string, fn interface{}) func() error {
	return func() error {
		// Initialize if needed
		if err := g.Init(); err != nil {
			return err
		}

		// Show deprecation warning
		if g.adapter.deprecationWarns {
			utils.Warn("Legacy target '%s' called. Consider migrating to new namespace structure.", name)
		}

		// Call the function using reflection
		fnValue := reflect.ValueOf(fn)
		fnType := fnValue.Type()

		// Handle different function signatures
		switch fnType.NumIn() {
		case 0:
			// func() or func() error
			results := fnValue.Call(nil)
			if len(results) > 0 {
				if err, ok := results[0].Interface().(error); ok && err != nil {
					return err
				}
			}
			return nil

		default:
			return fmt.Errorf("unsupported target signature for %s", name)
		}
	}
}

// Env provides legacy environment variable access
func (g *GlobalCompat) Env(key, defaultValue string) string {
	env := env.NewOSEnvironment()
	return env.GetWithDefault(key, defaultValue)
}

// SetEnv provides legacy environment variable setting
func (g *GlobalCompat) SetEnv(key, value string) error {
	env := env.NewOSEnvironment()
	return env.Set(key, value)
}

// FileExists provides legacy file existence check
func (g *GlobalCompat) FileExists(path string) bool {
	fileOps := fileops.New()
	return fileOps.File.Exists(path)
}

// ReadFile provides legacy file reading
func (g *GlobalCompat) ReadFile(path string) ([]byte, error) {
	fileOps := fileops.New()
	return fileOps.File.ReadFile(path)
}

// WriteFile provides legacy file writing
func (g *GlobalCompat) WriteFile(path string, data []byte) error {
	fileOps := fileops.New()
	return fileOps.File.WriteFile(path, data, 0o644)
}

// MkdirAll provides legacy directory creation
func (g *GlobalCompat) MkdirAll(path string) error {
	fileOps := fileops.New()
	return fileOps.File.MkdirAll(path, 0o755)
}

// RemoveAll provides legacy directory removal
func (g *GlobalCompat) RemoveAll(path string) error {
	fileOps := fileops.New()
	return fileOps.File.RemoveAll(path)
}

// GOOS provides legacy GOOS access
func (g *GlobalCompat) GOOS() string {
	return runtime.GOOS
}

// GOARCH provides legacy GOARCH access
func (g *GlobalCompat) GOARCH() string {
	return runtime.GOARCH
}

// Verbose provides legacy verbose check
func (g *GlobalCompat) Verbose() bool {
	return os.Getenv("MAGE_VERBOSE") != "" ||
		os.Getenv("MAGEFILE_VERBOSE") != "" ||
		os.Getenv("VERBOSE") != "" ||
		os.Getenv("V") != ""
}

// Debug provides legacy debug check
func (g *GlobalCompat) Debug() bool {
	return os.Getenv("MAGE_DEBUG") != "" ||
		os.Getenv("DEBUG") != ""
}

// legacyTargets stores registered legacy targets
var legacyTargets = make(map[string]func() error)

// GetLegacyTargets returns all registered legacy targets
func GetLegacyTargets() map[string]func() error {
	return legacyTargets
}

// WrapNamespace wraps a namespace for legacy compatibility
func WrapNamespace(ns interface{}) interface{} {
	// Use reflection to wrap all methods
	nsValue := reflect.ValueOf(ns)
	nsType := nsValue.Type()

	// Create wrapper struct dynamically
	fields := []reflect.StructField{
		{
			Name:      "wrapped",
			Type:      nsType,
			Anonymous: true,
		},
	}

	wrapperType := reflect.StructOf(fields)
	wrapper := reflect.New(wrapperType).Elem()
	wrapper.Field(0).Set(nsValue)

	return wrapper.Interface()
}

// RegisterCompatibilityAliases registers common aliases for backward compatibility
func RegisterCompatibilityAliases() {
	// Register common function aliases
	aliases := map[string]func() error{
		// Build aliases
		"build":   mage.Build{}.Default,
		"compile": mage.Build{}.Default,
		"make":    mage.Build{}.Default,

		// Test aliases
		"test":     mage.Test{}.Unit,
		"unittest": mage.Test{}.Unit,
		"check":    mage.Test{}.All,

		// Lint aliases
		"lint": mage.Lint{}.All,
		"vet":  mage.Lint{}.Vet,
		"fmt":  mage.Format{}.Check,

		// Clean aliases
		"clean":     mage.Clean{}.All,
		"cleanall":  mage.Clean{}.All,
		"distclean": mage.Clean{}.Full,

		// Docker aliases
		"docker":    func() error { return mage.Docker{}.Build("app") },
		"image":     func() error { return mage.Docker{}.Build("app") },
		"container": func() error { return mage.Docker{}.Build("app") },

		// Dependency aliases
		"deps":   mage.Deps{}.Download,
		"dep":    mage.Deps{}.Download,
		"vendor": mage.Deps{}.Vendor,
		"mod":    mage.Deps{}.Download,

		// Release aliases
		"release": mage.Releases{}.Create,
		"publish": mage.Releases{}.Publish,
		"dist":    mage.Releases{}.Create,

		// Install aliases
		"install":   mage.Install{}.Local,
		"uninstall": mage.Install{}.Uninstall,
	}

	for name, fn := range aliases {
		legacyTargets[name] = fn
	}
}

// MigrateCommand provides the migration command
func MigrateCommand() error {
	adapter := NewV1Adapter()

	// Check what needs migration
	needsConfigMigration := false
	needsTargetMigration := false

	legacyConfigs := []string{".mage.yml", "mage.yaml", "mage.yml", "Magefile.yml"}
	for _, cfg := range legacyConfigs {
		if adapter.files.Exists(cfg) {
			needsConfigMigration = true
			break
		}
	}

	legacyTargets := adapter.CheckLegacyTargets()
	if len(legacyTargets) > 0 {
		needsTargetMigration = true
	}

	if !needsConfigMigration && !needsTargetMigration {
		utils.Success("No migration needed! Your project is already using the latest format.")
		return nil
	}

	// Perform migrations
	if needsConfigMigration {
		if err := adapter.MigrateConfig(); err != nil {
			return err
		}
	}

	if needsTargetMigration {
		if err := adapter.MigrateLegacyTargets(); err != nil {
			return err
		}
	}

	return nil
}

// CheckCompatibility checks for compatibility issues
func CheckCompatibility() error {
	utils.Header("Compatibility Check")

	adapter := NewV1Adapter()
	issues := []string{}

	// Check for legacy configs
	legacyConfigs := []string{".mage.yml", "mage.yaml", "mage.yml", "Magefile.yml"}
	for _, cfg := range legacyConfigs {
		if adapter.files.Exists(cfg) {
			issues = append(issues, fmt.Sprintf("Legacy config file: %s", cfg))
		}
	}

	// Check for legacy targets
	legacyTargets := adapter.CheckLegacyTargets()
	for _, target := range legacyTargets {
		issues = append(issues, fmt.Sprintf("Legacy target file: %s", target))
	}

	// Check for old environment variables
	oldEnvVars := []string{"MAGEFILE_VERBOSE", "MAGEFILE_DEBUG", "MAGEFILE_GOCMD"}
	envHelper := env.NewOSEnvironment()
	for _, v := range oldEnvVars {
		if envHelper.GetWithDefault(v, "") != "" {
			issues = append(issues, fmt.Sprintf("Legacy environment variable: %s", v))
		}
	}

	if len(issues) == 0 {
		utils.Success("No compatibility issues found!")
		return nil
	}

	utils.Warn("Found %d compatibility issue(s):", len(issues))
	for _, issue := range issues {
		fmt.Printf("  - %s\n", issue)
	}

	fmt.Println("\nRun 'mage migrate' to automatically fix these issues.")

	return nil
}
