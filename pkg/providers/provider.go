// Package providers defines interfaces and implementations for cloud/platform providers
package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// Static errors for err113 compliance
var (
	ErrProviderNotFound = errors.New("provider not found")
)

// Provider defines the interface for cloud/platform providers
type Provider interface {
	// Core provider methods
	Name() string                            // Provider name (e.g., "aws", "azure", "gcp")
	Initialize(config *ProviderConfig) error // Initialize provider with configuration
	Validate() error                         // Validate provider configuration
	Health() (*HealthStatus, error)          // Check provider health
	Close() error                            // Cleanup provider resources

	// Service interfaces
	Compute() ComputeService       // Compute/VM services
	Storage() StorageService       // Storage services
	Network() NetworkService       // Network services
	Container() ContainerService   // Container services
	Database() DatabaseService     // Database services
	Security() SecurityService     // Security services
	Monitoring() MonitoringService // Monitoring services
	Serverless() ServerlessService // Serverless services
	AI() AIService                 // AI/ML services

	// Advanced features
	Cost() CostService                 // Cost management
	Compliance() ComplianceService     // Compliance services
	Disaster() DisasterRecoveryService // Disaster recovery
	Edge() EdgeService                 // Edge computing
	Quantum() QuantumService           // Quantum computing
}

// ProviderConfig holds provider configuration
type ProviderConfig struct {
	// Authentication
	Credentials Credentials `json:"credentials"`
	Region      string      `json:"region"`
	Endpoint    string      `json:"endpoint,omitempty"`

	// Options
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	EnableCache bool          `json:"enable_cache"`

	// Advanced
	CustomHeaders map[string]string `json:"custom_headers,omitempty"`
	ProxyURL      string            `json:"proxy_url,omitempty"`
	TLSConfig     *TLSConfig        `json:"tls_config,omitempty"`
}

// Credentials holds authentication credentials
type Credentials struct {
	Type      string            `json:"type"` // "key", "token", "oauth", "cert"
	AccessKey string            `json:"access_key,omitempty"`
	SecretKey string            `json:"secret_key,omitempty"`
	Token     string            `json:"token,omitempty"`
	CertPath  string            `json:"cert_path,omitempty"`
	KeyPath   string            `json:"key_path,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CAPath             string `json:"ca_path,omitempty"`
	CertPath           string `json:"cert_path,omitempty"`
	KeyPath            string `json:"key_path,omitempty"`
	MinVersion         string `json:"min_version,omitempty"`
}

// HealthStatus represents provider health
type HealthStatus struct {
	Healthy     bool                     `json:"healthy"`
	Status      string                   `json:"status"`
	Services    map[string]ServiceHealth `json:"services"`
	LastChecked time.Time                `json:"last_checked"`
	Latency     time.Duration            `json:"latency"`
}

// ServiceHealth represents individual service health
type ServiceHealth struct {
	Available    bool          `json:"available"`
	ResponseTime time.Duration `json:"response_time"`
	Error        string        `json:"error,omitempty"`
}

// ComputeService defines compute/VM operations
type ComputeService interface {
	// Instance operations
	CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error)
	GetInstance(ctx context.Context, id string) (*Instance, error)
	ListInstances(ctx context.Context, filter *InstanceFilter) ([]*Instance, error)
	UpdateInstance(ctx context.Context, id string, req *UpdateInstanceRequest) error
	DeleteInstance(ctx context.Context, id string) error

	// Instance lifecycle
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string) error
	RestartInstance(ctx context.Context, id string) error

	// Advanced operations
	ResizeInstance(ctx context.Context, id string, size string) error
	SnapshotInstance(ctx context.Context, id string, name string) (*Snapshot, error)
	CloneInstance(ctx context.Context, id string, req *CloneRequest) (*Instance, error)
}

// StorageService defines storage operations
type StorageService interface {
	// Bucket operations
	CreateBucket(ctx context.Context, req *CreateBucketRequest) (*Bucket, error)
	GetBucket(ctx context.Context, name string) (*Bucket, error)
	ListBuckets(ctx context.Context) ([]*Bucket, error)
	DeleteBucket(ctx context.Context, name string) error

	// Object operations
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, opts *PutOptions) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket string, prefix string) ([]*Object, error)

	// Advanced operations
	MultipartUpload(ctx context.Context, bucket, key string, reader io.Reader) error
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	SetObjectACL(ctx context.Context, bucket, key string, acl *ACL) error
}

// NetworkService defines network operations
type NetworkService interface {
	// VPC operations
	CreateVPC(ctx context.Context, req *CreateVPCRequest) (*VPC, error)
	GetVPC(ctx context.Context, id string) (*VPC, error)
	ListVPCs(ctx context.Context) ([]*VPC, error)
	DeleteVPC(ctx context.Context, id string) error

	// Subnet operations
	CreateSubnet(ctx context.Context, vpcID string, req *CreateSubnetRequest) (*Subnet, error)
	GetSubnet(ctx context.Context, id string) (*Subnet, error)
	ListSubnets(ctx context.Context, vpcID string) ([]*Subnet, error)
	DeleteSubnet(ctx context.Context, id string) error

	// Security operations
	CreateSecurityGroup(ctx context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, error)
	UpdateSecurityRules(ctx context.Context, groupID string, rules []*SecurityRule) error

	// Load balancing
	CreateLoadBalancer(ctx context.Context, req *CreateLoadBalancerRequest) (*LoadBalancer, error)
	UpdateLoadBalancer(ctx context.Context, id string, req *UpdateLoadBalancerRequest) error
}

// ContainerService defines container operations
type ContainerService interface {
	// Cluster operations
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)
	GetCluster(ctx context.Context, id string) (*Cluster, error)
	ListClusters(ctx context.Context) ([]*Cluster, error)
	UpdateCluster(ctx context.Context, id string, req *UpdateClusterRequest) error
	DeleteCluster(ctx context.Context, id string) error

	// Container operations
	DeployContainer(ctx context.Context, clusterID string, req *DeployRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, id string) (*Deployment, error)
	UpdateDeployment(ctx context.Context, id string, req *UpdateDeploymentRequest) error
	ScaleDeployment(ctx context.Context, id string, replicas int) error
	DeleteDeployment(ctx context.Context, id string) error

	// Service mesh
	EnableServiceMesh(ctx context.Context, clusterID string, config *ServiceMeshConfig) error
	ConfigureTrafficPolicy(ctx context.Context, policy *TrafficPolicy) error
}

// DatabaseService defines database operations
type DatabaseService interface {
	// Database operations
	CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*Database, error)
	GetDatabase(ctx context.Context, id string) (*Database, error)
	ListDatabases(ctx context.Context) ([]*Database, error)
	UpdateDatabase(ctx context.Context, id string, req *UpdateDatabaseRequest) error
	DeleteDatabase(ctx context.Context, id string) error

	// Backup operations
	CreateBackup(ctx context.Context, dbID string, name string) (*Backup, error)
	RestoreBackup(ctx context.Context, backupID string, targetDB string) error
	ListBackups(ctx context.Context, dbID string) ([]*Backup, error)

	// Scaling operations
	ScaleDatabase(ctx context.Context, dbID string, req *ScaleRequest) error
	EnableReadReplica(ctx context.Context, dbID string, region string) (*Database, error)
	EnableMultiMaster(ctx context.Context, dbID string, regions []string) error
}

// SecurityService defines security operations
type SecurityService interface {
	// Identity and Access Management
	CreateRole(ctx context.Context, req *CreateRoleRequest) (*Role, error)
	CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error)
	AttachPolicy(ctx context.Context, roleID, policyID string) error

	// Secrets management
	CreateSecret(ctx context.Context, req *CreateSecretRequest) (*Secret, error)
	GetSecret(ctx context.Context, id string) (*Secret, error)
	RotateSecret(ctx context.Context, id string) error

	// Compliance and auditing
	EnableAuditLogging(ctx context.Context, config *AuditConfig) error
	GetComplianceReport(ctx context.Context, standard string) (*ComplianceReport, error)

	// Encryption
	CreateKMSKey(ctx context.Context, req *CreateKeyRequest) (*KMSKey, error)
	Encrypt(ctx context.Context, keyID string, data []byte) ([]byte, error)
	Decrypt(ctx context.Context, keyID string, data []byte) ([]byte, error)
}

// MonitoringService defines monitoring operations
type MonitoringService interface {
	// Metrics
	PutMetric(ctx context.Context, metric *Metric) error
	GetMetrics(ctx context.Context, query *MetricQuery) ([]*MetricData, error)
	CreateDashboard(ctx context.Context, req *CreateDashboardRequest) (*Dashboard, error)

	// Logging
	CreateLogGroup(ctx context.Context, name string) error
	PutLogs(ctx context.Context, group string, logs []*LogEntry) error
	QueryLogs(ctx context.Context, query *LogQuery) ([]*LogEntry, error)

	// Alerting
	CreateAlert(ctx context.Context, req *CreateAlertRequest) (*Alert, error)
	UpdateAlert(ctx context.Context, id string, req *UpdateAlertRequest) error
	ListAlerts(ctx context.Context) ([]*Alert, error)

	// Tracing
	PutTrace(ctx context.Context, trace *Trace) error
	GetTrace(ctx context.Context, id string) (*Trace, error)
	QueryTraces(ctx context.Context, query *TraceQuery) ([]*Trace, error)
}

// ServerlessService defines serverless operations
type ServerlessService interface {
	// Function operations
	CreateFunction(ctx context.Context, req *CreateFunctionRequest) (*Function, error)
	UpdateFunction(ctx context.Context, id string, req *UpdateFunctionRequest) error
	InvokeFunction(ctx context.Context, id string, payload []byte) ([]byte, error)
	DeleteFunction(ctx context.Context, id string) error

	// Event operations
	CreateEventTrigger(ctx context.Context, functionID string, trigger *EventTrigger) error
	CreateAPIGateway(ctx context.Context, req *CreateAPIGatewayRequest) (*APIGateway, error)

	// Workflow operations
	CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*Workflow, error)
	ExecuteWorkflow(ctx context.Context, id string, input map[string]interface{}) (*WorkflowExecution, error)
}

// AIService defines AI/ML operations
type AIService interface {
	// Model operations
	CreateModel(ctx context.Context, req *CreateModelRequest) (*AIModel, error)
	TrainModel(ctx context.Context, id string, dataset *Dataset) (*TrainingJob, error)
	DeployModel(ctx context.Context, id string, config *DeploymentConfig) (*ModelEndpoint, error)
	Predict(ctx context.Context, endpointID string, data interface{}) (interface{}, error)

	// Data operations
	CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*Dataset, error)
	PreprocessData(ctx context.Context, datasetID string, pipeline *Pipeline) error

	// Advanced AI
	CreateNeuralNetwork(ctx context.Context, architecture *NetworkArchitecture) (*NeuralNetwork, error)
	FineTuneModel(ctx context.Context, modelID string, dataset *Dataset) (*AIModel, error)
	ExplainPrediction(ctx context.Context, modelID string, input interface{}) (*Explanation, error)
}

// CostService defines cost management operations
type CostService interface {
	GetCurrentSpend(ctx context.Context) (*SpendSummary, error)
	GetForecast(ctx context.Context, period time.Duration) (*CostForecast, error)
	SetBudget(ctx context.Context, req *SetBudgetRequest) (*Budget, error)
	GetRecommendations(ctx context.Context) ([]*CostRecommendation, error)
	EnableCostAlerts(ctx context.Context, config *AlertConfig) error
}

// ComplianceService defines compliance operations
type ComplianceService interface {
	RunComplianceCheck(ctx context.Context, standard string) (*ComplianceResult, error)
	GetComplianceStatus(ctx context.Context) (*ComplianceStatus, error)
	RemediateIssue(ctx context.Context, issueID string) error
	GenerateComplianceReport(ctx context.Context, req *ReportRequest) (*Report, error)
	EnableContinuousCompliance(ctx context.Context, standards []string) error
}

// DisasterRecoveryService defines disaster recovery operations
type DisasterRecoveryService interface {
	CreateBackupPlan(ctx context.Context, req *CreateBackupPlanRequest) (*BackupPlan, error)
	TestFailover(ctx context.Context, planID string) (*FailoverTest, error)
	InitiateFailover(ctx context.Context, planID string) (*Failover, error)
	GetRPO(ctx context.Context) (time.Duration, error)
	GetRTO(ctx context.Context) (time.Duration, error)
}

// EdgeService defines edge computing operations
type EdgeService interface {
	DeployToEdge(ctx context.Context, req *EdgeDeployRequest) (*EdgeDeployment, error)
	ListEdgeLocations(ctx context.Context) ([]*EdgeLocation, error)
	UpdateEdgeConfig(ctx context.Context, locationID string, config *EdgeConfig) error
	GetEdgeMetrics(ctx context.Context, locationID string) (*EdgeMetrics, error)
}

// QuantumService defines quantum computing operations (futuristic!)
type QuantumService interface {
	CreateQuantumCircuit(ctx context.Context, req *CreateCircuitRequest) (*QuantumCircuit, error)
	RunQuantumJob(ctx context.Context, circuitID string, shots int) (*QuantumResult, error)
	GetQuantumState(ctx context.Context, jobID string) (*QuantumState, error)
	OptimizeWithQuantum(ctx context.Context, problem *OptimizationProblem) (*Solution, error)
}

// Registry holds registered providers
type Registry struct {
	providers map[string]ProviderFactory
}

// ProviderFactory creates provider instances
type ProviderFactory func(config *ProviderConfig) (Provider, error)

// Global registry
var globalRegistry = &Registry{ //nolint:gochecknoglobals // Required for provider registry singleton
	providers: make(map[string]ProviderFactory),
}

// Register registers a provider factory
func Register(name string, factory ProviderFactory) {
	globalRegistry.providers[name] = factory
}

// RegisterAllProviders registers all available providers
func RegisterAllProviders() {
	// This function will be called to register providers instead of using init()
	// Import statements will trigger the registration
}

// Get returns a provider instance
func Get(name string, config *ProviderConfig) (Provider, error) {
	factory, ok := globalRegistry.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	return factory(config)
}

// List returns all registered provider names
func List() []string {
	names := make([]string, 0, len(globalRegistry.providers))
	for name := range globalRegistry.providers {
		names = append(names, name)
	}
	return names
}
