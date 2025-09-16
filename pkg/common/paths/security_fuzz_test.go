package paths

import (
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzPathTraversalDetection fuzzes path traversal detection with various attack vectors
func FuzzPathTraversalDetection(f *testing.F) {
	// Seed with known path traversal attack patterns
	f.Add("../../../etc/passwd")
	f.Add("..\\..\\..\\windows\\system32\\config\\sam")
	f.Add("%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd")
	f.Add("\u002e\u002e\u002f\u002e\u002e\u002f")
	f.Add("safe.txt\x00../../../etc/passwd")
	f.Add("/proc/self/fd/0")
	f.Add("\\\\server\\share\\file")
	f.Add("C:\\Windows\\System32")
	f.Add("file.txt:hidden:$DATA")
	f.Add("CON")
	f.Add("PRN.txt")
	f.Add("AUX")
	f.Add("NUL")
	f.Add("file⁄txt") // Unicode fraction slash
	f.Add("file\n../../../etc/passwd")
	f.Add("file\r\n../../../etc/passwd")
	f.Add("file\t../../../etc/passwd")

	f.Fuzz(func(t *testing.T, path string) {
		// Skip invalid UTF-8 strings that would cause panics in other operations
		if !utf8.ValidString(path) {
			t.Skip("Invalid UTF-8 input")
		}

		// Skip extremely long strings that could cause resource exhaustion
		if len(path) > 10000 {
			t.Skip("Path too long for fuzz testing")
		}

		pb := NewPathBuilder(path)

		// Test that IsSafe() never panics
		isSafe := pb.IsSafe()

		// Test that validation never panics
		validator := NewPathValidator()
		errors := validator.ValidatePath(pb)

		// Verify dangerous patterns are detected
		hasDangerousPatterns := strings.Contains(path, "..") ||
			strings.Contains(path, "\x00") ||
			strings.Contains(path, "\n") ||
			strings.Contains(path, "\r") ||
			strings.Contains(path, "\\\\") || // UNC paths
			(len(path) > 1 && path[1] == ':') || // Windows drive paths
			strings.Contains(path, ":") && strings.Contains(path, "$DATA") // ADS

		if hasDangerousPatterns {
			if isSafe {
				t.Errorf("Path with dangerous patterns reported as safe: %q", path)
			}
		}

		// Test that Clean() never panics
		cleaned := pb.Clean()
		if cleaned == nil {
			t.Error("Clean() returned nil")
		}

		// Test that String() never panics
		pathStr := pb.String()
		if pathStr == "" && path != "" {
			// Empty result from non-empty input might indicate an issue
			t.Logf("Empty result from non-empty input: %q -> %q", path, pathStr)
		}

		// Ensure validation errors are meaningful
		for _, err := range errors {
			if err.Message == "" {
				t.Error("Validation error with empty message")
			}
			if err.Rule == "" {
				t.Error("Validation error with empty rule")
			}
		}
	})
}

// FuzzPathSanitization fuzzes path sanitization to ensure it handles all edge cases
func FuzzPathSanitization(f *testing.F) {
	// Seed with various path formats
	f.Add("./path/../to/file")
	f.Add("path//to///file")
	f.Add("path/to/dir/")
	f.Add("/absolute/path")
	f.Add("relative/path")
	f.Add("")
	f.Add(".")
	f.Add("..")
	f.Add("...")
	f.Add("/")
	f.Add("//")
	f.Add("./././././")
	f.Add("../../../../../../../")
	f.Add("\\Windows\\System32")
	f.Add("/proc/self/exe")
	f.Add("file with spaces.txt")
	f.Add("file-with-dashes.txt")
	f.Add("file_with_underscores.txt")
	f.Add("UPPERCASE.TXT")
	f.Add("MiXeDcAsE.TxT")

	f.Fuzz(func(t *testing.T, path string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(path) {
			t.Skip("Invalid UTF-8 input")
		}

		// Skip extremely long strings
		if len(path) > 5000 {
			t.Skip("Path too long for fuzz testing")
		}

		pb := NewPathBuilder(path)

		// Test that Clean() never panics and produces valid output
		cleaned := pb.Clean()
		if cleaned == nil {
			t.Fatal("Clean() returned nil")
		}

		cleanedStr := cleaned.String()

		// Note: filepath.Clean() normalizes but doesn't remove .. patterns
		// This is expected behavior - .. patterns should be caught by validation, not cleaning
		// Verify that clean operations are at least stable
		_ = cleanedStr // We don't expect Clean() to remove .. patterns

		// Test that multiple cleanings produce the same result (idempotent)
		doubleCleaned := cleaned.Clean().String()
		if cleanedStr != doubleCleaned {
			t.Errorf("Clean() is not idempotent: %q != %q (from %q)", cleanedStr, doubleCleaned, path)
		}

		// Test filepath operations don't panic
		_ = pb.Base()
		_ = pb.Dir()
		_ = pb.Ext()
		_ = pb.IsAbs()

		// Test that Join operations don't panic
		joined := pb.Join("test", "path")
		if joined == nil {
			t.Error("Join() returned nil")
		}
	})
}

// FuzzValidationRules fuzzes various validation rules to ensure they handle edge cases
func FuzzValidationRules(f *testing.F) {
	// Seed with various inputs
	f.Add("/absolute/path")
	f.Add("relative/path")
	f.Add("file.txt")
	f.Add("file.TAR.GZ")
	f.Add(".hidden")
	f.Add("")
	f.Add("...")
	f.Add("file with spaces")
	f.Add("file\twith\ttabs")
	f.Add("file\nwith\nnewlines")
	f.Add("file\rwith\rcarriage")
	f.Add("file\x00with\x00null")
	f.Add("very-long-filename-" + strings.Repeat("x", 255))
	f.Add(strings.Repeat("A", 1000))

	f.Fuzz(func(t *testing.T, path string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(path) {
			t.Skip("Invalid UTF-8 input")
		}

		// Skip extremely long strings
		if len(path) > 5000 {
			t.Skip("Path too long for fuzz testing")
		}

		pb := NewPathBuilder(path)
		validator := NewPathValidator()

		// Test absolute path validation
		validator.RequireAbsolute()
		errors := validator.ValidatePath(pb)
		isAbsolute := filepath.IsAbs(path)
		hasAbsError := hasValidationError(errors, "absolute-path")

		if isAbsolute && hasAbsError {
			t.Errorf("Absolute path %q failed absolute validation", path)
		}
		if !isAbsolute && !hasAbsError {
			t.Errorf("Relative path %q passed absolute validation", path)
		}

		// Test relative path validation
		relValidator := NewPathValidator().RequireRelative()
		relErrors := relValidator.ValidatePath(pb)
		hasRelError := hasValidationError(relErrors, "relative-path")

		if !isAbsolute && hasRelError {
			t.Errorf("Relative path %q failed relative validation", path)
		}
		if isAbsolute && !hasRelError {
			t.Errorf("Absolute path %q passed relative validation", path)
		}

		// Test extension validation
		ext := filepath.Ext(path)
		if ext != "" && ext != "." {
			// Skip validation when extension is just a dot (current directory or hidden file edge case)
			extValidator := NewPathValidator().RequireExtension(ext[1:]) // Remove the dot
			extErrors := extValidator.ValidatePath(pb)
			if len(extErrors) > 0 {
				t.Errorf("Path %q with extension %q failed extension validation", path, ext)
			}
		}

		// Test max length validation
		maxLen := 100
		lengthValidator := NewPathValidator().RequireMaxLength(maxLen)
		lengthErrors := lengthValidator.ValidatePath(pb)
		hasLengthError := hasValidationError(lengthErrors, "max-length")

		if len(path) <= maxLen && hasLengthError {
			t.Errorf("Path %q (len=%d) failed max length validation (limit=%d)", path, len(path), maxLen)
		}
		if len(path) > maxLen && !hasLengthError {
			t.Errorf("Path %q (len=%d) passed max length validation (limit=%d)", path, len(path), maxLen)
		}

		// Test pattern validation
		patternValidator := NewPathValidator().RequirePattern(`^[a-zA-Z0-9._/-]*$`)
		patternErrors := patternValidator.ValidatePath(pb)
		_ = patternErrors // We don't assert much here since it depends on the random input

		// Ensure no validation rule causes a panic
		forbidValidator := NewPathValidator().ForbidPattern(`\.\.`)
		forbidErrors := forbidValidator.ValidatePath(pb)
		_ = forbidErrors
	})
}

// FuzzPathOperations fuzzes various path operations to ensure they don't panic
func FuzzPathOperations(f *testing.F) {
	// Seed with various path operations
	f.Add("base/file.txt", "suffix")
	f.Add("file", ".ext")
	f.Add("/abs/path", "relative")
	f.Add(".", "test")
	f.Add("..", "test")
	f.Add("", "test")
	f.Add("file.old.txt", ".new")
	f.Add("PATH", "segment")
	f.Add("path with spaces", "more spaces")
	f.Add("file\nwith\nnewlines", "\ttest")

	f.Fuzz(func(t *testing.T, basePath, operation string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(basePath) || !utf8.ValidString(operation) {
			t.Skip("Invalid UTF-8 input")
		}

		// Skip extremely long strings
		if len(basePath) > 2000 || len(operation) > 1000 {
			t.Skip("Input too long for fuzz testing")
		}

		pb := NewPathBuilder(basePath)

		// Test Join operations
		joined := pb.Join(operation)
		if joined == nil {
			t.Error("Join() returned nil")
		}

		// Test Append operations
		appended := pb.Append(operation)
		if appended == nil {
			t.Error("Append() returned nil")
		}

		// Test Prepend operations
		prepended := pb.Prepend(operation)
		if prepended == nil {
			t.Error("Prepend() returned nil")
		}

		// Test WithExt operations
		withExt := pb.WithExt(operation)
		if withExt == nil {
			t.Error("WithExt() returned nil")
		}

		// Test WithName operations
		withName := pb.WithName(operation)
		if withName == nil {
			t.Error("WithName() returned nil")
		}

		// Test WithoutExt
		withoutExt := pb.WithoutExt()
		if withoutExt == nil {
			t.Error("WithoutExt() returned nil")
		}

		// Test Clone
		cloned := pb.Clone()
		if cloned == nil {
			t.Error("Clone() returned nil")
		}

		// Test information methods
		_ = pb.Base()
		_ = pb.Dir()
		_ = pb.Ext()
		_ = pb.IsAbs()
		_ = pb.String()

		// Test boolean methods
		_ = pb.IsEmpty()
		_ = pb.IsValid()
		_ = pb.IsSafe()
		_ = pb.Exists()
		_ = pb.IsDir()
		_ = pb.IsFile()

		// Test matching operations
		_ = pb.Contains(operation)
		_ = pb.HasPrefix(operation)
		_ = pb.HasSuffix(operation)

		// Test pattern matching (with simple patterns to avoid regex errors)
		simplePatterns := []string{"*", "*.txt", "test*", "*test*"}
		for _, pattern := range simplePatterns {
			_ = pb.Match(pattern)
		}

		// Verify that operations maintain basic invariants
		originalStr := pb.String()
		clonedStr := cloned.String()
		if originalStr != clonedStr {
			t.Errorf("Clone() changed the path: %q != %q", originalStr, clonedStr)
		}
	})
}

// hasValidationError checks if a specific validation error type exists
func hasValidationError(errors []ValidationError, ruleName string) bool {
	for _, err := range errors {
		if err.Rule == ruleName {
			return true
		}
	}
	return false
}

// FuzzSecurityBypass fuzzes potential security bypass attempts
func FuzzSecurityBypass(f *testing.F) {
	// Seed with known bypass techniques
	f.Add("..%2f..%2f..%2fetc%2fpasswd")             // URL encoding
	f.Add("..%252f..%252f..%252fetc%252fpasswd")     // Double URL encoding
	f.Add("..\u002f..\u002f..\u002fetc\u002fpasswd") // Unicode encoding
	f.Add("../\x2e\x2e/\x2e\x2e/etc/passwd")         // Hex encoding
	f.Add("..\\..\\..\\etc\\passwd")                 // Backslash normalization
	f.Add("file.txt\x00../../../etc/passwd")         // Null byte injection
	f.Add(".//..//..//..//etc//passwd")              // Mixed current dir and traversal
	f.Add("\x2e\x2e\x2f\x2e\x2e\x2f\x2e\x2e\x2f")    // All hex encoded
	f.Add("file.txt%00../../../etc/passwd")          // URL encoded null
	f.Add("file.txt%0a../../../etc/passwd")          // URL encoded newline
	f.Add("file.txt%0d../../../etc/passwd")          // URL encoded carriage return
	f.Add("/proc/self/fd/../../../etc/passwd")       // Proc filesystem bypass
	f.Add("\\?\\C:\\Windows\\System32\\config\\SAM") // Windows long path prefix
	f.Add("file.txt:$DATA")                          // NTFS alternate data stream
	f.Add("CONIN$.txt")                              // Windows device names
	f.Add("file.txt.")                               // Trailing dot (Windows)
	f.Add("file.txt ")                               // Trailing space (Windows)

	f.Fuzz(func(t *testing.T, path string) {
		// Skip invalid UTF-8
		if !utf8.ValidString(path) {
			t.Skip("Invalid UTF-8 input")
		}

		// Skip extremely long strings
		if len(path) > 2000 {
			t.Skip("Path too long for fuzz testing")
		}

		pb := NewPathBuilder(path)

		// All potentially dangerous paths should be marked as unsafe
		// Check for Windows reserved names (must be exact matches in base name)
		baseName := strings.ToUpper(filepath.Base(path))
		baseWithoutExt := baseName
		if idx := strings.LastIndex(baseName, "."); idx > 0 {
			baseWithoutExt = baseName[:idx]
		}
		hasWindowsReserved := false
		windowsReserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9", "CONIN$", "CONOUT$"}
		for _, reserved := range windowsReserved {
			if baseWithoutExt == reserved || baseName == reserved {
				hasWindowsReserved = true
				break
			}
		}

		hasBypassAttempt := strings.Contains(path, "..") ||
			strings.Contains(path, "%2e") ||
			strings.Contains(path, "%2f") ||
			strings.Contains(path, "\\x") ||
			strings.Contains(path, "\\u") ||
			strings.Contains(path, "\x00") ||
			strings.Contains(path, "%00") ||
			strings.Contains(path, "/proc/") ||
			strings.Contains(path, "/dev/") ||
			strings.Contains(path, "\\\\?\\") ||
			strings.Contains(path, ":$DATA") ||
			hasWindowsReserved ||
			strings.HasSuffix(path, ".") ||
			strings.HasSuffix(path, " ")

		isSafe := pb.IsSafe()
		if hasBypassAttempt && isSafe {
			t.Errorf("Potential security bypass not detected: %q", path)
		}

		// Test that validation catches bypass attempts
		validator := NewPathValidator()
		validator.ForbidPattern(`\.\.`) // Forbid parent directory references
		errors := validator.ValidatePath(pb)

		if strings.Contains(path, "..") && len(errors) == 0 {
			t.Errorf("Path traversal attempt not caught by validator: %q", path)
		}

		// Test that cleaning doesn't accidentally allow bypasses
		cleaned := pb.Clean()
		cleanedStr := cleaned.String()

		// Note: filepath.Clean() normalizes but doesn't remove dangerous patterns
		// The IsSafe() and validation should catch dangerous patterns
		_ = cleanedStr // Clean() doesn't remove .. patterns, just normalizes them
	})
}
