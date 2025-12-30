package testutil

import (
	"github.com/stretchr/testify/mock"
)

// MockGoOperations provides test mocking for Go toolchain operations.
// It implements the mage.GoOperations interface.
type MockGoOperations struct {
	mock.Mock
}

// GetModuleName returns the current Go module name from go.mod.
func (m *MockGoOperations) GetModuleName() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// GetVersion returns the current project version.
func (m *MockGoOperations) GetVersion() string {
	args := m.Called()
	return args.String(0)
}

// GetGoVulnCheckVersion returns the govulncheck tool version.
func (m *MockGoOperations) GetGoVulnCheckVersion() string {
	args := m.Called()
	return args.String(0)
}

// NewMockGoOperations creates a new mock Go operations instance.
func NewMockGoOperations() *MockGoOperations {
	return &MockGoOperations{}
}

// MockGoBuilder provides a fluent interface for setting up Go operation mocks.
type MockGoBuilder struct {
	ops *MockGoOperations
}

// NewMockGoBuilder creates a new builder with an associated mock.
func NewMockGoBuilder() (*MockGoOperations, *MockGoBuilder) {
	ops := NewMockGoOperations()
	return ops, &MockGoBuilder{ops: ops}
}

// WithModuleName sets up a module name expectation.
func (b *MockGoBuilder) WithModuleName(name string, err error) *MockGoBuilder {
	b.ops.On("GetModuleName").Return(name, err).Maybe()
	return b
}

// WithVersion sets up a version expectation.
func (b *MockGoBuilder) WithVersion(version string) *MockGoBuilder {
	b.ops.On("GetVersion").Return(version).Maybe()
	return b
}

// WithGoVulnCheckVersion sets up a govulncheck version expectation.
func (b *MockGoBuilder) WithGoVulnCheckVersion(version string) *MockGoBuilder {
	b.ops.On("GetGoVulnCheckVersion").Return(version).Maybe()
	return b
}

// Build returns the configured mock.
func (b *MockGoBuilder) Build() *MockGoOperations {
	return b.ops
}
