package mage

import "testing"

// stubGoMinor overrides the active-Go detector for a test, restoring it on cleanup.
func stubGoMinor(t *testing.T, major, minor int, ok bool) {
	t.Helper()
	orig := detectGoMinorForToolSelection
	t.Cleanup(func() { detectGoMinorForToolSelection = orig })
	detectGoMinorForToolSelection = func() (int, int, bool) { return major, minor, ok }
}

func TestParseGoMajorMinor(t *testing.T) {
	tests := []struct {
		in           string
		major, minor int
		ok           bool
	}{
		{"1.26.0", 1, 26, true},
		{"1.26", 1, 26, true},
		{"1.26.x", 1, 26, true},
		{"go1.25.3", 1, 25, true},
		{"2.0.0", 2, 0, true},
		{"", 0, 0, false},
		{"abc", 0, 0, false},
		{"1", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			major, minor, ok := parseGoMajorMinor(tt.in)
			if ok != tt.ok {
				t.Fatalf("parseGoMajorMinor(%q) ok = %v, want %v", tt.in, ok, tt.ok)
			}
			if ok && (major != tt.major || minor != tt.minor) {
				t.Errorf("parseGoMajorMinor(%q) = %d.%d, want %d.%d", tt.in, major, minor, tt.major, tt.minor)
			}
		})
	}
}

// TestGetToolVersion_DualVersion verifies that gofumpt's version is chosen by the
// active Go toolchain when the _LATEST pins are present, without forcing repos on an
// older Go to upgrade alongside a tool bump.
func TestGetToolVersion_DualVersion(t *testing.T) {
	const (
		baseline = "v0.11.0"
		latest   = "v0.12.0"
	)

	t.Run("no _LATEST set uses baseline (inert)", func(t *testing.T) {
		t.Setenv("MAGE_X_GOFUMPT_VERSION", baseline)
		// detector must not even be consulted; make it fail loudly if it is.
		stubGoMinor(t, 1, 30, true)
		if got := GetToolVersion("gofumpt"); got != baseline {
			t.Errorf("GetToolVersion = %q, want %q", got, baseline)
		}
	})

	t.Run("newer Go selects _LATEST", func(t *testing.T) {
		t.Setenv("MAGE_X_GOFUMPT_VERSION", baseline)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST", latest)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST_MIN_GO", "1.26")
		stubGoMinor(t, 1, 26, true)
		if got := GetToolVersion("gofumpt"); got != latest {
			t.Errorf("GetToolVersion = %q, want %q", got, latest)
		}
	})

	t.Run("older Go keeps baseline", func(t *testing.T) {
		t.Setenv("MAGE_X_GOFUMPT_VERSION", baseline)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST", latest)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST_MIN_GO", "1.26")
		stubGoMinor(t, 1, 25, true)
		if got := GetToolVersion("gofumpt"); got != baseline {
			t.Errorf("GetToolVersion = %q, want %q", got, baseline)
		}
	})

	t.Run("undetectable Go is conservative (baseline)", func(t *testing.T) {
		t.Setenv("MAGE_X_GOFUMPT_VERSION", baseline)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST", latest)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST_MIN_GO", "1.26")
		stubGoMinor(t, 0, 0, false)
		if got := GetToolVersion("gofumpt"); got != baseline {
			t.Errorf("GetToolVersion = %q, want %q", got, baseline)
		}
	})

	t.Run("missing _LATEST_MIN_GO defaults to 1.26", func(t *testing.T) {
		t.Setenv("MAGE_X_GOFUMPT_VERSION", baseline)
		t.Setenv("MAGE_X_GOFUMPT_VERSION_LATEST", latest)
		// no _LATEST_MIN_GO; default 1.26. Go 1.25 => baseline.
		stubGoMinor(t, 1, 25, true)
		if got := GetToolVersion("gofumpt"); got != baseline {
			t.Errorf("GetToolVersion = %q, want %q", got, baseline)
		}
	})
}
