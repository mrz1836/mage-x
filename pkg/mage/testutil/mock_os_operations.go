package testutil

import (
	"os"

	"github.com/stretchr/testify/mock"
)

// MockOSOperations provides test mocking for OS operations.
// It implements the mage.OSOperations interface.
type MockOSOperations struct {
	mock.Mock
}

// Getenv retrieves the value of the environment variable named by the key.
func (m *MockOSOperations) Getenv(key string) string {
	args := m.Called(key)
	return args.String(0)
}

// Remove removes the named file or (empty) directory.
func (m *MockOSOperations) Remove(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

// TempDir returns the default directory to use for temporary files.
func (m *MockOSOperations) TempDir() string {
	args := m.Called()
	return args.String(0)
}

// FileExists checks if a file exists at the given path.
func (m *MockOSOperations) FileExists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

// WriteFile writes data to the named file, creating it if necessary.
func (m *MockOSOperations) WriteFile(path string, data []byte, perm os.FileMode) error {
	args := m.Called(path, data, perm)
	return args.Error(0)
}

// Symlink creates newname as a symbolic link to oldname.
func (m *MockOSOperations) Symlink(oldname, newname string) error {
	args := m.Called(oldname, newname)
	return args.Error(0)
}

// Readlink returns the destination of the named symbolic link.
func (m *MockOSOperations) Readlink(name string) (string, error) {
	args := m.Called(name)
	return args.String(0), args.Error(1)
}

// NewMockOSOperations creates a new mock OS operations instance.
func NewMockOSOperations() *MockOSOperations {
	return &MockOSOperations{}
}

// MockOSBuilder provides a fluent interface for setting up OS operation mocks.
type MockOSBuilder struct {
	ops *MockOSOperations
}

// NewMockOSBuilder creates a new builder with an associated mock.
func NewMockOSBuilder() (*MockOSOperations, *MockOSBuilder) {
	ops := NewMockOSOperations()
	return ops, &MockOSBuilder{ops: ops}
}

// WithEnv sets up an environment variable expectation.
func (b *MockOSBuilder) WithEnv(key, value string) *MockOSBuilder {
	b.ops.On("Getenv", key).Return(value).Maybe()
	return b
}

// WithTempDir sets up the temp directory expectation.
func (b *MockOSBuilder) WithTempDir(dir string) *MockOSBuilder {
	b.ops.On("TempDir").Return(dir).Maybe()
	return b
}

// WithFileExists sets up a file existence expectation.
func (b *MockOSBuilder) WithFileExists(path string, exists bool) *MockOSBuilder {
	b.ops.On("FileExists", path).Return(exists).Maybe()
	return b
}

// WithRemoveSuccess sets up a successful remove expectation.
func (b *MockOSBuilder) WithRemoveSuccess(path string) *MockOSBuilder {
	b.ops.On("Remove", path).Return(nil).Maybe()
	return b
}

// WithRemoveError sets up a failing remove expectation.
func (b *MockOSBuilder) WithRemoveError(path string, err error) *MockOSBuilder {
	b.ops.On("Remove", path).Return(err).Maybe()
	return b
}

// WithWriteFileSuccess sets up a successful write file expectation.
func (b *MockOSBuilder) WithWriteFileSuccess(path string) *MockOSBuilder {
	b.ops.On("WriteFile", path, mock.Anything, mock.Anything).Return(nil).Maybe()
	return b
}

// WithSymlinkSuccess sets up a successful symlink expectation.
func (b *MockOSBuilder) WithSymlinkSuccess(oldname, newname string) *MockOSBuilder {
	b.ops.On("Symlink", oldname, newname).Return(nil).Maybe()
	return b
}

// WithReadlink sets up a readlink expectation.
func (b *MockOSBuilder) WithReadlink(name, target string, err error) *MockOSBuilder {
	b.ops.On("Readlink", name).Return(target, err).Maybe()
	return b
}

// Build returns the configured mock.
func (b *MockOSBuilder) Build() *MockOSOperations {
	return b.ops
}
