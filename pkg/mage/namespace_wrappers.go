// Package mage provides wrapper implementations for namespace interfaces
package mage

import "sync/atomic"

//nolint:gochecknoglobals // Required for unique instance IDs across all wrappers
var instanceCounter int64

// buildNamespaceWrapper wraps the Build namespace to implement BuildNamespace interface
type buildNamespaceWrapper struct {
	Build

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure buildNamespaceWrapper implements BuildNamespace
var _ BuildNamespace = (*buildNamespaceWrapper)(nil)

// testNamespaceWrapper wraps the Test namespace to implement TestNamespace interface
type testNamespaceWrapper struct {
	Test

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure testNamespaceWrapper implements TestNamespace
var _ TestNamespace = (*testNamespaceWrapper)(nil)

// lintNamespaceWrapper wraps the Lint namespace to implement LintNamespace interface
type lintNamespaceWrapper struct {
	Lint

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure lintNamespaceWrapper implements LintNamespace
var _ LintNamespace = (*lintNamespaceWrapper)(nil)

// formatNamespaceWrapper wraps the Format namespace to implement FormatNamespace interface
type formatNamespaceWrapper struct {
	Format

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure formatNamespaceWrapper implements FormatNamespace
var _ FormatNamespace = (*formatNamespaceWrapper)(nil)

// depsNamespaceWrapper wraps the Deps namespace to implement DepsNamespace interface
type depsNamespaceWrapper struct {
	Deps

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure depsNamespaceWrapper implements DepsNamespace
var _ DepsNamespace = (*depsNamespaceWrapper)(nil)

// gitNamespaceWrapper wraps the Git namespace to implement GitNamespace interface
type gitNamespaceWrapper struct {
	Git

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure gitNamespaceWrapper implements GitNamespace
var _ GitNamespace = (*gitNamespaceWrapper)(nil)

// releaseNamespaceWrapper wraps the Release namespace to implement ReleaseNamespace interface
type releaseNamespaceWrapper struct {
	Release

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure releaseNamespaceWrapper implements ReleaseNamespace
var _ ReleaseNamespace = (*releaseNamespaceWrapper)(nil)

// docsNamespaceWrapper wraps the Docs namespace to implement DocsNamespace interface
type docsNamespaceWrapper struct {
	Docs

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure docsNamespaceWrapper implements DocsNamespace
var _ DocsNamespace = (*docsNamespaceWrapper)(nil)

// toolsNamespaceWrapper wraps the Tools namespace to implement ToolsNamespace interface
type toolsNamespaceWrapper struct {
	Tools

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure toolsNamespaceWrapper implements ToolsNamespace
var _ ToolsNamespace = (*toolsNamespaceWrapper)(nil)

// generateNamespaceWrapper wraps the Generate namespace to implement GenerateNamespace interface
type generateNamespaceWrapper struct {
	Generate

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure generateNamespaceWrapper implements GenerateNamespace
var _ GenerateNamespace = (*generateNamespaceWrapper)(nil)

// updateNamespaceWrapper wraps the Update namespace to implement UpdateNamespace interface
type updateNamespaceWrapper struct {
	Update

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure updateNamespaceWrapper implements UpdateNamespace
var _ UpdateNamespace = (*updateNamespaceWrapper)(nil)

// modNamespaceWrapper wraps the Mod namespace to implement ModNamespace interface
type modNamespaceWrapper struct {
	Mod

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure modNamespaceWrapper implements ModNamespace
var _ ModNamespace = (*modNamespaceWrapper)(nil)

// metricsNamespaceWrapper wraps the Metrics to implement MetricsNamespace interface
type metricsNamespaceWrapper struct {
	Metrics

	instanceID int64 // Unique identifier to ensure different instances
}

// Ensure metricsNamespaceWrapper implements MetricsNamespace
var _ MetricsNamespace = (*metricsNamespaceWrapper)(nil)

// Factory functions for creating namespace implementations
// These maintain the exact same API but use consolidated implementation

// NewBuildNamespace creates a new BuildNamespace implementation
func NewBuildNamespace() BuildNamespace {
	return &buildNamespaceWrapper{
		Build:      Build{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewTestNamespace creates a new TestNamespace implementation
func NewTestNamespace() TestNamespace {
	return &testNamespaceWrapper{
		Test:       Test{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewLintNamespace creates a new LintNamespace implementation
func NewLintNamespace() LintNamespace {
	return &lintNamespaceWrapper{
		Lint:       Lint{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewFormatNamespace creates a new FormatNamespace implementation
func NewFormatNamespace() FormatNamespace {
	return &formatNamespaceWrapper{
		Format:     Format{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewDepsNamespace creates a new DepsNamespace implementation
func NewDepsNamespace() DepsNamespace {
	return &depsNamespaceWrapper{
		Deps:       Deps{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewGitNamespace creates a new GitNamespace implementation
func NewGitNamespace() GitNamespace {
	return &gitNamespaceWrapper{
		Git:        Git{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewReleaseNamespace creates a new ReleaseNamespace implementation
func NewReleaseNamespace() ReleaseNamespace {
	return &releaseNamespaceWrapper{
		Release:    Release{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewDocsNamespace creates a new DocsNamespace implementation
func NewDocsNamespace() DocsNamespace {
	return &docsNamespaceWrapper{
		Docs:       Docs{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewToolsNamespace creates a new ToolsNamespace implementation
func NewToolsNamespace() ToolsNamespace {
	return &toolsNamespaceWrapper{
		Tools:      Tools{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewSecurityNamespace creates a new SecurityNamespace implementation
func NewSecurityNamespace() SecurityNamespace {
	// return &securityNamespaceWrapper{Security{}}
	return nil // Temporarily disabled
}

// NewGenerateNamespace creates a new GenerateNamespace implementation
func NewGenerateNamespace() GenerateNamespace {
	return &generateNamespaceWrapper{
		Generate:   Generate{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewUpdateNamespace creates a new UpdateNamespace implementation
func NewUpdateNamespace() UpdateNamespace {
	return &updateNamespaceWrapper{
		Update:     Update{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewModNamespace creates a new ModNamespace implementation
func NewModNamespace() ModNamespace {
	return &modNamespaceWrapper{
		Mod:        Mod{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}

// NewMetricsNamespace creates a new MetricsNamespace implementation
func NewMetricsNamespace() MetricsNamespace {
	return &metricsNamespaceWrapper{
		Metrics:    Metrics{},
		instanceID: atomic.AddInt64(&instanceCounter, 1),
	}
}
