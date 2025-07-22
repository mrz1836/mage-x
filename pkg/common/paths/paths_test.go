package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathBuilder_BasicOperations(t *testing.T) {
	// Test Join
	pb := NewPathBuilder("home").Join("user", "documents")
	expected := filepath.Join("home", "user", "documents")
	assert.Equal(t, expected, pb.String(), "Join() should create correct path")

	// Test Dir
	dir := pb.Dir()
	expectedDir := filepath.Dir(expected)
	assert.Equal(t, expectedDir, dir.String(), "Dir() should return parent directory")

	// Test Base
	base := pb.Base()
	expectedBase := filepath.Base(expected)
	assert.Equal(t, expectedBase, base, "Base() should return filename")

	// Test Clean
	dirty := NewPathBuilder("home//user/../user/./documents")
	clean := dirty.Clean()
	assert.Equal(t, expected, clean.String(), "Clean() should normalize path")
}

func TestPathBuilder_Extensions(t *testing.T) {
	pb := NewPathBuilder("test.txt")

	// Test Ext
	assert.Equal(t, ".txt", pb.Ext(), "Ext() should return file extension")

	// Test WithExt
	withGo := pb.WithExt(".go")
	assert.Equal(t, "test.go", withGo.String(), "WithExt() should change extension")

	// Test WithExt without dot
	withJs := pb.WithExt("js")
	assert.Equal(t, "test.js", withJs.String(), "WithExt() should handle extension without dot")

	// Test WithoutExt
	withoutExt := pb.WithoutExt()
	assert.Equal(t, "test", withoutExt.String(), "WithoutExt() should remove extension")
}

func TestPathBuilder_Modifications(t *testing.T) {
	pb := NewPathBuilder("test.txt")

	// Test Append
	appended := pb.Append("_backup")
	assert.Equal(t, "test_backup.txt", appended.String(), "Append() should add suffix before extension")

	// Test Prepend
	prepended := pb.Prepend("old_")
	assert.Equal(t, "old_test.txt", prepended.String(), "Prepend() should add prefix to filename")

	// Test WithName
	withName := pb.WithName("new.go")
	assert.Equal(t, "new.go", withName.String(), "WithName() should replace filename")
}

func TestPathBuilder_TempOperations(t *testing.T) {
	// Test Temp file
	tempFile, err := Temp("test_*")
	require.NoError(t, err, "Temp() should not fail")

	tempPath := tempFile.String()
	assert.Contains(t, tempPath, "test_", "Temp file name should contain pattern")

	// Test TempDir
	tempDir, err := TempDir("testdir_*")
	require.NoError(t, err, "TempDir() should not fail")
	defer os.RemoveAll(tempDir.String())

	assert.True(t, tempDir.Exists(), "TempDir should exist")
	assert.True(t, tempDir.IsDir(), "TempDir should be a directory")
}

func TestPathBuilder_FileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := TempDir("pathtest_*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir.String())

	// Test Create
	testFile := tempDir.Join("test.txt")
	require.NoError(t, testFile.Create(), "Create() should not fail")

	assert.True(t, testFile.Exists(), "Created file should exist")
	assert.True(t, testFile.IsFile(), "Created path should be a file")

	// Test CreateDir
	testSubDir := tempDir.Join("subdir")
	require.NoError(t, testSubDir.CreateDir(), "CreateDir() should not fail")

	assert.True(t, testSubDir.IsDir(), "Created path should be a directory")

	// Test Remove
	require.NoError(t, testFile.Remove(), "Remove() should not fail")

	assert.False(t, testFile.Exists(), "Removed file should not exist")
}

func TestPathBuilder_Listing(t *testing.T) {
	// Create a temporary directory with some files
	tempDir, err := TempDir("listtest_*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir.String())

	// Create test files and directories
	testFile1 := tempDir.Join("file1.txt")
	testFile2 := tempDir.Join("file2.go")
	testSubDir := tempDir.Join("subdir")

	testFile1.Create()
	testFile2.Create()
	testSubDir.CreateDir()

	// Test List
	entries, err := tempDir.List()
	require.NoError(t, err, "List() should not fail")

	assert.Len(t, entries, 3, "List() should return 3 entries")

	// Test ListFiles
	files, err := tempDir.ListFiles()
	require.NoError(t, err, "ListFiles() should not fail")

	assert.Len(t, files, 2, "ListFiles() should return 2 files")

	// Test ListDirs
	dirs, err := tempDir.ListDirs()
	require.NoError(t, err, "ListDirs() should not fail")

	assert.Len(t, dirs, 1, "ListDirs() should return 1 directory")

	// Test Glob
	goFiles, err := tempDir.Glob("*.go")
	require.NoError(t, err, "Glob() should not fail")

	assert.Len(t, goFiles, 1, "Glob() should return 1 .go file")
}

func TestPathBuilder_Validation(t *testing.T) {
	// Test valid path
	validPath := NewPathBuilder("valid/path")
	assert.True(t, validPath.IsValid(), "Valid path should be valid")

	// Test path with unsafe components
	unsafePath := NewPathBuilder("../unsafe/path")
	assert.False(t, unsafePath.IsSafe(), "Unsafe path should not be safe")

	// Test empty path (filepath.Clean("") returns ".")
	emptyPath := NewPathBuilder("")
	assert.Equal(t, ".", emptyPath.String(), "Empty path should be cleaned to '.'")

	// Create actual empty path for testing
	actualEmptyPath := &DefaultPathBuilder{path: "", options: GetDefaultOptions()}
	assert.True(t, actualEmptyPath.IsEmpty(), "Actual empty path should be empty")
	assert.False(t, actualEmptyPath.IsValid(), "Actual empty path should not be valid")
}

func TestPathBuilder_Matching(t *testing.T) {
	pb := NewPathBuilder("test.txt")

	// Test Match
	assert.True(t, pb.Match("*.txt"), "Should match *.txt pattern")
	assert.False(t, pb.Match("*.go"), "Should not match *.go pattern")

	// Test Contains
	assert.True(t, pb.Contains("test"), "Should contain 'test'")

	// Test HasPrefix and HasSuffix
	longPath := NewPathBuilder("/home/user/documents/test.txt")
	assert.True(t, longPath.HasPrefix("/home"), "Should have prefix '/home'")
	assert.True(t, longPath.HasSuffix(".txt"), "Should have suffix '.txt'")
}

func TestPathBuilder_Clone(t *testing.T) {
	original := NewPathBuilder("test.txt")
	clone := original.Clone()

	assert.Equal(t, original.String(), clone.String(), "Clone should have same path")

	// Modify clone and ensure original is unchanged
	modified := clone.WithExt(".go")
	assert.NotEqual(t, original.String(), modified.String(), "Original should not be affected by clone modification")
}

func TestPathMatcher(t *testing.T) {
	matcher := NewPathMatcher()

	// Test AddPattern
	err := matcher.AddPattern("*.txt")
	require.NoError(t, err, "AddPattern() should not fail")

	// Test Match
	assert.True(t, matcher.Match("test.txt"), "Should match test.txt")
	assert.False(t, matcher.Match("test.go"), "Should not match test.go")

	// Test multiple patterns
	matcher.AddPattern("*.go")
	assert.True(t, matcher.Match("test.go"), "Should match test.go after adding pattern")

	// Test case sensitivity
	matcher.SetCaseSensitive(false)
	assert.True(t, matcher.Match("TEST.TXT"), "Should match TEST.TXT when case insensitive")

	// Test pattern removal
	err = matcher.RemovePattern("*.txt")
	require.NoError(t, err, "RemovePattern() should not fail")

	assert.False(t, matcher.Match("test.txt"), "Should not match test.txt after pattern removal")
}

func TestPathValidator(t *testing.T) {
	validator := NewPathValidator()

	// Test RequireAbsolute
	validator.RequireAbsolute()

	errors := validator.Validate("relative/path")
	assert.NotEmpty(t, errors, "Should have validation errors for relative path")

	absPath := "/absolute/path"
	errors = validator.Validate(absPath)
	assert.Empty(t, errors, "Should not have validation errors for absolute path")

	// Test extension validation
	extValidator := NewPathValidator()
	extValidator.RequireExtension("txt", "go")

	errors = extValidator.Validate("test.txt")
	assert.Empty(t, errors, "Should not have errors for .txt file")

	errors = extValidator.Validate("test.js")
	assert.NotEmpty(t, errors, "Should have errors for .js file")
}

func TestPathSet(t *testing.T) {
	set := NewPathSet()

	// Test Add
	assert.True(t, set.Add("path1"), "Add() should return true for new path")
	assert.False(t, set.Add("path1"), "Add() should return false for existing path")

	// Test Contains
	assert.True(t, set.Contains("path1"), "Should contain added path")

	// Test Size
	set.Add("path2")
	set.Add("path3")
	assert.Equal(t, 3, set.Size(), "Size() should be 3")

	// Test Remove
	assert.True(t, set.Remove("path2"), "Remove() should return true for existing path")
	assert.False(t, set.Contains("path2"), "Should not contain removed path")

	// Test set operations
	set1 := NewPathSetFromPaths([]string{"a", "b", "c"})
	set2 := NewPathSetFromPaths([]string{"c", "d", "e"})

	// Test Union
	union := set1.Union(set2)
	assert.Equal(t, 5, union.Size(), "Union size should be 5")

	// Test Intersection
	intersection := set1.Intersection(set2)
	assert.Equal(t, 1, intersection.Size(), "Intersection should have size 1")
	assert.True(t, intersection.Contains("c"), "Intersection should contain 'c'")

	// Test Difference
	difference := set1.Difference(set2)
	assert.Equal(t, 2, difference.Size(), "Difference should have size 2")
	assert.True(t, difference.Contains("a"), "Difference should contain 'a'")
	assert.True(t, difference.Contains("b"), "Difference should contain 'b'")
}

func TestPackageConvenienceFunctions(t *testing.T) {
	// Test Join
	joined := Join("home", "user", "documents")
	expected := filepath.Join("home", "user", "documents")
	assert.Equal(t, expected, joined.String(), "Join() should create correct path")

	// Test FromString
	pb := FromString("test/path")
	assert.Equal(t, "test/path", pb.String(), "FromString() should return correct path")

	// Test GlobPaths (create some test files first)
	tempDir, err := TempDir("globtest_*")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tempDir.String())

	// Create test files
	tempDir.Join("test1.txt").Create()
	tempDir.Join("test2.txt").Create()
	tempDir.Join("test1.go").Create()

	pattern := filepath.Join(tempDir.String(), "*.txt")
	matches, err := GlobPaths(pattern)
	require.NoError(t, err, "GlobPaths() should not fail")

	assert.Len(t, matches, 2, "GlobPaths() should return 2 matches")
}

func TestPathBuilder_RelativeOperations(t *testing.T) {
	// Create absolute paths for testing
	absPath1 := "/home/user/documents/file.txt"
	absPath2 := "/home/user"

	pb1 := NewPathBuilder(absPath1)
	pb2 := NewPathBuilder(absPath2)

	// Test Rel
	rel, err := pb1.Rel(absPath2)
	require.NoError(t, err, "Rel() should not fail")

	expected := "documents/file.txt"
	assert.Equal(t, expected, rel.String(), "Rel() should return correct relative path")

	// Test RelTo
	relTo, err := pb1.RelTo(pb2)
	require.NoError(t, err, "RelTo() should not fail")

	assert.Equal(t, expected, relTo.String(), "RelTo() should return correct relative path")
}

// Benchmark tests

func BenchmarkPathBuilder_Join(b *testing.B) {
	pb := NewPathBuilder("home")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.Join("user", "documents")
	}
}

func BenchmarkPathBuilder_String(b *testing.B) {
	pb := NewPathBuilder("home/user/documents/file.txt")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.String()
	}
}

func BenchmarkPathMatcher_Match(b *testing.B) {
	matcher := NewPathMatcher()
	matcher.AddPattern("*.txt")
	matcher.AddPattern("*.go")
	matcher.AddPattern("test_*")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.Match("test_file.txt")
	}
}

func BenchmarkPathSet_Add(b *testing.B) {
	set := NewPathSet()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Add("path_" + string(rune(i)))
	}
}

func BenchmarkPathBuilder_Exists(b *testing.B) {
	pb := NewPathBuilder("/")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.Exists()
	}
}

func BenchmarkPathValidator_Validate(b *testing.B) {
	validator := NewPathValidator()
	validator.RequireAbsolute()
	validator.RequireExtension("txt", "go")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate("/home/user/test.txt")
	}
}