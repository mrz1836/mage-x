//go:build integration
// +build integration

package integration

import (
	"fmt"
	"strings"
)

// countTreeDepth accurately counts the maximum depth in a dependency tree output
// Returns the maximum depth found in the tree structure
func countTreeDepth(output string) int {
	lines := strings.Split(output, "\n")
	maxDepth := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Count the depth by analyzing the tree structure
		// Root has no indentation (depth 0)
		// Level 1 has "├──" or "└──" with some spacing (depth 1)
		// Level 2 has "│   ├──" or "│   └──" (depth 2)
		// Each "│   " pattern indicates one level of nesting

		if strings.Contains(line, "├──") || strings.Contains(line, "└──") {
			depth := 0

			// Convert to runes to handle Unicode characters properly
			runes := []rune(line)
			i := 0

			// Count vertical pipes that indicate nesting levels
			for i < len(runes) {
				// Look for "│   " pattern (pipe followed by spaces)
				if i+4 <= len(runes) && string(runes[i:i+4]) == "│   " {
					depth++
					i += 4
				} else if i+3 <= len(runes) && string(runes[i:i+3]) == "│  " {
					// Handle slight variations in spacing
					depth++
					i += 3
				} else if runes[i] == '├' || runes[i] == '└' {
					// Found the branch symbol, this line represents depth+1
					depth++
					break
				} else if runes[i] == ' ' || runes[i] == '\t' {
					// Skip whitespace
					i++
				} else {
					// Hit content, break
					break
				}
			}

			if depth > maxDepth {
				maxDepth = depth
			}
		} else if strings.Contains(line, "testmodule") || strings.Contains(line, "Dependency Tree:") {
			// Root module or header - depth 0
			// Don't update maxDepth as this could be a header
		}
	}

	return maxDepth
}

// verifyDepthLimit checks that the tree output respects the specified depth limit
func verifyDepthLimit(output string, expectedMaxDepth int) bool {
	actualMaxDepth := countTreeDepth(output)
	return actualMaxDepth <= expectedMaxDepth
}

// hasTreeSymbols checks if the output contains dependency tree symbols
func hasTreeSymbols(output string) bool {
	return strings.Contains(output, "├──") ||
		strings.Contains(output, "└──") ||
		strings.Contains(output, "│")
}

// countDependencyLines counts how many lines contain actual dependencies (not headers)
func countDependencyLines(output string) int {
	lines := strings.Split(output, "\n")
	count := 0

	for _, line := range lines {
		// Count lines that contain dependency names but aren't headers
		if (strings.Contains(line, "├──") || strings.Contains(line, "└──")) &&
			!strings.Contains(line, "Dependency Tree:") &&
			!strings.Contains(line, "testmodule") {
			count++
		}
	}

	return count
}

// extractDependencyNames extracts all dependency names from the tree output
func extractDependencyNames(output string) []string {
	lines := strings.Split(output, "\n")
	var deps []string

	for _, line := range lines {
		if strings.Contains(line, "├──") || strings.Contains(line, "└──") {
			// Extract the part after the tree symbols
			parts := strings.Fields(line)
			for _, part := range parts {
				// Look for package names (containing dots or slashes)
				if strings.Contains(part, ".") || strings.Contains(part, "/") {
					// Clean up version info if present
					if idx := strings.Index(part, "@"); idx != -1 {
						part = part[:idx]
					}
					deps = append(deps, part)
					break
				}
			}
		}
	}

	return deps
}

// compareTreeOutputs compares two tree outputs and returns analysis
func compareTreeOutputs(output1, output2 string) (deeper, moreLines bool, analysis string) {
	depth1 := countTreeDepth(output1)
	depth2 := countTreeDepth(output2)

	lines1 := len(strings.Split(output1, "\n"))
	lines2 := len(strings.Split(output2, "\n"))

	deps1 := countDependencyLines(output1)
	deps2 := countDependencyLines(output2)

	deeper = depth2 > depth1
	moreLines = lines2 > lines1

	analysis = fmt.Sprintf("Comparison:\n  Output1: depth=%d, lines=%d, deps=%d\n  Output2: depth=%d, lines=%d, deps=%d",
		depth1, lines1, deps1, depth2, lines2, deps2)

	return deeper, moreLines, analysis
}
