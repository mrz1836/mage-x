package exec

import (
	"errors"
	"strings"
	"testing"
)

// Static test errors for err113 compliance
var (
	errExitStatus1      = errors.New("exit status 1")
	errNotFound         = errors.New("not found")
	errTestsFailed      = errors.New("tests failed")
	errGeneric          = errors.New("error")
	errPermissionDenied = errors.New("permission denied")
	errBuildFailed      = errors.New("build failed")
)

func TestCommandError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmdName string
		args    []string
		err     error
		wantMsg string
	}{
		{
			name:    "simple command with args",
			cmdName: "echo",
			args:    []string{"hello"},
			err:     errExitStatus1,
			wantMsg: "command failed [echo hello]: exit status 1",
		},
		{
			name:    "command with no args",
			cmdName: "ls",
			args:    []string{},
			err:     errNotFound,
			wantMsg: "command failed [ls ]: not found",
		},
		{
			name:    "command with multiple args",
			cmdName: "go",
			args:    []string{"test", "-v", "./..."},
			err:     errTestsFailed,
			wantMsg: "command failed [go test -v ./...]: tests failed",
		},
		{
			name:    "nil args treated as empty",
			cmdName: "pwd",
			args:    nil,
			err:     errGeneric,
			wantMsg: "command failed [pwd ]: error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CommandError(tt.cmdName, tt.args, tt.err)
			if got.Error() != tt.wantMsg {
				t.Errorf("CommandError() = %q, want %q", got.Error(), tt.wantMsg)
			}
			// Verify error wrapping works
			if !errors.Is(got, tt.err) {
				t.Error("CommandError() should wrap the original error")
			}
		})
	}
}

func TestCommandErrorWithOutput(t *testing.T) {
	t.Parallel()

	baseErr := errExitStatus1

	tests := []struct {
		name         string
		cmdName      string
		args         []string
		err          error
		output       string
		wantContains string
		wantNewline  bool
	}{
		{
			name:         "with output includes newline and output",
			cmdName:      "go",
			args:         []string{"build"},
			err:          baseErr,
			output:       "undefined: foo",
			wantContains: "undefined: foo",
			wantNewline:  true,
		},
		{
			name:         "empty output no newline",
			cmdName:      "echo",
			args:         []string{"test"},
			err:          baseErr,
			output:       "",
			wantContains: "command failed [echo test]: exit status 1",
			wantNewline:  false,
		},
		{
			name:         "whitespace-only output treated as empty",
			cmdName:      "ls",
			args:         []string{"-la"},
			err:          baseErr,
			output:       "   \n\t  ",
			wantContains: "command failed [ls -la]: exit status 1",
			wantNewline:  false,
		},
		{
			name:         "output with surrounding whitespace is trimmed",
			cmdName:      "cat",
			args:         []string{"file.txt"},
			err:          baseErr,
			output:       "  error message  \n",
			wantContains: "error message",
			wantNewline:  true,
		},
		{
			name:         "multiline output preserved",
			cmdName:      "make",
			args:         []string{"build"},
			err:          baseErr,
			output:       "line1\nline2\nline3",
			wantContains: "line1\nline2\nline3",
			wantNewline:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CommandErrorWithOutput(tt.cmdName, tt.args, tt.err, tt.output)
			msg := got.Error()

			if !strings.Contains(msg, tt.wantContains) {
				t.Errorf("CommandErrorWithOutput() = %q, want to contain %q", msg, tt.wantContains)
			}

			hasNewline := strings.Contains(msg, "\n")
			if hasNewline != tt.wantNewline {
				t.Errorf("hasNewline = %v, want %v", hasNewline, tt.wantNewline)
			}

			// Verify error wrapping
			if !errors.Is(got, tt.err) {
				t.Error("CommandErrorWithOutput() should wrap the original error")
			}
		})
	}
}

func TestCommandErrorInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmdName string
		args    []string
		dir     string
		err     error
		wantMsg string
	}{
		{
			name:    "includes directory in message",
			cmdName: "ls",
			args:    []string{"-la"},
			dir:     "/tmp/test",
			err:     errPermissionDenied,
			wantMsg: "command failed [ls -la] in /tmp/test: permission denied",
		},
		{
			name:    "relative directory",
			cmdName: "go",
			args:    []string{"build"},
			dir:     "./cmd/app",
			err:     errBuildFailed,
			wantMsg: "command failed [go build] in ./cmd/app: build failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CommandErrorInDir(tt.cmdName, tt.args, tt.dir, tt.err)
			if got.Error() != tt.wantMsg {
				t.Errorf("CommandErrorInDir() = %q, want %q", got.Error(), tt.wantMsg)
			}
			if !errors.Is(got, tt.err) {
				t.Error("CommandErrorInDir() should wrap the original error")
			}
		})
	}
}

func TestCommandErrorInDirWithOutput(t *testing.T) {
	t.Parallel()

	baseErr := errExitStatus1

	tests := []struct {
		name         string
		cmdName      string
		args         []string
		dir          string
		err          error
		output       string
		wantContains string
		wantNewline  bool
	}{
		{
			name:         "with output includes directory and output",
			cmdName:      "go",
			args:         []string{"test"},
			dir:          "/project/pkg",
			err:          baseErr,
			output:       "FAIL: TestFoo",
			wantContains: "in /project/pkg",
			wantNewline:  true,
		},
		{
			name:         "empty output includes directory only",
			cmdName:      "make",
			args:         []string{"all"},
			dir:          "/build",
			err:          baseErr,
			output:       "",
			wantContains: "in /build: exit status 1",
			wantNewline:  false,
		},
		{
			name:         "whitespace-only output treated as empty",
			cmdName:      "npm",
			args:         []string{"test"},
			dir:          "/app",
			err:          baseErr,
			output:       "   \t\n   ",
			wantContains: "in /app: exit status 1",
			wantNewline:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CommandErrorInDirWithOutput(tt.cmdName, tt.args, tt.dir, tt.err, tt.output)
			msg := got.Error()

			if !strings.Contains(msg, tt.wantContains) {
				t.Errorf("CommandErrorInDirWithOutput() = %q, want to contain %q", msg, tt.wantContains)
			}

			hasNewline := strings.Contains(msg, "\n")
			if hasNewline != tt.wantNewline {
				t.Errorf("hasNewline = %v, want %v", hasNewline, tt.wantNewline)
			}

			if !errors.Is(got, tt.err) {
				t.Error("CommandErrorInDirWithOutput() should wrap the original error")
			}
		})
	}
}
