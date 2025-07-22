// Package providers defines types used by cloud/platform providers
package providers

import (
	"time"
)

// Instance represents a compute instance
type Instance struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	State     string                 `json:"state"`
	Region    string                 `json:"region"`
	Zone      string                 `json:"zone"`
	PublicIP  string                 `json:"public_ip"`
	PrivateIP string                 `json:"private_ip"`
	CreatedAt time.Time              `json:"created_at"`
	Tags      map[string]string      `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// CreateInstanceRequest represents instance creation request
type CreateInstanceRequest struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Image          string            `json:"image"`
	Region         string            `json:"region"`
	Zone           string            `json:"zone"`
	KeyPair        string            `json:"key_pair,omitempty"`
	SecurityGroups []string          `json:"security_groups,omitempty"`
	UserData       string            `json:"user_data,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	DiskSize       int               `json:"disk_size,omitempty"`
	NetworkConfig  *NetworkConfig    `json:"network_config,omitempty"`
}

// UpdateInstanceRequest represents instance update request
type UpdateInstanceRequest struct {
	Name           *string           `json:"name,omitempty"`
	Type           *string           `json:"type,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	SecurityGroups []string          `json:"security_groups,omitempty"`
}

// InstanceFilter represents instance list filter
type InstanceFilter struct {
	States  []string          `json:"states,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
	Regions []string          `json:"regions,omitempty"`
	Types   []string          `json:"types,omitempty"`
}

// Snapshot represents an instance snapshot
type Snapshot struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	InstanceID  string    `json:"instance_id"`
	State       string    `json:"state"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
}

// CloneRequest represents instance clone request
type CloneRequest struct {
	Name string `json:"name"`
	Zone string `json:"zone,omitempty"`
	Type string `json:"type,omitempty"`
}

// NetworkConfig represents network configuration
type NetworkConfig struct {
	VPCID    string `json:"vpc_id"`
	SubnetID string `json:"subnet_id"`
	PublicIP bool   `json:"public_ip"`
	IPv6     bool   `json:"ipv6"`
}

// Bucket represents a storage bucket
type Bucket struct {
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	CreatedAt    time.Time         `json:"created_at"`
	Versioning   bool              `json:"versioning"`
	Encryption   bool              `json:"encryption"`
	PublicAccess bool              `json:"public_access"`
	Tags         map[string]string `json:"tags"`
}

// CreateBucketRequest represents bucket creation request
type CreateBucketRequest struct {
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	Versioning   bool              `json:"versioning"`
	Encryption   bool              `json:"encryption"`
	PublicAccess bool              `json:"public_access"`
	Tags         map[string]string `json:"tags,omitempty"`
	ACL          string            `json:"acl,omitempty"`
}

// Object represents a storage object
type Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	StorageClass string    `json:"storage_class"`
	ContentType  string    `json:"content_type"`
}

// PutOptions represents object upload options
type PutOptions struct {
	ContentType  string            `json:"content_type,omitempty"`
	CacheControl string            `json:"cache_control,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	ACL          string            `json:"acl,omitempty"`
	Encryption   string            `json:"encryption,omitempty"`
}

// ACL represents access control list
type ACL struct {
	Owner  string   `json:"owner"`
	Grants []*Grant `json:"grants"`
}

// Grant represents an ACL grant
type Grant struct {
	Grantee    string `json:"grantee"`
	Permission string `json:"permission"`
}

// VPC represents a virtual private cloud
type VPC struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	CIDR   string            `json:"cidr"`
	Region string            `json:"region"`
	State  string            `json:"state"`
	Tags   map[string]string `json:"tags"`
}

// CreateVPCRequest represents VPC creation request
type CreateVPCRequest struct {
	Name      string            `json:"name"`
	CIDR      string            `json:"cidr"`
	Region    string            `json:"region"`
	EnableDNS bool              `json:"enable_dns"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Subnet represents a network subnet
type Subnet struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	VPCID  string `json:"vpc_id"`
	CIDR   string `json:"cidr"`
	Zone   string `json:"zone"`
	State  string `json:"state"`
	Public bool   `json:"public"`
}

// CreateSubnetRequest represents subnet creation request
type CreateSubnetRequest struct {
	Name   string            `json:"name"`
	CIDR   string            `json:"cidr"`
	Zone   string            `json:"zone"`
	Public bool              `json:"public"`
	Tags   map[string]string `json:"tags,omitempty"`
}

// SecurityGroup represents a security group
type SecurityGroup struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	VPCID       string          `json:"vpc_id"`
	Rules       []*SecurityRule `json:"rules"`
}

// CreateSecurityGroupRequest represents security group creation
type CreateSecurityGroupRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	VPCID       string          `json:"vpc_id"`
	Rules       []*SecurityRule `json:"rules,omitempty"`
}

// SecurityRule represents a security rule
type SecurityRule struct {
	Direction   string `json:"direction"` // "ingress" or "egress"
	Protocol    string `json:"protocol"`
	FromPort    int    `json:"from_port"`
	ToPort      int    `json:"to_port"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

// LoadBalancer represents a load balancer
type LoadBalancer struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	State        string         `json:"state"`
	DNSName      string         `json:"dns_name"`
	Listeners    []*Listener    `json:"listeners"`
	TargetGroups []*TargetGroup `json:"target_groups"`
}

// CreateLoadBalancerRequest represents load balancer creation
type CreateLoadBalancerRequest struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Subnets        []string          `json:"subnets"`
	SecurityGroups []string          `json:"security_groups,omitempty"`
	Listeners      []*Listener       `json:"listeners"`
	Tags           map[string]string `json:"tags,omitempty"`
}

// UpdateLoadBalancerRequest represents load balancer update
type UpdateLoadBalancerRequest struct {
	Name           *string           `json:"name,omitempty"`
	SecurityGroups []string          `json:"security_groups,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
}

// Listener represents a load balancer listener
type Listener struct {
	Protocol      string `json:"protocol"`
	Port          int    `json:"port"`
	TargetGroupID string `json:"target_group_id"`
	Certificate   string `json:"certificate,omitempty"`
}

// TargetGroup represents a load balancer target group
type TargetGroup struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Protocol    string       `json:"protocol"`
	Port        int          `json:"port"`
	HealthCheck *HealthCheck `json:"health_check"`
	Targets     []string     `json:"targets"`
}

// HealthCheck represents a health check configuration
type HealthCheck struct {
	Protocol           string        `json:"protocol"`
	Port               int           `json:"port"`
	Path               string        `json:"path"`
	Interval           time.Duration `json:"interval"`
	Timeout            time.Duration `json:"timeout"`
	HealthyThreshold   int           `json:"healthy_threshold"`
	UnhealthyThreshold int           `json:"unhealthy_threshold"`
}

// Cluster represents a container cluster
type Cluster struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "kubernetes", "ecs", "swarm"
	Version   string    `json:"version"`
	State     string    `json:"state"`
	Region    string    `json:"region"`
	NodeCount int       `json:"node_count"`
	Endpoint  string    `json:"endpoint"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateClusterRequest represents cluster creation request
type CreateClusterRequest struct {
	Name          string                `json:"name"`
	Type          string                `json:"type"`
	Version       string                `json:"version"`
	Region        string                `json:"region"`
	NodeCount     int                   `json:"node_count"`
	NodeType      string                `json:"node_type"`
	NetworkConfig *ClusterNetworkConfig `json:"network_config,omitempty"`
	Tags          map[string]string     `json:"tags,omitempty"`
}

// UpdateClusterRequest represents cluster update request
type UpdateClusterRequest struct {
	Version   *string           `json:"version,omitempty"`
	NodeCount *int              `json:"node_count,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// ClusterNetworkConfig represents cluster network configuration
type ClusterNetworkConfig struct {
	VPCID          string   `json:"vpc_id"`
	SubnetIDs      []string `json:"subnet_ids"`
	SecurityGroups []string `json:"security_groups"`
}

// Deployment represents a container deployment
type Deployment struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ClusterID string    `json:"cluster_id"`
	Image     string    `json:"image"`
	Replicas  int       `json:"replicas"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeployRequest represents container deployment request
type DeployRequest struct {
	Name        string                `json:"name"`
	Image       string                `json:"image"`
	Replicas    int                   `json:"replicas"`
	CPU         string                `json:"cpu,omitempty"`
	Memory      string                `json:"memory,omitempty"`
	Environment map[string]string     `json:"environment,omitempty"`
	Ports       []int                 `json:"ports,omitempty"`
	Command     []string              `json:"command,omitempty"`
	HealthCheck *ContainerHealthCheck `json:"health_check,omitempty"`
}

// UpdateDeploymentRequest represents deployment update request
type UpdateDeploymentRequest struct {
	Image       *string           `json:"image,omitempty"`
	Replicas    *int              `json:"replicas,omitempty"`
	CPU         *string           `json:"cpu,omitempty"`
	Memory      *string           `json:"memory,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// ContainerHealthCheck represents container health check
type ContainerHealthCheck struct {
	Type     string        `json:"type"` // "http", "tcp", "exec"
	Path     string        `json:"path,omitempty"`
	Port     int           `json:"port,omitempty"`
	Command  []string      `json:"command,omitempty"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// ServiceMeshConfig represents service mesh configuration
type ServiceMeshConfig struct {
	Type    string `json:"type"` // "istio", "linkerd", "consul"
	MTLS    bool   `json:"mtls"`
	Tracing bool   `json:"tracing"`
	Metrics bool   `json:"metrics"`
}

// TrafficPolicy represents traffic management policy
type TrafficPolicy struct {
	Name    string         `json:"name"`
	Service string         `json:"service"`
	Rules   []*TrafficRule `json:"rules"`
}

// TrafficRule represents a traffic rule
type TrafficRule struct {
	Weight  int               `json:"weight"`
	Version string            `json:"version"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Database represents a managed database
type Database struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Engine    string    `json:"engine"`
	Version   string    `json:"version"`
	State     string    `json:"state"`
	Endpoint  string    `json:"endpoint"`
	Port      int       `json:"port"`
	Size      string    `json:"size"`
	Storage   int       `json:"storage"`
	MultiAZ   bool      `json:"multi_az"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateDatabaseRequest represents database creation request
type CreateDatabaseRequest struct {
	Name     string            `json:"name"`
	Engine   string            `json:"engine"`
	Version  string            `json:"version"`
	Size     string            `json:"size"`
	Storage  int               `json:"storage"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	MultiAZ  bool              `json:"multi_az"`
	Backup   bool              `json:"backup"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// UpdateDatabaseRequest represents database update request
type UpdateDatabaseRequest struct {
	Size     *string           `json:"size,omitempty"`
	Storage  *int              `json:"storage,omitempty"`
	Version  *string           `json:"version,omitempty"`
	Password *string           `json:"password,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// Backup represents a database backup
type Backup struct {
	ID         string    `json:"id"`
	DatabaseID string    `json:"database_id"`
	Name       string    `json:"name"`
	State      string    `json:"state"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// ScaleRequest represents database scaling request
type ScaleRequest struct {
	Size    string `json:"size"`
	IOPS    int    `json:"iops,omitempty"`
	Storage int    `json:"storage,omitempty"`
}

// Role represents an IAM role
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Policies    []string  `json:"policies"`
	TrustPolicy string    `json:"trust_policy"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TrustPolicy string   `json:"trust_policy"`
	Policies    []string `json:"policies,omitempty"`
}

// Policy represents an IAM policy
type Policy struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Document    string    `json:"document"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreatePolicyRequest represents policy creation request
type CreatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Document    string `json:"document"`
}

// Secret represents a managed secret
type Secret struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Version      string    `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	NextRotation time.Time `json:"next_rotation,omitempty"`
}

// CreateSecretRequest represents secret creation request
type CreateSecretRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Value        string            `json:"value"`
	AutoRotate   bool              `json:"auto_rotate"`
	RotationDays int               `json:"rotation_days,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// AuditConfig represents audit logging configuration
type AuditConfig struct {
	Enabled   bool     `json:"enabled"`
	LogGroup  string   `json:"log_group"`
	Events    []string `json:"events"`
	Retention int      `json:"retention_days"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	Standard    string              `json:"standard"`
	Score       float64             `json:"score"`
	Passed      int                 `json:"passed"`
	Failed      int                 `json:"failed"`
	Warnings    int                 `json:"warnings"`
	GeneratedAt time.Time           `json:"generated_at"`
	Details     []*ComplianceDetail `json:"details"`
}

// ComplianceDetail represents compliance check detail
type ComplianceDetail struct {
	Rule     string `json:"rule"`
	Status   string `json:"status"`
	Resource string `json:"resource"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// KMSKey represents a key management service key
type KMSKey struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Algorithm    string    `json:"algorithm"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	NextRotation time.Time `json:"next_rotation,omitempty"`
}

// CreateKeyRequest represents KMS key creation request
type CreateKeyRequest struct {
	Name       string            `json:"name"`
	Algorithm  string            `json:"algorithm"`
	Usage      string            `json:"usage"`
	AutoRotate bool              `json:"auto_rotate"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Metric represents a monitoring metric
type Metric struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit"`
	Timestamp  time.Time         `json:"timestamp"`
	Dimensions map[string]string `json:"dimensions"`
}

// MetricQuery represents a metric query
type MetricQuery struct {
	Namespace  string            `json:"namespace"`
	MetricName string            `json:"metric_name"`
	Dimensions map[string]string `json:"dimensions,omitempty"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Period     time.Duration     `json:"period"`
	Statistic  string            `json:"statistic"`
}

// MetricData represents metric data points
type MetricData struct {
	Label      string      `json:"label"`
	Timestamps []time.Time `json:"timestamps"`
	Values     []float64   `json:"values"`
	Unit       string      `json:"unit"`
}

// Dashboard represents a monitoring dashboard
type Dashboard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Widgets     []*Widget `json:"widgets"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateDashboardRequest represents dashboard creation request
type CreateDashboardRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Widgets     []*Widget `json:"widgets"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Query      interface{}            `json:"query"`
	Properties map[string]interface{} `json:"properties"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	Source    string                 `json:"source"`
	Fields    map[string]interface{} `json:"fields"`
}

// LogQuery represents a log query
type LogQuery struct {
	LogGroup  string    `json:"log_group"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Filter    string    `json:"filter,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Condition   *AlertCondition `json:"condition"`
	Actions     []*AlertAction  `json:"actions"`
	State       string          `json:"state"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
}

// CreateAlertRequest represents alert creation request
type CreateAlertRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Condition   *AlertCondition `json:"condition"`
	Actions     []*AlertAction  `json:"actions"`
	Enabled     bool            `json:"enabled"`
}

// UpdateAlertRequest represents alert update request
type UpdateAlertRequest struct {
	Name        *string         `json:"name,omitempty"`
	Description *string         `json:"description,omitempty"`
	Condition   *AlertCondition `json:"condition,omitempty"`
	Actions     []*AlertAction  `json:"actions,omitempty"`
	Enabled     *bool           `json:"enabled,omitempty"`
}

// AlertCondition represents alert condition
type AlertCondition struct {
	MetricName        string        `json:"metric_name"`
	Namespace         string        `json:"namespace"`
	Statistic         string        `json:"statistic"`
	Threshold         float64       `json:"threshold"`
	Comparison        string        `json:"comparison"`
	Period            time.Duration `json:"period"`
	EvaluationPeriods int           `json:"evaluation_periods"`
}

// AlertAction represents alert action
type AlertAction struct {
	Type       string            `json:"type"` // "email", "sms", "webhook"
	Target     string            `json:"target"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Trace represents a distributed trace
type Trace struct {
	ID            string        `json:"id"`
	ServiceName   string        `json:"service_name"`
	OperationName string        `json:"operation_name"`
	StartTime     time.Time     `json:"start_time"`
	Duration      time.Duration `json:"duration"`
	Status        string        `json:"status"`
	Spans         []*Span       `json:"spans"`
}

// Span represents a trace span
type Span struct {
	ID            string            `json:"id"`
	ParentID      string            `json:"parent_id,omitempty"`
	OperationName string            `json:"operation_name"`
	ServiceName   string            `json:"service_name"`
	StartTime     time.Time         `json:"start_time"`
	Duration      time.Duration     `json:"duration"`
	Status        string            `json:"status"`
	Tags          map[string]string `json:"tags"`
	Logs          []*SpanLog        `json:"logs"`
}

// SpanLog represents a span log entry
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// TraceQuery represents a trace query
type TraceQuery struct {
	ServiceName   string            `json:"service_name,omitempty"`
	OperationName string            `json:"operation_name,omitempty"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	MinDuration   time.Duration     `json:"min_duration,omitempty"`
	MaxDuration   time.Duration     `json:"max_duration,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
	Limit         int               `json:"limit,omitempty"`
}

// Function represents a serverless function
type Function struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Runtime      string            `json:"runtime"`
	Handler      string            `json:"handler"`
	CodeSize     int64             `json:"code_size"`
	Timeout      time.Duration     `json:"timeout"`
	Memory       int               `json:"memory"`
	Environment  map[string]string `json:"environment"`
	State        string            `json:"state"`
	LastModified time.Time         `json:"last_modified"`
}

// CreateFunctionRequest represents function creation request
type CreateFunctionRequest struct {
	Name        string            `json:"name"`
	Runtime     string            `json:"runtime"`
	Handler     string            `json:"handler"`
	Code        []byte            `json:"code"`
	Timeout     time.Duration     `json:"timeout"`
	Memory      int               `json:"memory"`
	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// UpdateFunctionRequest represents function update request
type UpdateFunctionRequest struct {
	Code        []byte            `json:"code,omitempty"`
	Handler     *string           `json:"handler,omitempty"`
	Timeout     *time.Duration    `json:"timeout,omitempty"`
	Memory      *int              `json:"memory,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// EventTrigger represents a function event trigger
type EventTrigger struct {
	Type       string                 `json:"type"` // "http", "schedule", "queue", "storage"
	Source     string                 `json:"source"`
	Properties map[string]interface{} `json:"properties"`
}

// APIGateway represents an API gateway
type APIGateway struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Endpoint    string    `json:"endpoint"`
	Routes      []*Route  `json:"routes"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAPIGatewayRequest represents API gateway creation
type CreateAPIGatewayRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Routes      []*Route          `json:"routes"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// Route represents an API gateway route
type Route struct {
	Path       string `json:"path"`
	Method     string `json:"method"`
	FunctionID string `json:"function_id"`
	Auth       bool   `json:"auth"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

// Workflow represents a serverless workflow
type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Definition  string    `json:"definition"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateWorkflowRequest represents workflow creation request
type CreateWorkflowRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Definition  string            `json:"definition"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// WorkflowExecution represents a workflow execution
type WorkflowExecution struct {
	ID         string                 `json:"id"`
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"`
	Input      map[string]interface{} `json:"input"`
	Output     map[string]interface{} `json:"output,omitempty"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time,omitempty"`
}

// AIModel represents an AI/ML model
type AIModel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Framework string    `json:"framework"`
	Version   string    `json:"version"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateModelRequest represents model creation request
type CreateModelRequest struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Framework    string            `json:"framework"`
	Architecture interface{}       `json:"architecture"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// Dataset represents a training dataset
type Dataset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	Location  string    `json:"location"`
	Format    string    `json:"format"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateDatasetRequest represents dataset creation request
type CreateDatasetRequest struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Source string            `json:"source"`
	Format string            `json:"format"`
	Tags   map[string]string `json:"tags,omitempty"`
}

// TrainingJob represents a model training job
type TrainingJob struct {
	ID        string             `json:"id"`
	ModelID   string             `json:"model_id"`
	DatasetID string             `json:"dataset_id"`
	Status    string             `json:"status"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time,omitempty"`
	Metrics   map[string]float64 `json:"metrics"`
}

// DeploymentConfig represents model deployment configuration
type DeploymentConfig struct {
	InstanceType  string `json:"instance_type"`
	InstanceCount int    `json:"instance_count"`
	AutoScale     bool   `json:"auto_scale"`
	MaxInstances  int    `json:"max_instances,omitempty"`
}

// ModelEndpoint represents a deployed model endpoint
type ModelEndpoint struct {
	ID        string    `json:"id"`
	ModelID   string    `json:"model_id"`
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Pipeline represents a data processing pipeline
type Pipeline struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Steps []*PipelineStep `json:"steps"`
}

// PipelineStep represents a pipeline step
type PipelineStep struct {
	Name   string                 `json:"name"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// NetworkArchitecture represents neural network architecture
type NetworkArchitecture struct {
	Layers       []*Layer `json:"layers"`
	Optimizer    string   `json:"optimizer"`
	LossFunction string   `json:"loss_function"`
	Metrics      []string `json:"metrics"`
}

// Layer represents a neural network layer
type Layer struct {
	Type       string                 `json:"type"`
	Units      int                    `json:"units,omitempty"`
	Activation string                 `json:"activation,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

// NeuralNetwork represents a neural network
type NeuralNetwork struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Architecture *NetworkArchitecture `json:"architecture"`
	State        string               `json:"state"`
	Parameters   int64                `json:"parameters"`
}

// Explanation represents model prediction explanation
type Explanation struct {
	Prediction    interface{}        `json:"prediction"`
	Confidence    float64            `json:"confidence"`
	Features      map[string]float64 `json:"features"`
	Importance    map[string]float64 `json:"importance"`
	Visualization string             `json:"visualization,omitempty"`
}

// SpendSummary represents current spend summary
type SpendSummary struct {
	Total     float64            `json:"total"`
	ByService map[string]float64 `json:"by_service"`
	ByRegion  map[string]float64 `json:"by_region"`
	ByTag     map[string]float64 `json:"by_tag"`
	Period    string             `json:"period"`
	Currency  string             `json:"currency"`
}

// CostForecast represents cost forecast
type CostForecast struct {
	Period     time.Duration      `json:"period"`
	Predicted  float64            `json:"predicted"`
	Confidence float64            `json:"confidence"`
	Breakdown  map[string]float64 `json:"breakdown"`
}

// SetBudgetRequest represents budget setting request
type SetBudgetRequest struct {
	Name   string            `json:"name"`
	Amount float64           `json:"amount"`
	Period string            `json:"period"` // "monthly", "quarterly", "yearly"
	Alerts []*BudgetAlert    `json:"alerts"`
	Tags   map[string]string `json:"tags,omitempty"`
}

// Budget represents a cost budget
type Budget struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Amount       float64        `json:"amount"`
	Period       string         `json:"period"`
	CurrentSpend float64        `json:"current_spend"`
	Percentage   float64        `json:"percentage"`
	Alerts       []*BudgetAlert `json:"alerts"`
}

// BudgetAlert represents a budget alert
type BudgetAlert struct {
	Threshold  float64  `json:"threshold"`
	Type       string   `json:"type"` // "percentage" or "amount"
	Recipients []string `json:"recipients"`
}

// CostRecommendation represents a cost saving recommendation
type CostRecommendation struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Resource    string  `json:"resource"`
	Description string  `json:"description"`
	Savings     float64 `json:"savings"`
	Impact      string  `json:"impact"`
	Effort      string  `json:"effort"`
}

// AlertConfig represents cost alert configuration
type AlertConfig struct {
	Thresholds []*CostThreshold `json:"thresholds"`
	Recipients []string         `json:"recipients"`
	Frequency  string           `json:"frequency"`
}

// CostThreshold represents a cost threshold
type CostThreshold struct {
	Amount  float64           `json:"amount"`
	Service string            `json:"service,omitempty"`
	Tag     map[string]string `json:"tag,omitempty"`
}

// ComplianceResult represents compliance check result
type ComplianceResult struct {
	Standard  string     `json:"standard"`
	Passed    bool       `json:"passed"`
	Score     float64    `json:"score"`
	Findings  []*Finding `json:"findings"`
	Timestamp time.Time  `json:"timestamp"`
}

// Finding represents a compliance finding
type Finding struct {
	ID          string `json:"id"`
	Rule        string `json:"rule"`
	Resource    string `json:"resource"`
	Status      string `json:"status"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
}

// ComplianceStatus represents overall compliance status
type ComplianceStatus struct {
	Standards   map[string]float64 `json:"standards"`
	Overall     float64            `json:"overall"`
	Trend       string             `json:"trend"`
	LastChecked time.Time          `json:"last_checked"`
}

// ReportRequest represents report generation request
type ReportRequest struct {
	Type      string    `json:"type"`
	Format    string    `json:"format"` // "pdf", "json", "csv"
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Standards []string  `json:"standards"`
}

// Report represents a compliance report
type Report struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Format      string    `json:"format"`
	URL         string    `json:"url"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// CreateBackupPlanRequest represents backup plan creation
type CreateBackupPlanRequest struct {
	Name       string   `json:"name"`
	Resources  []string `json:"resources"`
	Schedule   string   `json:"schedule"`
	Retention  int      `json:"retention_days"`
	Regions    []string `json:"regions"`
	Encryption bool     `json:"encryption"`
}

// BackupPlan represents a disaster recovery backup plan
type BackupPlan struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Resources  []string  `json:"resources"`
	Schedule   string    `json:"schedule"`
	Retention  int       `json:"retention_days"`
	LastBackup time.Time `json:"last_backup"`
	NextBackup time.Time `json:"next_backup"`
}

// FailoverTest represents a failover test result
type FailoverTest struct {
	ID        string        `json:"id"`
	PlanID    string        `json:"plan_id"`
	Status    string        `json:"status"`
	Duration  time.Duration `json:"duration"`
	RPO       time.Duration `json:"rpo"`
	RTO       time.Duration `json:"rto"`
	Timestamp time.Time     `json:"timestamp"`
}

// Failover represents an active failover
type Failover struct {
	ID           string    `json:"id"`
	PlanID       string    `json:"plan_id"`
	Status       string    `json:"status"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	SourceRegion string    `json:"source_region"`
	TargetRegion string    `json:"target_region"`
}

// EdgeDeployRequest represents edge deployment request
type EdgeDeployRequest struct {
	Name        string      `json:"name"`
	Application string      `json:"application"`
	Locations   []string    `json:"locations"`
	Config      *EdgeConfig `json:"config"`
}

// EdgeDeployment represents an edge deployment
type EdgeDeployment struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Application string          `json:"application"`
	Locations   []*EdgeLocation `json:"locations"`
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
}

// EdgeLocation represents an edge location
type EdgeLocation struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Region    string        `json:"region"`
	Type      string        `json:"type"` // "pop", "cell-tower", "iot-hub"
	Capacity  int           `json:"capacity"`
	Available int           `json:"available"`
	Latency   time.Duration `json:"latency"`
}

// EdgeConfig represents edge configuration
type EdgeConfig struct {
	AutoScale    bool `json:"auto_scale"`
	MinInstances int  `json:"min_instances"`
	MaxInstances int  `json:"max_instances"`
	CacheSize    int  `json:"cache_size_gb"`
}

// EdgeMetrics represents edge location metrics
type EdgeMetrics struct {
	LocationID   string        `json:"location_id"`
	Requests     int64         `json:"requests"`
	Latency      time.Duration `json:"latency"`
	Bandwidth    int64         `json:"bandwidth_bytes"`
	CacheHitRate float64       `json:"cache_hit_rate"`
	ErrorRate    float64       `json:"error_rate"`
}

// CreateCircuitRequest represents quantum circuit creation
type CreateCircuitRequest struct {
	Name   string         `json:"name"`
	Qubits int            `json:"qubits"`
	Gates  []*QuantumGate `json:"gates"`
}

// QuantumCircuit represents a quantum circuit
type QuantumCircuit struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Qubits    int            `json:"qubits"`
	Gates     []*QuantumGate `json:"gates"`
	Depth     int            `json:"depth"`
	CreatedAt time.Time      `json:"created_at"`
}

// QuantumGate represents a quantum gate
type QuantumGate struct {
	Type       string    `json:"type"`
	Target     []int     `json:"target"`
	Control    []int     `json:"control,omitempty"`
	Parameters []float64 `json:"parameters,omitempty"`
}

// QuantumResult represents quantum computation result
type QuantumResult struct {
	ID            string             `json:"id"`
	CircuitID     string             `json:"circuit_id"`
	Shots         int                `json:"shots"`
	Counts        map[string]int     `json:"counts"`
	Probabilities map[string]float64 `json:"probabilities"`
	ExecutionTime time.Duration      `json:"execution_time"`
}

// QuantumState represents quantum state
type QuantumState struct {
	StateVector   []complex128 `json:"state_vector"`
	Probabilities []float64    `json:"probabilities"`
	Entanglement  float64      `json:"entanglement"`
}

// OptimizationProblem represents a quantum optimization problem
type OptimizationProblem struct {
	Type        string   `json:"type"`
	Variables   int      `json:"variables"`
	Constraints []string `json:"constraints"`
	Objective   string   `json:"objective"`
}

// Solution represents an optimization solution
type Solution struct {
	Values     map[string]float64 `json:"values"`
	Cost       float64            `json:"cost"`
	Iterations int                `json:"iterations"`
	Optimal    bool               `json:"optimal"`
}
