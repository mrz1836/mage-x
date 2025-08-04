package azure

import (
	"context"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/providers"
	"github.com/stretchr/testify/suite"
)

// AzureAdvancedServicesTestSuite tests Azure advanced service implementations
type AzureAdvancedServicesTestSuite struct {
	suite.Suite

	provider *Provider
}

// SetupTest runs before each test
func (ts *AzureAdvancedServicesTestSuite) SetupTest() {
	config := providers.ProviderConfig{
		Region: "eastus",
		Credentials: providers.Credentials{
			Type:      "key",
			AccessKey: "azure-client-id",
			SecretKey: "azure-client-secret",
			Extra: map[string]string{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"tenant_id":       "87654321-4321-4321-4321-210987654321",
			},
		},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableCache: true,
	}

	p, err := New(&config)
	ts.Require().NoError(err)
	var ok bool
	ts.provider, ok = p.(*Provider)
	ts.Require().True(ok, "failed to cast provider to *Provider")
}

// TestAzureMonitoringService tests Azure monitoring service operations
func (ts *AzureAdvancedServicesTestSuite) TestAzureMonitoringService() {
	ctx := context.Background()
	monitoring := ts.provider.Monitoring()
	ts.Require().NotNil(monitoring)

	ts.Run("Metric operations", func() {
		// PutMetric
		metric := &providers.Metric{
			Name:      "cpu.utilization",
			Namespace: "Azure/VirtualMachines",
			Value:     85.5,
			Unit:      "Percent",
			Timestamp: time.Now(),
			Dimensions: map[string]string{
				"ResourceName":  "web-vm-01",
				"ResourceGroup": "production-rg",
				"Region":        "eastus",
			},
		}
		err := monitoring.PutMetric(ctx, metric)
		ts.Require().NoError(err)

		// GetMetrics
		query := &providers.MetricQuery{
			Namespace:  "Azure/VirtualMachines",
			MetricName: "cpu.utilization",
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now(),
			Period:     5 * time.Minute,
			Statistic:  "Average",
			Dimensions: map[string]string{
				"ResourceName": "web-vm-01",
			},
		}
		metrics, err := monitoring.GetMetrics(ctx, query)
		ts.Require().NoError(err)
		ts.Require().NotNil(metrics)
	})

	ts.Run("Dashboard operations", func() {
		// CreateDashboard
		dashboardReq := &providers.CreateDashboardRequest{
			Name:        "test-dashboard",
			Description: "Test dashboard for monitoring",
			Widgets: []*providers.Widget{
				{
					ID:    "cpu-widget",
					Type:  "line-chart",
					Title: "CPU Utilization",
					Query: map[string]interface{}{
						"metric":    "cpu.utilization",
						"timespan":  "1h",
						"statistic": "Average",
					},
					Properties: map[string]interface{}{
						"color": "blue",
						"unit":  "percent",
					},
				},
				{
					ID:    "memory-widget",
					Type:  "area-chart",
					Title: "Memory Usage",
					Query: map[string]interface{}{
						"metric":    "memory.usage",
						"timespan":  "1h",
						"statistic": "Average",
					},
				},
			},
		}
		dashboard, err := monitoring.CreateDashboard(ctx, dashboardReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(dashboard)
		ts.Require().Equal("test-dashboard", dashboard.Name)
		ts.Require().Equal("dash-123", dashboard.ID)
	})

	ts.Run("Log operations", func() {
		logGroupName := "application-logs"

		// CreateLogGroup
		err := monitoring.CreateLogGroup(ctx, logGroupName)
		ts.Require().NoError(err)

		// PutLogs
		logEntries := []*providers.LogEntry{
			{
				Timestamp: time.Now(),
				Level:     "INFO",
				Message:   "Application started successfully",
				Source:    "web-server",
				Fields: map[string]interface{}{
					"service": "web-api",
					"version": "1.2.3",
					"pid":     12345,
				},
			},
			{
				Timestamp: time.Now().Add(1 * time.Second),
				Level:     "WARN",
				Message:   "High memory usage detected",
				Source:    "web-server",
				Fields: map[string]interface{}{
					"memory_usage": "85%",
					"threshold":    "80%",
				},
			},
		}
		err = monitoring.PutLogs(ctx, logGroupName, logEntries)
		ts.Require().NoError(err)

		// QueryLogs
		logQuery := &providers.LogQuery{
			LogGroup:  logGroupName,
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Filter:    "level = 'INFO'",
			Limit:     100,
		}
		logs, err := monitoring.QueryLogs(ctx, logQuery)
		ts.Require().NoError(err)
		ts.Require().NotNil(logs)
	})

	ts.Run("Alert operations", func() {
		// CreateAlert
		alertReq := &providers.CreateAlertRequest{
			Name:        "high-cpu-alert",
			Description: "Alert when CPU usage exceeds 80%",
			Condition: &providers.AlertCondition{
				MetricName:        "cpu.utilization",
				Namespace:         "Azure/VirtualMachines",
				Statistic:         "Average",
				Threshold:         80.0,
				Comparison:        "GreaterThanThreshold",
				Period:            5 * time.Minute,
				EvaluationPeriods: 2,
			},
			Actions: []*providers.AlertAction{
				{
					Type:   "email",
					Target: "devops@example.com",
					Properties: map[string]string{
						"subject": "High CPU Alert",
						"body":    "CPU usage has exceeded 80%",
					},
				},
				{
					Type:   "webhook",
					Target: "https://hooks.slack.com/services/...",
					Properties: map[string]string{
						"method":  "POST",
						"channel": "#alerts",
					},
				},
			},
			Enabled: true,
		}
		alert, err := monitoring.CreateAlert(ctx, alertReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(alert)
		ts.Require().Equal("high-cpu-alert", alert.Name)
		ts.Require().Equal("alert-123", alert.ID)

		// UpdateAlert
		enabled := false
		updateReq := &providers.UpdateAlertRequest{
			Enabled: &enabled,
			Condition: &providers.AlertCondition{
				Threshold: 85.0, // Increase threshold
			},
		}
		err = monitoring.UpdateAlert(ctx, alert.ID, updateReq)
		ts.Require().NoError(err)

		// ListAlerts
		alerts, err := monitoring.ListAlerts(ctx)
		ts.Require().NoError(err)
		ts.Require().NotNil(alerts)
	})

	ts.Run("Trace operations", func() {
		// PutTrace
		trace := &providers.Trace{
			ID:            "trace-12345",
			ServiceName:   "web-api",
			OperationName: "getUserProfile",
			StartTime:     time.Now(),
			Duration:      250 * time.Millisecond,
			Status:        "OK",
			Spans: []*providers.Span{
				{
					ID:            "span-1",
					OperationName: "database.query",
					ServiceName:   "postgres",
					StartTime:     time.Now().Add(10 * time.Millisecond),
					Duration:      80 * time.Millisecond,
					Status:        "OK",
					Tags: map[string]string{
						"db.statement": "SELECT * FROM users WHERE id = $1",
						"db.type":      "postgresql",
					},
				},
				{
					ID:            "span-2",
					OperationName: "cache.get",
					ServiceName:   "redis",
					StartTime:     time.Now().Add(100 * time.Millisecond),
					Duration:      5 * time.Millisecond,
					Status:        "OK",
					Tags: map[string]string{
						"cache.key": "user:123:profile",
						"cache.hit": "false",
					},
				},
			},
		}
		err := monitoring.PutTrace(ctx, trace)
		ts.Require().NoError(err)

		// GetTrace
		retrievedTrace, err := monitoring.GetTrace(ctx, trace.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedTrace)
		ts.Require().Equal(trace.ID, retrievedTrace.ID)

		// QueryTraces
		traceQuery := &providers.TraceQuery{
			ServiceName:   "web-api",
			OperationName: "getUserProfile",
			StartTime:     time.Now().Add(-1 * time.Hour),
			EndTime:       time.Now(),
			MinDuration:   100 * time.Millisecond,
			MaxDuration:   500 * time.Millisecond,
			Tags: map[string]string{
				"http.method": "GET",
			},
		}
		traces, err := monitoring.QueryTraces(ctx, traceQuery)
		ts.Require().NoError(err)
		ts.Require().NotNil(traces)
	})
}

// TestAzureServerlessService tests Azure serverless service operations
func (ts *AzureAdvancedServicesTestSuite) TestAzureServerlessService() {
	ctx := context.Background()
	serverless := ts.provider.Serverless()
	ts.Require().NotNil(serverless)

	ts.Run("Function operations", func() {
		// CreateFunction
		functionReq := &providers.CreateFunctionRequest{
			Name:    "test-function",
			Runtime: "dotnet6",
			Handler: "TestFunction.Run",
			Code:    []byte("using System; public static class TestFunction { public static string Run() { return \"Hello Azure!\"; } }"),
			Timeout: 30 * time.Second,
			Memory:  512,
			Environment: map[string]string{
				"ENVIRONMENT": "test",
				"LOG_LEVEL":   "info",
			},
			Tags: map[string]string{
				"project": "test-project",
				"team":    "backend",
			},
		}
		function, err := serverless.CreateFunction(ctx, functionReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(function)
		ts.Require().Equal("test-function", function.Name)
		ts.Require().Equal("func-123", function.ID)

		// UpdateFunction
		memory := 1024
		timeout := 60 * time.Second
		updateReq := &providers.UpdateFunctionRequest{
			Memory:  &memory,
			Timeout: &timeout,
			Environment: map[string]string{
				"ENVIRONMENT": "production",
				"LOG_LEVEL":   "debug",
			},
		}
		err = serverless.UpdateFunction(ctx, function.ID, updateReq)
		ts.Require().NoError(err)

		// InvokeFunction
		payload := []byte(`{"name": "Azure", "message": "Hello from test!"}`)
		result, err := serverless.InvokeFunction(ctx, function.ID, payload)
		ts.Require().NoError(err)
		ts.Require().Equal([]byte("result"), result)

		// CreateEventTrigger
		eventTrigger := &providers.EventTrigger{
			Type:   "blob",
			Source: "test-storage-account",
			Properties: map[string]interface{}{
				"event":         "blob.created",
				"containerName": "uploads",
				"blobPrefix":    "documents/",
			},
		}
		err = serverless.CreateEventTrigger(ctx, function.ID, eventTrigger)
		ts.Require().NoError(err)

		// DeleteFunction
		err = serverless.DeleteFunction(ctx, function.ID)
		ts.Require().NoError(err)
	})

	ts.Run("API Gateway operations", func() {
		// CreateAPIGateway
		apiReq := &providers.CreateAPIGatewayRequest{
			Name:        "test-api",
			Description: "Test API Gateway for Azure Functions",
			Routes: []*providers.Route{
				{
					Path:   "/api/v1/users",
					Method: "GET",
				},
				{
					Path:   "/api/v1/users",
					Method: "POST",
				},
			},
			Tags: map[string]string{
				"environment": "test",
				"team":        "backend",
			},
		}
		api, err := serverless.CreateAPIGateway(ctx, apiReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(api)
		ts.Require().Equal("test-api", api.Name)
		ts.Require().Equal("api-123", api.ID)
	})

	ts.Run("Workflow operations", func() {
		// CreateWorkflow
		workflowReq := &providers.CreateWorkflowRequest{
			Name:        "test-workflow",
			Description: "Test Azure Logic App workflow",
			Definition: `{
				"$schema": "https://schema.management.azure.com/providers/Microsoft.Logic/schemas/2016-06-01/workflowdefinition.json#",
				"actions": {
					"HTTP": {
						"type": "Http",
						"inputs": {
							"method": "GET",
							"uri": "https://api.example.com/data"
						}
					}
				},
				"triggers": {
					"manual": {
						"type": "Request",
						"kind": "Http"
					}
				}
			}`,
		}
		workflow, err := serverless.CreateWorkflow(ctx, workflowReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(workflow)
		ts.Require().Equal("test-workflow", workflow.Name)
		ts.Require().Equal("wf-123", workflow.ID)

		// ExecuteWorkflow
		input := map[string]interface{}{
			"userId": 12345,
			"action": "processData",
			"data": map[string]interface{}{
				"items":    []string{"item1", "item2", "item3"},
				"priority": "high",
			},
		}
		execution, err := serverless.ExecuteWorkflow(ctx, workflow.ID, input)
		ts.Require().NoError(err)
		ts.Require().NotNil(execution)
		ts.Require().Equal("exec-123", execution.ID)
		ts.Require().Equal(workflow.ID, execution.WorkflowID)
	})
}

// TestAzureAIService tests Azure AI service operations
func (ts *AzureAdvancedServicesTestSuite) TestAzureAIService() {
	ctx := context.Background()
	ai := ts.provider.AI()
	ts.Require().NotNil(ai)

	ts.Run("Dataset operations", func() {
		// CreateDataset
		datasetReq := &providers.CreateDatasetRequest{
			Name:   "training-images",
			Type:   "image",
			Source: "https://testaccount.blob.core.windows.net/datasets/images/",
			Format: "jpeg",
			Tags: map[string]string{
				"project": "image-classification",
				"version": "v1.0",
			},
		}
		dataset, err := ai.CreateDataset(ctx, datasetReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(dataset)
		ts.Require().Equal("training-images", dataset.Name)
		ts.Require().Equal("dataset-123", dataset.ID)

		// PreprocessData
		pipeline := &providers.Pipeline{
			Steps: []*providers.PipelineStep{
				{
					Name: "resize",
					Type: "transform",
					Config: map[string]interface{}{
						"width":  224,
						"height": 224,
					},
				},
				{
					Name: "normalize",
					Type: "transform",
					Config: map[string]interface{}{
						"mean": []float64{0.485, 0.456, 0.406},
						"std":  []float64{0.229, 0.224, 0.225},
					},
				},
			},
		}
		err = ai.PreprocessData(ctx, dataset.ID, pipeline)
		ts.Require().NoError(err)
	})

	ts.Run("Model operations", func() {
		// CreateModel
		modelReq := &providers.CreateModelRequest{
			Name:      "image-classifier",
			Type:      "classification",
			Framework: "pytorch",
			Tags: map[string]string{
				"purpose": "image-classification",
				"team":    "ml-engineering",
			},
		}
		model, err := ai.CreateModel(ctx, modelReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(model)
		ts.Require().Equal("image-classifier", model.Name)
		ts.Require().Equal("model-123", model.ID)

		// TrainModel
		dataset := &providers.Dataset{
			ID:   "dataset-123",
			Name: "training-images",
			Type: "image",
		}
		trainingJob, err := ai.TrainModel(ctx, model.ID, dataset)
		ts.Require().NoError(err)
		ts.Require().NotNil(trainingJob)
		ts.Require().Equal("job-123", trainingJob.ID)
		ts.Require().Equal(model.ID, trainingJob.ModelID)

		// DeployModel
		deployConfig := &providers.DeploymentConfig{
			InstanceType:  "Standard_NC6s_v3", // GPU instance
			InstanceCount: 2,
			AutoScale:     true,
			MaxInstances:  10,
		}
		endpoint, err := ai.DeployModel(ctx, model.ID, deployConfig)
		ts.Require().NoError(err)
		ts.Require().NotNil(endpoint)
		ts.Require().Equal("endpoint-123", endpoint.ID)
		ts.Require().Equal(model.ID, endpoint.ModelID)

		// Predict
		inputData := map[string]interface{}{
			"image_url":            "https://example.com/test-image.jpg",
			"confidence_threshold": 0.8,
		}
		prediction, err := ai.Predict(ctx, endpoint.ID, inputData)
		ts.Require().NoError(err)
		ts.Require().Equal("prediction", prediction)

		// FineTuneModel
		fineTuneDataset := &providers.Dataset{
			ID:   "dataset-456",
			Name: "fine-tune-data",
			Type: "image",
		}
		fineTunedModel, err := ai.FineTuneModel(ctx, model.ID, fineTuneDataset)
		ts.Require().NoError(err)
		ts.Require().NotNil(fineTunedModel)
		ts.Require().Contains(fineTunedModel.ID, model.ID+"-tuned")

		// ExplainPrediction
		explanation, err := ai.ExplainPrediction(ctx, endpoint.ID, inputData)
		ts.Require().NoError(err)
		ts.Require().NotNil(explanation)
		ts.Require().Equal("explained", explanation.Prediction)
	})

	ts.Run("Neural Network operations", func() {
		// CreateNeuralNetwork
		architecture := &providers.NetworkArchitecture{
			Layers: []*providers.Layer{
				{
					Type:       "conv2d",
					Units:      32,
					Activation: "relu",
					Config: map[string]interface{}{
						"kernel_size": 3,
						"padding":     "same",
					},
				},
				{
					Type: "maxpool2d",
					Config: map[string]interface{}{
						"pool_size": 2,
					},
				},
				{
					Type:       "conv2d",
					Units:      64,
					Activation: "relu",
					Config: map[string]interface{}{
						"kernel_size": 3,
						"padding":     "same",
					},
				},
				{
					Type: "flatten",
				},
				{
					Type:       "dense",
					Units:      128,
					Activation: "relu",
					Config: map[string]interface{}{
						"dropout": 0.5,
					},
				},
				{
					Type:       "dense",
					Units:      10,
					Activation: "softmax",
				},
			},
			Optimizer:    "adam",
			LossFunction: "categorical_crossentropy",
			Metrics:      []string{"accuracy", "precision", "recall"},
		}

		neuralNetwork, err := ai.CreateNeuralNetwork(ctx, architecture)
		ts.Require().NoError(err)
		ts.Require().NotNil(neuralNetwork)
		ts.Require().Equal("nn-123", neuralNetwork.ID)
	})
}

// TestAzureAdvancedServicesTestSuite runs the Azure advanced services test suite
func TestAzureAdvancedServicesTestSuite(t *testing.T) {
	suite.Run(t, new(AzureAdvancedServicesTestSuite))
}
