package errors

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealRegistry_NewRegistryHasDefaults verifies new registry has default errors registered
func TestRealRegistry_NewRegistryHasDefaults(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	// Check some default error codes are registered
	defaultCodes := []ErrorCode{
		ErrUnknown,
		ErrInternal,
		ErrInvalidArgument,
		ErrNotFound,
		ErrTimeout,
		ErrBuildFailed,
		ErrTestFailed,
		ErrFileNotFound,
		ErrFileAccessDenied,
		ErrConfigNotFound,
		ErrConfigInvalid,
	}

	for _, code := range defaultCodes {
		assert.True(t, registry.Contains(code), "Registry should contain default code: %s", code)
	}
}

// TestRealRegistry_RegisterDuplicateFails verifies second registration fails
func TestRealRegistry_RegisterDuplicateFails(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	err := registry.Register("CUSTOM_ERROR", "First registration")
	require.NoError(t, err, "First registration should succeed")

	err = registry.Register("CUSTOM_ERROR", "Second registration")
	require.Error(t, err, "Second registration should fail")
	assert.ErrorIs(t, err, errCodeAlreadyRegistered, "Error should be errCodeAlreadyRegistered")
}

// TestRealRegistry_UnregisterNonexistent verifies unregistering unknown code returns error
func TestRealRegistry_UnregisterNonexistent(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	err := registry.Unregister("NONEXISTENT_CODE")
	require.Error(t, err, "Unregistering unknown code should fail")
	assert.ErrorIs(t, err, errCodeNotFound, "Error should be errCodeNotFound")
}

// TestRealRegistry_UnregisterExisting verifies unregistering existing code succeeds
func TestRealRegistry_UnregisterExisting(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	err := registry.Register("TO_BE_REMOVED", "Will be removed")
	require.NoError(t, err)

	assert.True(t, registry.Contains("TO_BE_REMOVED"))

	err = registry.Unregister("TO_BE_REMOVED")
	require.NoError(t, err, "Unregistering existing code should succeed")

	assert.False(t, registry.Contains("TO_BE_REMOVED"), "Code should be removed after unregister")
}

// TestRealRegistry_RegisterWithSeverity verifies severity is properly stored
func TestRealRegistry_RegisterWithSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     ErrorCode
		desc     string
		severity Severity
	}{
		{
			name:     "debug severity",
			code:     "DEBUG_ERROR",
			desc:     "Debug level error",
			severity: SeverityDebug,
		},
		{
			name:     "warning severity",
			code:     "WARNING_ERROR",
			desc:     "Warning level error",
			severity: SeverityWarning,
		},
		{
			name:     "critical severity",
			code:     "CRITICAL_ERROR",
			desc:     "Critical level error",
			severity: SeverityCritical,
		},
		{
			name:     "fatal severity",
			code:     "FATAL_ERROR",
			desc:     "Fatal level error",
			severity: SeverityFatal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := NewErrorRegistry()
			err := registry.RegisterWithSeverity(tt.code, tt.desc, tt.severity)
			require.NoError(t, err)

			def, exists := registry.Get(tt.code)
			require.True(t, exists)
			assert.Equal(t, tt.severity, def.Severity, "Severity should match")
			assert.Equal(t, tt.desc, def.Description, "Description should match")
		})
	}
}

// TestRealRegistry_GetNonexistent verifies Get returns false for unknown code
func TestRealRegistry_GetNonexistent(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	_, exists := registry.Get("NONEXISTENT")
	assert.False(t, exists, "Get should return false for unknown code")
}

// TestRealRegistry_ListReturnsAllSorted verifies List() returns all codes sorted
func TestRealRegistry_ListReturnsAllSorted(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	list := registry.List()

	require.NotEmpty(t, list, "List should not be empty")

	// Verify sorted order
	for i := 1; i < len(list); i++ {
		assert.Less(t, list[i-1].Code, list[i].Code,
			"List should be sorted: %s should be before %s", list[i-1].Code, list[i].Code)
	}
}

// TestRealRegistry_ListByPrefixFilters verifies only matching prefixes returned
func TestRealRegistry_ListByPrefixFilters(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	// Register some errors with different prefixes
	require.NoError(t, registry.Register("TEST_PREFIX_A", "Test A"))
	require.NoError(t, registry.Register("TEST_PREFIX_B", "Test B"))
	require.NoError(t, registry.Register("OTHER_PREFIX_C", "Other C"))

	// Filter by prefix
	testPrefixErrors := registry.ListByPrefix("TEST_PREFIX_")

	assert.Len(t, testPrefixErrors, 2, "Should return 2 errors with TEST_PREFIX_")

	for _, def := range testPrefixErrors {
		assert.Contains(t, string(def.Code), "TEST_PREFIX_",
			"All results should have TEST_PREFIX_ prefix")
	}
}

// TestRealRegistry_ListByPrefixNoMatch verifies empty result for no matches
func TestRealRegistry_ListByPrefixNoMatch(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	result := registry.ListByPrefix("NONEXISTENT_PREFIX_")
	assert.Empty(t, result, "Should return empty list for no matches")
}

// TestRealRegistry_ListBySeverityFilters verifies only matching severities returned
func TestRealRegistry_ListBySeverityFilters(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	// Register errors with specific severities
	require.NoError(t, registry.RegisterWithSeverity("SEV_CRITICAL_1", "Critical 1", SeverityCritical))
	require.NoError(t, registry.RegisterWithSeverity("SEV_CRITICAL_2", "Critical 2", SeverityCritical))
	require.NoError(t, registry.RegisterWithSeverity("SEV_WARNING_1", "Warning 1", SeverityWarning))

	criticalErrors := registry.ListBySeverity(SeverityCritical)

	// Should include our 2 critical errors (plus any defaults)
	criticalCount := 0
	for _, def := range criticalErrors {
		if def.Code == "SEV_CRITICAL_1" || def.Code == "SEV_CRITICAL_2" {
			criticalCount++
		}
		assert.Equal(t, SeverityCritical, def.Severity, "All results should have Critical severity")
	}

	assert.Equal(t, 2, criticalCount, "Should find both registered critical errors")
}

// TestRealRegistry_ListBySeverityNoMatch verifies empty result for no matches
func TestRealRegistry_ListBySeverityNoMatch(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	// Clear and use empty registry
	require.NoError(t, registry.Clear())

	result := registry.ListBySeverity(SeverityFatal)
	assert.Empty(t, result, "Should return empty list for no matches")
}

// TestRealRegistry_ClearRemovesAll verifies Clear() empties registry
func TestRealRegistry_ClearRemovesAll(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	// Verify not empty initially
	require.NotEmpty(t, registry.List(), "Registry should have default errors")

	err := registry.Clear()
	require.NoError(t, err, "Clear should not return error")

	list := registry.List()
	assert.Empty(t, list, "Registry should be empty after Clear")
}

// TestRealRegistry_ContainsAfterClear verifies Contains returns false after Clear
func TestRealRegistry_ContainsAfterClear(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	assert.True(t, registry.Contains(ErrBuildFailed), "Should contain default error before clear")

	require.NoError(t, registry.Clear())

	assert.False(t, registry.Contains(ErrBuildFailed), "Should not contain error after clear")
}

// TestRealRegistry_ExtractCategory verifies extractCategory() parses codes correctly
func TestRealRegistry_ExtractCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		code     ErrorCode
		expected string
	}{
		{
			name:     "BUILD_FAILED extracts build",
			code:     "ERR_BUILD_FAILED",
			expected: "build",
		},
		{
			name:     "FILE_NOT_FOUND extracts file",
			code:     "ERR_FILE_NOT_FOUND",
			expected: "file",
		},
		{
			name:     "single word returns general",
			code:     "UNKNOWN",
			expected: "general",
		},
		{
			name:     "CONFIG_INVALID extracts config",
			code:     "ERR_CONFIG_INVALID",
			expected: "config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractCategory(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRealRegistry_IsRetryable verifies isRetryable() identifies retryable codes
func TestRealRegistry_IsRetryable(t *testing.T) {
	t.Parallel()

	retryableCodes := []ErrorCode{
		ErrInternal,
		ErrTimeout,
		ErrResourceExhausted,
		ErrUnavailable,
		ErrBuildFailed,
		ErrTestFailed,
		ErrCommandTimeout,
	}

	nonRetryableCodes := []ErrorCode{
		ErrInvalidArgument,
		ErrNotFound,
		ErrFileNotFound,
		ErrPermissionDenied,
		ErrConfigInvalid,
	}

	for _, code := range retryableCodes {
		assert.True(t, isRetryable(code), "%s should be retryable", code)
	}

	for _, code := range nonRetryableCodes {
		assert.False(t, isRetryable(code), "%s should not be retryable", code)
	}
}

// TestRealRegistry_ExtractTags verifies extractTags() creates proper tags
func TestRealRegistry_ExtractTags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		code         ErrorCode
		expectedTags []string
	}{
		{
			name:         "BUILD_FAILED has failure tag",
			code:         "ERR_BUILD_FAILED",
			expectedTags: []string{"build", "failure"},
		},
		{
			name:         "COMMAND_TIMEOUT has timeout tag",
			code:         "ERR_COMMAND_TIMEOUT",
			expectedTags: []string{"command", "timeout"},
		},
		{
			name:         "INVALID_INPUT has validation tag",
			code:         "ERR_INVALID_INPUT",
			expectedTags: []string{"invalid", "validation"},
		},
		{
			name:         "FILE_NOT_FOUND has missing tag",
			code:         "ERR_FILE_NOT_FOUND",
			expectedTags: []string{"file", "missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractTags(tt.code)
			for _, expected := range tt.expectedTags {
				assert.Contains(t, result, expected, "Tags should contain %s", expected)
			}
		})
	}
}

// TestRealRegistry_Concurrent verifies concurrent read/write operations are safe
func TestRealRegistry_Concurrent(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			code := ErrorCode("CONCURRENT_" + string(rune('A'+idx%26)))
			_ = registry.Register(code, "Concurrent error") //nolint:errcheck // intentionally ignoring duplicate registration errors
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.List()
			_ = registry.ListByPrefix("ERR_")
			_ = registry.ListBySeverity(SeverityError)
			_ = registry.Contains(ErrBuildFailed)
			_, _ = registry.Get(ErrBuildFailed)
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

// TestRealRegistry_ErrorDefinitionFields verifies all ErrorDefinition fields are set
func TestRealRegistry_ErrorDefinitionFields(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	err := registry.RegisterWithSeverity("TEST_FULL_DEF", "Full definition test", SeverityCritical)
	require.NoError(t, err)

	def, exists := registry.Get("TEST_FULL_DEF")
	require.True(t, exists)

	assert.Equal(t, ErrorCode("TEST_FULL_DEF"), def.Code)
	assert.Equal(t, "Full definition test", def.Description)
	assert.Equal(t, SeverityCritical, def.Severity)
	assert.NotEmpty(t, def.Category, "Category should be extracted")
	assert.NotNil(t, def.Tags, "Tags should not be nil")
}

// TestRealRegistry_RegisterDefaultSeverity verifies Register uses SeverityError as default
func TestRealRegistry_RegisterDefaultSeverity(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	err := registry.Register("DEFAULT_SEVERITY_TEST", "Test default severity")
	require.NoError(t, err)

	def, exists := registry.Get("DEFAULT_SEVERITY_TEST")
	require.True(t, exists)

	assert.Equal(t, SeverityError, def.Severity, "Default severity should be SeverityError")
}

// TestRealRegistry_ListReturnsCopy verifies List returns a copy, not internal slice
func TestRealRegistry_ListReturnsCopy(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	list1 := registry.List()
	list2 := registry.List()

	require.NotEmpty(t, list1)

	// Modify list1
	list1[0].Description = "MODIFIED"

	// list2 should not be affected
	assert.NotEqual(t, "MODIFIED", list2[0].Description, "List should return copies")
}

// TestRealRegistry_ListByPrefixSorted verifies ListByPrefix returns sorted results
func TestRealRegistry_ListByPrefixSorted(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	require.NoError(t, registry.Register("SORT_TEST_C", "C"))
	require.NoError(t, registry.Register("SORT_TEST_A", "A"))
	require.NoError(t, registry.Register("SORT_TEST_B", "B"))

	result := registry.ListByPrefix("SORT_TEST_")

	require.Len(t, result, 3)
	assert.Equal(t, ErrorCode("SORT_TEST_A"), result[0].Code)
	assert.Equal(t, ErrorCode("SORT_TEST_B"), result[1].Code)
	assert.Equal(t, ErrorCode("SORT_TEST_C"), result[2].Code)
}

// TestRealRegistry_ListBySeveritySorted verifies ListBySeverity returns sorted results
func TestRealRegistry_ListBySeveritySorted(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()
	require.NoError(t, registry.Clear())

	require.NoError(t, registry.RegisterWithSeverity("ZSORT", "Z", SeverityWarning))
	require.NoError(t, registry.RegisterWithSeverity("ASORT", "A", SeverityWarning))
	require.NoError(t, registry.RegisterWithSeverity("MSORT", "M", SeverityWarning))

	result := registry.ListBySeverity(SeverityWarning)

	require.Len(t, result, 3)
	assert.Equal(t, ErrorCode("ASORT"), result[0].Code)
	assert.Equal(t, ErrorCode("MSORT"), result[1].Code)
	assert.Equal(t, ErrorCode("ZSORT"), result[2].Code)
}

// TestRealRegistry_RetryableProperty verifies Retryable property is set correctly
func TestRealRegistry_RetryableProperty(t *testing.T) {
	t.Parallel()

	registry := NewErrorRegistry()

	// Get a known retryable error
	def, exists := registry.Get(ErrTimeout)
	require.True(t, exists)
	assert.True(t, def.Retryable, "ErrTimeout should be marked as retryable")

	// Get a known non-retryable error
	def, exists = registry.Get(ErrNotFound)
	require.True(t, exists)
	assert.False(t, def.Retryable, "ErrNotFound should not be marked as retryable")
}
