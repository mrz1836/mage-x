package paths

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// DefaultPathBuilder implements PathBuilder using standard library
type DefaultPathBuilder struct {
	path    string
	options PathOptions
	fileOps *fileops.FileOps
}

// NewPathBuilder creates a new path builder
func NewPathBuilder(path string) *DefaultPathBuilder {
	return &DefaultPathBuilder{
		path: filepath.Clean(path),
		options: PathOptions{
			CreateMode:    0o755,
			CreateParents: true,
			BufferSize:    8192,
		},
		fileOps: fileops.New(),
	}
}

// NewPathBuilderWithOptions creates a new path builder with options
func NewPathBuilderWithOptions(path string, options PathOptions) *DefaultPathBuilder {
	return &DefaultPathBuilder{
		path:    filepath.Clean(path),
		options: options,
	}
}

// Basic operations

// Join appends path elements to the current path
func (pb *DefaultPathBuilder) Join(elements ...string) PathBuilder {
	allElements := append([]string{pb.path}, elements...)
	return &DefaultPathBuilder{
		path:    filepath.Join(allElements...),
		options: pb.options,
	}
}

// Dir returns the directory component of the path
func (pb *DefaultPathBuilder) Dir() PathBuilder {
	return &DefaultPathBuilder{
		path:    filepath.Dir(pb.path),
		options: pb.options,
	}
}

// Base returns the base name of the path
func (pb *DefaultPathBuilder) Base() string {
	return filepath.Base(pb.path)
}

// Ext returns the extension of the path
func (pb *DefaultPathBuilder) Ext() string {
	return filepath.Ext(pb.path)
}

// Clean returns a cleaned version of the path
func (pb *DefaultPathBuilder) Clean() PathBuilder {
	return &DefaultPathBuilder{
		path:    filepath.Clean(pb.path),
		options: pb.options,
	}
}

// Abs returns an absolute representation of the path
func (pb *DefaultPathBuilder) Abs() (PathBuilder, error) {
	abs, err := filepath.Abs(pb.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	return &DefaultPathBuilder{
		path:    abs,
		options: pb.options,
	}, nil
}

// Path operations

// Append adds a suffix to the path (before extension)
func (pb *DefaultPathBuilder) Append(suffix string) PathBuilder {
	ext := filepath.Ext(pb.path)
	base := strings.TrimSuffix(pb.path, ext)
	return &DefaultPathBuilder{
		path:    base + suffix + ext,
		options: pb.options,
	}
}

// Prepend adds a prefix to the filename
func (pb *DefaultPathBuilder) Prepend(prefix string) PathBuilder {
	dir := filepath.Dir(pb.path)
	base := filepath.Base(pb.path)
	return &DefaultPathBuilder{
		path:    filepath.Join(dir, prefix+base),
		options: pb.options,
	}
}

// WithExt changes the extension
func (pb *DefaultPathBuilder) WithExt(ext string) PathBuilder {
	if !strings.HasPrefix(ext, ".") && ext != "" {
		ext = "." + ext
	}
	base := strings.TrimSuffix(pb.path, filepath.Ext(pb.path))
	return &DefaultPathBuilder{
		path:    base + ext,
		options: pb.options,
	}
}

// WithoutExt removes the extension
func (pb *DefaultPathBuilder) WithoutExt() PathBuilder {
	return pb.WithExt("")
}

// WithName changes the filename (keeping directory)
func (pb *DefaultPathBuilder) WithName(name string) PathBuilder {
	dir := filepath.Dir(pb.path)
	return &DefaultPathBuilder{
		path:    filepath.Join(dir, name),
		options: pb.options,
	}
}

// Relative operations

// Rel returns a relative path from basepath
func (pb *DefaultPathBuilder) Rel(basepath string) (PathBuilder, error) {
	rel, err := filepath.Rel(basepath, pb.path)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	return &DefaultPathBuilder{
		path:    rel,
		options: pb.options,
	}, nil
}

// RelTo returns a relative path to target
func (pb *DefaultPathBuilder) RelTo(target PathBuilder) (PathBuilder, error) {
	return pb.Rel(target.String())
}

// Information

// String returns the string representation of the path
func (pb *DefaultPathBuilder) String() string {
	return pb.path
}

// IsAbs returns true if the path is absolute
func (pb *DefaultPathBuilder) IsAbs() bool {
	return filepath.IsAbs(pb.path)
}

// IsDir returns true if the path is a directory
func (pb *DefaultPathBuilder) IsDir() bool {
	info, err := os.Stat(pb.path)
	return err == nil && info.IsDir()
}

// IsFile returns true if the path is a file
func (pb *DefaultPathBuilder) IsFile() bool {
	info, err := os.Stat(pb.path)
	return err == nil && !info.IsDir()
}

// Exists returns true if the path exists
func (pb *DefaultPathBuilder) Exists() bool {
	_, err := os.Stat(pb.path)
	return err == nil
}

// Size returns the size of the file
func (pb *DefaultPathBuilder) Size() int64 {
	info, err := os.Stat(pb.path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// ModTime returns the modification time
func (pb *DefaultPathBuilder) ModTime() time.Time {
	info, err := os.Stat(pb.path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// Mode returns the file mode
func (pb *DefaultPathBuilder) Mode() fs.FileMode {
	info, err := os.Stat(pb.path)
	if err != nil {
		return 0
	}
	return info.Mode()
}

// Directory operations

// Walk walks the directory tree
func (pb *DefaultPathBuilder) Walk(fn WalkFunc) error {
	return filepath.WalkDir(pb.path, func(path string, d fs.DirEntry, err error) error {
		var info fs.FileInfo
		if d != nil {
			var infoErr error
			info, infoErr = d.Info()
			if infoErr != nil {
				// Continue with nil info - let the walk function handle it
				info = nil
			}
		}
		pathBuilder := NewPathBuilder(path)
		return fn(pathBuilder, info, err)
	})
}

// List returns all entries in the directory
func (pb *DefaultPathBuilder) List() ([]PathBuilder, error) {
	entries, err := os.ReadDir(pb.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := make([]PathBuilder, 0, len(entries))
	for _, entry := range entries {
		path := filepath.Join(pb.path, entry.Name())
		result = append(result, NewPathBuilder(path))
	}

	return result, nil
}

// ListFiles returns only files in the directory
func (pb *DefaultPathBuilder) ListFiles() ([]PathBuilder, error) {
	entries, err := pb.List()
	if err != nil {
		return nil, err
	}

	result := make([]PathBuilder, 0)
	for _, entry := range entries {
		if entry.IsFile() {
			result = append(result, entry)
		}
	}

	return result, nil
}

// ListDirs returns only directories
func (pb *DefaultPathBuilder) ListDirs() ([]PathBuilder, error) {
	entries, err := pb.List()
	if err != nil {
		return nil, err
	}

	result := make([]PathBuilder, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry)
		}
	}

	return result, nil
}

// Glob returns paths matching the pattern
func (pb *DefaultPathBuilder) Glob(pattern string) ([]PathBuilder, error) {
	fullPattern := filepath.Join(pb.path, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern failed: %w", err)
	}

	result := make([]PathBuilder, 0, len(matches))
	for _, match := range matches {
		result = append(result, NewPathBuilder(match))
	}

	return result, nil
}

// Validation

// Validate checks if the path is valid
func (pb *DefaultPathBuilder) Validate() error {
	if pb.path == "" {
		return &ValidationError{
			Path:    pb.path,
			Rule:    "non-empty",
			Message: "path cannot be empty",
			Code:    "EMPTY_PATH",
		}
	}

	// Check for unsafe characters
	if !pb.options.AllowUnsafePaths {
		if strings.Contains(pb.path, "..") {
			return &ValidationError{
				Path:    pb.path,
				Rule:    "safe-path",
				Message: "path contains unsafe '..' component",
				Code:    "UNSAFE_PATH",
			}
		}
	}

	// Check path length
	if pb.options.MaxPathLength > 0 && len(pb.path) > pb.options.MaxPathLength {
		return &ValidationError{
			Path:    pb.path,
			Rule:    "max-length",
			Message: fmt.Sprintf("path exceeds maximum length of %d", pb.options.MaxPathLength),
			Code:    "PATH_TOO_LONG",
		}
	}

	// Check base path restriction
	if pb.options.RestrictToBasePath != "" {
		abs, err := filepath.Abs(pb.path)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}

		baseAbs, err := filepath.Abs(pb.options.RestrictToBasePath)
		if err != nil {
			return fmt.Errorf("failed to resolve base path: %w", err)
		}

		rel, err := filepath.Rel(baseAbs, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			return &ValidationError{
				Path:    pb.path,
				Rule:    "base-path",
				Message: fmt.Sprintf("path is outside of base path %s", pb.options.RestrictToBasePath),
				Code:    "OUTSIDE_BASE_PATH",
			}
		}
	}

	return nil
}

// IsValid returns true if the path is valid
func (pb *DefaultPathBuilder) IsValid() bool {
	return pb.Validate() == nil
}

// IsEmpty returns true if the path is empty
func (pb *DefaultPathBuilder) IsEmpty() bool {
	return pb.path == ""
}

// IsSafe returns true if the path is considered safe
func (pb *DefaultPathBuilder) IsSafe() bool {
	return !strings.Contains(pb.path, "..")
}

// Modification

// Create creates the file
func (pb *DefaultPathBuilder) Create() error {
	if pb.options.CreateParents {
		dir := filepath.Dir(pb.path)
		if err := os.MkdirAll(dir, pb.options.CreateMode); err != nil {
			return fmt.Errorf("failed to create parent directories: %w", err)
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if !pb.options.OverwriteExisting {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(pb.path, flags, pb.options.CreateMode)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return file.Close()
}

// CreateDir creates the directory
func (pb *DefaultPathBuilder) CreateDir() error {
	return os.Mkdir(pb.path, pb.options.CreateMode)
}

// CreateDirAll creates the directory and all parent directories
func (pb *DefaultPathBuilder) CreateDirAll() error {
	return os.MkdirAll(pb.path, pb.options.CreateMode)
}

// Remove removes the file or empty directory
func (pb *DefaultPathBuilder) Remove() error {
	return os.Remove(pb.path)
}

// RemoveAll removes the path and any children it contains
func (pb *DefaultPathBuilder) RemoveAll() error {
	return os.RemoveAll(pb.path)
}

// Copy copies the file or directory to dest
func (pb *DefaultPathBuilder) Copy(dest PathBuilder) error {
	return pb.copyRecursive(pb.path, dest.String())
}

// Move moves the file or directory to dest
func (pb *DefaultPathBuilder) Move(dest PathBuilder) error {
	return os.Rename(pb.path, dest.String())
}

// Links

// Readlink returns the target of a symbolic link
func (pb *DefaultPathBuilder) Readlink() (PathBuilder, error) {
	target, err := os.Readlink(pb.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read link: %w", err)
	}
	return NewPathBuilder(target), nil
}

// Symlink creates a symbolic link to target
func (pb *DefaultPathBuilder) Symlink(target PathBuilder) error {
	return os.Symlink(target.String(), pb.path)
}

// Matching

// Match returns true if the path matches the pattern
func (pb *DefaultPathBuilder) Match(pattern string) bool {
	matched, err := filepath.Match(pattern, pb.Base())
	if err != nil {
		// Invalid pattern - return false
		return false
	}
	return matched
}

// Contains returns true if the path contains the substring
func (pb *DefaultPathBuilder) Contains(sub string) bool {
	return strings.Contains(pb.path, sub)
}

// HasPrefix returns true if the path has the prefix
func (pb *DefaultPathBuilder) HasPrefix(prefix string) bool {
	return strings.HasPrefix(pb.path, prefix)
}

// HasSuffix returns true if the path has the suffix
func (pb *DefaultPathBuilder) HasSuffix(suffix string) bool {
	return strings.HasSuffix(pb.path, suffix)
}

// Cloning

// Clone creates a copy of the path builder
func (pb *DefaultPathBuilder) Clone() PathBuilder {
	return &DefaultPathBuilder{
		path:    pb.path,
		options: pb.options,
	}
}

// Helper methods

// copyRecursive recursively copies files and directories
func (pb *DefaultPathBuilder) copyRecursive(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return pb.copyDir(src, dst, srcInfo)
	}
	return pb.copyFile(src, dst, srcInfo)
}

// copyFile copies a single file
func (pb *DefaultPathBuilder) copyFile(src, dst string, srcInfo fs.FileInfo) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			// Log the error but don't fail the operation
			log.Printf("failed to close source file %s: %v", src, closeErr)
		}
	}()

	// Create destination directory if needed
	if pb.options.CreateParents {
		dstDir := filepath.Dir(dst)
		if mkdirErr := os.MkdirAll(dstDir, pb.options.CreateMode); mkdirErr != nil {
			return mkdirErr
		}
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil {
			// Log the error but don't fail the operation
			log.Printf("failed to close destination file %s: %v", dst, closeErr)
		}
	}()

	// Copy file contents
	buffer := make([]byte, pb.options.BufferSize)
	for {
		n, err := srcFile.Read(buffer)
		if n > 0 {
			if _, writeErr := dstFile.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}

	// Preserve attributes if requested
	if pb.options.PreserveMode {
		if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
			return err
		}
	}

	if pb.options.PreserveMtime {
		if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
			return err
		}
	}

	return nil
}

// copyDir copies a directory recursively
func (pb *DefaultPathBuilder) copyDir(src, dst string, srcInfo fs.FileInfo) error {
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := pb.copyRecursive(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// Package-level convenience functions

// Join creates a path by joining elements
func Join(elements ...string) PathBuilder {
	return NewPathBuilder(filepath.Join(elements...))
}

// FromString creates a path builder from a string
func FromString(path string) PathBuilder {
	return NewPathBuilder(path)
}

// Temp creates a path builder for a temporary file
func Temp(pattern string) (PathBuilder, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	path := file.Name()
	if closeErr := file.Close(); closeErr != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", closeErr)
	}
	if removeErr := os.Remove(path); removeErr != nil {
		return nil, fmt.Errorf("failed to remove temp file: %w", removeErr)
	}
	return NewPathBuilder(path), nil
}

// TempDir creates a path builder for a temporary directory
func TempDir(pattern string) (PathBuilder, error) {
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return nil, err
	}
	return NewPathBuilder(dir), nil
}

// Exists checks if a path exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// IsFile checks if a path is a file
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// GlobPaths returns all paths matching the pattern
func GlobPaths(pattern string) ([]PathBuilder, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	result := make([]PathBuilder, 0, len(matches))
	for _, match := range matches {
		result = append(result, NewPathBuilder(match))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result, nil
}
