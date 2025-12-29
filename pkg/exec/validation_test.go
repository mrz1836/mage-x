package exec

import (
	"context"
	"errors"
	"testing"
)

func TestValidateCommandArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		arg     string
		wantErr error
	}{
		// Valid arguments
		{name: "simple arg", arg: "hello", wantErr: nil},
		{name: "path arg", arg: "/path/to/file", wantErr: nil},
		{name: "flag arg", arg: "--verbose", wantErr: nil},
		{name: "url", arg: "https://example.com/path?query=1|2", wantErr: nil},
		{name: "regex with pipe", arg: "^test|foo$", wantErr: nil},

		// Invalid arguments
		{name: "command substitution $()", arg: "$(whoami)", wantErr: ErrDangerousPattern},
		{name: "command substitution backtick", arg: "`whoami`", wantErr: ErrDangerousPattern},
		{name: "command chaining &&", arg: "foo && bar", wantErr: ErrDangerousPattern},
		{name: "command chaining ||", arg: "foo || bar", wantErr: ErrDangerousPattern},
		{name: "command separator", arg: "foo; bar", wantErr: ErrDangerousPattern},
		{name: "redirect >", arg: "foo > bar", wantErr: ErrDangerousPattern},
		{name: "redirect <", arg: "foo < bar", wantErr: ErrDangerousPattern},
		{name: "pipe without regex", arg: "test | cat", wantErr: ErrDangerousPipePattern},
		{name: "pipe with cat bypass", arg: "test | cat /etc/passwd ^", wantErr: ErrDangerousPipePattern},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateCommandArg(tt.arg)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateCommandArg(%q) = %v, want nil", tt.arg, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateCommandArg(%q) = nil, want error containing %v", tt.arg, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateCommandArg(%q) = %v, want error containing %v", tt.arg, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		// Valid paths
		{name: "simple path", path: "file.txt", wantErr: nil},
		{name: "relative path", path: "dir/file.txt", wantErr: nil},
		{name: "tmp path", path: "/tmp/file.txt", wantErr: nil},

		// Invalid paths
		{name: "null byte", path: "file\x00.txt", wantErr: ErrPathContainsNull},
		{name: "newline", path: "file\n.txt", wantErr: ErrPathContainsControl},
		{name: "carriage return", path: "file\r.txt", wantErr: ErrPathContainsControl},
		{name: "path traversal", path: "../etc/passwd", wantErr: ErrPathTraversal},
		{name: "path traversal middle", path: "foo/../bar", wantErr: ErrPathTraversal},
		{name: "absolute path", path: "/etc/passwd", wantErr: ErrAbsolutePathNotAllowed},
		{name: "windows drive", path: "C:\\Windows", wantErr: ErrAbsolutePathNotAllowed},
		{name: "unc path", path: "\\\\server\\share", wantErr: ErrAbsolutePathNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidatePath(tt.path)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidatePath(%q) = %v, want nil", tt.path, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidatePath(%q) = nil, want error containing %v", tt.path, tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidatePath(%q) = %v, want error containing %v", tt.path, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidatingExecutor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name        string
		allowedCmds []string
		cmdName     string
		args        []string
		wantErr     bool
	}{
		{
			name:        "allowed command",
			allowedCmds: []string{"echo", "ls"},
			cmdName:     "echo",
			args:        []string{"hello"},
			wantErr:     false,
		},
		{
			name:        "disallowed command",
			allowedCmds: []string{"echo"},
			cmdName:     "ls",
			args:        []string{},
			wantErr:     true,
		},
		{
			name:        "empty whitelist allows all",
			allowedCmds: []string{},
			cmdName:     "ls",
			args:        []string{},
			wantErr:     false,
		},
		{
			name:        "dangerous argument",
			allowedCmds: []string{},
			cmdName:     "echo",
			args:        []string{"$(whoami)"},
			wantErr:     true,
		},
		{
			name:        "path traversal in command",
			allowedCmds: []string{},
			cmdName:     "../bin/sh",
			args:        []string{},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			base := NewBase()
			var opts []ValidatingOption
			if len(tt.allowedCmds) > 0 {
				opts = append(opts, WithAllowedCommands(tt.allowedCmds))
			}
			v := NewValidatingExecutor(base, opts...)

			err := v.Execute(ctx, tt.cmdName, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatingExecutor_ExecuteOutput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	base := NewBase()
	v := NewValidatingExecutor(base)

	// Valid command
	output, err := v.ExecuteOutput(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("ExecuteOutput() error = %v, want nil", err)
	}
	if output == "" {
		t.Error("ExecuteOutput() output is empty")
	}

	// Invalid command (dangerous pattern)
	_, err = v.ExecuteOutput(ctx, "echo", "$(whoami)")
	if err == nil {
		t.Error("ExecuteOutput() expected error for dangerous pattern")
	}
}

func TestContainsSuspiciousPipeCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{name: "no pipe", arg: "hello world", want: false},
		{name: "pipe with cat", arg: "test | cat", want: true},
		{name: "pipe with rm", arg: "test | rm -rf", want: true},
		{name: "pipe with bash", arg: "test | bash", want: true},
		{name: "pipe with safe text", arg: "test | something", want: false},
		{name: "regex or pattern", arg: "foo|bar", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsSuspiciousPipeCommand(tt.arg)
			if got != tt.want {
				t.Errorf("containsSuspiciousPipeCommand(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}
