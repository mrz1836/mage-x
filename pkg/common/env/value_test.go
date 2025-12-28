package env

import "testing"

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
