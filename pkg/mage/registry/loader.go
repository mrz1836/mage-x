package registry

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Loader handles dynamic loading of magefiles and command discovery
type Loader struct {
	registry *Registry
	verbose  bool
}

// NewLoader creates a new magefile loader
func NewLoader(registry *Registry) *Loader {
	if registry == nil {
		registry = Global()
	}
	return &Loader{
		registry: registry,
		verbose:  os.Getenv("MAGE_X_VERBOSE") == "true",
	}
}

// DiscoverUserCommands discovers commands in magefiles/ directory or magefile.go without loading them
// Returns the list of commands that would be available for delegation
func (l *Loader) DiscoverUserCommands(dir string) ([]CommandInfo, error) {
	// First, check for magefiles/ directory (preferred by standard mage)
	magefilesDir := filepath.Join(dir, "magefiles")
	if info, err := os.Stat(magefilesDir); err == nil && info.IsDir() {
		if l.verbose {
			utils.Info("Found magefiles/ directory, scanning for Go files")
		}

		// Parse all Go files in the magefiles directory
		commands, err := l.parseMagefilesDir(magefilesDir)
		if err != nil {
			return nil, fmt.Errorf("failed to parse magefiles directory: %w", err)
		}

		if l.verbose {
			utils.Info("%s", fmt.Sprintf("Discovered %d custom commands in magefiles/ directory", len(commands)))
		}
		return commands, nil
	}

	// Fallback to root magefile.go
	magefilePath := filepath.Join(dir, "magefile.go")
	if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
		if l.verbose {
			utils.Info("No magefile.go or magefiles/ directory found, using built-in commands only")
		}
		return nil, nil
	}

	// Parse the magefile to discover commands
	commands, err := l.parseMagefile(magefilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse magefile: %w", err)
	}

	if l.verbose {
		utils.Info("%s", fmt.Sprintf("Discovered %d custom commands in magefile.go", len(commands)))
	}

	return commands, nil
}

// CommandInfo holds information about a discovered command
type CommandInfo struct {
	Name        string
	IsNamespace bool
	Namespace   string
	Method      string
	Description string
}

// parseMagefile parses a magefile to discover the targets mage can run
func (l *Loader) parseMagefile(path string) ([]CommandInfo, error) {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", path, err)
	}

	return mageTargets([]*ast.File{file}), nil
}

// parseMagefilesDir parses all Go files in the magefiles directory to discover
// the targets mage can run. A namespace type may be declared in a different
// file from its methods, so the files are examined together.
func (l *Loader) parseMagefilesDir(dir string) ([]CommandInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read magefiles directory: %w", err)
	}

	var files []*ast.File
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		// Only process .go files
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		// Skip test files
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		file, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, parser.ParseComments)
		if err != nil {
			if l.verbose {
				utils.Info("Warning: failed to parse %s: %v", entry.Name(), err)
			}
			continue // Skip files that can't be parsed, don't fail the entire discovery
		}
		files = append(files, file)
	}

	return mageTargets(files), nil
}

// mageTargets returns the targets mage would find in a magefile package: the
// exported functions with a target signature, and methods with one on types
// declared as mg.Namespace. Other exported functions, methods and types aren't
// targets, and mage refuses to run them.
func mageTargets(files []*ast.File) []CommandInfo {
	namespaces := make(map[string]bool)
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && isMageNamespace(ts) {
					namespaces[ts.Name.Name] = true
				}
			}
		}
	}

	var commands []CommandInfo
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !fn.Name.IsExported() || !isMageTarget(fn.Type) {
				continue
			}

			cmd := CommandInfo{
				Name:        fn.Name.Name,
				Description: extractDescription(fn.Doc),
			}
			if fn.Recv != nil {
				namespace := getReceiverType(fn.Recv)
				if !namespaces[namespace] {
					continue // mage only runs methods of mg.Namespace types
				}
				cmd.IsNamespace = true
				cmd.Namespace = namespace
				cmd.Method = fn.Name.Name
			}
			commands = append(commands, cmd)
		}
	}

	return commands
}

// isMageTarget reports whether a function signature is one mage runs as a
// target: an optional context.Context first, then only string, int, float64,
// bool or time.Duration arguments (plain or pointer), returning nothing or an
// error. This mirrors mage's own parse package.
func isMageTarget(ft *ast.FuncType) bool {
	params := ft.Params.List
	if len(params) > 0 && isSelector(params[0].Type, "context", "Context") {
		if len(params[0].Names) > 1 {
			return false // mage takes a single context
		}
		params = params[1:]
	}
	for _, param := range params {
		if !isMageArgType(param.Type) {
			return false
		}
	}

	switch ft.Results.NumFields() {
	case 0:
		return true
	case 1:
		ident, ok := ft.Results.List[0].Type.(*ast.Ident)
		return ok && ident.Name == "error"
	default:
		return false
	}
}

// isMageArgType reports whether mage accepts a target argument of this type
func isMageArgType(expr ast.Expr) bool {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X // optional argument
	}
	if ident, ok := expr.(*ast.Ident); ok {
		switch ident.Name {
		case "string", "int", "float64", "bool":
			return true
		}
	}
	return isSelector(expr, "time", "Duration")
}

// isSelector reports whether expr is the qualified identifier pkg.name
func isSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == pkg && sel.Sel.Name == name
}

// Helper functions

// isMageNamespace reports whether a type spec declares a mage namespace, which
// mage recognizes only as mg.Namespace spelled with the "mg" import name
func isMageNamespace(ts *ast.TypeSpec) bool {
	return isSelector(ts.Type, "mg", "Namespace")
}

// extractDescription extracts the description from doc comments
func extractDescription(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var lines []string
	for _, comment := range doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			lines = append(lines, text)
		}
	}

	return strings.Join(lines, " ")
}

// getReceiverType extracts the receiver type name from a method
func getReceiverType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}

	field := recv.List[0]
	if field.Type == nil {
		return ""
	}

	switch t := field.Type.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	}

	return ""
}
