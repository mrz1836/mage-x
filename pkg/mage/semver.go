package mage

import (
	"fmt"
	"strconv"
	"strings"
)

// SemanticVersion represents a semantic version (major.minor.patch).
// It provides parsing, comparison, and manipulation methods for version strings.
type SemanticVersion struct {
	major int
	minor int
	patch int
}

// ParseSemanticVersion parses a version string into a SemanticVersion.
// It accepts versions with or without the "v" prefix (e.g., "v1.2.3" or "1.2.3").
// Returns an error if the version string is malformed.
func ParseSemanticVersion(version string) (SemanticVersion, error) {
	clean := strings.TrimPrefix(version, "v")
	parts := strings.Split(clean, ".")

	if len(parts) != 3 {
		return SemanticVersion{}, fmt.Errorf("%w: %s (expected X.Y.Z format)", errInvalidVersionFormat, version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("%w: %q in %s", errInvalidMajorVersion, parts[0], version)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("%w: %q in %s", errInvalidMinorVersion, parts[1], version)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("%w: %q in %s", errInvalidPatchVersion, parts[2], version)
	}

	if major < 0 || minor < 0 || patch < 0 {
		return SemanticVersion{}, fmt.Errorf("%w: negative component in %s", errInvalidVersionFormat, version)
	}

	return SemanticVersion{major: major, minor: minor, patch: patch}, nil
}

// String returns the version in "vX.Y.Z" format.
func (sv SemanticVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", sv.major, sv.minor, sv.patch)
}

// Major returns the major version component.
func (sv SemanticVersion) Major() int {
	return sv.major
}

// Minor returns the minor version component.
func (sv SemanticVersion) Minor() int {
	return sv.minor
}

// Patch returns the patch version component.
func (sv SemanticVersion) Patch() int {
	return sv.patch
}

// Compare returns -1 if sv < other, 0 if sv == other, 1 if sv > other.
// Comparison is done component-wise: major, then minor, then patch.
func (sv SemanticVersion) Compare(other SemanticVersion) int {
	if sv.major != other.major {
		if sv.major < other.major {
			return -1
		}
		return 1
	}

	if sv.minor != other.minor {
		if sv.minor < other.minor {
			return -1
		}
		return 1
	}

	if sv.patch != other.patch {
		if sv.patch < other.patch {
			return -1
		}
		return 1
	}

	return 0
}

// IsNewerThan returns true if sv is strictly greater than other.
func (sv SemanticVersion) IsNewerThan(other SemanticVersion) bool {
	return sv.Compare(other) > 0
}

// Bump returns a new SemanticVersion with the specified component incremented.
// Valid bump types are "major", "minor", and "patch".
// - "major": increments major, resets minor and patch to 0
// - "minor": increments minor, resets patch to 0
// - "patch": increments patch
func (sv SemanticVersion) Bump(bumpType string) (SemanticVersion, error) {
	switch bumpType {
	case "major":
		return SemanticVersion{
			major: sv.major + 1,
			minor: 0,
			patch: 0,
		}, nil
	case "minor":
		return SemanticVersion{
			major: sv.major,
			minor: sv.minor + 1,
			patch: 0,
		}, nil
	case "patch":
		return SemanticVersion{
			major: sv.major,
			minor: sv.minor,
			patch: sv.patch + 1,
		}, nil
	default:
		return SemanticVersion{}, fmt.Errorf("%w: %q", errInvalidBumpType, bumpType)
	}
}

// IsBetween returns true if sv is strictly between start and end (exclusive).
// Returns false if sv equals start or end, or if the range is invalid.
func (sv SemanticVersion) IsBetween(start, end SemanticVersion) bool {
	return sv.Compare(start) > 0 && sv.Compare(end) < 0
}
