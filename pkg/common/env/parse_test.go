package env

import (
	"testing"
)

func TestPositive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value int
		want  bool
	}{
		{1, true},
		{100, true},
		{0, false},
		{-1, false},
		{-100, false},
	}
	for _, tt := range tests {
		if got := Positive(tt.value); got != tt.want {
			t.Errorf("Positive(%d) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestNonNegative(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value int
		want  bool
	}{
		{0, true},
		{1, true},
		{100, true},
		{-1, false},
		{-100, false},
	}
	for _, tt := range tests {
		if got := NonNegative(tt.value); got != tt.want {
			t.Errorf("NonNegative(%d) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestPositiveFloat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value float64
		want  bool
	}{
		{0.1, true},
		{1.0, true},
		{100.5, true},
		{0.0, false},
		{-0.1, false},
		{-100.5, false},
	}
	for _, tt := range tests {
		if got := PositiveFloat(tt.value); got != tt.want {
			t.Errorf("PositiveFloat(%f) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		setEnv    bool
		validator func(int) bool
		wantVal   int
		wantOK    bool
	}{
		{
			name:      "positive value with Positive validator",
			envKey:    "TEST_PARSEINT_POS",
			envValue:  "42",
			setEnv:    true,
			validator: Positive,
			wantVal:   42,
			wantOK:    true,
		},
		{
			name:      "zero with NonNegative validator",
			envKey:    "TEST_PARSEINT_ZERO",
			envValue:  "0",
			setEnv:    true,
			validator: NonNegative,
			wantVal:   0,
			wantOK:    true,
		},
		{
			name:      "zero with Positive validator fails",
			envKey:    "TEST_PARSEINT_ZERO_POS",
			envValue:  "0",
			setEnv:    true,
			validator: Positive,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "negative rejected by Positive",
			envKey:    "TEST_PARSEINT_NEG",
			envValue:  "-5",
			setEnv:    true,
			validator: Positive,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "negative rejected by NonNegative",
			envKey:    "TEST_PARSEINT_NEG2",
			envValue:  "-1",
			setEnv:    true,
			validator: NonNegative,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "nil validator accepts any parsed value",
			envKey:    "TEST_PARSEINT_NIL",
			envValue:  "-100",
			setEnv:    true,
			validator: nil,
			wantVal:   -100,
			wantOK:    true,
		},
		{
			name:      "empty string returns false",
			envKey:    "TEST_PARSEINT_EMPTY",
			envValue:  "",
			setEnv:    true,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "unset variable returns false",
			envKey:    "TEST_PARSEINT_UNSET",
			envValue:  "",
			setEnv:    false,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "non-numeric string returns false",
			envKey:    "TEST_PARSEINT_ABC",
			envValue:  "abc",
			setEnv:    true,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "float value parses integer portion (Sscanf behavior)",
			envKey:    "TEST_PARSEINT_FLOAT",
			envValue:  "3.14",
			setEnv:    true,
			validator: nil,
			wantVal:   3,
			wantOK:    true,
		},
		{
			name:      "whitespace only returns false",
			envKey:    "TEST_PARSEINT_WS",
			envValue:  "  ",
			setEnv:    true,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "value with inline comment is cleaned",
			envKey:    "TEST_PARSEINT_COMMENT",
			envValue:  "123 #comment",
			setEnv:    true,
			validator: nil,
			wantVal:   123,
			wantOK:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			gotVal, gotOK := ParseInt(tt.envKey, tt.validator)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("ParseInt(%q) = (%d, %v), want (%d, %v)",
					tt.envKey, gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name      string
		envKey    string
		envValue  string
		setEnv    bool
		validator func(float64) bool
		wantVal   float64
		wantOK    bool
	}{
		{
			name:      "positive float",
			envKey:    "TEST_PARSEFLOAT_POS",
			envValue:  "3.14",
			setEnv:    true,
			validator: nil,
			wantVal:   3.14,
			wantOK:    true,
		},
		{
			name:      "integer parses as float",
			envKey:    "TEST_PARSEFLOAT_INT",
			envValue:  "42",
			setEnv:    true,
			validator: nil,
			wantVal:   42.0,
			wantOK:    true,
		},
		{
			name:      "zero with PositiveFloat validator fails",
			envKey:    "TEST_PARSEFLOAT_ZERO",
			envValue:  "0",
			setEnv:    true,
			validator: PositiveFloat,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "negative rejected by PositiveFloat",
			envKey:    "TEST_PARSEFLOAT_NEG",
			envValue:  "-2.5",
			setEnv:    true,
			validator: PositiveFloat,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "empty string returns false",
			envKey:    "TEST_PARSEFLOAT_EMPTY",
			envValue:  "",
			setEnv:    true,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "unset variable returns false",
			envKey:    "TEST_PARSEFLOAT_UNSET",
			envValue:  "",
			setEnv:    false,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
		{
			name:      "non-numeric string returns false",
			envKey:    "TEST_PARSEFLOAT_ABC",
			envValue:  "abc",
			setEnv:    true,
			validator: nil,
			wantVal:   0,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			gotVal, gotOK := ParseFloat(tt.envKey, tt.validator)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("ParseFloat(%q) = (%f, %v), want (%f, %v)",
					tt.envKey, gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestParseStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		setEnv   bool
		want     []string
	}{
		{
			name:     "simple comma-separated list",
			envKey:   "TEST_SLICE_SIMPLE",
			envValue: "a,b,c",
			setEnv:   true,
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "list with spaces around elements",
			envKey:   "TEST_SLICE_SPACES",
			envValue: " a , b , c ",
			setEnv:   true,
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "empty elements filtered out",
			envKey:   "TEST_SLICE_EMPTY",
			envValue: "a,,b,,,c",
			setEnv:   true,
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "single value",
			envKey:   "TEST_SLICE_SINGLE",
			envValue: "only",
			setEnv:   true,
			want:     []string{"only"},
		},
		{
			name:     "empty string returns nil",
			envKey:   "TEST_SLICE_EMPTYSTR",
			envValue: "",
			setEnv:   true,
			want:     nil,
		},
		{
			name:     "unset variable returns nil",
			envKey:   "TEST_SLICE_UNSET",
			envValue: "",
			setEnv:   false,
			want:     nil,
		},
		{
			name:     "whitespace only returns nil",
			envKey:   "TEST_SLICE_WS",
			envValue: "  ,  ,  ",
			setEnv:   true,
			want:     nil,
		},
		{
			name:     "value with inline comment",
			envKey:   "TEST_SLICE_COMMENT",
			envValue: "a,b,c #this is a comment",
			setEnv:   true,
			want:     []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			got := ParseStringSlice(tt.envKey)
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("ParseStringSlice(%q) = %v, want %v", tt.envKey, got, tt.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		setEnv   bool
		wantVal  bool
		wantOK   bool
	}{
		// True values
		{name: "true lowercase", envKey: "TEST_BOOL_TRUE", envValue: "true", setEnv: true, wantVal: true, wantOK: true},
		{name: "TRUE uppercase", envKey: "TEST_BOOL_TRUE2", envValue: "TRUE", setEnv: true, wantVal: true, wantOK: true},
		{name: "True mixed case", envKey: "TEST_BOOL_TRUE3", envValue: "True", setEnv: true, wantVal: true, wantOK: true},
		{name: "1", envKey: "TEST_BOOL_ONE", envValue: "1", setEnv: true, wantVal: true, wantOK: true},
		{name: "yes", envKey: "TEST_BOOL_YES", envValue: "yes", setEnv: true, wantVal: true, wantOK: true},
		{name: "YES", envKey: "TEST_BOOL_YES2", envValue: "YES", setEnv: true, wantVal: true, wantOK: true},
		{name: "on", envKey: "TEST_BOOL_ON", envValue: "on", setEnv: true, wantVal: true, wantOK: true},
		{name: "ON", envKey: "TEST_BOOL_ON2", envValue: "ON", setEnv: true, wantVal: true, wantOK: true},

		// False values
		{name: "false lowercase", envKey: "TEST_BOOL_FALSE", envValue: "false", setEnv: true, wantVal: false, wantOK: true},
		{name: "FALSE uppercase", envKey: "TEST_BOOL_FALSE2", envValue: "FALSE", setEnv: true, wantVal: false, wantOK: true},
		{name: "0", envKey: "TEST_BOOL_ZERO", envValue: "0", setEnv: true, wantVal: false, wantOK: true},
		{name: "no", envKey: "TEST_BOOL_NO", envValue: "no", setEnv: true, wantVal: false, wantOK: true},
		{name: "off", envKey: "TEST_BOOL_OFF", envValue: "off", setEnv: true, wantVal: false, wantOK: true},

		// Invalid/not found
		{name: "empty string", envKey: "TEST_BOOL_EMPTY", envValue: "", setEnv: true, wantVal: false, wantOK: false},
		{name: "unset variable", envKey: "TEST_BOOL_UNSET", envValue: "", setEnv: false, wantVal: false, wantOK: false},
		{name: "invalid value", envKey: "TEST_BOOL_INVALID", envValue: "maybe", setEnv: true, wantVal: false, wantOK: false},
		{name: "2 is not boolean", envKey: "TEST_BOOL_TWO", envValue: "2", setEnv: true, wantVal: false, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.envKey, tt.envValue)
			}

			gotVal, gotOK := ParseBool(tt.envKey)
			if gotVal != tt.wantVal || gotOK != tt.wantOK {
				t.Errorf("ParseBool(%q) = (%v, %v), want (%v, %v)",
					tt.envKey, gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

// stringSlicesEqual compares two string slices for equality.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
