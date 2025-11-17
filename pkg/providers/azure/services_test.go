package azure

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/providers"
)

// AzureServicesTestSuite tests Azure service implementations
type AzureServicesTestSuite struct {
	suite.Suite

	provider *Provider
}

// SetupTest runs before each test
func (ts *AzureServicesTestSuite) SetupTest() {
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

// TestAzureComputeService tests Azure compute service operations
func (ts *AzureServicesTestSuite) TestAzureComputeService() {
	ctx := context.Background()
	compute := ts.provider.Compute()
	ts.Require().NotNil(compute)

	ts.Run("CreateInstance", func() {
		req := &providers.CreateInstanceRequest{
			Name:           "test-vm",
			Type:           "Standard_D2s_v3",
			Image:          "Ubuntu-20.04-LTS",
			Region:         "eastus",
			Zone:           "eastus-1",
			KeyPair:        "my-key-pair",
			SecurityGroups: []string{"default"},
			UserData:       "#!/bin/bash\necho 'Hello Azure'",
			Tags: map[string]string{
				"Environment": "test",
				"Purpose":     "unit-testing",
			},
			DiskSize: 30,
		}

		instance, err := compute.CreateInstance(ctx, req)
		ts.Require().NoError(err)
		ts.Require().NotNil(instance)
		ts.Require().Equal("test-vm", instance.Name)
		ts.Require().Equal("Standard_D2s_v3", instance.Type)
		ts.Require().Equal("Creating", instance.State)
		ts.Require().Equal("eastus", instance.Region)
		ts.Require().Equal("eastus-1", instance.Zone)
		ts.Require().Contains(instance.ID, "sub-id")
		ts.Require().Contains(instance.ID, "Microsoft.Compute/virtualMachines")
		ts.Require().Equal(req.Tags, instance.Tags)
	})

	ts.Run("GetInstance", func() {
		id := "/subscriptions/test-sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/test-vm"
		instance, err := compute.GetInstance(ctx, id)
		ts.Require().NoError(err)
		ts.Require().NotNil(instance)
		ts.Require().Equal(id, instance.ID)
		ts.Require().Equal("azure-vm-1", instance.Name)
		ts.Require().Equal("Standard_D2s_v3", instance.Type)
		ts.Require().Equal("Running", instance.State)
	})

	ts.Run("ListInstances", func() {
		filter := &providers.InstanceFilter{
			States:  []string{"Running", "Stopped"},
			Regions: []string{"eastus", "westus2"},
			Types:   []string{"Standard_D2s_v3", "Standard_B2s"},
		}
		instances, err := compute.ListInstances(ctx, filter)
		ts.Require().NoError(err)
		ts.Require().NotNil(instances)
		ts.Require().Len(instances, 1)
		ts.Require().Equal("web-vm", instances[0].Name)
		ts.Require().Equal("Standard_B2s", instances[0].Type)
		ts.Require().Equal("Running", instances[0].State)
	})

	ts.Run("Instance lifecycle operations", func() {
		instanceID := "vm-test-lifecycle"

		// Test all lifecycle operations
		name := "updated-vm"
		ts.Require().NoError(compute.UpdateInstance(ctx, instanceID, &providers.UpdateInstanceRequest{Name: &name}))
		ts.Require().NoError(compute.StartInstance(ctx, instanceID))
		ts.Require().NoError(compute.StopInstance(ctx, instanceID))
		ts.Require().NoError(compute.RestartInstance(ctx, instanceID))
		ts.Require().NoError(compute.ResizeInstance(ctx, instanceID, "Standard_D4s_v3"))
		ts.Require().NoError(compute.DeleteInstance(ctx, instanceID))
	})

	ts.Run("SnapshotInstance", func() {
		instanceID := "vm-for-snapshot"
		snapshotName := "test-snapshot"

		snapshot, err := compute.SnapshotInstance(ctx, instanceID, snapshotName)
		ts.Require().NoError(err)
		ts.Require().NotNil(snapshot)
		ts.Require().Equal(snapshotName, snapshot.Name)
		ts.Require().Equal(instanceID, snapshot.InstanceID)
		ts.Require().Equal("Creating", snapshot.State)
		ts.Require().Contains(snapshot.ID, "snap-")
	})

	ts.Run("CloneInstance", func() {
		sourceID := "source-vm"
		cloneReq := &providers.CloneRequest{
			Name: "cloned-vm",
			Type: "Standard_D2s_v3",
			Zone: "eastus-2",
		}

		clonedInstance, err := compute.CloneInstance(ctx, sourceID, cloneReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(clonedInstance)
		ts.Require().Equal("cloned-vm", clonedInstance.Name)
		ts.Require().Equal("Standard_D2s_v3", clonedInstance.Type)
		ts.Require().Equal("eastus-2", clonedInstance.Zone)
	})
}

// TestAzureStorageService tests Azure storage service operations
func (ts *AzureServicesTestSuite) TestAzureStorageService() {
	ctx := context.Background()
	storage := ts.provider.Storage()
	ts.Require().NotNil(storage)

	ts.Run("CreateBucket", func() {
		req := &providers.CreateBucketRequest{
			Name:         "test-container",
			Region:       "eastus",
			Versioning:   true,
			Encryption:   true,
			PublicAccess: false,
			Tags: map[string]string{
				"Project":     "test-project",
				"Environment": "dev",
			},
		}

		bucket, err := storage.CreateBucket(ctx, req)
		ts.Require().NoError(err)
		ts.Require().NotNil(bucket)
		ts.Require().Equal("test-container", bucket.Name)
		ts.Require().Equal("eastus", bucket.Region)
		ts.Require().True(bucket.Versioning)
		ts.Require().True(bucket.Encryption)
		ts.Require().False(bucket.PublicAccess)
		ts.Require().Equal(req.Tags, bucket.Tags)
	})

	ts.Run("GetBucket", func() {
		bucketName := "existing-container"
		bucket, err := storage.GetBucket(ctx, bucketName)
		ts.Require().NoError(err)
		ts.Require().NotNil(bucket)
		ts.Require().Equal(bucketName, bucket.Name)
		ts.Require().Equal("eastus", bucket.Region) // From provider config
	})

	ts.Run("ListBuckets", func() {
		buckets, err := storage.ListBuckets(ctx)
		ts.Require().NoError(err)
		ts.Require().NotNil(buckets)
		ts.Require().Len(buckets, 1)
		ts.Require().Equal("container1", buckets[0].Name)
	})

	ts.Run("Object operations", func() {
		bucketName := "test-container"
		objectKey := "test-file.txt"
		content := "Hello Azure Blob Storage!"

		// PutObject
		opts := &providers.PutOptions{
			ContentType:  "text/plain",
			CacheControl: "max-age=3600",
			Metadata: map[string]string{
				"uploaded-by": "unit-test",
				"purpose":     "testing",
			},
		}
		err := storage.PutObject(ctx, bucketName, objectKey, strings.NewReader(content), opts)
		ts.Require().NoError(err)

		// ListObjects
		objects, err := storage.ListObjects(ctx, bucketName, "test-")
		ts.Require().NoError(err)
		ts.Require().NotNil(objects)
		ts.Require().Len(objects, 1)
		ts.Require().Equal("test-file.txt", objects[0].Key)

		// GetObject (returns error for Azure implementation)
		_, err = storage.GetObject(ctx, bucketName, objectKey)
		ts.Require().Error(err)
		ts.Require().ErrorIs(err, ErrNotImplemented)

		// MultipartUpload
		largeContent := strings.Repeat("Large content chunk ", 1000)
		err = storage.MultipartUpload(ctx, bucketName, "large-file.txt", strings.NewReader(largeContent))
		ts.Require().NoError(err)

		// GeneratePresignedURL
		url, err := storage.GeneratePresignedURL(ctx, bucketName, objectKey, time.Hour)
		ts.Require().NoError(err)
		ts.Require().Contains(url, "blob.core.windows.net")
		ts.Require().Contains(url, bucketName)
		ts.Require().Contains(url, objectKey)

		// SetObjectACL
		acl := &providers.ACL{
			Owner: "test-owner",
			Grants: []*providers.Grant{
				{Grantee: "user1", Permission: "READ"},
				{Grantee: "user2", Permission: "WRITE"},
			},
		}
		err = storage.SetObjectACL(ctx, bucketName, objectKey, acl)
		ts.Require().NoError(err)

		// DeleteObject
		err = storage.DeleteObject(ctx, bucketName, objectKey)
		ts.Require().NoError(err)
	})

	ts.Run("DeleteBucket", func() {
		err := storage.DeleteBucket(ctx, "test-container")
		ts.Require().NoError(err)
	})
}

// TestAzureNetworkService tests Azure network service operations
func (ts *AzureServicesTestSuite) TestAzureNetworkService() {
	ctx := context.Background()
	network := ts.provider.Network()
	ts.Require().NotNil(network)

	ts.Run("VPC operations", func() {
		// CreateVPC
		vpcReq := &providers.CreateVPCRequest{
			Name:      "test-vnet",
			CIDR:      "10.0.0.0/16",
			Region:    "eastus",
			EnableDNS: true,
			Tags: map[string]string{
				"Purpose": "testing",
				"Team":    "devops",
			},
		}
		vpc, err := network.CreateVPC(ctx, vpcReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(vpc)
		ts.Require().Equal("test-vnet", vpc.Name)
		ts.Require().Equal("10.0.0.0/16", vpc.CIDR)
		ts.Require().Equal("eastus", vpc.Region)
		ts.Require().Equal("Succeeded", vpc.State)
		ts.Require().Contains(vpc.ID, "vnet-test-vnet")
		ts.Require().Equal(vpcReq.Tags, vpc.Tags)

		// GetVPC
		retrievedVPC, err := network.GetVPC(ctx, vpc.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedVPC)
		ts.Require().Equal(vpc.ID, retrievedVPC.ID)
		ts.Require().Equal("azure-vnet", retrievedVPC.Name)
		ts.Require().Equal("Succeeded", retrievedVPC.State)

		// ListVPCs
		vpcs, err := network.ListVPCs(ctx)
		ts.Require().NoError(err)
		ts.Require().NotNil(vpcs)
		ts.Require().Len(vpcs, 1)
		ts.Require().Equal("default", vpcs[0].Name)

		// DeleteVPC
		err = network.DeleteVPC(ctx, vpc.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Subnet operations", func() {
		vpcID := "vnet-12345"

		// CreateSubnet
		subnetReq := &providers.CreateSubnetRequest{
			Name:   "test-subnet",
			CIDR:   "10.0.1.0/24",
			Zone:   "eastus-1",
			Public: true,
		}
		subnet, err := network.CreateSubnet(ctx, vpcID, subnetReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(subnet)
		ts.Require().Equal("test-subnet", subnet.Name)
		ts.Require().Equal(vpcID, subnet.VPCID)
		ts.Require().Equal("10.0.1.0/24", subnet.CIDR)
		ts.Require().Equal("eastus-1", subnet.Zone)
		ts.Require().Equal("Succeeded", subnet.State)
		ts.Require().True(subnet.Public)
		ts.Require().Contains(subnet.ID, "subnet-test-subnet")

		// GetSubnet
		retrievedSubnet, err := network.GetSubnet(ctx, subnet.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedSubnet)
		ts.Require().Equal(subnet.ID, retrievedSubnet.ID)
		ts.Require().Equal("azure-subnet", retrievedSubnet.Name)

		// ListSubnets
		subnets, err := network.ListSubnets(ctx, vpcID)
		ts.Require().NoError(err)
		ts.Require().NotNil(subnets)
		ts.Require().Len(subnets, 1)
		ts.Require().Equal(vpcID, subnets[0].VPCID)

		// DeleteSubnet
		err = network.DeleteSubnet(ctx, subnet.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Security Group operations", func() {
		// CreateSecurityGroup
		sgReq := &providers.CreateSecurityGroupRequest{
			Name:        "test-nsg",
			Description: "Test network security group",
			VPCID:       "vnet-12345",
			Rules: []*providers.SecurityRule{
				{
					Direction:   "inbound",
					Protocol:    "tcp",
					FromPort:    80,
					ToPort:      80,
					Source:      "0.0.0.0/0",
					Description: "Allow HTTP",
				},
				{
					Direction:   "inbound",
					Protocol:    "tcp",
					FromPort:    443,
					ToPort:      443,
					Source:      "0.0.0.0/0",
					Description: "Allow HTTPS",
				},
			},
		}
		sg, err := network.CreateSecurityGroup(ctx, sgReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(sg)
		ts.Require().Equal("test-nsg", sg.Name)
		ts.Require().Equal("Test network security group", sg.Description)
		ts.Require().Equal("vnet-12345", sg.VPCID)
		ts.Require().Len(sg.Rules, 2)
		ts.Require().Contains(sg.ID, "nsg-test-nsg")

		// UpdateSecurityRules
		newRules := []*providers.SecurityRule{
			{
				Direction: "inbound",
				Protocol:  "tcp",
				FromPort:  22,
				ToPort:    22,
				Source:    "10.0.0.0/8",
			},
		}
		err = network.UpdateSecurityRules(ctx, sg.ID, newRules)
		ts.Require().NoError(err)
	})

	ts.Run("Load Balancer operations", func() {
		// CreateLoadBalancer
		lbReq := &providers.CreateLoadBalancerRequest{
			Name:    "test-lb",
			Type:    "application",
			Subnets: []string{"subnet-1", "subnet-2"},
			Listeners: []*providers.Listener{
				{
					Protocol:      "HTTP",
					Port:          80,
					TargetGroupID: "tg-web-servers",
				},
				{
					Protocol:      "HTTPS",
					Port:          443,
					TargetGroupID: "tg-web-servers",
					Certificate:   "cert-123",
				},
			},
		}
		lb, err := network.CreateLoadBalancer(ctx, lbReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(lb)
		ts.Require().Equal("test-lb", lb.Name)
		ts.Require().Equal("application", lb.Type)
		ts.Require().Equal("Succeeded", lb.State)
		ts.Require().Contains(lb.DNSName, "test-lb.eastus.cloudapp.azure.com")
		ts.Require().Contains(lb.ID, "lb-test-lb")
		ts.Require().Len(lb.Listeners, 2)

		// UpdateLoadBalancer
		lbName := "updated-lb"
		updateReq := &providers.UpdateLoadBalancerRequest{
			Name: &lbName,
		}
		err = network.UpdateLoadBalancer(ctx, lb.ID, updateReq)
		ts.Require().NoError(err)
	})
}

// TestAzureContainerService tests Azure container service operations
func (ts *AzureServicesTestSuite) TestAzureContainerService() {
	ctx := context.Background()
	container := ts.provider.Container()
	ts.Require().NotNil(container)

	ts.Run("Cluster operations", func() {
		// CreateCluster
		clusterReq := &providers.CreateClusterRequest{
			Name:      "test-aks",
			Type:      "kubernetes",
			Version:   "1.27.3",
			Region:    "eastus",
			NodeCount: 3,
			NodeType:  "Standard_D2s_v3",
		}
		cluster, err := container.CreateCluster(ctx, clusterReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(cluster)
		ts.Require().Equal("test-aks", cluster.Name)
		ts.Require().Equal("kubernetes", cluster.Type)
		ts.Require().Equal("1.27.3", cluster.Version)
		ts.Require().Equal("Creating", cluster.State)
		ts.Require().Equal("eastus", cluster.Region)
		ts.Require().Equal(3, cluster.NodeCount)
		ts.Require().Contains(cluster.ID, "aks-test-aks")

		// GetCluster
		retrievedCluster, err := container.GetCluster(ctx, cluster.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedCluster)
		ts.Require().Equal(cluster.ID, retrievedCluster.ID)
		ts.Require().Equal("aks-cluster", retrievedCluster.Name)
		ts.Require().Equal("Succeeded", retrievedCluster.State)

		// ListClusters
		clusters, err := container.ListClusters(ctx)
		ts.Require().NoError(err)
		ts.Require().NotNil(clusters)
		ts.Require().Len(clusters, 1)
		ts.Require().Equal("production", clusters[0].Name)

		// UpdateCluster
		nodeCount := 5
		updateReq := &providers.UpdateClusterRequest{NodeCount: &nodeCount}
		err = container.UpdateCluster(ctx, cluster.ID, updateReq)
		ts.Require().NoError(err)

		// DeleteCluster
		err = container.DeleteCluster(ctx, cluster.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Deployment operations", func() {
		clusterID := "aks-production"

		// DeployContainer
		deployReq := &providers.DeployRequest{
			Name:     "web-app",
			Image:    "nginx:latest",
			Replicas: 3,
			Environment: map[string]string{
				"ENV":  "production",
				"PORT": "8080",
			},
			Ports: []int{80, 443},
		}
		deployment, err := container.DeployContainer(ctx, clusterID, deployReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(deployment)
		ts.Require().Equal("web-app", deployment.Name)
		ts.Require().Equal(clusterID, deployment.ClusterID)
		ts.Require().Equal("nginx:latest", deployment.Image)
		ts.Require().Equal(3, deployment.Replicas)
		ts.Require().Equal("Running", deployment.State)
		ts.Require().Contains(deployment.ID, "deploy-web-app")

		// GetDeployment
		retrievedDeployment, err := container.GetDeployment(ctx, deployment.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedDeployment)
		ts.Require().Equal(deployment.ID, retrievedDeployment.ID)
		ts.Require().Equal("azure-deployment", retrievedDeployment.Name)

		// UpdateDeployment
		replicas := 5
		updateReq := &providers.UpdateDeploymentRequest{Replicas: &replicas}
		err = container.UpdateDeployment(ctx, deployment.ID, updateReq)
		ts.Require().NoError(err)

		// ScaleDeployment
		err = container.ScaleDeployment(ctx, deployment.ID, 10)
		ts.Require().NoError(err)

		// DeleteDeployment
		err = container.DeleteDeployment(ctx, deployment.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Service Mesh operations", func() {
		clusterID := "aks-production"

		// EnableServiceMesh
		meshConfig := &providers.ServiceMeshConfig{
			Type:    "istio",
			MTLS:    true,
			Tracing: true,
			Metrics: true,
		}
		err := container.EnableServiceMesh(ctx, clusterID, meshConfig)
		ts.Require().NoError(err)

		// ConfigureTrafficPolicy
		trafficPolicy := &providers.TrafficPolicy{
			Name:    "canary-policy",
			Service: "web-service",
			Rules: []*providers.TrafficRule{
				{Weight: 90, Version: "v1"},
				{Weight: 10, Version: "v2"},
			},
		}
		err = container.ConfigureTrafficPolicy(ctx, trafficPolicy)
		ts.Require().NoError(err)
	})
}

// TestAzureDatabaseService tests Azure database service operations
func (ts *AzureServicesTestSuite) TestAzureDatabaseService() {
	ctx := context.Background()
	database := ts.provider.Database()
	ts.Require().NotNil(database)

	ts.Run("Database operations", func() {
		// CreateDatabase
		dbReq := &providers.CreateDatabaseRequest{
			Name:     "test-sqldb",
			Engine:   "sqlserver",
			Version:  "2019",
			Size:     "Standard_S2",
			Storage:  250,
			Username: "dbadmin",
			Password: "SecurePassword123!",
			MultiAZ:  true,
			Backup:   true,
		}
		db, err := database.CreateDatabase(ctx, dbReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(db)
		ts.Require().Equal("test-sqldb", db.Name)
		ts.Require().Equal("sqlserver", db.Engine)
		ts.Require().Equal("2019", db.Version)
		ts.Require().Equal("Creating", db.State)
		ts.Require().Equal("Standard_S2", db.Size)
		ts.Require().Equal(250, db.Storage)
		ts.Require().True(db.MultiAZ)
		ts.Require().Contains(db.ID, "sql-test-sqldb")

		// GetDatabase
		retrievedDB, err := database.GetDatabase(ctx, db.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedDB)
		ts.Require().Equal(db.ID, retrievedDB.ID)
		ts.Require().Equal("azure-sql", retrievedDB.Name)
		ts.Require().Equal("Online", retrievedDB.State)

		// ListDatabases
		databases, err := database.ListDatabases(ctx)
		ts.Require().NoError(err)
		ts.Require().NotNil(databases)
		ts.Require().Len(databases, 1)
		ts.Require().Equal("production", databases[0].Name)

		// UpdateDatabase
		size := "Standard_S4"
		updateReq := &providers.UpdateDatabaseRequest{Size: &size}
		err = database.UpdateDatabase(ctx, db.ID, updateReq)
		ts.Require().NoError(err)

		// ScaleDatabase
		scaleReq := &providers.ScaleRequest{
			Size:    "Standard_S4",
			Storage: 500,
			IOPS:    2000,
		}
		err = database.ScaleDatabase(ctx, db.ID, scaleReq)
		ts.Require().NoError(err)

		// EnableReadReplica
		replica, err := database.EnableReadReplica(ctx, db.ID, "westus2")
		ts.Require().NoError(err)
		ts.Require().NotNil(replica)
		ts.Require().Contains(replica.ID, "replica-")
		ts.Require().Equal("replica", replica.Name)

		// EnableMultiMaster
		regions := []string{"eastus", "westus2", "centralus"}
		err = database.EnableMultiMaster(ctx, db.ID, regions)
		ts.Require().NoError(err)

		// DeleteDatabase
		err = database.DeleteDatabase(ctx, db.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Backup operations", func() {
		dbID := "sql-production-db"

		// CreateBackup
		backup, err := database.CreateBackup(ctx, dbID, "manual-backup-test")
		ts.Require().NoError(err)
		ts.Require().NotNil(backup)
		ts.Require().Equal("manual-backup-test", backup.Name)
		ts.Require().Equal(dbID, backup.DatabaseID)
		ts.Require().Equal("InProgress", backup.State)
		ts.Require().Contains(backup.ID, "backup-manual-backup-test")

		// ListBackups
		backups, err := database.ListBackups(ctx, dbID)
		ts.Require().NoError(err)
		ts.Require().NotNil(backups)
		ts.Require().Len(backups, 1)
		ts.Require().Equal(dbID, backups[0].DatabaseID)

		// RestoreBackup
		err = database.RestoreBackup(ctx, backup.ID, "restored-db")
		ts.Require().NoError(err)
	})
}

// TestAzureSecurityService tests Azure security service operations
func (ts *AzureServicesTestSuite) TestAzureSecurityService() {
	ctx := context.Background()
	security := ts.provider.Security()
	ts.Require().NotNil(security)

	ts.Run("Role and Policy operations", func() {
		// CreateRole
		roleReq := &providers.CreateRoleRequest{
			Name:        "test-role",
			Description: "Test role for unit testing",
			Policies:    []string{"policy-1", "policy-2"},
		}
		role, err := security.CreateRole(ctx, roleReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(role)
		ts.Require().Equal("test-role", role.Name)
		ts.Require().Equal("role-123", role.ID)

		// CreatePolicy
		policyReq := &providers.CreatePolicyRequest{
			Name:        "test-policy",
			Description: "Test policy for unit testing",
			Document:    `{"Version": "1.0", "Statement": [{"Effect": "Allow", "Action": "storage:read"}]}`,
		}
		policy, err := security.CreatePolicy(ctx, policyReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(policy)
		ts.Require().Equal("test-policy", policy.Name)
		ts.Require().Equal("policy-123", policy.ID)

		// AttachPolicy
		err = security.AttachPolicy(ctx, role.ID, policy.ID)
		ts.Require().NoError(err)
	})

	ts.Run("Secret operations", func() {
		// CreateSecret
		secretReq := &providers.CreateSecretRequest{
			Name:         "test-secret",
			Description:  "Test secret for unit testing",
			Value:        "super-secret-value",
			AutoRotate:   true,
			RotationDays: 30,
			Tags: map[string]string{
				"Environment": "test",
				"Purpose":     "unit-testing",
			},
		}
		secret, err := security.CreateSecret(ctx, secretReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(secret)
		ts.Require().Equal("test-secret", secret.Name)
		ts.Require().Equal("secret-123", secret.ID)

		// GetSecret
		retrievedSecret, err := security.GetSecret(ctx, secret.ID)
		ts.Require().NoError(err)
		ts.Require().NotNil(retrievedSecret)
		ts.Require().Equal(secret.ID, retrievedSecret.ID)

		// RotateSecret
		err = security.RotateSecret(ctx, secret.ID)
		ts.Require().NoError(err)
	})

	ts.Run("KMS operations", func() {
		// CreateKMSKey
		keyReq := &providers.CreateKeyRequest{
			Name:       "test-key",
			Algorithm:  "RSA-2048",
			Usage:      "encrypt",
			AutoRotate: true,
		}
		key, err := security.CreateKMSKey(ctx, keyReq)
		ts.Require().NoError(err)
		ts.Require().NotNil(key)
		ts.Require().Equal("test-key", key.Name)
		ts.Require().Equal("key-123", key.ID)

		// Encrypt and Decrypt
		originalData := []byte("sensitive information")
		encryptedData, err := security.Encrypt(ctx, key.ID, originalData)
		ts.Require().NoError(err)
		ts.Require().Equal(originalData, encryptedData) // Mock just returns original data

		decryptedData, err := security.Decrypt(ctx, key.ID, encryptedData)
		ts.Require().NoError(err)
		ts.Require().Equal(originalData, decryptedData)
	})

	ts.Run("Compliance operations", func() {
		// EnableAuditLogging
		auditConfig := &providers.AuditConfig{
			Enabled:  true,
			LogGroup: "audit-logs",
		}
		err := security.EnableAuditLogging(ctx, auditConfig)
		ts.Require().NoError(err)

		// GetComplianceReport
		report, err := security.GetComplianceReport(ctx, "ISO27001")
		ts.Require().NoError(err)
		ts.Require().NotNil(report)
		ts.Require().Equal("ISO27001", report.Standard)
	})
}

// TestAzureServicesTestSuite runs the Azure services test suite
func TestAzureServicesTestSuite(t *testing.T) {
	suite.Run(t, new(AzureServicesTestSuite))
}
