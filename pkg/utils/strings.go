package utils

import (
	"strings"
)

// ParseNonEmptyLines parses command output into non-empty lines.
// It trims leading/trailing whitespace from the output, splits by newlines,
// and filters out empty lines. This is commonly used to process command output.
//
// Example:
//
//	output := "file1.go\nfile2.go\n\nfile3.go\n"
//	files := utils.ParseNonEmptyLines(output)
//	// files = []string{"file1.go", "file2.go", "file3.go"}
func ParseNonEmptyLines(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
