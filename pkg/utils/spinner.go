// Package utils provides utility functions for mage tasks
package utils

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Spinner provides an animated progress indicator
type Spinner struct {
	mu       sync.Mutex
	frames   []string
	message  string
	delay    time.Duration
	active   bool
	paused   bool
	stopCh   chan struct{}
	pauseCh  chan struct{}
	resumeCh chan struct{}
	current  int
}

// SpinnerStyle represents different spinner animation styles
type SpinnerStyle int

const (
	// SpinnerStyleDots is the dots spinner style
	SpinnerStyleDots SpinnerStyle = iota
	// SpinnerStyleLine is the line spinner style
	SpinnerStyleLine
	// SpinnerStyleCircle is the circle spinner style
	SpinnerStyleCircle
	// SpinnerStyleSquare is the square spinner style
	SpinnerStyleSquare
	// SpinnerStyleArrow is the arrow spinner style
	SpinnerStyleArrow
	// SpinnerStyleBounce is the bouncing ball spinner style
	SpinnerStyleBounce
)

// SpinnerFrameRegistry provides thread-safe access to spinner frame configurations
type SpinnerFrameRegistry struct {
	once sync.Once
	data map[SpinnerStyle][]string
}

// NewSpinnerFrameRegistry creates a new spinner frame registry
func NewSpinnerFrameRegistry() *SpinnerFrameRegistry {
	return &SpinnerFrameRegistry{}
}

// GetFrames returns the spinner frame configurations with thread-safe initialization
func (r *SpinnerFrameRegistry) GetFrames() map[SpinnerStyle][]string {
	r.once.Do(func() {
		r.data = map[SpinnerStyle][]string{
			SpinnerStyleDots:   {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
			SpinnerStyleLine:   {"-", "\\", "|", "/"},
			SpinnerStyleCircle: {"◐", "◓", "◑", "◒"},
			SpinnerStyleSquare: {"◰", "◳", "◲", "◱"},
			SpinnerStyleArrow:  {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
			SpinnerStyleBounce: {"⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈"},
		}
	})
	return r.data
}

// NewSpinner creates a new spinner with default style
func NewSpinner(message string) *Spinner {
	return NewSpinnerWithStyle(message, SpinnerStyleDots)
}

// NewSpinnerWithStyle creates a new spinner with a specific style
func NewSpinnerWithStyle(message string, style SpinnerStyle) *Spinner {
	return NewSpinnerWithStyleAndRegistry(message, style, newDefaultSpinnerFrameRegistry())
}

// NewSpinnerWithStyleAndRegistry creates a new spinner with a specific style and frame registry
func NewSpinnerWithStyleAndRegistry(message string, style SpinnerStyle, registry *SpinnerFrameRegistry) *Spinner {
	spinnerData := registry.GetFrames()
	frames, ok := spinnerData[style]
	if !ok {
		frames = spinnerData[SpinnerStyleDots]
	}

	return &Spinner{
		frames:   frames,
		message:  message,
		delay:    100 * time.Millisecond,
		stopCh:   make(chan struct{}),
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
	}
}

// newDefaultSpinnerFrameRegistry creates the default registry instance
func newDefaultSpinnerFrameRegistry() *SpinnerFrameRegistry {
	return NewSpinnerFrameRegistry()
}

// Start starts the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go s.animate()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	close(s.stopCh)

	// Clear the spinner line
	if _, err := fmt.Fprint(os.Stdout, "\r\033[K"); err != nil {
		// Continue if write fails
		log.Printf("failed to clear spinner line: %v", err)
	}
}

// Pause temporarily pauses the spinner
func (s *Spinner) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active || s.paused {
		return
	}

	s.paused = true
	s.pauseCh <- struct{}{}
}

// Resume resumes a paused spinner
func (s *Spinner) Resume() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active || !s.paused {
		return
	}

	s.paused = false
	s.resumeCh <- struct{}{}
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// SetDelay sets the animation delay
func (s *Spinner) SetDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delay = delay
}

// animate runs the spinner animation loop
func (s *Spinner) animate() {
	ticker := time.NewTicker(s.delay)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return

		case <-s.pauseCh:
			// Clear the spinner line when pausing
			if _, err := fmt.Fprint(os.Stdout, "\r\033[K"); err != nil {
				// Continue if write fails
				log.Printf("failed to clear spinner line: %v", err)
			}

			// Wait for resume
			<-s.resumeCh

		case <-ticker.C:
			s.mu.Lock()
			frame := s.frames[s.current]
			msg := s.message
			s.current = (s.current + 1) % len(s.frames)
			s.mu.Unlock()

			// Use carriage return to overwrite the line
			if _, err := fmt.Fprintf(os.Stdout, "\r%s %s", frame, msg); err != nil {
				// Continue if write fails
				log.Printf("failed to write spinner frame: %v", err)
			}
		}
	}
}

// MultiSpinner manages multiple spinners for parallel tasks
type MultiSpinner struct {
	mu       sync.Mutex
	spinners map[string]*TaskSpinner
	active   bool
	stopCh   chan struct{}
	registry *SpinnerFrameRegistry
}

// TaskSpinner represents a spinner for a specific task
type TaskSpinner struct {
	name    string
	message string
	status  TaskStatus
	frames  []string
	current int
}

// TaskStatus represents the status of a task
type TaskStatus int

const (
	// TaskStatusPending indicates a pending task
	TaskStatusPending TaskStatus = iota
	// TaskStatusRunning indicates a running task
	TaskStatusRunning
	// TaskStatusSuccess indicates a successful task
	TaskStatusSuccess
	// TaskStatusFailed indicates a failed task
	TaskStatusFailed
)

// NewMultiSpinner creates a new multi-spinner
func NewMultiSpinner() *MultiSpinner {
	return NewMultiSpinnerWithRegistry(newDefaultSpinnerFrameRegistry())
}

// NewMultiSpinnerWithRegistry creates a new multi-spinner with a specific frame registry
func NewMultiSpinnerWithRegistry(registry *SpinnerFrameRegistry) *MultiSpinner {
	return &MultiSpinner{
		spinners: make(map[string]*TaskSpinner),
		stopCh:   make(chan struct{}),
		registry: registry,
	}
}

// AddTask adds a new task to the multi-spinner
func (m *MultiSpinner) AddTask(name, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.spinners[name] = &TaskSpinner{
		name:    name,
		message: message,
		status:  TaskStatusPending,
		frames:  m.registry.GetFrames()[SpinnerStyleDots],
	}
}

// UpdateTask updates a task's status and message
func (m *MultiSpinner) UpdateTask(name string, status TaskStatus, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if spinner, ok := m.spinners[name]; ok {
		spinner.status = status
		if message != "" {
			spinner.message = message
		}
	}
}

// Start starts the multi-spinner animation
func (m *MultiSpinner) Start() {
	m.mu.Lock()
	if m.active {
		m.mu.Unlock()
		return
	}
	m.active = true
	m.mu.Unlock()

	go m.animate()
}

// Stop stops the multi-spinner animation
func (m *MultiSpinner) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return
	}

	m.active = false
	close(m.stopCh)

	// Clear all spinner lines
	for range m.spinners {
		if _, err := fmt.Fprint(os.Stdout, "\033[1A\033[K"); err != nil {
			// Continue if write fails
			log.Printf("failed to clear multiline spinner: %v", err)
		}
	}
}

// animate runs the multi-spinner animation loop
func (m *MultiSpinner) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Initial render
	m.render()

	for {
		select {
		case <-m.stopCh:
			return

		case <-ticker.C:
			m.render()
		}
	}
}

// render renders all spinners
func (m *MultiSpinner) render() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Move cursor to beginning of spinner area
	for i := 0; i < len(m.spinners); i++ {
		if _, err := fmt.Fprint(os.Stdout, "\033[1A"); err != nil {
			// Continue if write fails
			log.Printf("failed to move cursor up: %v", err)
		}
	}

	// Render each spinner
	for _, spinner := range m.spinners {
		var icon string

		switch spinner.status {
		case TaskStatusPending:
			icon = "⏸️ "
		case TaskStatusRunning:
			icon = spinner.frames[spinner.current]
			spinner.current = (spinner.current + 1) % len(spinner.frames)
		case TaskStatusSuccess:
			icon = "✅"
		case TaskStatusFailed:
			icon = "❌"
		}

		// Clear line and print status
		if _, err := fmt.Fprintf(os.Stdout, "\033[K  %s %s: %s\n", icon, spinner.name, spinner.message); err != nil {
			// Continue if write fails
			log.Printf("failed to write multiline spinner status: %v", err)
		}
	}
}

// ProgressTree represents a hierarchical progress display
type ProgressTree struct {
	mu       sync.Mutex
	root     *ProgressNode
	renderer *treeRenderer
}

// ProgressNode represents a node in the progress tree
type ProgressNode struct {
	name     string
	status   TaskStatus
	progress int
	total    int
	children []*ProgressNode
	parent   *ProgressNode
}

// treeRenderer handles rendering of the progress tree
type treeRenderer struct {
	useColor bool
	symbols  treeSymbols
	registry *TreeSymbolRegistry
}

// treeSymbols contains symbols for tree rendering
type treeSymbols struct {
	branch     string
	lastBranch string
	vertical   string
	empty      string
}

// TreeSymbolType represents the style of tree symbols to use
type TreeSymbolType int

const (
	// TreeSymbolTypeUnicode uses Unicode box drawing characters
	TreeSymbolTypeUnicode TreeSymbolType = iota
	// TreeSymbolTypeASCII uses ASCII characters for compatibility
	TreeSymbolTypeASCII
)

// TreeSymbolRegistry provides thread-safe access to tree symbol configurations
type TreeSymbolRegistry struct {
	unicodeOnce sync.Once
	unicodeData treeSymbols
	asciiOnce   sync.Once
	asciiData   treeSymbols
}

// NewTreeSymbolRegistry creates a new tree symbol registry
func NewTreeSymbolRegistry() *TreeSymbolRegistry {
	return &TreeSymbolRegistry{}
}

// GetUnicodeSymbols returns the Unicode tree drawing symbols with thread-safe initialization
func (r *TreeSymbolRegistry) GetUnicodeSymbols() treeSymbols {
	r.unicodeOnce.Do(func() {
		r.unicodeData = treeSymbols{
			branch:     "├─",
			lastBranch: "└─",
			vertical:   "│ ",
			empty:      "  ",
		}
	})
	return r.unicodeData
}

// GetASCIISymbols returns the ASCII tree drawing symbols with thread-safe initialization
func (r *TreeSymbolRegistry) GetASCIISymbols() treeSymbols {
	r.asciiOnce.Do(func() {
		r.asciiData = treeSymbols{
			branch:     "|-",
			lastBranch: "`-",
			vertical:   "| ",
			empty:      "  ",
		}
	})
	return r.asciiData
}

// GetSymbols returns the tree symbols for the specified type
func (r *TreeSymbolRegistry) GetSymbols(symbolType TreeSymbolType) treeSymbols {
	switch symbolType {
	case TreeSymbolTypeUnicode:
		return r.GetUnicodeSymbols()
	case TreeSymbolTypeASCII:
		return r.GetASCIISymbols()
	default:
		return r.GetUnicodeSymbols()
	}
}

// newDefaultTreeSymbolRegistry creates the default tree symbol registry instance
func newDefaultTreeSymbolRegistry() *TreeSymbolRegistry {
	return NewTreeSymbolRegistry()
}

// NewProgressTree creates a new progress tree
func NewProgressTree(name string) *ProgressTree {
	return NewProgressTreeWithRegistry(name, newDefaultTreeSymbolRegistry(), TreeSymbolTypeUnicode)
}

// NewProgressTreeWithRegistry creates a new progress tree with a specific symbol registry and type
func NewProgressTreeWithRegistry(name string, registry *TreeSymbolRegistry, symbolType TreeSymbolType) *ProgressTree {
	return &ProgressTree{
		root: &ProgressNode{
			name:   name,
			status: TaskStatusPending,
		},
		renderer: &treeRenderer{
			useColor: shouldUseColor(),
			symbols:  registry.GetSymbols(symbolType),
			registry: registry,
		},
	}
}

// AddTask adds a task to the progress tree
func (p *ProgressTree) AddTask(parent, name string, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	parentNode := p.findNode(p.root, parent)
	if parentNode == nil {
		parentNode = p.root
	}

	node := &ProgressNode{
		name:   name,
		status: TaskStatusPending,
		total:  total,
		parent: parentNode,
	}

	parentNode.children = append(parentNode.children, node)
}

// UpdateTask updates a task's progress
func (p *ProgressTree) UpdateTask(name string, progress int, status TaskStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()

	node := p.findNode(p.root, name)
	if node != nil {
		node.progress = progress
		node.status = status

		// Update parent progress if all children are complete
		p.updateParentProgress(node.parent)
	}
}

// Render renders the progress tree
func (p *ProgressTree) Render() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, err := fmt.Fprintln(os.Stdout, ""); err != nil {
		// Continue if write fails
		log.Printf("failed to write newline: %v", err)
	}
	p.renderer.renderNode(p.root, "", true)
}

// findNode finds a node by name
func (p *ProgressTree) findNode(node *ProgressNode, name string) *ProgressNode {
	if node.name == name {
		return node
	}

	for _, child := range node.children {
		if found := p.findNode(child, name); found != nil {
			return found
		}
	}

	return nil
}

// updateParentProgress updates parent node progress based on children
func (p *ProgressTree) updateParentProgress(node *ProgressNode) {
	if node == nil || len(node.children) == 0 {
		return
	}

	totalProgress := 0
	totalItems := 0
	allComplete := true
	anyFailed := false

	for _, child := range node.children {
		if child.total > 0 {
			totalProgress += child.progress
			totalItems += child.total
		}

		if child.status != TaskStatusSuccess {
			allComplete = false
		}

		if child.status == TaskStatusFailed {
			anyFailed = true
		}
	}

	if totalItems > 0 {
		node.progress = totalProgress
		node.total = totalItems
	}

	if anyFailed {
		node.status = TaskStatusFailed
	} else if allComplete && len(node.children) > 0 {
		node.status = TaskStatusSuccess
	} else if totalProgress > 0 {
		node.status = TaskStatusRunning
	}

	// Recursively update parent
	p.updateParentProgress(node.parent)
}

// renderNode renders a node and its children
func (r *treeRenderer) renderNode(node *ProgressNode, prefix string, isLast bool) {
	// Render current node
	var statusIcon string
	switch node.status {
	case TaskStatusPending:
		statusIcon = "○"
	case TaskStatusRunning:
		statusIcon = "◐"
	case TaskStatusSuccess:
		statusIcon = "●"
	case TaskStatusFailed:
		statusIcon = "✗"
	}

	// Build the line
	line := prefix
	if node.parent != nil {
		if isLast {
			line += r.symbols.lastBranch
		} else {
			line += r.symbols.branch
		}
	}

	line += fmt.Sprintf(" %s %s", statusIcon, node.name)

	// Add progress bar if applicable
	if node.total > 0 {
		percent := float64(node.progress) / float64(node.total) * 100
		line += fmt.Sprintf(" [%d/%d] %.0f%%", node.progress, node.total, percent)
	}

	if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
		// Continue if write fails
		log.Printf("failed to write fancy line: %v", err)
	}

	// Render children
	childPrefix := prefix
	if node.parent != nil {
		if isLast {
			childPrefix += r.symbols.empty
		} else {
			childPrefix += r.symbols.vertical
		}
	}

	for i, child := range node.children {
		r.renderNode(child, childPrefix, i == len(node.children)-1)
	}
}
