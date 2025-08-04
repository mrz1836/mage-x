package mage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors for err113 compliance - removed unused errors

// Test fixtures

func createTestWorkflowDefinition() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "test-workflow",
		Description: "Test workflow for unit testing",
		Version:     "1.0.0",
		Steps: []WorkflowStep{
			{
				Name:    "step1",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"hello"},
			},
			{
				Name:       "step2",
				Type:       "shell",
				Command:    "echo",
				Args:       []string{"world"},
				Parallel:   true,
				RetryCount: 2,
			},
		},
		Variables: map[string]interface{}{
			"test_var": "test_value",
		},
		Settings: WorkflowSettings{
			Timeout:          "5m",
			MaxRetries:       3,
			FailureStrategy:  "stop",
			NotificationMode: "on_failure",
			Environment:      map[string]string{"TEST": "true"},
		},
		Triggers:    []WorkflowTrigger{},
		LastUpdated: time.Now(),
	}
}

func createTestExecution() WorkflowExecution {
	workflow := createTestWorkflowDefinition()
	return WorkflowExecution{
		ID:        "test-exec-123",
		Workflow:  workflow,
		StartTime: time.Now(),
		Status:    "running",
		Results:   make([]StepResult, len(workflow.Steps)),
		Context: ExecutionContext{
			Variables:   map[string]interface{}{"test": "value"},
			Environment: map[string]string{"ENV": "test"},
			Metadata:    map[string]string{"meta": "data"},
		},
	}
}

// Test environment setup helper
func withTestEnvironment(t *testing.T, envVars map[string]string, testFunc func()) {
	// Store original environment values - include WORKFLOW key to ensure it's handled
	originalValues := make(map[string]string)
	for key := range envVars {
		originalValues[key] = os.Getenv(key)
	}
	// Always store WORKFLOW to ensure proper cleanup
	if _, exists := originalValues["WORKFLOW"]; !exists {
		originalValues["WORKFLOW"] = os.Getenv("WORKFLOW")
	}

	// Set test environment variables
	for key, value := range envVars {
		if value == "" {
			err := os.Unsetenv(key)
			require.NoError(t, err)
		} else {
			err := os.Setenv(key, value)
			require.NoError(t, err)
		}
	}

	// If testing empty environments, explicitly unset WORKFLOW
	if len(envVars) == 0 {
		err := os.Unsetenv("WORKFLOW")
		require.NoError(t, err)
	}

	// Cleanup after test
	defer func() {
		for key, originalValue := range originalValues {
			if originalValue == "" {
				err := os.Unsetenv(key)
				require.NoError(t, err)
			} else {
				err := os.Setenv(key, originalValue)
				require.NoError(t, err)
			}
		}
	}()

	testFunc()
}

// Test WorkflowExecution.Execute() with environment variables
func TestWorkflow_Execute_EnvironmentVariables(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing workflow environment variable",
			envVars: map[string]string{
				"WORKFLOW": "", // Explicitly set to empty to ensure it's unset
			},
			expectError: true,
			errorMsg:    "WORKFLOW environment variable is required",
		},
		{
			name: "workflow environment variable set",
			envVars: map[string]string{
				"WORKFLOW": "test-workflow",
			},
			expectError: true, // Will fail because workflow doesn't exist
			errorMsg:    "failed to load workflow 'test-workflow'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTestEnvironment(t, tt.envVars, func() {
				workflow := Workflow{}
				err := workflow.Execute()

				if tt.expectError {
					require.Error(t, err)
					if tt.errorMsg != "" {
						assert.Contains(t, err.Error(), tt.errorMsg)
					}
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

// Test executeWorkflowSteps function with actual workflow steps
func TestExecuteWorkflowSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		steps          []WorkflowStep
		expectError    bool
		expectedStatus string
	}{
		{
			name: "successful execution with parallel steps",
			steps: []WorkflowStep{
				{
					Name:     "step1",
					Type:     "shell",
					Command:  "echo",
					Args:     []string{"hello"},
					Parallel: false,
				},
				{
					Name:     "step2",
					Type:     "shell",
					Command:  "echo",
					Args:     []string{"world"},
					Parallel: true,
				},
			},
			expectError: false,
		},
		{
			name: "step failure without continue on error",
			steps: []WorkflowStep{
				{
					Name:            "failing-step",
					Type:            "unsupported", // This will cause failure
					Command:         "nonexistent",
					ContinueOnError: false,
				},
			},
			expectError: true,
		},
		{
			name: "step failure with continue on error",
			steps: []WorkflowStep{
				{
					Name:            "failing-step",
					Type:            "unsupported", // This will cause failure
					Command:         "nonexistent",
					ContinueOnError: true,
				},
			},
			expectError: false, // Should not fail because ContinueOnError is true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			execution := createTestExecution()
			execution.Workflow.Steps = tt.steps
			execution.Results = make([]StepResult, len(tt.steps))

			err := executeWorkflowSteps(&execution)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify that results were populated
			assert.Len(t, execution.Results, len(tt.steps))
		})
	}
}

// Test executeWorkflowStep function with different step types
func TestExecuteWorkflowStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		step           WorkflowStep
		execContext    ExecutionContext
		expectedStatus string
		expectRetries  bool
	}{
		{
			name: "successful step execution",
			step: WorkflowStep{
				Name:    "test-step",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"hello"},
			},
			execContext:    ExecutionContext{},
			expectedStatus: "completed",
			expectRetries:  false,
		},
		{
			name: "step with conditions not met",
			step: WorkflowStep{
				Name:    "conditional-step",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"hello"},
				Conditions: []StepCondition{
					{
						Type:  "variable",
						Field: "nonexistent",
						Value: "value",
					},
				},
			},
			execContext:    ExecutionContext{Variables: map[string]interface{}{}},
			expectedStatus: "skipped",
			expectRetries:  false,
		},
		{
			name: "step with retry logic",
			step: WorkflowStep{
				Name:       "retry-step",
				Type:       "unsupported", // Will fail
				Command:    "echo",
				Args:       []string{"hello"},
				RetryCount: 3,
			},
			execContext:    ExecutionContext{},
			expectedStatus: "failed",
			expectRetries:  true,
		},
		{
			name: "step with timeout setting",
			step: WorkflowStep{
				Name:    "timeout-step",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"hello"},
				Timeout: "1s",
			},
			execContext:    ExecutionContext{},
			expectedStatus: "completed", // Mock execution completes quickly
			expectRetries:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			result := executeWorkflowStep(ctx, &tt.step, tt.execContext)

			assert.Equal(t, tt.expectedStatus, result.Status)
			assert.Equal(t, tt.step.Name, result.Step.Name)
			assert.NotZero(t, result.Duration)

			if tt.expectRetries {
				assert.Positive(t, result.RetryCount)
			}
		})
	}
}

// Test step command execution types
func TestExecuteStepCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		step        WorkflowStep
		execContext ExecutionContext
		expectError bool
		errorMsg    string
	}{
		{
			name: "shell command",
			step: WorkflowStep{
				Type:    "shell",
				Command: "echo",
				Args:    []string{"hello"},
			},
			execContext: ExecutionContext{},
			expectError: false,
		},
		{
			name: "command type",
			step: WorkflowStep{
				Type:    "command",
				Command: "ls",
				Args:    []string{"-la"},
			},
			execContext: ExecutionContext{},
			expectError: false,
		},
		{
			name: "script command",
			step: WorkflowStep{
				Type:    "script",
				Command: "test.sh",
			},
			execContext: ExecutionContext{},
			expectError: false,
		},
		{
			name: "http command",
			step: WorkflowStep{
				Type:    "http",
				Command: "GET",
				Args:    []string{"https://api.example.com"},
			},
			execContext: ExecutionContext{},
			expectError: false,
		},
		{
			name: "notification command",
			step: WorkflowStep{
				Type:    "notification",
				Command: "slack",
				Args:    []string{"#general", "Test message"},
			},
			execContext: ExecutionContext{},
			expectError: false,
		},
		{
			name: "unsupported step type",
			step: WorkflowStep{
				Type:    "unsupported",
				Command: "unknown",
			},
			execContext: ExecutionContext{},
			expectError: true,
			errorMsg:    "unsupported step type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			output, err := executeStepCommand(ctx, &tt.step, tt.execContext)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, output)
			}
		})
	}
}

// Test workflow validation
func TestValidateWorkflowDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		workflow WorkflowDefinition
		expected bool
	}{
		{
			name:     "valid workflow",
			workflow: createTestWorkflowDefinition(),
			expected: true,
		},
		{
			name: "workflow without name",
			workflow: WorkflowDefinition{
				Name:        "",
				Description: "Test",
				Steps: []WorkflowStep{
					{Name: "step1", Type: "shell"},
				},
			},
			expected: false,
		},
		{
			name: "workflow without steps",
			workflow: WorkflowDefinition{
				Name:        "test",
				Description: "Test",
				Steps:       []WorkflowStep{},
			},
			expected: false,
		},
		{
			name: "workflow with invalid step",
			workflow: WorkflowDefinition{
				Name:        "test",
				Description: "Test",
				Steps: []WorkflowStep{
					{Name: "", Type: "shell"}, // Invalid: no name
				},
			},
			expected: false,
		},
		{
			name: "workflow with step without type",
			workflow: WorkflowDefinition{
				Name:        "test",
				Description: "Test",
				Steps: []WorkflowStep{
					{Name: "step1", Type: ""}, // Invalid: no type
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := validateWorkflowDefinition(&tt.workflow)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test step condition evaluation
func TestEvaluateStepConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		conditions  []StepCondition
		execContext ExecutionContext
		expected    bool
	}{
		{
			name:       "no conditions",
			conditions: []StepCondition{},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"test": "value"},
			},
			expected: true,
		},
		{
			name: "variable condition met",
			conditions: []StepCondition{
				{Type: "variable", Field: "test_var", Value: "expected"},
			},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"test_var": "expected"},
			},
			expected: true,
		},
		{
			name: "variable condition not met",
			conditions: []StepCondition{
				{Type: "variable", Field: "test_var", Value: "expected"},
			},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"test_var": "different"},
			},
			expected: false,
		},
		{
			name: "environment condition met",
			conditions: []StepCondition{
				{Type: "environment", Field: "ENV_VAR", Value: "prod"},
			},
			execContext: ExecutionContext{
				Environment: map[string]string{"ENV_VAR": "prod"},
			},
			expected: true,
		},
		{
			name: "environment condition not met",
			conditions: []StepCondition{
				{Type: "environment", Field: "ENV_VAR", Value: "prod"},
			},
			execContext: ExecutionContext{
				Environment: map[string]string{"ENV_VAR": "dev"},
			},
			expected: false,
		},
		{
			name: "multiple conditions all met",
			conditions: []StepCondition{
				{Type: "variable", Field: "var1", Value: "value1"},
				{Type: "environment", Field: "env1", Value: "envvalue1"},
			},
			execContext: ExecutionContext{
				Variables:   map[string]interface{}{"var1": "value1"},
				Environment: map[string]string{"env1": "envvalue1"},
			},
			expected: true,
		},
		{
			name: "multiple conditions one not met",
			conditions: []StepCondition{
				{Type: "variable", Field: "var1", Value: "value1"},
				{Type: "environment", Field: "env1", Value: "envvalue1"},
			},
			execContext: ExecutionContext{
				Variables:   map[string]interface{}{"var1": "different"},
				Environment: map[string]string{"env1": "envvalue1"},
			},
			expected: false,
		},
		{
			name: "unknown condition type defaults to true",
			conditions: []StepCondition{
				{Type: "unknown", Field: "field", Value: "value"},
			},
			execContext: ExecutionContext{},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := evaluateStepConditions(tt.conditions, tt.execContext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test individual condition evaluation
func TestEvaluateCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		condition   StepCondition
		execContext ExecutionContext
		expected    bool
	}{
		{
			name: "variable condition with string value",
			condition: StepCondition{
				Type:  "variable",
				Field: "test_var",
				Value: "expected",
			},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"test_var": "expected"},
			},
			expected: true,
		},
		{
			name: "variable condition with numeric value",
			condition: StepCondition{
				Type:  "variable",
				Field: "numeric_var",
				Value: "123",
			},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"numeric_var": 123},
			},
			expected: true,
		},
		{
			name: "variable condition field not found",
			condition: StepCondition{
				Type:  "variable",
				Field: "nonexistent",
				Value: "any",
			},
			execContext: ExecutionContext{
				Variables: map[string]interface{}{"other": "value"},
			},
			expected: false,
		},
		{
			name: "environment condition met",
			condition: StepCondition{
				Type:  "environment",
				Field: "ENV_VAR",
				Value: "production",
			},
			execContext: ExecutionContext{
				Environment: map[string]string{"ENV_VAR": "production"},
			},
			expected: true,
		},
		{
			name: "environment condition not met",
			condition: StepCondition{
				Type:  "environment",
				Field: "ENV_VAR",
				Value: "production",
			},
			execContext: ExecutionContext{
				Environment: map[string]string{"ENV_VAR": "development"},
			},
			expected: false,
		},
		{
			name: "environment field not found",
			condition: StepCondition{
				Type:  "environment",
				Field: "MISSING_VAR",
				Value: "any",
			},
			execContext: ExecutionContext{
				Environment: map[string]string{"OTHER_VAR": "value"},
			},
			expected: false,
		},
		{
			name: "unknown condition type returns true",
			condition: StepCondition{
				Type:  "unknown",
				Field: "field",
				Value: "value",
			},
			execContext: ExecutionContext{},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := evaluateCondition(tt.condition, tt.execContext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test workflow step grouping by dependencies
func TestGroupStepsByDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		steps          []WorkflowStep
		expectedGroups int
		description    string
	}{
		{
			name: "no parallel steps",
			steps: []WorkflowStep{
				{Name: "step1", Parallel: false},
				{Name: "step2", Parallel: false},
				{Name: "step3", Parallel: false},
			},
			expectedGroups: 3,
			description:    "Each non-parallel step should be in its own group",
		},
		{
			name: "all parallel steps",
			steps: []WorkflowStep{
				{Name: "step1", Parallel: true},
				{Name: "step2", Parallel: true},
				{Name: "step3", Parallel: true},
			},
			expectedGroups: 1,
			description:    "All parallel steps should be in the same group",
		},
		{
			name: "mixed parallel and sequential",
			steps: []WorkflowStep{
				{Name: "step1", Parallel: false},
				{Name: "step2", Parallel: true},
				{Name: "step3", Parallel: true},
				{Name: "step4", Parallel: false},
			},
			expectedGroups: 2,
			description:    "Current implementation groups adjacent parallel steps together",
		},
		{
			name:           "empty steps",
			steps:          []WorkflowStep{},
			expectedGroups: 0,
			description:    "Empty steps should result in no groups",
		},
		{
			name: "single step",
			steps: []WorkflowStep{
				{Name: "only-step", Parallel: false},
			},
			expectedGroups: 1,
			description:    "Single step should create one group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			groups := groupStepsByDependencies(tt.steps)
			assert.Len(t, groups, tt.expectedGroups, tt.description)

			// Verify all steps are included
			totalSteps := 0
			for _, group := range groups {
				totalSteps += len(group)
			}
			assert.Equal(t, len(tt.steps), totalSteps, "All steps should be included in groups")

			// Verify no empty groups
			for i, group := range groups {
				assert.NotEmpty(t, group, "Group %d should not be empty", i)
			}
		})
	}
}

// Test workflow template creation
func TestCreateWorkflowFromTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		workflowName  string
		templateType  string
		expectedDesc  string
		expectedSteps int
	}{
		{
			name:          "basic template",
			workflowName:  "test-basic",
			templateType:  "basic",
			expectedDesc:  "Basic workflow template",
			expectedSteps: 1,
		},
		{
			name:          "ci template",
			workflowName:  "test-ci",
			templateType:  "ci",
			expectedDesc:  "Continuous Integration workflow",
			expectedSteps: 4,
		},
		{
			name:          "deploy template",
			workflowName:  "test-deploy",
			templateType:  "deploy",
			expectedDesc:  "Deployment workflow",
			expectedSteps: 2,
		},
		{
			name:          "unknown template defaults to basic",
			workflowName:  "test-unknown",
			templateType:  "unknown",
			expectedDesc:  "Basic workflow template",
			expectedSteps: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workflow := createWorkflowFromTemplate(tt.workflowName, tt.templateType)

			assert.Equal(t, tt.workflowName, workflow.Name)
			assert.Equal(t, tt.expectedDesc, workflow.Description)
			assert.Equal(t, "1.0.0", workflow.Version)
			assert.Len(t, workflow.Steps, tt.expectedSteps)
			assert.NotNil(t, workflow.Variables)
			assert.NotNil(t, workflow.Settings.Environment)
			assert.Equal(t, "30m", workflow.Settings.Timeout)
			assert.Equal(t, 3, workflow.Settings.MaxRetries)
			assert.Equal(t, "stop", workflow.Settings.FailureStrategy)
			assert.Equal(t, "on_failure", workflow.Settings.NotificationMode)
			assert.False(t, workflow.LastUpdated.IsZero())
		})
	}
}

// Test workflow template details
func TestGetWorkflowTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		templateType     string
		expectedSteps    []string
		expectedCommands []string
	}{
		{
			name:             "basic template",
			templateType:     "basic",
			expectedSteps:    []string{"hello"},
			expectedCommands: []string{"echo"},
		},
		{
			name:             "ci template",
			templateType:     "ci",
			expectedSteps:    []string{"checkout", "build", "test", "lint"},
			expectedCommands: []string{"git", "go", "go", "golangci-lint"},
		},
		{
			name:             "deploy template",
			templateType:     "deploy",
			expectedSteps:    []string{"build", "deploy"},
			expectedCommands: []string{"go", "kubectl"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			template := getWorkflowTemplate(tt.templateType)

			assert.Len(t, template.Steps, len(tt.expectedSteps))

			for i, expectedStep := range tt.expectedSteps {
				assert.Equal(t, expectedStep, template.Steps[i].Name)
				assert.Equal(t, tt.expectedCommands[i], template.Steps[i].Command)
				assert.Equal(t, "shell", template.Steps[i].Type)
			}
		})
	}
}

// Integration test for workflow discovery with temporary files
func TestWorkflowDiscovery_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Parallel()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "workflow-test")
	require.NoError(t, err)
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	// Create test workflow files
	workflow1 := createTestWorkflowDefinition()
	workflow1.Name = "workflow1"
	workflow1Data, err := json.MarshalIndent(workflow1, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "workflow1.json"), workflow1Data, 0o600)
	require.NoError(t, err)

	workflow2 := createTestWorkflowDefinition()
	workflow2.Name = "workflow2"
	workflow2Data, err := json.MarshalIndent(workflow2, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tempDir, "workflow2.json"), workflow2Data, 0o600)
	require.NoError(t, err)

	// Create a non-JSON file that should be ignored
	err = os.WriteFile(filepath.Join(tempDir, "not-a-workflow.txt"), []byte("ignore me"), 0o600)
	require.NoError(t, err)

	// Create an invalid JSON file
	err = os.WriteFile(filepath.Join(tempDir, "invalid.json"), []byte("invalid json"), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name          string
		directory     string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "discover valid workflows",
			directory:     tempDir,
			expectedCount: 2, // Only valid JSON files should be discovered
			expectError:   false,
		},
		{
			name:          "nonexistent directory",
			directory:     "/nonexistent/path",
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflows, err := discoverWorkflows(tt.directory)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, workflows, tt.expectedCount)

				if tt.expectedCount > 0 {
					// Verify workflow names
					names := make([]string, len(workflows))
					for i, w := range workflows {
						names[i] = w.Name
					}
					assert.Contains(t, names, "workflow1")
					assert.Contains(t, names, "workflow2")
				}
			}
		})
	}
}

// Benchmark tests for performance-critical paths
func BenchmarkExecuteWorkflowSteps(b *testing.B) {
	execution := createTestExecution()
	execution.Workflow.Steps = []WorkflowStep{
		{Name: "step1", Type: "shell", Command: "echo", Args: []string{"hello"}},
		{Name: "step2", Type: "shell", Command: "echo", Args: []string{"world"}},
		{Name: "step3", Type: "shell", Command: "echo", Args: []string{"benchmark"}},
	}
	execution.Results = make([]StepResult, len(execution.Workflow.Steps))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := executeWorkflowSteps(&execution)
		_ = err // Intentionally ignore for benchmark
	}
}

func BenchmarkValidateWorkflowDefinition(b *testing.B) {
	workflow := createTestWorkflowDefinition()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateWorkflowDefinition(&workflow)
	}
}

func BenchmarkEvaluateStepConditions(b *testing.B) {
	conditions := []StepCondition{
		{Type: "variable", Field: "var1", Value: "value1"},
		{Type: "environment", Field: "env1", Value: "envvalue1"},
		{Type: "variable", Field: "var2", Value: "value2"},
	}
	execContext := ExecutionContext{
		Variables: map[string]interface{}{
			"var1": "value1",
			"var2": "value2",
		},
		Environment: map[string]string{
			"env1": "envvalue1",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = evaluateStepConditions(conditions, execContext)
	}
}

func BenchmarkGroupStepsByDependencies(b *testing.B) {
	steps := []WorkflowStep{
		{Name: "step1", Parallel: false},
		{Name: "step2", Parallel: true},
		{Name: "step3", Parallel: true},
		{Name: "step4", Parallel: false},
		{Name: "step5", Parallel: true},
		{Name: "step6", Parallel: true},
		{Name: "step7", Parallel: false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = groupStepsByDependencies(steps)
	}
}

// Fuzz tests for input validation
func FuzzValidateWorkflowDefinition(f *testing.F) {
	// Add seed corpus
	validWorkflow := createTestWorkflowDefinition()
	validData, err := json.Marshal(validWorkflow)
	if err != nil {
		return // Skip invalid JSON
	}
	f.Add(string(validData))

	// Add edge cases
	f.Add(`{"name":"","steps":[]}`)
	f.Add(`{"name":"test","steps":[{"name":"","type":""}]}`)

	f.Fuzz(func(t *testing.T, data string) {
		var workflow WorkflowDefinition
		if err := json.Unmarshal([]byte(data), &workflow); err != nil {
			return // Skip invalid JSON
		}

		// This should not panic regardless of input
		_ = validateWorkflowDefinition(&workflow)
	})
}

func FuzzEvaluateCondition(f *testing.F) {
	// Add seed corpus
	f.Add("variable", "test_field", "test_value")
	f.Add("environment", "ENV_VAR", "prod")
	f.Add("unknown", "field", "value")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, condType, field, value string) {
		condition := StepCondition{
			Type:  condType,
			Field: field,
			Value: value,
		}
		execContext := ExecutionContext{
			Variables: map[string]interface{}{
				"test_field": "test_value",
				"other_var":  123,
			},
			Environment: map[string]string{
				"ENV_VAR":   "prod",
				"OTHER_ENV": "dev",
			},
		}

		// This should not panic regardless of input
		_ = evaluateCondition(condition, execContext)
	})
}

// Example tests for documentation
func ExampleWorkflow_Execute() {
	// Set up environment
	err := os.Setenv("WORKFLOW", "example-workflow")
	if err != nil {
		fmt.Printf("Failed to set environment variable: %v\n", err)
		return
	}
	defer func() {
		if unsetErr := os.Unsetenv("WORKFLOW"); unsetErr != nil {
			fmt.Printf("Failed to unset environment variable: %v\n", unsetErr)
		}
	}()

	workflow := Workflow{}
	err = workflow.Execute()
	if err != nil {
		fmt.Printf("Workflow execution failed: %v\n", err)
	}
}

func Example_createWorkflowFromTemplate() {
	workflow := createWorkflowFromTemplate("my-ci-workflow", "ci")
	fmt.Printf("Created workflow: %s\n", workflow.Name)
	fmt.Printf("Description: %s\n", workflow.Description)
	fmt.Printf("Steps: %d\n", len(workflow.Steps))
	// Output:
	// Created workflow: my-ci-workflow
	// Description: Continuous Integration workflow
	// Steps: 4
}

func Example_validateWorkflowDefinition() {
	workflow := WorkflowDefinition{
		Name:        "example-workflow",
		Description: "Example workflow",
		Steps: []WorkflowStep{
			{
				Name:    "hello",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"Hello, World!"},
			},
		},
	}

	isValid := validateWorkflowDefinition(&workflow)
	fmt.Printf("Workflow is valid: %t\n", isValid)
	// Output:
	// Workflow is valid: true
}

func Example_evaluateStepConditions() {
	conditions := []StepCondition{
		{Type: "variable", Field: "deploy_env", Value: "production"},
		{Type: "environment", Field: "CI", Value: "true"},
	}

	execContext := ExecutionContext{
		Variables:   map[string]interface{}{"deploy_env": "production"},
		Environment: map[string]string{"CI": "true"},
	}

	shouldExecute := evaluateStepConditions(conditions, execContext)
	fmt.Printf("Step should execute: %t\n", shouldExecute)
	// Output:
	// Step should execute: true
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Parallel()

	t.Run("generateExecutionID", func(t *testing.T) {
		id1 := generateExecutionID()
		time.Sleep(1 * time.Second) // Ensure different timestamp
		id2 := generateExecutionID()

		assert.NotEqual(t, id1, id2)
		assert.True(t, strings.HasPrefix(id1, "exec-"))
		assert.True(t, strings.HasPrefix(id2, "exec-"))
	})

	t.Run("getWorkflowPath", func(t *testing.T) {
		path := getWorkflowPath("test-workflow")
		assert.True(t, strings.HasSuffix(path, "test-workflow.json"))
		assert.Contains(t, path, ".mage/workflows")
	})

	t.Run("getWorkflowsDirectory", func(t *testing.T) {
		dir := getWorkflowsDirectory()
		assert.Equal(t, filepath.Join(".mage", "workflows"), dir)
	})

	t.Run("getExecutionsDirectory", func(t *testing.T) {
		dir := getExecutionsDirectory()
		assert.Equal(t, filepath.Join(".mage", "executions"), dir)
	})
}

// Test error conditions and edge cases
func TestErrorConditions(t *testing.T) {
	t.Parallel()

	t.Run("executeWorkflowStep_context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		step := WorkflowStep{
			Name:    "canceled-step",
			Type:    "shell",
			Command: "echo",
			Args:    []string{"hello"},
		}

		result := executeWorkflowStep(ctx, &step, ExecutionContext{})
		// Should complete despite cancellation since mock doesn't check context
		assert.NotEmpty(t, result.Step.Name)
	})

	t.Run("groupStepsByDependencies_empty_steps", func(t *testing.T) {
		groups := groupStepsByDependencies([]WorkflowStep{})
		assert.Empty(t, groups)
	})

	t.Run("evaluateStepConditions_empty_conditions", func(t *testing.T) {
		result := evaluateStepConditions([]StepCondition{}, ExecutionContext{})
		assert.True(t, result) // Empty conditions should return true
	})

	t.Run("executeShellCommand_invalid_timeout", func(t *testing.T) {
		ctx := context.Background()
		step := WorkflowStep{
			Name:    "invalid-timeout",
			Type:    "shell",
			Command: "echo",
			Args:    []string{"hello"},
			Timeout: "invalid-duration",
		}

		// Should not panic with invalid timeout
		output, err := executeShellCommand(ctx, &step)
		require.NoError(t, err) // Mock implementation succeeds
		assert.NotEmpty(t, output)
	})
}

// Race condition tests
func TestConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency tests in short mode")
	}

	t.Parallel()

	t.Run("parallel_workflow_execution", func(t *testing.T) {
		execution := createTestExecution()
		execution.Workflow.Steps = []WorkflowStep{
			{Name: "parallel1", Type: "shell", Command: "echo", Args: []string{"1"}, Parallel: true},
			{Name: "parallel2", Type: "shell", Command: "echo", Args: []string{"2"}, Parallel: true},
			{Name: "parallel3", Type: "shell", Command: "echo", Args: []string{"3"}, Parallel: true},
		}
		execution.Results = make([]StepResult, len(execution.Workflow.Steps))

		// Run with race detector to catch any race conditions
		err := executeWorkflowSteps(&execution)
		require.NoError(t, err)

		// Verify all steps were executed
		assert.Len(t, execution.Results, 3)
		for _, result := range execution.Results {
			assert.NotEmpty(t, result.Step.Name)
		}
	})

	t.Run("concurrent_workflow_validation", func(t *testing.T) {
		workflow := createTestWorkflowDefinition()

		// Run validation concurrently
		const numGoroutines = 10
		results := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				results <- validateWorkflowDefinition(&workflow)
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			result := <-results
			assert.True(t, result)
		}
	})

	t.Run("concurrent_condition_evaluation", func(t *testing.T) {
		conditions := []StepCondition{
			{Type: "variable", Field: "var1", Value: "value1"},
			{Type: "environment", Field: "env1", Value: "envvalue1"},
		}
		execContext := ExecutionContext{
			Variables:   map[string]interface{}{"var1": "value1"},
			Environment: map[string]string{"env1": "envvalue1"},
		}

		// Run condition evaluation concurrently
		const numGoroutines = 20
		results := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				results <- evaluateStepConditions(conditions, execContext)
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			result := <-results
			assert.True(t, result)
		}
	})
}

// Test complex workflow scenarios
func TestComplexWorkflowScenarios(t *testing.T) {
	t.Parallel()

	t.Run("workflow_with_mixed_step_types", func(t *testing.T) {
		execution := createTestExecution()
		execution.Workflow.Steps = []WorkflowStep{
			{Name: "setup", Type: "shell", Command: "echo", Args: []string{"setup"}},
			{Name: "build", Type: "command", Command: "make", Args: []string{"build"}},
			{Name: "test", Type: "script", Command: "test.sh"},
			{Name: "notify", Type: "notification", Command: "slack", Args: []string{"#ci", "Build complete"}},
		}
		execution.Results = make([]StepResult, len(execution.Workflow.Steps))

		err := executeWorkflowSteps(&execution)
		require.NoError(t, err)

		// Verify all steps completed
		for _, result := range execution.Results {
			assert.Equal(t, "completed", result.Status)
		}
	})

	t.Run("workflow_with_conditional_steps", func(t *testing.T) {
		execution := createTestExecution()
		execution.Context.Variables["deploy_env"] = "production"
		execution.Context.Environment["CI"] = "true"

		execution.Workflow.Steps = []WorkflowStep{
			{
				Name:    "always-run",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"always"},
			},
			{
				Name:    "conditional-run",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"conditional"},
				Conditions: []StepCondition{
					{Type: "variable", Field: "deploy_env", Value: "production"},
				},
			},
			{
				Name:    "skipped-step",
				Type:    "shell",
				Command: "echo",
				Args:    []string{"skipped"},
				Conditions: []StepCondition{
					{Type: "variable", Field: "deploy_env", Value: "development"},
				},
			},
		}
		execution.Results = make([]StepResult, len(execution.Workflow.Steps))

		err := executeWorkflowSteps(&execution)
		require.NoError(t, err)

		// Verify step statuses
		assert.Equal(t, "completed", execution.Results[0].Status) // always-run
		assert.Equal(t, "completed", execution.Results[1].Status) // conditional-run
		assert.Equal(t, "skipped", execution.Results[2].Status)   // skipped-step
	})
}
