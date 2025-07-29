package providers

import (
	"testing"
)

func TestProviderCreation(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{"Local provider", "local", false},
		{"AWS provider", "aws", false},
		{"Azure provider", "azure", false},
		{"GCP provider", "gcp", false},
		{"Docker provider", "docker", false},
		{"Kubernetes provider", "kubernetes", false},
		{"Unknown provider", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.provider)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProvider() expected error for %s", tt.provider)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if provider == nil {
				t.Error("Provider should not be nil")
			}

			// Test that all required components are available
			testProviderComponents(t, provider)
		})
	}
}

func testProviderComponents(t *testing.T, provider Provider) {
	// Test Config component
	configLoader := provider.Config()
	if configLoader == nil {
		t.Error("Config component should not be nil")
	}

	// Test Environment component
	environment := provider.Environment()
	if environment == nil {
		t.Error("Environment component should not be nil")
	}

	// Test FileOperator component
	fileOps := provider.FileOperator()
	if fileOps == nil {
		t.Error("FileOperator component should not be nil")
	}

	// Test PathBuilder component
	pathBuilder := provider.PathBuilder()
	if pathBuilder == nil {
		t.Error("PathBuilder component should not be nil")
	}
}

func TestLocalProvider(t *testing.T) {
	provider, err := NewProvider("local")
	if err != nil {
		t.Fatalf("Failed to create local provider: %v", err)
	}

	// Test local provider specific behavior
	configLoader := provider.Config()
	config, err := configLoader.Load()
	if err != nil {
		t.Errorf("Local config load failed: %v", err)
	}

	if config == nil {
		t.Error("Config should not be nil")
	}

	// Test environment operations
	env := provider.Environment()
	testKey := "TEST_KEY"
	testValue := "test_value"

	err = env.Set(testKey, testValue)
	if err != nil {
		t.Errorf("Failed to set environment variable: %v", err)
	}

	retrieved := env.Get(testKey)
	if retrieved != testValue {
		t.Errorf("Expected %s, got %s", testValue, retrieved)
	}
}

func TestAWSProvider(t *testing.T) {
	provider, err := NewProvider("aws")
	if err != nil {
		t.Fatalf("Failed to create AWS provider: %v", err)
	}

	// Test AWS-specific components
	configLoader := provider.Config()
	if configLoader == nil {
		t.Error("AWS config loader should not be nil")
	}

	// AWS provider should have cloud-specific capabilities
	env := provider.Environment()
	if env == nil {
		t.Error("AWS environment should not be nil")
	}

	// Test that we can access cloud metadata (mock)
	region := env.GetWithDefault("AWS_REGION", "us-east-1")
	if region == "" {
		t.Error("AWS region should have a default value")
	}
}

func TestAzureProvider(t *testing.T) {
	provider, err := NewProvider("azure")
	if err != nil {
		t.Fatalf("Failed to create Azure provider: %v", err)
	}

	// Test Azure-specific components
	configLoader := provider.Config()
	if configLoader == nil {
		t.Error("Azure config loader should not be nil")
	}

	env := provider.Environment()
	if env == nil {
		t.Error("Azure environment should not be nil")
	}

	// Test Azure-specific environment variables
	subscriptionId := env.GetWithDefault("AZURE_SUBSCRIPTION_ID", "")
	if subscriptionId != "" {
		t.Logf("Azure subscription ID found: %s", subscriptionId)
	}
}

func TestGCPProvider(t *testing.T) {
	provider, err := NewProvider("gcp")
	if err != nil {
		t.Fatalf("Failed to create GCP provider: %v", err)
	}

	// Test GCP-specific components
	configLoader := provider.Config()
	if configLoader == nil {
		t.Error("GCP config loader should not be nil")
	}

	env := provider.Environment()
	if env == nil {
		t.Error("GCP environment should not be nil")
	}

	// Test GCP-specific environment variables
	projectId := env.GetWithDefault("GOOGLE_CLOUD_PROJECT", "default-project")
	if projectId == "" {
		t.Error("GCP project should have a default value")
	}
}

func TestDockerProvider(t *testing.T) {
	provider, err := NewProvider("docker")
	if err != nil {
		t.Fatalf("Failed to create Docker provider: %v", err)
	}

	// Test Docker-specific components
	fileOps := provider.FileOperator()
	if fileOps == nil {
		t.Error("Docker file operator should not be nil")
	}

	// Test Docker-specific path handling
	pathBuilder := provider.PathBuilder()
	if pathBuilder == nil {
		t.Error("Docker path builder should not be nil")
	}

	// Docker paths should be Unix-style
	dockerPath := pathBuilder.Join("app", "config", "app.yaml")
	expected := "app/config/app.yaml"
	if dockerPath.String() != expected {
		t.Errorf("Expected %s, got %s", expected, dockerPath.String())
	}
}

func TestKubernetesProvider(t *testing.T) {
	provider, err := NewProvider("kubernetes")
	if err != nil {
		t.Fatalf("Failed to create Kubernetes provider: %v", err)
	}

	// Test Kubernetes-specific components
	env := provider.Environment()
	if env == nil {
		t.Error("Kubernetes environment should not be nil")
	}

	// Test Kubernetes-specific environment variables
	namespace := env.GetWithDefault("POD_NAMESPACE", "default")
	if namespace == "" {
		t.Error("Kubernetes namespace should have a default value")
	}

	serviceName := env.GetWithDefault("POD_NAME", "unknown")
	if serviceName == "" {
		t.Error("Kubernetes pod name should have a default value")
	}
}

func TestProviderSwitching(t *testing.T) {
	// Test that we can switch between providers
	localProvider, err := NewProvider("local")
	if err != nil {
		t.Fatalf("Failed to create local provider: %v", err)
	}

	awsProvider, err := NewProvider("aws")
	if err != nil {
		t.Fatalf("Failed to create AWS provider: %v", err)
	}

	// Providers should be different instances
	if localProvider == awsProvider {
		t.Error("Local and AWS providers should be different instances")
	}

	// But should provide the same interface
	testProviderComponents(t, localProvider)
	testProviderComponents(t, awsProvider)
}

func TestProviderConfiguration(t *testing.T) {
	provider, err := NewProvider("local")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	configLoader := provider.Config()

	// Test configuration loading
	config, err := configLoader.Load()
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	// Test default configuration
	defaults := configLoader.GetDefaults()
	if defaults == nil {
		t.Error("Default config should not be nil")
	}

	// Test configuration validation
	err = configLoader.Validate(config)
	if err != nil {
		t.Errorf("Config validation failed: %v", err)
	}
}

func TestProviderFileOperations(t *testing.T) {
	provider, err := NewProvider("local")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	fileOps := provider.FileOperator()

	// Test basic file operations
	tempFile := "/tmp/test_provider_file.txt"
	testData := []byte("test data")

	// Write file
	err = fileOps.WriteFile(tempFile, testData, 0o644)
	if err != nil {
		t.Errorf("Failed to write file: %v", err)
	}

	// Check if file exists
	if !fileOps.Exists(tempFile) {
		t.Error("File should exist after writing")
	}

	// Read file
	readData, err := fileOps.ReadFile(tempFile)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("Expected %s, got %s", string(testData), string(readData))
	}

	// Clean up
	err = fileOps.Remove(tempFile)
	if err != nil {
		t.Errorf("Failed to remove file: %v", err)
	}
}

func TestProviderPathOperations(t *testing.T) {
	provider, err := NewProvider("local")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	pathBuilder := provider.PathBuilder()

	// Test path operations
	joined := pathBuilder.Join("home", "user", "documents", "file.txt")
	if joined.String() == "" {
		t.Error("Joined path should not be empty")
	}

	base := joined.Base()
	if base != "file.txt" {
		t.Errorf("Expected file.txt, got %s", base)
	}

	dir := joined.Dir()
	if dir.String() == "" {
		t.Error("Directory should not be empty")
	}

	ext := joined.Ext()
	if ext != ".txt" {
		t.Errorf("Expected .txt, got %s", ext)
	}
}

// Benchmark provider creation
func BenchmarkProviderCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		provider, err := NewProvider("local")
		if err != nil {
			b.Fatalf("Failed to create provider: %v", err)
		}
		_ = provider
	}
}

// Benchmark provider component access
func BenchmarkProviderComponentAccess(b *testing.B) {
	provider, err := NewProvider("local")
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = provider.Config()
		_ = provider.Environment()
		_ = provider.FileOperator()
		_ = provider.PathBuilder()
	}
}
