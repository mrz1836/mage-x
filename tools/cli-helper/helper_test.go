package clihelper

import (
	"testing"
)

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		args     []string
		expected string
	}{
		{
			name:     "command only",
			cmd:      "go",
			args:     []string{},
			expected: "go",
		},
		{
			name:     "command with args",
			cmd:      "go",
			args:     []string{"test", "./..."},
			expected: "go test ./...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCommand(tt.cmd, tt.args)
			if result != tt.expected {
				t.Errorf("FormatCommand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "empty args",
			args:         []string{},
			expectedCmd:  "",
			expectedArgs: nil,
		},
		{
			name:         "single arg",
			args:         []string{"test"},
			expectedCmd:  "test",
			expectedArgs: []string{},
		},
		{
			name:         "multiple args",
			args:         []string{"test", "./...", "-v"},
			expectedCmd:  "test",
			expectedArgs: []string{"./...", "-v"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := ParseArgs(tt.args)
			if cmd != tt.expectedCmd {
				t.Errorf("ParseArgs() cmd = %v, want %v", cmd, tt.expectedCmd)
			}
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("ParseArgs() args length = %v, want %v", len(args), len(tt.expectedArgs))
			}
		})
	}
}
