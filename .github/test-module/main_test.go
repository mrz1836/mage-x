package main

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	// Test that root command is properly initialized
	rootCmd := newRootCmd()
	if rootCmd == nil {
		t.Error("rootCmd should not be nil")
		return
	}

	if rootCmd.Use != "test-module" {
		t.Errorf("Expected Use to be 'test-module', got '%s'", rootCmd.Use)
	}
}
