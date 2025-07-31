package aws

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/providers"
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
	ts.provider, err = New(&ts.config)
	ts.Require().NoError(err)
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
		ts.Require().Equal("aws", ts.provider.Name())
	})

	ts.Run("Provider initialization", func() {
		awsProvider, ok := ts.provider.(*Provider)
		ts.Require().True(ok, "Provider should be of type *Provider")
		ts.Require().Equal("us-east-1", awsProvider.region)
		ts.Require().NotNil(awsProvider.services)
		ts.Require().NotNil(awsProvider.services.compute)
		ts.Require().NotNil(awsProvider.services.storage)
		ts.Require().NotNil(awsProvider.services.network)
	})

	ts.Run("Provider validation with valid credentials", func() {
		err := ts.provider.Validate()
		ts.Require().NoError(err)
	})

	ts.Run("Provider validation with missing credentials", func() {
		invalidConfig := providers.ProviderConfig{
			Region: "us-east-1",
			Credentials: providers.Credentials{
				Type: "key",
				// Missing AccessKey and SecretKey
			},
		}

		provider, err := New(&invalidConfig)
		ts.Require().NoError(err) // Creation should succeed

		err = provider.Validate()
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "AWS access key and secret key are required")
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

		provider, err := New(&invalidConfig)
		ts.Require().NoError(err) // Creation should succeed

		err = provider.Validate()
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "AWS region is required")
	})

	ts.Run("Provider health check", func() {
		health, err := ts.provider.Health()
		ts.Require().NoError(err)
		ts.Require().NotNil(health)
		ts.Require().True(health.Healthy)
		ts.Require().Equal("healthy", health.Status)
		ts.Require().NotEmpty(health.Services)

		// Check specific service health
		ts.Require().Contains(health.Services, "ec2")
		ts.Require().Contains(health.Services, "s3")
		ts.Require().Contains(health.Services, "rds")

		ec2Health := health.Services["ec2"]
		ts.Require().True(ec2Health.Available)
		ts.Require().Greater(ec2Health.ResponseTime, time.Duration(0))
	})

	ts.Run("Provider close", func() {
		err := ts.provider.Close()
		ts.Require().NoError(err)
	})
}

// TestAWSServiceAccessors tests service accessor methods
func (ts *AWSProviderTestSuite) TestAWSServiceAccessors() {
	ts.Run("All service accessors return non-nil services", func() {
		ts.Require().NotNil(ts.provider.Compute())
		ts.Require().NotNil(ts.provider.Storage())
		ts.Require().NotNil(ts.provider.Network())
		ts.Require().NotNil(ts.provider.Container())
		ts.Require().NotNil(ts.provider.Database())
		ts.Require().NotNil(ts.provider.Security())
		ts.Require().NotNil(ts.provider.Monitoring())
		ts.Require().NotNil(ts.provider.Serverless())
		ts.Require().NotNil(ts.provider.AI())
		ts.Require().NotNil(ts.provider.Cost())
		ts.Require().NotNil(ts.provider.Compliance())
		ts.Require().NotNil(ts.provider.Disaster())
		ts.Require().NotNil(ts.provider.Edge())
		ts.Require().NotNil(ts.provider.Quantum())
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
		ts.Require().NoError(err)
		ts.Require().NotNil(instance)
		ts.Require().Equal("test-ec2-instance", instance.Name)
		ts.Require().Equal("t3.medium", instance.Type)
		ts.Require().Equal("pending", instance.State)
		ts.Require().Equal("us-east-1", instance.Region)
		ts.Require().Equal("us-east-1a", instance.Zone)
		ts.Require().Contains(instance.ID, "i-")
		ts.Require().Equal("test", instance.Tags["Environment"])
	})

	ts.Run("Get EC2 instance", func() {
		instance, err := compute.GetInstance(ctx, "i-1234567890abcdef0")
		ts.Require().NoError(err)
		ts.Require().NotNil(instance)
		ts.Require().Equal("i-1234567890abcdef0", instance.ID)
		ts.Require().Equal("example-instance", instance.Name)
		ts.Require().Equal("running", instance.State)
		ts.Require().NotEmpty(instance.PublicIP)
		ts.Require().NotEmpty(instance.PrivateIP)
	})

	ts.Run("List EC2 instances", func() {
		filter := &providers.InstanceFilter{
			States:  []string{"running", "stopped"},
			Regions: []string{"us-east-1"},
		}

		instances, err := compute.ListInstances(ctx, filter)
		ts.Require().NoError(err)
		ts.Require().NotEmpty(instances)
		ts.Require().Len(instances, 2)

		// Check first instance
		webServer := instances[0]
		ts.Require().Equal("web-server-1", webServer.Name)
		ts.Require().Equal("t3.large", webServer.Type)
		ts.Require().Equal("running", webServer.State)

		// Check second instance
		dbServer := instances[1]
		ts.Require().Equal("db-server-1", dbServer.Name)
		ts.Require().Equal("r5.xlarge", dbServer.Type)
		ts.Require().Empty(dbServer.PublicIP) // Private instance
	})

	ts.Run("Instance lifecycle operations", func() {
		instanceID := "i-1234567890abcdef0"

		err := compute.StartInstance(ctx, instanceID)
		ts.Require().NoError(err)

		err = compute.StopInstance(ctx, instanceID)
		ts.Require().NoError(err)

		err = compute.RestartInstance(ctx, instanceID)
		ts.Require().NoError(err)

		err = compute.ResizeInstance(ctx, instanceID, "t3.large")
		ts.Require().NoError(err)
	})

	ts.Run("Instance snapshot", func() {
		instanceID := "i-1234567890abcdef0"
		snapshotName := "test-snapshot"

		snapshot, err := compute.SnapshotInstance(ctx, instanceID, snapshotName)
		ts.Require().NoError(err)
		ts.Require().NotNil(snapshot)
		ts.Require().Equal(snapshotName, snapshot.Name)
		ts.Require().Equal(instanceID, snapshot.InstanceID)
		ts.Require().Equal("pending", snapshot.State)
		ts.Require().Contains(snapshot.ID, "snap-")
		ts.Require().Positive(snapshot.Size)
	})

	ts.Run("Instance clone", func() {
		instanceID := "i-1234567890abcdef0"
		cloneReq := &providers.CloneRequest{
			Name: "cloned-instance",
			Zone: "us-east-1b",
			Type: "t3.medium",
		}

		clonedInstance, err := compute.CloneInstance(ctx, instanceID, cloneReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(clonedInstance)
		ts.Require().Equal("cloned-instance", clonedInstance.Name)
		ts.Require().Equal("t3.medium", clonedInstance.Type)
		ts.Require().Contains(clonedInstance.ID, "i-")
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
		ts.Require().NoError(err)
		ts.Require().NotNil(bucket)
		ts.Require().Equal("test-mage-bucket", bucket.Name)
		ts.Require().Equal("us-east-1", bucket.Region)
		ts.Require().True(bucket.Versioning)
		ts.Require().True(bucket.Encryption)
		ts.Require().False(bucket.PublicAccess)
		ts.Require().Equal("testing", bucket.Tags["Purpose"])
	})

	ts.Run("Get S3 bucket", func() {
		bucket, err := storage.GetBucket(ctx, "test-mage-bucket")
		ts.Require().NoError(err)
		ts.Require().NotNil(bucket)
		ts.Require().Equal("test-mage-bucket", bucket.Name)
		ts.Require().Equal("us-east-1", bucket.Region)
		ts.Require().True(bucket.Versioning)
		ts.Require().True(bucket.Encryption)
	})

	ts.Run("List S3 buckets", func() {
		buckets, err := storage.ListBuckets(ctx)
		ts.Require().NoError(err)
		ts.Require().NotEmpty(buckets)
		ts.Require().Len(buckets, 3)

		bucketNames := make([]string, len(buckets))
		for i, bucket := range buckets {
			bucketNames[i] = bucket.Name
		}
		ts.Require().Contains(bucketNames, "my-app-assets")
		ts.Require().Contains(bucketNames, "my-app-backups")
		ts.Require().Contains(bucketNames, "my-app-logs")
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
		ts.Require().NoError(err)

		// List objects
		objects, err := storage.ListObjects(ctx, bucketName, "test/")
		ts.Require().NoError(err)
		ts.Require().NotEmpty(objects)
		ts.Require().Len(objects, 2)

		// Check object properties
		file1 := objects[0]
		ts.Require().Contains(file1.Key, "file1.txt")
		ts.Require().Equal(int64(1024), file1.Size)
		ts.Require().Equal("text/plain", file1.ContentType)

		file2 := objects[1]
		ts.Require().Contains(file2.Key, "file2.jpg")
		ts.Require().Equal(int64(2048576), file2.Size)
		ts.Require().Equal("image/jpeg", file2.ContentType)

		// Delete object
		err = storage.DeleteObject(ctx, bucketName, objectKey)
		ts.Require().NoError(err)
	})

	ts.Run("Advanced S3 operations", func() {
		bucketName := "test-mage-bucket"

		// Multipart upload
		err := storage.MultipartUpload(ctx, bucketName, "large-file.dat", strings.NewReader("large file content"))
		ts.Require().NoError(err)

		// Generate presigned URL
		url, err := storage.GeneratePresignedURL(ctx, bucketName, "test-file.txt", time.Hour)
		ts.Require().NoError(err)
		ts.Require().Contains(url, bucketName)
		ts.Require().Contains(url, "test-file.txt")
		ts.Require().Contains(url, "amazonaws.com")
		ts.Require().Contains(url, "token=")
		ts.Require().Contains(url, "expires=")

		// Set object ACL
		acl := &providers.ACL{
			Owner: "bucket-owner",
			Grants: []*providers.Grant{
				{Grantee: "user1", Permission: "READ"},
				{Grantee: "user2", Permission: "WRITE"},
			},
		}
		err = storage.SetObjectACL(ctx, bucketName, "test-file.txt", acl)
		ts.Require().NoError(err)
	})

	ts.Run("Delete S3 bucket", func() {
		err := storage.DeleteBucket(ctx, "test-mage-bucket")
		ts.Require().NoError(err)
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
		ts.Require().NoError(err)
		ts.Require().NotNil(vpc)
		ts.Require().Equal("test-vpc", vpc.Name)
		ts.Require().Equal("10.0.0.0/16", vpc.CIDR)
		ts.Require().Equal("us-east-1", vpc.Region)
		ts.Require().Equal("available", vpc.State)
		ts.Require().Contains(vpc.ID, "vpc-")
		ts.Require().Equal("test", vpc.Tags["Environment"])

		// Get VPC
		retrievedVPC, err := network.GetVPC(ctx, vpc.ID)
		ts.Require().NoError(err)
		ts.Require().Equal(vpc.ID, retrievedVPC.ID)
		ts.Require().Equal("my-vpc", retrievedVPC.Name)
		ts.Require().Equal("available", retrievedVPC.State)

		// List VPCs
		vpcs, err := network.ListVPCs(ctx)
		ts.Require().NoError(err)
		ts.Require().NotEmpty(vpcs)
		ts.Require().Len(vpcs, 2)

		defaultVPC := vpcs[0]
		ts.Require().Equal("vpc-12345", defaultVPC.ID)
		ts.Require().Equal("default", defaultVPC.Name)
		ts.Require().Equal("172.31.0.0/16", defaultVPC.CIDR)

		prodVPC := vpcs[1]
		ts.Require().Equal("vpc-67890", prodVPC.ID)
		ts.Require().Equal("production", prodVPC.Name)
		ts.Require().Equal("10.0.0.0/16", prodVPC.CIDR)

		// Delete VPC
		err = network.DeleteVPC(ctx, vpc.ID)
		ts.Require().NoError(err)
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
		ts.Require().NoError(err)
		ts.Require().NotNil(subnet)
		ts.Require().Equal("test-subnet", subnet.Name)
		ts.Require().Equal(vpcID, subnet.VPCID)
		ts.Require().Equal("10.0.1.0/24", subnet.CIDR)
		ts.Require().Equal("us-east-1a", subnet.Zone)
		ts.Require().True(subnet.Public)
		ts.Require().Equal("available", subnet.State)
		ts.Require().Contains(subnet.ID, "subnet-")

		// Get subnet
		retrievedSubnet, err := network.GetSubnet(ctx, subnet.ID)
		ts.Require().NoError(err)
		ts.Require().Equal(subnet.ID, retrievedSubnet.ID)

		// List subnets
		subnets, err := network.ListSubnets(ctx, vpcID)
		ts.Require().NoError(err)
		ts.Require().NotEmpty(subnets)
		ts.Require().Len(subnets, 2)

		publicSubnet := subnets[0]
		ts.Require().Equal("public-1a", publicSubnet.Name)
		ts.Require().True(publicSubnet.Public)

		privateSubnet := subnets[1]
		ts.Require().Equal("private-1b", privateSubnet.Name)
		ts.Require().False(privateSubnet.Public)

		// Delete subnet
		err = network.DeleteSubnet(ctx, subnet.ID)
		ts.Require().NoError(err)
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
		ts.Require().NoError(err)
		ts.Require().NotNil(sg)
		ts.Require().Equal("test-security-group", sg.Name)
		ts.Require().Equal("Test security group for mage", sg.Description)
		ts.Require().Equal("vpc-12345", sg.VPCID)
		ts.Require().Len(sg.Rules, 3)
		ts.Require().Contains(sg.ID, "sg-")

		// Check rules
		httpRule := sg.Rules[0]
		ts.Require().Equal("ingress", httpRule.Direction)
		ts.Require().Equal("tcp", httpRule.Protocol)
		ts.Require().Equal(80, httpRule.FromPort)
		ts.Require().Equal(80, httpRule.ToPort)

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
		ts.Require().NoError(err)
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
		ts.Require().NoError(err)
		ts.Require().NotNil(lb)
		ts.Require().Equal("test-load-balancer", lb.Name)
		ts.Require().Equal("application", lb.Type)
		ts.Require().Equal("provisioning", lb.State)
		ts.Require().Contains(lb.DNSName, "elb.amazonaws.com")
		ts.Require().Len(lb.Listeners, 2)
		ts.Require().Contains(lb.ID, "lb-")

		// Check listeners
		httpListener := lb.Listeners[0]
		ts.Require().Equal("HTTP", httpListener.Protocol)
		ts.Require().Equal(80, httpListener.Port)

		httpsListener := lb.Listeners[1]
		ts.Require().Equal("HTTPS", httpsListener.Protocol)
		ts.Require().Equal(443, httpsListener.Port)
		ts.Require().NotEmpty(httpsListener.Certificate)

		// Update load balancer
		updateReq := &providers.UpdateLoadBalancerRequest{
			Name:           stringPtr("updated-load-balancer"),
			SecurityGroups: []string{"sg-new"},
			Tags: map[string]string{
				"Updated": "true",
			},
		}
		err = network.UpdateLoadBalancer(ctx, lb.ID, updateReq)
		ts.Require().NoError(err)
	})
}

// TestHelperFunctions tests AWS provider helper functions
func (ts *AWSProviderTestSuite) TestHelperFunctions() {
	ts.Run("ID generation", func() {
		id1 := generateID()
		id2 := generateID()

		ts.Require().NotEmpty(id1)
		ts.Require().NotEmpty(id2)
		ts.Require().NotEqual(id1, id2) // Should be unique
	})

	ts.Run("IP generation", func() {
		publicIP := generateIP()
		privateIP := generatePrivateIP()

		ts.Require().NotEmpty(publicIP)
		ts.Require().NotEmpty(privateIP)

		// Check IP format
		ts.Require().True(strings.HasPrefix(publicIP, "54."))
		ts.Require().True(strings.HasPrefix(privateIP, "10.0."))

		// Should contain 4 octets
		ts.Require().Len(strings.Split(publicIP, "."), 4)
		ts.Require().Len(strings.Split(privateIP, "."), 4)
	})
}

// TestAWSProviderRegistration tests AWS provider registration
func (ts *AWSProviderTestSuite) TestAWSProviderRegistration() {
	ts.Run("AWS provider is registered", func() {
		// The init() function should have registered the AWS provider
		providerNames := providers.List()
		ts.Require().Contains(providerNames, "aws")
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

		provider, err := providers.Get("aws", &config)
		ts.Require().NoError(err)
		ts.Require().NotNil(provider)
		ts.Require().Equal("aws", provider.Name())

		// Verify it's the correct type
		awsProvider, ok := provider.(*Provider)
		ts.Require().True(ok)
		ts.Require().Equal("us-west-2", awsProvider.region)
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
