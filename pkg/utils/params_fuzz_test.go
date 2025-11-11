package utils

import (
	"strings"
	"testing"
)

// FuzzParseParams tests the ParseParams function for correct parsing behavior
// with various edge cases and malformed inputs.
//
// This fuzzer focuses on:
// - Key=value parsing with edge cases
// - Boolean flag parsing
// - Empty strings and whitespace handling
// - Multiple = signs in values
// - Special characters and Unicode
// - Very long strings
func FuzzParseParams(f *testing.F) {
	// Seed corpus with known patterns and edge cases
	testCases := [][]string{
		// Normal cases
		{"key=value"},
		{"flag"},
		{"key=value", "flag"},
		{"a=1", "b=2", "c=3"},

		// Empty and whitespace
		{""},
		{" "},
		{"\t"},
		{"\n"},
		{"", ""},
		{" ", " "},

		// Keys with whitespace
		{" key=value"},
		{"key =value"},
		{" key =value"},
		{"key= value"},
		{"key=value "},
		{" key = value "},

		// Empty keys or values
		{"=value"},
		{"key="},
		{"="},
		{"=="},
		{"==="},

		// Multiple = signs
		{"key=value=extra"},
		{"key=a=b=c"},
		{"url=https://example.com"},
		{"equation=x=y+z"},

		// Boolean flags with whitespace
		{" flag"},
		{"flag "},
		{" flag "},
		{"\tflag"},
		{"\nflag"},

		// Special characters in keys
		{"key-with-dash=value"},
		{"key_with_underscore=value"},
		{"key.with.dot=value"},
		{"key:with:colon=value"},
		{"key@with@at=value"},

		// Special characters in values
		{"key=value-with-dash"},
		{"key=value_with_underscore"},
		{"key=value.with.dot"},
		{"key=value:with:colon"},
		{"key=value with spaces"},

		// Unicode
		{"key=日本語"}, //nolint:gosmopolitan // Testing Unicode values
		{"clé=valeur"},
		{"ключ=значение"},
		{"🔑=🎯"},

		// Very long strings
		{strings.Repeat("a", 1000) + "=value"},
		{"key=" + strings.Repeat("v", 1000)},
		{strings.Repeat("flag", 100)},

		// Mixed valid and edge cases
		{"valid=1", "", "flag", " ", "key=value"},
		{"", "a=b", "", "c=d", ""},

		// Null bytes
		{"key\x00=value"},
		{"key=value\x00"},
		{"\x00"},

		// Control characters
		{"key\r=value"},
		{"key\n=value"},
		{"key\t=value"},

		// Quote characters (not parsed as quotes, just strings)
		{"key=\"value\""},
		{"key='value'"},
		{"\"key\"=value"},
		{"'key'=value"},

		// Numeric keys and values
		{"123=456"},
		{"0=0"},
		{"-1=1"},

		// Only whitespace between =
		{"key= =value"},
		{"key=\t\n"},

		// Multiple arguments with edge cases
		{"a=1", "b=2", "c", "d=4", "", "e"},
	}

	// Convert test cases to individual arguments for fuzzer
	// We'll serialize the string slice as a single string with a delimiter
	for _, tc := range testCases {
		// Join with a delimiter that's unlikely to appear naturally
		joined := strings.Join(tc, "\x1E") // ASCII Record Separator
		f.Add(joined)
	}

	// Fuzzing function
	f.Fuzz(func(t *testing.T, argsStr string) {
		// Split back into slice
		var args []string
		if argsStr == "" {
			args = []string{}
		} else {
			args = strings.Split(argsStr, "\x1E")
		}

		// Call the function under test
		result := ParseParams(args)

		// Invariants that must always hold:

		// 1. Function should never panic
		// (implicitly tested by fuzzer)

		// 2. Result should never be nil
		if result == nil {
			t.Errorf("ParseParams returned nil map")
			return
		}

		// 3. All keys in result should be non-empty after trimming
		for key := range result {
			if strings.TrimSpace(key) == "" {
				t.Errorf("Result contains empty key: key=%q", key)
			}
		}

		// 4. Validate parsing logic - build expected result (last value wins for duplicates)
		expected := make(map[string]string)
		for _, arg := range args {
			trimmedArg := strings.TrimSpace(arg)
			if trimmedArg == "" {
				continue
			}

			if strings.Contains(arg, "=") {
				// Should parse as key=value
				parts := strings.SplitN(arg, "=", 2)
				key := strings.TrimSpace(parts[0])
				expectedValue := strings.TrimSpace(parts[1])
				if key != "" {
					expected[key] = expectedValue
				}
			} else {
				// Should parse as boolean flag
				key := strings.TrimSpace(arg)
				if key != "" {
					expected[key] = TrueValue
				}
			}
		}

		// Compare expected with actual result
		if len(result) != len(expected) {
			t.Errorf("Result size mismatch: expected=%d, actual=%d, expected=%v, result=%v",
				len(expected), len(result), expected, result)
		}

		for key, expectedValue := range expected {
			if actualValue, exists := result[key]; !exists {
				t.Errorf("Key not found in result: key=%q, expected=%q", key, expectedValue)
			} else if actualValue != expectedValue {
				t.Errorf("Value mismatch: key=%q, expected=%q, actual=%q", key, expectedValue, actualValue)
			}
		}

		for key, actualValue := range result {
			if expectedValue, exists := expected[key]; !exists {
				t.Errorf("Unexpected key in result: key=%q, actual=%q", key, actualValue)
			} else if actualValue != expectedValue {
				t.Errorf("Value mismatch: key=%q, expected=%q, actual=%q", key, expectedValue, actualValue)
			}
		}

		// 5. Result size should not exceed number of valid args
		// (some args might have empty keys after trimming, which are skipped)
		validArgCount := 0
		seenKeys := make(map[string]bool)
		for _, arg := range args {
			var key string
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				key = strings.TrimSpace(parts[0])
			} else {
				key = strings.TrimSpace(arg)
			}
			if key != "" && !seenKeys[key] {
				seenKeys[key] = true
				validArgCount++
			}
		}

		if len(result) > validArgCount {
			t.Errorf("Result has more keys than expected: args=%v, result=%v, expected_max=%d, actual=%d",
				args, result, validArgCount, len(result))
		}
	})
}
