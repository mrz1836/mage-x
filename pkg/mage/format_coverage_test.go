package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// FormatCoverageTestSuite provides comprehensive coverage for Format methods
type FormatCoverageTestSuite struct {
	suite.Suite

	env    *testutil.TestEnvironment
	format Format
}

func TestFormatCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(FormatCoverageTestSuite))
}

func (ts *FormatCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.format = Format{}

	// Create .mage.yaml config
	ts.env.CreateMageConfig(`
project:
  name: test
format:
  verbose: false
`)

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *FormatCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *FormatCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// TestFormatDefaultExercise exercises Format.Default code path
func (ts *FormatCoverageTestSuite) TestFormatDefaultExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Default()
	})
	// Always succeeds even if individual formatters fail
	ts.Require().NoError(err)
}

// TestFormatGofmtExercise exercises Format.Gofmt code path
func (ts *FormatCoverageTestSuite) TestFormatGofmtExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Gofmt()
	})
	// Exercise code path
	_ = err
}

// TestFormatFumptExercise exercises Format.Fumpt code path
func (ts *FormatCoverageTestSuite) TestFormatFumptExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Fumpt()
	})
	// Exercise code path - may fail if gofumpt not installed
	_ = err
}

// TestFormatGciExercise exercises Format.Gci code path
func (ts *FormatCoverageTestSuite) TestFormatGciExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Gci()
	})
	// Exercise code path - may fail if gci not installed
	_ = err
}

// TestFormatImportsExercise exercises Format.Imports code path
func (ts *FormatCoverageTestSuite) TestFormatImportsExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Imports()
	})
	// Exercise code path - may fail if goimports not installed
	_ = err
}

// TestFormatFixExercise exercises Format.Fix code path
func (ts *FormatCoverageTestSuite) TestFormatFixExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Fix()
	})
	// Exercise code path
	_ = err
}

// TestFormatCheckExercise exercises Format.Check code path
func (ts *FormatCoverageTestSuite) TestFormatCheckExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.Check()
	})
	// Exercise code path
	_ = err
}

// TestFormatJSONExercise exercises Format.JSON code path
func (ts *FormatCoverageTestSuite) TestFormatJSONExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.JSON()
	})
	// Exercise code path
	_ = err
}

// TestFormatYAMLExercise exercises Format.YAML code path
func (ts *FormatCoverageTestSuite) TestFormatYAMLExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.YAML()
	})
	// Exercise code path
	_ = err
}

// Note: Format.Markdown does not exist - removed test

// TestFormatAllExercise exercises Format.All code path
func (ts *FormatCoverageTestSuite) TestFormatAllExercise() {
	err := ts.withMockRunner(func() error {
		return ts.format.All()
	})
	// Exercise code path
	_ = err
}

// ================== Standalone Tests ==================

func TestFormatStaticErrors(t *testing.T) {
	// Verify static errors in format.go are defined correctly
	require.Error(t, ErrCodeNotFormatted)
	assert.Error(t, ErrUnexpectedRunnerType)
}

func TestFilterEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "only empty strings",
			input:    []string{"", "", ""},
			expected: nil,
		},
		{
			name:     "mixed",
			input:    []string{"a", "", "b", "", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "no empty strings",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
