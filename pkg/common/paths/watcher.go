// Package paths provides advanced path watching capabilities
package paths

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"path/filepath"
	"sync"
	"time"
)

// Error definitions for watcher operations
var (
	ErrWatcherMockError = errors.New("mock error")
)

// DefaultPathWatcher implements the PathWatcher interface
type DefaultPathWatcher struct {
	mu           sync.RWMutex
	watchedPaths map[string]EventMask
	events       chan *PathEvent
	errors       chan error
	bufferSize   int
	recursive    bool
	debounce     time.Duration
	done         chan struct{}
	cancel       context.CancelFunc
	running      bool
	lastSeen     map[string]time.Time
}

// NewPathWatcher creates a new path watcher with default settings
func NewPathWatcher() PathWatcher {
	done := make(chan struct{})

	return &DefaultPathWatcher{
		watchedPaths: make(map[string]EventMask),
		events:       make(chan *PathEvent, 1000),
		errors:       make(chan error, 100),
		bufferSize:   1000,
		recursive:    true,
		debounce:     100 * time.Millisecond,
		done:         done,
		cancel:       func() { close(done) },
		lastSeen:     make(map[string]time.Time),
	}
}

// Watch starts watching a path for specific events
func (w *DefaultPathWatcher) Watch(path string, events EventMask) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	w.watchedPaths[absPath] = events

	// Start watching if not already running
	if !w.running {
		w.running = true
		go w.watchLoop()
	}

	return nil
}

// WatchPath starts watching a PathBuilder for specific events
func (w *DefaultPathWatcher) WatchPath(path PathBuilder, events EventMask) error {
	return w.Watch(path.String(), events)
}

// Unwatch stops watching a specific path
func (w *DefaultPathWatcher) Unwatch(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	delete(w.watchedPaths, absPath)
	return nil
}

// UnwatchPath stops watching a specific PathBuilder
func (w *DefaultPathWatcher) UnwatchPath(path PathBuilder) error {
	return w.Unwatch(path.String())
}

// Events returns the channel for receiving path events
func (w *DefaultPathWatcher) Events() <-chan *PathEvent {
	return w.events
}

// Errors returns the channel for receiving error events
func (w *DefaultPathWatcher) Errors() <-chan error {
	return w.errors
}

// Close stops the watcher and closes all channels
func (w *DefaultPathWatcher) Close() error {
	w.cancel()

	w.mu.Lock()
	defer w.mu.Unlock()

	w.running = false
	close(w.events)
	close(w.errors)

	return nil
}

// SetBufferSize sets the buffer size for event channels
func (w *DefaultPathWatcher) SetBufferSize(size int) PathWatcher {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.bufferSize = size
	return w
}

// SetRecursive sets whether to watch directories recursively
func (w *DefaultPathWatcher) SetRecursive(recursive bool) PathWatcher {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.recursive = recursive
	return w
}

// SetDebounce sets the debounce duration for events
func (w *DefaultPathWatcher) SetDebounce(duration time.Duration) PathWatcher {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.debounce = duration
	return w
}

// IsWatching returns true if the path is being watched
func (w *DefaultPathWatcher) IsWatching(path string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	_, exists := w.watchedPaths[absPath]
	return exists
}

// WatchedPaths returns all currently watched paths
func (w *DefaultPathWatcher) WatchedPaths() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	paths := make([]string, 0, len(w.watchedPaths))
	for path := range w.watchedPaths {
		paths = append(paths, path)
	}

	return paths
}

// watchLoop is the main watching loop
func (w *DefaultPathWatcher) watchLoop() {
	ticker := time.NewTicker(w.debounce)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ticker.C:
			w.checkForChanges()
		}
	}
}

// checkForChanges checks for file system changes in watched paths
func (w *DefaultPathWatcher) checkForChanges() {
	w.mu.RLock()
	watchedPaths := make(map[string]EventMask)
	for path, events := range w.watchedPaths {
		watchedPaths[path] = events
	}
	w.mu.RUnlock()

	for watchPath, eventMask := range watchedPaths {
		err := filepath.Walk(watchPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				// Log error but continue walking to check other files
				log.Printf("[DEBUG] path watcher: skipping path %s due to error: %v", path, err)
				return nil
			}

			// Skip if not recursive and not in the root directory
			if !w.recursive && filepath.Dir(path) != watchPath {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			w.checkFileForChanges(path, info, eventMask)
			return nil
		})
		if err != nil {
			select {
			case w.errors <- err:
			default:
				// Error channel is full, drop the error
			}
		}
	}
}

// checkFileForChanges checks a specific file for changes
func (w *DefaultPathWatcher) checkFileForChanges(path string, info fs.FileInfo, eventMask EventMask) {
	modTime := info.ModTime()
	lastSeen, exists := w.lastSeen[path]

	if !exists {
		// New file
		w.lastSeen[path] = modTime
		if eventMask&EventCreate != 0 {
			w.emitEvent(&PathEvent{
				Path:   path,
				Op:     EventCreate,
				Time:   time.Now(),
				Info:   info,
				Source: "watcher",
			})
		}
	} else if modTime.After(lastSeen) {
		// Modified file
		w.lastSeen[path] = modTime
		if eventMask&EventWrite != 0 {
			w.emitEvent(&PathEvent{
				Path:   path,
				Op:     EventWrite,
				Time:   time.Now(),
				Info:   info,
				Source: "watcher",
			})
		}
	}
}

// emitEvent sends an event to the events channel
func (w *DefaultPathWatcher) emitEvent(event *PathEvent) {
	select {
	case w.events <- event:
	default:
		// Events channel is full, drop the event
	}
}

// MockPathWatcher implements PathWatcher for testing
type MockPathWatcher struct {
	WatchCalls         []MockWatchCall
	WatchPathCalls     []MockWatchPathCall
	UnwatchCalls       []string
	UnwatchPathCalls   []PathBuilder
	IsWatchingCalls    []string
	WatchedPathsCalls  int
	SetBufferSizeCalls []int
	SetRecursiveCalls  []bool
	SetDebounceCalls   []time.Duration
	ShouldError        bool
	WatchedPathsList   []string
	MockEvents         chan *PathEvent
	MockErrors         chan error
}

// MockWatchCall represents a call to the Watch method in MockPathWatcher
type MockWatchCall struct {
	Path   string
	Events EventMask
}

// MockWatchPathCall represents a call to the WatchPath method in MockPathWatcher
type MockWatchPathCall struct {
	Path   PathBuilder
	Events EventMask
}

// MockPathBuilderForWatcher is a simple interface for testing

// NewMockPathWatcher creates a new mock path watcher for testing
func NewMockPathWatcher() *MockPathWatcher {
	return &MockPathWatcher{
		WatchCalls:       make([]MockWatchCall, 0),
		WatchPathCalls:   make([]MockWatchPathCall, 0),
		UnwatchCalls:     make([]string, 0),
		UnwatchPathCalls: make([]PathBuilder, 0),
		IsWatchingCalls:  make([]string, 0),
		MockEvents:       make(chan *PathEvent, 100),
		MockErrors:       make(chan error, 100),
	}
}

// Watch starts watching a path for events in the mock watcher
func (m *MockPathWatcher) Watch(path string, events EventMask) error {
	m.WatchCalls = append(m.WatchCalls, MockWatchCall{Path: path, Events: events})
	if m.ShouldError {
		return ErrWatcherMockError
	}
	return nil
}

// WatchPath starts watching a PathBuilder for events in the mock watcher
func (m *MockPathWatcher) WatchPath(path PathBuilder, events EventMask) error {
	m.WatchPathCalls = append(m.WatchPathCalls, MockWatchPathCall{Path: path, Events: events})
	if m.ShouldError {
		return ErrWatcherMockError
	}
	return nil
}

// Unwatch stops watching a path in the mock watcher
func (m *MockPathWatcher) Unwatch(path string) error {
	m.UnwatchCalls = append(m.UnwatchCalls, path)
	if m.ShouldError {
		return ErrWatcherMockError
	}
	return nil
}

// UnwatchPath stops watching a PathBuilder in the mock watcher
func (m *MockPathWatcher) UnwatchPath(path PathBuilder) error {
	m.UnwatchPathCalls = append(m.UnwatchPathCalls, path)
	if m.ShouldError {
		return ErrWatcherMockError
	}
	return nil
}

// Events returns the events channel from the mock watcher
func (m *MockPathWatcher) Events() <-chan *PathEvent {
	return m.MockEvents
}

// Errors returns the errors channel from the mock watcher
func (m *MockPathWatcher) Errors() <-chan error {
	return m.MockErrors
}

// Close closes the mock watcher and its channels
func (m *MockPathWatcher) Close() error {
	if m.ShouldError {
		return ErrWatcherMockError
	}
	close(m.MockEvents)
	close(m.MockErrors)
	return nil
}

// SetBufferSize sets the buffer size for the mock watcher
func (m *MockPathWatcher) SetBufferSize(size int) PathWatcher {
	m.SetBufferSizeCalls = append(m.SetBufferSizeCalls, size)
	return m
}

// SetRecursive sets recursive watching for the mock watcher
func (m *MockPathWatcher) SetRecursive(recursive bool) PathWatcher {
	m.SetRecursiveCalls = append(m.SetRecursiveCalls, recursive)
	return m
}

// SetDebounce sets the debounce duration for the mock watcher
func (m *MockPathWatcher) SetDebounce(duration time.Duration) PathWatcher {
	m.SetDebounceCalls = append(m.SetDebounceCalls, duration)
	return m
}

// IsWatching checks if a path is being watched by the mock watcher
func (m *MockPathWatcher) IsWatching(path string) bool {
	m.IsWatchingCalls = append(m.IsWatchingCalls, path)
	for _, watched := range m.WatchedPathsList {
		if watched == path {
			return true
		}
	}
	return false
}

// WatchedPaths returns all paths being watched by the mock watcher
func (m *MockPathWatcher) WatchedPaths() []string {
	m.WatchedPathsCalls++
	return m.WatchedPathsList
}
