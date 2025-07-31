//go:build go1.18
// +build go1.18

package security

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// FuzzValidateCommandArg tests command argument validation with fuzzing
func FuzzValidateCommandArg(f *testing.F) {
	// Add seed corpus with known injection attempts
	testcases := []string{
		// Normal arguments
		"--flag",
		"-v",
		"filename.txt",
		"/path/to/file",
		"value",
		"key=value",
		"http://example.com",
		"https://example.com/path?query=value",
		"user@host.com",
		"192.168.1.1",
		"[::1]",
		"localhost:8080",

		// Shell injection attempts
		"$(whoami)",
		"$(echo test)",
		"`whoami`",
		"`echo test`",
		"&&whoami",
		"||whoami",
		";whoami",
		"|whoami",
		">output.txt",
		"<input.txt",
		">>output.txt",
		"2>&1",
		"$(echo${IFS}test)",
		"${IFS}",
		"$IFS",

		// Complex injection attempts
		"test$(whoami)test",
		"test`whoami`test",
		"test&&whoami&&test",
		"test||whoami||test",
		"test;whoami;test",
		"test|whoami|test",
		"$(whoami$(echo test))",
		"`whoami`$(echo test)",
		"test\nwhoami",
		"test\rwhoami",
		"test\r\nwhoami",

		// Path traversal attempts
		"../../etc/passwd",
		"..\\..\\windows\\system32",
		"/etc/passwd",
		"C:\\Windows\\System32\\cmd.exe",
		"\\\\server\\share",

		// Special characters
		"test\x00null",
		"test\nnewline",
		"test\ttab",
		"test\rcarriage",
		"test space",
		"test\"quote",
		"test'quote",
		"test\\backslash",
		"test$dollar",
		"test#hash",
		"test@at",
		"test!exclaim",
		"test%percent",
		"test^caret",
		"test&ampersand",
		"test*asterisk",
		"test(paren)",
		"test[bracket]",
		"test{brace}",
		"test~tilde",

		// Unicode and emoji
		"test🚀rocket",
		"测试",
		"テスト",
		"тест",
		"🚀",
		strings.Repeat("🚀", 100),

		// Long strings
		strings.Repeat("a", 1000),
		strings.Repeat("a", 10000),
		strings.Repeat("$(whoami)", 100),

		// Empty and whitespace
		"",
		" ",
		"\t",
		"\n",
		"\r",
		"\r\n",
		"   ",

		// URLs with potential issues
		"http://example.com/;whoami",
		"https://example.com/$(whoami)",
		"http://example.com|whoami",
		"https://example.com`whoami`",

		// Regex patterns (should be allowed)
		"^test.*$",
		"[a-z]+",
		"test|other",
		"(test|other)",
		"test{1,3}",
		"test?",
		"test*",
		"test+",

		// Environment variable references (should be careful)
		"$HOME",
		"$PATH",
		"${HOME}",
		"${PATH}",
		"%HOME%",
		"%PATH%",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, arg string) {
		// The function should not panic on any input
		err := ValidateCommandArg(arg)

		// Check for consistency
		if err == nil {
			// If accepted, should not contain obvious dangerous patterns
			dangerousPatterns := []string{
				"$(",
				"`",
				"&&",
				"||",
				";",
				"$(echo",
				"${IFS}",
			}

			for _, pattern := range dangerousPatterns {
				assert.NotContains(t, arg, pattern,
					"ValidateCommandArg accepted argument with dangerous pattern '%s': %s", pattern, arg)
			}

			// Special handling for pipe - should only be in regex or URLs
			if strings.Contains(arg, "|") {
				isRegex := strings.ContainsAny(arg, "^$[]()+*?.{}\\")
				isURL := strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://")
				assert.True(t, isRegex || isURL,
					"ValidateCommandArg accepted pipe outside of regex/URL context: %s", arg)
			}

			// Redirects should not be allowed
			assert.NotContains(t, arg, ">", "Argument contains output redirect")
			assert.NotContains(t, arg, "<", "Argument contains input redirect")
		}

		// If the function rejected invalid UTF-8, that's correct behavior
		if err != nil && strings.Contains(err.Error(), "invalid UTF-8") {
			assert.False(t, utf8.ValidString(arg), "Function rejected valid UTF-8")
		}
	})
}

// FuzzValidatePath tests path validation with fuzzing
func FuzzValidatePath(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		// Normal paths
		"file.txt",
		"dir/file.txt",
		"dir/subdir/file.txt",
		"./file.txt",
		"./dir/file.txt",
		"dir/./file.txt",
		"dir/../dir/file.txt",
		"/tmp/file.txt",
		"/tmp/dir/file.txt",

		// Path traversal attempts
		"../file.txt",
		"../../file.txt",
		"../../../etc/passwd",
		"dir/../../../etc/passwd",
		"dir/../../etc/passwd",
		"./../../etc/passwd",
		"./../etc/passwd",
		"..\\..\\windows\\system32",
		"dir\\..\\..\\windows\\system32",

		// Absolute paths
		"/etc/passwd",
		"/home/user/file.txt",
		"C:\\Windows\\System32\\cmd.exe",
		"C:/Windows/System32/cmd.exe",
		"D:\\file.txt",
		"\\\\server\\share\\file.txt",
		"//server/share/file.txt",

		// Special cases
		"",
		".",
		"..",
		"...",
		"/",
		"\\",
		"//",
		"\\\\",

		// Tricky paths
		"file..txt",
		"file...txt",
		"dir..dir/file.txt",
		"dir/..file.txt",
		"dir/file..txt",
		".hidden",
		"..hidden",
		"...hidden",
		"dir/.hidden",
		"dir/..hidden",

		// With special characters
		"file\x00.txt",
		"file\n.txt",
		"file\r\n.txt",
		"file with spaces.txt",
		"file;cmd.txt",
		"file&cmd.txt",
		"file|cmd.txt",
		"file>output.txt",
		"file<input.txt",
		"file`cmd`.txt",
		"file$(cmd).txt",
		"file${var}.txt",

		// Unicode paths
		"файл.txt",
		"文件.txt",
		"🚀.txt",
		"dir/🚀/file.txt",

		// Long paths
		strings.Repeat("a", 255) + ".txt",
		"dir/" + strings.Repeat("b", 255) + "/file.txt",
		strings.Repeat("dir/", 50) + "file.txt",

		// Windows special paths
		"CON",
		"PRN",
		"AUX",
		"NUL",
		"COM1",
		"LPT1",
		"con.txt",
		"prn.txt",

		// URL-like paths
		"http://example.com/file.txt",
		"https://example.com/file.txt",
		"file://localhost/path/file.txt",
		"ftp://server/file.txt",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, path string) {
		// The function should not panic
		err := ValidatePath(path)

		// Check for dangerous patterns - if present, validation should have failed
		cleaned := filepath.Clean(path)
		if strings.Contains(cleaned, "../") || strings.Contains(cleaned, "..\\") ||
			cleaned == ".." || strings.HasSuffix(cleaned, "/..") ||
			strings.Contains(path, "\x00") || strings.Contains(path, "\n") ||
			strings.Contains(path, "\r") {
			assert.Error(t, err, "Path with dangerous patterns should be rejected: cleaned=%s, original=%s", cleaned, path)
			return
		}

		if err == nil {
			// Clean the path for verification
			cleaned := filepath.Clean(path)

			// Should not be absolute path (except /tmp)
			if !strings.HasPrefix(cleaned, "/tmp") {
				assert.False(t, filepath.IsAbs(cleaned),
					"ValidatePath accepted absolute path outside /tmp: %s", path)

				// Check for Windows absolute paths
				if runtime.GOOS != "windows" {
					assert.False(t, len(path) > 1 && path[1] == ':',
						"Path looks like Windows drive path: %s", path)
				}

				// Check for UNC paths
				assert.False(t, strings.HasPrefix(path, "\\\\"),
					"Path is UNC path: %s", path)
			}
		}
	})
}

// FuzzFilterEnvironment tests environment variable filtering with fuzzing
func FuzzFilterEnvironment(f *testing.F) {
	// Create executor for testing
	executor := NewSecureExecutor()

	// Add seed corpus
	testcases := [][]string{
		// Normal environment variables
		{"PATH=/usr/bin:/usr/local/bin", "HOME=/home/user", "USER=testuser"},

		// Sensitive variables that should be filtered
		{"AWS_SECRET_ACCESS_KEY=secret123", "PATH=/usr/bin"},
		{"GITHUB_TOKEN=ghp_xxxxxxxxxxxx", "USER=test"},
		{"GITLAB_TOKEN=glpat-xxxxxxxxxxxx", "HOME=/home/user"},
		{"NPM_TOKEN=npm_xxxxxxxxxxxx", "NODE_ENV=production"},
		{"DOCKER_PASSWORD=secret", "DOCKER_USER=user"},
		{"DATABASE_PASSWORD=dbpass123", "DATABASE_HOST=localhost"},
		{"API_KEY=key123", "API_URL=https://api.example.com"},
		{"SECRET_KEY=secret", "APP_ENV=production"},
		{"PRIVATE_KEY=-----BEGIN RSA PRIVATE KEY-----", "PUBLIC_KEY=-----BEGIN PUBLIC KEY-----"},

		// Edge cases
		{"AWS_SECRET=not_exactly_match", "AWS_SECRETS=plural"},
		{"SECRET_API_KEY=should_filter", "API_SECRET_KEY=should_filter"},
		{"SECRETKEY=no_underscore", "SECRET=exact_match"},
		{"MY_SECRET=not_prefix", "SECRET_VALUE=yes_prefix"},

		// Malformed entries
		{"NOEQUALS", "=NOKEY", "KEY=", "=", ""},
		{"KEY=VALUE=EXTRA", "KEY==DOUBLE"},

		// Special characters in values
		{"KEY=value with spaces", "KEY=value\nwith\nnewlines"},
		{"KEY=value\twith\ttabs", "KEY=value;with;semicolons"},
		{"KEY=value|with|pipes", "KEY=value&with&ampersands"},
		{"KEY=value$with$dollars", "KEY=value`with`backticks"},

		// Unicode
		{"KEY=значение", "KEY=値", "KEY=🚀"},
		{"КЛЮЧ=value", "键=value", "🚀=value"},

		// Very long entries
		{"KEY=" + strings.Repeat("a", 1000)},
		{strings.Repeat("A", 255) + "=value"},
		{"AWS_SECRET_ACCESS_KEY=" + strings.Repeat("x", 10000)},
	}

	// Flatten for fuzzing
	for _, envList := range testcases {
		for _, env := range envList {
			f.Add(env)
		}
	}

	f.Fuzz(func(t *testing.T, envVar string) {
		// Create a slice with this environment variable
		env := []string{envVar}

		// Should not panic
		filtered := executor.filterEnvironment(env)

		// Verify filtering logic
		if len(filtered) == 0 {
			// Variable was filtered out
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				varName := strings.ToUpper(parts[0])

				// Check if it should have been filtered
				sensitivePrefix := []string{
					"AWS_SECRET", "GITHUB_TOKEN", "GITLAB_TOKEN", "NPM_TOKEN",
					"DOCKER_PASSWORD", "DATABASE_PASSWORD", "API_KEY", "SECRET", "PRIVATE_KEY",
				}

				wasFiltered := false
				for _, prefix := range sensitivePrefix {
					if strings.HasPrefix(varName, prefix) &&
						(len(varName) == len(prefix) || varName[len(prefix)] == '_') {
						wasFiltered = true
						break
					}
				}

				assert.True(t, wasFiltered || len(parts) < 2,
					"Variable was filtered but shouldn't have been: %s", envVar)
			}
		} else {
			// Variable was kept
			assert.Len(t, filtered, 1, "Filter changed number of variables")
			assert.Equal(t, envVar, filtered[0], "Filter modified variable")

			// Verify it's not sensitive
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				varName := strings.ToUpper(parts[0])

				// These prefixes should be filtered
				sensitivePrefix := []string{
					"AWS_SECRET", "GITHUB_TOKEN", "GITLAB_TOKEN", "NPM_TOKEN",
					"DOCKER_PASSWORD", "DATABASE_PASSWORD", "API_KEY", "SECRET", "PRIVATE_KEY",
				}

				for _, prefix := range sensitivePrefix {
					if strings.HasPrefix(varName, prefix) &&
						(len(varName) == len(prefix) || varName[len(prefix)] == '_') {
						assert.Fail(t, fmt.Sprintf("Sensitive variable was not filtered: %s", envVar))
					}
				}
			}
		}
	})
}

// FuzzFilterEnvironmentMultiple tests filtering multiple environment variables
func FuzzFilterEnvironmentMultiple(f *testing.F) {
	executor := NewSecureExecutor()

	// Add seed corpus with multiple variables
	testcases := []string{
		"PATH=/usr/bin",
		"HOME=/home/user",
		"USER=testuser",
		"AWS_SECRET_ACCESS_KEY=secret",
		"GITHUB_TOKEN=token",
		"API_KEY=key",
		"SECRET=value",
		"MY_SECRET=value",
		"SECRET_KEY=value",
		"NORMAL_VAR=value",
		"",
		"MALFORMED",
		"=NO_KEY",
		"NO_VALUE=",
	}

	for i := 0; i < len(testcases); i += 3 {
		env1 := testcases[i]
		env2 := ""
		env3 := ""
		if i+1 < len(testcases) {
			env2 = testcases[i+1]
		}
		if i+2 < len(testcases) {
			env3 = testcases[i+2]
		}
		f.Add(env1, env2, env3)
	}

	f.Fuzz(func(t *testing.T, env1, env2, env3 string) {
		// Create environment with three variables
		env := []string{env1, env2, env3}

		// Should not panic
		filtered := executor.filterEnvironment(env)

		// Filtered should not have more items than original
		assert.LessOrEqual(t, len(filtered), len(env),
			"Filtered environment has more items than original")

		// Each filtered item should be from the original
		for _, item := range filtered {
			assert.Contains(t, env, item,
				"Filtered environment contains item not in original: %s", item)
		}

		// Check that filtering doesn't introduce new duplicates
		// Count occurrences in original
		originalCount := make(map[string]int)
		for _, item := range env {
			originalCount[item]++
		}

		// Count occurrences in filtered
		filteredCount := make(map[string]int)
		for _, item := range filtered {
			filteredCount[item]++
		}

		// Filtered count should not exceed original count for any item
		for item, count := range filteredCount {
			assert.LessOrEqual(t, count, originalCount[item],
				"Filtered environment has more occurrences of '%s' than original", item)
		}
	})
}
