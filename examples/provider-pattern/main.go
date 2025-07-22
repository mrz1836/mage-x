// Package main demonstrates using the provider pattern
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrz1836/go-mage/pkg/providers"
	_ "github.com/mrz1836/go-mage/pkg/providers/aws"   // Register AWS provider
	_ "github.com/mrz1836/go-mage/pkg/providers/azure" // Register Azure provider
)

func main() {
	// Example 1: Using AWS provider
	fmt.Println("=== AWS Provider Example ===")
	if err := useAWSProvider(); err != nil {
		log.Printf("AWS provider error: %v", err)
	}

	fmt.Println("\n=== Azure Provider Example ===")
	if err := useAzureProvider(); err != nil {
		log.Printf("Azure provider error: %v", err)
	}

	fmt.Println("\n=== Multi-Cloud Example ===")
	if err := multiCloudExample(); err != nil {
		log.Printf("Multi-cloud error: %v", err)
	}
}

func useAWSProvider() error {
	// Configure AWS provider
	config := providers.ProviderConfig{
		Credentials: providers.Credentials{
			Type:      "key",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
			SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		Region:     "us-east-1",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	// Get AWS provider instance
	provider, err := providers.Get("aws", config)
	if err != nil {
		return fmt.Errorf("failed to get AWS provider: %w", err)
	}
	defer provider.Close()

	// Check provider health
	health, err := provider.Health()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	fmt.Printf("Provider health: %s\n", health.Status)

	// Use compute service
	ctx := context.Background()
	compute := provider.Compute()

	// Create an instance
	instance, err := compute.CreateInstance(ctx, &providers.CreateInstanceRequest{
		Name:   "demo-instance",
		Type:   "t3.micro",
		Image:  "ami-0c02fb55956c7d316", // Amazon Linux 2
		Region: "us-east-1",
		Zone:   "us-east-1a",
		Tags: map[string]string{
			"Environment": "demo",
			"CreatedBy":   "provider-example",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	fmt.Printf("Created instance: %s (ID: %s)\n", instance.Name, instance.ID)

	// Use storage service
	storage := provider.Storage()

	// Create a bucket
	bucket, err := storage.CreateBucket(ctx, &providers.CreateBucketRequest{
		Name:       "my-demo-bucket-" + fmt.Sprintf("%d", time.Now().Unix()),
		Region:     "us-east-1",
		Versioning: true,
		Encryption: true,
		Tags: map[string]string{
			"Purpose": "demo",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	fmt.Printf("Created bucket: %s\n", bucket.Name)

	// Use container service
	container := provider.Container()

	// Create a Kubernetes cluster
	cluster, err := container.CreateCluster(ctx, &providers.CreateClusterRequest{
		Name:      "demo-eks-cluster",
		Type:      "kubernetes",
		Version:   "1.27",
		Region:    "us-east-1",
		NodeCount: 3,
		NodeType:  "t3.medium",
	})
	if err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	fmt.Printf("Created cluster: %s (Type: %s)\n", cluster.Name, cluster.Type)

	// Use AI service
	ai := provider.AI()

	// Create a dataset
	dataset, err := ai.CreateDataset(ctx, &providers.CreateDatasetRequest{
		Name:   "demo-dataset",
		Type:   "image",
		Source: "s3://my-bucket/data",
		Format: "jpeg",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	fmt.Printf("Created dataset: %s\n", dataset.Name)

	// Use monitoring service
	monitoring := provider.Monitoring()

	// Put a metric
	err = monitoring.PutMetric(ctx, &providers.Metric{
		Name:      "DemoMetric",
		Namespace: "Demo/Application",
		Value:     42.0,
		Unit:      "Count",
		Timestamp: time.Now(),
		Dimensions: map[string]string{
			"Environment": "demo",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to put metric: %w", err)
	}

	fmt.Println("Successfully put metric")

	return nil
}

func useAzureProvider() error {
	// Configure Azure provider
	config := providers.ProviderConfig{
		Credentials: providers.Credentials{
			Type:      "key",
			AccessKey: "client-id",
			SecretKey: "client-secret",
			Extra: map[string]string{
				"subscription_id": "12345678-1234-1234-1234-123456789012",
				"tenant_id":       "87654321-4321-4321-4321-210987654321",
			},
		},
		Region:  "eastus",
		Timeout: 30 * time.Second,
	}

	// Get Azure provider instance
	provider, err := providers.Get("azure", config)
	if err != nil {
		return fmt.Errorf("failed to get Azure provider: %w", err)
	}
	defer provider.Close()

	ctx := context.Background()

	// Create a virtual network
	network := provider.Network()
	vpc, err := network.CreateVPC(ctx, &providers.CreateVPCRequest{
		Name:   "demo-vnet",
		CIDR:   "10.0.0.0/16",
		Region: "eastus",
		Tags: map[string]string{
			"Environment": "demo",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create VNet: %w", err)
	}

	fmt.Printf("Created VNet: %s (CIDR: %s)\n", vpc.Name, vpc.CIDR)

	// Create a database
	database := provider.Database()
	db, err := database.CreateDatabase(ctx, &providers.CreateDatabaseRequest{
		Name:     "demo-db",
		Engine:   "azuresql",
		Version:  "12.0",
		Size:     "S0",
		Storage:  10,
		Username: "adminuser",
		Password: "ComplexPassword123!",
	})
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	fmt.Printf("Created database: %s (Engine: %s)\n", db.Name, db.Engine)

	return nil
}

func multiCloudExample() error {
	ctx := context.Background()

	// Define provider configurations
	providersConfig := map[string]providers.ProviderConfig{
		"aws": {
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "aws-access-key",
				SecretKey: "aws-secret-key",
			},
			Region: "us-east-1",
		},
		"azure": {
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "azure-client-id",
				SecretKey: "azure-client-secret",
				Extra: map[string]string{
					"subscription_id": "sub-id",
					"tenant_id":       "tenant-id",
				},
			},
			Region: "eastus",
		},
	}

	// Deploy same application to multiple clouds
	for providerName, config := range providersConfig {
		fmt.Printf("\nDeploying to %s...\n", providerName)

		provider, err := providers.Get(providerName, config)
		if err != nil {
			log.Printf("Failed to get %s provider: %v", providerName, err)
			continue
		}
		defer provider.Close()

		// Create compute instance
		compute := provider.Compute()
		instance, err := compute.CreateInstance(ctx, &providers.CreateInstanceRequest{
			Name:   fmt.Sprintf("multi-cloud-app-%s", providerName),
			Type:   getInstanceType(providerName),
			Region: config.Region,
			Tags: map[string]string{
				"Environment": "multi-cloud",
				"Provider":    providerName,
			},
		})
		if err != nil {
			log.Printf("Failed to create instance on %s: %v", providerName, err)
			continue
		}

		fmt.Printf("✓ Created instance on %s: %s\n", providerName, instance.ID)

		// Deploy container
		container := provider.Container()
		deployment, err := container.DeployContainer(ctx, "default-cluster", &providers.DeployRequest{
			Name:     "multi-cloud-app",
			Image:    "nginx:latest",
			Replicas: 3,
			Ports:    []int{80},
			Environment: map[string]string{
				"CLOUD_PROVIDER": providerName,
			},
		})
		if err != nil {
			log.Printf("Failed to deploy container on %s: %v", providerName, err)
			continue
		}

		fmt.Printf("✓ Deployed container on %s: %s\n", providerName, deployment.ID)

		// Set up monitoring
		monitoring := provider.Monitoring()
		alert, err := monitoring.CreateAlert(ctx, &providers.CreateAlertRequest{
			Name:        "high-cpu-alert",
			Description: "Alert when CPU usage is high",
			Condition: &providers.AlertCondition{
				MetricName: "CPUUtilization",
				Namespace:  getMetricNamespace(providerName),
				Statistic:  "Average",
				Threshold:  80.0,
				Comparison: "GreaterThanThreshold",
				Period:     5 * time.Minute,
			},
			Actions: []*providers.AlertAction{
				{
					Type:   "email",
					Target: "alerts@example.com",
				},
			},
			Enabled: true,
		})
		if err != nil {
			log.Printf("Failed to create alert on %s: %v", providerName, err)
			continue
		}

		fmt.Printf("✓ Created monitoring alert on %s: %s\n", providerName, alert.ID)
	}

	fmt.Println("\nMulti-cloud deployment completed!")
	return nil
}

// Helper functions

func getInstanceType(provider string) string {
	switch provider {
	case "aws":
		return "t3.medium"
	case "azure":
		return "Standard_B2s"
	case "gcp":
		return "n1-standard-1"
	default:
		return "medium"
	}
}

func getMetricNamespace(provider string) string {
	switch provider {
	case "aws":
		return "AWS/EC2"
	case "azure":
		return "Microsoft.Compute/virtualMachines"
	case "gcp":
		return "compute.googleapis.com/instance"
	default:
		return "compute"
	}
}
