package utils

import (
	"os"
	"path/filepath"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	return fileops.GetDefault().File.Exists(path)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	return fileops.GetDefault().File.Exists(path) && fileops.GetDefault().File.IsDir(path)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return fileops.GetDefault().File.MkdirAll(path, fileops.PermDir)
}

// CleanDir removes and recreates a directory
func CleanDir(path string) error {
	if err := fileops.GetDefault().File.RemoveAll(path); err != nil {
		// Ignore "not exists" errors, similar to original behavior
		if !os.IsNotExist(err) {
			return err
		}
	}
	return fileops.GetDefault().File.MkdirAll(path, fileops.PermDir)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	return fileops.GetDefault().File.Copy(src, dst)
}

// FindFiles finds files matching a pattern
func FindFiles(_, pattern string) ([]string, error) {
	return findFiles(".", pattern)
}

func findFiles(root, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		// Use fileops to check if it's a file (not directory)
		if matched && fileops.GetDefault().File.Exists(path) && !fileops.GetDefault().File.IsDir(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
