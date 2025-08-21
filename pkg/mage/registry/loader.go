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

// DiscoverUserCommands discovers commands in magefile.go without loading them
// Returns the list of commands that would be available for delegation
func (l *Loader) DiscoverUserCommands(dir string) ([]CommandInfo, error) {
	magefilePath := filepath.Join(dir, "magefile.go")

	// Check if magefile exists
	if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
		if l.verbose {
			utils.Info("No magefile.go found, using built-in commands only")
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

// parseMagefile parses a magefile to discover exported functions
func (l *Loader) parseMagefile(path string) ([]CommandInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var commands []CommandInfo

	// Look for exported functions and types
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Check if it's an exported function
			if d.Name.IsExported() {
				cmd := CommandInfo{
					Name:        d.Name.Name,
					Description: extractDescription(d.Doc),
				}

				// Check if it's a method (namespace function)
				if d.Recv != nil && len(d.Recv.List) > 0 {
					cmd.IsNamespace = true
					cmd.Namespace = getReceiverType(d.Recv)
					cmd.Method = d.Name.Name
				}

				commands = append(commands, cmd)
			}

		case *ast.GenDecl:
			// Look for type aliases that might be namespace re-exports
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
						// Skip if this is just a mg.Namespace alias (these are framework types, not custom commands)
						if isNamespaceAlias(ts) {
							continue
						}
						// This might be a real namespace type
						commands = append(commands, CommandInfo{
							Name:        ts.Name.Name,
							IsNamespace: true,
							Namespace:   ts.Name.Name,
							Description: extractDescription(d.Doc),
						})
					}
				}
			}
		}
	}

	return commands, nil
}

// Helper functions

// isNamespaceAlias checks if a type spec is just a mg.Namespace alias
func isNamespaceAlias(ts *ast.TypeSpec) bool {
	if ts.Type == nil {
		return false
	}

	// Check if the type is a selector expression like mg.Namespace
	if sel, ok := ts.Type.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "mg" && sel.Sel.Name == "Namespace"
		}
	}

	return false
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
