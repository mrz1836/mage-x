// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errOperationCanceled = errors.New("operation canceled")
	errGoModExists       = errors.New("go.mod already exists")
	errModuleRequired    = errors.New("module name is required")
	errUnsupportedFormat = errors.New("unsupported format")
)

// Mod namespace for Go module management tasks
type Mod mg.Namespace

// Download downloads all module dependencies
func (Mod) Download() error {
	utils.Header("Downloading Module Dependencies")

	// Show current module
	module, err := utils.GetModuleName()
	if err == nil {
		utils.Info("Module: %s", module)
	}

	// Download dependencies
	utils.Info("Downloading dependencies...")
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	// Verify downloads
	utils.Info("Verifying downloads...")
	if err := GetRunner().RunCmd("go", "mod", "verify"); err != nil {
		utils.Warn("Verification failed: %v", err)
	}

	utils.Success("Dependencies downloaded successfully")
	return nil
}

// Tidy cleans up go.mod and go.sum
func (Mod) Tidy() error {
	utils.Header("Tidying Module Dependencies")

	// Run go mod tidy
	utils.Info("Running go mod tidy...")
	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Check if anything changed
	output, err := GetRunner().RunCmdOutput("git", "status", "--porcelain", "go.mod", "go.sum")
	if err == nil && strings.TrimSpace(output) != "" {
		utils.Info("Module files were updated:")
		utils.Info("%s", output)
	} else {
		utils.Success("Module files are already tidy")
	}

	return nil
}

// Update updates all dependencies to latest versions
func (Mod) Update() error {
	utils.Header("Updating Dependencies")

	// List outdated dependencies first
	utils.Info("Checking for updates...")
	output, err := GetRunner().RunCmdOutput("go", "list", "-u", "-m", "all")
	if err != nil {
		return fmt.Errorf("failed to list dependencies: %w", err)
	}

	// Count updates available
	updates := 0
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			updates++
			utils.Info("  %s", line)
		}
	}

	if updates == 0 {
		utils.Success("All dependencies are up to date")
		return nil
	}

	utils.Info("Found %d updates available", updates)

	// Update dependencies
	utils.Info("Updating dependencies...")
	if err := GetRunner().RunCmd("go", "get", "-u", "./..."); err != nil {
		return fmt.Errorf("failed to update dependencies: %w", err)
	}

	// Run tidy after update
	utils.Info("Running go mod tidy...")
	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	utils.Success("Dependencies updated successfully")
	return nil
}

// Clean removes the module cache
func (Mod) Clean() error {
	utils.Header("Cleaning Module Cache")

	utils.Warn("This will remove all cached modules!")
	utils.Info("Module cache location: %s", getModCache())

	// Check if FORCE is set
	if GetMageXEnv("FORCE") != trueValue {
		utils.Error("Set MAGE_X_FORCE=true to confirm module cache deletion")
		return errOperationCanceled
	}

	// Clean module cache
	utils.Info("Cleaning module cache...")
	if err := GetRunner().RunCmd("go", "clean", "-modcache"); err != nil {
		return fmt.Errorf("failed to clean module cache: %w", err)
	}

	utils.Success("Module cache cleaned")
	return nil
}

// Graph generates a dependency graph with tree visualization (use depth=3 format=json filter=pattern show_versions=false)
func (Mod) Graph(args ...string) error {
	utils.Header("Generating Dependency Graph")

	// Parse command-line parameters
	params := utils.ParseParams(args)

	depth := 0 // 0 = unlimited
	if depthStr := utils.GetParam(params, "depth", "0"); depthStr != "0" {
		if d, err := strconv.Atoi(depthStr); err == nil {
			depth = d
		}
	}
	showVersions := utils.GetParam(params, "show_versions", trueValue) == trueValue
	filter := utils.GetParam(params, "filter", "")
	format := strings.ToLower(utils.GetParam(params, "format", "tree"))

	// Get module graph
	utils.Info("Analyzing dependencies...")
	output, err := GetRunner().RunCmdOutput("go", "mod", "graph")
	if err != nil {
		return fmt.Errorf("failed to generate dependency graph: %w", err)
	}

	// Parse the graph into a tree structure
	graph := parseModGraph(output)

	// Get root module name
	rootModule, err := utils.GetModuleName()
	if err != nil {
		utils.Debug("Failed to get module name: %v", err)
		rootModule = statusUnknown
	}

	// Apply filter if specified
	if filter != "" {
		graph = filterGraph(graph, filter)
	}

	// Display the graph based on format
	switch format {
	case "tree":
		displayTreeGraph(graph, rootModule, showVersions, depth)
	case "json":
		displayJSONGraph(graph, rootModule)
	case "dot":
		displayDotGraph(graph)
	case "mermaid":
		displayMermaidGraph(graph)
	default:
		return fmt.Errorf("%w: %s (supported: tree, json, dot, mermaid)", errUnsupportedFormat, format)
	}

	// Display summary statistics
	displayGraphStats(graph, rootModule)

	// Save full graph if requested (standard env var)
	if graphFile := GetMageXEnv("GRAPH_FILE"); graphFile != "" {
		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(graphFile, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write graph file: %w", err)
		}
		utils.Success("Full dependency graph saved to: %s", graphFile)
	}

	return nil
}

// DependencyNode represents a node in the dependency tree
type DependencyNode struct {
	Name         string
	Version      string
	Dependencies []*DependencyNode
	Visited      bool // for cycle detection
}

// DependencyGraph represents the full dependency graph
type DependencyGraph struct {
	Nodes map[string]*DependencyNode
	Edges map[string][]string
}

// parseModGraph parses go mod graph output into a structured graph
func parseModGraph(output string) *DependencyGraph {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: make(map[string][]string),
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		parent := parts[0]
		dep := parts[1]

		// Create parent node if it doesn't exist
		if _, exists := graph.Nodes[parent]; !exists {
			name, ver := parseModuleNameVersion(parent)
			graph.Nodes[parent] = &DependencyNode{
				Name:         name,
				Version:      ver,
				Dependencies: []*DependencyNode{},
			}
		}

		// Create dependency node if it doesn't exist
		if _, exists := graph.Nodes[dep]; !exists {
			name, ver := parseModuleNameVersion(dep)
			graph.Nodes[dep] = &DependencyNode{
				Name:         name,
				Version:      ver,
				Dependencies: []*DependencyNode{},
			}
		}

		// Add edge
		graph.Edges[parent] = append(graph.Edges[parent], dep)
		graph.Nodes[parent].Dependencies = append(graph.Nodes[parent].Dependencies, graph.Nodes[dep])
	}

	return graph
}

// parseModuleNameVersion splits module@version into name and version
func parseModuleNameVersion(module string) (string, string) {
	if parts := strings.Split(module, "@"); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return module, ""
}

// filterGraph filters the graph to only include modules matching the filter
func filterGraph(graph *DependencyGraph, filter string) *DependencyGraph {
	filteredGraph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
		Edges: make(map[string][]string),
	}

	// Find matching nodes
	for key, node := range graph.Nodes {
		if strings.Contains(strings.ToLower(node.Name), strings.ToLower(filter)) {
			filteredGraph.Nodes[key] = node
			filteredGraph.Edges[key] = graph.Edges[key]
		}
	}

	return filteredGraph
}

// displayTreeGraph displays the dependency graph as a tree
func displayTreeGraph(graph *DependencyGraph, rootModule string, showVersions bool, maxDepth int) {
	if rootNode, exists := graph.Nodes[rootModule]; exists {
		utils.Info("Dependency Tree:")
		fmt.Print("\n")
		displayNode(rootNode, "", true, showVersions, 0, maxDepth, make(map[string]bool))
	} else {
		utils.Warn("Root module not found in graph: %s", rootModule)
	}
}

// displayNode recursively displays a dependency node with tree formatting
func displayNode(node *DependencyNode, prefix string, isLast, showVersions bool, depth, maxDepth int, visited map[string]bool) {
	// Check depth limit - depth > maxDepth ensures we show nodes AT maxDepth level
	// For example, depth=1 shows root (0) and direct children (1)
	if maxDepth > 0 && depth > maxDepth {
		return
	}

	// Format node name
	nodeName := node.Name
	if showVersions && node.Version != "" {
		nodeName += "@" + node.Version
	}

	// Check for cycles
	nodeKey := nodeName
	if visited[nodeKey] {
		fmt.Printf("%s%s %s (cycle detected)\n", prefix, getTreeSymbol(isLast), nodeName)
		return
	}
	visited[nodeKey] = true

	// Print current node
	fmt.Printf("%s%s %s\n", prefix, getTreeSymbol(isLast), nodeName)

	// Print dependencies
	for i, dep := range node.Dependencies {
		isLastDep := i == len(node.Dependencies)-1
		newPrefix := prefix + getTreePrefix(isLast)
		displayNode(dep, newPrefix, isLastDep, showVersions, depth+1, maxDepth, visited)
	}

	// Remove from visited to allow showing the same module in different subtrees
	delete(visited, nodeKey)
}

// getTreeSymbol returns the appropriate tree drawing symbol
func getTreeSymbol(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}

// getTreePrefix returns the prefix for child nodes
func getTreePrefix(parentIsLast bool) string {
	if parentIsLast {
		return "    "
	}
	return "│   "
}

// displayJSONGraph displays the dependency graph as JSON
func displayJSONGraph(graph *DependencyGraph, rootModule string) {
	utils.Info("Dependency Graph (JSON format):")
	fmt.Print("\n")

	// Convert graph to a simple structure for JSON serialization
	type jsonNode struct {
		Name         string     `json:"name"`
		Version      string     `json:"version,omitempty"`
		Dependencies []jsonNode `json:"dependencies,omitempty"`
	}

	var convertNode func(*DependencyNode, map[string]bool) jsonNode
	convertNode = func(node *DependencyNode, visited map[string]bool) jsonNode {
		nodeKey := node.Name + "@" + node.Version
		result := jsonNode{
			Name:    node.Name,
			Version: node.Version,
		}

		if !visited[nodeKey] {
			visited[nodeKey] = true
			for _, dep := range node.Dependencies {
				result.Dependencies = append(result.Dependencies, convertNode(dep, visited))
			}
		}

		return result
	}

	if rootNode, exists := graph.Nodes[rootModule]; exists {
		jsonData := convertNode(rootNode, make(map[string]bool))
		if jsonBytes, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			fmt.Print(string(jsonBytes) + "\n")
		} else {
			utils.Error("Failed to marshal JSON: %v", err)
		}
	} else {
		utils.Warn("Root module not found in graph: %s", rootModule)
	}
}

// displayDotGraph displays the dependency graph in DOT format for graphviz
func displayDotGraph(graph *DependencyGraph) {
	utils.Info("Dependency Graph (DOT format):")
	fmt.Print("\n")
	fmt.Print("digraph dependencies {\n")
	fmt.Print("  rankdir=TB;\n")
	fmt.Print("  node [shape=box, style=rounded];\n")
	fmt.Print("\n")

	// Add nodes
	for _, node := range graph.Nodes {
		nodeName := node.Name
		if node.Version != "" {
			nodeName += "\\n" + node.Version
		}
		fmt.Printf("  \"%s\" [label=\"%s\"];\n", node.Name, nodeName)
	}

	fmt.Print("\n")

	// Add edges
	for parent, deps := range graph.Edges {
		for _, dep := range deps {
			depName, _ := parseModuleNameVersion(dep)
			parentName, _ := parseModuleNameVersion(parent)
			fmt.Printf("  \"%s\" -> \"%s\";\n", parentName, depName)
		}
	}

	fmt.Print("}\n")
}

// displayMermaidGraph displays the dependency graph in Mermaid format
func displayMermaidGraph(graph *DependencyGraph) {
	utils.Info("Dependency Graph (Mermaid format):")
	fmt.Print("\n")
	fmt.Print("graph TD;\n")

	// Create node IDs
	nodeIDs := make(map[string]string)
	idCounter := 1
	for key, node := range graph.Nodes {
		nodeIDs[key] = fmt.Sprintf("N%d", idCounter)
		idCounter++
		nodeName := node.Name
		if node.Version != "" {
			nodeName += "<br/>" + node.Version
		}
		fmt.Printf("  %s[\"%s\"];\n", nodeIDs[key], nodeName)
	}

	fmt.Print("\n")

	// Add edges
	for parent, deps := range graph.Edges {
		for _, dep := range deps {
			fmt.Printf("  %s --> %s;\n", nodeIDs[parent], nodeIDs[dep])
		}
	}
}

// displayGraphStats displays summary statistics about the dependency graph
func displayGraphStats(graph *DependencyGraph, rootModule string) {
	fmt.Print("\n")
	utils.Header("Dependency Statistics")

	// Count direct dependencies
	directDeps := 0
	if deps, exists := graph.Edges[rootModule]; exists {
		directDeps = len(deps)
	}

	// Count total unique modules
	totalModules := len(graph.Nodes)

	// Calculate maximum depth
	maxDepth := calculateMaxDepth(graph, rootModule)

	// Find duplicate dependencies (same module, different versions)
	duplicates := findDuplicateDependencies(graph)

	utils.Info("Direct dependencies: %d", directDeps)
	utils.Info("Total unique modules: %d", totalModules)
	utils.Info("Maximum dependency depth: %d", maxDepth)

	if len(duplicates) > 0 {
		utils.Warn("Modules with multiple versions:")
		for module, versions := range duplicates {
			utils.Warn("  %s: %s", module, strings.Join(versions, ", "))
		}
	}
}

// calculateMaxDepth calculates the maximum dependency depth from the root module
func calculateMaxDepth(graph *DependencyGraph, rootModule string) int {
	var calculateDepth func(string, map[string]bool) int
	calculateDepth = func(module string, visited map[string]bool) int {
		if visited[module] {
			return 0 // Avoid infinite recursion on cycles
		}
		visited[module] = true
		defer func() { delete(visited, module) }()

		maxChildDepth := 0
		if deps, exists := graph.Edges[module]; exists {
			for _, dep := range deps {
				childDepth := calculateDepth(dep, visited)
				if childDepth > maxChildDepth {
					maxChildDepth = childDepth
				}
			}
		}
		return maxChildDepth + 1
	}

	return calculateDepth(rootModule, make(map[string]bool)) - 1 // Subtract 1 to not count the root
}

// findDuplicateDependencies finds modules that appear with different versions
func findDuplicateDependencies(graph *DependencyGraph) map[string][]string {
	moduleVersions := make(map[string]map[string]bool)

	// Collect all versions for each module
	for _, node := range graph.Nodes {
		if node.Version != "" {
			if moduleVersions[node.Name] == nil {
				moduleVersions[node.Name] = make(map[string]bool)
			}
			moduleVersions[node.Name][node.Version] = true
		}
	}

	// Find modules with multiple versions
	duplicates := make(map[string][]string)
	for module, versions := range moduleVersions {
		if len(versions) > 1 {
			versionList := make([]string, 0, len(versions))
			for version := range versions {
				versionList = append(versionList, version)
			}
			duplicates[module] = versionList
		}
	}

	return duplicates
}

// Why shows why a module is needed
func (Mod) Why(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: usage: mage mod:why <module1> [module2]", errModuleRequired)
	}

	utils.Header("Module Dependency Analysis")

	// Analyze each provided module
	for i, module := range args {
		if i > 0 {
			fmt.Print("\n")
		}

		utils.Info("Analyzing why %s is needed...", module)

		// Run go mod why
		output, err := GetRunner().RunCmdOutput("go", "mod", "why", module)
		if err != nil {
			return fmt.Errorf("failed to analyze module %s: %w", module, err)
		}

		utils.Info("Dependency path:")
		utils.Info("%s", output)

		// Also check if it's a direct dependency
		directDeps, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{.Require}}", "all")
		if err == nil && strings.Contains(directDeps, module) {
			utils.Info("This is a DIRECT dependency")
		} else {
			utils.Info("This is an INDIRECT dependency")
		}
	}

	return nil
}

// Vendor vendors all dependencies
func (Mod) Vendor() error {
	utils.Header("Vendoring Dependencies")

	// Check if vendor directory exists
	vendorExists := utils.DirExists("vendor")

	// Run go mod vendor
	utils.Info("Vendoring dependencies...")
	if err := GetRunner().RunCmd("go", "mod", "vendor"); err != nil {
		return fmt.Errorf("vendoring failed: %w", err)
	}

	// Show vendor directory size
	if size, err := getDirSize("vendor"); err == nil {
		utils.Info("Vendor directory size: %s", formatBytes(size))
	}

	if !vendorExists {
		utils.Success("Dependencies vendored successfully")
		utils.Info("Remember to add vendor/ to your .gitignore")
	} else {
		utils.Success("Vendor directory updated")
	}

	return nil
}

// Init initializes a new Go module
func (Mod) Init() error {
	utils.Header("Initializing Go Module")

	// Check if go.mod already exists
	if utils.FileExists("go.mod") {
		return errGoModExists
	}

	// Get module name
	moduleName := GetMageXEnv("MODULE")
	if moduleName == "" {
		// Try to infer from git remote
		if remote, err := GetRunner().RunCmdOutput("git", "remote", "get-url", "origin"); err == nil {
			remote = strings.TrimSpace(remote)
			// Convert git URL to module path
			moduleName = gitURLToModulePath(remote)
		}
	}

	if moduleName == "" {
		return fmt.Errorf("%w. Usage: MAGE_X_MODULE=github.com/user/repo magex mod:init", errModuleRequired)
	}

	// Initialize module
	utils.Info("Initializing module: %s", moduleName)
	if err := GetRunner().RunCmd("go", "mod", "init", moduleName); err != nil {
		return fmt.Errorf("module initialization failed: %w", err)
	}

	utils.Success("Module initialized successfully")
	utils.Info("Created go.mod for %s", moduleName)

	return nil
}

// Helper functions

// getModCache returns the module cache directory
func getModCache() string {
	if cache := os.Getenv("GOMODCACHE"); cache != "" {
		return cache
	}

	// Default locations
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return filepath.Join(gopath, "pkg", "mod")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to a reasonable default
		utils.Debug("Failed to get user home directory: %v", err)
		return filepath.Join(".", "go", "pkg", "mod")
	}
	return filepath.Join(home, "go", "pkg", "mod")
}

// getDirSize calculates directory size

// gitURLToModulePath converts a git URL to a Go module path
func gitURLToModulePath(gitURL string) string {
	// Remove protocol
	gitURL = strings.TrimPrefix(gitURL, "https://")
	gitURL = strings.TrimPrefix(gitURL, "http://")
	gitURL = strings.TrimPrefix(gitURL, "git@")
	gitURL = strings.TrimPrefix(gitURL, "ssh://git@")

	// Convert git@github.com:user/repo.git to github.com/user/repo
	gitURL = strings.Replace(gitURL, ":", "/", 1)

	// Remove .git suffix
	gitURL = strings.TrimSuffix(gitURL, ".git")

	return gitURL
}

// Additional methods for Mod namespace required by tests

// Verify verifies module dependencies
func (Mod) Verify() error {
	runner := GetRunner()
	return runner.RunCmd("go", "mod", "verify")
}

// Edit edits go.mod from tools or scripts
func (Mod) Edit(args ...string) error {
	runner := GetRunner()
	cmdArgs := append([]string{"mod", "edit"}, args...)
	return runner.RunCmd("go", cmdArgs...)
}

// Get adds dependencies
func (Mod) Get(packages ...string) error {
	runner := GetRunner()
	cmdArgs := append([]string{"get"}, packages...)
	return runner.RunCmd("go", cmdArgs...)
}

// List lists modules
func (Mod) List(pattern ...string) error {
	runner := GetRunner()
	cmdArgs := []string{"list", "-m"}
	if len(pattern) > 0 {
		cmdArgs = append(cmdArgs, pattern...)
	} else {
		cmdArgs = append(cmdArgs, "all")
	}
	return runner.RunCmd("go", cmdArgs...)
}
