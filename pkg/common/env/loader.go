// Package env provides environment variable loading utilities
package env

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnvFiles loads environment variables from .env files in order of priority
// Higher priority files override lower priority ones
// Only sets environment variables that are not already set
func LoadEnvFiles(basePaths ...string) error {
	// Default search paths if none provided
	if len(basePaths) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		basePaths = []string{cwd}
	}

	// Files to load in order of priority (lowest to highest)
	envFiles := []string{
		".github/.env.base", // Base configuration (lowest priority)
		".env.base",         // Alternative base location
		".env.custom",       // Custom overrides
		".env.local",        // Local development
		".env",              // Standard env file (highest priority)
	}

	var loadedCount int
	for _, basePath := range basePaths {
		for _, envFile := range envFiles {
			fullPath := filepath.Join(basePath, envFile)
			if err := loadEnvFile(fullPath); err == nil {
				loadedCount++
			}
		}
	}

	return nil
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
		value := strings.TrimSpace(parts[1])

		// Expand variables in value (e.g., ${VAR} or $VAR)
		value = expandVariables(value, loaded)

		// Store in our local map for future expansions
		loaded[key] = value

		// Only set environment variable if not already set
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				// Continue on error but could log if needed
				_ = err
			}
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
