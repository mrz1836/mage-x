package env

import (
	"testing"
)

func TestCleanValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "simple value without comment",
			input: "value",
			want:  "value",
		},
		{
			name:  "value with space-prefixed comment",
			input: "value #comment",
			want:  "value",
		},
		{
			name:  "value with tab-prefixed comment",
			input: "value\t#comment",
			want:  "value",
		},
		{
			name:  "comment-only line",
			input: "#comment",
			want:  "",
		},
		{
			name:  "comment-only line with leading whitespace",
			input: "  #comment",
			want:  "",
		},
		{
			name:  "leading and trailing whitespace",
			input: "  value  ",
			want:  "value",
		},
		{
			name:  "hash in value without space prefix preserved",
			input: "value#notcomment",
			want:  "value#notcomment",
		},
		{
			name:  "multiple space-prefixed comments takes first",
			input: "value #first #second",
			want:  "value",
		},
		{
			name:  "trailing whitespace before comment",
			input: "value   #comment",
			want:  "value",
		},
		{
			name:  "path with hash that looks like comment",
			input: "/path/to/file#anchor",
			want:  "/path/to/file#anchor",
		},
		{
			name:  "URL with fragment",
			input: "https://example.com#section",
			want:  "https://example.com#section",
		},
		{
			name:  "color hex code",
			input: "#FF0000",
			want:  "",
		},
		{
			name:  "value equals sign preserved",
			input: "FOO=bar",
			want:  "FOO=bar",
		},
		{
			name:  "mixed tab and space before comment",
			input: "value \t#comment",
			want:  "value",
		},
		{
			name:  "only whitespace",
			input: "   ",
			want:  "",
		},
		{
			name:  "tab-prefixed comment after value with spaces",
			input: "some value\t#inline comment",
			want:  "some value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CleanValue(tt.input)
			if got != tt.want {
				t.Errorf("CleanValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetOr(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		setEnv       bool
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_GETOR_SET",
			envValue:     "actual_value",
			setEnv:       true,
			defaultValue: "default",
			want:         "actual_value",
		},
		{
			name:         "returns default when not set",
			key:          "TEST_GETOR_NOTSET",
			envValue:     "",
			setEnv:       false,
			defaultValue: "my_default",
			want:         "my_default",
		},
		{
			name:         "returns default when env is empty string",
			key:          "TEST_GETOR_EMPTY",
			envValue:     "",
			setEnv:       true,
			defaultValue: "fallback",
			want:         "fallback",
		},
		{
			name:         "cleans value with inline comment",
			key:          "TEST_GETOR_COMMENT",
			envValue:     "value #inline comment",
			setEnv:       true,
			defaultValue: "default",
			want:         "value",
		},
		{
			name:         "returns default for comment-only value",
			key:          "TEST_GETOR_ONLYCOMMENT",
			envValue:     "#this is only a comment",
			setEnv:       true,
			defaultValue: "default_for_comment",
			want:         "default_for_comment",
		},
		{
			name:         "trims whitespace from value",
			key:          "TEST_GETOR_WHITESPACE",
			envValue:     "  trimmed  ",
			setEnv:       true,
			defaultValue: "default",
			want:         "trimmed",
		},
		{
			name:         "returns default for whitespace-only",
			key:          "TEST_GETOR_WSONLY",
			envValue:     "   ",
			setEnv:       true,
			defaultValue: "default_ws",
			want:         "default_ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() here due to env manipulation
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := GetOr(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetOr(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		want     string
	}{
		{
			name:     "returns cleaned value when set",
			key:      "TEST_MUSTGET_SET",
			envValue: "value #comment",
			setEnv:   true,
			want:     "value",
		},
		{
			name:     "returns empty for unset",
			key:      "TEST_MUSTGET_NOTSET",
			envValue: "",
			setEnv:   false,
			want:     "",
		},
		{
			name:     "returns empty for empty value",
			key:      "TEST_MUSTGET_EMPTY",
			envValue: "",
			setEnv:   true,
			want:     "",
		},
		{
			name:     "returns empty for whitespace only",
			key:      "TEST_MUSTGET_WS",
			envValue: "   ",
			setEnv:   true,
			want:     "",
		},
		{
			name:     "trims and returns value",
			key:      "TEST_MUSTGET_TRIM",
			envValue: "  hello  ",
			setEnv:   true,
			want:     "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}

			got := MustGet(tt.key)
			if got != tt.want {
				t.Errorf("MustGet(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
