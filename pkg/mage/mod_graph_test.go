//go:build integration
// +build integration

package mage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors
var (
	errModGraphTestFailed = errors.New("mod graph test failed")
	errModGraphCmdFailed  = errors.New("mod graph command failed")
)

// ModGraphTestSuite defines the test suite for mod graph functions
type ModGraphTestSuite struct {
	suite.Suite
	env *testutil.TestEnvironment
	mod Mod
}

// SetupTest runs before each test
func (ts *ModGraphTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.mod = Mod{}
}

// TearDownTest runs after each test
func (ts *ModGraphTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestParseModuleNameVersion tests module name/version parsing
func (ts *ModGraphTestSuite) TestParseModuleNameVersion() {
	tests := []struct {
		name        string
		module      string
		wantName    string
		wantVersion string
	}{
		{
			name:        "module with version",
			module:      "github.com/pkg/errors@v0.9.1",
			wantName:    "github.com/pkg/errors",
			wantVersion: "v0.9.1",
		},
		{
			name:        "module without version",
			module:      "github.com/pkg/errors",
			wantName:    "github.com/pkg/errors",
			wantVersion: "",
		},
		{
			name:        "module with complex version",
			module:      "golang.org/x/text@v0.3.7-0.20210503195748-5c7c50ebbd4f",
			wantName:    "golang.org/x/text",
			wantVersion: "v0.3.7-0.20210503195748-5c7c50ebbd4f",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			name, version := parseModuleNameVersion(tt.module)
			ts.Assert().Equal(tt.wantName, name)
			ts.Assert().Equal(tt.wantVersion, version)
		})
	}
}

// TestParseModGraph tests dependency graph parsing
func (ts *ModGraphTestSuite) TestParseModGraph() {
	tests := []struct {
		name      string
		input     string
		wantNodes int
		wantEdges map[string]int
	}{
		{
			name:      "empty graph",
			input:     "",
			wantNodes: 0,
			wantEdges: map[string]int{},
		},
		{
			name: "single dependency",
			input: `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
`,
			wantNodes: 2,
			wantEdges: map[string]int{
				"github.com/test/app@v1.0.0": 1,
			},
		},
		{
			name: "multiple dependencies",
			input: `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
github.com/test/app@v1.0.0 github.com/stretchr/testify@v1.8.0
github.com/pkg/errors@v0.9.1 github.com/golang/protobuf@v1.5.2
`,
			wantNodes: 4,
			wantEdges: map[string]int{
				"github.com/test/app@v1.0.0":   2,
				"github.com/pkg/errors@v0.9.1": 1,
			},
		},
		{
			name: "with malformed lines",
			input: `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
invalid line without space
github.com/test/app@v1.0.0 github.com/stretchr/testify@v1.8.0
`,
			wantNodes: 3,
			wantEdges: map[string]int{
				"github.com/test/app@v1.0.0": 2,
			},
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			graph := parseModGraph(tt.input)

			ts.Assert().Len(graph.Nodes, tt.wantNodes)

			for parent, expectedEdges := range tt.wantEdges {
				edges, exists := graph.Edges[parent]
				ts.Assert().True(exists, "parent should have edges")
				ts.Assert().Len(edges, expectedEdges)
			}
		})
	}
}

// TestFilterGraph tests graph filtering
func (ts *ModGraphTestSuite) TestFilterGraph() {
	// Create a sample graph
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"github.com/stretchr/testify@v1.8.0": {
				Name:    "github.com/stretchr/testify",
				Version: "v1.8.0",
			},
			"github.com/pkg/errors@v0.9.1": {
				Name:    "github.com/pkg/errors",
				Version: "v0.9.1",
			},
			"golang.org/x/text@v0.3.7": {
				Name:    "golang.org/x/text",
				Version: "v0.3.7",
			},
		},
		Edges: make(map[string][]string),
	}

	tests := []struct {
		name      string
		filter    string
		wantNodes int
	}{
		{
			name:      "filter by testify",
			filter:    "testify",
			wantNodes: 1,
		},
		{
			name:      "filter by github",
			filter:    "github",
			wantNodes: 2,
		},
		{
			name:      "filter by golang",
			filter:    "golang",
			wantNodes: 1,
		},
		{
			name:      "filter no matches",
			filter:    "nonexistent",
			wantNodes: 0,
		},
		{
			name:      "case insensitive",
			filter:    "TESTIFY",
			wantNodes: 1,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			filtered := filterGraph(graph, tt.filter)
			ts.Assert().Len(filtered.Nodes, tt.wantNodes)
		})
	}
}

// TestCalculateMaxDepth tests maximum depth calculation
func (ts *ModGraphTestSuite) TestCalculateMaxDepth() {
	tests := []struct {
		name      string
		graph     *DependencyGraph
		rootKey   string
		wantDepth int
	}{
		{
			name: "no dependencies",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"root@v1.0.0": {Name: "root", Version: "v1.0.0"},
				},
				Edges: map[string][]string{},
			},
			rootKey:   "root@v1.0.0",
			wantDepth: 0,
		},
		{
			name: "single level",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"root@v1.0.0": {Name: "root", Version: "v1.0.0"},
					"dep1@v1.0.0": {Name: "dep1", Version: "v1.0.0"},
				},
				Edges: map[string][]string{
					"root@v1.0.0": {"dep1@v1.0.0"},
				},
			},
			rootKey:   "root@v1.0.0",
			wantDepth: 1,
		},
		{
			name: "multiple levels",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"root@v1.0.0": {Name: "root", Version: "v1.0.0"},
					"dep1@v1.0.0": {Name: "dep1", Version: "v1.0.0"},
					"dep2@v1.0.0": {Name: "dep2", Version: "v1.0.0"},
				},
				Edges: map[string][]string{
					"root@v1.0.0": {"dep1@v1.0.0"},
					"dep1@v1.0.0": {"dep2@v1.0.0"},
				},
			},
			rootKey:   "root@v1.0.0",
			wantDepth: 2,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			depth := calculateMaxDepth(tt.graph, tt.rootKey)
			ts.Assert().Equal(tt.wantDepth, depth)
		})
	}
}

// TestFindDuplicateDependencies tests duplicate dependency detection
func (ts *ModGraphTestSuite) TestFindDuplicateDependencies() {
	tests := []struct {
		name           string
		graph          *DependencyGraph
		wantDuplicates int
	}{
		{
			name: "no duplicates",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"github.com/pkg/errors@v0.9.1": {
						Name:    "github.com/pkg/errors",
						Version: "v0.9.1",
					},
				},
			},
			wantDuplicates: 0,
		},
		{
			name: "single duplicate",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"github.com/pkg/errors@v0.9.1": {
						Name:    "github.com/pkg/errors",
						Version: "v0.9.1",
					},
					"github.com/pkg/errors@v0.8.0": {
						Name:    "github.com/pkg/errors",
						Version: "v0.8.0",
					},
				},
			},
			wantDuplicates: 1,
		},
		{
			name: "multiple duplicates",
			graph: &DependencyGraph{
				Nodes: map[string]*DependencyNode{
					"github.com/pkg/errors@v0.9.1": {
						Name:    "github.com/pkg/errors",
						Version: "v0.9.1",
					},
					"github.com/pkg/errors@v0.8.0": {
						Name:    "github.com/pkg/errors",
						Version: "v0.8.0",
					},
					"golang.org/x/text@v0.3.7": {
						Name:    "golang.org/x/text",
						Version: "v0.3.7",
					},
					"golang.org/x/text@v0.3.6": {
						Name:    "golang.org/x/text",
						Version: "v0.3.6",
					},
				},
			},
			wantDuplicates: 2,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			duplicates := findDuplicateDependencies(tt.graph)
			ts.Assert().Len(duplicates, tt.wantDuplicates)
		})
	}
}

// TestGetTreeSymbol tests tree symbol generation
func (ts *ModGraphTestSuite) TestGetTreeSymbol() {
	tests := []struct {
		name   string
		isLast bool
		want   string
	}{
		{
			name:   "last item",
			isLast: true,
			want:   "└── ",
		},
		{
			name:   "not last item",
			isLast: false,
			want:   "├── ",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			result := getTreeSymbol(tt.isLast)
			ts.Assert().Equal(tt.want, result)
		})
	}
}

// TestGetTreePrefix tests tree prefix generation
func (ts *ModGraphTestSuite) TestGetTreePrefix() {
	tests := []struct {
		name         string
		parentIsLast bool
		wantPrefix   string
	}{
		{
			name:         "parent is last",
			parentIsLast: true,
			wantPrefix:   "    ",
		},
		{
			name:         "parent is not last",
			parentIsLast: false,
			wantPrefix:   "│   ",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			result := getTreePrefix(tt.parentIsLast)
			ts.Assert().Equal(tt.wantPrefix, result)
		})
	}
}

// TestModGraph_TreeFormat tests tree output format
func (ts *ModGraphTestSuite) TestModGraph_TreeFormat() {
	// Create mock graph output
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
github.com/test/app@v1.0.0 github.com/stretchr/testify@v1.8.0
`

	// Mock commands
	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			// Test with tree format
			return ts.mod.Graph("format=tree", "depth=2")
		},
	)

	ts.Assert().NoError(err)
}

// TestModGraph_JSONFormat tests JSON output format
func (ts *ModGraphTestSuite) TestModGraph_JSONFormat() {
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
`

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph("format=json")
		},
	)

	ts.Assert().NoError(err)
}

// TestModGraph_DOTFormat tests DOT output format
func (ts *ModGraphTestSuite) TestModGraph_DOTFormat() {
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
`

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph("format=dot")
		},
	)

	ts.Assert().NoError(err)
}

// TestModGraph_MermaidFormat tests Mermaid output format
func (ts *ModGraphTestSuite) TestModGraph_MermaidFormat() {
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
`

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph("format=mermaid")
		},
	)

	ts.Assert().NoError(err)
}

// TestModGraph_UnsupportedFormat tests unsupported format error
func (ts *ModGraphTestSuite) TestModGraph_UnsupportedFormat() {
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
`

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph("format=invalid")
		},
	)

	ts.Assert().Error(err)
	ts.Assert().Contains(err.Error(), "unsupported format")
}

// TestModGraph_WithFilter tests filtering
func (ts *ModGraphTestSuite) TestModGraph_WithFilter() {
	mockGraphOutput := `github.com/test/app@v1.0.0 github.com/pkg/errors@v0.9.1
github.com/test/app@v1.0.0 github.com/stretchr/testify@v1.8.0
`

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(mockGraphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph("filter=testify", "format=tree")
		},
	)

	ts.Assert().NoError(err)
}

// TestDisplayJSONGraph tests JSON graph display
func (ts *ModGraphTestSuite) TestDisplayJSONGraph() {
	graph := &DependencyGraph{
		Nodes: map[string]*DependencyNode{
			"root@v1.0.0": {
				Name:         "root",
				Version:      "v1.0.0",
				Dependencies: []*DependencyNode{},
			},
		},
	}

	// This should not panic
	ts.Assert().NotPanics(func() {
		displayJSONGraph(graph, "root@v1.0.0")
	})
}

// TestWriteAWSINI_RoundTrip tests that parsing and writing are inverse operations
func (ts *ModGraphTestSuite) TestINIRoundTrip() {
	originalContent := `[default]
aws_access_key_id = KEY1
aws_secret_access_key = SECRET1

[production]
region = us-west-2
mfa_serial = arn:aws:iam::123:mfa/user
`

	// Parse
	ini := parseAWSINI([]byte(originalContent))

	// Write
	output := writeAWSINI(ini)

	// Parse again
	ini2 := parseAWSINI(output)

	// Should have same number of sections
	ts.Assert().Equal(len(ini.Sections), len(ini2.Sections))

	// Check values are preserved
	for _, section1 := range ini.Sections {
		found := false
		for _, section2 := range ini2.Sections {
			if section1.Name == section2.Name {
				found = true
				for key, value1 := range section1.Values {
					value2, exists := section2.Values[key]
					ts.Assert().True(exists, "key %s should exist in second parse", key)
					ts.Assert().Equal(value1, value2, "values for key %s should match", key)
				}
				break
			}
		}
		ts.Assert().True(found, "section %s should exist in second parse", section1.Name)
	}
}

// TestModGraphTestSuite runs the test suite
func TestModGraphTestSuite(t *testing.T) {
	suite.Run(t, new(ModGraphTestSuite))
}
