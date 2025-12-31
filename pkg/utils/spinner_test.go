package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpinner_Basic(t *testing.T) {
	t.Run("NewSpinner creates spinner with default style", func(t *testing.T) {
		spinner := NewSpinner("Loading...")
		assert.Equal(t, "Loading...", spinner.message)
		registry := newDefaultSpinnerFrameRegistry()
		assert.Equal(t, registry.GetFrames()[SpinnerStyleDots], spinner.frames)
		assert.Equal(t, 100*time.Millisecond, spinner.delay)
		assert.False(t, spinner.active)
		assert.False(t, spinner.paused)
	})

	t.Run("NewSpinnerWithStyle creates spinner with specific style", func(t *testing.T) {
		spinner := NewSpinnerWithStyle("Processing...", SpinnerStyleLine)
		assert.Equal(t, "Processing...", spinner.message)
		registry := newDefaultSpinnerFrameRegistry()
		assert.Equal(t, registry.GetFrames()[SpinnerStyleLine], spinner.frames)
		assert.False(t, spinner.active)
	})

	t.Run("NewSpinnerWithStyle with invalid style uses default", func(t *testing.T) {
		invalidStyle := SpinnerStyle(999)
		spinner := NewSpinnerWithStyle("Test", invalidStyle)
		registry := newDefaultSpinnerFrameRegistry()
		assert.Equal(t, registry.GetFrames()[SpinnerStyleDots], spinner.frames)
	})
}

func TestSpinner_Lifecycle(t *testing.T) {
	t.Run("Start and Stop", func(t *testing.T) {
		spinner := NewSpinner("Test")

		// Start spinner
		spinner.Start()
		assert.True(t, spinner.active)

		// Allow it to run briefly
		time.Sleep(50 * time.Millisecond)

		// Stop spinner
		spinner.Stop()
		assert.False(t, spinner.active)
	})

	t.Run("Multiple Start calls", func(t *testing.T) {
		spinner := NewSpinner("Test")

		// First start
		spinner.Start()
		assert.True(t, spinner.active)

		// Second start should not change state
		spinner.Start()
		assert.True(t, spinner.active)

		spinner.Stop()
	})

	t.Run("Stop when not active", func(t *testing.T) {
		spinner := NewSpinner("Test")

		// Stop when not started - should not panic
		assert.NotPanics(t, func() {
			spinner.Stop()
		})
		assert.False(t, spinner.active)
	})
}

func TestSpinner_PauseResume(t *testing.T) {
	t.Run("Pause and Resume", func(t *testing.T) {
		spinner := NewSpinner("Test")
		spinner.Start()

		// Allow it to run
		time.Sleep(50 * time.Millisecond)

		// Pause
		spinner.Pause()
		assert.True(t, spinner.paused)

		// Resume
		spinner.Resume()
		assert.False(t, spinner.paused)

		spinner.Stop()
	})

	t.Run("Pause when not active", func(t *testing.T) {
		spinner := NewSpinner("Test")

		// Pause when not started - should not panic
		assert.NotPanics(t, func() {
			spinner.Pause()
		})
		assert.False(t, spinner.paused)
	})

	t.Run("Resume when not paused", func(t *testing.T) {
		spinner := NewSpinner("Test")
		spinner.Start()

		// Resume when not paused - should not panic
		assert.NotPanics(t, func() {
			spinner.Resume()
		})

		spinner.Stop()
	})

	t.Run("Multiple pause calls", func(t *testing.T) {
		spinner := NewSpinner("Test")
		spinner.Start()

		// First pause
		spinner.Pause()
		assert.True(t, spinner.paused)

		// Second pause - should not change state
		spinner.Pause()
		assert.True(t, spinner.paused)

		spinner.Resume()
		spinner.Stop()
	})
}

func TestSpinner_MessageUpdate(t *testing.T) {
	t.Run("UpdateMessage changes message", func(t *testing.T) {
		spinner := NewSpinner("Initial")
		assert.Equal(t, "Initial", spinner.message)

		spinner.UpdateMessage("Updated")
		assert.Equal(t, "Updated", spinner.message)
	})

	t.Run("UpdateMessage while running", func(t *testing.T) {
		spinner := NewSpinner("Initial")
		spinner.Start()

		time.Sleep(50 * time.Millisecond)

		spinner.UpdateMessage("Running Update")
		assert.Equal(t, "Running Update", spinner.message)

		spinner.Stop()
	})
}

func TestSpinner_DelayConfiguration(t *testing.T) {
	t.Run("SetDelay changes animation delay", func(t *testing.T) {
		spinner := NewSpinner("Test")
		originalDelay := spinner.delay

		newDelay := 200 * time.Millisecond
		spinner.SetDelay(newDelay)
		assert.Equal(t, newDelay, spinner.delay)
		assert.NotEqual(t, originalDelay, spinner.delay)
	})
}

func TestSpinnerStyles(t *testing.T) {
	styles := []SpinnerStyle{
		SpinnerStyleDots,
		SpinnerStyleLine,
		SpinnerStyleCircle,
		SpinnerStyleSquare,
		SpinnerStyleArrow,
		SpinnerStyleBounce,
	}

	for _, style := range styles {
		t.Run(string(rune(style)), func(t *testing.T) {
			spinner := NewSpinnerWithStyle("Test", style)
			registry := newDefaultSpinnerFrameRegistry()
			expectedFrames := registry.GetFrames()[style]
			assert.Equal(t, expectedFrames, spinner.frames)
			assert.NotEmpty(t, spinner.frames)
		})
	}
}

func TestMultiSpinner(t *testing.T) {
	t.Run("NewMultiSpinner creates empty multi-spinner", func(t *testing.T) {
		ms := NewMultiSpinner()
		assert.NotNil(t, ms)
		assert.NotNil(t, ms.spinners)
		assert.Empty(t, ms.spinners)
		assert.False(t, ms.active)
	})

	t.Run("AddTask adds task to multi-spinner", func(t *testing.T) {
		ms := NewMultiSpinner()

		ms.AddTask("task1", "First task")
		assert.Len(t, ms.spinners, 1)

		task := ms.spinners["task1"]
		require.NotNil(t, task)
		assert.Equal(t, "task1", task.name)
		assert.Equal(t, "First task", task.message)
		assert.Equal(t, TaskStatusPending, task.status)
	})

	t.Run("UpdateTask updates task status and message", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "First task")

		ms.UpdateTask("task1", TaskStatusRunning, "Running task")

		task := ms.spinners["task1"]
		require.NotNil(t, task)
		assert.Equal(t, TaskStatusRunning, task.status)
		assert.Equal(t, "Running task", task.message)
	})

	t.Run("UpdateTask with empty message keeps original", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "Original message")

		ms.UpdateTask("task1", TaskStatusRunning, "")

		task := ms.spinners["task1"]
		require.NotNil(t, task)
		assert.Equal(t, TaskStatusRunning, task.status)
		assert.Equal(t, "Original message", task.message)
	})

	t.Run("UpdateTask for non-existent task", func(t *testing.T) {
		ms := NewMultiSpinner()

		// Should not panic
		assert.NotPanics(t, func() {
			ms.UpdateTask("nonexistent", TaskStatusRunning, "test")
		})
	})

	t.Run("Start and Stop multi-spinner", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "Test task")

		// Start
		ms.Start()
		assert.True(t, ms.active)

		// Allow it to run briefly
		time.Sleep(50 * time.Millisecond)

		// Stop
		ms.Stop()
		assert.False(t, ms.active)
	})

	t.Run("Multiple Start calls on multi-spinner", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "Test task")

		// First start
		ms.Start()
		assert.True(t, ms.active)

		// Second start should not change state
		ms.Start()
		assert.True(t, ms.active)

		ms.Stop()
	})

	t.Run("Stop when not active", func(t *testing.T) {
		ms := NewMultiSpinner()

		// Stop when not started - should not panic
		assert.NotPanics(t, func() {
			ms.Stop()
		})
		assert.False(t, ms.active)
	})
}

func TestTaskStatus(t *testing.T) {
	// Test that all task statuses are defined
	statuses := []TaskStatus{
		TaskStatusPending,
		TaskStatusRunning,
		TaskStatusSuccess,
		TaskStatusFailed,
	}

	for i, status := range statuses {
		assert.Equal(t, TaskStatus(i), status)
	}
}

func TestProgressTree(t *testing.T) {
	t.Run("NewProgressTree creates tree with root", func(t *testing.T) {
		tree := NewProgressTree("Root Task")
		assert.NotNil(t, tree)
		assert.NotNil(t, tree.root)
		assert.Equal(t, "Root Task", tree.root.name)
		assert.Equal(t, TaskStatusPending, tree.root.status)
		assert.NotNil(t, tree.renderer)
	})

	t.Run("AddTask adds child task", func(t *testing.T) {
		tree := NewProgressTree("Root")

		tree.AddTask("Root", "Child Task", 100)

		assert.Len(t, tree.root.children, 1)
		child := tree.root.children[0]
		assert.Equal(t, "Child Task", child.name)
		assert.Equal(t, 100, child.total)
		assert.Equal(t, TaskStatusPending, child.status)
		assert.Equal(t, tree.root, child.parent)
	})

	t.Run("AddTask with non-existent parent uses root", func(t *testing.T) {
		tree := NewProgressTree("Root")

		tree.AddTask("NonExistent", "Child Task", 50)

		assert.Len(t, tree.root.children, 1)
		child := tree.root.children[0]
		assert.Equal(t, "Child Task", child.name)
		assert.Equal(t, tree.root, child.parent)
	})

	t.Run("UpdateTask updates progress and status", func(t *testing.T) {
		tree := NewProgressTree("Root")
		tree.AddTask("Root", "Child", 100)

		tree.UpdateTask("Child", 50, TaskStatusRunning)

		child := tree.root.children[0]
		assert.Equal(t, 50, child.progress)
		assert.Equal(t, TaskStatusRunning, child.status)
	})

	t.Run("UpdateTask for non-existent task", func(t *testing.T) {
		tree := NewProgressTree("Root")

		// Should not panic
		assert.NotPanics(t, func() {
			tree.UpdateTask("NonExistent", 50, TaskStatusRunning)
		})
	})

	t.Run("Render doesn't panic", func(t *testing.T) {
		tree := NewProgressTree("Root")
		tree.AddTask("Root", "Child", 100)
		tree.UpdateTask("Child", 50, TaskStatusRunning)

		// Should not panic - we can't easily test the output
		assert.NotPanics(t, func() {
			tree.Render()
		})
	})

	t.Run("findNode finds correct node", func(t *testing.T) {
		tree := NewProgressTree("Root")
		tree.AddTask("Root", "Child1", 100)
		tree.AddTask("Root", "Child2", 50)

		// Test finding root
		found := tree.findNode(tree.root, "Root")
		assert.Equal(t, tree.root, found)

		// Test finding children
		found = tree.findNode(tree.root, "Child1")
		assert.NotNil(t, found)
		assert.Equal(t, "Child1", found.name)

		found = tree.findNode(tree.root, "Child2")
		assert.NotNil(t, found)
		assert.Equal(t, "Child2", found.name)

		// Test finding non-existent
		found = tree.findNode(tree.root, "NonExistent")
		assert.Nil(t, found)
	})
}

func TestTreeSymbols(t *testing.T) {
	t.Run("unicode tree symbols are defined", func(t *testing.T) {
		registry := newDefaultTreeSymbolRegistry()
		unicodeSymbols := registry.GetUnicodeSymbols()
		assert.NotEmpty(t, unicodeSymbols.branch)
		assert.NotEmpty(t, unicodeSymbols.lastBranch)
		assert.NotEmpty(t, unicodeSymbols.vertical)
		assert.NotEmpty(t, unicodeSymbols.empty)
	})

	t.Run("ascii tree symbols are defined", func(t *testing.T) {
		registry := newDefaultTreeSymbolRegistry()
		asciiSymbols := registry.GetASCIISymbols()
		assert.NotEmpty(t, asciiSymbols.branch)
		assert.NotEmpty(t, asciiSymbols.lastBranch)
		assert.NotEmpty(t, asciiSymbols.vertical)
		assert.NotEmpty(t, asciiSymbols.empty)
	})

	t.Run("GetSymbols returns correct type", func(t *testing.T) {
		registry := NewTreeSymbolRegistry()

		// Unicode type
		unicodeSymbols := registry.GetSymbols(TreeSymbolTypeUnicode)
		assert.NotEmpty(t, unicodeSymbols.branch)

		// ASCII type
		asciiSymbols := registry.GetSymbols(TreeSymbolTypeASCII)
		assert.NotEmpty(t, asciiSymbols.branch)
		assert.NotEqual(t, unicodeSymbols.branch, asciiSymbols.branch)

		// Default (invalid) type should return unicode
		defaultSymbols := registry.GetSymbols(TreeSymbolType(999))
		assert.Equal(t, unicodeSymbols.branch, defaultSymbols.branch)
	})
}

func TestProgressTreeWithCustomRegistry(t *testing.T) {
	t.Run("NewProgressTreeWithRegistry creates tree with custom registry", func(t *testing.T) {
		registry := NewTreeSymbolRegistry()
		tree := NewProgressTreeWithRegistry("Root Task", registry, TreeSymbolTypeASCII)
		assert.NotNil(t, tree)
		assert.NotNil(t, tree.root)
		assert.Equal(t, "Root Task", tree.root.name)
		assert.NotNil(t, tree.renderer)
	})

	t.Run("Progress tree with nested children", func(t *testing.T) {
		tree := NewProgressTree("Root")
		tree.AddTask("Root", "Child1", 100)
		tree.AddTask("Child1", "Grandchild1", 50)

		// Update grandchild to completion
		tree.UpdateTask("Grandchild1", 50, TaskStatusSuccess)

		// Verify parent progress update logic is triggered
		assert.NotPanics(t, func() {
			tree.Render()
		})
	})

	t.Run("Progress tree with multiple children at same level", func(t *testing.T) {
		tree := NewProgressTree("Root")
		tree.AddTask("Root", "Child1", 100)
		tree.AddTask("Root", "Child2", 100)
		tree.AddTask("Root", "Child3", 100)

		// Update all children to completion
		tree.UpdateTask("Child1", 100, TaskStatusSuccess)
		tree.UpdateTask("Child2", 100, TaskStatusSuccess)
		tree.UpdateTask("Child3", 100, TaskStatusSuccess)

		assert.NotPanics(t, func() {
			tree.Render()
		})
	})
}

// Integration test for complex spinner workflow
func TestSpinnerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("complex spinner workflow", func(t *testing.T) {
		spinner := NewSpinner("Initial task")

		// Start spinner
		spinner.Start()
		time.Sleep(20 * time.Millisecond)

		// Update message
		spinner.UpdateMessage("Processing...")
		time.Sleep(20 * time.Millisecond)

		// Pause and resume
		spinner.Pause()
		time.Sleep(10 * time.Millisecond)
		spinner.Resume()
		time.Sleep(20 * time.Millisecond)

		// Final update and stop
		spinner.UpdateMessage("Completing...")
		time.Sleep(20 * time.Millisecond)
		spinner.Stop()

		// Verify final state
		assert.False(t, spinner.active)
		assert.False(t, spinner.paused)
		assert.Equal(t, "Completing...", spinner.message)
	})

	t.Run("multi-spinner workflow", func(t *testing.T) {
		ms := NewMultiSpinner()

		// Add multiple tasks
		ms.AddTask("compile", "Compiling source code")
		ms.AddTask("test", "Running tests")
		ms.AddTask("build", "Building artifacts")

		// Start multi-spinner
		ms.Start()
		time.Sleep(50 * time.Millisecond)

		// Update tasks through different states
		ms.UpdateTask("compile", TaskStatusRunning, "Compiling modules...")
		time.Sleep(30 * time.Millisecond)

		ms.UpdateTask("compile", TaskStatusSuccess, "Compilation complete")
		ms.UpdateTask("test", TaskStatusRunning, "Running unit tests...")
		time.Sleep(30 * time.Millisecond)

		ms.UpdateTask("test", TaskStatusSuccess, "All tests passed")
		ms.UpdateTask("build", TaskStatusRunning, "Creating build artifacts...")
		time.Sleep(30 * time.Millisecond)

		ms.UpdateTask("build", TaskStatusSuccess, "Build complete")
		time.Sleep(20 * time.Millisecond)

		// Stop multi-spinner
		ms.Stop()

		// Verify final states
		assert.False(t, ms.active)
		assert.Equal(t, TaskStatusSuccess, ms.spinners["compile"].status)
		assert.Equal(t, TaskStatusSuccess, ms.spinners["test"].status)
		assert.Equal(t, TaskStatusSuccess, ms.spinners["build"].status)
	})
}

// Benchmark tests
func BenchmarkSpinner_StartStop(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spinner := NewSpinner("Benchmark test")
		spinner.Start()
		spinner.Stop()
	}
}

func BenchmarkSpinner_UpdateMessage(b *testing.B) {
	spinner := NewSpinner("Initial")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spinner.UpdateMessage("Updated message")
	}
}

func BenchmarkMultiSpinner_AddTask(b *testing.B) {
	ms := NewMultiSpinner()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.AddTask(string(rune(i)), "Task")
	}
}

func BenchmarkMultiSpinner_UpdateTask(b *testing.B) {
	ms := NewMultiSpinner()
	ms.AddTask("task", "Test task")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.UpdateTask("task", TaskStatusRunning, "Running")
	}
}
