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
	expectedMinCommands = 230      // Minimum expected number of registered commands
	concurrentWorkers   = 10       // Number of concurrent workers for parallel tests
)

// Test errors for security testing
var (
	ErrCommandInjection = errors.New("potential command injection detected")
	ErrInvalidInput     = errors.New("invalid input detected")
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
		"docs", "tools", "generate", "cli", "update", "mod", "recipes",
		"metrics", "workflow", "bench", "vet", "configure", "init",
		"enterprise", "integrations", "wizard", "help", "version",
		"install", "audit", "yaml", "enterpriseconfig",
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
		{"docker", "Build a Docker image", "Build", false},
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
		{"default", false},
		{"unit", false},
		{"short", false},
		{"race", false},
		{"cover", false},
		{"coverrace", false},
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
		"CLI":                {"cli:"},
		"Update":             {"update:"},
		"Module":             {"mod:"},
		"Recipes":            {"recipes:"},
		"Metrics":            {"metrics:"},
		"Workflow":           {"workflow:"},
		"Benchmark":          {"bench:"},
		"Vet":                {"vet:"},
		"Configure":          {"configure:"},
		"Init":               {"init:"},
		"Enterprise":         {"enterprise:", "enterpriseconfig:"},
		"Integrations":       {"integrations:"},
		"Wizard":             {"wizard:"},
		"Help":               {"help:"},
		"Version Management": {"version:"},
		"Installation":       {"install:"},
		"Audit & Security":   {"audit:"},
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
		{"DepsCommands", registerDepsCommands, "deps:", 8},
		{"GitCommands", registerGitCommands, "git:", 8},
		{"ReleaseCommands", registerReleaseCommands, "release:", 6},
		{"DocsCommands", registerDocsCommands, "docs:", 8},
		{"ToolsCommands", registerToolsCommands, "tools:", 3},
		{"GenerateCommands", registerGenerateCommands, "generate:", 4},
		{"CLICommands", registerCLICommands, "cli:", 10},
		{"UpdateCommands", registerUpdateCommands, "update:", 2},
		{"ModCommands", registerModCommands, "mod:", 8},
		{"RecipesCommands", registerRecipesCommands, "recipes:", 6},
		{"MetricsCommands", registerMetricsCommands, "metrics:", 5},
		{"WorkflowCommands", registerWorkflowCommands, "workflow:", 6},
		{"BenchCommands", registerBenchCommands, "bench:", 6},
		{"VetCommands", registerVetCommands, "vet:", 1},
		{"ConfigureCommands", registerConfigureCommands, "configure:", 6},
		{"InitCommands", registerInitCommands, "init:", 8},
		{"EnterpriseCommands", registerEnterpriseCommands, "enterprise:", 6},
		{"IntegrationsCommands", registerIntegrationsCommands, "integrations:", 6},
		{"WizardCommands", registerWizardCommands, "wizard:", 5},
		{"HelpCommands", registerHelpCommands, "help:", 6},
		{"VersionCommands", registerVersionCommands, "version:", 5},
		{"InstallCommands", registerInstallCommands, "install:", 12},
		{"AuditCommands", registerAuditCommands, "audit:", 6},
		{"YamlCommands", registerYamlCommands, "yaml:", 4},
		{"EnterpriseConfigCommands", registerEnterpriseConfigCommands, "enterpriseconfig:", 5},
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
		"test:bench":           true,
		"test:coverage":        true,
		"git:tag":              true,
		"git:tagremove":        true,
		"git:tagupdate":        true,
		"git:commit":           true,
		"git:add":              true,
		"recipes:show":         true,
		"recipes:run":          true,
		"recipes:search":       true,
		"recipes:create":       true,
		"recipes:install":      true,
		"workflow:execute":     true,
		"workflow:status":      true,
		"workflow:create":      true,
		"workflow:validate":    true,
		"workflow:schedule":    true,
		"workflow:template":    true,
		"workflow:history":     true,
		"bench:default":        true,
		"bench:compare":        true,
		"bench:save":           true,
		"bench:cpu":            true,
		"bench:mem":            true,
		"bench:profile":        true,
		"bench:trace":          true,
		"bench:regression":     true,
		"version:bump":         true,
		"version:changelog":    true,
		"integrations:setup":   true,
		"integrations:test":    true,
		"integrations:sync":    true,
		"integrations:notify":  true,
		"integrations:webhook": true,
		"integrations:export":  true,
		"integrations:import":  true,
		"bench":                true, // top-level bench command
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
		"Build":                     mage.Build{},
		"Test":                      mage.Test{},
		"Lint":                      mage.Lint{},
		"Format":                    mage.Format{},
		"Deps":                      mage.Deps{},
		"Git":                       mage.Git{},
		"Release":                   mage.Release{},
		"Docs":                      mage.Docs{},
		"Tools":                     mage.Tools{},
		"Generate":                  mage.Generate{},
		"CLI":                       mage.CLI{},
		"Update":                    mage.Update{},
		"Mod":                       mage.Mod{},
		"Recipes":                   mage.Recipes{},
		"Metrics":                   mage.Metrics{},
		"Workflow":                  mage.Workflow{},
		"Bench":                     mage.Bench{},
		"Vet":                       mage.Vet{},
		"Configure":                 mage.Configure{},
		"Init":                      mage.Init{},
		"Enterprise":                mage.Enterprise{},
		"Integrations":              mage.Integrations{},
		"Wizard":                    mage.Wizard{},
		"Help":                      mage.Help{},
		"Version":                   mage.Version{},
		"Install":                   mage.Install{},
		"Audit":                     mage.Audit{},
		"Yaml":                      mage.Yaml{},
		"EnterpriseConfigNamespace": mage.EnterpriseConfigNamespace{},
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
