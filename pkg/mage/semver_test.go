package mage

import (
	"errors"
	"testing"
)

func TestParseSemanticVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   error
	}{
		// Valid versions with v prefix
		{name: "basic version", input: "v1.2.3", wantMajor: 1, wantMinor: 2, wantPatch: 3},
		{name: "zero version", input: "v0.0.0", wantMajor: 0, wantMinor: 0, wantPatch: 0},
		{name: "large numbers", input: "v99.88.77", wantMajor: 99, wantMinor: 88, wantPatch: 77},
		{name: "very large", input: "v100.200.300", wantMajor: 100, wantMinor: 200, wantPatch: 300},

		// Valid versions without v prefix
		{name: "no prefix", input: "1.2.3", wantMajor: 1, wantMinor: 2, wantPatch: 3},
		{name: "no prefix zeros", input: "0.0.1", wantMajor: 0, wantMinor: 0, wantPatch: 1},

		// Invalid versions
		{name: "empty string", input: "", wantErr: errInvalidVersionFormat},
		{name: "only v", input: "v", wantErr: errInvalidVersionFormat},
		{name: "too few parts", input: "v1.2", wantErr: errInvalidVersionFormat},
		{name: "too many parts", input: "v1.2.3.4", wantErr: errInvalidVersionFormat},
		{name: "non-numeric major", input: "va.2.3", wantErr: errInvalidMajorVersion},
		{name: "non-numeric minor", input: "v1.b.3", wantErr: errInvalidMinorVersion},
		{name: "non-numeric patch", input: "v1.2.c", wantErr: errInvalidPatchVersion},
		{name: "alpha suffix", input: "v1.2.3-alpha", wantErr: errInvalidPatchVersion},
		{name: "spaces", input: "v1. 2.3", wantErr: errInvalidMinorVersion},
		{name: "negative major", input: "v-1.2.3", wantErr: errInvalidVersionFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemanticVersion(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ParseSemanticVersion(%q) expected error %v, got nil", tt.input, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ParseSemanticVersion(%q) error = %v, want %v", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSemanticVersion(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got.Major() != tt.wantMajor || got.Minor() != tt.wantMinor || got.Patch() != tt.wantPatch {
				t.Errorf("ParseSemanticVersion(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, got.Major(), got.Minor(), got.Patch(),
					tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
		})
	}
}

func TestSemanticVersion_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "with prefix", input: "v1.2.3", want: "v1.2.3"},
		{name: "without prefix", input: "1.2.3", want: "v1.2.3"},
		{name: "zeros", input: "v0.0.0", want: "v0.0.0"},
		{name: "large", input: "v100.200.300", want: "v100.200.300"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv, err := ParseSemanticVersion(tt.input)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.input, err)
			}

			got := sv.String()
			if got != tt.want {
				t.Errorf("SemanticVersion.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSemanticVersion_RoundTrip(t *testing.T) {
	versions := []string{"v0.0.0", "v1.0.0", "v1.2.3", "v99.88.77"}

	for _, v := range versions {
		t.Run(v, func(t *testing.T) {
			sv1, err := ParseSemanticVersion(v)
			if err != nil {
				t.Fatalf("first parse failed: %v", err)
			}

			sv2, err := ParseSemanticVersion(sv1.String())
			if err != nil {
				t.Fatalf("second parse failed: %v", err)
			}

			if sv1.Compare(sv2) != 0 {
				t.Errorf("round-trip failed: %v -> %v -> %v", v, sv1.String(), sv2.String())
			}
		})
	}
}

func TestSemanticVersion_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		// Equal versions
		{name: "equal basic", a: "v1.2.3", b: "v1.2.3", want: 0},
		{name: "equal zeros", a: "v0.0.0", b: "v0.0.0", want: 0},

		// Major version differences
		{name: "major greater", a: "v2.0.0", b: "v1.0.0", want: 1},
		{name: "major less", a: "v1.0.0", b: "v2.0.0", want: -1},
		{name: "major dominates minor", a: "v2.0.0", b: "v1.99.99", want: 1},

		// Minor version differences
		{name: "minor greater", a: "v1.2.0", b: "v1.1.0", want: 1},
		{name: "minor less", a: "v1.1.0", b: "v1.2.0", want: -1},
		{name: "minor dominates patch", a: "v1.2.0", b: "v1.1.99", want: 1},

		// Patch version differences
		{name: "patch greater", a: "v1.2.4", b: "v1.2.3", want: 1},
		{name: "patch less", a: "v1.2.3", b: "v1.2.4", want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := ParseSemanticVersion(tt.a)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.a, err)
			}
			b, err := ParseSemanticVersion(tt.b)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.b, err)
			}

			got := a.Compare(b)
			if got != tt.want {
				t.Errorf("(%s).Compare(%s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSemanticVersion_IsNewerThan(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{name: "newer major", a: "v2.0.0", b: "v1.0.0", want: true},
		{name: "newer minor", a: "v1.2.0", b: "v1.1.0", want: true},
		{name: "newer patch", a: "v1.1.2", b: "v1.1.1", want: true},
		{name: "older major", a: "v1.0.0", b: "v2.0.0", want: false},
		{name: "older minor", a: "v1.1.0", b: "v1.2.0", want: false},
		{name: "older patch", a: "v1.1.1", b: "v1.1.2", want: false},
		{name: "equal", a: "v1.2.3", b: "v1.2.3", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := ParseSemanticVersion(tt.a)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.a, err)
			}
			b, err := ParseSemanticVersion(tt.b)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.b, err)
			}

			got := a.IsNewerThan(b)
			if got != tt.want {
				t.Errorf("(%s).IsNewerThan(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSemanticVersion_Bump(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		bumpType string
		want     string
		wantErr  error
	}{
		// Patch bumps
		{name: "patch basic", input: "v1.2.3", bumpType: "patch", want: "v1.2.4"},
		{name: "patch from zero", input: "v1.2.0", bumpType: "patch", want: "v1.2.1"},
		{name: "patch rollover", input: "v1.2.99", bumpType: "patch", want: "v1.2.100"},

		// Minor bumps
		{name: "minor basic", input: "v1.2.3", bumpType: "minor", want: "v1.3.0"},
		{name: "minor from zero", input: "v1.0.5", bumpType: "minor", want: "v1.1.0"},
		{name: "minor rollover", input: "v1.99.5", bumpType: "minor", want: "v1.100.0"},

		// Major bumps
		{name: "major basic", input: "v1.2.3", bumpType: "major", want: "v2.0.0"},
		{name: "major from zero", input: "v0.5.5", bumpType: "major", want: "v1.0.0"},
		{name: "major rollover", input: "v99.5.5", bumpType: "major", want: "v100.0.0"},

		// Invalid bump types
		{name: "invalid type", input: "v1.2.3", bumpType: "invalid", wantErr: errInvalidBumpType},
		{name: "empty type", input: "v1.2.3", bumpType: "", wantErr: errInvalidBumpType},
		{name: "MAJOR uppercase", input: "v1.2.3", bumpType: "MAJOR", wantErr: errInvalidBumpType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sv, err := ParseSemanticVersion(tt.input)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.input, err)
			}

			got, err := sv.Bump(tt.bumpType)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Bump(%q) expected error %v, got nil", tt.bumpType, tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Bump(%q) error = %v, want %v", tt.bumpType, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Bump(%q) unexpected error: %v", tt.bumpType, err)
				return
			}

			if got.String() != tt.want {
				t.Errorf("(%s).Bump(%q) = %s, want %s", tt.input, tt.bumpType, got.String(), tt.want)
			}
		})
	}
}

func TestSemanticVersion_IsBetween(t *testing.T) {
	tests := []struct {
		name   string
		middle string
		start  string
		end    string
		want   bool
	}{
		// Strictly between
		{name: "between major", middle: "v2.0.0", start: "v1.0.0", end: "v3.0.0", want: true},
		{name: "between minor", middle: "v1.2.0", start: "v1.1.0", end: "v1.3.0", want: true},
		{name: "between patch", middle: "v1.1.2", start: "v1.1.1", end: "v1.1.3", want: true},

		// Equal to boundary
		{name: "equal to start", middle: "v1.0.0", start: "v1.0.0", end: "v2.0.0", want: false},
		{name: "equal to end", middle: "v2.0.0", start: "v1.0.0", end: "v2.0.0", want: false},
		{name: "equal to both", middle: "v1.0.0", start: "v1.0.0", end: "v1.0.0", want: false},

		// Outside range
		{name: "before start", middle: "v0.5.0", start: "v1.0.0", end: "v2.0.0", want: false},
		{name: "after end", middle: "v2.5.0", start: "v1.0.0", end: "v2.0.0", want: false},

		// Inverted range (start > end)
		{name: "inverted range", middle: "v1.5.0", start: "v2.0.0", end: "v1.0.0", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middle, err := ParseSemanticVersion(tt.middle)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.middle, err)
			}
			start, err := ParseSemanticVersion(tt.start)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.start, err)
			}
			end, err := ParseSemanticVersion(tt.end)
			if err != nil {
				t.Fatalf("ParseSemanticVersion(%q) unexpected error: %v", tt.end, err)
			}

			got := middle.IsBetween(start, end)
			if got != tt.want {
				t.Errorf("(%s).IsBetween(%s, %s) = %v, want %v",
					tt.middle, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestSemanticVersion_Accessors(t *testing.T) {
	sv, err := ParseSemanticVersion("v10.20.30")
	if err != nil {
		t.Fatalf("ParseSemanticVersion failed: %v", err)
	}

	if got := sv.Major(); got != 10 {
		t.Errorf("Major() = %d, want 10", got)
	}
	if got := sv.Minor(); got != 20 {
		t.Errorf("Minor() = %d, want 20", got)
	}
	if got := sv.Patch(); got != 30 {
		t.Errorf("Patch() = %d, want 30", got)
	}
}

func TestSemanticVersion_ImmutableBump(t *testing.T) {
	original, err := ParseSemanticVersion("v1.2.3")
	if err != nil {
		t.Fatalf("ParseSemanticVersion failed: %v", err)
	}
	bumped, err := original.Bump("patch")
	if err != nil {
		t.Fatalf("Bump failed: %v", err)
	}

	if original.Patch() != 3 {
		t.Errorf("Original version was mutated: patch = %d, want 3", original.Patch())
	}
	if bumped.Patch() != 4 {
		t.Errorf("Bumped version incorrect: patch = %d, want 4", bumped.Patch())
	}
}
