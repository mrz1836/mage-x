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

func TestRootCommandShortDescription(t *testing.T) {
	// Test the short description
	rootCmd := newRootCmd()
	expected := "A test module for multi-module support"
	if rootCmd.Short != expected {
		t.Errorf("Expected Short to be '%s', got '%s'", expected, rootCmd.Short)
	}
}

func TestRootCommandLongDescription(t *testing.T) {
	// Test the long description
	rootCmd := newRootCmd()
	expected := "This is a test module to demonstrate multi-module support in mage-x."
	if rootCmd.Long != expected {
		t.Errorf("Expected Long to be '%s', got '%s'", expected, rootCmd.Long)
	}
}
