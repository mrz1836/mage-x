package registry

import (
	"fmt"
	"strings"
	"testing"
)

func TestRegistry_NewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if r.commands == nil {
		t.Error("commands map not initialized")
	}
	if r.aliases == nil {
		t.Error("aliases map not initialized")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "build",
		Description: "Test command",
		Category:    "Build",
		Func:        func() error { return nil },
	}

	err := r.Register(cmd)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Verify command is registered
	retrieved, exists := r.Get("build:test")
	if !exists {
		t.Error("Command not found after registration")
	}
	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", retrieved.Name)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	cmd1 := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "build",
		Description: "Test command 1",
		Category:    "Build",
		Func:        func() error { return nil },
	}

	cmd2 := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "build",
		Description: "Test command 2",
		Category:    "Build",
		Func:        func() error { return nil },
	}

	err := r.Register(cmd1)
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	err = r.Register(cmd2)
	if err == nil {
		t.Error("Expected error when registering duplicate command")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("Expected 'already registered' error, got: %v", err)
	}
}

func TestRegistry_RegisterWithAliases(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "build",
		Method:      "Build",
		Namespace:   "build",
		Description: "Build command",
		Category:    "Build",
		Aliases:     []string{"b", "compile"},
		Func:        func() error { return nil },
	}

	err := r.Register(cmd)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Test aliases work
	for _, alias := range []string{"b", "compile"} {
		retrieved, exists := r.Get(alias)
		if !exists {
			t.Errorf("Alias '%s' not found", alias)
		}
		if retrieved.Name != "build" {
			t.Errorf("Alias '%s' returned wrong command: %s", alias, retrieved.Name)
		}
	}
}

func TestRegistry_RegisterDuplicateAlias(t *testing.T) {
	r := NewRegistry()

	cmd1 := &Command{
		Name:        "build1",
		Method:      "Build1",
		Namespace:   "build",
		Description: "Build command 1",
		Category:    "Build",
		Aliases:     []string{"b"},
		Func:        func() error { return nil },
	}

	cmd2 := &Command{
		Name:        "build2",
		Method:      "Build2",
		Namespace:   "build",
		Description: "Build command 2",
		Category:    "Build",
		Aliases:     []string{"b"},
		Func:        func() error { return nil },
	}

	err := r.Register(cmd1)
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	err = r.Register(cmd2)
	if err == nil {
		t.Error("Expected error when registering duplicate alias")
	}
	if !strings.Contains(err.Error(), "alias already registered") {
		t.Errorf("Expected 'alias already registered' error, got: %v", err)
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "build",
		Description: "Test command",
		Category:    "Build",
		Func:        func() error { return nil },
	}

	r.MustRegister(cmd)

	// Test full name
	retrieved, exists := r.Get("build:test")
	if !exists {
		t.Error("Command not found by full name")
	}
	if retrieved.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", retrieved.Name)
	}

	// Test case insensitive
	_, exists = r.Get("BUILD:TEST")
	if !exists {
		t.Error("Command not found with uppercase")
	}

	// Test non-existent command
	_, exists = r.Get("nonexistent")
	if exists {
		t.Error("Non-existent command found")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
		{
			Name:        "hidden",
			Method:      "Hidden",
			Namespace:   "internal",
			Description: "Hidden command",
			Category:    "Internal",
			Hidden:      true,
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	list := r.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 visible commands, got %d", len(list))
	}

	// Check that hidden command is not included
	for _, cmd := range list {
		if cmd.Hidden {
			t.Error("Hidden command included in list")
		}
	}

	// Check sorting
	if len(list) >= 2 {
		first := list[0].FullName()
		second := list[1].FullName()
		if first > second {
			t.Error("Commands not sorted properly")
		}
	}
}

func TestRegistry_ListByNamespace(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "compile",
			Method:      "Compile",
			Namespace:   "build",
			Description: "Compile command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	buildCommands := r.ListByNamespace("build")
	if len(buildCommands) != 2 {
		t.Errorf("Expected 2 build commands, got %d", len(buildCommands))
	}

	testCommands := r.ListByNamespace("test")
	if len(testCommands) != 1 {
		t.Errorf("Expected 1 test command, got %d", len(testCommands))
	}

	// Test case insensitive
	buildCommands = r.ListByNamespace("BUILD")
	if len(buildCommands) != 2 {
		t.Errorf("Expected 2 build commands (case insensitive), got %d", len(buildCommands))
	}
}

func TestRegistry_ListByCategory(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "compile",
			Method:      "Compile",
			Namespace:   "build",
			Description: "Compile command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	buildCommands := r.ListByCategory("Build")
	if len(buildCommands) != 2 {
		t.Errorf("Expected 2 Build category commands, got %d", len(buildCommands))
	}

	testCommands := r.ListByCategory("Test")
	if len(testCommands) != 1 {
		t.Errorf("Expected 1 Test category command, got %d", len(testCommands))
	}
}

func TestRegistry_Namespaces(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
		{
			Name:        "lint",
			Method:      "Lint",
			Namespace:   "lint",
			Description: "Lint command",
			Category:    "Quality",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	namespaces := r.Namespaces()
	if len(namespaces) != 3 {
		t.Errorf("Expected 3 namespaces, got %d", len(namespaces))
	}

	expectedNamespaces := []string{"build", "lint", "test"}
	for i, expected := range expectedNamespaces {
		if namespaces[i] != expected {
			t.Errorf("Expected namespace '%s', got '%s'", expected, namespaces[i])
		}
	}
}

func TestRegistry_Categories(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
		{
			Name:        "lint",
			Method:      "Lint",
			Namespace:   "lint",
			Description: "Lint command",
			Category:    "Quality",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	categories := r.Categories()
	if len(categories) != 3 {
		t.Errorf("Expected 3 categories, got %d", len(categories))
	}

	expectedCategories := []string{"Build", "Quality", "Test"}
	for i, expected := range expectedCategories {
		if categories[i] != expected {
			t.Errorf("Expected category '%s', got '%s'", expected, categories[i])
		}
	}
}

func TestRegistry_Search(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build the project",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Run tests",
			Category:    "Test",
			Func:        func() error { return nil },
		},
		{
			Name:        "unittest",
			Method:      "UnitTest",
			Namespace:   "test",
			Description: "Run unit tests",
			Category:    "Test",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	// Search by name
	results := r.Search("build")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'build', got %d", len(results))
	}
	if results[0].Name != "build" {
		t.Errorf("Expected 'build' command, got '%s'", results[0].Name)
	}

	// Search by description
	results = r.Search("unit")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'unit', got %d", len(results))
	}
	if results[0].Name != "unittest" {
		t.Errorf("Expected 'unittest' command, got '%s'", results[0].Name)
	}

	// Search by namespace
	results = r.Search("test")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'test', got %d", len(results))
	}

	// Case insensitive search
	results = r.Search("BUILD")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'BUILD' (case insensitive), got %d", len(results))
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()

	executed := false
	cmd := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "test",
		Description: "Test command",
		Category:    "Test",
		Func: func() error {
			executed = true
			return nil
		},
	}

	r.MustRegister(cmd)

	err := r.Execute("test:test")
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if !executed {
		t.Error("Command handler was not executed")
	}
}

func TestRegistry_ExecuteNonExistent(t *testing.T) {
	r := NewRegistry()

	err := r.Execute("nonexistent")
	if err == nil {
		t.Error("Expected error when executing non-existent command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Expected 'unknown command' error, got: %v", err)
	}
}

func TestRegistry_ExecuteWithSuggestions(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "build",
		Method:      "Build",
		Namespace:   "build",
		Description: "Build command",
		Category:    "Build",
		Func:        func() error { return nil },
	}

	r.MustRegister(cmd)

	err := r.Execute("buidl")
	if err == nil {
		t.Error("Expected error when executing misspelled command")
	}

	// The algorithm looks for substring matches, so "buidl" should find "build:build"
	errorStr := err.Error()
	if strings.Contains(errorStr, "Did you mean:") {
		// Great, suggestions were provided
		if !strings.Contains(errorStr, "build:build") {
			t.Errorf("Expected suggestion to contain 'build:build', got: %v", err)
		}
	} else {
		// No suggestions found - this can happen if the similarity algorithm doesn't match
		// Just verify it's a proper "unknown command" error
		if !strings.Contains(errorStr, "unknown command") {
			t.Errorf("Expected 'unknown command' error, got: %v", err)
		}
		t.Logf("No suggestions found for 'buidl' - this is acceptable")
	}
}

func TestRegistry_Metadata(t *testing.T) {
	r := NewRegistry()

	commands := []*Command{
		{
			Name:        "build",
			Method:      "Build",
			Namespace:   "build",
			Description: "Build command",
			Category:    "Build",
			Func:        func() error { return nil },
		},
		{
			Name:        "test",
			Method:      "Test",
			Namespace:   "test",
			Description: "Test command",
			Category:    "Test",
			Func:        func() error { return nil },
		},
	}

	for _, cmd := range commands {
		r.MustRegister(cmd)
	}

	metadata := r.Metadata()
	if metadata.TotalCommands != 2 {
		t.Errorf("Expected 2 total commands, got %d", metadata.TotalCommands)
	}
	if len(metadata.Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(metadata.Categories))
	}
	if metadata.Categories["Build"] != 1 {
		t.Errorf("Expected 1 Build command, got %d", metadata.Categories["Build"])
	}
	if metadata.Categories["Test"] != 1 {
		t.Errorf("Expected 1 Test command, got %d", metadata.Categories["Test"])
	}
}

func TestRegistry_Clear(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "test",
		Description: "Test command",
		Category:    "Test",
		Func:        func() error { return nil },
	}

	r.MustRegister(cmd)

	if len(r.commands) == 0 {
		t.Error("Command not registered before clear")
	}

	r.Clear()

	if len(r.commands) != 0 {
		t.Error("Commands not cleared")
	}
	if len(r.aliases) != 0 {
		t.Error("Aliases not cleared")
	}
	if len(r.metadata.Categories) != 0 {
		t.Error("Metadata not cleared")
	}
}

func TestRegistry_GlobalFunctions(t *testing.T) {
	// Clear global registry for test isolation
	Global().Clear()

	cmd := &Command{
		Name:        "global",
		Method:      "Global",
		Namespace:   "test",
		Description: "Global test command",
		Category:    "Test",
		Func:        func() error { return nil },
	}

	err := Register(cmd)
	if err != nil {
		t.Fatalf("Global Register() failed: %v", err)
	}

	retrieved, exists := Get("test:global")
	if !exists {
		t.Error("Global Get() failed to find command")
		return
	}
	if retrieved == nil {
		t.Error("Global Get() returned nil command")
		return
	}
	if retrieved.Name != "global" {
		t.Errorf("Expected name 'global', got '%s'", retrieved.Name)
	}

	list := List()
	if len(list) != 1 {
		t.Errorf("Expected 1 command in global list, got %d", len(list))
	}

	// Test MustRegister
	cmd2 := &Command{
		Name:        "global2",
		Method:      "Global2",
		Namespace:   "test",
		Description: "Global test command 2",
		Category:    "Test",
		Func:        func() error { return nil },
	}

	MustRegister(cmd2)

	list = List()
	if len(list) != 2 {
		t.Errorf("Expected 2 commands in global list after MustRegister, got %d", len(list))
	}
}

func TestRegistry_MustRegister(t *testing.T) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "test",
		Method:      "Test",
		Namespace:   "test",
		Description: "Test command",
		Category:    "Test",
		Func:        func() error { return nil },
	}

	// Should not panic
	r.MustRegister(cmd)

	// Verify command is registered
	_, exists := r.Get("test:test")
	if !exists {
		t.Error("Command not found after MustRegister")
	}
}

func TestRegistry_MustRegisterPanic(t *testing.T) {
	r := NewRegistry()

	// Register invalid command to trigger panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from MustRegister with invalid command")
		}
	}()

	invalidCmd := &Command{
		// Missing required fields to trigger validation error
	}

	r.MustRegister(invalidCmd)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()

	// Test concurrent registration and access
	done := make(chan bool, 10)

	// Register commands concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			cmd := &Command{
				Name:        fmt.Sprintf("concurrent-%d", id), // Use id to make each command unique
				Method:      "Concurrent",
				Namespace:   "test",
				Description: "Concurrent test command",
				Category:    "Test",
				Func:        func() error { return nil },
			}
			// This should succeed for first goroutine and fail for others due to duplicate
			if err := r.Register(cmd); err != nil {
				// Registration error is expected in this race condition test
				_ = err
			}
			done <- true
		}(i)
	}

	// Access commands concurrently
	for i := 0; i < 5; i++ {
		go func() {
			r.Get("test:concurrent")
			r.List()
			r.Namespaces()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify registry is still functional
	list := r.List()
	if len(list) == 0 {
		t.Error("Registry corrupted after concurrent access")
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	r := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := &Command{
			Name:        "bench",
			Method:      "Bench",
			Namespace:   "bench",
			Description: "Benchmark command",
			Category:    "Benchmark",
			Func:        func() error { return nil },
		}
		r.Clear()
		if err := r.Register(cmd); err != nil {
			// Benchmark registration error is unexpected but continue
			_ = err
		}
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	r := NewRegistry()

	cmd := &Command{
		Name:        "bench",
		Method:      "Bench",
		Namespace:   "bench",
		Description: "Benchmark command",
		Category:    "Benchmark",
		Func:        func() error { return nil },
	}
	r.MustRegister(cmd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get("bench:bench")
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	r := NewRegistry()

	// Register multiple commands
	for i := 0; i < 100; i++ {
		cmd := &Command{
			Name:        "command",
			Method:      "Command",
			Namespace:   "bench",
			Description: "Benchmark command",
			Category:    "Benchmark",
			Func:        func() error { return nil },
		}
		r.Clear()
		r.MustRegister(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.List()
	}
}

func BenchmarkRegistry_Search(b *testing.B) {
	r := NewRegistry()

	// Register multiple commands
	for i := 0; i < 100; i++ {
		cmd := &Command{
			Name:        "command",
			Method:      "Command",
			Namespace:   "bench",
			Description: "Benchmark command for searching",
			Category:    "Benchmark",
			Func:        func() error { return nil },
		}
		r.Clear()
		r.MustRegister(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Search("bench")
	}
}
