package testhelpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TempWorkspace provides a temporary workspace for tests
type TempWorkspace struct {
	t       *testing.T
	rootDir string
	dirs    map[string]string
	cleanup []func()
}

// NewTempWorkspace creates a new temporary workspace
func NewTempWorkspace(t *testing.T, name string) *TempWorkspace {
	t.Helper()

	if name == "" {
		name = "workspace"
	}

	rootDir, err := os.MkdirTemp("", fmt.Sprintf("mage-%s-*", name))
	if err != nil {
		t.Fatalf("Failed to create temp workspace: %v", err)
	}

	tw := &TempWorkspace{
		t:       t,
		rootDir: rootDir,
		dirs:    make(map[string]string),
		cleanup: []func(){},
	}

	// Register cleanup
	t.Cleanup(tw.Cleanup)

	return tw
}

// Root returns the root directory of the workspace
func (tw *TempWorkspace) Root() string {
	return tw.rootDir
}

// Dir creates or gets a named directory in the workspace
func (tw *TempWorkspace) Dir(name string) string {
	tw.t.Helper()

	if dir, exists := tw.dirs[name]; exists {
		return dir
	}

	dir := filepath.Join(tw.rootDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		tw.t.Fatalf("Failed to create directory %s: %v", name, err)
	}

	tw.dirs[name] = dir
	return dir
}

// Path returns the full path for a relative path in the workspace
func (tw *TempWorkspace) Path(parts ...string) string {
	allParts := append([]string{tw.rootDir}, parts...)
	return filepath.Join(allParts...)
}

// WriteFile writes a file to the workspace
func (tw *TempWorkspace) WriteFile(path string, content []byte) string {
	tw.t.Helper()

	fullPath := tw.Path(path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		tw.t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		tw.t.Fatalf("Failed to write file: %v", err)
	}

	return fullPath
}

// WriteTextFile writes a text file to the workspace
func (tw *TempWorkspace) WriteTextFile(path, content string) string {
	return tw.WriteFile(path, []byte(content))
}

// ReadFile reads a file from the workspace
func (tw *TempWorkspace) ReadFile(path string) []byte {
	tw.t.Helper()

	fullPath := tw.Path(path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		tw.t.Fatalf("Failed to read file: %v", err)
	}

	return data
}

// ReadTextFile reads a text file from the workspace
func (tw *TempWorkspace) ReadTextFile(path string) string {
	return string(tw.ReadFile(path))
}

// CopyFile copies a file within the workspace
func (tw *TempWorkspace) CopyFile(src, dst string) {
	tw.t.Helper()

	srcPath := tw.Path(src)
	dstPath := tw.Path(dst)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		tw.t.Fatalf("Failed to create destination directory: %v", err)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		tw.t.Fatalf("Failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		tw.t.Fatalf("Failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		tw.t.Fatalf("Failed to copy file: %v", err)
	}
}

// CopyDir copies a directory within the workspace
func (tw *TempWorkspace) CopyDir(src, dst string) {
	tw.t.Helper()

	srcPath := tw.Path(src)
	dstPath := tw.Path(dst)

	err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		// Calculate destination path
		destPath := filepath.Join(dstPath, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
	if err != nil {
		tw.t.Fatalf("Failed to copy directory: %v", err)
	}
}

// Move moves a file or directory within the workspace
func (tw *TempWorkspace) Move(src, dst string) {
	tw.t.Helper()

	srcPath := tw.Path(src)
	dstPath := tw.Path(dst)

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		tw.t.Fatalf("Failed to create destination directory: %v", err)
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		tw.t.Fatalf("Failed to move: %v", err)
	}
}

// Remove removes a file or directory from the workspace
func (tw *TempWorkspace) Remove(path string) {
	tw.t.Helper()

	fullPath := tw.Path(path)
	if err := os.RemoveAll(fullPath); err != nil {
		tw.t.Fatalf("Failed to remove: %v", err)
	}
}

// Exists checks if a path exists in the workspace
func (tw *TempWorkspace) Exists(path string) bool {
	fullPath := tw.Path(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// IsDir checks if a path is a directory
func (tw *TempWorkspace) IsDir(path string) bool {
	fullPath := tw.Path(path)
	info, err := os.Stat(fullPath)
	return err == nil && info.IsDir()
}

// IsFile checks if a path is a file
func (tw *TempWorkspace) IsFile(path string) bool {
	fullPath := tw.Path(path)
	info, err := os.Stat(fullPath)
	return err == nil && !info.IsDir()
}

// Chmod changes file permissions
func (tw *TempWorkspace) Chmod(path string, mode os.FileMode) {
	tw.t.Helper()

	fullPath := tw.Path(path)
	if err := os.Chmod(fullPath, mode); err != nil {
		tw.t.Fatalf("Failed to chmod: %v", err)
	}
}

// Symlink creates a symbolic link
func (tw *TempWorkspace) Symlink(target, link string) {
	tw.t.Helper()

	targetPath := tw.Path(target)
	linkPath := tw.Path(link)

	// Create link directory if needed
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		tw.t.Fatalf("Failed to create link directory: %v", err)
	}

	if err := os.Symlink(targetPath, linkPath); err != nil {
		tw.t.Fatalf("Failed to create symlink: %v", err)
	}
}

// ListFiles lists all files in a directory
func (tw *TempWorkspace) ListFiles(dir string) []string {
	tw.t.Helper()

	fullPath := tw.Path(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		tw.t.Fatalf("Failed to list files: %v", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files
}

// ListDirs lists all directories in a directory
func (tw *TempWorkspace) ListDirs(dir string) []string {
	tw.t.Helper()

	fullPath := tw.Path(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		tw.t.Fatalf("Failed to list directories: %v", err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs
}

// Walk walks the workspace tree
func (tw *TempWorkspace) Walk(fn filepath.WalkFunc) {
	tw.t.Helper()

	if err := filepath.Walk(tw.rootDir, fn); err != nil {
		tw.t.Fatalf("Failed to walk workspace: %v", err)
	}
}

// AddCleanup adds a cleanup function
func (tw *TempWorkspace) AddCleanup(fn func()) {
	tw.cleanup = append(tw.cleanup, fn)
}

// Cleanup cleans up the workspace
func (tw *TempWorkspace) Cleanup() {
	// Run cleanup functions in reverse order
	for i := len(tw.cleanup) - 1; i >= 0; i-- {
		tw.cleanup[i]()
	}

	// Remove workspace
	os.RemoveAll(tw.rootDir)
}

// AssertExists asserts that a path exists
func (tw *TempWorkspace) AssertExists(path string) {
	tw.t.Helper()

	if !tw.Exists(path) {
		tw.t.Errorf("Expected %s to exist", path)
	}
}

// AssertNotExists asserts that a path does not exist
func (tw *TempWorkspace) AssertNotExists(path string) {
	tw.t.Helper()

	if tw.Exists(path) {
		tw.t.Errorf("Expected %s to not exist", path)
	}
}

// AssertFileContains asserts that a file contains text
func (tw *TempWorkspace) AssertFileContains(path, expected string) {
	tw.t.Helper()

	content := tw.ReadTextFile(path)
	if !strings.Contains(content, expected) {
		tw.t.Errorf("File %s does not contain expected text: %s", path, expected)
	}
}

// AssertFileEquals asserts that a file equals expected content
func (tw *TempWorkspace) AssertFileEquals(path, expected string) {
	tw.t.Helper()

	content := tw.ReadTextFile(path)
	if content != expected {
		tw.t.Errorf("File %s content mismatch.\nExpected:\n%s\nActual:\n%s",
			path, expected, content)
	}
}

// WorkspaceBuilder provides a fluent interface for building workspaces
type WorkspaceBuilder struct {
	workspace *TempWorkspace
}

// NewWorkspaceBuilder creates a new workspace builder
func NewWorkspaceBuilder(t *testing.T) *WorkspaceBuilder {
	return &WorkspaceBuilder{
		workspace: NewTempWorkspace(t, "builder"),
	}
}

// WithFile adds a file to the workspace
func (wb *WorkspaceBuilder) WithFile(path, content string) *WorkspaceBuilder {
	wb.workspace.WriteTextFile(path, content)
	return wb
}

// WithDir adds a directory to the workspace
func (wb *WorkspaceBuilder) WithDir(path string) *WorkspaceBuilder {
	wb.workspace.Dir(path)
	return wb
}

// WithGoModule adds a go.mod file
func (wb *WorkspaceBuilder) WithGoModule(module string) *WorkspaceBuilder {
	content := fmt.Sprintf("module %s\n\ngo 1.24\n", module)
	return wb.WithFile("go.mod", content)
}

// WithMagefile adds a magefile
func (wb *WorkspaceBuilder) WithMagefile() *WorkspaceBuilder {
	content := `// +build mage

package main

import "fmt"

func Build() error {
	fmt.Println("Building...")
	return nil
}
`
	return wb.WithFile("magefile.go", content)
}

// Build returns the workspace
func (wb *WorkspaceBuilder) Build() *TempWorkspace {
	return wb.workspace
}

// SandboxedWorkspace provides an isolated workspace with restricted access
type SandboxedWorkspace struct {
	*TempWorkspace
	allowedPaths map[string]bool
}

// NewSandboxedWorkspace creates a new sandboxed workspace
func NewSandboxedWorkspace(t *testing.T) *SandboxedWorkspace {
	return &SandboxedWorkspace{
		TempWorkspace: NewTempWorkspace(t, "sandbox"),
		allowedPaths:  make(map[string]bool),
	}
}

// AllowPath allows access to a path
func (sw *SandboxedWorkspace) AllowPath(path string) {
	sw.allowedPaths[path] = true
}

// isAllowed checks if a path is allowed
func (sw *SandboxedWorkspace) isAllowed(path string) bool {
	// Always allow paths within the workspace
	if filepath.HasPrefix(path, sw.rootDir) {
		return true
	}

	// Check allowed paths
	for allowed := range sw.allowedPaths {
		if filepath.HasPrefix(path, allowed) {
			return true
		}
	}

	return false
}

// WriteFile overrides to check permissions
func (sw *SandboxedWorkspace) WriteFile(path string, content []byte) string {
	sw.t.Helper()

	fullPath := sw.Path(path)
	if !sw.isAllowed(fullPath) {
		sw.t.Fatalf("Access denied to path: %s", fullPath)
	}

	return sw.TempWorkspace.WriteFile(path, content)
}

// ReadFile overrides to check permissions
func (sw *SandboxedWorkspace) ReadFile(path string) []byte {
	sw.t.Helper()

	fullPath := sw.Path(path)
	if !sw.isAllowed(fullPath) {
		sw.t.Fatalf("Access denied to path: %s", fullPath)
	}

	return sw.TempWorkspace.ReadFile(path)
}
