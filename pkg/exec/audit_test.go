package exec

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static errors for err113 compliance
var errAuditLogFailed = errors.New("audit log failed")

// mockAuditLogger captures audit events for testing
type mockAuditLogger struct {
	mu     sync.Mutex
	events []AuditEvent
}

func (m *mockAuditLogger) LogEvent(event AuditEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockAuditLogger) getEvents() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]AuditEvent, len(m.events))
	copy(result, m.events)
	return result
}

// failingAuditLogger always returns an error
type failingAuditLogger struct{}

func (f *failingAuditLogger) LogEvent(_ AuditEvent) error {
	return errAuditLogFailed
}

func TestAuditingExecutor_Execute(t *testing.T) {
	t.Run("logs successful command execution", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		err := executor.Execute(ctx, "echo", "hello")

		require.NoError(t, err)

		events := logger.getEvents()
		require.Len(t, events, 1)

		event := events[0]
		assert.Equal(t, "echo", event.Command)
		assert.Equal(t, []string{"hello"}, event.Args)
		assert.True(t, event.Success)
		assert.Equal(t, 0, event.ExitCode)
		assert.NotZero(t, event.Timestamp)
		assert.NotZero(t, event.Duration)
	})

	t.Run("logs failed command execution", func(t *testing.T) {
		base := NewBase() // Not dry run - will actually try to execute
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		err := executor.Execute(ctx, "nonexistentcommand12345")

		require.Error(t, err)

		events := logger.getEvents()
		require.Len(t, events, 1)

		event := events[0]
		assert.Equal(t, "nonexistentcommand12345", event.Command)
		assert.False(t, event.Success)
		assert.NotEqual(t, 0, event.ExitCode)
	})

	t.Run("continues on audit log failure", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		executor := NewAuditingExecutor(base, WithAuditLogger(&failingAuditLogger{}))

		ctx := context.Background()
		err := executor.Execute(ctx, "echo", "hello")

		// Should not fail even though audit logging failed
		require.NoError(t, err)
	})

	t.Run("skips logging when no logger configured", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		executor := NewAuditingExecutor(base) // No logger

		ctx := context.Background()
		err := executor.Execute(ctx, "echo", "hello")

		require.NoError(t, err)
	})
}

func TestAuditingExecutor_ExecuteOutput(t *testing.T) {
	t.Run("logs command with output", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		output, err := executor.ExecuteOutput(ctx, "echo", "hello")

		require.NoError(t, err)
		assert.Empty(t, output) // Dry run returns empty

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "echo", events[0].Command)
		assert.True(t, events[0].Success)
	})
}

func TestAuditingExecutor_Options(t *testing.T) {
	t.Run("WithAuditWorkingDir sets working directory", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base,
			WithAuditLogger(logger),
			WithAuditWorkingDir("/test/dir"),
		)

		ctx := context.Background()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "/test/dir", events[0].WorkingDir)
	})

	t.Run("WithAuditDryRun adds dry run metadata", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base,
			WithAuditLogger(logger),
			WithAuditDryRun(true),
		)

		ctx := context.Background()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "true", events[0].Metadata["dry_run"])
	})

	t.Run("WithAuditMetadata adds custom metadata", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base,
			WithAuditLogger(logger),
			WithAuditMetadata("build_id", "12345"),
			WithAuditMetadata("version", "1.0.0"),
		)

		ctx := context.Background()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "12345", events[0].Metadata["build_id"])
		assert.Equal(t, "1.0.0", events[0].Metadata["version"])
	})
}

func TestAuditEvent_Fields(t *testing.T) {
	t.Run("captures user", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")

		events := logger.getEvents()
		require.Len(t, events, 1)
		// User should be set (either actual username or "unknown")
		assert.NotEmpty(t, events[0].User)
	})

	t.Run("captures duration", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		start := time.Now()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")
		elapsed := time.Since(start)

		events := logger.getEvents()
		require.Len(t, events, 1)
		// Duration should be reasonable (less than elapsed time + buffer)
		assert.LessOrEqual(t, events[0].Duration, elapsed+time.Millisecond*100)
	})

	t.Run("includes environment subset", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		//nolint:errcheck // Testing audit capture, not execution success
		_ = executor.Execute(ctx, "echo", "hello")

		events := logger.getEvents()
		require.Len(t, events, 1)
		// Environment map should exist
		assert.NotNil(t, events[0].Environment)
	})
}

func TestDefaultAuditLogger(t *testing.T) {
	t.Run("LogEvent returns nil", func(t *testing.T) {
		logger := &DefaultAuditLogger{}
		event := AuditEvent{
			Timestamp: time.Now(),
			Command:   "echo",
			Args:      []string{"test"},
			Success:   true,
		}

		err := logger.LogEvent(event)
		assert.NoError(t, err)
	})
}

func TestAuditingExecutor_ImplementsInterface(t *testing.T) {
	base := NewBase()
	executor := NewAuditingExecutor(base)

	var _ Executor = executor
}

func TestAuditingExecutor_ExecuteWithEnv(t *testing.T) {
	t.Run("logs command with environment", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		err := executor.ExecuteWithEnv(ctx, []string{"TEST_VAR=value"}, "echo", "hello")

		require.NoError(t, err)

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "echo", events[0].Command)
		assert.Equal(t, []string{"hello"}, events[0].Args)
		assert.True(t, events[0].Success)
	})
}

func TestAuditingExecutor_ExecuteInDir(t *testing.T) {
	t.Run("logs command with directory", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		err := executor.ExecuteInDir(ctx, "/tmp", "echo", "hello")

		require.NoError(t, err)

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "echo", events[0].Command)
		assert.True(t, events[0].Success)
	})
}

func TestAuditingExecutor_ExecuteOutputInDir(t *testing.T) {
	t.Run("logs command with output and directory", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		output, err := executor.ExecuteOutputInDir(ctx, "/tmp", "echo", "hello")

		require.NoError(t, err)
		assert.Empty(t, output) // Dry run returns empty

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "echo", events[0].Command)
		assert.True(t, events[0].Success)
	})
}

func TestAuditingExecutor_ExecuteStreaming(t *testing.T) {
	t.Run("logs streaming command", func(t *testing.T) {
		base := NewBase(WithDryRun(true))
		logger := &mockAuditLogger{}
		executor := NewAuditingExecutor(base, WithAuditLogger(logger))

		ctx := context.Background()
		var stdout, stderr strings.Builder
		err := executor.ExecuteStreaming(ctx, &stdout, &stderr, "echo", "hello")

		require.NoError(t, err)

		events := logger.getEvents()
		require.Len(t, events, 1)
		assert.Equal(t, "echo", events[0].Command)
		assert.Equal(t, []string{"hello"}, events[0].Args)
		assert.True(t, events[0].Success)
	})
}

func BenchmarkAuditingExecutor(b *testing.B) {
	base := NewBase(WithDryRun(true))
	logger := &mockAuditLogger{}
	executor := NewAuditingExecutor(base, WithAuditLogger(logger))

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//nolint:errcheck // Benchmarking execution, not checking errors
		_ = executor.Execute(ctx, "echo", "benchmark")
	}
}
