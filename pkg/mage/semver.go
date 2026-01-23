package mage

import (
	"fmt"
	"strconv"
	"strings"
)

// SemanticVersion represents a semantic version (major.minor.patch[-prerelease]).
// It provides parsing, comparison, and manipulation methods for version strings.
// Pre-release versions (e.g., v1.0.0-beta, v1.0.0-rc.1) are also supported.
type SemanticVersion struct {
	major      int
	minor      int
	patch      int
	prerelease string // Optional pre-release identifier (e.g., "beta", "rc.1", "alpha.2")
}

// ParseSemanticVersion parses a version string into a SemanticVersion.
// It accepts versions with or without the "v" prefix (e.g., "v1.2.3" or "1.2.3").
// It also supports pre-release suffixes (e.g., "v1.2.3-beta", "v1.2.3-rc.1").
// Returns an error if the version string is malformed.
func ParseSemanticVersion(version string) (SemanticVersion, error) {
	clean := strings.TrimPrefix(version, "v")

	// Extract pre-release suffix if present (e.g., "1.2.3-beta" -> "1.2.3", "beta")
	var prerelease string
	if idx := strings.Index(clean, "-"); idx != -1 {
		prerelease = clean[idx+1:]
		clean = clean[:idx]
	}

	// Also handle build metadata (+build) by stripping it
	if idx := strings.Index(clean, "+"); idx != -1 {
		clean = clean[:idx]
	}

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

	return SemanticVersion{major: major, minor: minor, patch: patch, prerelease: prerelease}, nil
}

// String returns the version in "vX.Y.Z" or "vX.Y.Z-prerelease" format.
func (sv SemanticVersion) String() string {
	if sv.prerelease != "" {
		return fmt.Sprintf("v%d.%d.%d-%s", sv.major, sv.minor, sv.patch, sv.prerelease)
	}
	return fmt.Sprintf("v%d.%d.%d", sv.major, sv.minor, sv.patch)
}

// BaseString returns the version in "vX.Y.Z" format without pre-release suffix.
func (sv SemanticVersion) BaseString() string {
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

// Prerelease returns the pre-release identifier (e.g., "beta", "rc.1").
// Returns empty string for stable releases.
func (sv SemanticVersion) Prerelease() string {
	return sv.prerelease
}

// IsPrerelease returns true if this version has a pre-release suffix.
func (sv SemanticVersion) IsPrerelease() bool {
	return sv.prerelease != ""
}

// Compare returns -1 if sv < other, 0 if sv == other, 1 if sv > other.
// Comparison is done component-wise: major, then minor, then patch, then prerelease.
// Per semver spec, a pre-release version has lower precedence than the normal version
// (e.g., 1.0.0-alpha < 1.0.0).
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

	// Handle pre-release comparison per semver spec:
	// - A version without pre-release > a version with pre-release (same M.m.p)
	// - If both have pre-release, compare lexically
	if sv.prerelease == "" && other.prerelease == "" {
		return 0
	}
	if sv.prerelease == "" && other.prerelease != "" {
		// sv is stable, other is prerelease -> sv is higher
		return 1
	}
	if sv.prerelease != "" && other.prerelease == "" {
		// sv is prerelease, other is stable -> sv is lower
		return -1
	}

	// Both have prereleases - compare them
	return comparePrerelease(sv.prerelease, other.prerelease)
}

// comparePrerelease compares two pre-release identifiers according to semver spec.
// Identifiers consisting of only digits are compared numerically.
// Identifiers with letters or hyphens are compared lexically.
// Numeric identifiers always have lower precedence than non-numeric.
// A larger set of pre-release fields has a higher precedence than a smaller set.
func comparePrerelease(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}

	for i := 0; i < minLen; i++ {
		cmp := comparePrereleaseIdentifier(partsA[i], partsB[i])
		if cmp != 0 {
			return cmp
		}
	}

	// If all compared parts are equal, the one with more parts is greater
	if len(partsA) < len(partsB) {
		return -1
	}
	if len(partsA) > len(partsB) {
		return 1
	}
	return 0
}

// comparePrereleaseIdentifier compares a single pre-release identifier part.
func comparePrereleaseIdentifier(a, b string) int {
	aNum, aErr := strconv.Atoi(a)
	bNum, bErr := strconv.Atoi(b)

	// Both numeric - compare numerically
	if aErr == nil && bErr == nil {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// Numeric < non-numeric
	if aErr == nil && bErr != nil {
		return -1
	}
	if aErr != nil && bErr == nil {
		return 1
	}

	// Both non-numeric - compare lexically
	if a < b {
		return -1
	}
	if a > b {
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
