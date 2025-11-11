package utils

import (
	"errors"
	"strings"
	"testing"
)

// FuzzParsePlatform tests the ParsePlatform function for correct parsing of
// platform strings in "os/arch" format with various edge cases.
//
// This fuzzer focuses on:
// - Valid os/arch combinations
// - Missing or extra slashes
// - Empty OS or arch components
// - Special characters and Unicode
// - Very long strings
// - Error handling consistency
func FuzzParsePlatform(f *testing.F) {
	// Seed corpus with known patterns and edge cases
	testCases := []string{
		// Valid cases
		"linux/amd64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
		"linux/386",
		"linux/arm",
		"linux/arm64",
		"freebsd/amd64",
		"openbsd/amd64",
		"netbsd/amd64",

		// Edge cases with slashes
		"/",
		"//",
		"///",
		"linux/",
		"/amd64",
		"/linux/amd64",
		"linux/amd64/",
		"linux/amd64/extra",
		"linux//amd64",
		"linux///amd64",

		// Empty components
		"",
		" ",
		"/ ",
		" /",
		" / ",
		"/amd64",
		"linux/",

		// No slash
		"linux",
		"amd64",
		"linuxamd64",

		// Multiple platforms (invalid)
		"linux/amd64/darwin/arm64",
		"linux/amd64/windows/386",

		// Special characters
		"linux-gnu/amd64",
		"linux.gnu/amd64",
		"linux_gnu/amd64",
		"linux:gnu/amd64",
		"linux@gnu/amd64",
		"linux$gnu/amd64",

		// Unicode
		"日本語/amd64", //nolint:gosmopolitan // Testing Unicode OS names
		"linux/日本語", //nolint:gosmopolitan // Testing Unicode arch names
		"тэг/тэг",
		"🐧/💻",

		// Case variations
		"Linux/AMD64",
		"LINUX/AMDG64",
		"LiNuX/AmD64",

		// Numbers
		"386/amd64",
		"linux/386",
		"123/456",

		// Very long strings
		strings.Repeat("a", 1000) + "/amd64",
		"linux/" + strings.Repeat("a", 1000),
		strings.Repeat("linux/amd64/", 100),

		// Whitespace
		" linux/amd64",
		"linux/amd64 ",
		" linux/amd64 ",
		"linux / amd64",
		"linux/ amd64",
		"linux /amd64",
		"\tlinux/amd64",
		"linux/amd64\t",
		"linux\t/amd64",
		"linux/\tamd64",

		// Control characters
		"linux\n/amd64",
		"linux/\namd64",
		"linux\r/amd64",
		"linux/\ramd64",
		"\nlinux/amd64",
		"linux/amd64\n",

		// Null bytes
		"linux\x00/amd64",
		"linux/\x00amd64",
		"\x00linux/amd64",
		"linux/amd64\x00",

		// Quote characters
		"\"linux\"/amd64",
		"linux/\"amd64\"",
		"'linux'/amd64",
		"linux/'amd64'",

		// Backslashes (Windows-style)
		"linux\\amd64",
		"linux\\/amd64",
		"linux/\\amd64",

		// Dots
		"./amd64",
		"linux/.",
		"../amd64",
		"linux/..",
		"./linux/amd64",
		"linux/amd64/.",

		// Dashes and underscores (common in real OS/arch names)
		"linux-gnu/amd64",
		"linux_gnu/amd64",
		"linux/amd64-v1",
		"linux/amd64_v1",

		// Mixed valid and invalid
		"linux/amd64/extra/parts",
		"/linux/amd64/",
		"//linux//amd64//",

		// Only slashes with whitespace
		" / / ",
		"/ / /",

		// Real-world-like values
		"js/wasm",
		"plan9/amd64",
		"solaris/amd64",
		"android/arm64",
		"ios/arm64",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, s string) {
		// Call the function under test
		platform, err := ParsePlatform(s)

		// Invariants that must always hold:

		// 1. Function should never panic
		// (implicitly tested by fuzzer)

		// 2. Valid input should have exactly one slash
		slashCount := strings.Count(s, "/")

		if slashCount == 1 { //nolint:nestif // Fuzz test validation logic requires nesting
			// Should succeed if both parts are non-empty
			parts := strings.Split(s, "/")
			if parts[0] != "" && parts[1] != "" {
				// Should parse successfully
				if err != nil {
					t.Errorf("Valid input returned error: s=%q, err=%v", s, err)
				}
				// Platform should match input
				if platform.OS != parts[0] || platform.Arch != parts[1] {
					t.Errorf("Platform mismatch: s=%q, expected OS=%q Arch=%q, got OS=%q Arch=%q",
						s, parts[0], parts[1], platform.OS, platform.Arch)
				}
			} else {
				// Empty OS or Arch should return error
				if err == nil {
					t.Errorf("Empty component should return error: s=%q, platform=%v", s, platform)
				}
			}
		} else {
			// Wrong number of slashes should return error
			if err == nil {
				t.Errorf("Invalid format should return error: s=%q, slashCount=%d, platform=%v",
					s, slashCount, platform)
			}
		}

		// 3. Error should be errInvalidPlatformFormat or wrapped version
		if err != nil {
			if !errors.Is(err, errInvalidPlatformFormat) {
				t.Errorf("Error is not errInvalidPlatformFormat: s=%q, err=%v", s, err)
			}
		}

		// 4. If successful, OS and Arch should be non-empty
		if err == nil {
			if platform.OS == "" {
				t.Errorf("Success with empty OS: s=%q, platform=%v", s, platform)
			}
			if platform.Arch == "" {
				t.Errorf("Success with empty Arch: s=%q, platform=%v", s, platform)
			}
		}

		// 5. If error, Platform should be zero value
		if err != nil {
			if platform.OS != "" || platform.Arch != "" {
				t.Errorf("Error should return zero value Platform: s=%q, platform=%v", s, platform)
			}
		}

		// 6. Validate the parsing logic manually
		parts := strings.Split(s, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" { //nolint:nestif // Fuzz test validation requires nesting
			// Should succeed
			if err != nil {
				t.Errorf("Valid two-part input returned error: s=%q, parts=%v, err=%v", s, parts, err)
			} else {
				// Check that platform matches
				if platform.OS != parts[0] {
					t.Errorf("OS mismatch: s=%q, expected=%q, actual=%q", s, parts[0], platform.OS)
				}
				if platform.Arch != parts[1] {
					t.Errorf("Arch mismatch: s=%q, expected=%q, actual=%q", s, parts[1], platform.Arch)
				}
			}
		} else {
			// Should fail
			if err == nil {
				t.Errorf("Invalid input succeeded: s=%q, parts=%v, platform=%v", s, parts, platform)
			}
		}

		// 7. Multiple slashes or no slashes should always fail
		if slashCount == 0 || slashCount > 1 {
			if err == nil {
				t.Errorf("Wrong slash count should fail: s=%q, slashCount=%d", s, slashCount)
			}
		}

		// 8. Empty string should fail
		if s == "" && err == nil {
			t.Errorf("Empty string should return error")
		}

		// 9. String with only slash should fail
		if s == "/" && err == nil {
			t.Errorf("Single slash should return error")
		}

		// 10. Consistency check: parse should be deterministic
		platform2, err2 := ParsePlatform(s)
		if (err == nil) != (err2 == nil) {
			t.Errorf("Inconsistent error return: s=%q, err1=%v, err2=%v", s, err, err2)
		}
		if err == nil && platform != platform2 {
			t.Errorf("Inconsistent platform return: s=%q, platform1=%v, platform2=%v", s, platform, platform2)
		}
	})
}
