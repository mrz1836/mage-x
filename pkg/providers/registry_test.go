package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// Error definitions for registry tests
var (
	ErrFactoryError = errors.New("factory error")
	ErrConfigIsNil  = errors.New("config is nil")
)

// RegistryTestSuite tests the provider registry in isolation
type RegistryTestSuite struct {
	suite.Suite

	registry *Registry
}

// SetupTest creates a new registry for each test
func (ts *RegistryTestSuite) SetupTest() {
	ts.registry = NewRegistry()
}

// TestRegistryCreation tests registry creation
func (ts *RegistryTestSuite) TestRegistryCreation() {
	ts.Run("New registry is empty", func() {
		ts.Require().NotNil(ts.registry)
		ts.Require().Empty(ts.registry.List())
		ts.Require().False(ts.registry.HasProvider("any"))
	})
}

// TestRegistryOperations tests basic registry operations
func (ts *RegistryTestSuite) TestRegistryOperations() {
	ts.Run("Register and retrieve provider", func() {
		factory := func(config *ProviderConfig) (Provider, error) {
			return &mockProvider{config: *config, providerName: "test"}, nil
		}

		ts.registry.Register("test", factory)
		ts.Require().True(ts.registry.HasProvider("test"))
		ts.Require().Contains(ts.registry.List(), "test")

		config := &ProviderConfig{Region: "us-east-1"}
		provider, err := ts.registry.Get("test", config)
		ts.Require().NoError(err)
		ts.Require().Equal("test", provider.Name())
	})

	ts.Run("Get non-existent provider", func() {
		_, err := ts.registry.Get("nonexistent", &ProviderConfig{})
		ts.Require().Error(err)
		ts.Require().ErrorIs(err, ErrProviderNotFound)
		ts.Require().Contains(err.Error(), "nonexistent")
	})

	ts.Run("Factory returns error", func() {
		factory := func(config *ProviderConfig) (Provider, error) {
			return nil, ErrFactoryError
		}

		ts.registry.Register("failing", factory)
		ts.Require().True(ts.registry.HasProvider("failing"))

		_, err := ts.registry.Get("failing", &ProviderConfig{})
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "factory error")
	})
}

// TestRegistryConcurrency tests thread safety of registry operations
func (ts *RegistryTestSuite) TestRegistryConcurrency() {
	ts.Run("Concurrent registration and access", func() {
		const numGoroutines = 50
		var wg sync.WaitGroup

		// Concurrent registration
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				providerName := "concurrent-" + string(rune('A'+id%26))
				factory := func(config *ProviderConfig) (Provider, error) {
					return &mockProvider{config: *config, providerName: providerName}, nil
				}

				ts.registry.Register(providerName, factory)
			}(i)
		}

		// Concurrent access
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(_ int) {
				defer wg.Done()

				// Try to access registry methods (use assert in goroutines)
				providers := ts.registry.List()
				ts.NotNil(providers)

				// Check if any provider exists
				if len(providers) > 0 {
					exists := ts.registry.HasProvider(providers[0])
					ts.True(exists)
				}
			}(i)
		}

		wg.Wait()

		// Verify final state
		providers := ts.registry.List()
		ts.Require().NotEmpty(providers)
		for _, providerName := range providers {
			ts.Require().True(ts.registry.HasProvider(providerName))
		}
	})
}

// TestRegistryManagerSingleton tests the singleton registry manager
func (ts *RegistryTestSuite) TestRegistryManagerSingleton() {
	ts.Run("Registry manager returns same instance", func() {
		// Get multiple instances of the default registry manager
		manager1 := getDefaultRegistryManager()
		manager2 := getDefaultRegistryManager()

		ts.Require().Same(manager1, manager2, "Should return same singleton instance")
	})

	ts.Run("Registry lazy initialization", func() {
		// Create a new registry manager for testing
		manager := newRegistryManager()

		// First call should create the registry
		reg1 := manager.getOrCreateRegistry()
		ts.Require().NotNil(reg1)

		// Second call should return the same registry
		reg2 := manager.getOrCreateRegistry()
		ts.Require().Same(reg1, reg2, "Should return same registry instance")
	})

	ts.Run("Registry manager thread safety", func() {
		manager := newRegistryManager()
		const numGoroutines = 20
		registries := make([]*Registry, numGoroutines)
		var wg sync.WaitGroup

		// Concurrent access to get registry
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				registries[index] = manager.getOrCreateRegistry()
			}(i)
		}

		wg.Wait()

		// All should be the same instance
		baseRegistry := registries[0]
		for i := 1; i < numGoroutines; i++ {
			ts.Require().Same(baseRegistry, registries[i],
				"All goroutines should get the same registry instance")
		}
	})
}

// TestRegistryEdgeCases tests edge cases and error conditions
func (ts *RegistryTestSuite) TestRegistryEdgeCases() {
	ts.Run("Empty provider name registration", func() {
		factory := func(config *ProviderConfig) (Provider, error) {
			return &mockProvider{}, nil
		}

		// Should not panic with empty name
		ts.Require().NotPanics(func() {
			ts.registry.Register("", factory)
		})

		ts.Require().True(ts.registry.HasProvider(""))

		provider, err := ts.registry.Get("", &ProviderConfig{})
		ts.Require().NoError(err)
		ts.Require().NotNil(provider)
	})

	ts.Run("Nil config handling", func() {
		factory := func(config *ProviderConfig) (Provider, error) {
			if config == nil {
				return nil, ErrConfigIsNil
			}
			return &mockProvider{config: *config}, nil
		}

		ts.registry.Register("niltest", factory)

		_, err := ts.registry.Get("niltest", nil)
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "config is nil")
	})

	ts.Run("Multiple registrations overwrite", func() {
		// First registration
		factory1 := func(config *ProviderConfig) (Provider, error) {
			return &mockProvider{config: *config, providerName: "first"}, nil
		}
		ts.registry.Register("overwrite", factory1)

		// Second registration (should overwrite)
		factory2 := func(config *ProviderConfig) (Provider, error) {
			return &mockProvider{config: *config, providerName: "second"}, nil
		}
		ts.registry.Register("overwrite", factory2)

		// Should get the second provider
		provider, err := ts.registry.Get("overwrite", &ProviderConfig{})
		ts.Require().NoError(err)
		ts.Require().Equal("second", provider.Name())
	})
}

// TestRegistryTestSuite runs the registry test suite
func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

// ErrorTestSuite tests error handling and edge cases
type ErrorTestSuite struct {
	suite.Suite
}

// TestStaticErrors tests package-level error definitions
func (ts *ErrorTestSuite) TestStaticErrors() {
	ts.Run("ErrProviderNotFound properties", func() {
		ts.Require().Error(ErrProviderNotFound)
		ts.Require().Equal("provider not found", ErrProviderNotFound.Error())
	})

	ts.Run("Error wrapping with provider name", func() {
		providerName := "test-provider"
		wrappedErr := fmt.Errorf("wrapped: %w: %s", ErrProviderNotFound, providerName)

		// Should be able to unwrap to find original error
		ts.Require().Contains(wrappedErr.Error(), ErrProviderNotFound.Error())
		ts.Require().Contains(wrappedErr.Error(), providerName)
	})
}

// TestProviderInterfaceCompliance tests that mock provider implements all interfaces
func (ts *ErrorTestSuite) TestProviderInterfaceCompliance() {
	ts.Run("Mock provider implements Provider interface", func() {
		provider := &mockProvider{}
		err := provider.Initialize(&ProviderConfig{})
		ts.Require().NoError(err)
		ts.Require().NotNil(provider)

		// Test that all service methods return non-nil
		ts.Require().NotNil(provider.Compute())
		ts.Require().NotNil(provider.Storage())
		ts.Require().NotNil(provider.Network())
		ts.Require().NotNil(provider.Container())
		ts.Require().NotNil(provider.Database())
		ts.Require().NotNil(provider.Security())
		ts.Require().NotNil(provider.Monitoring())
		ts.Require().NotNil(provider.Serverless())
		ts.Require().NotNil(provider.AI())
		ts.Require().NotNil(provider.Cost())
		ts.Require().NotNil(provider.Compliance())
		ts.Require().NotNil(provider.Disaster())
		ts.Require().NotNil(provider.Edge())
		ts.Require().NotNil(provider.Quantum())
	})

	ts.Run("Provider basic operations", func() {
		provider := &mockProvider{}

		// Test name before initialization
		ts.Require().Equal("mock", provider.Name())

		// Test initialization
		config := &ProviderConfig{
			Region: "us-west-2",
			Credentials: Credentials{
				Type:      "key",
				AccessKey: "test-key",
				SecretKey: "test-secret",
			},
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		}

		err := provider.Initialize(config)
		ts.Require().NoError(err)

		// Test validation
		err = provider.Validate()
		ts.Require().NoError(err)

		// Test health check
		health, err := provider.Health()
		ts.Require().NoError(err)
		ts.Require().NotNil(health)
		ts.Require().True(health.Healthy)
		ts.Require().Equal("healthy", health.Status)
		ts.Require().Equal(10*time.Millisecond, health.Latency)

		// Test close
		err = provider.Close()
		ts.Require().NoError(err)
	})
}

// TestServiceInterfaceEdgeCases tests edge cases for service interfaces
func (ts *ErrorTestSuite) TestServiceInterfaceEdgeCases() {
	ctx := context.Background()
	provider := &mockProvider{}
	err := provider.Initialize(&ProviderConfig{})
	ts.Require().NoError(err)

	ts.Run("ServerlessService edge cases", func() {
		serverless := provider.Serverless()

		// Test function with minimal configuration
		funcReq := &CreateFunctionRequest{
			Name:    "minimal-func",
			Runtime: "go1.x",
			Handler: "main",
			Code:    []byte("package main\nfunc main() {}"),
			Timeout: 30 * time.Second,
			Memory:  128,
		}

		function, err := serverless.CreateFunction(ctx, funcReq)
		ts.Require().NoError(err)
		ts.Require().Equal("minimal-func", function.Name)

		// Test function invocation
		payload := []byte(`{"test": "data"}`)
		result, err := serverless.InvokeFunction(ctx, function.ID, payload)
		ts.Require().NoError(err)
		ts.Require().Equal([]byte("result"), result)

		// Test workflow creation
		workflowReq := &CreateWorkflowRequest{
			Name:        "test-workflow",
			Description: "Test workflow",
			Definition:  `{"StartAt": "Hello", "States": {"Hello": {"Type": "Task"}}}`,
		}

		workflow, err := serverless.CreateWorkflow(ctx, workflowReq)
		ts.Require().NoError(err)
		ts.Require().Equal("test-workflow", workflow.Name)

		// Test workflow execution
		input := map[string]interface{}{"param": "value"}
		execution, err := serverless.ExecuteWorkflow(ctx, workflow.ID, input)
		ts.Require().NoError(err)
		ts.Require().Equal(workflow.ID, execution.WorkflowID)
	})

	ts.Run("MonitoringService comprehensive operations", func() {
		monitoring := provider.Monitoring()

		// Test metric creation and querying
		metric := &Metric{
			Name:      "test.metric",
			Namespace: "TestApp",
			Value:     42.0,
			Unit:      "Count",
			Timestamp: time.Now(),
			Dimensions: map[string]string{
				"Environment": "test",
				"Service":     "api",
			},
		}

		err := monitoring.PutMetric(ctx, metric)
		ts.Require().NoError(err)

		// Query metrics
		query := &MetricQuery{
			Namespace:  "TestApp",
			MetricName: "test.metric",
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now(),
			Period:     5 * time.Minute,
			Statistic:  "Average",
		}

		metrics, err := monitoring.GetMetrics(ctx, query)
		ts.Require().NoError(err)
		ts.Require().NotNil(metrics)

		// Test alert creation with complex conditions
		alertReq := &CreateAlertRequest{
			Name:        "high-error-rate",
			Description: "Alert when error rate exceeds threshold",
			Condition: &AlertCondition{
				MetricName:        "error.rate",
				Namespace:         "TestApp",
				Statistic:         "Average",
				Threshold:         0.05, // 5%
				Comparison:        "GreaterThanThreshold",
				Period:            5 * time.Minute,
				EvaluationPeriods: 2,
			},
			Actions: []*AlertAction{
				{
					Type:   "email",
					Target: "alerts@example.com",
					Properties: map[string]string{
						"subject": "High Error Rate Alert",
					},
				},
				{
					Type:   "webhook",
					Target: "https://webhook.example.com/alert",
					Properties: map[string]string{
						"method": "POST",
					},
				},
			},
			Enabled: true,
		}

		alert, err := monitoring.CreateAlert(ctx, alertReq)
		ts.Require().NoError(err)
		ts.Require().Equal("high-error-rate", alert.Name)
		ts.Require().True(alert.Enabled)
		ts.Require().Len(alert.Actions, 2)
	})

	ts.Run("ComplianceService operations", func() {
		compliance := provider.Compliance()

		// Test compliance check
		result, err := compliance.RunComplianceCheck(ctx, "SOC2")
		ts.Require().NoError(err)
		ts.Require().Equal("SOC2", result.Standard)

		// Test compliance status
		status, err := compliance.GetComplianceStatus(ctx)
		ts.Require().NoError(err)
		ts.Require().InEpsilon(0.95, status.Overall, 0.001)

		// Test report generation
		reportReq := &ReportRequest{
			Type:      "compliance",
			Format:    "pdf",
			StartDate: time.Now().AddDate(0, -1, 0), // Last month
			EndDate:   time.Now(),
			Standards: []string{"SOC2", "GDPR", "HIPAA"},
		}

		report, err := compliance.GenerateComplianceReport(ctx, reportReq)
		ts.Require().NoError(err)
		ts.Require().Equal("compliance", report.Type)
		ts.Require().Equal("pdf", report.Format)
	})
}

// TestErrorTestSuite runs the error test suite
func TestErrorTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorTestSuite))
}
