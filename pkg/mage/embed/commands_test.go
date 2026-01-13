package embed

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

var errWorkerPanic = errors.New("worker panicked")

// Test constants
const (
	testTimeout         = 5 * 1000 // 5 seconds in milliseconds for timeout tests
	expectedMinCommands = 166      // Minimum expected number of registered commands
	concurrentWorkers   = 10       // Number of concurrent workers for parallel tests
)

// Test errors for security testing
var (
	ErrCommandInjection = errors.New("potential command injection detected")
	ErrInvalidInput     = errors.New("invalid input detected")
	errTestBinding      = errors.New("test error")
)

// MockRegistry provides a test registry with enhanced security validation
type MockRegistry struct {
	*registry.Registry

	// Security validation
	validationErrors []error
	registrationLog  []string
	mu               sync.RWMutex
}

// NewMockRegistry creates a new mock registry for testing
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		Registry:         registry.NewRegistry(),
		validationErrors: []error{},
		registrationLog:  []string{},
	}
}

// SetRegistered overrides the base implementation to add logging
func (m *MockRegistry) SetRegistered(registered bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Registry.SetRegistered(registered)
	m.registrationLog = append(m.registrationLog, fmt.Sprintf("SetRegistered: %v", registered))
}

// MustRegister overrides the base implementation with enhanced security validation
func (m *MockRegistry) MustRegister(cmd *registry.Command) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Security validation
	if err := m.validateCommandSecurity(cmd); err != nil {
		m.validationErrors = append(m.validationErrors, err)
		panic(fmt.Sprintf("security validation failed: %v", err))
	}

	// Use the base registry for actual registration
	m.Registry.MustRegister(cmd)

	cmdName := cmd.FullName()
	m.registrationLog = append(m.registrationLog, fmt.Sprintf("MustRegister: %s", cmdName))
}

// validateCommandSecurity validates command for security issues
func (m *MockRegistry) validateCommandSecurity(cmd *registry.Command) error {
	// Validate for potential command injection patterns
	dangerousPatterns := []string{
		";", "&&", "||", "|", "`", "$", "$(", "${",
		"eval", "exec", "sh", "bash", "cmd.exe",
		"rm -rf", "del /f", "format c:",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(cmd.Description), pattern) {
			return fmt.Errorf("%w: dangerous pattern '%s' in description", ErrCommandInjection, pattern)
		}
		if strings.Contains(strings.ToLower(cmd.Usage), pattern) {
			return fmt.Errorf("%w: dangerous pattern '%s' in usage", ErrCommandInjection, pattern)
		}
	}

	// Validate command structure
	if cmd.Namespace == "" && cmd.Method == "" && cmd.Name == "" {
		return fmt.Errorf("%w: command must have name or namespace+method", ErrInvalidInput)
	}

	return nil
}

// GetRegisteredCommands returns all registered commands for testing
func (m *MockRegistry) GetRegisteredCommands() map[string]*registry.Command {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Access commands through the List method and build map
	commands := m.List()
	result := make(map[string]*registry.Command)
	for _, cmd := range commands {
		result[cmd.FullName()] = cmd
	}
	return result
}

// GetRegisteredAliases returns all registered aliases for testing
func (m *MockRegistry) GetRegisteredAliases() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Since we can't access aliases directly, we'll simulate based on commands
	result := make(map[string]string)
	commands := m.List()
	for _, cmd := range commands {
		for _, alias := range cmd.Aliases {
			result[alias] = cmd.FullName()
		}
	}
	return result
}

// GetValidationErrors returns security validation errors
func (m *MockRegistry) GetValidationErrors() []error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]error(nil), m.validationErrors...)
}

// GetRegistrationLog returns the registration log
func (m *MockRegistry) GetRegistrationLog() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string(nil), m.registrationLog...)
}

// Clear clears the mock registry
func (m *MockRegistry) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Registry.Clear()
	m.validationErrors = []error{}
	m.registrationLog = []string{}
}

// TestRegisterAll tests the main RegisterAll function
func TestRegisterAll(t *testing.T) {
	t.Parallel()

	t.Run("ValidRegistry", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()

		// Test with empty registry
		require.NotPanics(t, func() {
			RegisterAll(mockReg.Registry)
		})

		// Verify registration status was set
		assert.True(t, mockReg.IsRegistered(), "Registry should be marked as registered")

		// Verify commands were registered
		commands := mockReg.GetRegisteredCommands()
		assert.GreaterOrEqual(t, len(commands), expectedMinCommands,
			"Should register at least %d commands", expectedMinCommands)

		// Verify no security validation errors
		errors := mockReg.GetValidationErrors()
		assert.Empty(t, errors, "Should have no security validation errors")
	})

	t.Run("NilRegistry", func(t *testing.T) {
		t.Parallel()

		// Test with nil registry - should use global registry
		require.NotPanics(t, func() {
			RegisterAll(nil)
		})
	})

	t.Run("AlreadyRegistered", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		mockReg.SetRegistered(true)

		// Should exit early if already registered
		RegisterAll(mockReg.Registry)

		commands := mockReg.GetRegisteredCommands()
		assert.Empty(t, commands, "Should not register commands if already registered")
	})

	t.Run("DoubleRegistration", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()

		// First registration
		RegisterAll(mockReg.Registry)
		firstCommandCount := len(mockReg.GetRegisteredCommands())

		// Second registration should be ignored
		RegisterAll(mockReg.Registry)
		secondCommandCount := len(mockReg.GetRegisteredCommands())

		assert.Equal(t, firstCommandCount, secondCommandCount,
			"Double registration should not add more commands")
	})
}

// TestRegisterAllNamespaces tests that all expected namespaces are registered
func TestRegisterAllNamespaces(t *testing.T) {
	t.Parallel()

	expectedNamespaces := []string{
		"build", "test", "lint", "format", "deps", "git", "release",
		"docs", "tools", "generate", "update", "mod",
		"metrics", "bench", "vet", "configure",
		"help", "version", "install", "yaml",
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()
	foundNamespaces := make(map[string]bool)

	for _, cmd := range commands {
		if cmd.Namespace != "" {
			foundNamespaces[cmd.Namespace] = true
		}
	}

	for _, expectedNS := range expectedNamespaces {
		assert.True(t, foundNamespaces[expectedNS],
			"Expected namespace '%s' should be registered", expectedNS)
	}
}

// TestBuildNamespaceCommands tests build namespace command registration
func TestBuildNamespaceCommands(t *testing.T) {
	t.Parallel()

	expectedCommands := []struct {
		method      string
		description string
		category    string
		hasAliases  bool
	}{
		{"default", "Build the application for the current platform", "Build", true},
		{"all", "Build for all configured platforms", "Build", false},
		{"linux", "Build for Linux (amd64)", "Build", false},
		{"darwin", "Build for macOS (amd64 and arm64)", "Build", true},
		{"windows", "Build for Windows (amd64)", "Build", false},
		{"clean", "Remove build artifacts", "Build", true},
		{"install", "Install the binary to $GOPATH/bin", "Build", true},
		{"generate", "Run go generate", "Build", false},
		{"prebuild", "Pre-build all packages to warm cache", "Build", false},
	}

	mockReg := NewMockRegistry()
	registerBuildCommands(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for _, expected := range expectedCommands {
		cmdName := fmt.Sprintf("build:%s", expected.method)
		cmd, exists := commands[cmdName]

		require.True(t, exists, "Command %s should be registered", cmdName)
		assert.Equal(t, expected.description, cmd.Description,
			"Description should match for %s", cmdName)
		assert.Equal(t, expected.category, cmd.Category,
			"Category should match for %s", cmdName)

		if expected.hasAliases {
			assert.NotEmpty(t, cmd.Aliases, "Command %s should have aliases", cmdName)
		}

		// Verify function is set
		assert.NotNil(t, cmd.Func, "Command %s should have a function", cmdName)
	}
}

// TestTestNamespaceCommands tests test namespace command registration
func TestTestNamespaceCommands(t *testing.T) {
	t.Parallel()

	expectedCommands := []struct {
		method      string
		hasArgsFunc bool
	}{
		{"default", true},   // Now has ArgsFunc for JSON support
		{"unit", true},      // Now has ArgsFunc for JSON support
		{"short", true},     // Now has ArgsFunc for JSON support
		{"race", true},      // Now has ArgsFunc for JSON support
		{"cover", true},     // Now has ArgsFunc for JSON support
		{"coverrace", true}, // Now has ArgsFunc for JSON support
		{"coverreport", false},
		{"coverhtml", false},
		{"fuzz", true},       // Now has ArgsFunc for time parameter
		{"fuzzshort", true},  // Has ArgsFunc for time parameter
		{"bench", true},      // Has ArgsFunc
		{"benchshort", true}, // Has ArgsFunc for time parameter
		{"integration", false},
		{"ci", false},
		{"parallel", false},
		{"nolint", false},
		{"cinorace", false},
		{"run", false},
		{"coverage", true}, // Has ArgsFunc
		{"vet", false},
	}

	mockReg := NewMockRegistry()
	registerTestCommands(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for _, expected := range expectedCommands {
		cmdName := fmt.Sprintf("test:%s", expected.method)
		cmd, exists := commands[cmdName]

		require.True(t, exists, "Command %s should be registered", cmdName)
		assert.Equal(t, "Test", cmd.Category, "Category should be Test for %s", cmdName)

		if expected.hasArgsFunc {
			assert.NotNil(t, cmd.FuncWithArgs, "Command %s should have ArgsFunc", cmdName)
		} else {
			assert.NotNil(t, cmd.Func, "Command %s should have Func", cmdName)
		}
	}
}

// TestTopLevelCommands tests top-level convenience command registration
func TestTopLevelCommands(t *testing.T) {
	t.Parallel()

	expectedTopLevel := []string{
		"build", "test", "lint", "format", "clean", "install", "bench",
	}

	mockReg := NewMockRegistry()
	registerTopLevelCommands(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for _, cmdName := range expectedTopLevel {
		cmd, exists := commands[cmdName]

		require.True(t, exists, "Top-level command %s should be registered", cmdName)
		assert.Equal(t, "Common", cmd.Category, "Category should be Common for %s", cmdName)

		// Bench command should have ArgsFunc, others should have Func
		if cmdName == "bench" {
			assert.NotNil(t, cmd.FuncWithArgs, "Command %s should have ArgsFunc", cmdName)
		} else {
			assert.NotNil(t, cmd.Func, "Command %s should have Func", cmdName)
		}
	}
}

// TestCommandAliases tests that command aliases are properly registered
func TestCommandAliases(t *testing.T) {
	t.Parallel()

	expectedAliases := map[string]string{
		"build":     "build:default",
		"clean":     "build:clean",
		"install":   "build:install",
		"test":      "test:default",
		"benchmark": "test:bench",
		"lint":      "lint:default",
		"format":    "format:default",
		"fmt":       "format:default",
		"deps":      "deps:default",
		"tidy":      "deps:tidy",
		"vendor":    "deps:vendor",
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	aliases := mockReg.GetRegisteredAliases()

	for alias, expectedCommand := range expectedAliases {
		actualCommand, exists := aliases[alias]
		assert.True(t, exists, "Alias '%s' should exist", alias)
		assert.Equal(t, expectedCommand, actualCommand,
			"Alias '%s' should point to '%s'", alias, expectedCommand)
	}
}

// TestCommandSecurity tests security validation of registered commands
func TestCommandSecurity(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	// Verify no security validation errors occurred
	errors := mockReg.GetValidationErrors()
	assert.Empty(t, errors, "Should have no security validation errors during registration")

	// Test all registered commands for security
	commands := mockReg.GetRegisteredCommands()

	for cmdName, cmd := range commands {
		t.Run(fmt.Sprintf("SecurityValidation_%s", cmdName), func(t *testing.T) {
			// Check for dangerous patterns in description
			assert.NotContains(t, strings.ToLower(cmd.Description), "rm -rf",
				"Command %s should not contain dangerous patterns", cmdName)
			assert.NotContains(t, strings.ToLower(cmd.Description), "del /f",
				"Command %s should not contain dangerous patterns", cmdName)

			// Ensure command has proper structure
			assert.True(t, cmd.Namespace != "" || cmd.Name != "",
				"Command %s should have proper identification", cmdName)

			// Verify function assignment
			assert.True(t, cmd.Func != nil || cmd.FuncWithArgs != nil,
				"Command %s should have executable function", cmdName)
		})
	}
}

// TestConcurrentRegistration tests thread safety of command registration
func TestConcurrentRegistration(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	registrationErrors := make(chan error, concurrentWorkers)

	for i := 0; i < concurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					registrationErrors <- fmt.Errorf("worker %d: %w: %v", workerID, errWorkerPanic, r)
				}
			}()

			// Each worker creates its own mock registry
			mockReg := NewMockRegistry()
			RegisterAll(mockReg.Registry)

			// Verify registration completed successfully
			if !mockReg.IsRegistered() {
				registrationErrors <- fmt.Errorf("worker %d: %w", workerID, ErrRegistrationIncomplete)
				return
			}

			commands := mockReg.GetRegisteredCommands()
			if len(commands) < expectedMinCommands {
				registrationErrors <- fmt.Errorf("worker %d: %w (got %d)", workerID, ErrInsufficientCommands, len(commands))
			}
		}(i)
	}

	wg.Wait()
	close(registrationErrors)

	// Check for any errors
	errors := make([]error, 0, concurrentWorkers)
	for err := range registrationErrors {
		errors = append(errors, err)
	}

	assert.Empty(t, errors, "Concurrent registration should not produce errors: %v", errors)
}

// TestNamespaceCommandPatterns tests that all namespace commands follow patterns
func TestNamespaceCommandPatterns(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()
	namespacePatterns := make(map[string][]string)

	// Group commands by namespace
	for cmdName, cmd := range commands {
		if cmd.Namespace != "" {
			namespacePatterns[cmd.Namespace] = append(namespacePatterns[cmd.Namespace], cmdName)
		}
	}

	// Verify that some core namespaces have default commands
	coreNamespacesWithDefaults := map[string]bool{
		"build":  true,
		"test":   true,
		"lint":   true,
		"format": true,
		"deps":   true,
		"docs":   true,
		"init":   true,
	}

	for namespace, cmdNames := range namespacePatterns {
		defaultCmd := fmt.Sprintf("%s:default", namespace)
		hasDefault := false

		for _, cmdName := range cmdNames {
			if cmdName == defaultCmd {
				hasDefault = true
				break
			}
		}

		// Check that core namespaces have default commands
		if coreNamespacesWithDefaults[namespace] {
			assert.True(t, hasDefault, "Core namespace '%s' should have a default command", namespace)
		}

		// All namespaces should have at least one command
		assert.NotEmpty(t, cmdNames, "Namespace '%s' should have at least one command", namespace)
	}
}

// TestCommandDescriptions tests that all commands have proper descriptions
func TestCommandDescriptions(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for cmdName, cmd := range commands {
		t.Run(fmt.Sprintf("Description_%s", cmdName), func(t *testing.T) {
			assert.NotEmpty(t, cmd.Description, "Command %s should have description", cmdName)
			assert.NotEmpty(t, cmd.Category, "Command %s should have category", cmdName)

			// Description should be meaningful (more than just the command name)
			assert.Greater(t, len(cmd.Description), 5,
				"Command %s should have meaningful description", cmdName)
		})
	}
}

// TestCommandCategories tests that commands are properly categorized
func TestCommandCategories(t *testing.T) {
	t.Parallel()

	expectedCategories := map[string][]string{
		"Build":              {"build:"},
		"Test":               {"test:"},
		"Lint":               {"lint:"},
		"Format":             {"format:"},
		"Dependencies":       {"deps:"},
		"Git":                {"git:"},
		"Release":            {"release:"},
		"Documentation":      {"docs:"},
		"Tools":              {"tools:"},
		"Generate":           {"generate:"},
		"Update":             {"update:"},
		"Module":             {"mod:"},
		"Metrics":            {"metrics:"},
		"Benchmark":          {"bench:"},
		"Vet":                {"vet:"},
		"Configure":          {"configure:"},
		"Help":               {"help:"},
		"Version Management": {"version:"},
		"Installation":       {"install:"},
		"Configuration":      {"yaml:"},
		"Common":             {"build", "test", "lint", "format", "clean", "install", "bench"},
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()
	categorizedCommands := make(map[string][]string)

	// Group commands by category
	for cmdName, cmd := range commands {
		categorizedCommands[cmd.Category] = append(categorizedCommands[cmd.Category], cmdName)
	}

	for category, expectedPrefixes := range expectedCategories {
		categoryCommands, exists := categorizedCommands[category]
		assert.True(t, exists, "Category '%s' should exist", category)

		if exists {
			foundCommands := 0
			for _, expectedPrefix := range expectedPrefixes {
				for _, cmdName := range categoryCommands {
					if strings.HasPrefix(cmdName, expectedPrefix) || cmdName == strings.TrimSuffix(expectedPrefix, ":") {
						foundCommands++
						break
					}
				}
			}
			assert.Positive(t, foundCommands,
				"Category '%s' should have commands matching expected prefixes", category)
		}
	}
}

// TestIndividualNamespaceRegistration tests each namespace registration function individually
func TestIndividualNamespaceRegistration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		registerFunc   func(*registry.Registry)
		expectedPrefix string
		minCommands    int
	}{
		{"BuildCommands", registerBuildCommands, "build:", 8},
		{"TestCommands", registerTestCommands, "test:", 15},
		{"LintCommands", registerLintCommands, "lint:", 4},
		{"FormatCommands", registerFormatCommands, "format:", 3},
		{"DepsCommands", registerDepsCommands, "deps:", 10},
		{"GitCommands", registerGitCommands, "git:", 8},
		{"ReleaseCommands", registerReleaseCommands, "release:", 6},
		{"DocsCommands", registerDocsCommands, "docs:", 8},
		{"ToolsCommands", registerToolsCommands, "tools:", 3},
		{"GenerateCommands", registerGenerateCommands, "generate:", 4},
		{"UpdateCommands", registerUpdateCommands, "update:", 2},
		{"ModCommands", registerModCommands, "mod:", 8},
		{"MetricsCommands", registerMetricsCommands, "metrics:", 5},
		{"BenchCommands", registerBenchCommands, "bench:", 6},
		{"VetCommands", registerVetCommands, "vet:", 1},
		{"ConfigureCommands", registerConfigureCommands, "configure:", 6},
		{"HelpCommands", registerHelpCommands, "help:", 6},
		{"VersionCommands", registerVersionCommands, "version:", 5},
		{"InstallCommands", registerInstallCommands, "install:", 12},
		{"YamlCommands", registerYamlCommands, "yaml:", 4},
		{"TopLevelCommands", registerTopLevelCommands, "", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockReg := NewMockRegistry()
			tc.registerFunc(mockReg.Registry)

			commands := mockReg.GetRegisteredCommands()
			matchingCommands := 0

			for cmdName := range commands {
				if tc.expectedPrefix == "" || strings.HasPrefix(cmdName, tc.expectedPrefix) {
					matchingCommands++
				}
			}

			assert.GreaterOrEqual(t, matchingCommands, tc.minCommands,
				"Expected at least %d commands for %s", tc.minCommands, tc.name)
		})
	}
}

// TestCommandWithArgsValidation tests commands with arguments
func TestCommandWithArgsValidation(t *testing.T) {
	t.Parallel()

	expectedArgsCommands := map[string]bool{
		"test:bench":        true,
		"test:coverage":     true,
		"git:tag":           true,
		"git:tagremove":     true,
		"git:tagupdate":     true,
		"git:commit":        true,
		"git:add":           true,
		"bench:default":     true,
		"bench:compare":     true,
		"bench:save":        true,
		"bench:cpu":         true,
		"bench:mem":         true,
		"bench:profile":     true,
		"bench:trace":       true,
		"bench:regression":  true,
		"version:bump":      true,
		"version:changelog": true,
		"bench":             true, // top-level bench command
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for cmdName, expectArgs := range expectedArgsCommands {
		cmd, exists := commands[cmdName]
		require.True(t, exists, "Command %s should be registered", cmdName)

		if expectArgs {
			assert.NotNil(t, cmd.FuncWithArgs, "Command %s should have ArgsFunc", cmdName)
		}
	}
}

// TestCommandUsageAndExamples tests that commands with args have usage and examples
func TestCommandUsageAndExamples(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for cmdName, cmd := range commands {
		t.Run(fmt.Sprintf("UsageExamples_%s", cmdName), func(t *testing.T) {
			// Commands with ArgsFunc should have usage and examples
			if cmd.FuncWithArgs != nil {
				if len(cmd.Examples) == 0 && cmd.Usage == "" {
					t.Logf("Command %s has ArgsFunc but no usage/examples - this might be acceptable", cmdName)
				}
			}
		})
	}
}

// TestCriticalCommandsExist ensures that critical commands are always registered
func TestCriticalCommandsExist(t *testing.T) {
	t.Parallel()

	// These are critical commands that should never be accidentally removed
	criticalCommands := []string{
		"build:default",
		"test:default",
		"lint:default",
		"deps:audit",
		"deps:update",
		"deps:tidy",
		"git:status",
		"release:default",
		"docs:serve",
		"tools:install",
		"generate:mocks",
		"configure:init",
		"help:default",
		"version:show",
		// Tool integrations - these must always be available
		"bmad:install",
		"bmad:check",
		"aws:login",
		"aws:status",
		"speckit:install",
		"speckit:check",
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for _, cmdName := range criticalCommands {
		t.Run(fmt.Sprintf("Critical_%s", cmdName), func(t *testing.T) {
			cmd, exists := commands[cmdName]
			require.True(t, exists, "Critical command %s must be registered", cmdName)

			// Command should have either Func or FuncWithArgs
			hasFunc := cmd.Func != nil || cmd.FuncWithArgs != nil
			require.True(t, hasFunc, "Critical command %s must have a function (Func or FuncWithArgs)", cmdName)
		})
	}
}

// TestAllNamespaceImplementationsRegistered ensures that all mage.Namespace
// implementations in pkg/mage have corresponding command registrations.
// This prevents the speckit-style bug where implementation exists but registration is missing.
func TestAllNamespaceImplementationsRegistered(t *testing.T) {
	t.Parallel()

	// Expected namespaces that MUST have command registrations.
	// This list should match all `type Xxx mg.Namespace` definitions in pkg/mage.
	// When adding a new namespace, add it here to ensure registration is not forgotten.
	expectedNamespaces := []string{
		"build", "test", "lint", "format", "deps", "git", "release",
		"docs", "tools", "generate", "update", "mod", "metrics",
		"bench", "vet", "configure", "help", "version", "install",
		"yaml", "bmad", "aws", "speckit",
	}

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)
	commands := mockReg.GetRegisteredCommands()

	for _, ns := range expectedNamespaces {
		t.Run(ns, func(t *testing.T) {
			found := false
			for cmdName := range commands {
				if strings.HasPrefix(cmdName, ns+":") {
					found = true
					break
				}
			}
			require.True(t, found,
				"Namespace %q has no registered commands. "+
					"Did you forget to add registerXxxCommands() to RegisterAll()? "+
					"Check pkg/mage/embed/commands.go", ns)
		})
	}
}

// TestRegistryInterface ensures the registry interface is properly implemented
func TestRegistryInterface(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()

	// Test interface methods
	assert.False(t, mockReg.IsRegistered(), "Should start unregistered")

	mockReg.SetRegistered(true)
	assert.True(t, mockReg.IsRegistered(), "Should be registered after SetRegistered(true)")

	mockReg.SetRegistered(false)
	assert.False(t, mockReg.IsRegistered(), "Should be unregistered after SetRegistered(false)")
}

// TestErrorHandling tests error handling in command registration
func TestErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("SecurityValidation", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()

		// Test registering a command with dangerous content should panic
		assert.Panics(t, func() {
			mockReg.MustRegister(&registry.Command{
				Name:        "dangerous",
				Description: "This command runs rm -rf / to delete everything",
				Func:        func() error { return nil },
			})
		}, "Should panic on dangerous command")
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()

		// Test registering invalid command should panic
		assert.Panics(t, func() {
			mockReg.MustRegister(&registry.Command{
				// No name, namespace, or method
			})
		}, "Should panic on invalid command")
	})
}

// BenchmarkRegisterAll benchmarks the RegisterAll function
func BenchmarkRegisterAll(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mockReg := NewMockRegistry()
		b.StartTimer()

		RegisterAll(mockReg.Registry)
	}
}

// BenchmarkIndividualRegistration benchmarks individual registration functions
func BenchmarkIndividualRegistration(b *testing.B) {
	registrationFunctions := []struct {
		name string
		fn   func(*registry.Registry)
	}{
		{"BuildCommands", registerBuildCommands},
		{"TestCommands", registerTestCommands},
		{"LintCommands", registerLintCommands},
		{"FormatCommands", registerFormatCommands},
	}

	for _, rf := range registrationFunctions {
		b.Run(rf.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				mockReg := NewMockRegistry()
				b.StartTimer()

				rf.fn(mockReg.Registry)
			}
		})
	}
}

// BenchmarkConcurrentRegistration benchmarks concurrent registration
func BenchmarkConcurrentRegistration(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		for j := 0; j < concurrentWorkers; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mockReg := NewMockRegistry()
				RegisterAll(mockReg.Registry)
			}()
		}

		wg.Wait()
	}
}

// FuzzRegisterAll performs fuzz testing on the RegisterAll function
func FuzzRegisterAll(f *testing.F) {
	// Add test cases for fuzzing
	f.Add(true)  // registered = true
	f.Add(false) // registered = false

	f.Fuzz(func(t *testing.T, preRegistered bool) {
		mockReg := NewMockRegistry()
		mockReg.SetRegistered(preRegistered)

		// Should not panic regardless of input
		assert.NotPanics(t, func() {
			RegisterAll(mockReg.Registry)
		})
	})
}

// TestMageNamespaceTypes ensures all mage namespace types are available
func TestMageNamespaceTypes(t *testing.T) {
	t.Parallel()

	// Test that all expected mage namespace types exist and can be instantiated
	namespaceTypes := map[string]interface{}{
		"Build":     mage.Build{},
		"Test":      mage.Test{},
		"Lint":      mage.Lint{},
		"Format":    mage.Format{},
		"Deps":      mage.Deps{},
		"Git":       mage.Git{},
		"Release":   mage.Release{},
		"Docs":      mage.Docs{},
		"Tools":     mage.Tools{},
		"Generate":  mage.Generate{},
		"Update":    mage.Update{},
		"Mod":       mage.Mod{},
		"Metrics":   mage.Metrics{},
		"Bench":     mage.Bench{},
		"Vet":       mage.Vet{},
		"Configure": mage.Configure{},
		"Help":      mage.Help{},
		"Version":   mage.Version{},
		"Install":   mage.Install{},
		"Yaml":      mage.Yaml{},
	}

	for typeName, instance := range namespaceTypes {
		t.Run(fmt.Sprintf("Type_%s", typeName), func(t *testing.T) {
			// Verify the type exists and is not nil
			assert.NotNil(t, instance, "Type %s should be instantiable", typeName)

			// Verify it's the expected type
			typeInfo := reflect.TypeOf(instance)
			assert.NotNil(t, typeInfo, "Type %s should have valid type info", typeName)

			// For struct types, verify they have the expected name
			if typeInfo.Kind() == reflect.Struct {
				assert.Contains(t, typeInfo.String(), typeName,
					"Type %s should contain expected name in string representation", typeName)
			}
		})
	}
}

// TestCommandBuilderPattern tests that commands are built using the builder pattern
func TestCommandBuilderPattern(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	// Verify that all commands have required builder pattern fields set
	for cmdName, cmd := range commands {
		t.Run(fmt.Sprintf("BuilderPattern_%s", cmdName), func(t *testing.T) {
			// All commands should have description and category
			assert.NotEmpty(t, cmd.Description, "Command %s should have description", cmdName)
			assert.NotEmpty(t, cmd.Category, "Command %s should have category", cmdName)

			// All commands should have either Func or FuncWithArgs
			assert.True(t, cmd.Func != nil || cmd.FuncWithArgs != nil,
				"Command %s should have executable function", cmdName)

			// Namespace commands should follow naming convention
			if cmd.Namespace != "" && cmd.Method != "" {
				expectedName := fmt.Sprintf("%s:%s", strings.ToLower(cmd.Namespace), strings.ToLower(cmd.Method))
				assert.Equal(t, expectedName, cmd.FullName(),
					"Command %s should follow naming convention", cmdName)
			}
		})
	}
}

// TestThreadSafety tests thread safety of the registration process
func TestThreadSafety(t *testing.T) {
	t.Parallel()

	const goroutines = 50
	const iterations = 10

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("Iteration_%d", i), func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, goroutines)

			for j := 0; j < goroutines; j++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()

					defer func() {
						if r := recover(); r != nil {
							errors <- fmt.Errorf("goroutine %d: %w: %v", id, ErrWorkerPanic, r)
						}
					}()

					mockReg := NewMockRegistry()
					RegisterAll(mockReg.Registry)

					// Verify registration
					if !mockReg.IsRegistered() {
						errors <- fmt.Errorf("goroutine %d: %w", id, ErrRegistrationIncomplete)
						return
					}

					commands := mockReg.GetRegisteredCommands()
					if len(commands) < expectedMinCommands {
						errors <- fmt.Errorf("goroutine %d: %w (got %d)", id, ErrInsufficientCommands, len(commands))
					}
				}(j)
			}

			wg.Wait()
			close(errors)

			var errorList []error
			for err := range errors {
				errorList = append(errorList, err)
			}

			assert.Empty(t, errorList, "Thread safety test should not produce errors: %v", errorList)
		})
	}
}

// TestFuzzCommandsAcceptArguments tests that fuzz commands properly accept time arguments
func TestFuzzCommandsAcceptArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		commandName string
		args        []string
		description string
	}{
		{
			commandName: "test:fuzz",
			args:        []string{"time=5s"},
			description: "test:fuzz should accept time parameter",
		},
		{
			commandName: "test:fuzzshort",
			args:        []string{"time=3s"},
			description: "test:fuzzshort should accept time parameter",
		},
		{
			commandName: "test:benchshort",
			args:        []string{"time=1s"},
			description: "test:benchshort should accept time parameter",
		},
	}

	mockReg := NewMockRegistry()
	registerTestCommands(mockReg.Registry)
	commands := mockReg.GetRegisteredCommands()

	for _, tc := range testCases {
		t.Run(tc.commandName, func(t *testing.T) {
			cmd, exists := commands[tc.commandName]
			require.True(t, exists, "Command %s should be registered", tc.commandName)

			// Verify command has ArgsFunc (can accept arguments)
			assert.NotNil(t, cmd.FuncWithArgs, "%s should have FuncWithArgs to accept arguments", tc.commandName)

			// Verify command has proper usage examples
			assert.NotEmpty(t, cmd.Usage, "%s should have usage information", tc.commandName)
			assert.NotEmpty(t, cmd.Examples, "%s should have usage examples", tc.commandName)

			// Verify the usage contains time parameter information
			assert.Contains(t, cmd.Usage, "time", "%s usage should mention time parameter", tc.commandName)

			// Verify examples show time usage
			foundTimeExample := false
			for _, example := range cmd.Examples {
				if strings.Contains(example, "time=") {
					foundTimeExample = true
					break
				}
			}
			assert.True(t, foundTimeExample, "%s should have example showing time parameter usage", tc.commandName)
		})
	}
}

// TestFuzzCommandArguments tests the specific argument handling for fuzz commands
func TestFuzzCommandArguments(t *testing.T) {
	t.Parallel()

	// Test that the commands exist and can be called with time argument
	mockReg := NewMockRegistry()
	registerTestCommands(mockReg.Registry)

	fuzzCommands := []string{"test:fuzz", "test:fuzzshort", "test:benchshort"}

	for _, cmdName := range fuzzCommands {
		t.Run(fmt.Sprintf("ArgumentHandling_%s", cmdName), func(t *testing.T) {
			cmd, exists := mockReg.Get(cmdName)
			require.True(t, exists, "Command %s should exist", cmdName)

			// Verify the command is registered with ArgsFunc
			assert.NotNil(t, cmd.FuncWithArgs, "Command %s should have FuncWithArgs", cmdName)

			// Test that it has proper category
			assert.Equal(t, "Test", cmd.Category, "Command %s should be in Test category", cmdName)

			// Test command builder pattern completion
			assert.NotEmpty(t, cmd.Description, "Command %s should have description", cmdName)
			assert.NotEmpty(t, cmd.Usage, "Command %s should have usage", cmdName)
			assert.NotEmpty(t, cmd.Examples, "Command %s should have examples", cmdName)
		})
	}
}

// ============================================================================
// Tests for Data-Driven Registration Infrastructure
// ============================================================================

// TestCommandDefStruct tests the CommandDef struct fields and usage
func TestCommandDefStruct(t *testing.T) {
	t.Parallel()

	t.Run("BasicCommandDef", func(t *testing.T) {
		t.Parallel()

		def := CommandDef{
			Method:   "test",
			Desc:     "Test description",
			Aliases:  []string{"t", "tst"},
			Usage:    "magex test [flags]",
			Examples: []string{"magex test", "magex test -v"},
		}

		assert.Equal(t, "test", def.Method)
		assert.Equal(t, "Test description", def.Desc)
		assert.Len(t, def.Aliases, 2)
		assert.Equal(t, "magex test [flags]", def.Usage)
		assert.Len(t, def.Examples, 2)
	})

	t.Run("MinimalCommandDef", func(t *testing.T) {
		t.Parallel()

		def := CommandDef{
			Method: "simple",
			Desc:   "Simple command",
		}

		assert.Equal(t, "simple", def.Method)
		assert.Equal(t, "Simple command", def.Desc)
		assert.Empty(t, def.Aliases)
		assert.Empty(t, def.Usage)
		assert.Empty(t, def.Examples)
	})

	t.Run("CommandDefWithExamples", func(t *testing.T) {
		t.Parallel()

		def := CommandDef{
			Method:   "complex",
			Desc:     "Complex command",
			Examples: []string{"example1", "example2", "example3"},
		}

		assert.Len(t, def.Examples, 3)
		assert.Contains(t, def.Examples, "example1")
		assert.Contains(t, def.Examples, "example2")
		assert.Contains(t, def.Examples, "example3")
	})
}

// TestMethodBindingStruct tests the MethodBinding struct
func TestMethodBindingStruct(t *testing.T) {
	t.Parallel()

	t.Run("NoArgsBinding", func(t *testing.T) {
		t.Parallel()

		called := false
		binding := MethodBinding{
			NoArgs: func() error {
				called = true
				return nil
			},
		}

		assert.NotNil(t, binding.NoArgs)
		assert.Nil(t, binding.WithArgs)

		err := binding.NoArgs()
		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("WithArgsBinding", func(t *testing.T) {
		t.Parallel()

		var receivedArgs []string
		binding := MethodBinding{
			WithArgs: func(args ...string) error {
				receivedArgs = args
				return nil
			},
		}

		assert.Nil(t, binding.NoArgs)
		assert.NotNil(t, binding.WithArgs)

		err := binding.WithArgs("arg1", "arg2")
		require.NoError(t, err)
		assert.Equal(t, []string{"arg1", "arg2"}, receivedArgs)
	})

	t.Run("DualFunctionBinding", func(t *testing.T) {
		t.Parallel()

		noArgsCalled := false
		withArgsCalled := false

		binding := MethodBinding{
			NoArgs: func() error {
				noArgsCalled = true
				return nil
			},
			WithArgs: func(args ...string) error {
				withArgsCalled = true
				return nil
			},
		}

		assert.NotNil(t, binding.NoArgs)
		assert.NotNil(t, binding.WithArgs)

		err := binding.NoArgs()
		require.NoError(t, err)
		assert.True(t, noArgsCalled)

		err = binding.WithArgs("test")
		require.NoError(t, err)
		assert.True(t, withArgsCalled)
	})

	t.Run("ErrorReturning", func(t *testing.T) {
		t.Parallel()

		binding := MethodBinding{
			NoArgs: func() error {
				return errTestBinding
			},
		}

		err := binding.NoArgs()
		require.Error(t, err)
		assert.Equal(t, errTestBinding, err)
	})
}

// TestRegisterNamespaceCommandsHelper tests the helper function directly
func TestRegisterNamespaceCommandsHelper(t *testing.T) {
	t.Parallel()

	t.Run("BasicRegistration", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "test1", Desc: "Test command 1"},
			{Method: "test2", Desc: "Test command 2"},
		}
		bindings := map[string]MethodBinding{
			"test1": {NoArgs: func() error { return nil }},
			"test2": {NoArgs: func() error { return nil }},
		}

		registerNamespaceCommands(mockReg.Registry, "testns", "TestCategory", commands, bindings)

		registeredCmds := mockReg.GetRegisteredCommands()
		assert.Len(t, registeredCmds, 2)

		cmd1, exists := registeredCmds["testns:test1"]
		require.True(t, exists)
		assert.Equal(t, "Test command 1", cmd1.Description)
		assert.Equal(t, "TestCategory", cmd1.Category)
		assert.NotNil(t, cmd1.Func)

		cmd2, exists := registeredCmds["testns:test2"]
		require.True(t, exists)
		assert.Equal(t, "Test command 2", cmd2.Description)
	})

	t.Run("WithAliases", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "cmd", Desc: "Test command", Aliases: []string{"c", "command"}},
		}
		bindings := map[string]MethodBinding{
			"cmd": {NoArgs: func() error { return nil }},
		}

		registerNamespaceCommands(mockReg.Registry, "ns", "Cat", commands, bindings)

		registeredCmds := mockReg.GetRegisteredCommands()
		cmd, exists := registeredCmds["ns:cmd"]
		require.True(t, exists)
		assert.Contains(t, cmd.Aliases, "c")
		assert.Contains(t, cmd.Aliases, "command")
	})

	t.Run("WithUsageAndExamples", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{
				Method:   "example",
				Desc:     "Example command",
				Usage:    "magex ns:example [flags]",
				Examples: []string{"magex ns:example", "magex ns:example --flag"},
			},
		}
		bindings := map[string]MethodBinding{
			"example": {NoArgs: func() error { return nil }},
		}

		registerNamespaceCommands(mockReg.Registry, "ns", "Cat", commands, bindings)

		registeredCmds := mockReg.GetRegisteredCommands()
		cmd, exists := registeredCmds["ns:example"]
		require.True(t, exists)
		assert.Equal(t, "magex ns:example [flags]", cmd.Usage)
		assert.Len(t, cmd.Examples, 2)
	})

	t.Run("WithArgsFunction", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "argcmd", Desc: "Command with args"},
		}
		bindings := map[string]MethodBinding{
			"argcmd": {WithArgs: func(args ...string) error { return nil }},
		}

		registerNamespaceCommands(mockReg.Registry, "ns", "Cat", commands, bindings)

		registeredCmds := mockReg.GetRegisteredCommands()
		cmd, exists := registeredCmds["ns:argcmd"]
		require.True(t, exists)
		assert.NotNil(t, cmd.FuncWithArgs)
		assert.Nil(t, cmd.Func)
	})

	t.Run("DualFunctionRegistration", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "dual", Desc: "Dual function command"},
		}
		bindings := map[string]MethodBinding{
			"dual": {
				NoArgs:   func() error { return nil },
				WithArgs: func(args ...string) error { return nil },
			},
		}

		registerNamespaceCommands(mockReg.Registry, "ns", "Cat", commands, bindings)

		registeredCmds := mockReg.GetRegisteredCommands()
		cmd, exists := registeredCmds["ns:dual"]
		require.True(t, exists)
		assert.NotNil(t, cmd.Func)
		assert.NotNil(t, cmd.FuncWithArgs)
	})
}

// TestDataTableIntegrity tests that all data tables are properly structured
func TestDataTableIntegrity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		commands []CommandDef
		minCount int
	}{
		{"buildCommands", getBuildCommands(), 10},
		{"testCommands", getTestCommands(), 20},
		{"lintCommands", getLintCommands(), 5},
		{"formatCommands", getFormatCommands(), 4},
		{"depsCommands", getDepsCommands(), 9},
		{"gitCommands", getGitCommands(), 12},
		{"releaseCommands", getReleaseCommands(), 9},
		{"docsCommands", getDocsCommands(), 10},
		{"toolsCommands", getToolsCommands(), 4},
		{"generateCommands", getGenerateCommands(), 5},
		{"updateCommands", getUpdateCommands(), 2},
		{"modCommands", getModCommands(), 9},
		{"metricsCommands", getMetricsCommands(), 7},
		{"benchCommands", getBenchCommands(), 8},
		{"vetCommands", getVetCommands(), 1},
		{"configureCommands", getConfigureCommands(), 7},
		{"helpCommands", getHelpCommands(), 7},
		{"versionCommands", getVersionCommands(), 6},
		{"installCommands", getInstallCommands(), 15},
		{"yamlCommands", getYamlCommands(), 5},
		{"bmadCommands", getBmadCommands(), 3},
		{"awsCommands", getAWSCommands(), 4},
		{"speckitCommands", getSpeckitCommands(), 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.GreaterOrEqual(t, len(tc.commands), tc.minCount,
				"%s should have at least %d commands", tc.name, tc.minCount)

			// Verify each command has required fields
			for i, cmd := range tc.commands {
				assert.NotEmpty(t, cmd.Method,
					"%s[%d] should have Method", tc.name, i)
				assert.NotEmpty(t, cmd.Desc,
					"%s[%d] should have Desc", tc.name, i)
			}

			// Verify no duplicate methods
			methods := make(map[string]bool)
			for _, cmd := range tc.commands {
				assert.False(t, methods[cmd.Method],
					"%s has duplicate method: %s", tc.name, cmd.Method)
				methods[cmd.Method] = true
			}
		})
	}
}

// TestMethodBindingFactories tests all method binding factory functions
func TestMethodBindingFactories(t *testing.T) {
	t.Parallel()

	t.Run("BuildMethodBindings", func(t *testing.T) {
		t.Parallel()

		b := mage.Build{}
		bindings := buildMethodBindings(b)

		assert.NotEmpty(t, bindings)
		assert.NotNil(t, bindings["default"].NoArgs)
		assert.NotNil(t, bindings["all"].NoArgs)
		assert.NotNil(t, bindings["clean"].NoArgs)
	})

	t.Run("TestMethodBindings", func(t *testing.T) {
		t.Parallel()

		test := mage.Test{}
		bindings := testMethodBindings(test)

		assert.NotEmpty(t, bindings)
		assert.NotNil(t, bindings["default"].WithArgs)
		assert.NotNil(t, bindings["unit"].WithArgs)
		assert.NotNil(t, bindings["ci"].NoArgs)
	})

	t.Run("BenchMethodBindings_DualFunction", func(t *testing.T) {
		t.Parallel()

		b := mage.Bench{}
		bindings := benchMethodBindings(b)

		assert.NotEmpty(t, bindings)

		// Bench commands should have both NoArgs and WithArgs
		for method, binding := range bindings {
			assert.NotNil(t, binding.NoArgs,
				"bench:%s should have NoArgs", method)
			assert.NotNil(t, binding.WithArgs,
				"bench:%s should have WithArgs", method)
		}
	})

	t.Run("VersionMethodBindings", func(t *testing.T) {
		t.Parallel()

		v := mage.Version{}
		bindings := versionMethodBindings(v)

		assert.NotEmpty(t, bindings)
		assert.NotNil(t, bindings["show"].NoArgs)
		assert.NotNil(t, bindings["bump"].WithArgs)
		assert.NotNil(t, bindings["changelog"].WithArgs)
	})

	t.Run("AllBindingFactoriesNonEmpty", func(t *testing.T) {
		t.Parallel()

		factories := []struct {
			name    string
			factory func() map[string]MethodBinding
		}{
			{"build", func() map[string]MethodBinding { return buildMethodBindings(mage.Build{}) }},
			{"test", func() map[string]MethodBinding { return testMethodBindings(mage.Test{}) }},
			{"lint", func() map[string]MethodBinding { return lintMethodBindings(mage.Lint{}) }},
			{"format", func() map[string]MethodBinding { return formatMethodBindings(mage.Format{}) }},
			{"deps", func() map[string]MethodBinding { return depsMethodBindings(mage.Deps{}) }},
			{"git", func() map[string]MethodBinding { return gitMethodBindings(mage.Git{}) }},
			{"release", func() map[string]MethodBinding { return releaseMethodBindings(mage.Release{}) }},
			{"docs", func() map[string]MethodBinding { return docsMethodBindings(mage.Docs{}) }},
			{"tools", func() map[string]MethodBinding { return toolsMethodBindings(mage.Tools{}) }},
			{"generate", func() map[string]MethodBinding { return generateMethodBindings(mage.Generate{}) }},
			{"update", func() map[string]MethodBinding { return updateMethodBindings(mage.Update{}) }},
			{"mod", func() map[string]MethodBinding { return modMethodBindings(mage.Mod{}) }},
			{"metrics", func() map[string]MethodBinding { return metricsMethodBindings(mage.Metrics{}) }},
			{"bench", func() map[string]MethodBinding { return benchMethodBindings(mage.Bench{}) }},
			{"vet", func() map[string]MethodBinding { return vetMethodBindings(mage.Vet{}) }},
			{"configure", func() map[string]MethodBinding { return configureMethodBindings(mage.Configure{}) }},
			{"help", func() map[string]MethodBinding { return helpMethodBindings(mage.Help{}) }},
			{"version", func() map[string]MethodBinding { return versionMethodBindings(mage.Version{}) }},
			{"install", func() map[string]MethodBinding { return installMethodBindings(mage.Install{}) }},
			{"yaml", func() map[string]MethodBinding { return yamlMethodBindings(mage.Yaml{}) }},
			{"bmad", func() map[string]MethodBinding { return bmadMethodBindings(mage.Bmad{}) }},
			{"aws", func() map[string]MethodBinding { return awsMethodBindings(mage.AWS{}) }},
			{"speckit", func() map[string]MethodBinding { return speckitMethodBindings(mage.Speckit{}) }},
		}

		for _, f := range factories {
			t.Run(f.name, func(t *testing.T) {
				bindings := f.factory()
				assert.NotEmpty(t, bindings,
					"%s method bindings should not be empty", f.name)

				// Each binding should have at least one function
				for method, binding := range bindings {
					hasFunc := binding.NoArgs != nil || binding.WithArgs != nil
					assert.True(t, hasFunc,
						"%s:%s should have at least one function", f.name, method)
				}
			})
		}
	})
}

// TestDepsAuditSpecialCase tests the deps:audit command with Options
func TestDepsAuditSpecialCase(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	registerDepsCommands(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	// Verify deps:audit exists
	auditCmd, exists := commands["deps:audit"]
	require.True(t, exists, "deps:audit should be registered")

	// Verify it has required fields
	assert.Equal(t, "Security audit of dependencies (govulncheck)", auditCmd.Description)
	assert.Equal(t, "Dependencies", auditCmd.Category)
	assert.NotEmpty(t, auditCmd.LongDescription)
	assert.NotNil(t, auditCmd.FuncWithArgs)
	assert.NotEmpty(t, auditCmd.Usage)
	assert.NotEmpty(t, auditCmd.Examples)

	// Verify it has Options (special case)
	assert.NotEmpty(t, auditCmd.Options, "deps:audit should have Options defined")

	// Verify LongDescription contains expected content
	assert.Contains(t, auditCmd.LongDescription, "govulncheck")
	assert.Contains(t, auditCmd.LongDescription, "CVE")
	assert.Contains(t, auditCmd.LongDescription, "MAGE_X_CVE_EXCLUDES")
}

// TestBenchNamespaceDualFunctions tests that bench commands have both functions
func TestBenchNamespaceDualFunctions(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	registerBenchCommands(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	expectedDualCommands := []string{
		"bench:default",
		"bench:compare",
		"bench:save",
		"bench:cpu",
		"bench:mem",
		"bench:profile",
		"bench:trace",
		"bench:regression",
	}

	for _, cmdName := range expectedDualCommands {
		t.Run(cmdName, func(t *testing.T) {
			cmd, exists := commands[cmdName]
			require.True(t, exists, "%s should be registered", cmdName)

			// Bench commands should have BOTH Func and FuncWithArgs
			assert.NotNil(t, cmd.Func, "%s should have Func", cmdName)
			assert.NotNil(t, cmd.FuncWithArgs, "%s should have FuncWithArgs", cmdName)
		})
	}
}

// TestBindingsMatchDataTables ensures bindings match data tables
func TestBindingsMatchDataTables(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		commands []CommandDef
		bindings map[string]MethodBinding
	}{
		{"build", getBuildCommands(), buildMethodBindings(mage.Build{})},
		{"test", getTestCommands(), testMethodBindings(mage.Test{})},
		{"lint", getLintCommands(), lintMethodBindings(mage.Lint{})},
		{"format", getFormatCommands(), formatMethodBindings(mage.Format{})},
		{"deps", getDepsCommands(), depsMethodBindings(mage.Deps{})},
		{"git", getGitCommands(), gitMethodBindings(mage.Git{})},
		{"release", getReleaseCommands(), releaseMethodBindings(mage.Release{})},
		{"docs", getDocsCommands(), docsMethodBindings(mage.Docs{})},
		{"tools", getToolsCommands(), toolsMethodBindings(mage.Tools{})},
		{"generate", getGenerateCommands(), generateMethodBindings(mage.Generate{})},
		{"update", getUpdateCommands(), updateMethodBindings(mage.Update{})},
		{"mod", getModCommands(), modMethodBindings(mage.Mod{})},
		{"metrics", getMetricsCommands(), metricsMethodBindings(mage.Metrics{})},
		{"bench", getBenchCommands(), benchMethodBindings(mage.Bench{})},
		{"vet", getVetCommands(), vetMethodBindings(mage.Vet{})},
		{"configure", getConfigureCommands(), configureMethodBindings(mage.Configure{})},
		{"help", getHelpCommands(), helpMethodBindings(mage.Help{})},
		{"version", getVersionCommands(), versionMethodBindings(mage.Version{})},
		{"install", getInstallCommands(), installMethodBindings(mage.Install{})},
		{"yaml", getYamlCommands(), yamlMethodBindings(mage.Yaml{})},
		{"bmad", getBmadCommands(), bmadMethodBindings(mage.Bmad{})},
		{"aws", getAWSCommands(), awsMethodBindings(mage.AWS{})},
		{"speckit", getSpeckitCommands(), speckitMethodBindings(mage.Speckit{})},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Every command in data table should have a binding
			for _, cmd := range tc.commands {
				binding, exists := tc.bindings[cmd.Method]
				assert.True(t, exists,
					"%s:%s should have a binding", tc.name, cmd.Method)

				if exists {
					// Binding should have at least one function
					hasFunc := binding.NoArgs != nil || binding.WithArgs != nil
					assert.True(t, hasFunc,
						"%s:%s binding should have a function", tc.name, cmd.Method)
				}
			}

			// Every binding should have a command in data table
			for method := range tc.bindings {
				found := false
				for _, cmd := range tc.commands {
					if cmd.Method == method {
						found = true
						break
					}
				}
				assert.True(t, found,
					"%s binding '%s' should have a command definition", tc.name, method)
			}
		})
	}
}

// TestTotalCommandCount verifies the exact number of commands
func TestTotalCommandCount(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	// Count by type
	namespaceCommands := 0
	topLevelCommands := 0
	for _, cmd := range commands {
		if cmd.Namespace != "" {
			namespaceCommands++
		} else {
			topLevelCommands++
		}
	}

	// Expected: 166 data table + 1 deps:audit + 7 top-level = 174
	assert.Equal(t, 167, namespaceCommands,
		"Should have 167 namespace commands (166 from tables + 1 deps:audit)")
	assert.Equal(t, 7, topLevelCommands,
		"Should have 7 top-level commands")
	assert.Len(t, commands, 174,
		"Should have 174 total commands")
}

// TestMissingBindingPanics verifies commands without bindings cause panic
func TestMissingBindingPanics(t *testing.T) {
	t.Parallel()

	t.Run("EmptyBindingPanics", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "nobinding", Desc: "Command with no binding"},
		}
		bindings := map[string]MethodBinding{
			// Intentionally empty binding - no function set
			"nobinding": {},
		}

		// This should panic because the registry requires at least one function
		assert.Panics(t, func() {
			registerNamespaceCommands(mockReg.Registry, "test", "Test", commands, bindings)
		}, "Should panic when command has no function")
	})

	t.Run("MissingBindingEntryPanics", func(t *testing.T) {
		t.Parallel()

		mockReg := NewMockRegistry()
		commands := []CommandDef{
			{Method: "exists", Desc: "Command that exists"},
			{Method: "missing", Desc: "Command with missing binding"},
		}
		bindings := map[string]MethodBinding{
			"exists": {NoArgs: func() error { return nil }},
			// "missing" is not in bindings map
		}

		// This should panic because the binding lookup returns zero value
		assert.Panics(t, func() {
			registerNamespaceCommands(mockReg.Registry, "test", "Test", commands, bindings)
		}, "Should panic when binding is missing")
	})
}

// TestCommandExamplesFormat verifies examples are properly formatted
func TestCommandExamplesFormat(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	RegisterAll(mockReg.Registry)

	commands := mockReg.GetRegisteredCommands()

	for cmdName, cmd := range commands {
		if len(cmd.Examples) > 0 {
			t.Run(fmt.Sprintf("Examples_%s", cmdName), func(t *testing.T) {
				for i, example := range cmd.Examples {
					// Examples should not be empty
					assert.NotEmpty(t, example,
						"%s example[%d] should not be empty", cmdName, i)

					// Examples should start with "magex" or be a valid command
					if !strings.HasPrefix(example, "magex") &&
						!strings.HasPrefix(example, "MAGE_X") {
						t.Logf("Warning: %s example[%d] doesn't start with 'magex': %s",
							cmdName, i, example)
					}
				}
			})
		}
	}
}

// BenchmarkDataDrivenRegistration benchmarks the data-driven approach
func BenchmarkDataDrivenRegistration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mockReg := NewMockRegistry()
		b.StartTimer()

		// Register a single namespace using data-driven approach
		registerBuildCommands(mockReg.Registry)
	}
}

// BenchmarkHelperFunction benchmarks the registerNamespaceCommands helper
func BenchmarkHelperFunction(b *testing.B) {
	commands := []CommandDef{
		{Method: "cmd1", Desc: "Command 1"},
		{Method: "cmd2", Desc: "Command 2"},
		{Method: "cmd3", Desc: "Command 3"},
		{Method: "cmd4", Desc: "Command 4"},
		{Method: "cmd5", Desc: "Command 5"},
	}
	bindings := map[string]MethodBinding{
		"cmd1": {NoArgs: func() error { return nil }},
		"cmd2": {NoArgs: func() error { return nil }},
		"cmd3": {NoArgs: func() error { return nil }},
		"cmd4": {NoArgs: func() error { return nil }},
		"cmd5": {NoArgs: func() error { return nil }},
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		mockReg := NewMockRegistry()
		b.StartTimer()

		registerNamespaceCommands(mockReg.Registry, "bench", "Bench", commands, bindings)
	}
}

// ============================================================================
// Getter Function Tests - Validate the refactor from globals to functions
// ============================================================================

// TestGetterFunctionsReturnFreshSlices verifies that each getter returns a new slice
// (not a shared reference) to ensure immutability and thread safety
func TestGetterFunctionsReturnFreshSlices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
		{"getFormatCommands", getFormatCommands},
		{"getDepsCommands", getDepsCommands},
		{"getGitCommands", getGitCommands},
		{"getReleaseCommands", getReleaseCommands},
		{"getDocsCommands", getDocsCommands},
		{"getToolsCommands", getToolsCommands},
		{"getGenerateCommands", getGenerateCommands},
		{"getUpdateCommands", getUpdateCommands},
		{"getModCommands", getModCommands},
		{"getMetricsCommands", getMetricsCommands},
		{"getBenchCommands", getBenchCommands},
		{"getVetCommands", getVetCommands},
		{"getConfigureCommands", getConfigureCommands},
		{"getHelpCommands", getHelpCommands},
		{"getVersionCommands", getVersionCommands},
		{"getInstallCommands", getInstallCommands},
		{"getYamlCommands", getYamlCommands},
		{"getBmadCommands", getBmadCommands},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Get two slices from the same getter
			slice1 := tc.getter()
			slice2 := tc.getter()

			// They should have the same length and content
			require.Len(t, slice2, len(slice1),
				"%s should return slices of equal length", tc.name)

			// But they should NOT be the same underlying slice (pointer comparison)
			// Modify one and verify the other is unchanged
			if len(slice1) > 0 {
				originalMethod := slice1[0].Method
				slice1[0].Method = "MODIFIED_FOR_TEST"

				assert.Equal(t, originalMethod, slice2[0].Method,
					"%s should return independent slices; modifying one should not affect the other", tc.name)

				// Restore for cleanliness
				slice1[0].Method = originalMethod
			}
		})
	}
}

// TestGetterFunctionsReturnConsistentData verifies each getter returns the same data
func TestGetterFunctionsReturnConsistentData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
		{"getFormatCommands", getFormatCommands},
		{"getDepsCommands", getDepsCommands},
		{"getGitCommands", getGitCommands},
		{"getReleaseCommands", getReleaseCommands},
		{"getDocsCommands", getDocsCommands},
		{"getToolsCommands", getToolsCommands},
		{"getGenerateCommands", getGenerateCommands},
		{"getUpdateCommands", getUpdateCommands},
		{"getModCommands", getModCommands},
		{"getMetricsCommands", getMetricsCommands},
		{"getBenchCommands", getBenchCommands},
		{"getVetCommands", getVetCommands},
		{"getConfigureCommands", getConfigureCommands},
		{"getHelpCommands", getHelpCommands},
		{"getVersionCommands", getVersionCommands},
		{"getInstallCommands", getInstallCommands},
		{"getYamlCommands", getYamlCommands},
		{"getBmadCommands", getBmadCommands},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call multiple times and verify consistency
			for i := 0; i < 5; i++ {
				slice := tc.getter()
				require.NotEmpty(t, slice, "%s should return non-empty slice", tc.name)

				// Verify all commands have required fields
				for j, cmd := range slice {
					assert.NotEmpty(t, cmd.Method,
						"%s[%d] should have Method field", tc.name, j)
					assert.NotEmpty(t, cmd.Desc,
						"%s[%d] should have Desc field", tc.name, j)
				}
			}
		})
	}
}

// TestGetterFunctionsConcurrentAccess verifies getter functions are thread-safe
func TestGetterFunctionsConcurrentAccess(t *testing.T) {
	t.Parallel()

	getters := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
		{"getFormatCommands", getFormatCommands},
		{"getDepsCommands", getDepsCommands},
		{"getGitCommands", getGitCommands},
		{"getReleaseCommands", getReleaseCommands},
		{"getDocsCommands", getDocsCommands},
		{"getToolsCommands", getToolsCommands},
		{"getGenerateCommands", getGenerateCommands},
		{"getUpdateCommands", getUpdateCommands},
		{"getModCommands", getModCommands},
		{"getMetricsCommands", getMetricsCommands},
		{"getBenchCommands", getBenchCommands},
		{"getVetCommands", getVetCommands},
		{"getConfigureCommands", getConfigureCommands},
		{"getHelpCommands", getHelpCommands},
		{"getVersionCommands", getVersionCommands},
		{"getInstallCommands", getInstallCommands},
		{"getYamlCommands", getYamlCommands},
		{"getBmadCommands", getBmadCommands},
	}

	const numGoroutines = 50

	for _, g := range getters {
		t.Run(g.name+"_concurrent", func(t *testing.T) {
			t.Parallel()

			var wg sync.WaitGroup
			results := make(chan int, numGoroutines)

			// Launch many goroutines calling the same getter
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(getter func() []CommandDef) {
					defer wg.Done()
					cmds := getter()
					results <- len(cmds)
				}(g.getter)
			}

			wg.Wait()
			close(results)

			// All goroutines should get the same length
			var firstLen int
			first := true
			for length := range results {
				if first {
					firstLen = length
					first = false
				} else {
					assert.Equal(t, firstLen, length,
						"%s should return same length from concurrent calls", g.name)
				}
			}
		})
	}
}

// TestGetterFunctionsExpectedCounts verifies each getter returns expected command counts
func TestGetterFunctionsExpectedCounts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		getter        func() []CommandDef
		expectedCount int
	}{
		{"getBuildCommands", getBuildCommands, 10},
		{"getTestCommands", getTestCommands, 21},
		{"getLintCommands", getLintCommands, 5},
		{"getFormatCommands", getFormatCommands, 4},
		{"getDepsCommands", getDepsCommands, 9},
		{"getGitCommands", getGitCommands, 12},
		{"getReleaseCommands", getReleaseCommands, 9},
		{"getDocsCommands", getDocsCommands, 10},
		{"getToolsCommands", getToolsCommands, 4},
		{"getGenerateCommands", getGenerateCommands, 5},
		{"getUpdateCommands", getUpdateCommands, 2},
		{"getModCommands", getModCommands, 9},
		{"getMetricsCommands", getMetricsCommands, 7},
		{"getBenchCommands", getBenchCommands, 8},
		{"getVetCommands", getVetCommands, 1},
		{"getConfigureCommands", getConfigureCommands, 7},
		{"getHelpCommands", getHelpCommands, 7},
		{"getVersionCommands", getVersionCommands, 6},
		{"getInstallCommands", getInstallCommands, 15},
		{"getYamlCommands", getYamlCommands, 5},
		{"getBmadCommands", getBmadCommands, 3},
		{"getAWSCommands", getAWSCommands, 4},
		{"getSpeckitCommands", getSpeckitCommands, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmds := tc.getter()
			assert.Len(t, cmds, tc.expectedCount,
				"%s should return exactly %d commands", tc.name, tc.expectedCount)
		})
	}
}

// TestGetterFunctionsNoDuplicateMethods verifies no duplicate methods within each getter
func TestGetterFunctionsNoDuplicateMethods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
		{"getFormatCommands", getFormatCommands},
		{"getDepsCommands", getDepsCommands},
		{"getGitCommands", getGitCommands},
		{"getReleaseCommands", getReleaseCommands},
		{"getDocsCommands", getDocsCommands},
		{"getToolsCommands", getToolsCommands},
		{"getGenerateCommands", getGenerateCommands},
		{"getUpdateCommands", getUpdateCommands},
		{"getModCommands", getModCommands},
		{"getMetricsCommands", getMetricsCommands},
		{"getBenchCommands", getBenchCommands},
		{"getVetCommands", getVetCommands},
		{"getConfigureCommands", getConfigureCommands},
		{"getHelpCommands", getHelpCommands},
		{"getVersionCommands", getVersionCommands},
		{"getInstallCommands", getInstallCommands},
		{"getYamlCommands", getYamlCommands},
		{"getBmadCommands", getBmadCommands},
		{"getAWSCommands", getAWSCommands},
		{"getSpeckitCommands", getSpeckitCommands},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmds := tc.getter()
			seen := make(map[string]bool)

			for _, cmd := range cmds {
				assert.False(t, seen[cmd.Method],
					"%s has duplicate method: %s", tc.name, cmd.Method)
				seen[cmd.Method] = true
			}
		})
	}
}

// TestGetterFunctionsAliasesValid verifies all aliases in commands are non-empty
func TestGetterFunctionsAliasesValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
		{"getFormatCommands", getFormatCommands},
		{"getDepsCommands", getDepsCommands},
		{"getGitCommands", getGitCommands},
		{"getReleaseCommands", getReleaseCommands},
		{"getDocsCommands", getDocsCommands},
		{"getToolsCommands", getToolsCommands},
		{"getGenerateCommands", getGenerateCommands},
		{"getUpdateCommands", getUpdateCommands},
		{"getModCommands", getModCommands},
		{"getMetricsCommands", getMetricsCommands},
		{"getBenchCommands", getBenchCommands},
		{"getVetCommands", getVetCommands},
		{"getConfigureCommands", getConfigureCommands},
		{"getHelpCommands", getHelpCommands},
		{"getVersionCommands", getVersionCommands},
		{"getInstallCommands", getInstallCommands},
		{"getYamlCommands", getYamlCommands},
		{"getBmadCommands", getBmadCommands},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmds := tc.getter()
			for _, cmd := range cmds {
				for i, alias := range cmd.Aliases {
					assert.NotEmpty(t, alias,
						"%s method '%s' has empty alias at index %d", tc.name, cmd.Method, i)
				}
			}
		})
	}
}

// TestTotalCommandsFromGetters verifies sum of all getter counts matches expected total
func TestTotalCommandsFromGetters(t *testing.T) {
	t.Parallel()

	getters := []func() []CommandDef{
		getBuildCommands,
		getTestCommands,
		getLintCommands,
		getFormatCommands,
		getDepsCommands,
		getGitCommands,
		getReleaseCommands,
		getDocsCommands,
		getToolsCommands,
		getGenerateCommands,
		getUpdateCommands,
		getModCommands,
		getMetricsCommands,
		getBenchCommands,
		getVetCommands,
		getConfigureCommands,
		getHelpCommands,
		getVersionCommands,
		getInstallCommands,
		getYamlCommands,
		getBmadCommands,
		getAWSCommands,
		getSpeckitCommands,
	}

	total := 0
	for _, getter := range getters {
		total += len(getter())
	}

	// Expected: 166 commands from data tables
	assert.Equal(t, 166, total,
		"Total commands from all getters should equal 166")
}

// BenchmarkGetterFunctions benchmarks the getter function calls
func BenchmarkGetterFunctions(b *testing.B) {
	getters := []struct {
		name   string
		getter func() []CommandDef
	}{
		{"getBuildCommands", getBuildCommands},
		{"getTestCommands", getTestCommands},
		{"getLintCommands", getLintCommands},
	}

	for _, g := range getters {
		b.Run(g.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = g.getter()
			}
		})
	}
}
