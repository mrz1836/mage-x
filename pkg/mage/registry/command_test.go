package registry

import (
	"errors"
	"strings"
	"testing"
)

// Test errors
var (
	errTest     = errors.New("test error")
	errTestArgs = errors.New("test args error")
)

const (
	testCategory = "Test"
)

func TestCommand_FullName(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected string
	}{
		{
			name: "namespace and method",
			command: Command{
				Namespace: "Build",
				Method:    "Linux",
			},
			expected: "build:linux",
		},
		{
			name: "name only",
			command: Command{
				Name: "TestCommand",
			},
			expected: "testcommand",
		},
		{
			name: "empty method",
			command: Command{
				Namespace: "Build",
				Method:    "",
			},
			expected: "",
		},
		{
			name: "case sensitivity",
			command: Command{
				Namespace: "BUILD",
				Method:    "LINUX",
			},
			expected: "build:linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.FullName()
			if result != tt.expected {
				t.Errorf("FullName() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestCommand_IsNamespace(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		expected bool
	}{
		{
			name: "default method",
			command: Command{
				Namespace: "build",
				Method:    "default",
			},
			expected: true,
		},
		{
			name: "empty method",
			command: Command{
				Namespace: "build",
				Method:    "",
			},
			expected: true,
		},
		{
			name: "specific method",
			command: Command{
				Namespace: "build",
				Method:    "linux",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.IsNamespace()
			if result != tt.expected {
				t.Errorf("IsNamespace() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCommand_Validate(t *testing.T) {
	tests := []struct {
		name        string
		command     Command
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid command with name and func",
			command: Command{
				Name: "test",
				Func: func() error { return nil },
			},
			expectError: false,
		},
		{
			name: "valid command with namespace/method and func",
			command: Command{
				Namespace: "build",
				Method:    "linux",
				Func:      func() error { return nil },
			},
			expectError: false,
		},
		{
			name: "valid command with args func",
			command: Command{
				Name:         "test",
				FuncWithArgs: func(args ...string) error { return nil },
			},
			expectError: false,
		},
		{
			name: "invalid command - no name or namespace/method",
			command: Command{
				Func: func() error { return nil },
			},
			expectError: true,
			errorMsg:    "command must have either Name or Namespace+Method",
		},
		{
			name: "invalid command - partial namespace/method",
			command: Command{
				Namespace: "build",
				Func:      func() error { return nil },
			},
			expectError: true,
			errorMsg:    "command must have either Name or Namespace+Method",
		},
		{
			name: "invalid command - no function",
			command: Command{
				Name: "test",
			},
			expectError: true,
			errorMsg:    "has no executable function",
		},
		{
			name: "valid command - both func and args func",
			command: Command{
				Name:         "test",
				Func:         func() error { return nil },
				FuncWithArgs: func(args ...string) error { return nil },
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.command.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Validate() error = %q, expected to contain %q", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCommand_Execute(t *testing.T) {
	t.Run("execute without args", func(t *testing.T) {
		executed := false
		cmd := Command{
			Name: "test",
			Func: func() error {
				executed = true
				return nil
			},
		}

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if !executed {
			t.Error("Execute() did not call the function")
		}
	})

	t.Run("execute with args - args function available", func(t *testing.T) {
		var receivedArgs []string
		cmd := Command{
			Name: "test",
			FuncWithArgs: func(args ...string) error {
				receivedArgs = args
				return nil
			},
		}

		args := []string{"arg1", "arg2"}
		err := cmd.Execute(args...)
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if len(receivedArgs) != 2 || receivedArgs[0] != "arg1" || receivedArgs[1] != "arg2" {
			t.Errorf("Execute() args = %v, expected %v", receivedArgs, args)
		}
	})

	t.Run("execute with args - fallback to regular function", func(t *testing.T) {
		executed := false
		cmd := Command{
			Name: "test",
			Func: func() error {
				executed = true
				return nil
			},
		}

		err := cmd.Execute("arg1", "arg2")
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if !executed {
			t.Error("Execute() did not call the function")
		}
	})

	t.Run("execute with function error", func(t *testing.T) {
		expectedError := errTest
		cmd := Command{
			Name: "test",
			Func: func() error {
				return expectedError
			},
		}

		err := cmd.Execute()
		if !errors.Is(err, expectedError) {
			t.Errorf("Execute() error = %v, expected %v", err, expectedError)
		}
	})

	t.Run("execute with args function error", func(t *testing.T) {
		expectedError := errTestArgs
		cmd := Command{
			Name: "test",
			FuncWithArgs: func(args ...string) error {
				return expectedError
			},
		}

		err := cmd.Execute("arg")
		if !errors.Is(err, expectedError) {
			t.Errorf("Execute() error = %v, expected %v", err, expectedError)
		}
	})

	t.Run("execute with no function", func(t *testing.T) {
		cmd := Command{
			Name: "test",
		}

		err := cmd.Execute()
		if err == nil {
			t.Error("Execute() expected error for command with no function")
		}
		if !strings.Contains(err.Error(), "no executable function") {
			t.Errorf("Execute() error = %q, expected to contain 'no executable function'", err.Error())
		}
	})

	t.Run("execute deprecated command", func(t *testing.T) {
		cmd := Command{
			Name:       "test",
			Deprecated: "Use 'newtest' instead",
			Func:       func() error { return nil },
		}

		// We can't easily capture stdout in a unit test, so we just verify
		// that the function still executes correctly
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
	})
}

func TestCommandBuilder_NewCommand(t *testing.T) {
	builder := NewCommand("test")
	if builder == nil {
		t.Fatal("NewCommand() returned nil")
	}
	if builder.cmd.Name != "test" {
		t.Errorf("NewCommand() name = %q, expected %q", builder.cmd.Name, "test")
	}
}

func TestCommandBuilder_NewNamespaceCommand(t *testing.T) {
	builder := NewNamespaceCommand("build", "linux")
	if builder == nil {
		t.Fatal("NewNamespaceCommand() returned nil")
	}
	if builder.cmd.Namespace != "build" {
		t.Errorf("NewNamespaceCommand() namespace = %q, expected %q", builder.cmd.Namespace, "build")
	}
	if builder.cmd.Method != "linux" {
		t.Errorf("NewNamespaceCommand() method = %q, expected %q", builder.cmd.Method, "linux")
	}
}

func TestCommandBuilder_FluentInterface(t *testing.T) {
	testFunc := func() error { return nil }
	testArgsFunc := func(args ...string) error { return nil }

	cmd := NewCommand("test").
		WithDescription("Test command").
		WithFunc(testFunc).
		WithArgsFunc(testArgsFunc).
		WithAliases("t", "test-cmd").
		WithCategory("Test").
		WithDependencies("dep1", "dep2").
		Hidden().
		Deprecated("Use newtest instead").
		Since("v1.0.0").
		MustBuild()

	if cmd.Name != "test" {
		t.Errorf("Name = %q, expected %q", cmd.Name, "test")
	}
	if cmd.Description != "Test command" {
		t.Errorf("Description = %q, expected %q", cmd.Description, "Test command")
	}
	if cmd.Func == nil {
		t.Error("Func not set")
	}
	if cmd.FuncWithArgs == nil {
		t.Error("FuncWithArgs not set")
	}
	if len(cmd.Aliases) != 2 || cmd.Aliases[0] != "t" || cmd.Aliases[1] != "test-cmd" {
		t.Errorf("Aliases = %v, expected %v", cmd.Aliases, []string{"t", "test-cmd"})
	}
	if cmd.Category != testCategory {
		t.Errorf("Category = %q, expected %q", cmd.Category, testCategory)
	}
	if len(cmd.Dependencies) != 2 || cmd.Dependencies[0] != "dep1" || cmd.Dependencies[1] != "dep2" {
		t.Errorf("Dependencies = %v, expected %v", cmd.Dependencies, []string{"dep1", "dep2"})
	}
	if !cmd.Hidden {
		t.Error("Expected command to be hidden")
	}
	if cmd.Deprecated != "Use newtest instead" {
		t.Errorf("Deprecated = %q, expected %q", cmd.Deprecated, "Use newtest instead")
	}
	if cmd.Since != "v1.0.0" {
		t.Errorf("Since = %q, expected %q", cmd.Since, "v1.0.0")
	}
}

func TestCommandBuilder_Build(t *testing.T) {
	t.Run("valid command", func(t *testing.T) {
		cmd, err := NewCommand("test").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			t.Errorf("Build() unexpected error: %v", err)
		}
		if cmd == nil {
			t.Error("Build() returned nil command")
		}
	})

	t.Run("invalid command", func(t *testing.T) {
		_, err := NewCommand("").Build()

		if err == nil {
			t.Error("Build() expected error for invalid command")
		}
	})
}

func TestCommandBuilder_MustBuild(t *testing.T) {
	t.Run("valid command", func(t *testing.T) {
		cmd := NewCommand("test").
			WithFunc(func() error { return nil }).
			MustBuild()

		if cmd == nil {
			t.Error("MustBuild() returned nil command")
		}
	})

	t.Run("invalid command panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustBuild() expected panic for invalid command")
			}
		}()

		NewCommand("").MustBuild()
	})
}

func TestCommandMetadata(t *testing.T) {
	metadata := CommandMetadata{
		TotalCommands: 100,
		Namespaces:    []string{"build", "test", "lint"},
		Categories: map[string]int{
			"Build": 20,
			"Test":  30,
			"Lint":  10,
		},
		Version: "v1.0.0",
	}

	if metadata.TotalCommands != 100 {
		t.Errorf("TotalCommands = %d, expected %d", metadata.TotalCommands, 100)
	}
	if len(metadata.Namespaces) != 3 {
		t.Errorf("Namespaces length = %d, expected %d", len(metadata.Namespaces), 3)
	}
	if metadata.Categories["Build"] != 20 {
		t.Errorf("Build category count = %d, expected %d", metadata.Categories["Build"], 20)
	}
	if metadata.Version != "v1.0.0" {
		t.Errorf("Version = %q, expected %q", metadata.Version, "v1.0.0")
	}
}

func BenchmarkCommand_Execute(b *testing.B) {
	cmd := Command{
		Name: "bench",
		Func: func() error { return nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cmd.Execute(); err != nil {
			// Benchmark execution error is expected, continue
			_ = err
		}
	}
}

func BenchmarkCommand_ExecuteWithArgs(b *testing.B) {
	cmd := Command{
		Name:         "bench",
		FuncWithArgs: func(args ...string) error { return nil },
	}

	args := []string{"arg1", "arg2", "arg3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cmd.Execute(args...); err != nil {
			// Benchmark execution error is expected, continue
			_ = err
		}
	}
}

func BenchmarkCommand_FullName(b *testing.B) {
	cmd := Command{
		Namespace: "build",
		Method:    "linux",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.FullName()
	}
}

func BenchmarkCommandBuilder_Build(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := NewCommand("bench").
			WithDescription("Benchmark command").
			WithFunc(func() error { return nil }).
			WithCategory("Benchmark").
			Build(); err != nil {
			// Benchmark build error is unexpected but continue
			_ = err
		}
	}
}

func TestCommand_ExecutePrecedence(t *testing.T) {
	// Test that FuncWithArgs takes precedence when args are provided
	argsExecuted := false
	funcExecuted := false

	cmd := Command{
		Name: "test",
		Func: func() error {
			funcExecuted = true
			return nil
		},
		FuncWithArgs: func(args ...string) error {
			argsExecuted = true
			return nil
		},
	}

	// Execute with args - should call FuncWithArgs
	err := cmd.Execute("arg1")
	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}
	if !argsExecuted {
		t.Error("FuncWithArgs not called when args provided")
	}
	if funcExecuted {
		t.Error("Func called when FuncWithArgs should take precedence")
	}

	// Reset flags
	argsExecuted = false
	funcExecuted = false

	// Execute without args - should call Func
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Execute() unexpected error: %v", err)
	}
	if !funcExecuted {
		t.Error("Func not called when no args provided")
	}
	if argsExecuted {
		t.Error("FuncWithArgs called when no args provided")
	}
}

func TestCommand_ExecuteEdgeCases(t *testing.T) {
	t.Run("execute with empty args slice", func(t *testing.T) {
		funcExecuted := false
		cmd := Command{
			Name: "test",
			Func: func() error {
				funcExecuted = true
				return nil
			},
			FuncWithArgs: func(args ...string) error {
				t.Error("FuncWithArgs should not be called with empty args slice")
				return nil
			},
		}

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if !funcExecuted {
			t.Error("Func not called with empty args")
		}
	})

	t.Run("execute with nil args", func(t *testing.T) {
		funcExecuted := false
		cmd := Command{
			Name: "test",
			Func: func() error {
				funcExecuted = true
				return nil
			},
		}

		err := cmd.Execute(nil...)
		if err != nil {
			t.Errorf("Execute() unexpected error: %v", err)
		}
		if !funcExecuted {
			t.Error("Func not called with nil args")
		}
	})
}
