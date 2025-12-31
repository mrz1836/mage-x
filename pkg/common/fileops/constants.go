// Package fileops provides file system operations and constants.
package fileops

// File extension constants for config file detection.
// These should be used instead of hardcoded string literals.
const (
	// ExtYAML is the .yaml file extension
	ExtYAML = ".yaml"
	// ExtYML is the .yml file extension (common alternative)
	ExtYML = ".yml"
	// ExtJSON is the .json file extension
	ExtJSON = ".json"
)

// Format string constants for config file format specification.
// These match the extensions but without the leading dot.
const (
	// FormatYAML specifies YAML format
	FormatYAML = "yaml"
	// FormatYML specifies YML format (common alternative)
	FormatYML = "yml"
	// FormatJSON specifies JSON format
	FormatJSON = "json"
)
