package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestHelp_Default tests the default help output
func TestHelp_Default(t *testing.T) {
	h := Help{}
	err := h.Default()
	require.NoError(t, err, "Default help should not error")
}

// TestHelp_Commands tests listing all available commands
func TestHelp_Commands(t *testing.T) {
	h := Help{}
	err := h.Commands()
	require.NoError(t, err, "Commands listing should not error")
}

// TestHelp_Command tests getting help for specific commands
func TestHelp_Command(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "missing command name",
			commandName: "",
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "nonexistent command",
			commandName: "nonexistent-command-12345",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.commandName != "" {
				t.Setenv("COMMAND", tt.commandName)
			}

			h := Help{}
			err := h.Command()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestHelp_Examples tests the examples display
func TestHelp_Examples(t *testing.T) {
	h := Help{}
	err := h.Examples()
	require.NoError(t, err, "Examples display should not error")
}

// TestHelp_GettingStarted tests the getting started guide
func TestHelp_GettingStarted(t *testing.T) {
	h := Help{}
	err := h.GettingStarted()
	require.NoError(t, err, "Getting started guide should not error")
}

// TestHelp_Completions tests shell completions generation
func TestHelp_Completions(t *testing.T) {
	tests := []struct {
		name        string
		shell       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "bash completions",
			shell:   "bash",
			wantErr: false,
		},
		{
			name:    "zsh completions",
			shell:   "zsh",
			wantErr: false,
		},
		{
			name:    "fish completions",
			shell:   "fish",
			wantErr: false,
		},
		{
			name:        "unsupported shell",
			shell:       "invalid-shell",
			wantErr:     true,
			errContains: "unsupported",
		},
		{
			name:        "no shell specified",
			shell:       "",
			wantErr:     true,
			errContains: "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shell != "" {
				t.Setenv("SHELL", tt.shell)
			}

			h := Help{}
			err := h.Completions()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestHelp_Topics tests help topics display
func TestHelp_Topics(t *testing.T) {
	h := Help{}
	err := h.Topics()
	require.NoError(t, err, "Topics display should not error")
}

// TestHelpCommand_Structure tests HelpCommand structure
func TestHelpCommand_Structure(t *testing.T) {
	tests := []struct {
		name string
		cmd  HelpCommand
	}{
		{
			name: "complete command",
			cmd: HelpCommand{
				Name:        "test",
				Namespace:   "mage",
				Description: "Test command",
				Usage:       "magex test",
				Examples:    []string{"magex test", "magex test --verbose"},
				Options: []HelpOption{
					{
						Name:        "verbose",
						Description: "Enable verbose output",
						Default:     "false",
						Required:    false,
					},
				},
				SeeAlso: []string{"build", "lint"},
			},
		},
		{
			name: "minimal command",
			cmd: HelpCommand{
				Name:        "minimal",
				Description: "Minimal command",
				Usage:       "magex minimal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.cmd.Name)
			assert.NotEmpty(t, tt.cmd.Description)
			assert.NotEmpty(t, tt.cmd.Usage)
		})
	}
}

// TestHelpOption_Structure tests HelpOption structure
func TestHelpOption_Structure(t *testing.T) {
	tests := []struct {
		name   string
		option HelpOption
	}{
		{
			name: "required option",
			option: HelpOption{
				Name:        "output",
				Description: "Output file",
				Required:    true,
			},
		},
		{
			name: "optional option with default",
			option: HelpOption{
				Name:        "verbose",
				Description: "Verbose output",
				Default:     "false",
				Required:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.option.Name)
			assert.NotEmpty(t, tt.option.Description)
		})
	}
}

// TestFormatOption tests option formatting
func TestFormatOption(t *testing.T) {
	tests := []struct {
		name     string
		option   HelpOption
		contains []string
	}{
		{
			name: "required option",
			option: HelpOption{
				Name:        "output",
				Description: "Output file",
				Required:    true,
			},
			contains: []string{"Output file", "(required)"},
		},
		{
			name: "optional option with default",
			option: HelpOption{
				Name:        "verbose",
				Description: "Verbose output",
				Default:     "false",
				Required:    false,
			},
			contains: []string{"Verbose output", "[default: false]"},
		},
		{
			name: "simple option",
			option: HelpOption{
				Name:        "help",
				Description: "Show help",
				Required:    false,
			},
			contains: []string{"Show help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOption(tt.option)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// TestHelpErrorConstants tests error constants
func TestHelpErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "help command required error",
			err:  errHelpCommandRequired,
			want: "required",
		},
		{
			name: "unsupported shell error",
			err:  errUnsupportedShell,
			want: "unsupported",
		},
		{
			name: "command not found error",
			err:  errCommandNotFound,
			want: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.want)
		})
	}
}

// TestGetAllCommandsUnit tests getAllCommands helper function
func TestGetAllCommandsUnit(t *testing.T) {
	t.Run("returns help commands from registry", func(t *testing.T) {
		commands := getAllCommands()

		// Returns a slice (may be empty if no commands registered in unit test context)
		assert.NotNil(t, commands)

		// Each command should have required fields if any are present
		for _, cmd := range commands {
			assert.NotEmpty(t, cmd.Name, "command should have a name")
		}
	})

	t.Run("converts registry commands correctly", func(t *testing.T) {
		commands := getAllCommands()

		// Find a known command and verify conversion
		for _, cmd := range commands {
			if cmd.Namespace != "" {
				// Namespaced commands should have namespace populated
				assert.NotEmpty(t, cmd.Name)
				break
			}
		}
	})
}

// TestGetCommandHelpUnit tests getCommandHelp helper function
func TestGetCommandHelpUnit(t *testing.T) {
	t.Run("returns error for nonexistent command", func(t *testing.T) {
		_, err := getCommandHelp("nonexistent-command-xyz-12345")

		require.Error(t, err)
		assert.ErrorIs(t, err, errCommandNotFound)
	})

	t.Run("returns command help for valid command", func(t *testing.T) {
		// Get all commands first to find a valid one
		commands := getAllCommands()
		if len(commands) == 0 {
			t.Skip("no commands registered")
		}

		// Use the first available command
		validName := commands[0].Name

		cmd, err := getCommandHelp(validName)

		require.NoError(t, err)
		assert.Equal(t, validName, cmd.Name)
	})
}

// TestShowHelpNamespaceUnit tests showHelpNamespace helper function
func TestShowHelpNamespaceUnit(t *testing.T) {
	t.Run("handles namespace with no commands", func(t *testing.T) {
		// Get global registry
		reg := registry.Global()

		// Try a namespace that likely has no commands
		err := showHelpNamespace(reg, "nonexistent-namespace-xyz")

		// Should not error, just print message
		require.NoError(t, err)
	})

	t.Run("displays commands for valid namespace", func(t *testing.T) {
		reg := registry.Global()
		namespaces := reg.Namespaces()

		if len(namespaces) == 0 {
			t.Skip("no namespaces registered")
		}

		// Use the first available namespace
		err := showHelpNamespace(reg, namespaces[0])

		require.NoError(t, err)
	})
}

// TestHelpCommandNamespace tests Help.Command with namespace input
func TestHelpCommandNamespace(t *testing.T) {
	t.Run("shows namespace help when namespace provided", func(t *testing.T) {
		// Get a valid namespace
		reg := registry.Global()
		namespaces := reg.Namespaces()

		if len(namespaces) == 0 {
			t.Skip("no namespaces registered")
		}

		t.Setenv("COMMAND", namespaces[0])

		h := Help{}
		err := h.Command()

		require.NoError(t, err)
	})
}

// TestGenerateBashCompletions tests bash completion generation
func TestGenerateBashCompletions(t *testing.T) {
	t.Run("generates bash completion script", func(t *testing.T) {
		err := generateBashCompletions()
		require.NoError(t, err)
	})
}

// TestGenerateZshCompletions tests zsh completion generation
func TestGenerateZshCompletions(t *testing.T) {
	t.Run("generates zsh completion script", func(t *testing.T) {
		err := generateZshCompletions()
		require.NoError(t, err)
	})
}

// TestGenerateFishCompletions tests fish completion generation
func TestGenerateFishCompletions(t *testing.T) {
	t.Run("generates fish completion script", func(t *testing.T) {
		err := generateFishCompletions()
		require.NoError(t, err)
	})
}
