package azure

import (
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/providers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AzureProviderTestSuite defines the test suite for Azure provider
type AzureProviderTestSuite struct {
	suite.Suite
	provider providers.Provider
	config   providers.ProviderConfig
}

// SetupTest runs before each test
func (ts *AzureProviderTestSuite) SetupTest() {
	ts.config = providers.ProviderConfig{
		Region: "eastus",
		Credentials: providers.Credentials{
			Type:      "key",
			AccessKey: "client-id-example",
			SecretKey: "client-secret-example",
			Extra: map[string]string{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"tenant_id":       "87654321-4321-4321-4321-210987654321",
			},
		},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableCache: true,
	}

	var err error
	ts.provider, err = New(ts.config)
	require.NoError(ts.T(), err)
}

// TearDownTest runs after each test
func (ts *AzureProviderTestSuite) TearDownTest() {
	if ts.provider != nil {
		if err := ts.provider.Close(); err != nil {
			ts.T().Logf("Warning: failed to close Azure provider in teardown: %v", err)
		}
	}
}

// TestAzureProviderBasics tests basic Azure provider functionality
func (ts *AzureProviderTestSuite) TestAzureProviderBasics() {
	ts.Run("Provider name", func() {
		require.Equal(ts.T(), "azure", ts.provider.Name())
	})

	ts.Run("Provider initialization", func() {
		azureProvider, ok := ts.provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "12345678-1234-1234-1234-123456789012", azureProvider.subscription)
		require.NotNil(ts.T(), azureProvider.services)
		require.NotNil(ts.T(), azureProvider.services.compute)
		require.NotNil(ts.T(), azureProvider.services.storage)
		require.NotNil(ts.T(), azureProvider.services.network)
	})

	ts.Run("Provider validation with valid key credentials", func() {
		validConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "client-id",
				SecretKey: "client-secret",
				Extra: map[string]string{
					"subscription_id": "sub-123",
				},
			},
		}

		provider, err := New(validConfig)
		require.NoError(ts.T(), err)

		err = provider.Validate()
		require.NoError(ts.T(), err)
	})

	ts.Run("Provider validation with valid certificate credentials", func() {
		certConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:     "cert",
				CertPath: "/path/to/cert.pem",
				KeyPath:  "/path/to/key.pem",
				Extra: map[string]string{
					"subscription_id": "sub-123",
				},
			},
		}

		provider, err := New(certConfig)
		require.NoError(ts.T(), err)

		err = provider.Validate()
		require.NoError(ts.T(), err)
	})

	ts.Run("Provider validation with missing subscription ID", func() {
		invalidConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "client-id",
				SecretKey: "client-secret",
				// Missing subscription_id in Extra
			},
		}

		_, err := New(invalidConfig)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure subscription ID is required")
	})

	ts.Run("Provider validation with missing key credentials", func() {
		invalidConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type: "key",
				// Missing AccessKey and SecretKey
				Extra: map[string]string{
					"subscription_id": "sub-123",
				},
			},
		}

		provider, err := New(invalidConfig)
		require.NoError(ts.T(), err) // Creation should succeed

		err = provider.Validate()
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure client ID and client secret are required")
	})

	ts.Run("Provider validation with missing certificate paths", func() {
		invalidConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type: "cert",
				// Missing CertPath and KeyPath
				Extra: map[string]string{
					"subscription_id": "sub-123",
				},
			},
		}

		provider, err := New(invalidConfig)
		require.NoError(ts.T(), err) // Creation should succeed

		err = provider.Validate()
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure certificate and key paths are required")
	})

	ts.Run("Provider health check", func() {
		health, err := ts.provider.Health()
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), health)
		require.True(ts.T(), health.Healthy)
		require.Equal(ts.T(), "healthy", health.Status)
		require.NotEmpty(ts.T(), health.Services)

		// Check specific service health
		require.Contains(ts.T(), health.Services, "compute")
		require.Contains(ts.T(), health.Services, "storage")
		require.Contains(ts.T(), health.Services, "network")

		computeHealth := health.Services["compute"]
		require.True(ts.T(), computeHealth.Available)
		require.Greater(ts.T(), computeHealth.ResponseTime, time.Duration(0))
	})

	ts.Run("Provider close", func() {
		err := ts.provider.Close()
		require.NoError(ts.T(), err)
	})
}

// TestAzureServiceAccessors tests service accessor methods
func (ts *AzureProviderTestSuite) TestAzureServiceAccessors() {
	ts.Run("Core service accessors return non-nil services", func() {
		// Test core implemented services
		require.NotNil(ts.T(), ts.provider.Compute())
		require.NotNil(ts.T(), ts.provider.Storage())
		require.NotNil(ts.T(), ts.provider.Network())
		require.NotNil(ts.T(), ts.provider.Container())
		require.NotNil(ts.T(), ts.provider.Database())
		require.NotNil(ts.T(), ts.provider.Security())
		require.NotNil(ts.T(), ts.provider.Monitoring())
		require.NotNil(ts.T(), ts.provider.Serverless())
		require.NotNil(ts.T(), ts.provider.AI())
	})

	ts.Run("Placeholder services return nil", func() {
		// Test placeholder services that are not yet implemented
		require.Nil(ts.T(), ts.provider.Cost())
		require.Nil(ts.T(), ts.provider.Compliance())
		require.Nil(ts.T(), ts.provider.Disaster())
		require.Nil(ts.T(), ts.provider.Edge())
		require.Nil(ts.T(), ts.provider.Quantum())
	})

	ts.Run("Service types are correct", func() {
		// Verify we get the correct Azure-specific service implementations
		azureProvider, ok := ts.provider.(*Provider)
		require.True(ts.T(), ok)

		_, ok = azureProvider.services.compute.(*azureComputeService)
		require.True(ts.T(), ok, "Compute service should be Azure-specific implementation")

		_, ok = azureProvider.services.storage.(*azureStorageService)
		require.True(ts.T(), ok, "Storage service should be Azure-specific implementation")

		_, ok = azureProvider.services.network.(*azureNetworkService)
		require.True(ts.T(), ok, "Network service should be Azure-specific implementation")

		_, ok = azureProvider.services.container.(*azureContainerService)
		require.True(ts.T(), ok, "Container service should be Azure-specific implementation")

		_, ok = azureProvider.services.database.(*azureDatabaseService)
		require.True(ts.T(), ok, "Database service should be Azure-specific implementation")

		_, ok = azureProvider.services.security.(*azureSecurityService)
		require.True(ts.T(), ok, "Security service should be Azure-specific implementation")

		_, ok = azureProvider.services.monitoring.(*azureMonitoringService)
		require.True(ts.T(), ok, "Monitoring service should be Azure-specific implementation")

		_, ok = azureProvider.services.serverless.(*azureServerlessService)
		require.True(ts.T(), ok, "Serverless service should be Azure-specific implementation")

		_, ok = azureProvider.services.ai.(*azureAIService)
		require.True(ts.T(), ok, "AI service should be Azure-specific implementation")
	})
}

// TestAzureCredentialTypes tests different Azure credential types
func (ts *AzureProviderTestSuite) TestAzureCredentialTypes() {
	ts.Run("Service Principal (key) authentication", func() {
		config := providers.ProviderConfig{
			Region: "westus2",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "service-principal-client-id",
				SecretKey: "service-principal-client-secret",
				Extra: map[string]string{
					"subscription_id": "12345678-1234-1234-1234-123456789012",
					"tenant_id":       "87654321-4321-4321-4321-210987654321",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "azure", provider.Name())

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "12345678-1234-1234-1234-123456789012", azureProvider.subscription)
		require.Equal(ts.T(), config, azureProvider.config)
	})

	ts.Run("Certificate-based authentication", func() {
		config := providers.ProviderConfig{
			Region: "northeurope",
			Credentials: providers.Credentials{
				Type:     "cert",
				CertPath: "/etc/ssl/certs/azure-cert.pem",
				KeyPath:  "/etc/ssl/private/azure-key.pem",
				Extra: map[string]string{
					"subscription_id": "98765432-4321-4321-4321-210987654321",
					"tenant_id":       "12345678-1234-1234-1234-123456789012",
					"client_id":       "cert-based-client-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "azure", provider.Name())

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "98765432-4321-4321-4321-210987654321", azureProvider.subscription)
	})

	ts.Run("OAuth token authentication", func() {
		config := providers.ProviderConfig{
			Region: "southeastasia",
			Credentials: providers.Credentials{
				Type:  "oauth",
				Token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIs...",
				Extra: map[string]string{
					"subscription_id": "11111111-2222-3333-4444-555555555555",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "azure", provider.Name())

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "11111111-2222-3333-4444-555555555555", azureProvider.subscription)
	})
}

// TestAzureRegions tests Azure region handling
func (ts *AzureProviderTestSuite) TestAzureRegions() {
	regions := []string{
		"eastus",
		"eastus2",
		"westus",
		"westus2",
		"westus3",
		"centralus",
		"northcentralus",
		"southcentralus",
		"westcentralus",
		"canadacentral",
		"canadaeast",
		"brazilsouth",
		"northeurope",
		"westeurope",
		"ukwest",
		"uksouth",
		"francecentral",
		"francesouth",
		"germanywestcentral",
		"switzerlandnorth",
		"norwayeast",
		"japaneast",
		"japanwest",
		"eastasia",
		"southeastasia",
		"australiaeast",
		"australiasoutheast",
		"centralindia",
		"southindia",
		"westindia",
		"koreacentral",
		"koreasouth",
		"uaenorth",
		"southafricanorth",
	}

	for _, region := range regions {
		ts.Run("Region "+region, func() {
			config := providers.ProviderConfig{
				Region: region,
				Credentials: providers.Credentials{
					Type:      "key",
					AccessKey: "test-client-id",
					SecretKey: "test-client-secret",
					Extra: map[string]string{
						"subscription_id": "test-subscription-id",
					},
				},
			}

			provider, err := New(config)
			require.NoError(ts.T(), err)
			require.Equal(ts.T(), "azure", provider.Name())

			azureProvider, ok := provider.(*Provider)
			require.True(ts.T(), ok)
			require.Equal(ts.T(), region, azureProvider.config.Region)
		})
	}
}

// TestAzureProviderConfiguration tests various Azure configuration options
func (ts *AzureProviderTestSuite) TestAzureProviderConfiguration() {
	ts.Run("Custom endpoint configuration", func() {
		config := providers.ProviderConfig{
			Region:   "eastus",
			Endpoint: "https://management.usgovcloudapi.net", // Azure Government Cloud
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "gov-client-id",
				SecretKey: "gov-client-secret",
				Extra: map[string]string{
					"subscription_id": "gov-subscription-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "https://management.usgovcloudapi.net", azureProvider.config.Endpoint)
	})

	ts.Run("Custom timeout and retry configuration", func() {
		config := providers.ProviderConfig{
			Region:     "westeurope",
			Timeout:    60 * time.Second,
			MaxRetries: 5,
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "eu-client-id",
				SecretKey: "eu-client-secret",
				Extra: map[string]string{
					"subscription_id": "eu-subscription-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), 60*time.Second, azureProvider.config.Timeout)
		require.Equal(ts.T(), 5, azureProvider.config.MaxRetries)
	})

	ts.Run("Custom headers and proxy configuration", func() {
		config := providers.ProviderConfig{
			Region: "southeastasia",
			CustomHeaders: map[string]string{
				"User-Agent":      "MyApp/1.0 AzureGoSDK",
				"X-Custom-Header": "custom-value",
			},
			ProxyURL: "http://corporate-proxy.company.com:8080",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "asia-client-id",
				SecretKey: "asia-client-secret",
				Extra: map[string]string{
					"subscription_id": "asia-subscription-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "MyApp/1.0 AzureGoSDK", azureProvider.config.CustomHeaders["User-Agent"])
		require.Equal(ts.T(), "custom-value", azureProvider.config.CustomHeaders["X-Custom-Header"])
		require.Equal(ts.T(), "http://corporate-proxy.company.com:8080", azureProvider.config.ProxyURL)
	})

	ts.Run("TLS configuration", func() {
		config := providers.ProviderConfig{
			Region: "australiaeast",
			TLSConfig: &providers.TLSConfig{
				InsecureSkipVerify: false,
				CAPath:             "/etc/ssl/certs/azure-ca.pem",
				CertPath:           "/etc/ssl/certs/azure-client.pem",
				KeyPath:            "/etc/ssl/private/azure-client-key.pem",
				MinVersion:         "1.3",
			},
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "au-client-id",
				SecretKey: "au-client-secret",
				Extra: map[string]string{
					"subscription_id": "au-subscription-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err)

		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.NotNil(ts.T(), azureProvider.config.TLSConfig)
		require.False(ts.T(), azureProvider.config.TLSConfig.InsecureSkipVerify)
		require.Equal(ts.T(), "/etc/ssl/certs/azure-ca.pem", azureProvider.config.TLSConfig.CAPath)
		require.Equal(ts.T(), "1.3", azureProvider.config.TLSConfig.MinVersion)
	})
}

// TestAzureProviderErrors tests error handling in Azure provider
func (ts *AzureProviderTestSuite) TestAzureProviderErrors() {
	ts.Run("Invalid credential type", func() {
		config := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type: "invalid-type",
				Extra: map[string]string{
					"subscription_id": "test-subscription-id",
				},
			},
		}

		provider, err := New(config)
		require.NoError(ts.T(), err) // Creation should succeed

		err = provider.Validate()
		require.NoError(ts.T(), err) // Validation should pass for unknown types (handled gracefully)
	})

	ts.Run("Empty subscription ID", func() {
		config := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "test-client-id",
				SecretKey: "test-client-secret",
				Extra: map[string]string{
					"subscription_id": "",
				},
			},
		}

		_, err := New(config)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure subscription ID is required")
	})

	ts.Run("Missing subscription ID from Extra", func() {
		config := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "test-client-id",
				SecretKey: "test-client-secret",
				Extra:     map[string]string{}, // No subscription_id
			},
		}

		_, err := New(config)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure subscription ID is required")
	})

	ts.Run("Nil Extra map", func() {
		config := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "test-client-id",
				SecretKey: "test-client-secret",
				Extra:     nil, // Nil Extra map
			},
		}

		_, err := New(config)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "Azure subscription ID is required")
	})
}

// TestAzureProviderRegistration tests Azure provider registration
func (ts *AzureProviderTestSuite) TestAzureProviderRegistration() {
	ts.Run("Azure provider is registered", func() {
		// The init() function should have registered the Azure provider
		providerNames := providers.List()
		require.Contains(ts.T(), providerNames, "azure")
	})

	ts.Run("Get Azure provider from registry", func() {
		config := providers.ProviderConfig{
			Region: "westus2",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "registry-test-client-id",
				SecretKey: "registry-test-client-secret",
				Extra: map[string]string{
					"subscription_id": "registry-test-subscription-id",
				},
			},
		}

		provider, err := providers.Get("azure", config)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), provider)
		require.Equal(ts.T(), "azure", provider.Name())

		// Verify it's the correct type
		azureProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "registry-test-subscription-id", azureProvider.subscription)
		require.Equal(ts.T(), "westus2", azureProvider.config.Region)
	})
}

// TestAzureProviderComparison tests comparison with other providers
func (ts *AzureProviderTestSuite) TestAzureProviderComparison() {
	ts.Run("Azure provider has unique characteristics", func() {
		// Test Azure provider characteristics
		azureConfig := providers.ProviderConfig{
			Region: "eastus",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "azure-client-id",
				SecretKey: "azure-client-secret",
				Extra: map[string]string{
					"subscription_id": "azure-subscription-id",
				},
			},
		}

		azureProvider, err := providers.Get("azure", azureConfig)
		require.NoError(ts.T(), err)

		// Test Azure provider unique characteristics
		require.Equal(ts.T(), "azure", azureProvider.Name())

		// Test Azure health status contains Azure-specific services
		azureHealth, err := azureProvider.Health()
		require.NoError(ts.T(), err)
		require.True(ts.T(), azureHealth.Healthy)

		// Azure uses service names like compute, storage, network, aks, sql, keyvault, etc.
		require.Contains(ts.T(), azureHealth.Services, "compute")
		require.Contains(ts.T(), azureHealth.Services, "storage")
		require.Contains(ts.T(), azureHealth.Services, "network")
		require.Contains(ts.T(), azureHealth.Services, "aks")
		require.Contains(ts.T(), azureHealth.Services, "sql")
		require.Contains(ts.T(), azureHealth.Services, "keyvault")
		require.Contains(ts.T(), azureHealth.Services, "monitor")
		require.Contains(ts.T(), azureHealth.Services, "functions")
		require.Contains(ts.T(), azureHealth.Services, "cognitive")
	})
}

// TestAzureProviderTestSuite runs the test suite
func TestAzureProviderTestSuite(t *testing.T) {
	suite.Run(t, new(AzureProviderTestSuite))
}
