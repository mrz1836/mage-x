package utils

import (
	"regexp"
	"strings"
	"testing"
)

// FuzzParseBuildExpression tests the parseBuildExpression method for correct parsing
// of modern //go:build expressions with various edge cases.
//
// This fuzzer focuses on:
// - Boolean expressions with and/or/not operators
// - Nested parentheses
// - Go version constraints (should be filtered)
// - Special characters and malformed expressions
// - Unicode and very long expressions
func FuzzParseBuildExpression(f *testing.F) {
	// Seed corpus with known patterns and edge cases
	testCases := []string{
		// Normal cases
		"linux",
		"darwin && amd64",
		"linux || darwin",
		"!windows",
		"(linux || darwin) && amd64",
		"linux && (amd64 || arm64)",

		// Go version constraints (should be filtered out)
		"go1.20",
		"go1.21 && linux",
		"linux && go1.20",
		"go1.19 || go1.20",

		// Logical operators (should be filtered out)
		"and",
		"or",
		"not",
		"and && or || not",

		// Complex nested expressions
		"((linux || darwin) && amd64) || windows",
		"(((linux)))",
		"((linux && amd64) || (darwin && arm64))",

		// Multiple tags
		"linux && amd64 && cgo",
		"tag1 || tag2 || tag3 || tag4",
		"!tag1 && !tag2 && !tag3",

		// Edge cases with whitespace
		"  linux  ",
		"linux  &&  amd64",
		"( linux || darwin )",

		// Empty and minimal
		"",
		" ",
		"()",
		"(())",
		"( )",

		// Single characters
		"a",
		"x",
		"z9",

		// Numbers and underscores
		"tag_1",
		"tag123",
		"_tag",
		"tag_",

		// Multiple parentheses
		"((((linux))))",
		"(linux) || (darwin)",
		"((linux) && (amd64))",

		// Mixed operators
		"linux && !windows || darwin",
		"!linux && !windows && !darwin",
		"(linux && !cgo) || windows",

		// Special characters (should not break regex)
		"tag-with-dash",
		"tag.with.dot",
		"tag:with:colon",
		"tag@with@at",
		"tag$with$dollar",

		// Very long expressions
		strings.Repeat("linux && ", 50) + "amd64",
		strings.Repeat("(", 20) + "linux" + strings.Repeat(")", 20),

		// Unicode
		"日本語", //nolint:gosmopolitan // Testing Unicode handling
		"тэг",
		"🏷️",

		// Malformed expressions
		"&& linux",
		"linux &&",
		"|| darwin",
		"darwin ||",
		"linux && && amd64",
		"linux || || darwin",

		// Tags with operators in them (edge case)
		"android",  // contains 'and'
		"landor",   // contains 'and' and 'or'
		"ornament", // starts with 'or'
		"nothing",  // contains 'not'

		// Case variations
		"Linux",
		"LINUX",
		"LiNuX",

		// Numeric patterns
		"386",
		"go2",
		"go999",

		// Multiple spaces
		"linux    &&    amd64",
		"linux\t&&\tamd64",
		"linux\n&&\namd64",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, expr string) {
		// Create a BuildTagsDiscovery instance
		d := NewBuildTagsDiscovery(".", nil)

		// Call the method under test
		result := d.parseBuildExpression(expr)

		// Invariants that must always hold:

		// 1. Function should never panic
		// (implicitly tested by fuzzer)

		// 2. Result should never be nil (function ensures empty slice instead of nil)
		if result == nil {
			t.Errorf("parseBuildExpression returned nil: expr=%q", expr)
			return
		}

		// 3. Result should not contain logical operators
		for _, tag := range result {
			if tag == "and" || tag == "or" || tag == "not" {
				t.Errorf("Result contains logical operator: expr=%q, tag=%q", expr, tag)
			}
		}

		// 4. Result should not contain Go version tags
		for _, tag := range result {
			if strings.HasPrefix(tag, "go") && len(tag) >= 3 && tag[2] >= '0' && tag[2] <= '9' {
				t.Errorf("Result contains Go version tag: expr=%q, tag=%q", expr, tag)
			}
		}

		// 5. All tags should match the identifier pattern
		identifierRe := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
		for _, tag := range result {
			if !identifierRe.MatchString(tag) {
				t.Errorf("Result contains invalid identifier: expr=%q, tag=%q", expr, tag)
			}
		}

		// 6. Verify the extraction logic matches the implementation
		// Extract identifiers using the same regex as the implementation
		extractRe := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9_]*\b`)
		matches := extractRe.FindAllString(expr, -1)

		// Filter to expected tags (excluding operators and Go versions)
		var expectedTags []string
		for _, match := range matches {
			if match == "and" || match == "or" || match == "not" {
				continue
			}
			if strings.HasPrefix(match, "go") && len(match) >= 3 && match[2] >= '0' && match[2] <= '9' {
				continue
			}
			expectedTags = append(expectedTags, match)
		}

		// Result should match expected tags
		if len(result) != len(expectedTags) {
			t.Errorf("Result length mismatch: expr=%q, expected=%d, actual=%d, expectedTags=%v, result=%v",
				expr, len(expectedTags), len(result), expectedTags, result)
		}

		// Check that all expected tags are in result (order may vary)
		expectedSet := make(map[string]bool)
		for _, tag := range expectedTags {
			expectedSet[tag] = true
		}
		resultSet := make(map[string]bool)
		for _, tag := range result {
			resultSet[tag] = true
		}

		for tag := range expectedSet {
			if !resultSet[tag] {
				t.Errorf("Expected tag not in result: expr=%q, tag=%q, result=%v", expr, tag, result)
			}
		}
		for tag := range resultSet {
			if !expectedSet[tag] {
				t.Errorf("Unexpected tag in result: expr=%q, tag=%q, expected=%v", expr, tag, expectedTags)
			}
		}
	})
}

// FuzzParseLegacyBuildExpression tests the parseLegacyBuildExpression method for correct
// parsing of legacy // +build expressions with various edge cases.
//
// This fuzzer focuses on:
// - Space-separated OR conditions
// - Comma-separated AND conditions
// - Negation prefixes (!)
// - Empty tags and whitespace handling
// - Special characters and malformed expressions
func FuzzParseLegacyBuildExpression(f *testing.F) {
	// Seed corpus with known patterns and edge cases
	testCases := []string{
		// Normal cases
		"linux",
		"linux darwin",
		"linux,amd64",
		"!windows",
		"linux,!cgo",
		"linux darwin windows",

		// Mixed AND (comma) and OR (space)
		"linux,amd64 darwin,arm64",
		"!windows,amd64 !darwin,amd64",
		"integration,!windows integration,!darwin",

		// Multiple negations
		"!tag1",
		"!tag1 !tag2",
		"!tag1,!tag2",
		"!tag1,!tag2,!tag3",

		// Edge cases with whitespace
		"  linux  ",
		"linux  darwin",
		"linux ,amd64",
		"linux, amd64",
		"linux , amd64",

		// Empty and minimal
		"",
		" ",
		",",
		",,",
		"  ,  ",

		// Single characters
		"a",
		"x",
		"z9",

		// Commas without spaces
		"a,b,c,d",
		"!a,!b,!c",

		// Spaces without commas
		"a b c d",
		"!a !b !c",

		// Mixed patterns
		"a,b c,d",
		"!a,b !c,d",
		"a,!b c,!d",

		// Leading/trailing delimiters
		",linux",
		"linux,",
		" linux",
		"linux ",
		"!,linux",
		",!linux",

		// Multiple consecutive delimiters
		"linux,,amd64",
		"linux  darwin",
		"linux , , amd64",

		// Only delimiters
		"!",
		"!!",
		"!!!",
		"! !",

		// Negation patterns
		"!",
		"! tag",
		"tag !",
		"!!tag",
		"!!!tag",

		// Empty tags after split
		"linux,,,,amd64",
		"linux    darwin",

		// Numbers and underscores
		"tag_1",
		"tag123",
		"_tag",
		"tag_",
		"386",

		// Very long expressions
		strings.Repeat("tag ", 100),
		strings.Repeat("tag,", 100),
		strings.Repeat("!tag ", 50),

		// Unicode
		"日本語", //nolint:gosmopolitan // Testing Unicode handling
		"тэг",
		"🏷️",

		// Special characters
		"tag-with-dash",
		"tag.with.dot",
		"tag:with:colon",

		// Tab characters
		"linux\tdarwin",
		"linux\t\tdarwin",
		"linux,\tamd64",

		// Newlines (shouldn't appear but test anyway)
		"linux\ndarwin",
		"linux,\namd64",

		// Complex real-world examples
		"integration,!windows",
		"linux,386 linux,amd64 darwin,amd64",
		"!race,!cgo",

		// Case variations
		"Linux",
		"LINUX",
		"LiNuX",

		// Mixed case negations
		"!Linux",
		"!LINUX",
	}

	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, expr string) {
		// Create a BuildTagsDiscovery instance
		d := NewBuildTagsDiscovery(".", nil)

		// Call the method under test
		result := d.parseLegacyBuildExpression(expr)

		// Invariants that must always hold:

		// 1. Function should never panic
		// (implicitly tested by fuzzer)

		// 2. Result should never be nil
		if result == nil {
			t.Errorf("parseLegacyBuildExpression returned nil: expr=%q", expr)
			return
		}

		// 3. Result should not contain empty strings
		for _, tag := range result {
			if tag == "" {
				t.Errorf("Result contains empty tag: expr=%q, result=%v", expr, result)
			}
		}

		// 4. Result should not contain negation prefixes
		for _, tag := range result {
			if strings.HasPrefix(tag, "!") {
				t.Errorf("Result contains tag with negation prefix: expr=%q, tag=%q", expr, tag)
			}
		}

		// 5. Verify the parsing logic
		// Split by spaces (OR conditions)
		parts := strings.Fields(expr)
		var expectedTags []string

		for _, part := range parts {
			// Split by commas (AND conditions)
			commaParts := strings.Split(part, ",")
			for _, commaPart := range commaParts {
				// Remove all leading negation prefixes (matches implementation)
				tag := strings.TrimLeft(commaPart, "!")
				if tag != "" {
					expectedTags = append(expectedTags, tag)
				}
			}
		}

		// Check that result matches expected (order may vary, but length should match)
		if len(result) != len(expectedTags) {
			t.Errorf("Result length mismatch: expr=%q, expected=%d, actual=%d, expectedTags=%v, result=%v",
				expr, len(expectedTags), len(result), expectedTags, result)
		}

		// Check that all tags are present (use sets for comparison)
		expectedSet := make(map[string]bool)
		for _, tag := range expectedTags {
			expectedSet[tag] = true
		}
		resultSet := make(map[string]bool)
		for _, tag := range result {
			resultSet[tag] = true
		}

		for tag := range expectedSet {
			if !resultSet[tag] {
				t.Errorf("Expected tag not in result: expr=%q, tag=%q, result=%v", expr, tag, result)
			}
		}
		for tag := range resultSet {
			if !expectedSet[tag] {
				t.Errorf("Unexpected tag in result: expr=%q, tag=%q, expected=%v", expr, tag, expectedTags)
			}
		}

		// 6. Special case: if expression is only delimiters/whitespace, result should be empty
		trimmed := strings.TrimSpace(strings.Trim(expr, ",!"))
		if trimmed == "" && len(result) != 0 {
			t.Errorf("Non-empty result for delimiter-only expression: expr=%q, result=%v", expr, result)
		}
	})
}
