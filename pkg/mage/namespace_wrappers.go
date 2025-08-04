// Package mage provides wrapper implementations for namespace interfaces
package mage

// buildNamespaceWrapper wraps the Build namespace to implement BuildNamespace interface
type buildNamespaceWrapper struct {
	Build
}

// Ensure buildNamespaceWrapper implements BuildNamespace
var _ BuildNamespace = (*buildNamespaceWrapper)(nil)

// testNamespaceWrapper wraps the Test namespace to implement TestNamespace interface
type testNamespaceWrapper struct {
	Test
}

// Ensure testNamespaceWrapper implements TestNamespace
var _ TestNamespace = (*testNamespaceWrapper)(nil)

// lintNamespaceWrapper wraps the Lint namespace to implement LintNamespace interface
type lintNamespaceWrapper struct {
	Lint
}

// Ensure lintNamespaceWrapper implements LintNamespace
var _ LintNamespace = (*lintNamespaceWrapper)(nil)

// formatNamespaceWrapper wraps the Format namespace to implement FormatNamespace interface
type formatNamespaceWrapper struct {
	Format
}

// Ensure formatNamespaceWrapper implements FormatNamespace
var _ FormatNamespace = (*formatNamespaceWrapper)(nil)

// depsNamespaceWrapper wraps the Deps namespace to implement DepsNamespace interface
type depsNamespaceWrapper struct {
	Deps
}

// Ensure depsNamespaceWrapper implements DepsNamespace
var _ DepsNamespace = (*depsNamespaceWrapper)(nil)

// gitNamespaceWrapper wraps the Git namespace to implement GitNamespace interface
type gitNamespaceWrapper struct {
	Git
}

// Ensure gitNamespaceWrapper implements GitNamespace
var _ GitNamespace = (*gitNamespaceWrapper)(nil)

// releaseNamespaceWrapper wraps the Release namespace to implement ReleaseNamespace interface
type releaseNamespaceWrapper struct {
	Release
}

// Ensure releaseNamespaceWrapper implements ReleaseNamespace
var _ ReleaseNamespace = (*releaseNamespaceWrapper)(nil)

// docsNamespaceWrapper wraps the Docs namespace to implement DocsNamespace interface
type docsNamespaceWrapper struct {
	Docs
}

// Ensure docsNamespaceWrapper implements DocsNamespace
var _ DocsNamespace = (*docsNamespaceWrapper)(nil)

// toolsNamespaceWrapper wraps the Tools namespace to implement ToolsNamespace interface
type toolsNamespaceWrapper struct {
	Tools
}

// Ensure toolsNamespaceWrapper implements ToolsNamespace
var _ ToolsNamespace = (*toolsNamespaceWrapper)(nil)

// generateNamespaceWrapper wraps the Generate namespace to implement GenerateNamespace interface
type generateNamespaceWrapper struct {
	Generate
}

// Ensure generateNamespaceWrapper implements GenerateNamespace
var _ GenerateNamespace = (*generateNamespaceWrapper)(nil)

// cliNamespaceWrapper wraps the CLI namespace to implement CLINamespace interface
type cliNamespaceWrapper struct {
	CLI
}

// Ensure cliNamespaceWrapper implements CLINamespace
var _ CLINamespace = (*cliNamespaceWrapper)(nil)

// updateNamespaceWrapper wraps the Update namespace to implement UpdateNamespace interface
type updateNamespaceWrapper struct {
	Update
}

// Ensure updateNamespaceWrapper implements UpdateNamespace
var _ UpdateNamespace = (*updateNamespaceWrapper)(nil)

// modNamespaceWrapper wraps the Mod namespace to implement ModNamespace interface
type modNamespaceWrapper struct {
	Mod
}

// Ensure modNamespaceWrapper implements ModNamespace
var _ ModNamespace = (*modNamespaceWrapper)(nil)

// recipesNamespaceWrapper wraps the Recipes namespace to implement RecipesNamespace interface
type recipesNamespaceWrapper struct {
	Recipes
}

// Ensure recipesNamespaceWrapper implements RecipesNamespace
var _ RecipesNamespace = (*recipesNamespaceWrapper)(nil)

// metricsNamespaceWrapper wraps the Metrics to implement MetricsNamespace interface
type metricsNamespaceWrapper struct {
	Metrics
}

// Ensure metricsNamespaceWrapper implements MetricsNamespace
var _ MetricsNamespace = (*metricsNamespaceWrapper)(nil)

// workflowNamespaceWrapper wraps the Workflow namespace to implement WorkflowNamespace interface
type workflowNamespaceWrapper struct {
	Workflow
}

// Ensure workflowNamespaceWrapper implements WorkflowNamespace
var _ WorkflowNamespace = (*workflowNamespaceWrapper)(nil)

// Factory functions for creating namespace implementations
// These maintain the exact same API but use consolidated implementation

// NewBuildNamespace creates a new BuildNamespace implementation
func NewBuildNamespace() BuildNamespace {
	return &buildNamespaceWrapper{Build{}}
}

// NewTestNamespace creates a new TestNamespace implementation
func NewTestNamespace() TestNamespace {
	return &testNamespaceWrapper{Test{}}
}

// NewLintNamespace creates a new LintNamespace implementation
func NewLintNamespace() LintNamespace {
	return &lintNamespaceWrapper{Lint{}}
}

// NewFormatNamespace creates a new FormatNamespace implementation
func NewFormatNamespace() FormatNamespace {
	return &formatNamespaceWrapper{Format{}}
}

// NewDepsNamespace creates a new DepsNamespace implementation
func NewDepsNamespace() DepsNamespace {
	return &depsNamespaceWrapper{Deps{}}
}

// NewGitNamespace creates a new GitNamespace implementation
func NewGitNamespace() GitNamespace {
	return &gitNamespaceWrapper{Git{}}
}

// NewReleaseNamespace creates a new ReleaseNamespace implementation
func NewReleaseNamespace() ReleaseNamespace {
	return &releaseNamespaceWrapper{Release{}}
}

// NewDocsNamespace creates a new DocsNamespace implementation
func NewDocsNamespace() DocsNamespace {
	return &docsNamespaceWrapper{Docs{}}
}

// NewToolsNamespace creates a new ToolsNamespace implementation
func NewToolsNamespace() ToolsNamespace {
	return &toolsNamespaceWrapper{Tools{}}
}

// NewSecurityNamespace creates a new SecurityNamespace implementation
func NewSecurityNamespace() SecurityNamespace {
	// return &securityNamespaceWrapper{Security{}}
	return nil // Temporarily disabled
}

// NewGenerateNamespace creates a new GenerateNamespace implementation
func NewGenerateNamespace() GenerateNamespace {
	return &generateNamespaceWrapper{Generate{}}
}

// NewCLINamespace creates a new CLINamespace implementation
func NewCLINamespace() CLINamespace {
	return &cliNamespaceWrapper{CLI{}}
}

// NewUpdateNamespace creates a new UpdateNamespace implementation
func NewUpdateNamespace() UpdateNamespace {
	return &updateNamespaceWrapper{Update{}}
}

// NewModNamespace creates a new ModNamespace implementation
func NewModNamespace() ModNamespace {
	return &modNamespaceWrapper{Mod{}}
}

// NewRecipesNamespace creates a new RecipesNamespace implementation
func NewRecipesNamespace() RecipesNamespace {
	return &recipesNamespaceWrapper{Recipes{}}
}

// NewMetricsNamespace creates a new MetricsNamespace implementation
func NewMetricsNamespace() MetricsNamespace {
	return &metricsNamespaceWrapper{Metrics{}}
}

// NewWorkflowNamespace creates a new WorkflowNamespace implementation
func NewWorkflowNamespace() WorkflowNamespace {
	return &workflowNamespaceWrapper{Workflow{}}
}
