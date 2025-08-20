package registry

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static loader errors
var (
	ErrInvalidStructType = errors.New("expected struct type for namespace")
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

// LoadUserMagefile attempts to load a user's magefile.go if it exists
func (l *Loader) LoadUserMagefile(dir string) error {
	magefilePath := filepath.Join(dir, "magefile.go")

	// Check if magefile exists
	if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
		if l.verbose {
			utils.Info("No magefile.go found, using built-in commands only")
		}
		return nil
	}

	// Parse the magefile to discover commands
	commands, err := l.parseMagefile(magefilePath)
	if err != nil {
		return fmt.Errorf("failed to parse magefile: %w", err)
	}

	// Compile magefile to plugin for dynamic loading
	pluginPath, err := l.compileMagefile(magefilePath)
	if err != nil {
		return fmt.Errorf("failed to compile magefile: %w", err)
	}

	// Load the plugin
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to load magefile plugin: %w", err)
	}

	// Register discovered commands
	for _, cmdInfo := range commands {
		if err := l.registerUserCommand(p, cmdInfo); err != nil {
			if l.verbose {
				fmt.Printf("Warning: failed to register command %s: %v\n", cmdInfo.Name, err)
			}
		}
	}

	return nil
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
						// This might be a namespace type
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

// compileMagefile compiles the magefile to a plugin
func (l *Loader) compileMagefile(magefilePath string) (string, error) {
	// Create temporary directory for plugin
	tmpDir, err := os.MkdirTemp("", "magex-plugin-*")
	if err != nil {
		return "", err
	}

	pluginPath := filepath.Join(tmpDir, "magefile.so")

	// Build the plugin
	ctx := context.Background()
	// #nosec G204 - magefilePath is validated and controlled by the application
	cmd := exec.CommandContext(ctx, "go", "build", "-buildmode=plugin", "-o", pluginPath, magefilePath)
	cmd.Env = append(os.Environ(), "GO111MODULE=on")

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to build magefile plugin: %w\nOutput: %s", err, output)
	}

	return pluginPath, nil
}

// registerUserCommand registers a user-defined command from the plugin
func (l *Loader) registerUserCommand(p *plugin.Plugin, info CommandInfo) error {
	// Look up the symbol
	symbol, err := p.Lookup(info.Name)
	if err != nil {
		return fmt.Errorf("failed to find symbol %s: %w", info.Name, err)
	}

	// Try to cast to a function
	switch fn := symbol.(type) {
	case func() error:
		cmd := NewCommand(info.Name).
			WithDescription(info.Description).
			WithFunc(fn).
			WithCategory("User").
			MustBuild()

		return l.registry.Register(cmd)

	case func(args ...string) error:
		cmd := NewCommand(info.Name).
			WithDescription(info.Description).
			WithArgsFunc(fn).
			WithCategory("User").
			MustBuild()

		return l.registry.Register(cmd)

	default:
		// Try reflection for namespace types
		return l.registerNamespaceType(symbol, info)
	}
}

// registerNamespaceType attempts to register methods of a namespace type
func (l *Loader) registerNamespaceType(symbol interface{}, info CommandInfo) error {
	typ := reflect.TypeOf(symbol)
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("%w %s", ErrInvalidStructType, info.Name)
	}

	// Register methods
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !method.IsExported() {
			continue
		}

		// Create command for this method
		cmdName := fmt.Sprintf("%s:%s", strings.ToLower(info.Name), strings.ToLower(method.Name))

		// Create wrapper function
		fn := func() error {
			// Create instance and call method
			instance := reflect.New(typ).Elem()
			result := method.Func.Call([]reflect.Value{instance})

			if len(result) > 0 {
				if err, ok := result[0].Interface().(error); ok {
					return err
				}
			}
			return nil
		}

		cmd := NewCommand(cmdName).
			WithDescription(fmt.Sprintf("%s %s", info.Name, method.Name)).
			WithFunc(fn).
			WithCategory("User").
			MustBuild()

		if err := l.registry.Register(cmd); err != nil {
			if l.verbose {
				fmt.Printf("Warning: failed to register %s: %v\n", cmdName, err)
			}
		}
	}

	return nil
}

// LoadFromReader loads command definitions from a reader (for embedded commands)
func (l *Loader) LoadFromReader(r io.Reader) error {
	// This would load command definitions from embedded data
	// Used for built-in MAGE-X commands
	// Implementation depends on the format we choose for embedded commands
	return nil
}

// Helper functions

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
