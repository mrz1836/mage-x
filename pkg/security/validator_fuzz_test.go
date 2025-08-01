//go:build go1.18
// +build go1.18

package security

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FuzzValidateVersion tests version validation with fuzzing
func FuzzValidateVersion(f *testing.F) {
	// Add seed corpus with known edge cases
	testcases := []string{
		"1.0.0",
		"v1.0.0",
		"0.0.0",
		"1.2.3-alpha",
		"1.2.3-beta.1",
		"1.2.3+build.123",
		"1.2.3-rc.1+build.123",
		"999.999.999",
		"v1.0.0-",
		"1.0.0-",
		"1.0.0+",
		"1.0.0-+",
		"1.0.0-...",
		"1.0.0-/",
		"1.0.0-\\",
		"1.0.0-<script>",
		"1.0.0-${IFS}",
		"1.0.0-$(whoami)",
		"1.0.0-`whoami`",
		"1.0.0-;rm -rf /",
		"v1.0.0\x00",
		"1.0.0\n",
		"1.0.0\r\n",
		strings.Repeat("1", 1000) + ".0.0",
		"🚀1.0.0",
		"1.0.0🚀",
		"",
		" ",
		"\t",
		".",
		"..",
		"...",
		"v",
		"vv",
		"v.",
		"v..",
		"v...",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, version string) {
		// The function should not panic on any input
		err := ValidateVersion(version)

		// Check for dangerous patterns first - if present, validation should have failed
		if strings.Contains(version, "..") || strings.Contains(version, "$(") ||
			strings.Contains(version, "`") || strings.Contains(version, "\x00") ||
			strings.Contains(version, "\n") || strings.Contains(version, "\r") {
			assert.Error(t, err, "Version with dangerous patterns should be rejected: %s", version)
			return
		}

		// If no error, verify it's actually a valid version
		if err == nil {
			// Should match our version pattern
			cleaned := strings.TrimPrefix(version, "v")
			assert.Regexp(t, `^\d+\.\d+\.\d+(-[a-zA-Z0-9\-\.]+)?(\+[a-zA-Z0-9\-\.]+)?$`, cleaned,
				"ValidateVersion accepted invalid version: %s", version)
		}

		// If the function rejected invalid UTF-8, that's correct behavior
		if err != nil && strings.Contains(err.Error(), "invalid UTF-8") {
			assert.False(t, utf8.ValidString(version), "Function rejected valid UTF-8")
		}
	})
}

// FuzzValidateGitRef tests git reference validation with fuzzing
func FuzzValidateGitRef(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		"main",
		"master",
		"develop",
		"feature/new-feature",
		"release/1.0.0",
		"hotfix/urgent-fix",
		"v1.0.0",
		"refs/heads/main",
		"refs/tags/v1.0.0",
		"feature-123",
		"feature_123",
		"FEATURE",
		"feature/UPPERCASE",
		"feature/with.dots",
		"",
		" ",
		"..",
		"../../../etc/passwd",
		"feature/../../../etc/passwd",
		"feature/..",
		"~",
		"^",
		":",
		"feature:malicious",
		"feature malicious",
		"feature\ttab",
		"feature\nnewline",
		"feature;rm -rf /",
		"feature&&whoami",
		"feature||whoami",
		"feature`whoami`",
		"feature$(whoami)",
		"feature${IFS}",
		"*",
		"feature*",
		"?",
		"[",
		"feature[",
		"\\",
		"feature\\command",
		strings.Repeat("a", 1000),
		"feature/" + strings.Repeat("b", 255),
		"🚀feature",
		"feature🚀",
		"feature\x00null",
		"\x00",
		"\x1b[31mred\x1b[0m",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, ref string) {
		err := ValidateGitRef(ref)

		if err == nil {
			// Verify it only contains allowed characters
			assert.Regexp(t, `^[a-zA-Z0-9\-_\.\/]+$`, ref,
				"ValidateGitRef accepted ref with invalid characters: %s", ref)

			// Verify no dangerous patterns
			dangerousPatterns := []string{"..", "~", "^", ":", "\\", "*", "?", "[", " "}
			for _, pattern := range dangerousPatterns {
				assert.NotContains(t, ref, pattern,
					"Git ref contains dangerous pattern '%s': %s", pattern, ref)
			}

			// Should not be empty
			assert.NotEmpty(t, ref, "ValidateGitRef accepted empty ref")
		}
	})
}

// FuzzValidateFilename tests filename validation with fuzzing
func FuzzValidateFilename(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		"file.txt",
		"document.pdf",
		"image.png",
		"script.sh",
		"data.json",
		"config.yaml",
		"README.md",
		"test-file.txt",
		"test_file.txt",
		"123.txt",
		"file123.txt",
		"",
		" ",
		"..",
		"../passwd",
		"../../etc/passwd",
		"/etc/passwd",
		"C:\\Windows\\System32\\cmd.exe",
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
		".hidden",
		"..hidden",
		"...hidden",
		strings.Repeat("a", 255) + ".txt",
		strings.Repeat("a", 1000) + ".txt",
		"🚀.txt",
		"file🚀.txt",
		"file.txt",    // Russian file
		"wenjian.txt", // Chinese file
		"~file.txt",
		"$file.txt",
		"#file.txt",
		"@file.txt",
		"!file.txt",
		"%file.txt",
		"^file.txt",
		"*file.txt",
		"(file).txt",
		"[file].txt",
		"{file}.txt",
		"file?.txt",
		"file*.txt",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, filename string) {
		err := ValidateFilename(filename)

		// Check for dangerous patterns - if present, validation should have failed
		if filename == ".." || filename == "." || strings.Contains(filename, "/") ||
			strings.Contains(filename, "\\") || strings.Contains(filename, "\x00") ||
			filename == "" || strings.TrimSpace(filename) != filename {
			assert.Error(t, err, "Filename with dangerous patterns should be rejected: %s", filename)
			return
		}

		if err == nil {
			// Should match safe pattern
			assert.Regexp(t, `^[a-zA-Z0-9\-_\.]+$`, filename,
				"Filename contains unsafe characters: %s", filename)

			// Should not be empty
			assert.NotEmpty(t, filename, "ValidateFilename accepted empty filename")
		}
	})
}

// FuzzValidateURL tests URL validation with fuzzing
func FuzzValidateURL(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		"http://example.com",
		"https://example.com",
		"https://example.com/path",
		"https://example.com/path?query=value",
		"https://example.com:8080/path",
		"https://user:pass@example.com/path",
		"",
		" ",
		"ftp://example.com",
		"file:///etc/passwd",
		"javascript:alert('xss')",
		"data:text/html,<script>alert('xss')</script>",
		"vbscript:msgbox('xss')",
		"about:blank",
		"chrome://settings",
		"https://example.com/<script>alert('xss')</script>",
		"https://example.com/path?param=<script>alert('xss')</script>",
		"https://example.com/path#<script>alert('xss')</script>",
		"https://example.com/path%3Cscript%3Ealert('xss')%3C/script%3E",
		"https://example.com/path?onerror=alert('xss')",
		"https://example.com/path?onload=alert('xss')",
		"HTTP://EXAMPLE.COM",
		"HtTpS://ExAmPlE.cOm",
		"https://" + strings.Repeat("a", 1000) + ".com",
		"https://example.com/" + strings.Repeat("a", 10000),
		"https://🚀.com",
		"https://example.🚀",
		"https://exam ple.com",
		"https://example.com/path with spaces",
		"https://example.com\x00",
		"https://example.com\n",
		"https://example.com\r\n",
		"https://example.com\\path",
		"https://[::1]/path",
		"https://[2001:db8::1]/path",
		"https://127.0.0.1/path",
		"https://192.168.1.1/path",
		"//example.com/path",
		"///etc/passwd",
		"\\\\server\\share",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, url string) {
		err := ValidateURL(url)

		// Check for suspicious patterns - if present, validation should have failed
		lowerURL := strings.ToLower(url)
		suspiciousPatterns := []string{
			"javascript:", "data:", "vbscript:", "file:", "about:", "chrome:",
			"<script", "%3cscript", "onerror=", "onload=",
		}
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(lowerURL, pattern) {
				assert.Error(t, err, "URL with suspicious pattern '%s' should be rejected: %s", pattern, url)
				return
			}
		}

		// Check for other dangerous patterns
		if url == "" || strings.Contains(url, "\x00") || strings.Contains(url, "\n") ||
			strings.Contains(url, "\r") || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
			assert.Error(t, err, "URL with dangerous patterns should be rejected: %s", url)
			return
		}

		if err == nil {
			// Should start with http:// or https://
			assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"),
				"URL doesn't start with http:// or https://: %s", url)

			// Should not be empty
			assert.NotEmpty(t, url, "ValidateURL accepted empty URL")
		}
	})
}

// FuzzValidateEmail tests email validation with fuzzing
func FuzzValidateEmail(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user_name@example.com",
		"123@example.com",
		"user@subdomain.example.com",
		"user@example.co.uk",
		"",
		" ",
		"@",
		"user@",
		"@example.com",
		"user",
		"user@@example.com",
		"user@example@com",
		"user example@example.com",
		"user@exam ple.com",
		"user@example",
		"user@.com",
		"user@example.",
		"user@.",
		"user@..",
		"user@-example.com",
		"user@example-.com",
		"user@ex_ample.com",
		"user@" + strings.Repeat("a", 255) + ".com",
		strings.Repeat("a", 255) + "@example.com",
		"user@example.com\x00",
		"user@example.com\n",
		"user@example.com\r\n",
		"user\x00@example.com",
		"<script>@example.com",
		"user@<script>.com",
		"user@example.<script>",
		"🚀@example.com",
		"user@🚀.com",
		"user@example.🚀",
		"yonghu@example.com", // Chinese user
		"user@tatoeba.com",   // Japanese example
		"user@example.com;rm -rf /",
		"user@example.com&&whoami",
		"user@example.com||whoami",
		"user@example.com`whoami`",
		"user@example.com$(whoami)",
		"user+filter/../../etc/passwd@example.com",
		"user%2B@example.com",
		"user%40example.com",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, email string) {
		err := ValidateEmail(email)

		if err == nil {
			// Should have exactly one @
			parts := strings.Split(email, "@")
			require.Len(t, parts, 2, "Email should have exactly one @ symbol")

			// Both parts should not be empty
			assert.NotEmpty(t, parts[0], "Email local part is empty")
			assert.NotEmpty(t, parts[1], "Email domain part is empty")

			// Domain should have at least one dot
			assert.Contains(t, parts[1], ".", "Email domain should contain at least one dot")

			// Should not be just whitespace
			assert.NotEmpty(t, strings.TrimSpace(email), "Email is just whitespace")
		}
	})
}

// FuzzValidatePort tests port validation with fuzzing
func FuzzValidatePort(f *testing.F) {
	// Add seed corpus with edge cases
	testcases := []int{
		0, 1, 22, 80, 443, 1023, 1024, 3000, 8080, 8443,
		65534, 65535, 65536, -1, -80, -65535,
		999999, -999999, 1<<31 - 1, -(1 << 31),
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, port int) {
		err := ValidatePort(port)

		if err == nil {
			// Port should be in valid range
			assert.GreaterOrEqual(t, port, 1, "Port should be >= 1")
			assert.LessOrEqual(t, port, 65535, "Port should be <= 65535")
		} else {
			// If error, port should be out of range
			assert.True(t, port < 1 || port > 65535,
				"ValidatePort returned error for valid port %d", port)
		}
	})
}

// FuzzValidateEnvVar tests environment variable name validation with fuzzing
func FuzzValidateEnvVar(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		"PATH",
		"HOME",
		"USER",
		"_VAR",
		"VAR_NAME",
		"VAR123",
		"_",
		"__VAR__",
		"MY_LONG_VARIABLE_NAME",
		"",
		" ",
		"123VAR",
		"VAR-NAME",
		"VAR NAME",
		"VAR=VALUE",
		"VAR;",
		"VAR&",
		"VAR|",
		"VAR>",
		"VAR<",
		"VAR`",
		"VAR$",
		"VAR()",
		"VAR[]",
		"VAR{}",
		"VAR*",
		"VAR?",
		"VAR\x00",
		"VAR\n",
		"VAR\r\n",
		"🚀VAR",
		"VAR🚀",
		"VAR_🚀",
		strings.Repeat("A", 255),
		strings.Repeat("A", 1000),
		"PATH;rm -rf /",
		"$(whoami)",
		"`whoami`",
		"${IFS}",
		"VAR\tTAB",
		"VAR\nNEWLINE",
		"-VAR",
		".VAR",
		"/VAR",
		"\\VAR",
		"@VAR",
		"#VAR",
		"%VAR",
		"^VAR",
		"!VAR",
		"+VAR",
		"=VAR",
		"~VAR",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, name string) {
		err := ValidateEnvVar(name)

		if err == nil {
			// Should not be empty
			assert.NotEmpty(t, name, "ValidateEnvVar accepted empty name")

			// Should start with letter or underscore
			if name != "" {
				firstChar := name[0]
				assert.True(t, (firstChar >= 'a' && firstChar <= 'z') ||
					(firstChar >= 'A' && firstChar <= 'Z') || firstChar == '_',
					"Env var should start with letter or underscore: %s", name)
			}

			// Should only contain alphanumeric and underscore
			assert.Regexp(t, `^[a-zA-Z_][a-zA-Z0-9_]*$`, name,
				"Env var contains invalid characters: %s", name)
		}
	})
}
