package registry

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Static errors for registry operations
var (
	ErrCommandAlreadyRegistered = errors.New("command already registered")
	ErrAliasAlreadyExists       = errors.New("alias already registered")
	ErrUnknownCommand           = errors.New("unknown command")
)

// Registry manages all available MAGE-X commands
type Registry struct {
	mu         sync.RWMutex
	commands   map[string]*Command
	aliases    map[string]string // alias -> command name mapping
	registered bool              // tracks if commands have been registered

	// Metadata about the registry
	metadata CommandMetadata
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]*Command),
		aliases:  make(map[string]string),
		metadata: CommandMetadata{
			Categories:   make(map[string]int),
			CategoryInfo: make(map[string]CategoryInfo),
		},
	}
}

// Global registry holder
//
//nolint:gochecknoglobals // Necessary for singleton pattern
var globalRegistry struct {
	once     sync.Once
	instance *Registry
}

// Global returns the global registry instance using package-level singleton
func Global() *Registry {
	globalRegistry.once.Do(func() {
		globalRegistry.instance = NewRegistry()
	})

	return globalRegistry.instance
}

// IsRegistered returns whether commands have been registered
func (r *Registry) IsRegistered() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registered
}

// SetRegistered sets the registration status
func (r *Registry) SetRegistered(registered bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registered = registered
}

// Register adds a command to the registry
func (r *Registry) Register(cmd *Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	cmdName := cmd.FullName()

	// Check for duplicates
	if _, exists := r.commands[cmdName]; exists {
		return fmt.Errorf("%w: %s", ErrCommandAlreadyRegistered, cmdName)
	}

	// Register the command
	r.commands[cmdName] = cmd

	// Register aliases
	for _, alias := range cmd.Aliases {
		if existing, exists := r.aliases[alias]; exists {
			return fmt.Errorf("%w: %s for command %s", ErrAliasAlreadyExists, alias, existing)
		}
		r.aliases[alias] = cmdName
	}

	// Update metadata
	r.updateMetadata(cmd)

	return nil
}

// MustRegister adds a command to the registry, panicking on error
func (r *Registry) MustRegister(cmd *Command) {
	if err := r.Register(cmd); err != nil {
		panic(fmt.Sprintf("failed to register command: %v", err))
	}
}

// Get retrieves a command by name or alias
func (r *Registry) Get(name string) (*Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try direct lookup
	if cmd, exists := r.commands[strings.ToLower(name)]; exists {
		return cmd, true
	}

	// Try alias lookup
	if cmdName, exists := r.aliases[strings.ToLower(name)]; exists {
		return r.commands[cmdName], true
	}

	return nil, false
}

// List returns all registered commands
func (r *Registry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commands := make([]*Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		if !cmd.Hidden {
			commands = append(commands, cmd)
		}
	}

	// Sort by namespace and name
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Namespace != commands[j].Namespace {
			return commands[i].Namespace < commands[j].Namespace
		}
		return commands[i].Method < commands[j].Method
	})

	return commands
}

// ListByNamespace returns all commands in a namespace
func (r *Registry) ListByNamespace(namespace string) []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var commands []*Command
	nsLower := strings.ToLower(namespace)

	for _, cmd := range r.commands {
		if strings.ToLower(cmd.Namespace) == nsLower && !cmd.Hidden {
			commands = append(commands, cmd)
		}
	}

	// Sort by method name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Method < commands[j].Method
	})

	return commands
}

// ListByCategory returns all commands in a category
func (r *Registry) ListByCategory(category string) []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var commands []*Command
	for _, cmd := range r.commands {
		if cmd.Category == category && !cmd.Hidden {
			commands = append(commands, cmd)
		}
	}

	// Sort by full name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].FullName() < commands[j].FullName()
	})

	return commands
}

// Namespaces returns all registered namespaces
func (r *Registry) Namespaces() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.namespacesLocked()
}

// namespacesLocked returns all registered namespaces without acquiring the lock.
// Caller must hold at least RLock.
func (r *Registry) namespacesLocked() []string {
	namespaceMap := make(map[string]bool)
	for _, cmd := range r.commands {
		if cmd.Namespace != "" {
			namespaceMap[cmd.Namespace] = true
		}
	}

	namespaces := make([]string, 0, len(namespaceMap))
	for ns := range namespaceMap {
		namespaces = append(namespaces, ns)
	}

	sort.Strings(namespaces)
	return namespaces
}

// Categories returns all registered categories
func (r *Registry) Categories() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categories := make([]string, 0, len(r.metadata.Categories))
	for cat := range r.metadata.Categories {
		categories = append(categories, cat)
	}

	sort.Strings(categories)
	return categories
}

// CategorizedCommands returns commands organized by category with ordering
func (r *Registry) CategorizedCommands() map[string][]*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categorized := make(map[string][]*Command)
	for _, cmd := range r.commands {
		if !cmd.Hidden {
			category := cmd.Category
			if category == "" {
				category = "other"
			}
			categorized[category] = append(categorized[category], cmd)
		}
	}

	// Sort commands within each category
	for category := range categorized {
		sort.Slice(categorized[category], func(i, j int) bool {
			return categorized[category][i].FullName() < categorized[category][j].FullName()
		})
	}

	return categorized
}

// CategoryOrder returns categories in display order
func (r *Registry) CategoryOrder() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	type categoryWithOrder struct {
		name  string
		order int
	}

	categoriesWithOrder := make([]categoryWithOrder, 0, len(r.metadata.Categories))
	for category := range r.metadata.Categories {
		order := 99 // default order
		if info, exists := r.metadata.CategoryInfo[category]; exists {
			order = info.Order
		}
		categoriesWithOrder = append(categoriesWithOrder, categoryWithOrder{name: category, order: order})
	}

	// Sort by order, then by name
	sort.Slice(categoriesWithOrder, func(i, j int) bool {
		if categoriesWithOrder[i].order != categoriesWithOrder[j].order {
			return categoriesWithOrder[i].order < categoriesWithOrder[j].order
		}
		return categoriesWithOrder[i].name < categoriesWithOrder[j].name
	})

	categories := make([]string, len(categoriesWithOrder))
	for i, cat := range categoriesWithOrder {
		categories[i] = cat.name
	}

	return categories
}

// Search finds commands matching the query
func (r *Registry) Search(query string) []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []*Command

	for _, cmd := range r.commands {
		// Check name, namespace, method, description, tags
		if strings.Contains(strings.ToLower(cmd.Name), query) ||
			strings.Contains(strings.ToLower(cmd.Namespace), query) ||
			strings.Contains(strings.ToLower(cmd.Method), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) ||
			strings.Contains(strings.ToLower(cmd.LongDescription), query) ||
			r.searchTags(cmd.Tags, query) {
			matches = append(matches, cmd)
		}
	}

	return matches
}

// searchTags checks if any tags contain the query
func (r *Registry) searchTags(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

// Execute runs a command by name with optional arguments
func (r *Registry) Execute(name string, args ...string) error {
	cmd, exists := r.Get(name)
	if !exists {
		// Use comprehensive search for better suggestions
		suggestions := r.Search(name)
		if len(suggestions) > 0 {
			var suggestionNames []string
			for i, suggestion := range suggestions {
				if i >= 5 { // Limit to 5 suggestions
					break
				}
				suggestionNames = append(suggestionNames, suggestion.FullName())
			}
			return fmt.Errorf("%w '%s'. Did you mean: %s?",
				ErrUnknownCommand, name, strings.Join(suggestionNames, ", "))
		}
		return fmt.Errorf("%w: %s", ErrUnknownCommand, name)
	}

	// Execute dependencies first
	for _, dep := range cmd.Dependencies {
		if err := r.Execute(dep); err != nil {
			return fmt.Errorf("dependency '%s' failed: %w", dep, err)
		}
	}

	// Execute the command
	return cmd.Execute(args...)
}

// Metadata returns registry metadata with deep-copied maps to prevent race conditions
func (r *Registry) Metadata() CommandMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Deep copy Categories map to prevent shared reference issues
	categories := make(map[string]int, len(r.metadata.Categories))
	for k, v := range r.metadata.Categories {
		categories[k] = v
	}

	// Deep copy CategoryInfo map
	categoryInfo := make(map[string]CategoryInfo, len(r.metadata.CategoryInfo))
	for k, v := range r.metadata.CategoryInfo {
		categoryInfo[k] = v
	}

	// Get namespaces without re-acquiring lock
	namespaces := r.namespacesLocked()

	return CommandMetadata{
		TotalCommands: len(r.commands),
		Namespaces:    namespaces,
		Categories:    categories,
		CategoryInfo:  categoryInfo,
		Version:       r.metadata.Version,
	}
}

// Clear removes all registered commands (useful for testing)
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands = make(map[string]*Command)
	r.aliases = make(map[string]string)
	r.metadata = CommandMetadata{
		Categories:   make(map[string]int),
		CategoryInfo: make(map[string]CategoryInfo),
	}
}

// updateMetadata updates registry metadata when a command is added
func (r *Registry) updateMetadata(cmd *Command) {
	if cmd.Category != "" {
		r.metadata.Categories[cmd.Category]++
	}

	// Initialize category info if not present
	if r.metadata.CategoryInfo == nil {
		r.metadata.CategoryInfo = make(map[string]CategoryInfo)
	}

	// Update category info with standard categories
	if _, exists := r.metadata.CategoryInfo[cmd.Category]; !exists && cmd.Category != "" {
		r.metadata.CategoryInfo[cmd.Category] = r.getStandardCategoryInfo(cmd.Category)
	}
}

// getStandardCategoryInfo returns standard category information
func (r *Registry) getStandardCategoryInfo(category string) CategoryInfo {
	standardCategories := map[string]CategoryInfo{
		"core":     {Name: "Essential Operations", Icon: "🎯", Order: 1},
		"build":    {Name: "Build & Compilation", Icon: "🔨", Order: 2},
		"test":     {Name: "Testing & Quality", Icon: "🧪", Order: 3},
		"quality":  {Name: "Code Quality & Linting", Icon: "✨", Order: 4},
		"deps":     {Name: "Dependency Management", Icon: "📦", Order: 5},
		"tools":    {Name: "Development Tools", Icon: "🔧", Order: 6},
		"modules":  {Name: "Go Module Operations", Icon: "📋", Order: 7},
		"docs":     {Name: "Documentation", Icon: "📚", Order: 8},
		"git":      {Name: "Git Operations", Icon: "🔀", Order: 9},
		"version":  {Name: "Version Management", Icon: "🏷️", Order: 10},
		"metrics":  {Name: "Code Analysis & Metrics", Icon: "📊", Order: 11},
		"config":   {Name: "Configuration Management", Icon: "⚙️", Order: 13},
		"generate": {Name: "Code Generation", Icon: "🏗️", Order: 14},
		"init":     {Name: "Project Initialization", Icon: "🚀", Order: 15},
		"update":   {Name: "Update Management", Icon: "🔄", Order: 17},
		"help":     {Name: "Help System", Icon: "📖", Order: 18},
	}

	if info, exists := standardCategories[category]; exists {
		return info
	}

	// Default category info
	return CategoryInfo{
		Name:  cases.Title(language.English).String(category),
		Icon:  "📋",
		Order: 99,
	}
}

// Global registration functions for convenience

// Register adds a command to the global registry
func Register(cmd *Command) error {
	return Global().Register(cmd)
}

// MustRegister adds a command to the global registry, panicking on error
func MustRegister(cmd *Command) {
	Global().MustRegister(cmd)
}

// Get retrieves a command from the global registry
func Get(name string) (*Command, bool) {
	return Global().Get(name)
}

// List returns all commands from the global registry
func List() []*Command {
	return Global().List()
}

// Execute runs a command from the global registry
func Execute(name string, args ...string) error {
	return Global().Execute(name, args...)
}
