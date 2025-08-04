package providers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TypesTestSuite tests all the type definitions and data structures
type TypesTestSuite struct {
	suite.Suite
}

// TestInstanceTypes tests Instance and related types
func (ts *TypesTestSuite) TestInstanceTypes() {
	ts.Run("Instance structure validation", func() {
		now := time.Now()
		instance := &Instance{
			ID:        "i-1234567890abcdef0",
			Name:      "web-server-01",
			Type:      "t3.medium",
			State:     "running",
			Region:    "us-east-1",
			Zone:      "us-east-1a",
			PublicIP:  "54.123.45.67",
			PrivateIP: "10.0.1.100",
			CreatedAt: now,
			Tags: map[string]string{
				"Environment": "production",
				"Project":     "web-app",
				"Owner":       "devops-team",
			},
			Metadata: map[string]interface{}{
				"monitoring":   true,
				"backup":       true,
				"cost_center":  "engineering",
				"instance_age": 30,
			},
		}

		// Validate all fields are set correctly
		ts.Require().Equal("i-1234567890abcdef0", instance.ID)
		ts.Require().Equal("web-server-01", instance.Name)
		ts.Require().Equal("t3.medium", instance.Type)
		ts.Require().Equal("running", instance.State)
		ts.Require().Equal("us-east-1", instance.Region)
		ts.Require().Equal("us-east-1a", instance.Zone)
		ts.Require().Equal("54.123.45.67", instance.PublicIP)
		ts.Require().Equal("10.0.1.100", instance.PrivateIP)
		ts.Require().Equal(now, instance.CreatedAt)
		ts.Require().Len(instance.Tags, 3)
		ts.Require().Equal("production", instance.Tags["Environment"])
		ts.Require().Len(instance.Metadata, 4)
		monitoring, ok := instance.Metadata["monitoring"].(bool)
		ts.Require().True(ok, "monitoring metadata should be a bool")
		ts.Require().True(monitoring)
		instanceAge, ok := instance.Metadata["instance_age"].(int)
		ts.Require().True(ok, "instance_age metadata should be an int")
		ts.Require().Equal(30, instanceAge)
	})

	ts.Run("CreateInstanceRequest validation", func() {
		req := &CreateInstanceRequest{
			Name:           "test-instance",
			Type:           "t3.micro",
			Image:          "ami-0abcdef1234567890",
			Region:         "us-west-2",
			Zone:           "us-west-2a",
			KeyPair:        "my-key-pair",
			SecurityGroups: []string{"sg-123", "sg-456"},
			UserData:       "#!/bin/bash\necho 'Hello World'",
			Tags: map[string]string{
				"Purpose": "testing",
				"Auto":    "true",
			},
			DiskSize: 20,
			NetworkConfig: &NetworkConfig{
				VPCID:    "vpc-123456",
				SubnetID: "subnet-789012",
				PublicIP: true,
				IPv6:     false,
			},
		}

		ts.Require().Equal("test-instance", req.Name)
		ts.Require().Equal("t3.micro", req.Type)
		ts.Require().Equal("ami-0abcdef1234567890", req.Image)
		ts.Require().Len(req.SecurityGroups, 2)
		ts.Require().Contains(req.UserData, "Hello World")
		ts.Require().Equal(20, req.DiskSize)
		ts.Require().NotNil(req.NetworkConfig)
		ts.Require().True(req.NetworkConfig.PublicIP)
		ts.Require().False(req.NetworkConfig.IPv6)
	})

	ts.Run("InstanceFilter validation", func() {
		filter := &InstanceFilter{
			States:  []string{"running", "stopped"},
			Tags:    map[string]string{"Environment": "production"},
			Regions: []string{"us-east-1", "us-west-2"},
			Types:   []string{"t3.micro", "t3.small", "t3.medium"},
		}

		ts.Require().Len(filter.States, 2)
		ts.Require().Contains(filter.States, "running")
		ts.Require().Contains(filter.States, "stopped")
		ts.Require().Equal("production", filter.Tags["Environment"])
		ts.Require().Len(filter.Regions, 2)
		ts.Require().Len(filter.Types, 3)
	})
}

// TestStorageTypes tests storage-related types
func (ts *TypesTestSuite) TestStorageTypes() {
	ts.Run("Bucket and Object structures", func() {
		now := time.Now()
		bucket := &Bucket{
			Name:         "my-test-bucket",
			Region:       "us-east-1",
			CreatedAt:    now,
			Versioning:   true,
			Encryption:   true,
			PublicAccess: false,
			Tags: map[string]string{
				"Project":     "data-lake",
				"Environment": "staging",
			},
		}

		object := &Object{
			Key:          "documents/report.pdf",
			Size:         1048576, // 1MB
			LastModified: now,
			ETag:         "d41d8cd98f00b204e9800998ecf8427e",
			StorageClass: "STANDARD",
			ContentType:  "application/pdf",
		}

		ts.Require().Equal("my-test-bucket", bucket.Name)
		ts.Require().True(bucket.Versioning)
		ts.Require().True(bucket.Encryption)
		ts.Require().False(bucket.PublicAccess)
		ts.Require().Equal("data-lake", bucket.Tags["Project"])

		ts.Require().Equal("documents/report.pdf", object.Key)
		ts.Require().Equal(int64(1048576), object.Size)
		ts.Require().Equal("application/pdf", object.ContentType)
		ts.Require().Equal("STANDARD", object.StorageClass)
	})

	ts.Run("PutOptions and ACL structures", func() {
		opts := &PutOptions{
			ContentType:  "image/jpeg",
			CacheControl: "max-age=3600",
			Metadata: map[string]string{
				"photographer": "john-doe",
				"location":     "paris",
			},
			ACL:        "private",
			Encryption: "AES256",
		}

		acl := &ACL{
			Owner: "bucket-owner-id",
			Grants: []*Grant{
				{
					Grantee:    "user-id-1",
					Permission: "READ",
				},
				{
					Grantee:    "user-id-2",
					Permission: "WRITE",
				},
			},
		}

		ts.Require().Equal("image/jpeg", opts.ContentType)
		ts.Require().Equal("max-age=3600", opts.CacheControl)
		ts.Require().Equal("john-doe", opts.Metadata["photographer"])
		ts.Require().Equal("private", opts.ACL)

		ts.Require().Equal("bucket-owner-id", acl.Owner)
		ts.Require().Len(acl.Grants, 2)
		ts.Require().Equal("READ", acl.Grants[0].Permission)
		ts.Require().Equal("WRITE", acl.Grants[1].Permission)
	})
}

// TestNetworkTypes tests network-related types
func (ts *TypesTestSuite) TestNetworkTypes() {
	ts.Run("VPC and Subnet structures", func() {
		vpc := &VPC{
			ID:     "vpc-0123456789abcdef0",
			Name:   "main-vpc",
			CIDR:   "10.0.0.0/16",
			Region: "us-east-1",
			State:  "available",
			Tags: map[string]string{
				"Purpose": "production",
				"Team":    "infrastructure",
			},
		}

		subnet := &Subnet{
			ID:     "subnet-0123456789abcdef0",
			Name:   "public-subnet-1a",
			VPCID:  vpc.ID,
			CIDR:   "10.0.1.0/24",
			Zone:   "us-east-1a",
			State:  "available",
			Public: true,
		}

		ts.Require().Equal("vpc-0123456789abcdef0", vpc.ID)
		ts.Require().Equal("10.0.0.0/16", vpc.CIDR)
		ts.Require().Equal("available", vpc.State)
		ts.Require().Equal("infrastructure", vpc.Tags["Team"])

		ts.Require().Equal(vpc.ID, subnet.VPCID)
		ts.Require().Equal("10.0.1.0/24", subnet.CIDR)
		ts.Require().True(subnet.Public)
	})

	ts.Run("SecurityGroup and LoadBalancer structures", func() {
		sg := &SecurityGroup{
			ID:          "sg-0123456789abcdef0",
			Name:        "web-servers-sg",
			Description: "Security group for web servers",
			VPCID:       "vpc-0123456789abcdef0",
			Rules: []*SecurityRule{
				{
					Direction:   "ingress",
					Protocol:    "tcp",
					FromPort:    80,
					ToPort:      80,
					Source:      "0.0.0.0/0",
					Description: "HTTP access from anywhere",
				},
				{
					Direction:   "ingress",
					Protocol:    "tcp",
					FromPort:    443,
					ToPort:      443,
					Source:      "0.0.0.0/0",
					Description: "HTTPS access from anywhere",
				},
			},
		}

		lb := &LoadBalancer{
			ID:      "lb-0123456789abcdef0",
			Name:    "web-lb",
			Type:    "application",
			State:   "active",
			DNSName: "web-lb-123456789.us-east-1.elb.amazonaws.com",
			Listeners: []*Listener{
				{
					Protocol:      "HTTP",
					Port:          80,
					TargetGroupID: "tg-123456789",
				},
				{
					Protocol:      "HTTPS",
					Port:          443,
					TargetGroupID: "tg-123456789",
					Certificate:   "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012",
				},
			},
			TargetGroups: []*TargetGroup{
				{
					ID:       "tg-123456789",
					Name:     "web-servers",
					Protocol: "HTTP",
					Port:     80,
					HealthCheck: &HealthCheck{
						Protocol:           "HTTP",
						Port:               80,
						Path:               "/health",
						Interval:           30 * time.Second,
						Timeout:            5 * time.Second,
						HealthyThreshold:   2,
						UnhealthyThreshold: 3,
					},
					Targets: []string{"i-123", "i-456", "i-789"},
				},
			},
		}

		ts.Require().Equal("web-servers-sg", sg.Name)
		ts.Require().Len(sg.Rules, 2)
		ts.Require().Equal("ingress", sg.Rules[0].Direction)
		ts.Require().Equal(80, sg.Rules[0].FromPort)
		ts.Require().Equal(443, sg.Rules[1].FromPort)

		ts.Require().Equal("application", lb.Type)
		ts.Require().Contains(lb.DNSName, "elb.amazonaws.com")
		ts.Require().Len(lb.Listeners, 2)
		ts.Require().Len(lb.TargetGroups, 1)
		ts.Require().Equal("/health", lb.TargetGroups[0].HealthCheck.Path)
		ts.Require().Len(lb.TargetGroups[0].Targets, 3)
	})
}

// TestContainerTypes tests container-related types
func (ts *TypesTestSuite) TestContainerTypes() {
	ts.Run("Cluster and Deployment structures", func() {
		now := time.Now()
		cluster := &Cluster{
			ID:        "cluster-123456789",
			Name:      "production-k8s",
			Type:      "kubernetes",
			Version:   "1.27.3",
			State:     "ACTIVE",
			Region:    "us-west-2",
			NodeCount: 5,
			Endpoint:  "https://ABC123DEF456.gr7.us-west-2.eks.amazonaws.com",
			CreatedAt: now,
		}

		deployment := &Deployment{
			ID:        "deploy-987654321",
			Name:      "web-application",
			ClusterID: cluster.ID,
			Image:     "nginx:1.21-alpine",
			Replicas:  3,
			State:     "RUNNING",
			CreatedAt: now,
			UpdatedAt: now.Add(1 * time.Hour),
		}

		ts.Require().Equal("production-k8s", cluster.Name)
		ts.Require().Equal("kubernetes", cluster.Type)
		ts.Require().Equal("1.27.3", cluster.Version)
		ts.Require().Equal(5, cluster.NodeCount)
		ts.Require().Contains(cluster.Endpoint, "eks.amazonaws.com")

		ts.Require().Equal(cluster.ID, deployment.ClusterID)
		ts.Require().Equal("nginx:1.21-alpine", deployment.Image)
		ts.Require().Equal(3, deployment.Replicas)
		ts.Require().Equal("RUNNING", deployment.State)
	})

	ts.Run("ServiceMesh and TrafficPolicy structures", func() {
		meshConfig := &ServiceMeshConfig{
			Type:    "istio",
			MTLS:    true,
			Tracing: true,
			Metrics: true,
		}

		trafficPolicy := &TrafficPolicy{
			Name:    "canary-deployment",
			Service: "web-service",
			Rules: []*TrafficRule{
				{
					Weight:  90,
					Version: "v1.0",
					Headers: map[string]string{
						"user-type": "regular",
					},
				},
				{
					Weight:  10,
					Version: "v1.1",
					Headers: map[string]string{
						"user-type": "beta",
					},
				},
			},
		}

		ts.Require().Equal("istio", meshConfig.Type)
		ts.Require().True(meshConfig.MTLS)
		ts.Require().True(meshConfig.Tracing)

		ts.Require().Equal("canary-deployment", trafficPolicy.Name)
		ts.Require().Equal("web-service", trafficPolicy.Service)
		ts.Require().Len(trafficPolicy.Rules, 2)
		ts.Require().Equal(90, trafficPolicy.Rules[0].Weight)
		ts.Require().Equal("v1.1", trafficPolicy.Rules[1].Version)
	})
}

// TestDatabaseTypes tests database-related types
func (ts *TypesTestSuite) TestDatabaseTypes() {
	ts.Run("Database and Backup structures", func() {
		now := time.Now()
		database := &Database{
			ID:        "db-123456789abcdef0",
			Name:      "production-postgres",
			Engine:    "postgres",
			Version:   "14.8",
			State:     "available",
			Endpoint:  "prod-pg.cluster-abc123.us-east-1.rds.amazonaws.com",
			Port:      5432,
			Size:      "db.r5.xlarge",
			Storage:   500,
			MultiAZ:   true,
			CreatedAt: now,
		}

		backup := &Backup{
			ID:         "backup-987654321fedcba0",
			DatabaseID: database.ID,
			Name:       "daily-backup-20231201",
			State:      "available",
			Size:       10737418240, // 10GB
			CreatedAt:  now,
			ExpiresAt:  now.AddDate(0, 0, 30), // 30 days retention
		}

		ts.Require().Equal("production-postgres", database.Name)
		ts.Require().Equal("postgres", database.Engine)
		ts.Require().Equal("14.8", database.Version)
		ts.Require().Equal(5432, database.Port)
		ts.Require().True(database.MultiAZ)
		ts.Require().Equal(500, database.Storage)

		ts.Require().Equal(database.ID, backup.DatabaseID)
		ts.Require().Contains(backup.Name, "daily-backup")
		ts.Require().Equal(int64(10737418240), backup.Size)
		ts.Require().Equal("available", backup.State)
	})

	ts.Run("ScaleRequest validation", func() {
		scaleReq := &ScaleRequest{
			Size:    "db.r5.2xlarge",
			IOPS:    3000,
			Storage: 1000,
		}

		ts.Require().Equal("db.r5.2xlarge", scaleReq.Size)
		ts.Require().Equal(3000, scaleReq.IOPS)
		ts.Require().Equal(1000, scaleReq.Storage)
	})
}

// TestSecurityTypes tests security-related types
func (ts *TypesTestSuite) TestSecurityTypes() {
	ts.Run("Role and Policy structures", func() {
		now := time.Now()
		role := &Role{
			ID:          "role-123456789",
			Name:        "EC2InstanceRole",
			Description: "Role for EC2 instances to access S3",
			Policies:    []string{"policy-s3-read", "policy-cloudwatch-logs"},
			TrustPolicy: `{"Version": "2012-10-17", "Statement": []}`,
			CreatedAt:   now,
		}

		policy := &Policy{
			ID:          "policy-987654321",
			Name:        "S3ReadOnlyAccess",
			Description: "Allows read-only access to S3 buckets",
			Document:    `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Action": "s3:GetObject", "Resource": "*"}]}`,
			Version:     "v1",
			CreatedAt:   now,
		}

		ts.Require().Equal("EC2InstanceRole", role.Name)
		ts.Require().Contains(role.Description, "EC2 instances")
		ts.Require().Len(role.Policies, 2)
		ts.Require().Contains(role.TrustPolicy, "2012-10-17")

		ts.Require().Equal("S3ReadOnlyAccess", policy.Name)
		ts.Require().Contains(policy.Document, "s3:GetObject")
		ts.Require().Equal("v1", policy.Version)
	})

	ts.Run("Secret and KMSKey structures", func() {
		now := time.Now()
		secret := &Secret{
			ID:           "secret-abcdef123456789",
			Name:         "database-password",
			Description:  "Password for production database",
			Version:      "v1",
			CreatedAt:    now,
			UpdatedAt:    now.Add(1 * time.Hour),
			NextRotation: now.AddDate(0, 0, 30), // 30 days
		}

		kmsKey := &KMSKey{
			ID:           "key-fedcba987654321",
			Name:         "application-encryption-key",
			Algorithm:    "SYMMETRIC_DEFAULT",
			State:        "Enabled",
			CreatedAt:    now,
			NextRotation: now.AddDate(1, 0, 0), // 1 year
		}

		ts.Require().Equal("database-password", secret.Name)
		ts.Require().Equal("v1", secret.Version)
		ts.Require().Contains(secret.Description, "production database")

		ts.Require().Equal("application-encryption-key", kmsKey.Name)
		ts.Require().Equal("SYMMETRIC_DEFAULT", kmsKey.Algorithm)
		ts.Require().Equal("Enabled", kmsKey.State)
	})
}

// TestMonitoringTypes tests monitoring-related types
func (ts *TypesTestSuite) TestMonitoringTypes() {
	ts.Run("Metric and Alert structures", func() {
		now := time.Now()
		metric := &Metric{
			Name:      "cpu.utilization",
			Namespace: "AWS/EC2",
			Value:     75.5,
			Unit:      "Percent",
			Timestamp: now,
			Dimensions: map[string]string{
				"InstanceId":   "i-1234567890abcdef0",
				"InstanceType": "t3.medium",
			},
		}

		alert := &Alert{
			ID:          "alert-123456789",
			Name:        "high-cpu-utilization",
			Description: "Alert when CPU utilization exceeds 80%",
			Condition: &AlertCondition{
				MetricName:        "cpu.utilization",
				Namespace:         "AWS/EC2",
				Statistic:         "Average",
				Threshold:         80.0,
				Comparison:        "GreaterThanThreshold",
				Period:            5 * time.Minute,
				EvaluationPeriods: 2,
			},
			Actions: []*AlertAction{
				{
					Type:   "email",
					Target: "devops@example.com",
					Properties: map[string]string{
						"subject": "High CPU Alert",
					},
				},
			},
			State:     "OK",
			Enabled:   true,
			CreatedAt: now,
		}

		ts.Require().Equal("cpu.utilization", metric.Name)
		ts.Require().InEpsilon(75.5, metric.Value, 0.001)
		ts.Require().Equal("Percent", metric.Unit)
		ts.Require().Equal("i-1234567890abcdef0", metric.Dimensions["InstanceId"])

		ts.Require().Equal("high-cpu-utilization", alert.Name)
		ts.Require().InEpsilon(80.0, alert.Condition.Threshold, 0.001)
		ts.Require().Equal("GreaterThanThreshold", alert.Condition.Comparison)
		ts.Require().Len(alert.Actions, 1)
		ts.Require().Equal("email", alert.Actions[0].Type)
		ts.Require().True(alert.Enabled)
	})

	ts.Run("Dashboard and Trace structures", func() {
		now := time.Now()
		dashboard := &Dashboard{
			ID:          "dash-123456789",
			Name:        "Application Metrics",
			Description: "Dashboard for application performance metrics",
			Widgets: []*Widget{
				{
					ID:    "widget-1",
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
			},
			CreatedAt: now,
			UpdatedAt: now.Add(2 * time.Hour),
		}

		trace := &Trace{
			ID:            "trace-abc123def456",
			ServiceName:   "user-service",
			OperationName: "getUserProfile",
			StartTime:     now,
			Duration:      150 * time.Millisecond,
			Status:        "OK",
			Spans: []*Span{
				{
					ID:            "span-123",
					OperationName: "database.query",
					ServiceName:   "postgres",
					StartTime:     now.Add(10 * time.Millisecond),
					Duration:      50 * time.Millisecond,
					Status:        "OK",
					Tags: map[string]string{
						"db.statement": "SELECT * FROM users WHERE id = $1",
						"db.type":      "postgresql",
					},
				},
			},
		}

		ts.Require().Equal("Application Metrics", dashboard.Name)
		ts.Require().Len(dashboard.Widgets, 1)
		ts.Require().Equal("line-chart", dashboard.Widgets[0].Type)

		ts.Require().Equal("user-service", trace.ServiceName)
		ts.Require().Equal("getUserProfile", trace.OperationName)
		ts.Require().Equal(150*time.Millisecond, trace.Duration)
		ts.Require().Len(trace.Spans, 1)
		ts.Require().Equal("database.query", trace.Spans[0].OperationName)
	})
}

// TestAITypes tests AI/ML-related types
func (ts *TypesTestSuite) TestAITypes() {
	ts.Run("AIModel and Dataset structures", func() {
		now := time.Now()
		model := &AIModel{
			ID:        "model-abc123def456",
			Name:      "image-classifier",
			Type:      "classification",
			Framework: "tensorflow",
			Version:   "v1.0",
			State:     "trained",
			CreatedAt: now,
			UpdatedAt: now.Add(6 * time.Hour),
		}

		dataset := &Dataset{
			ID:        "dataset-fed456cba123",
			Name:      "training-images",
			Type:      "image",
			Size:      5368709120, // 5GB
			Location:  "s3://ml-bucket/datasets/images/",
			Format:    "jpeg",
			CreatedAt: now,
		}

		ts.Require().Equal("image-classifier", model.Name)
		ts.Require().Equal("classification", model.Type)
		ts.Require().Equal("tensorflow", model.Framework)
		ts.Require().Equal("trained", model.State)

		ts.Require().Equal("training-images", dataset.Name)
		ts.Require().Equal("image", dataset.Type)
		ts.Require().Equal(int64(5368709120), dataset.Size)
		ts.Require().Contains(dataset.Location, "s3://")
	})

	ts.Run("TrainingJob and NetworkArchitecture structures", func() {
		now := time.Now()
		trainingJob := &TrainingJob{
			ID:        "job-789012345abc",
			ModelID:   "model-abc123def456",
			DatasetID: "dataset-fed456cba123",
			Status:    "completed",
			StartTime: now,
			EndTime:   now.Add(2 * time.Hour),
			Metrics: map[string]float64{
				"accuracy":  0.95,
				"precision": 0.92,
				"recall":    0.89,
				"f1_score":  0.905,
			},
		}

		architecture := &NetworkArchitecture{
			Layers: []*Layer{
				{
					Type:       "dense",
					Units:      128,
					Activation: "relu",
					Config: map[string]interface{}{
						"dropout": 0.2,
					},
				},
				{
					Type:       "dense",
					Units:      64,
					Activation: "relu",
				},
				{
					Type:       "dense",
					Units:      10,
					Activation: "softmax",
				},
			},
			Optimizer:    "adam",
			LossFunction: "categorical_crossentropy",
			Metrics:      []string{"accuracy", "precision"},
		}

		ts.Require().Equal("completed", trainingJob.Status)
		ts.Require().InEpsilon(0.95, trainingJob.Metrics["accuracy"], 0.001)
		ts.Require().InEpsilon(0.905, trainingJob.Metrics["f1_score"], 0.001)
		ts.Require().Equal(2*time.Hour, trainingJob.EndTime.Sub(trainingJob.StartTime))

		ts.Require().Len(architecture.Layers, 3)
		ts.Require().Equal("dense", architecture.Layers[0].Type)
		ts.Require().Equal(128, architecture.Layers[0].Units)
		ts.Require().Equal("adam", architecture.Optimizer)
		ts.Require().Equal("categorical_crossentropy", architecture.LossFunction)
	})
}

// TestJSONSerialization tests JSON serialization/deserialization
func (ts *TypesTestSuite) TestJSONSerialization() {
	ts.Run("Instance JSON serialization", func() {
		original := &Instance{
			ID:        "i-1234567890abcdef0",
			Name:      "test-instance",
			Type:      "t3.medium",
			State:     "running",
			Region:    "us-east-1",
			Zone:      "us-east-1a",
			PublicIP:  "54.123.45.67",
			PrivateIP: "10.0.1.100",
			CreatedAt: time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC),
			Tags:      map[string]string{"env": "test"},
			Metadata:  map[string]interface{}{"monitoring": true},
		}

		// Serialize to JSON
		jsonData, err := json.Marshal(original)
		ts.Require().NoError(err)
		ts.Require().Contains(string(jsonData), "i-1234567890abcdef0")

		// Deserialize back
		var deserialized Instance
		err = json.Unmarshal(jsonData, &deserialized)
		ts.Require().NoError(err)

		// Verify all fields match
		ts.Require().Equal(original.ID, deserialized.ID)
		ts.Require().Equal(original.Name, deserialized.Name)
		ts.Require().Equal(original.Type, deserialized.Type)
		ts.Require().Equal(original.State, deserialized.State)
		ts.Require().Equal(original.Tags["env"], deserialized.Tags["env"])
		ts.Require().Equal(original.Metadata["monitoring"], deserialized.Metadata["monitoring"])
	})

	ts.Run("Complex nested structure JSON serialization", func() {
		original := &LoadBalancer{
			ID:      "lb-123",
			Name:    "test-lb",
			Type:    "application",
			State:   "active",
			DNSName: "test-lb.example.com",
			Listeners: []*Listener{
				{
					Protocol:      "HTTPS",
					Port:          443,
					TargetGroupID: "tg-123",
					Certificate:   "cert-arn",
				},
			},
			TargetGroups: []*TargetGroup{
				{
					ID:       "tg-123",
					Name:     "web-servers",
					Protocol: "HTTP",
					Port:     80,
					HealthCheck: &HealthCheck{
						Protocol:           "HTTP",
						Port:               80,
						Path:               "/health",
						Interval:           30 * time.Second,
						Timeout:            5 * time.Second,
						HealthyThreshold:   2,
						UnhealthyThreshold: 3,
					},
					Targets: []string{"i-123", "i-456"},
				},
			},
		}

		// Serialize and deserialize
		jsonData, err := json.Marshal(original)
		ts.Require().NoError(err)

		var deserialized LoadBalancer
		err = json.Unmarshal(jsonData, &deserialized)
		ts.Require().NoError(err)

		// Verify nested structures
		ts.Require().Equal(original.Name, deserialized.Name)
		ts.Require().Len(deserialized.Listeners, 1)
		ts.Require().Equal(original.Listeners[0].Protocol, deserialized.Listeners[0].Protocol)
		ts.Require().Len(deserialized.TargetGroups, 1)
		ts.Require().NotNil(deserialized.TargetGroups[0].HealthCheck)
		ts.Require().Equal(original.TargetGroups[0].HealthCheck.Path, deserialized.TargetGroups[0].HealthCheck.Path)
	})
}

// TestTypesTestSuite runs the types test suite
func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}
