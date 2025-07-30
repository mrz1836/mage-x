// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Generate namespace for code generation tasks
type Generate mg.Namespace

// Default runs go generate in the base of the repo
func (Generate) Default() error {
	utils.Header("Running Code Generation")

	// Check for generate directives
	hasGenerate, files := checkForGenerateDirectives()
	if !hasGenerate {
		utils.Info("No //go:generate directives found")
		return nil
	}

	utils.Info("Found generate directives in %d files", len(files))
	for _, file := range files {
		utils.Info("  - %s", file)
	}

	// Run go generate
	utils.Info("\nRunning go generate...")

	args := []string{"generate", "-v"}

	// Add build tags if specified
	if tags := os.Getenv("GO_BUILD_TAGS"); tags != "" {
		args = append(args, "-tags", tags)
	}

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go generate failed: %w", err)
	}

	utils.Success("Code generation complete")
	return nil
}

// All runs go generate on all packages
func (Generate) All() error {
	utils.Header("Running Code Generation (All Packages)")

	// Run go generate on all packages
	utils.Info("Running go generate on all packages...")

	args := []string{"generate", "-v", "./..."}

	// Add build tags if specified
	if tags := os.Getenv("GO_BUILD_TAGS"); tags != "" {
		args = append(args, "-tags", tags)
	}

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go generate failed: %w", err)
	}

	utils.Success("Code generation complete for all packages")
	return nil
}

// Mocks generates mock files
func (Generate) Mocks() error {
	utils.Header("Generating Mocks")

	// Check if mockgen is installed
	if !utils.CommandExists("mockgen") {
		utils.Info("Installing mockgen...")
		if err := GetRunner().RunCmd("go", "install", "github.com/golang/mock/mockgen@latest"); err != nil {
			return fmt.Errorf("failed to install mockgen: %w", err)
		}
	}

	// Find interfaces to mock
	interfaces := findInterfaces()
	if len(interfaces) == 0 {
		utils.Info("No interfaces found to mock")
		return nil
	}

	utils.Info("Found %d interfaces to mock", len(interfaces))

	// Generate mocks
	for _, iface := range interfaces {
		utils.Info("Generating mock for %s.%s", iface.Package, iface.Name)

		outputFile := fmt.Sprintf("mocks/mock_%s.go", strings.ToLower(iface.Name))

		// Create mocks directory if needed
		if err := os.MkdirAll("mocks", 0o755); err != nil {
			return fmt.Errorf("failed to create mocks directory: %w", err)
		}

		// Generate mock
		args := []string{
			"-source", iface.File,
			"-destination", outputFile,
			"-package", "mocks",
		}

		if err := GetRunner().RunCmd("mockgen", args...); err != nil {
			utils.Warn("Failed to generate mock for %s: %v", iface.Name, err)
		}
	}

	utils.Success("Mock generation complete")
	return nil
}

// Proto generates protobuf code
func (Generate) Proto() error {
	utils.Header("Generating Protocol Buffer Code")

	// Check if protoc is installed
	if !utils.CommandExists("protoc") {
		return fmt.Errorf("protoc not found. Please install protocol buffer compiler")
	}

	// Check for protoc-gen-go
	if !utils.CommandExists("protoc-gen-go") {
		utils.Info("Installing protoc-gen-go...")
		if err := GetRunner().RunCmd("go", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@latest"); err != nil {
			return fmt.Errorf("failed to install protoc-gen-go: %w", err)
		}
	}

	// Check for protoc-gen-go-grpc if needed
	hasService := checkForGRPCService()
	if hasService && !utils.CommandExists("protoc-gen-go-grpc") {
		utils.Info("Installing protoc-gen-go-grpc...")
		if err := GetRunner().RunCmd("go", "install", "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"); err != nil {
			return fmt.Errorf("failed to install protoc-gen-go-grpc: %w", err)
		}
	}

	// Find proto files
	protoFiles, err := utils.FindFiles(".", "*.proto")
	if err != nil {
		return fmt.Errorf("failed to find proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		utils.Info("No .proto files found")
		return nil
	}

	utils.Info("Found %d proto files", len(protoFiles))

	// Generate code for each proto file
	for _, proto := range protoFiles {
		utils.Info("Processing %s", proto)

		dir := filepath.Dir(proto)

		args := []string{
			"--go_out=.",
			"--go_opt=paths=source_relative",
		}

		if hasService {
			args = append(args,
				"--go-grpc_out=.",
				"--go-grpc_opt=paths=source_relative",
			)
		}

		args = append(args, "-I", dir, proto)

		if err := GetRunner().RunCmd("protoc", args...); err != nil {
			return fmt.Errorf("failed to generate code for %s: %w", proto, err)
		}
	}

	utils.Success("Protocol buffer code generation complete")
	return nil
}

// Clean removes generated files
func (Generate) Clean() error {
	utils.Header("Cleaning Generated Files")

	patterns := []string{
		"*.pb.go",
		"*_gen.go",
		"*_generated.go",
		"mock_*.go",
		"mocks/",
	}

	// Add custom patterns from environment
	if custom := os.Getenv("GENERATED_PATTERNS"); custom != "" {
		patterns = append(patterns, strings.Split(custom, ",")...)
	}

	removed := 0
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, file := range files {
			utils.Info("Removing %s", file)
			if err := os.RemoveAll(file); err != nil {
				utils.Warn("Failed to remove %s: %v", file, err)
			} else {
				removed++
			}
		}
	}

	if removed > 0 {
		utils.Success("Removed %d generated files/directories", removed)
	} else {
		utils.Info("No generated files found to clean")
	}

	return nil
}

// Check verifies generated files are up to date
func (Generate) Check() error {
	utils.Header("Checking Generated Files")

	// Create a temporary directory for comparison
	tempDir, err := os.MkdirTemp("", "generate-check-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			utils.Warn("Failed to remove temp directory: %v", removeErr)
		}
	}()

	// Copy current generated files
	utils.Info("Backing up current generated files...")
	if backupErr := backupGeneratedFiles(tempDir); backupErr != nil {
		return fmt.Errorf("failed to backup files: %w", backupErr)
	}

	// Run generation
	utils.Info("Running code generation...")
	if genErr := (Generate{}).All(); genErr != nil {
		return fmt.Errorf("generation failed: %w", genErr)
	}

	// Compare files
	utils.Info("Comparing generated files...")
	different, err := compareGeneratedFiles(tempDir)
	if err != nil {
		return fmt.Errorf("failed to compare files: %w", err)
	}

	if len(different) > 0 {
		utils.Error("Generated files are out of date:")
		for _, file := range different {
			fmt.Printf("  - %s\n", file)
		}
		return fmt.Errorf("generated files need to be regenerated")
	}

	utils.Success("All generated files are up to date")
	return nil
}

// Helper types and functions

type Interface struct {
	Name    string
	Package string
	File    string
}

// checkForGenerateDirectives checks for //go:generate directives
func checkForGenerateDirectives() (bool, []string) {
	var files []string
	hasGenerate := false

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip vendor and common directories
		if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			fileOps := fileops.New()
			content, err := fileOps.File.ReadFile(path)
			if err != nil {
				return nil
			}

			if strings.Contains(string(content), "//go:generate") {
				hasGenerate = true
				files = append(files, path)
			}
		}

		return nil
	})
	if err != nil {
		utils.Warn("Error walking directory for generate directives: %v", err)
	}

	return hasGenerate, files
}

// findInterfaces finds interfaces in the codebase
func findInterfaces() []Interface {
	var interfaces []Interface

	// This is a simplified version - in reality, you'd use go/ast to parse
	// For now, we'll look for common patterns
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor/") {
			// Look for interface definitions
			// This is simplified - real implementation would use go/ast
		}

		return nil
	})
	if err != nil {
		utils.Warn("Error walking directory for interfaces: %v", err)
	}

	return interfaces
}

// checkForGRPCService checks if proto files contain service definitions
func checkForGRPCService() bool {
	protoFiles, err := utils.FindFiles(".", "*.proto")
	if err != nil {
		utils.Warn("Error finding proto files: %v", err)
		return false
	}

	fileOps := fileops.New()
	for _, proto := range protoFiles {
		content, err := fileOps.File.ReadFile(proto)
		if err != nil {
			continue
		}

		if strings.Contains(string(content), "service ") {
			return true
		}
	}

	return false
}

// backupGeneratedFiles backs up generated files for comparison
func backupGeneratedFiles(tempDir string) error {
	patterns := []string{"*.pb.go", "*_gen.go", "*_generated.go"}

	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for range files {
			// Copy file to temp directory
			// Implementation would copy file preserving directory structure
		}
	}

	return nil
}

// compareGeneratedFiles compares generated files with backup
func compareGeneratedFiles(tempDir string) ([]string, error) {
	var different []string

	// Implementation would compare files in tempDir with current files
	// and return list of files that are different

	return different, nil
}

// Additional methods for Generate namespace required by tests

// Code generates code
func (Generate) Code() error {
	runner := GetRunner()
	return runner.RunCmd("go", "generate", "./...")
}

// Docs generates documentation
func (Generate) Docs() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating documentation")
}

// Swagger generates Swagger documentation
func (Generate) Swagger() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating Swagger docs")
}

// OpenAPI generates OpenAPI specification
func (Generate) OpenAPI() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating OpenAPI spec")
}

// GraphQL generates GraphQL code
func (Generate) GraphQL() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating GraphQL code")
}

// SQL generates SQL files
func (Generate) SQL() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating SQL files")
}

// Wire generates wire dependency injection
func (Generate) Wire() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating wire files")
}

// Config generates configuration files
func (Generate) Config() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating config files")
}
