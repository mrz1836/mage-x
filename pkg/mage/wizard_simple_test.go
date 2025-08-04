package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple validation function tests that don't rely on complex mocking
func TestValidationFunctions(t *testing.T) {
	t.Run("validateNonEmpty", func(t *testing.T) {
		require.NoError(t, validateNonEmpty("test"))
		require.Error(t, validateNonEmpty(""))
		assert.Error(t, validateNonEmpty("   "))
	})

	t.Run("validateDomain", func(t *testing.T) {
		assert.NoError(t, validateDomain("example.com"))
		assert.NoError(t, validateDomain("sub.example.com"))
		require.Error(t, validateDomain(""))
		require.Error(t, validateDomain("invalid..domain"))
	})

	t.Run("validateDuration", func(t *testing.T) {
		assert.NoError(t, validateDuration("30m"))
		assert.NoError(t, validateDuration("1h30m"))
		require.Error(t, validateDuration(""))
		require.Error(t, validateDuration("invalid"))
	})

	t.Run("validateTimezone", func(t *testing.T) {
		assert.NoError(t, validateTimezone("UTC"))
		assert.NoError(t, validateTimezone("America/New_York"))
		assert.NoError(t, validateTimezone("Custom/Timezone"))
	})
}

// Test wizard creation without complex input simulation
func TestWizardCreation(t *testing.T) {
	t.Run("NewEnterpriseWizard", func(t *testing.T) {
		wizard := NewEnterpriseWizard()
		assert.NotNil(t, wizard)
		assert.Equal(t, "Enterprise Setup", wizard.GetName())
		assert.Equal(t, "Complete enterprise configuration setup", wizard.GetDescription())
		assert.NotNil(t, wizard.Context)
		assert.NotNil(t, wizard.Context.Config)
		assert.Len(t, wizard.Steps, 10)
	})

	t.Run("step properties", func(t *testing.T) {
		tests := []struct {
			step        WizardStep
			name        string
			description string
			required    bool
		}{
			{&WelcomeStep{}, "Welcome", "Welcome and introduction", false},
			{&OrganizationStep{}, "Organization Configuration", "Configure organization settings", true},
			{&SecurityStep{}, "Security Configuration", "Configure security settings", true},
			{&IntegrationsStep{}, "Integration Configuration", "Configure enterprise integrations", false},
			{&WorkflowsStep{}, "Workflow Configuration", "Configure workflow settings", false},
			{&AnalyticsStep{}, "Analytics Configuration", "Configure analytics and reporting", false},
			{&DeploymentStep{}, "Deployment Configuration", "Configure deployment settings", false},
			{&ComplianceStep{}, "Compliance Configuration", "Configure compliance and governance", false},
			{&SummaryStep{}, "Configuration Summary", "Review and confirm configuration", true},
			{&FinalizationStep{}, "Finalization", "Save and finalize configuration", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.name, tt.step.GetName())
				assert.Equal(t, tt.description, tt.step.GetDescription())
				assert.Equal(t, tt.required, tt.step.IsRequired())
				assert.Empty(t, tt.step.GetDependencies())
			})
		}
	})
}

// Test other wizard types
func TestOtherWizardTypes(t *testing.T) {
	tests := []struct {
		name         string
		wizard       InteractiveWizard
		expectedName string
		expectedDesc string
	}{
		{"ProjectWizard", NewProjectWizard(), "Project Configuration", "Configure project-specific settings"},
		{"IntegrationWizard", NewIntegrationWizard(), "Integration Setup", "Configure enterprise integrations"},
		{"SecurityWizard", NewSecurityWizard(), "Security Configuration", "Configure security settings"},
		{"WorkflowWizard", NewWorkflowWizard(), "Workflow Configuration", "Configure workflow settings"},
		{"DeploymentWizard", NewDeploymentWizard(), "Deployment Configuration", "Configure deployment settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedName, tt.wizard.GetName())
			assert.Equal(t, tt.expectedDesc, tt.wizard.GetDescription())
			assert.NoError(t, tt.wizard.Run()) // These currently return nil
		})
	}
}

// Test integration type mapping
func TestGetIntegrationType(t *testing.T) {
	step := &IntegrationsStep{}

	tests := []struct {
		name     string
		expected string
	}{
		{"slack", "communication"},
		{"github", "source_control"},
		{"gitlab", "source_control"},
		{"jira", "issue_tracking"},
		{"jenkins", "ci_cd"},
		{"docker", "containerization"},
		{"kubernetes", "orchestration"},
		{"prometheus", "monitoring"},
		{"grafana", "monitoring"},
		{"aws", "cloud_provider"},
		{"azure", "cloud_provider"},
		{"gcp", "cloud_provider"},
		{"unknown", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := step.getIntegrationType(tt.name)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test error constants
func TestErrorConstants(t *testing.T) {
	tests := []struct {
		err      error
		expected string
	}{
		{errConfigRejected, "configuration rejected by user"},
		{errValueEmpty, "value cannot be empty"},
		{errDomainEmpty, "domain cannot be empty"},
		{errInvalidDomainFormat, "invalid domain format"},
		{errDurationEmpty, "duration cannot be empty"},
		{errInvalidDurationFormat, "invalid duration format (e.g., 30m, 1h, 2h30m)"},
		{errWizardCanceled, "wizard canceled by user"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}
