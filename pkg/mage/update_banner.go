// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// Banner color constants
const (
	bannerColorReset  = "\033[0m"
	bannerColorYellow = "\033[33m"
)

// Banner layout constants
const (
	// upgradeCmd is the command users should run to update
	upgradeCmd = "magex update:install"

	// versionDisplayWidth is the fixed width for version strings in the banner
	// Chosen to accommodate typical semver versions like "v1.2.3-beta.1"
	versionDisplayWidth = 12
)

// ShowUpdateBanner displays a banner when an update is available
func ShowUpdateBanner(result *UpdateCheckResult) {
	if result == nil || !result.UpdateAvailable {
		return
	}

	banner := formatUpdateBanner(result.CurrentVersion, result.LatestVersion)

	if shouldUseColorForBanner() {
		fmt.Printf("\n%s%s%s\n", bannerColorYellow, banner, bannerColorReset)
	} else {
		fmt.Printf("\n%s\n", banner)
	}
}

// formatUpdateBanner creates the update notification banner
func formatUpdateBanner(current, latest string) string {
	const boxWidth = 70 // characters between | and | (matches 70 dashes in border)

	currentPadded := padVersion(current, versionDisplayWidth)
	latestPadded := padVersion(latest, versionDisplayWidth)

	// Build version line and pad to box width
	versionLine := fmt.Sprintf("   Current: %s   Latest: %s", currentPadded, latestPadded)
	versionLine = padRight(versionLine, boxWidth)

	// Build command line and pad to box width
	cmdLine := "   " + upgradeCmd
	cmdLine = padRight(cmdLine, boxWidth)

	emptyLine := padRight("", boxWidth)

	lines := []string{
		"",
		"  +----------------------------------------------------------------------+",
		"  |" + emptyLine + "|",
		"  |" + padRight("   A new version of MAGE-X is available!", boxWidth) + "|",
		"  |" + emptyLine + "|",
		"  |" + versionLine + "|",
		"  |" + emptyLine + "|",
		"  |" + padRight("   Upgrade command:", boxWidth) + "|",
		"  |" + cmdLine + "|",
		"  |" + emptyLine + "|",
		"  +----------------------------------------------------------------------+",
		"",
	}

	return strings.Join(lines, "\n")
}

// FormatUpdateBannerFancy creates a fancier update notification banner with box drawing
func FormatUpdateBannerFancy(current, latest string) string {
	const boxWidth = 70 // characters between │ and │ (matches 70 dashes in border)

	currentPadded := padVersion(current, versionDisplayWidth)
	latestPadded := padVersion(latest, versionDisplayWidth)

	// Build version line and pad to box width
	versionLine := fmt.Sprintf("   Current: %s   Latest: %s", currentPadded, latestPadded)
	versionLine = padRight(versionLine, boxWidth)

	// Build command line and pad to box width
	cmdLine := "   " + upgradeCmd
	cmdLine = padRight(cmdLine, boxWidth)

	emptyLine := padRight("", boxWidth)

	lines := []string{
		"",
		"  ╭──────────────────────────────────────────────────────────────────────╮",
		"  │" + emptyLine + "│",
		"  │" + padRight("   A new version of MAGE-X is available!", boxWidth) + "│",
		"  │" + emptyLine + "│",
		"  │" + versionLine + "│",
		"  │" + emptyLine + "│",
		"  │" + padRight("   Upgrade:", boxWidth) + "│",
		"  │" + cmdLine + "│",
		"  │" + emptyLine + "│",
		"  ╰──────────────────────────────────────────────────────────────────────╯",
		"",
	}

	return strings.Join(lines, "\n")
}

// FormatUpdateBannerMinimal creates a minimal update notification
func FormatUpdateBannerMinimal(current, latest string) string {
	return fmt.Sprintf(
		"\n  Update available: %s -> %s\n  Run: %s\n",
		current, latest, upgradeCmd,
	)
}

// padVersion pads a version string to a fixed width (Unicode-safe)
// Uses rune count for proper handling of multi-byte characters
func padVersion(version string, width int) string {
	runeCount := utf8.RuneCountInString(version)
	if runeCount >= width {
		// Truncate to width runes, not bytes
		runes := []rune(version)
		return string(runes[:width])
	}
	return version + strings.Repeat(" ", width-runeCount)
}

// padRight pads a string to a fixed width on the right (Unicode-safe)
// Uses rune count for proper handling of multi-byte characters
func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		// Truncate to width runes, not bytes
		runes := []rune(s)
		return string(runes[:width])
	}
	return s + strings.Repeat(" ", width-runeCount)
}

// shouldUseColorForBanner determines if color output should be enabled for the banner
func shouldUseColorForBanner() bool {
	// Disable color in CI environments
	if os.Getenv("CI") != "" {
		return false
	}

	// Disable color if NO_COLOR is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Disable color if not a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// GetUpgradeCommand returns the upgrade command string
func GetUpgradeCommand() string {
	return upgradeCmd
}
