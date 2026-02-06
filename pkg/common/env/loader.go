// Package env provides environment variable loading utilities
package env

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Sentinel errors for env file operations
var (
	ErrNotDirectory = errors.New("path is not a directory")
	ErrNoEnvFiles   = errors.New("no .env files found in directory")
)

// LoadEnvFiles loads environment variables from .env files in order of priority
// Higher priority files override lower priority ones
// For each basePath, it first tries modular .github/env/*.env files (preferred),
// then falls back to legacy flat-file loading.
func LoadEnvFiles(basePaths ...string) error {
	// Default search paths if none provided
	if len(basePaths) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		basePaths = []string{cwd}
	}

	// Legacy files to load in order of priority (lowest to highest)
	envFiles := []string{
		".github/.env.base",   // Base configuration (lowest priority)
		".env.base",           // Alternative base location
		".github/.env.custom", // Custom overrides in .github directory
		".env.custom",         // Custom overrides in root directory
		".env.local",          // Local development
		".env",                // Standard env file (highest priority)
	}

	var loadedCount int
	for _, basePath := range basePaths {
		// Try modular .github/env/ directory first
		if envDir := findEnvDir(basePath); envDir != "" {
			if err := LoadEnvDir(envDir, isCI()); err != nil {
				return err
			}
			loadedCount++
			continue
		}

		// Fall back to legacy flat-file loading
		for _, envFile := range envFiles {
			fullPath := filepath.Join(basePath, envFile)
			if err := loadEnvFile(fullPath); err == nil {
				loadedCount++
			}
		}
	}

	return nil
}

// LoadEnvDir loads all *.env files from the given directory in lexicographic order.
// When skipLocal is true, 99-local.env is skipped (intended for CI environments).
func LoadEnvDir(dirPath string, skipLocal bool) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("failed to access directory %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %w", dirPath, ErrNotDirectory)
	}

	pattern := filepath.Join(dirPath, "*.env")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob %s: %w", pattern, err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("%s: %w", dirPath, ErrNoEnvFiles)
	}

	sort.Strings(matches)

	for _, file := range matches {
		if skipLocal && filepath.Base(file) == "99-local.env" {
			continue
		}
		if err := loadEnvFile(file); err != nil {
			return fmt.Errorf("failed to load %s: %w", file, err)
		}
	}

	return nil
}

// isCI returns true when running in a CI environment.
func isCI() bool {
	return os.Getenv("CI") == "true"
}

// hasEnvFiles checks if a directory exists and contains at least one .env file.
func hasEnvFiles(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil || !info.IsDir() {
		return false
	}

	matches, err := filepath.Glob(filepath.Join(dirPath, "*.env"))
	if err != nil {
		return false
	}

	return len(matches) > 0
}

// findEnvDir checks for a .github/env/ directory under basePath that contains *.env files.
// Returns the path to the env directory, or empty string if not found.
func findEnvDir(basePath string) string {
	envDir := filepath.Join(basePath, ".github", "env")
	if hasEnvFiles(envDir) {
		return envDir
	}
	return ""
}

// processValue processes a raw value from an env file, handling quotes and inline comments.
func processValue(value string) string {
	value = strings.TrimSpace(value)

	if len(value) == 0 {
		return value
	}

	// Handle double-quoted values
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}

	// Handle single-quoted values (no variable expansion, no inline comment stripping)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		return value[1 : len(value)-1]
	}

	// For unquoted values, strip inline comments (# preceded by whitespace)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	if idx := strings.Index(value, "\t#"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}

	return value
}

// loadEnvFile loads a single env file, parsing variable expansions
func loadEnvFile(filePath string) error {
	// #nosec G304 - filePath is constructed from known safe directories and env file names
	file, err := os.Open(filePath)
	if err != nil {
		return err // File doesn't exist or can't be read
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			// Log close error but don't fail the operation
			_ = cerr
		}
	}()

	// Store loaded variables for expansion
	loaded := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])

		// Strip optional "export " prefix
		key = strings.TrimPrefix(key, "export ")
		key = strings.TrimSpace(key)

		// Skip empty keys after prefix stripping
		if key == "" {
			continue
		}

		// Process value: handle quotes and inline comments
		value := processValue(parts[1])

		// Expand variables in value (e.g., ${VAR} or $VAR)
		value = expandVariables(value, loaded)

		// Store in our local map for future expansions
		loaded[key] = value

		// Set environment variable (later files override earlier ones)
		if err := os.Setenv(key, value); err != nil {
			// Continue on error but could log if needed
			_ = err
		}
	}

	return scanner.Err()
}

// expandVariables expands ${VAR} and $VAR references in a value
func expandVariables(value string, localVars map[string]string) string {
	// Handle ${VAR} format
	for strings.Contains(value, "${") {
		start := strings.Index(value, "${")
		if start == -1 {
			break
		}
		end := strings.Index(value[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varName := value[start+2 : end]
		var replacement string

		// Check local vars first, then environment
		if localValue, exists := localVars[varName]; exists {
			replacement = localValue
		} else {
			replacement = os.Getenv(varName)
		}

		value = value[:start] + replacement + value[end+1:]
	}

	// Handle $VAR format (simpler case)
	if strings.HasPrefix(value, "$") && !strings.Contains(value, " ") {
		varName := value[1:]
		if localValue, exists := localVars[varName]; exists {
			return localValue
		}
		return os.Getenv(varName)
	}

	return value
}

// LoadStartupEnv is a convenience function for loading env files at application startup
// It looks for .env files in the current directory and parent directories
func LoadStartupEnv() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Try current directory and up to 3 parent directories
	searchPaths := []string{cwd}
	currentPath := cwd
	for i := 0; i < 3; i++ {
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break // Reached root
		}
		searchPaths = append(searchPaths, parentPath)
		currentPath = parentPath
	}

	return LoadEnvFiles(searchPaths...)
}
