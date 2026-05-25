package mage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// errBoom is a static sentinel error used by table-driven tests to assert
// that runner errors propagate through Test.Run unchanged.
var errBoom = errors.New("boom")

// TestTestRun_ArgHandling verifies that Test.Run translates name= and pkg=
// parameters into the correct `go test` invocation, including back-compat
// behavior when no params are provided.
func TestTestRun_ArgHandling(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectArgs  []interface{} // expected args to RunCmd ("go", ...)
		runnerErr   error
		wantErr     bool
		wantErrorIs error
	}{
		{
			name:       "no args runs all packages",
			args:       nil,
			expectArgs: []interface{}{"go", "test", "./..."},
		},
		{
			name:       "empty args slice runs all packages",
			args:       []string{},
			expectArgs: []interface{}{"go", "test", "./..."},
		},
		{
			name:       "name only adds -run and defaults pkg to ./...",
			args:       []string{"name=TestFoo"},
			expectArgs: []interface{}{"go", "test", "-run", "TestFoo", "./..."},
		},
		{
			name:       "pkg only restricts to that package",
			args:       []string{"pkg=./internal/vault"},
			expectArgs: []interface{}{"go", "test", "./internal/vault"},
		},
		{
			name:       "name and pkg combine into -run + path",
			args:       []string{"name=TestSecureZero", "pkg=./internal/vault"},
			expectArgs: []interface{}{"go", "test", "-run", "TestSecureZero", "./internal/vault"},
		},
		{
			name:       "pkg before name is order-independent",
			args:       []string{"pkg=./pkg/utils", "name=TestParseParams"},
			expectArgs: []interface{}{"go", "test", "-run", "TestParseParams", "./pkg/utils"},
		},
		{
			name:       "unknown params are ignored",
			args:       []string{"flavor=spicy", "name=TestX"},
			expectArgs: []interface{}{"go", "test", "-run", "TestX", "./..."},
		},
		{
			name:        "runner errors propagate",
			args:        []string{"name=TestX"},
			expectArgs:  []interface{}{"go", "test", "-run", "TestX", "./..."},
			runnerErr:   errBoom,
			wantErr:     true,
			wantErrorIs: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockRunner(t, func(m *MockCommandRunner) {
				m.On("RunCmd", tt.expectArgs...).Return(tt.runnerErr).Once()

				err := Test{}.Run(tt.args...)
				if tt.wantErr {
					require.Error(t, err)
					if tt.wantErrorIs != nil {
						require.ErrorIs(t, err, tt.wantErrorIs)
					}
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}
