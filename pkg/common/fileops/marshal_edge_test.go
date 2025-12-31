package fileops

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONMarshalUnmarshalableTypes tests JSON marshal with types that can't be marshaled.
func TestJSONMarshalUnmarshalableTypes(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "Channel",
			value: make(chan int),
		},
		{
			name:  "Function",
			value: func() {},
		},
		{
			name:  "ComplexNumber",
			value: complex(1, 2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := op.Marshal(tt.value)
			require.Error(t, err)
		})
	}
}

// TestJSONMarshalIndentUnmarshalableTypes tests JSON MarshalIndent with types that can't be marshaled.
func TestJSONMarshalIndentUnmarshalableTypes(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "Channel",
			value: make(chan int),
		},
		{
			name:  "Function",
			value: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := op.MarshalIndent(tt.value, "", "  ")
			require.Error(t, err)
		})
	}
}

// TestYAMLMarshalUnmarshalableTypesPanics tests that YAML marshal panics for unmarshalable types.
// Note: yaml.v3 panics instead of returning an error for certain types like functions.
func TestYAMLMarshalUnmarshalableTypesPanics(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	tests := []struct {
		name  string
		value interface{}
	}{
		{
			name:  "Function",
			value: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// yaml.v3 panics for functions, so we need to recover
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic when marshaling function, but did not panic")
				}
			}()
			//nolint:errcheck // Intentionally not checking error - expecting panic
			_, _ = op.Marshal(tt.value)
		})
	}
}

// TestWriteJSONToReadOnlyDirectory tests WriteJSON when directory is read-only.
func TestWriteJSONToReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == goosWindows {
		t.Skip("Skipping on Windows - read-only directory behavior differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping as root - root can write to read-only directories")
	}

	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o555)) //nolint:gosec // G301: Test file - intentional permissions
	defer func() {
		if err := os.Chmod(readOnlyDir, 0o755); err != nil { //nolint:gosec // G302: Test file - intentional permissions
			t.Logf("cleanup chmod failed: %v", err)
		}
	}()

	path := filepath.Join(readOnlyDir, "test.json")
	err := op.WriteJSON(path, map[string]string{"key": "value"})
	require.Error(t, err)
}

// TestWriteJSONIndentToReadOnlyDirectory tests WriteJSONIndent when directory is read-only.
func TestWriteJSONIndentToReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == goosWindows {
		t.Skip("Skipping on Windows - read-only directory behavior differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping as root - root can write to read-only directories")
	}

	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o555)) //nolint:gosec // G301: Test file - intentional permissions
	defer func() {
		if err := os.Chmod(readOnlyDir, 0o755); err != nil { //nolint:gosec // G302: Test file - intentional permissions
			t.Logf("cleanup chmod failed: %v", err)
		}
	}()

	path := filepath.Join(readOnlyDir, "test.json")
	err := op.WriteJSONIndent(path, map[string]string{"key": "value"}, "", "  ")
	require.Error(t, err)
}

// TestWriteYAMLToReadOnlyDirectory tests WriteYAML when directory is read-only.
func TestWriteYAMLToReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == goosWindows {
		t.Skip("Skipping on Windows - read-only directory behavior differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping as root - root can write to read-only directories")
	}

	tmpDir := t.TempDir()
	op := NewDefaultYAMLOperator(nil)

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o555)) //nolint:gosec // G301: Test file - intentional permissions
	defer func() {
		if err := os.Chmod(readOnlyDir, 0o755); err != nil { //nolint:gosec // G302: Test file - intentional permissions
			t.Logf("cleanup chmod failed: %v", err)
		}
	}()

	path := filepath.Join(readOnlyDir, "test.yaml")
	err := op.WriteYAML(path, map[string]string{"key": "value"})
	require.Error(t, err)
}

// TestReadJSONFromNonExistent tests ReadJSON with non-existent file.
func TestReadJSONFromNonExistent(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	var result map[string]interface{}
	err := op.ReadJSON("/nonexistent/path/file.json", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// TestReadYAMLFromNonExistent tests ReadYAML with non-existent file.
func TestReadYAMLFromNonExistent(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	var result map[string]interface{}
	err := op.ReadYAML("/nonexistent/path/file.yaml", &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

// TestReadJSONInvalidContent tests ReadJSON with invalid JSON content.
func TestReadJSONInvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	path := filepath.Join(tmpDir, "invalid.json")
	require.NoError(t, os.WriteFile(path, []byte("not valid json {"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	var result map[string]interface{}
	err := op.ReadJSON(path, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal JSON")
}

// TestReadYAMLInvalidContent tests ReadYAML with invalid YAML content.
func TestReadYAMLInvalidContent(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultYAMLOperator(nil)

	path := filepath.Join(tmpDir, "invalid.yaml")
	// Invalid YAML - tabs in wrong places
	require.NoError(t, os.WriteFile(path, []byte(":\t:invalid:\n\t\t--"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	var result map[string]interface{}
	err := op.ReadYAML(path, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal YAML")
}

// TestJSONUnmarshalInvalidTarget tests JSON Unmarshal with invalid target.
func TestJSONUnmarshalInvalidTarget(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	data := []byte(`{"key": "value"}`)

	// Try to unmarshal into non-pointer
	var result string
	err := op.Unmarshal(data, result)
	require.Error(t, err)
}

// TestYAMLUnmarshalInvalidTarget tests YAML Unmarshal with invalid target.
func TestYAMLUnmarshalInvalidTarget(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	data := []byte("key: value")

	// Try to unmarshal into non-pointer
	var result string
	err := op.Unmarshal(data, result)
	require.Error(t, err)
}

// TestJSONRoundTripComplexStructure tests JSON with complex nested structures.
func TestJSONRoundTripComplexStructure(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	type nested struct {
		Values []int             `json:"values"`
		Map    map[string]string `json:"map"`
	}
	type complexStruct struct {
		Name   string   `json:"name"`
		Count  int      `json:"count"`
		Nested nested   `json:"nested"`
		Tags   []string `json:"tags"`
	}

	original := complexStruct{
		Name:  "test",
		Count: 42,
		Nested: nested{
			Values: []int{1, 2, 3, 4, 5},
			Map:    map[string]string{"a": "b", "c": "d"},
		},
		Tags: []string{"tag1", "tag2"},
	}

	path := filepath.Join(tmpDir, "complex.json")
	require.NoError(t, op.WriteJSON(path, original))

	var loaded complexStruct
	require.NoError(t, op.ReadJSON(path, &loaded))

	assert.Equal(t, original.Name, loaded.Name)
	assert.Equal(t, original.Count, loaded.Count)
	assert.Equal(t, original.Nested.Values, loaded.Nested.Values)
	assert.Equal(t, original.Tags, loaded.Tags)
}

// TestYAMLRoundTripComplexStructure tests YAML with complex nested structures.
func TestYAMLRoundTripComplexStructure(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultYAMLOperator(nil)

	type nested struct {
		Values []int             `yaml:"values"`
		Map    map[string]string `yaml:"map"`
	}
	type complexStruct struct {
		Name   string   `yaml:"name"`
		Count  int      `yaml:"count"`
		Nested nested   `yaml:"nested"`
		Tags   []string `yaml:"tags"`
	}

	original := complexStruct{
		Name:  "test",
		Count: 42,
		Nested: nested{
			Values: []int{1, 2, 3, 4, 5},
			Map:    map[string]string{"a": "b", "c": "d"},
		},
		Tags: []string{"tag1", "tag2"},
	}

	path := filepath.Join(tmpDir, "complex.yaml")
	require.NoError(t, op.WriteYAML(path, original))

	var loaded complexStruct
	require.NoError(t, op.ReadYAML(path, &loaded))

	assert.Equal(t, original.Name, loaded.Name)
	assert.Equal(t, original.Count, loaded.Count)
	assert.Equal(t, original.Nested.Values, loaded.Nested.Values)
	assert.Equal(t, original.Tags, loaded.Tags)
}

// TestWriteJSONWithMarshalError tests WriteJSON when marshal fails.
func TestWriteJSONWithMarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	path := filepath.Join(tmpDir, "test.json")

	// Try to marshal unmarshalable type
	err := op.WriteJSON(path, make(chan int))
	require.Error(t, err)

	// File should not exist
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}

// TestWriteJSONIndentWithMarshalError tests WriteJSONIndent when marshal fails.
func TestWriteJSONIndentWithMarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultJSONOperator(nil)

	path := filepath.Join(tmpDir, "test.json")

	// Try to marshal unmarshalable type
	err := op.WriteJSONIndent(path, make(chan int), "", "  ")
	require.Error(t, err)

	// File should not exist
	_, statErr := os.Stat(path)
	assert.True(t, os.IsNotExist(statErr))
}

// TestWriteYAMLWithMarshalPanic tests WriteYAML when marshal panics.
// Note: yaml.v3 panics for unmarshalable types like functions.
func TestWriteYAMLWithMarshalPanic(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultYAMLOperator(nil)

	path := filepath.Join(tmpDir, "test.yaml")

	// Try to marshal unmarshalable type (function) - yaml.v3 panics
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when marshaling function")
		}
		// File should not exist after panic recovery
		_, statErr := os.Stat(path)
		assert.True(t, os.IsNotExist(statErr))
	}()
	//nolint:errcheck // Intentionally not checking error - expecting panic
	_ = op.WriteYAML(path, func() {})
}

// TestJSONMarshalNilValue tests JSON marshal with nil value.
func TestJSONMarshalNilValue(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	data, err := op.Marshal(nil)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

// TestYAMLMarshalNilValue tests YAML marshal with nil value.
func TestYAMLMarshalNilValue(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	data, err := op.Marshal(nil)
	require.NoError(t, err)
	assert.Equal(t, "null\n", string(data))
}

// TestJSONMarshalEmptyStruct tests JSON marshal with empty struct.
func TestJSONMarshalEmptyStruct(t *testing.T) {
	op := NewDefaultJSONOperator(nil)

	type Empty struct{}
	data, err := op.Marshal(Empty{})
	require.NoError(t, err)
	assert.Equal(t, "{}", string(data))
}

// TestYAMLMarshalEmptyStruct tests YAML marshal with empty struct.
func TestYAMLMarshalEmptyStruct(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	type Empty struct{}
	data, err := op.Marshal(Empty{})
	require.NoError(t, err)
	assert.Equal(t, "{}\n", string(data))
}
