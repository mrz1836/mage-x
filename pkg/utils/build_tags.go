package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// BuildTagsDiscovery handles discovery of build tags in Go files
type BuildTagsDiscovery struct {
	rootPath    string
	excludeList map[string]bool
}

// NewBuildTagsDiscovery creates a new build tags discovery instance
func NewBuildTagsDiscovery(rootPath string, excludeTags []string) *BuildTagsDiscovery {
	excludeMap := make(map[string]bool)
	for _, tag := range excludeTags {
		excludeMap[strings.TrimSpace(tag)] = true
	}

	return &BuildTagsDiscovery{
		rootPath:    rootPath,
		excludeList: excludeMap,
	}
}

// DiscoverBuildTags scans for build tags in Go files and returns unique tags
func (d *BuildTagsDiscovery) DiscoverBuildTags() ([]string, error) {
	tagSet := make(map[string]bool)

	err := filepath.Walk(d.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// Skip vendor and node_modules directories
		if strings.Contains(path, "vendor/") || strings.Contains(path, "node_modules/") {
			return nil
		}

		// Extract tags from this file
		tags, err := d.extractBuildTagsFromFile(path)
		if err != nil {
			Debug("Failed to extract build tags from %s: %v", path, err)
			return nil // Continue processing other files
		}

		// Add unique tags to set
		for _, tag := range tags {
			if !d.excludeList[tag] && tag != "" {
				tagSet[tag] = true
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert set to sorted slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return tags, nil
}

// extractBuildTagsFromFile extracts build tags from a single Go file
func (d *BuildTagsDiscovery) extractBuildTagsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath) // #nosec G304 -- filepath is controlled by walk function
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			Debug("Failed to close file %s: %v", filePath, closeErr)
		}
	}()

	var tags []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Stop scanning after first non-comment, non-blank line (package declaration)
		if lineNum > 20 || (line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*")) {
			break
		}

		// Extract tags from go:build directives
		if strings.HasPrefix(line, "//go:build ") {
			expr := strings.TrimPrefix(line, "//go:build ")
			extractedTags := d.parseBuildExpression(expr)
			tags = append(tags, extractedTags...)
		}

		// Extract tags from legacy +build directives
		if strings.HasPrefix(line, "// +build ") {
			expr := strings.TrimPrefix(line, "// +build ")
			extractedTags := d.parseLegacyBuildExpression(expr)
			tags = append(tags, extractedTags...)
		}
	}

	return d.removeDuplicates(tags), scanner.Err()
}

// parseBuildExpression parses modern //go:build expressions
func (d *BuildTagsDiscovery) parseBuildExpression(expr string) []string {
	// Remove logical operators and parentheses, extract identifiers
	re := regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9_]*\b`)
	matches := re.FindAllString(expr, -1)

	tags := make([]string, 0, len(matches))
	for _, match := range matches {
		// Skip logical operators
		if match == "and" || match == "or" || match == "not" {
			continue
		}
		// Skip Go version tags (go1, go2, etc) as they're version constraints, not build tags
		if strings.HasPrefix(match, "go") && len(match) >= 3 && match[2] >= '0' && match[2] <= '9' {
			continue
		}
		tags = append(tags, match)
	}

	// Ensure we return an empty slice instead of nil for empty results
	if tags == nil {
		return []string{}
	}
	return tags
}

// parseLegacyBuildExpression parses legacy // +build expressions
func (d *BuildTagsDiscovery) parseLegacyBuildExpression(expr string) []string {
	var tags []string
	parts := strings.Fields(expr)

	for _, part := range parts {
		// In legacy build constraints, commas indicate AND relationships within a term
		// Split by comma to handle cases like "integration,!windows"
		commaParts := strings.Split(part, ",")

		for _, commaPart := range commaParts {
			// Remove all leading negation prefixes (handle multiple ! like !! or !!!)
			tag := strings.TrimLeft(commaPart, "!")
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Ensure we return an empty slice instead of nil for empty results
	if tags == nil {
		return []string{}
	}
	return tags
}

// removeDuplicates removes duplicate tags from slice
func (d *BuildTagsDiscovery) removeDuplicates(tags []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}

// DiscoverBuildTags is a convenience function that creates a discovery instance and runs it
func DiscoverBuildTags(rootPath string, excludeTags []string) ([]string, error) {
	discovery := NewBuildTagsDiscovery(rootPath, excludeTags)
	return discovery.DiscoverBuildTags()
}

// DiscoverBuildTagsFromCurrentDir discovers build tags from the current directory
func DiscoverBuildTagsFromCurrentDir(excludeTags []string) ([]string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return DiscoverBuildTags(currentDir, excludeTags)
}
