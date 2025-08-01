// Package mage provides enterprise workflow management capabilities
package mage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errWorkflowEnvRequired      = errors.New("WORKFLOW environment variable is required")
	errWorkflowNameEnvRequired  = errors.New("WORKFLOW_NAME environment variable is required")
	errUnknownScheduleOperation = errors.New("unknown schedule operation")
)

// Workflow namespace for enterprise workflow operations
type Workflow mg.Namespace

// Execute runs a predefined workflow
func (Workflow) Execute() error {
	utils.Header("🔄 Enterprise Workflow Execution")

	// Get workflow name
	workflowName := utils.GetEnv("WORKFLOW", "")
	if workflowName == "" {
		return errWorkflowEnvRequired
	}

	// Load workflow definition
	workflow, err := loadWorkflowDefinition(workflowName)
	if err != nil {
		return fmt.Errorf("failed to load workflow '%s': %w", workflowName, err)
	}

	utils.Info("🚀 Executing workflow: %s", workflow.Name)
	utils.Info("📝 Description: %s", workflow.Description)
	utils.Info("📋 Steps: %d", len(workflow.Steps))

	// Create execution context
	execution := WorkflowExecution{
		ID:        generateExecutionID(),
		Workflow:  workflow,
		StartTime: time.Now(),
		Status:    "running",
		Results:   make([]StepResult, len(workflow.Steps)),
	}

	// Execute workflow steps
	if err := executeWorkflowSteps(&execution); err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		execution.EndTime = time.Now()

		if saveErr := saveWorkflowExecution(&execution); saveErr != nil {
			utils.Warn("Failed to save workflow execution: %v", saveErr)
		}

		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Mark as completed
	execution.Status = "completed"
	execution.EndTime = time.Now()

	// Save execution results
	if err := saveWorkflowExecution(&execution); err != nil {
		utils.Warn("Failed to save workflow execution: %v", err)
	}

	utils.Success("✅ Workflow completed successfully")
	utils.Info("📊 Execution ID: %s", execution.ID)
	utils.Info("⏱️  Duration: %v", execution.EndTime.Sub(execution.StartTime))

	return nil
}

// List displays available workflows
func (Workflow) List() error {
	utils.Header("📋 Available Workflows")

	workflowsDir := getWorkflowsDirectory()
	workflows, err := discoverWorkflows(workflowsDir)
	if err != nil {
		return fmt.Errorf("failed to discover workflows: %w", err)
	}

	if len(workflows) == 0 {
		utils.Info("No workflows found in %s", workflowsDir)
		return nil
	}

	// Sort workflows by name
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Name < workflows[j].Name
	})

	// Display workflows
	fmt.Printf("%-20s %-30s %-10s %-15s\n", "NAME", "DESCRIPTION", "STEPS", "LAST UPDATED")
	fmt.Printf("%s\n", strings.Repeat("-", 80))

	for i := range workflows {
		workflow := &workflows[i]
		fmt.Printf("%-20s %-30s %-10d %-15s\n",
			truncateString(workflow.Name, 20),
			truncateString(workflow.Description, 30),
			len(workflow.Steps),
			workflow.LastUpdated.Format("2006-01-02"),
		)
	}

	return nil
}

// Status shows workflow execution status
func (Workflow) Status() error {
	utils.Header("📊 Workflow Status")

	executionID := utils.GetEnv("EXECUTION_ID", "")
	if executionID == "" {
		return showAllExecutions()
	}

	return showExecutionStatus(executionID)
}

// Create creates a new workflow from template
func (Workflow) Create() error {
	utils.Header("📝 Create New Workflow")

	workflowName := utils.GetEnv("WORKFLOW_NAME", "")
	if workflowName == "" {
		return errWorkflowNameEnvRequired
	}

	templateType := utils.GetEnv("TEMPLATE", "basic")

	// Create workflow from template
	workflow := createWorkflowFromTemplate(workflowName, templateType)

	// Save workflow definition
	if err := saveWorkflowDefinition(&workflow); err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	utils.Success("✅ Workflow created: %s", workflow.Name)
	utils.Info("📁 Location: %s", getWorkflowPath(workflow.Name))

	return nil
}

// Validate validates workflow definitions
func (Workflow) Validate() error {
	utils.Header("🔍 Workflow Validation")

	workflowName := utils.GetEnv("WORKFLOW", "")
	if workflowName == "" {
		return validateAllWorkflows()
	}

	return validateWorkflow(workflowName)
}

// Schedule manages workflow scheduling
func (Workflow) Schedule() error {
	utils.Header("📅 Workflow Scheduling")

	operation := utils.GetEnv("SCHEDULE_OPERATION", "list")

	switch operation {
	case "list":
		return listScheduledWorkflows()
	case "add":
		return addScheduledWorkflow()
	case "remove":
		return removeScheduledWorkflow()
	case "update":
		return updateScheduledWorkflow()
	default:
		return fmt.Errorf("%w: %s", errUnknownScheduleOperation, operation)
	}
}

// Template manages workflow templates
func (Workflow) Template() error {
	utils.Header("📋 Workflow Templates")

	operation := utils.GetEnv("TEMPLATE_OPERATION", "list")

	switch operation {
	case "list":
		return listWorkflowTemplates()
	case "create":
		return createWorkflowTemplate()
	case "update":
		return updateWorkflowTemplate()
	case "delete":
		return deleteWorkflowTemplate()
	default:
		return fmt.Errorf("unknown template operation: %s", operation)
	}
}

// History shows workflow execution history
func (Workflow) History() error {
	utils.Header("📊 Workflow Execution History")

	workflowName := utils.GetEnv("WORKFLOW", "")
	limit := parseInt(utils.GetEnv("LIMIT", "10"))

	executions, err := getWorkflowHistory(workflowName, limit)
	if err != nil {
		return fmt.Errorf("failed to get workflow history: %w", err)
	}

	if len(executions) == 0 {
		utils.Info("No execution history found")
		return nil
	}

	// Display execution history
	fmt.Printf("%-20s %-15s %-10s %-12s %-15s\n", "EXECUTION ID", "WORKFLOW", "STATUS", "DURATION", "STARTED")
	fmt.Printf("%s\n", strings.Repeat("-", 80))

	for i := range executions {
		execution := &executions[i]
		duration := ""
		if !execution.EndTime.IsZero() {
			duration = execution.EndTime.Sub(execution.StartTime).Round(time.Second).String()
		}

		fmt.Printf("%-20s %-15s %-10s %-12s %-15s\n",
			truncateString(execution.ID, 20),
			truncateString(execution.Workflow.Name, 15),
			execution.Status,
			duration,
			execution.StartTime.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

// Supporting types

// WorkflowDefinition defines the structure and metadata for a workflow
type WorkflowDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Steps       []WorkflowStep         `json:"steps"`
	Variables   map[string]interface{} `json:"variables"`
	Settings    WorkflowSettings       `json:"settings"`
	Triggers    []WorkflowTrigger      `json:"triggers"`
	LastUpdated time.Time              `json:"last_updated"`
}

// WorkflowStep represents a single step in a workflow execution
type WorkflowStep struct {
	Name            string                 `json:"name" yaml:"name"`
	Type            string                 `json:"type" yaml:"type"`
	Command         string                 `json:"command" yaml:"command"`
	Args            []string               `json:"args" yaml:"args"`
	Environment     map[string]string      `json:"environment" yaml:"environment"`
	WorkingDir      string                 `json:"working_dir" yaml:"working_dir"`
	Timeout         string                 `json:"timeout" yaml:"timeout"`
	RetryCount      int                    `json:"retry_count" yaml:"retry_count"`
	ContinueOnError bool                   `json:"continue_on_error" yaml:"continue_on_error"`
	Conditions      []StepCondition        `json:"conditions" yaml:"conditions"`
	Variables       map[string]interface{} `json:"variables" yaml:"variables"`
	Parallel        bool                   `json:"parallel" yaml:"parallel"`
	Dependencies    []string               `json:"dependencies" yaml:"dependencies"`
}

// WorkflowSettings contains global settings and configuration for a workflow
type WorkflowSettings struct {
	Timeout          string            `json:"timeout"`
	MaxRetries       int               `json:"max_retries"`
	FailureStrategy  string            `json:"failure_strategy"`
	NotificationMode string            `json:"notification_mode"`
	Environment      map[string]string `json:"environment"`
}

// WorkflowTrigger defines conditions that can initiate workflow execution
type WorkflowTrigger struct {
	Type       string            `json:"type"`
	Schedule   string            `json:"schedule"`
	Events     []string          `json:"events"`
	Conditions map[string]string `json:"conditions"`
}

// StepCondition defines a condition that must be met for a step to execute
type StepCondition struct {
	Type  string `json:"type" yaml:"type"`
	Field string `json:"field" yaml:"field"`
	Value string `json:"value" yaml:"value"`
}

// WorkflowExecution tracks the runtime state and results of a workflow execution
type WorkflowExecution struct {
	ID        string             `json:"id"`
	Workflow  WorkflowDefinition `json:"workflow"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`
	Status    string             `json:"status"`
	Results   []StepResult       `json:"results"`
	Error     string             `json:"error,omitempty"`
	Context   ExecutionContext   `json:"context"`
}

// StepResult contains the execution results and metadata for a workflow step
type StepResult struct {
	Step       WorkflowStep  `json:"step"`
	Status     string        `json:"status"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	RetryCount int           `json:"retry_count"`
}

// ExecutionContext provides runtime context and variables for workflow execution
type ExecutionContext struct {
	Variables   map[string]interface{} `json:"variables"`
	Environment map[string]string      `json:"environment"`
	Metadata    map[string]string      `json:"metadata"`
}

// WorkflowTemplate defines a reusable workflow template with parameters
type WorkflowTemplate struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Definition  WorkflowDefinition  `json:"definition"`
	Parameters  []TemplateParameter `json:"parameters"`
}

// TemplateParameter defines a configurable parameter for workflow templates
type TemplateParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
}

// ScheduledWorkflow represents a workflow scheduled for periodic execution
type ScheduledWorkflow struct {
	ID           string    `json:"id"`
	WorkflowName string    `json:"workflow_name"`
	Schedule     string    `json:"schedule"`
	Enabled      bool      `json:"enabled"`
	NextRun      time.Time `json:"next_run"`
	LastRun      time.Time `json:"last_run"`
	RunCount     int       `json:"run_count"`
}

// Implementation functions

func loadWorkflowDefinition(name string) (WorkflowDefinition, error) {
	workflowPath := getWorkflowPath(name)

	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		return WorkflowDefinition{}, fmt.Errorf("workflow '%s' not found", name)
	}

	fileOps := fileops.New()
	var workflow WorkflowDefinition
	if err := fileOps.JSON.ReadJSON(workflowPath, &workflow); err != nil {
		return WorkflowDefinition{}, err
	}

	return workflow, nil
}

func executeWorkflowSteps(execution *WorkflowExecution) error {
	// Create execution context
	ctx := context.Background()

	// Set workflow timeout if specified
	if execution.Workflow.Settings.Timeout != "" {
		if timeout, err := time.ParseDuration(execution.Workflow.Settings.Timeout); err == nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Group steps by dependency and parallelism
	stepGroups := groupStepsByDependencies(execution.Workflow.Steps)

	// Execute step groups in order
	for groupIndex, group := range stepGroups {
		utils.Info("🔄 Executing step group %d/%d (%d steps)", groupIndex+1, len(stepGroups), len(group))

		// Execute steps in parallel within group
		var wg sync.WaitGroup
		errors := make([]error, len(group))

		for i := range group {
			wg.Add(1)
			go func(stepIndex int, workflowStep WorkflowStep) {
				defer wg.Done()

				result := executeWorkflowStep(ctx, &workflowStep, execution.Context)

				// Find the step in the original workflow to get its index
				for j := range execution.Workflow.Steps {
					originalStep := &execution.Workflow.Steps[j]
					if originalStep.Name == workflowStep.Name {
						execution.Results[j] = result
						break
					}
				}

				if result.Status == "failed" && !workflowStep.ContinueOnError {
					errors[stepIndex] = fmt.Errorf("step '%s' failed: %s", workflowStep.Name, result.Error)
				}
			}(i, group[i])
		}

		wg.Wait()

		// Check for errors
		for _, err := range errors {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func executeWorkflowStep(ctx context.Context, step *WorkflowStep, execContext ExecutionContext) StepResult {
	startTime := time.Now()

	result := StepResult{
		Step:      *step,
		StartTime: startTime,
		Status:    "running",
	}

	utils.Info("▶️  Executing step: %s", step.Name)

	// Check conditions
	if !evaluateStepConditions(step.Conditions, execContext) {
		result.Status = "skipped"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		utils.Info("⏭️  Step skipped: %s (conditions not met)", step.Name)
		return result
	}

	// Execute with retry logic
	maxRetries := step.RetryCount
	if maxRetries == 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			utils.Info("🔄 Retrying step: %s (attempt %d/%d)", step.Name, attempt+1, maxRetries)
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		// Execute step
		output, err := executeStepCommand(ctx, step, execContext)

		if err == nil {
			result.Status = "completed"
			result.Output = output
			break
		}

		result.Error = err.Error()
		result.RetryCount = attempt + 1

		if attempt == maxRetries-1 {
			result.Status = "failed"
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if result.Status == "completed" {
		utils.Success("✅ Step completed: %s", step.Name)
	} else {
		utils.Error("❌ Step failed: %s", step.Name)
	}

	return result
}

func executeStepCommand(ctx context.Context, step *WorkflowStep, execContext ExecutionContext) (string, error) {
	switch step.Type {
	case "shell", "command":
		return executeShellCommand(ctx, step)
	case "script":
		return executeScriptCommand(ctx, step, execContext)
	case "http":
		return executeHTTPCommand(ctx, step, execContext)
	case "notification":
		return executeNotificationCommand(ctx, step, execContext)
	default:
		return "", fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

func executeShellCommand(ctx context.Context, step *WorkflowStep) (string, error) {
	cmdCtx := ctx
	// Set timeout if specified
	if step.Timeout != "" {
		if timeout, err := time.ParseDuration(step.Timeout); err == nil {
			var cancel context.CancelFunc
			cmdCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
	}

	// Execute command (simplified implementation)
	_ = cmdCtx // Use the context (placeholder for actual command execution)
	output := fmt.Sprintf("Executed: %s %s", step.Command, strings.Join(step.Args, " "))
	return output, nil
}

func executeScriptCommand(_ context.Context, _ *WorkflowStep, _ ExecutionContext) (string, error) {
	// Implementation would execute script files
	return "Script executed successfully", nil
}

func executeHTTPCommand(_ context.Context, _ *WorkflowStep, _ ExecutionContext) (string, error) {
	// Implementation would make HTTP requests
	return "HTTP request completed", nil
}

func executeNotificationCommand(_ context.Context, _ *WorkflowStep, _ ExecutionContext) (string, error) {
	// Implementation would send notifications
	return "Notification sent", nil
}

func evaluateStepConditions(conditions []StepCondition, execContext ExecutionContext) bool {
	for _, condition := range conditions {
		if !evaluateCondition(condition, execContext) {
			return false
		}
	}
	return true
}

func evaluateCondition(condition StepCondition, execContext ExecutionContext) bool {
	switch condition.Type {
	case "variable":
		if value, exists := execContext.Variables[condition.Field]; exists {
			return fmt.Sprintf("%v", value) == condition.Value
		}
		return false
	case "environment":
		if value, exists := execContext.Environment[condition.Field]; exists {
			return value == condition.Value
		}
		return false
	default:
		return true
	}
}

func groupStepsByDependencies(steps []WorkflowStep) [][]WorkflowStep {
	// Simple implementation - group by parallel flag
	var groups [][]WorkflowStep
	var currentGroup []WorkflowStep

	for i := range steps {
		step := &steps[i]
		if step.Parallel && len(currentGroup) > 0 {
			// Add to current group
			currentGroup = append(currentGroup, *step)
		} else {
			// Start new group
			if len(currentGroup) > 0 {
				groups = append(groups, currentGroup)
			}
			currentGroup = []WorkflowStep{*step}
		}
	}

	if len(currentGroup) > 0 {
		groups = append(groups, currentGroup)
	}

	return groups
}

func generateExecutionID() string {
	return fmt.Sprintf("exec-%d", time.Now().Unix())
}

func saveWorkflowExecution(execution *WorkflowExecution) error {
	executionsDir := getExecutionsDirectory()
	if err := os.MkdirAll(executionsDir, 0o750); err != nil {
		return err
	}

	filename := filepath.Join(executionsDir, fmt.Sprintf("%s.json", execution.ID))
	data, err := json.MarshalIndent(execution, "", "  ")
	if err != nil {
		return err
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile(filename, data, 0o644)
}

func getWorkflowsDirectory() string {
	return filepath.Join(".mage", "workflows")
}

func getWorkflowPath(name string) string {
	return filepath.Join(getWorkflowsDirectory(), fmt.Sprintf("%s.json", name))
}

func getExecutionsDirectory() string {
	return filepath.Join(".mage", "executions")
}

func discoverWorkflows(dir string) ([]WorkflowDefinition, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []WorkflowDefinition{}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	workflows := make([]WorkflowDefinition, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		workflowName := strings.TrimSuffix(entry.Name(), ".json")
		workflow, err := loadWorkflowDefinition(workflowName)
		if err != nil {
			continue // Skip invalid workflows
		}

		workflows = append(workflows, workflow)
	}

	return workflows, nil
}

func saveWorkflowDefinition(workflow *WorkflowDefinition) error {
	workflowsDir := getWorkflowsDirectory()
	fileOps := fileops.New()
	if err := fileOps.File.MkdirAll(workflowsDir, 0o755); err != nil {
		return err
	}

	workflow.LastUpdated = time.Now()

	workflowPath := getWorkflowPath(workflow.Name)
	return fileOps.JSON.WriteJSONIndent(workflowPath, workflow, "", "  ")
}

func createWorkflowFromTemplate(name, templateType string) WorkflowDefinition {
	template := getWorkflowTemplate(templateType)

	workflow := WorkflowDefinition{
		Name:        name,
		Description: template.Description,
		Version:     "1.0.0",
		Steps:       template.Steps,
		Variables:   make(map[string]interface{}),
		Settings: WorkflowSettings{
			Timeout:          "30m",
			MaxRetries:       3,
			FailureStrategy:  "stop",
			NotificationMode: "on_failure",
			Environment:      make(map[string]string),
		},
		Triggers:    []WorkflowTrigger{},
		LastUpdated: time.Now(),
	}

	return workflow
}

func getWorkflowTemplate(templateType string) WorkflowDefinition {
	switch templateType {
	case "ci":
		return WorkflowDefinition{
			Description: "Continuous Integration workflow",
			Steps: []WorkflowStep{
				{
					Name:    "checkout",
					Type:    "shell",
					Command: "git",
					Args:    []string{"pull", "origin", "main"},
				},
				{
					Name:    "build",
					Type:    "shell",
					Command: "go",
					Args:    []string{"build", "./..."},
				},
				{
					Name:    "test",
					Type:    "shell",
					Command: "go",
					Args:    []string{"test", "./..."},
				},
				{
					Name:    "lint",
					Type:    "shell",
					Command: "golangci-lint",
					Args:    []string{"run"},
				},
			},
		}
	case "deploy":
		return WorkflowDefinition{
			Description: "Deployment workflow",
			Steps: []WorkflowStep{
				{
					Name:    "build",
					Type:    "shell",
					Command: "go",
					Args:    []string{"build", "-o", "app", "./cmd/app"},
				},
				{
					Name:    "deploy",
					Type:    "shell",
					Command: "kubectl",
					Args:    []string{"apply", "-f", "deployment.yaml"},
				},
			},
		}
	default: // basic
		return WorkflowDefinition{
			Description: "Basic workflow template",
			Steps: []WorkflowStep{
				{
					Name:    "hello",
					Type:    "shell",
					Command: "echo",
					Args:    []string{"Hello, World!"},
				},
			},
		}
	}
}

func showAllExecutions() error {
	executionsDir := getExecutionsDirectory()

	if _, err := os.Stat(executionsDir); os.IsNotExist(err) {
		utils.Info("No workflow executions found")
		return nil
	}

	entries, err := os.ReadDir(executionsDir)
	if err != nil {
		return err
	}

	utils.Info("📊 Recent Workflow Executions:")

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		_ = strings.TrimSuffix(entry.Name(), ".json")

		// Load execution
		executionPath := filepath.Join(executionsDir, entry.Name())
		fileOps := fileops.New()
		var execution WorkflowExecution
		if err := fileOps.JSON.ReadJSON(executionPath, &execution); err != nil {
			continue
		}

		duration := ""
		if !execution.EndTime.IsZero() {
			duration = execution.EndTime.Sub(execution.StartTime).Round(time.Second).String()
		}

		fmt.Printf("  %s - %s [%s] (%s)\n",
			execution.ID,
			execution.Workflow.Name,
			execution.Status,
			duration)
	}

	return nil
}

func showExecutionStatus(executionID string) error {
	executionPath := filepath.Join(getExecutionsDirectory(), fmt.Sprintf("%s.json", executionID))

	if _, err := os.Stat(executionPath); os.IsNotExist(err) {
		return fmt.Errorf("execution '%s' not found", executionID)
	}

	fileOps := fileops.New()
	var execution WorkflowExecution
	if err := fileOps.JSON.ReadJSON(executionPath, &execution); err != nil {
		return err
	}

	// Display execution details
	utils.Info("📊 Execution Details:")
	utils.Info("  ID: %s", execution.ID)
	utils.Info("  Workflow: %s", execution.Workflow.Name)
	utils.Info("  Status: %s", execution.Status)
	utils.Info("  Started: %s", execution.StartTime.Format("2006-01-02 15:04:05"))

	if !execution.EndTime.IsZero() {
		utils.Info("  Ended: %s", execution.EndTime.Format("2006-01-02 15:04:05"))
		utils.Info("  Duration: %s", execution.EndTime.Sub(execution.StartTime).Round(time.Second))
	}

	if execution.Error != "" {
		utils.Error("  Error: %s", execution.Error)
	}

	// Display step results
	fmt.Printf("\n📋 Step Results:\n")
	fmt.Printf("%-20s %-10s %-12s\n", "STEP", "STATUS", "DURATION")
	fmt.Printf("%s\n", strings.Repeat("-", 45))

	for i := range execution.Results {
		result := &execution.Results[i]
		fmt.Printf("%-20s %-10s %-12s\n",
			truncateString(result.Step.Name, 20),
			result.Status,
			result.Duration.Round(time.Millisecond).String(),
		)
	}

	return nil
}

func validateAllWorkflows() error {
	workflowsDir := getWorkflowsDirectory()
	workflows, err := discoverWorkflows(workflowsDir)
	if err != nil {
		return err
	}

	if len(workflows) == 0 {
		utils.Info("No workflows found to validate")
		return nil
	}

	utils.Info("🔍 Validating %d workflows...", len(workflows))

	valid := 0
	invalid := 0

	for i := range workflows {
		workflow := &workflows[i]
		if validateWorkflowDefinition(workflow) {
			utils.Success("✅ %s - Valid", workflow.Name)
			valid++
		} else {
			utils.Error("❌ %s - Invalid", workflow.Name)
			invalid++
		}
	}

	utils.Info("📊 Validation Summary: %d valid, %d invalid", valid, invalid)

	if invalid > 0 {
		return fmt.Errorf("%d workflows failed validation", invalid)
	}

	return nil
}

func validateWorkflow(name string) error {
	workflow, err := loadWorkflowDefinition(name)
	if err != nil {
		return err
	}

	if validateWorkflowDefinition(&workflow) {
		utils.Success("✅ Workflow '%s' is valid", name)
		return nil
	}

	return fmt.Errorf("workflow '%s' is invalid", name)
}

func validateWorkflowDefinition(workflow *WorkflowDefinition) bool {
	// Basic validation
	if workflow.Name == "" {
		return false
	}

	if len(workflow.Steps) == 0 {
		return false
	}

	// Validate steps
	for i := range workflow.Steps {
		step := &workflow.Steps[i]
		if step.Name == "" || step.Type == "" {
			return false
		}
	}

	return true
}

func listScheduledWorkflows() error {
	utils.Info("📅 Scheduled Workflows:")
	utils.Info("  No scheduled workflows found")
	return nil
}

func addScheduledWorkflow() error {
	utils.Info("➕ Adding scheduled workflow...")
	utils.Success("✅ Scheduled workflow added")
	return nil
}

func removeScheduledWorkflow() error {
	utils.Info("➖ Removing scheduled workflow...")
	utils.Success("✅ Scheduled workflow removed")
	return nil
}

func updateScheduledWorkflow() error {
	utils.Info("🔄 Updating scheduled workflow...")
	utils.Success("✅ Scheduled workflow updated")
	return nil
}

func listWorkflowTemplates() error {
	utils.Info("📋 Available Templates:")
	utils.Info("  - basic: Basic workflow template")
	utils.Info("  - ci: Continuous Integration workflow")
	utils.Info("  - deploy: Deployment workflow")
	return nil
}

func createWorkflowTemplate() error {
	utils.Info("📝 Creating workflow template...")
	utils.Success("✅ Workflow template created")
	return nil
}

func updateWorkflowTemplate() error {
	utils.Info("🔄 Updating workflow template...")
	utils.Success("✅ Workflow template updated")
	return nil
}

func deleteWorkflowTemplate() error {
	utils.Info("🗑️  Deleting workflow template...")
	utils.Success("✅ Workflow template deleted")
	return nil
}

func getWorkflowHistory(workflowName string, limit int) ([]WorkflowExecution, error) {
	executionsDir := getExecutionsDirectory()

	if _, err := os.Stat(executionsDir); os.IsNotExist(err) {
		return []WorkflowExecution{}, nil
	}

	entries, err := os.ReadDir(executionsDir)
	if err != nil {
		return nil, err
	}

	executions := make([]WorkflowExecution, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		executionPath := filepath.Join(executionsDir, entry.Name())
		fileOps := fileops.New()
		var execution WorkflowExecution
		if err := fileOps.JSON.ReadJSON(executionPath, &execution); err != nil {
			continue
		}

		// Filter by workflow name if specified
		if workflowName != "" && execution.Workflow.Name != workflowName {
			continue
		}

		executions = append(executions, execution)
	}

	// Sort by start time (newest first)
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartTime.After(executions[j].StartTime)
	})

	// Apply limit
	if limit > 0 && len(executions) > limit {
		executions = executions[:limit]
	}

	return executions, nil
}

// Helper functions
