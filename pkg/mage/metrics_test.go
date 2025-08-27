package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatNumberWithCommas tests the formatNumberWithCommas helper function
func TestFormatNumberWithCommas(t *testing.T) {
	testCases := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "single digit",
			input:    5,
			expected: "5",
		},
		{
			name:     "double digit",
			input:    42,
			expected: "42",
		},
		{
			name:     "three digits",
			input:    123,
			expected: "123",
		},
		{
			name:     "four digits",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "five digits",
			input:    12345,
			expected: "12,345",
		},
		{
			name:     "six digits",
			input:    123456,
			expected: "123,456",
		},
		{
			name:     "seven digits",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "test file count example",
			input:    59475,
			expected: "59,475",
		},
		{
			name:     "go file count example",
			input:    53778,
			expected: "53,778",
		},
		{
			name:     "total count example",
			input:    113253,
			expected: "113,253",
		},
		{
			name:     "one million",
			input:    1000000,
			expected: "1,000,000",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "exactly thousand",
			input:    1000,
			expected: "1,000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumberWithCommas(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
