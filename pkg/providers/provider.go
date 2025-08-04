// Package providers defines interfaces and implementations for cloud/platform providers
package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
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
	ResizeInstance(ctx context.Context, id, size string) error
	SnapshotInstance(ctx context.Context, id, name string) (*Snapshot, error)
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
	ListObjects(ctx context.Context, bucket, prefix string) ([]*Object, error)

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
	CreateBackup(ctx context.Context, dbID, name string) (*Backup, error)
	RestoreBackup(ctx context.Context, backupID, targetDB string) error
	ListBackups(ctx context.Context, dbID string) ([]*Backup, error)

	// Scaling operations
	ScaleDatabase(ctx context.Context, dbID string, req *ScaleRequest) error
	EnableReadReplica(ctx context.Context, dbID, region string) (*Database, error)
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
	mu        sync.RWMutex
	providers map[string]ProviderFactory
}

// ProviderFactory creates provider instances
type ProviderFactory func(config *ProviderConfig) (Provider, error)

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ProviderFactory),
	}
}

// Register registers a provider factory with this registry
func (r *Registry) Register(name string, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = factory
}

// Get returns a provider instance from this registry
func (r *Registry) Get(name string, config *ProviderConfig) (Provider, error) {
	r.mu.RLock()
	factory, ok := r.providers[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	return factory(config)
}

// List returns all registered provider names from this registry
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// HasProvider checks if a provider is registered in this registry
func (r *Registry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.providers[name]
	return ok
}

// defaultRegistry holds the default registry instance used by package-level functions
// This maintains backward compatibility while removing true global state
type registryManager struct {
	mu       sync.RWMutex
	registry *Registry
	once     sync.Once
}

// newRegistryManager creates a new registry manager
func newRegistryManager() *registryManager {
	return &registryManager{}
}

// getOrCreateRegistry returns the default registry, creating it if necessary
func (rm *registryManager) getOrCreateRegistry() *Registry {
	rm.mu.RLock()
	if rm.registry != nil {
		defer rm.mu.RUnlock()
		return rm.registry
	}
	rm.mu.RUnlock()

	rm.once.Do(func() {
		rm.mu.Lock()
		defer rm.mu.Unlock()
		if rm.registry == nil {
			rm.registry = NewRegistry()
		}
	})

	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.registry
}

// GetRegistry returns the provider registry instance with thread-safe lazy initialization
// This function maintains backward compatibility with existing code
func GetRegistry() *Registry {
	return getDefaultRegistryManager().getOrCreateRegistry()
}

// Register registers a provider factory using the default registry
// This function maintains backward compatibility with existing code
func Register(name string, factory ProviderFactory) {
	GetRegistry().Register(name, factory)
}

// getDefaultRegistryManager returns the singleton registry manager instance
// This uses a closure-based singleton pattern for backward compatibility
//
//nolint:gochecknoglobals // Required for maintaining backward compatibility with shared provider registry
var getDefaultRegistryManager = func() func() *registryManager {
	var (
		once     sync.Once
		instance *registryManager
	)

	return func() *registryManager {
		once.Do(func() {
			instance = newRegistryManager()
		})
		return instance
	}
}()

// RegisterAllProviders registers all available providers
func RegisterAllProviders() {
	// This function will be called to register providers instead of using init()
	// Import statements will trigger the registration
}

// Get returns a provider instance using the default registry
// This function maintains backward compatibility with existing code
func Get(name string, config *ProviderConfig) (Provider, error) {
	return GetRegistry().Get(name, config)
}

// List returns all registered provider names using the default registry
// This function maintains backward compatibility with existing code
func List() []string {
	return GetRegistry().List()
}

// HasProvider checks if a provider is registered using the default registry
// This function maintains backward compatibility with existing code
func HasProvider(name string) bool {
	return GetRegistry().HasProvider(name)
}
