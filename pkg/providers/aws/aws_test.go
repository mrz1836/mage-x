package aws

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/providers"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AWSProviderTestSuite defines the test suite for AWS provider
type AWSProviderTestSuite struct {
	suite.Suite
	provider providers.Provider
	config   providers.ProviderConfig
}

// SetupTest runs before each test
func (ts *AWSProviderTestSuite) SetupTest() {
	ts.config = providers.ProviderConfig{
		Region: "us-east-1",
		Credentials: providers.Credentials{
			Type:      "key",
			AccessKey: "AKIAIOSFODNN7EXAMPLE",
			SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
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
func (ts *AWSProviderTestSuite) TearDownTest() {
	if ts.provider != nil {
		if err := ts.provider.Close(); err != nil {
			ts.T().Logf("Warning: failed to close AWS provider in teardown: %v", err)
		}
	}
}

// TestAWSProviderBasics tests basic AWS provider functionality
func (ts *AWSProviderTestSuite) TestAWSProviderBasics() {
	ts.Run("Provider name", func() {
		require.Equal(ts.T(), "aws", ts.provider.Name())
	})

	ts.Run("Provider initialization", func() {
		awsProvider, ok := ts.provider.(*Provider)
		require.True(ts.T(), ok, "Provider should be of type *Provider")
		require.Equal(ts.T(), "us-east-1", awsProvider.region)
		require.NotNil(ts.T(), awsProvider.services)
		require.NotNil(ts.T(), awsProvider.services.compute)
		require.NotNil(ts.T(), awsProvider.services.storage)
		require.NotNil(ts.T(), awsProvider.services.network)
	})

	ts.Run("Provider validation with valid credentials", func() {
		err := ts.provider.Validate()
		require.NoError(ts.T(), err)
	})

	ts.Run("Provider validation with missing credentials", func() {
		invalidConfig := providers.ProviderConfig{
			Region: "us-east-1",
			Credentials: providers.Credentials{
				Type: "key",
				// Missing AccessKey and SecretKey
			},
		}

		provider, err := New(invalidConfig)
		require.NoError(ts.T(), err) // Creation should succeed

		err = provider.Validate()
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "AWS access key and secret key are required")
	})

	ts.Run("Provider validation with missing region", func() {
		invalidConfig := providers.ProviderConfig{
			// Missing Region
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "test-key",
				SecretKey: "test-secret",
			},
		}

		provider, err := New(invalidConfig)
		require.NoError(ts.T(), err) // Creation should succeed

		err = provider.Validate()
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "AWS region is required")
	})

	ts.Run("Provider health check", func() {
		health, err := ts.provider.Health()
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), health)
		require.True(ts.T(), health.Healthy)
		require.Equal(ts.T(), "healthy", health.Status)
		require.NotEmpty(ts.T(), health.Services)

		// Check specific service health
		require.Contains(ts.T(), health.Services, "ec2")
		require.Contains(ts.T(), health.Services, "s3")
		require.Contains(ts.T(), health.Services, "rds")

		ec2Health := health.Services["ec2"]
		require.True(ts.T(), ec2Health.Available)
		require.Greater(ts.T(), ec2Health.ResponseTime, time.Duration(0))
	})

	ts.Run("Provider close", func() {
		err := ts.provider.Close()
		require.NoError(ts.T(), err)
	})
}

// TestAWSServiceAccessors tests service accessor methods
func (ts *AWSProviderTestSuite) TestAWSServiceAccessors() {
	ts.Run("All service accessors return non-nil services", func() {
		require.NotNil(ts.T(), ts.provider.Compute())
		require.NotNil(ts.T(), ts.provider.Storage())
		require.NotNil(ts.T(), ts.provider.Network())
		require.NotNil(ts.T(), ts.provider.Container())
		require.NotNil(ts.T(), ts.provider.Database())
		require.NotNil(ts.T(), ts.provider.Security())
		require.NotNil(ts.T(), ts.provider.Monitoring())
		require.NotNil(ts.T(), ts.provider.Serverless())
		require.NotNil(ts.T(), ts.provider.AI())
		require.NotNil(ts.T(), ts.provider.Cost())
		require.NotNil(ts.T(), ts.provider.Compliance())
		require.NotNil(ts.T(), ts.provider.Disaster())
		require.NotNil(ts.T(), ts.provider.Edge())
		require.NotNil(ts.T(), ts.provider.Quantum())
	})
}

// TestAWSComputeService tests AWS EC2 compute service
func (ts *AWSProviderTestSuite) TestAWSComputeService() {
	ctx := context.Background()
	compute := ts.provider.Compute()

	ts.Run("Create EC2 instance", func() {
		req := &providers.CreateInstanceRequest{
			Name:           "test-ec2-instance",
			Type:           "t3.medium",
			Image:          "ami-12345678",
			Region:         "us-east-1",
			Zone:           "us-east-1a",
			KeyPair:        "my-keypair",
			SecurityGroups: []string{"sg-12345"},
			Tags: map[string]string{
				"Environment": "test",
				"Project":     "mage",
			},
		}

		instance, err := compute.CreateInstance(ctx, req)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), instance)
		require.Equal(ts.T(), "test-ec2-instance", instance.Name)
		require.Equal(ts.T(), "t3.medium", instance.Type)
		require.Equal(ts.T(), "pending", instance.State)
		require.Equal(ts.T(), "us-east-1", instance.Region)
		require.Equal(ts.T(), "us-east-1a", instance.Zone)
		require.Contains(ts.T(), instance.ID, "i-")
		require.Equal(ts.T(), "test", instance.Tags["Environment"])
	})

	ts.Run("Get EC2 instance", func() {
		instance, err := compute.GetInstance(ctx, "i-1234567890abcdef0")
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), instance)
		require.Equal(ts.T(), "i-1234567890abcdef0", instance.ID)
		require.Equal(ts.T(), "example-instance", instance.Name)
		require.Equal(ts.T(), "running", instance.State)
		require.NotEmpty(ts.T(), instance.PublicIP)
		require.NotEmpty(ts.T(), instance.PrivateIP)
	})

	ts.Run("List EC2 instances", func() {
		filter := &providers.InstanceFilter{
			States:  []string{"running", "stopped"},
			Regions: []string{"us-east-1"},
		}

		instances, err := compute.ListInstances(ctx, filter)
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), instances)
		require.Len(ts.T(), instances, 2)

		// Check first instance
		webServer := instances[0]
		require.Equal(ts.T(), "web-server-1", webServer.Name)
		require.Equal(ts.T(), "t3.large", webServer.Type)
		require.Equal(ts.T(), "running", webServer.State)

		// Check second instance
		dbServer := instances[1]
		require.Equal(ts.T(), "db-server-1", dbServer.Name)
		require.Equal(ts.T(), "r5.xlarge", dbServer.Type)
		require.Empty(ts.T(), dbServer.PublicIP) // Private instance
	})

	ts.Run("Instance lifecycle operations", func() {
		instanceID := "i-1234567890abcdef0"

		err := compute.StartInstance(ctx, instanceID)
		require.NoError(ts.T(), err)

		err = compute.StopInstance(ctx, instanceID)
		require.NoError(ts.T(), err)

		err = compute.RestartInstance(ctx, instanceID)
		require.NoError(ts.T(), err)

		err = compute.ResizeInstance(ctx, instanceID, "t3.large")
		require.NoError(ts.T(), err)
	})

	ts.Run("Instance snapshot", func() {
		instanceID := "i-1234567890abcdef0"
		snapshotName := "test-snapshot"

		snapshot, err := compute.SnapshotInstance(ctx, instanceID, snapshotName)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), snapshot)
		require.Equal(ts.T(), snapshotName, snapshot.Name)
		require.Equal(ts.T(), instanceID, snapshot.InstanceID)
		require.Equal(ts.T(), "pending", snapshot.State)
		require.Contains(ts.T(), snapshot.ID, "snap-")
		require.Greater(ts.T(), snapshot.Size, int64(0))
	})

	ts.Run("Instance clone", func() {
		instanceID := "i-1234567890abcdef0"
		cloneReq := &providers.CloneRequest{
			Name: "cloned-instance",
			Zone: "us-east-1b",
			Type: "t3.medium",
		}

		clonedInstance, err := compute.CloneInstance(ctx, instanceID, cloneReq)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), clonedInstance)
		require.Equal(ts.T(), "cloned-instance", clonedInstance.Name)
		require.Equal(ts.T(), "t3.medium", clonedInstance.Type)
		require.Contains(ts.T(), clonedInstance.ID, "i-")
	})
}

// TestAWSStorageService tests AWS S3 storage service
func (ts *AWSProviderTestSuite) TestAWSStorageService() {
	ctx := context.Background()
	storage := ts.provider.Storage()

	ts.Run("Create S3 bucket", func() {
		req := &providers.CreateBucketRequest{
			Name:         "test-mage-bucket",
			Region:       "us-east-1",
			Versioning:   true,
			Encryption:   true,
			PublicAccess: false,
			Tags: map[string]string{
				"Purpose": "testing",
				"Owner":   "mage",
			},
			ACL: "private",
		}

		bucket, err := storage.CreateBucket(ctx, req)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), bucket)
		require.Equal(ts.T(), "test-mage-bucket", bucket.Name)
		require.Equal(ts.T(), "us-east-1", bucket.Region)
		require.True(ts.T(), bucket.Versioning)
		require.True(ts.T(), bucket.Encryption)
		require.False(ts.T(), bucket.PublicAccess)
		require.Equal(ts.T(), "testing", bucket.Tags["Purpose"])
	})

	ts.Run("Get S3 bucket", func() {
		bucket, err := storage.GetBucket(ctx, "test-mage-bucket")
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), bucket)
		require.Equal(ts.T(), "test-mage-bucket", bucket.Name)
		require.Equal(ts.T(), "us-east-1", bucket.Region)
		require.True(ts.T(), bucket.Versioning)
		require.True(ts.T(), bucket.Encryption)
	})

	ts.Run("List S3 buckets", func() {
		buckets, err := storage.ListBuckets(ctx)
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), buckets)
		require.Len(ts.T(), buckets, 3)

		bucketNames := make([]string, len(buckets))
		for i, bucket := range buckets {
			bucketNames[i] = bucket.Name
		}
		require.Contains(ts.T(), bucketNames, "my-app-assets")
		require.Contains(ts.T(), bucketNames, "my-app-backups")
		require.Contains(ts.T(), bucketNames, "my-app-logs")
	})

	ts.Run("Object operations", func() {
		bucketName := "test-mage-bucket"
		objectKey := "test/file.txt"
		content := "This is test content for S3 object"

		// Put object
		opts := &providers.PutOptions{
			ContentType:  "text/plain",
			CacheControl: "max-age=3600",
			Metadata: map[string]string{
				"uploaded-by": "mage-test",
				"version":     "1.0",
			},
			ACL:        "private",
			Encryption: "AES256",
		}
		err := storage.PutObject(ctx, bucketName, objectKey, strings.NewReader(content), opts)
		require.NoError(ts.T(), err)

		// List objects
		objects, err := storage.ListObjects(ctx, bucketName, "test/")
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), objects)
		require.Len(ts.T(), objects, 2)

		// Check object properties
		file1 := objects[0]
		require.Contains(ts.T(), file1.Key, "file1.txt")
		require.Equal(ts.T(), int64(1024), file1.Size)
		require.Equal(ts.T(), "text/plain", file1.ContentType)

		file2 := objects[1]
		require.Contains(ts.T(), file2.Key, "file2.jpg")
		require.Equal(ts.T(), int64(2048576), file2.Size)
		require.Equal(ts.T(), "image/jpeg", file2.ContentType)

		// Delete object
		err = storage.DeleteObject(ctx, bucketName, objectKey)
		require.NoError(ts.T(), err)
	})

	ts.Run("Advanced S3 operations", func() {
		bucketName := "test-mage-bucket"

		// Multipart upload
		err := storage.MultipartUpload(ctx, bucketName, "large-file.dat", strings.NewReader("large file content"))
		require.NoError(ts.T(), err)

		// Generate presigned URL
		url, err := storage.GeneratePresignedURL(ctx, bucketName, "test-file.txt", time.Hour)
		require.NoError(ts.T(), err)
		require.Contains(ts.T(), url, bucketName)
		require.Contains(ts.T(), url, "test-file.txt")
		require.Contains(ts.T(), url, "amazonaws.com")
		require.Contains(ts.T(), url, "token=")
		require.Contains(ts.T(), url, "expires=")

		// Set object ACL
		acl := &providers.ACL{
			Owner: "bucket-owner",
			Grants: []*providers.Grant{
				{Grantee: "user1", Permission: "READ"},
				{Grantee: "user2", Permission: "WRITE"},
			},
		}
		err = storage.SetObjectACL(ctx, bucketName, "test-file.txt", acl)
		require.NoError(ts.T(), err)
	})

	ts.Run("Delete S3 bucket", func() {
		err := storage.DeleteBucket(ctx, "test-mage-bucket")
		require.NoError(ts.T(), err)
	})
}

// TestAWSNetworkService tests AWS VPC/networking service
func (ts *AWSProviderTestSuite) TestAWSNetworkService() {
	ctx := context.Background()
	network := ts.provider.Network()

	ts.Run("VPC operations", func() {
		// Create VPC
		vpcReq := &providers.CreateVPCRequest{
			Name:      "test-vpc",
			CIDR:      "10.0.0.0/16",
			Region:    "us-east-1",
			EnableDNS: true,
			Tags: map[string]string{
				"Environment": "test",
				"Purpose":     "mage-testing",
			},
		}

		vpc, err := network.CreateVPC(ctx, vpcReq)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), vpc)
		require.Equal(ts.T(), "test-vpc", vpc.Name)
		require.Equal(ts.T(), "10.0.0.0/16", vpc.CIDR)
		require.Equal(ts.T(), "us-east-1", vpc.Region)
		require.Equal(ts.T(), "available", vpc.State)
		require.Contains(ts.T(), vpc.ID, "vpc-")
		require.Equal(ts.T(), "test", vpc.Tags["Environment"])

		// Get VPC
		retrievedVPC, err := network.GetVPC(ctx, vpc.ID)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), vpc.ID, retrievedVPC.ID)
		require.Equal(ts.T(), "my-vpc", retrievedVPC.Name)
		require.Equal(ts.T(), "available", retrievedVPC.State)

		// List VPCs
		vpcs, err := network.ListVPCs(ctx)
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), vpcs)
		require.Len(ts.T(), vpcs, 2)

		defaultVPC := vpcs[0]
		require.Equal(ts.T(), "vpc-12345", defaultVPC.ID)
		require.Equal(ts.T(), "default", defaultVPC.Name)
		require.Equal(ts.T(), "172.31.0.0/16", defaultVPC.CIDR)

		prodVPC := vpcs[1]
		require.Equal(ts.T(), "vpc-67890", prodVPC.ID)
		require.Equal(ts.T(), "production", prodVPC.Name)
		require.Equal(ts.T(), "10.0.0.0/16", prodVPC.CIDR)

		// Delete VPC
		err = network.DeleteVPC(ctx, vpc.ID)
		require.NoError(ts.T(), err)
	})

	ts.Run("Subnet operations", func() {
		vpcID := "vpc-12345"

		// Create subnet
		subnetReq := &providers.CreateSubnetRequest{
			Name:   "test-subnet",
			CIDR:   "10.0.1.0/24",
			Zone:   "us-east-1a",
			Public: true,
			Tags: map[string]string{
				"Type": "public",
			},
		}

		subnet, err := network.CreateSubnet(ctx, vpcID, subnetReq)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), subnet)
		require.Equal(ts.T(), "test-subnet", subnet.Name)
		require.Equal(ts.T(), vpcID, subnet.VPCID)
		require.Equal(ts.T(), "10.0.1.0/24", subnet.CIDR)
		require.Equal(ts.T(), "us-east-1a", subnet.Zone)
		require.True(ts.T(), subnet.Public)
		require.Equal(ts.T(), "available", subnet.State)
		require.Contains(ts.T(), subnet.ID, "subnet-")

		// Get subnet
		retrievedSubnet, err := network.GetSubnet(ctx, subnet.ID)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), subnet.ID, retrievedSubnet.ID)

		// List subnets
		subnets, err := network.ListSubnets(ctx, vpcID)
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), subnets)
		require.Len(ts.T(), subnets, 2)

		publicSubnet := subnets[0]
		require.Equal(ts.T(), "public-1a", publicSubnet.Name)
		require.True(ts.T(), publicSubnet.Public)

		privateSubnet := subnets[1]
		require.Equal(ts.T(), "private-1b", privateSubnet.Name)
		require.False(ts.T(), privateSubnet.Public)

		// Delete subnet
		err = network.DeleteSubnet(ctx, subnet.ID)
		require.NoError(ts.T(), err)
	})

	ts.Run("Security group operations", func() {
		sgReq := &providers.CreateSecurityGroupRequest{
			Name:        "test-security-group",
			Description: "Test security group for mage",
			VPCID:       "vpc-12345",
			Rules: []*providers.SecurityRule{
				{
					Direction:   "ingress",
					Protocol:    "tcp",
					FromPort:    80,
					ToPort:      80,
					Source:      "0.0.0.0/0",
					Description: "HTTP access",
				},
				{
					Direction:   "ingress",
					Protocol:    "tcp",
					FromPort:    443,
					ToPort:      443,
					Source:      "0.0.0.0/0",
					Description: "HTTPS access",
				},
				{
					Direction:   "egress",
					Protocol:    "tcp",
					FromPort:    0,
					ToPort:      65535,
					Source:      "0.0.0.0/0",
					Description: "All outbound traffic",
				},
			},
		}

		sg, err := network.CreateSecurityGroup(ctx, sgReq)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), sg)
		require.Equal(ts.T(), "test-security-group", sg.Name)
		require.Equal(ts.T(), "Test security group for mage", sg.Description)
		require.Equal(ts.T(), "vpc-12345", sg.VPCID)
		require.Len(ts.T(), sg.Rules, 3)
		require.Contains(ts.T(), sg.ID, "sg-")

		// Check rules
		httpRule := sg.Rules[0]
		require.Equal(ts.T(), "ingress", httpRule.Direction)
		require.Equal(ts.T(), "tcp", httpRule.Protocol)
		require.Equal(ts.T(), 80, httpRule.FromPort)
		require.Equal(ts.T(), 80, httpRule.ToPort)

		// Update security rules
		newRules := []*providers.SecurityRule{
			{
				Direction: "ingress",
				Protocol:  "tcp",
				FromPort:  22,
				ToPort:    22,
				Source:    "10.0.0.0/16",
			},
		}
		err = network.UpdateSecurityRules(ctx, sg.ID, newRules)
		require.NoError(ts.T(), err)
	})

	ts.Run("Load balancer operations", func() {
		lbReq := &providers.CreateLoadBalancerRequest{
			Name:           "test-load-balancer",
			Type:           "application",
			Subnets:        []string{"subnet-12345", "subnet-67890"},
			SecurityGroups: []string{"sg-12345"},
			Listeners: []*providers.Listener{
				{
					Protocol:      "HTTP",
					Port:          80,
					TargetGroupID: "tg-12345",
				},
				{
					Protocol:      "HTTPS",
					Port:          443,
					TargetGroupID: "tg-67890",
					Certificate:   "arn:aws:acm:us-east-1:123456789012:certificate/12345",
				},
			},
			Tags: map[string]string{
				"Environment": "test",
			},
		}

		lb, err := network.CreateLoadBalancer(ctx, lbReq)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), lb)
		require.Equal(ts.T(), "test-load-balancer", lb.Name)
		require.Equal(ts.T(), "application", lb.Type)
		require.Equal(ts.T(), "provisioning", lb.State)
		require.Contains(ts.T(), lb.DNSName, "elb.amazonaws.com")
		require.Len(ts.T(), lb.Listeners, 2)
		require.Contains(ts.T(), lb.ID, "lb-")

		// Check listeners
		httpListener := lb.Listeners[0]
		require.Equal(ts.T(), "HTTP", httpListener.Protocol)
		require.Equal(ts.T(), 80, httpListener.Port)

		httpsListener := lb.Listeners[1]
		require.Equal(ts.T(), "HTTPS", httpsListener.Protocol)
		require.Equal(ts.T(), 443, httpsListener.Port)
		require.NotEmpty(ts.T(), httpsListener.Certificate)

		// Update load balancer
		updateReq := &providers.UpdateLoadBalancerRequest{
			Name:           stringPtr("updated-load-balancer"),
			SecurityGroups: []string{"sg-new"},
			Tags: map[string]string{
				"Updated": "true",
			},
		}
		err = network.UpdateLoadBalancer(ctx, lb.ID, updateReq)
		require.NoError(ts.T(), err)
	})
}

// TestHelperFunctions tests AWS provider helper functions
func (ts *AWSProviderTestSuite) TestHelperFunctions() {
	ts.Run("ID generation", func() {
		id1 := generateID()
		id2 := generateID()

		require.NotEmpty(ts.T(), id1)
		require.NotEmpty(ts.T(), id2)
		require.NotEqual(ts.T(), id1, id2) // Should be unique
	})

	ts.Run("IP generation", func() {
		publicIP := generateIP()
		privateIP := generatePrivateIP()

		require.NotEmpty(ts.T(), publicIP)
		require.NotEmpty(ts.T(), privateIP)

		// Check IP format
		require.True(ts.T(), strings.HasPrefix(publicIP, "54."))
		require.True(ts.T(), strings.HasPrefix(privateIP, "10.0."))

		// Should contain 4 octets
		require.Len(ts.T(), strings.Split(publicIP, "."), 4)
		require.Len(ts.T(), strings.Split(privateIP, "."), 4)
	})
}

// TestAWSProviderRegistration tests AWS provider registration
func (ts *AWSProviderTestSuite) TestAWSProviderRegistration() {
	ts.Run("AWS provider is registered", func() {
		// The init() function should have registered the AWS provider
		providerNames := providers.List()
		require.Contains(ts.T(), providerNames, "aws")
	})

	ts.Run("Get AWS provider from registry", func() {
		config := providers.ProviderConfig{
			Region: "us-west-2",
			Credentials: providers.Credentials{
				Type:      "key",
				AccessKey: "test-key",
				SecretKey: "test-secret",
			},
		}

		provider, err := providers.Get("aws", config)
		require.NoError(ts.T(), err)
		require.NotNil(ts.T(), provider)
		require.Equal(ts.T(), "aws", provider.Name())

		// Verify it's the correct type
		awsProvider, ok := provider.(*Provider)
		require.True(ts.T(), ok)
		require.Equal(ts.T(), "us-west-2", awsProvider.region)
	})
}

// TestAWSProviderTestSuite runs the test suite
func TestAWSProviderTestSuite(t *testing.T) {
	suite.Run(t, new(AWSProviderTestSuite))
}

// Helper function for pointer to string
func stringPtr(s string) *string {
	return &s
}
