package utils

import (
	"reflect"
	"testing"
)

func TestParseNonEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "basic lines",
			input:    "file1.go\nfile2.go\nfile3.go",
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:     "lines with trailing newline",
			input:    "file1.go\nfile2.go\nfile3.go\n",
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:     "lines with empty lines",
			input:    "file1.go\n\nfile2.go\n\nfile3.go",
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:     "lines with whitespace",
			input:    "  file1.go  \n  file2.go  \n  file3.go  ",
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only whitespace",
			input:    "   \n\n   ",
			expected: []string{},
		},
		{
			name:     "single line",
			input:    "file1.go",
			expected: []string{"file1.go"},
		},
		{
			name:     "single line with newline",
			input:    "file1.go\n",
			expected: []string{"file1.go"},
		},
		{
			name:     "lines with leading/trailing whitespace",
			input:    "\n  file1.go\nfile2.go  \n  file3.go  \n\n",
			expected: []string{"file1.go", "file2.go", "file3.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseNonEmptyLines(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseNonEmptyLines() = %v, want %v", result, tt.expected)
			}
		})
	}
}
