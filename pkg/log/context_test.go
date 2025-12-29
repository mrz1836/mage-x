package log

import (
	"context"
	"sync"
	"testing"
)

func TestGetRequestIDFromContext_AllKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      interface{}
		value    string
		expected string
	}{
		{
			name:     "request_id key",
			key:      "request_id",
			value:    "req-123",
			expected: "req-123",
		},
		{
			name:     "requestId key",
			key:      "requestId",
			value:    "req-456",
			expected: "req-456",
		},
		{
			name:     "req_id key",
			key:      "req_id",
			value:    "req-789",
			expected: "req-789",
		},
		{
			name:     "trace_id key",
			key:      "trace_id",
			value:    "trace-abc",
			expected: "trace-abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.WithValue(context.Background(), tt.key, tt.value)
			got := GetRequestIDFromContext(ctx)
			if got != tt.expected {
				t.Errorf("GetRequestIDFromContext() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetRequestIDFromContext_NilContext(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1012: Intentionally testing nil context handling
	got := GetRequestIDFromContext(nil)
	if got != "" {
		t.Errorf("GetRequestIDFromContext(nil) = %q, want empty string", got)
	}
}

func TestGetRequestIDFromContext_EmptyContext(t *testing.T) {
	t.Parallel()

	got := GetRequestIDFromContext(context.Background())
	if got != "" {
		t.Errorf("GetRequestIDFromContext(empty) = %q, want empty string", got)
	}
}

func TestGetRequestIDFromContext_EmptyString(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx := context.WithValue(context.Background(), "request_id", "")
	got := GetRequestIDFromContext(ctx)
	if got != "" {
		t.Errorf("GetRequestIDFromContext(empty value) = %q, want empty string", got)
	}
}

func TestGetRequestIDFromContext_NonStringValue(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx := context.WithValue(context.Background(), "request_id", 12345)
	got := GetRequestIDFromContext(ctx)
	if got != "" {
		t.Errorf("GetRequestIDFromContext(non-string) = %q, want empty string", got)
	}
}

func TestGetRequestIDFromContext_Priority(t *testing.T) {
	t.Parallel()

	// When multiple keys are set, the first matching key should win
	// Order of keys: "request_id", "requestId", "req_id", "trace_id"
	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx := context.WithValue(context.Background(), "trace_id", "trace-last")
	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx = context.WithValue(ctx, "request_id", "req-first")

	got := GetRequestIDFromContext(ctx)
	if got != "req-first" {
		t.Errorf("GetRequestIDFromContext() = %q, want %q (first key should win)", got, "req-first")
	}
}

func TestGetRequestIDFromContext_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx := context.WithValue(context.Background(), "request_id", "concurrent-req")

	var wg sync.WaitGroup
	const numGoroutines = 100

	results := make([]string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = GetRequestIDFromContext(ctx)
		}(i)
	}

	wg.Wait()

	for i, result := range results {
		if result != "concurrent-req" {
			t.Errorf("goroutine %d: GetRequestIDFromContext() = %q, want %q", i, result, "concurrent-req")
		}
	}
}

func TestCopyFields(t *testing.T) {
	t.Parallel()

	t.Run("nil source", func(t *testing.T) {
		t.Parallel()
		result := copyFields(nil)
		if result == nil {
			t.Error("copyFields(nil) should return non-nil empty map")
		}
		if len(result) != 0 {
			t.Errorf("copyFields(nil) should return empty map, got %d items", len(result))
		}
	})

	t.Run("empty source", func(t *testing.T) {
		t.Parallel()
		src := make(Fields)
		result := copyFields(src)
		if result == nil {
			t.Error("copyFields(empty) should return non-nil empty map")
		}
		if len(result) != 0 {
			t.Errorf("copyFields(empty) should return empty map, got %d items", len(result))
		}
	})

	t.Run("populated source", func(t *testing.T) {
		t.Parallel()
		src := Fields{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		result := copyFields(src)

		if len(result) != 3 {
			t.Errorf("copyFields should copy all fields, got %d items", len(result))
		}

		if result["key1"] != "value1" {
			t.Errorf("key1 = %v, want %v", result["key1"], "value1")
		}
		if result["key2"] != 42 {
			t.Errorf("key2 = %v, want %v", result["key2"], 42)
		}
		if result["key3"] != true {
			t.Errorf("key3 = %v, want %v", result["key3"], true)
		}
	})

	t.Run("deep copy isolation", func(t *testing.T) {
		t.Parallel()
		src := Fields{"key": "original"}
		result := copyFields(src)

		// Modify the copy
		result["key"] = "modified"
		result["new"] = "added"

		// Original should be unchanged
		if src["key"] != "original" {
			t.Errorf("original should be unchanged, got key = %v", src["key"])
		}
		if _, exists := src["new"]; exists {
			t.Error("original should not have new key")
		}
	})
}
