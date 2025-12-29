package exec

import (
	"time"

	"github.com/mrz1836/mage-x/pkg/retry"
)

// Builder provides a fluent API for constructing executor chains
type Builder struct {
	base          *Base
	validators    []ValidatingOption
	timeout       *TimeoutExecutor
	retrying      *RetryingExecutor
	envFilter     *EnvFilteringExecutor
	auditing      *AuditingExecutor
	useValidate   bool
	useTimeout    bool
	useRetry      bool
	useEnvFilter  bool
	useAuditLog   bool
	auditLogger   AuditLogger
	auditMetadata map[string]string
}

// NewBuilder creates a new executor builder
func NewBuilder() *Builder {
	return &Builder{
		base: NewBase(),
	}
}

// WithWorkingDirectory sets the base working directory
func (b *Builder) WithWorkingDirectory(dir string) *Builder {
	b.base.WorkingDir = dir
	return b
}

// WithVerbose enables verbose logging
func (b *Builder) WithVerbose(verbose bool) *Builder {
	b.base.Verbose = verbose
	return b
}

// WithBaseEnv sets additional base environment variables
func (b *Builder) WithBaseEnv(env []string) *Builder {
	b.base.Env = env
	return b
}

// WithValidation adds command/argument validation
func (b *Builder) WithValidation(opts ...ValidatingOption) *Builder {
	b.useValidate = true
	b.validators = append(b.validators, opts...)
	return b
}

// WithTimeout adds timeout support with default timeout
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.useTimeout = true
	b.timeout = &TimeoutExecutor{
		DefaultTimeout: timeout,
	}
	return b
}

// WithAdaptiveTimeout adds timeout support with per-command timeouts
func (b *Builder) WithAdaptiveTimeout() *Builder {
	b.useTimeout = true
	b.timeout = &TimeoutExecutor{
		DefaultTimeout:  30 * time.Second,
		TimeoutResolver: NewAdaptiveTimeoutResolver(),
	}
	return b
}

// WithTimeoutResolver adds timeout support with custom resolver
func (b *Builder) WithTimeoutResolver(resolver TimeoutResolver) *Builder {
	b.useTimeout = true
	b.timeout = &TimeoutExecutor{
		DefaultTimeout:  30 * time.Second,
		TimeoutResolver: resolver,
	}
	return b
}

// WithRetry adds retry support with defaults
func (b *Builder) WithRetry(maxRetries int) *Builder {
	b.useRetry = true
	b.retrying = &RetryingExecutor{
		MaxRetries: maxRetries,
		Classifier: retry.NewNetworkClassifier(),
		Backoff:    retry.DefaultBackoff(),
	}
	return b
}

// WithRetryClassifier adds retry support with custom classifier
func (b *Builder) WithRetryClassifier(classifier retry.Classifier, maxRetries int) *Builder {
	b.useRetry = true
	b.retrying = &RetryingExecutor{
		MaxRetries: maxRetries,
		Classifier: classifier,
		Backoff:    retry.DefaultBackoff(),
	}
	return b
}

// WithEnvFiltering adds environment variable filtering
func (b *Builder) WithEnvFiltering(opts ...EnvFilterOption) *Builder {
	b.useEnvFilter = true
	b.envFilter = &EnvFilteringExecutor{
		SensitivePrefixes: DefaultSensitivePrefixes,
		Whitelist:         DefaultEnvWhitelist,
	}
	for _, opt := range opts {
		opt(b.envFilter)
	}
	return b
}

// WithDryRun enables dry run mode (log commands without executing)
func (b *Builder) WithDryRun(dryRun bool) *Builder {
	b.base.DryRun = dryRun
	return b
}

// WithAuditLogging enables audit logging with optional custom logger
func (b *Builder) WithAuditLogging(logger AuditLogger) *Builder {
	b.useAuditLog = true
	b.auditLogger = logger
	return b
}

// WithAuditMetadata adds custom metadata to audit events
func (b *Builder) WithAuditMetadata(key, value string) *Builder {
	if b.auditMetadata == nil {
		b.auditMetadata = make(map[string]string)
	}
	b.auditMetadata[key] = value
	return b
}

// Build constructs the executor chain
// The order of decorators from innermost to outermost is:
// Base -> EnvFilter -> Validation -> Audit -> Retry -> Timeout
func (b *Builder) Build() FullExecutor {
	var executor FullExecutor = b.base

	// Apply env filtering first (closest to execution)
	if b.useEnvFilter {
		b.envFilter.wrapped = executor
		executor = b.envFilter
	}

	// Apply validation
	if b.useValidate {
		executor = NewValidatingExecutor(executor, b.validators...)
	}

	// Apply audit logging
	if b.useAuditLog {
		opts := []AuditingOption{
			WithAuditWorkingDir(b.base.WorkingDir),
			WithAuditDryRun(b.base.DryRun),
		}
		if b.auditLogger != nil {
			opts = append(opts, WithAuditLogger(b.auditLogger))
		}
		for k, v := range b.auditMetadata {
			opts = append(opts, WithAuditMetadata(k, v))
		}
		b.auditing = NewAuditingExecutor(executor, opts...)
		executor = b.auditing
	}

	// Apply retry (wraps validation so retries include validation)
	if b.useRetry {
		b.retrying.wrapped = executor
		executor = b.retrying
	}

	// Apply timeout last (outermost - controls total time including retries)
	if b.useTimeout {
		b.timeout.wrapped = executor
		executor = b.timeout
	}

	return executor
}

// Secure creates an executor with common security settings:
// - Validation
// - Environment filtering
// - Adaptive timeouts
func Secure() FullExecutor {
	return NewBuilder().
		WithValidation().
		WithEnvFiltering().
		WithAdaptiveTimeout().
		Build()
}

// SecureWithRetry creates a secure executor with retry support
func SecureWithRetry(maxRetries int) FullExecutor {
	return NewBuilder().
		WithValidation().
		WithEnvFiltering().
		WithRetryClassifier(CommandClassifier, maxRetries).
		WithAdaptiveTimeout().
		Build()
}

// Simple creates a basic executor without any decorators
func Simple() Executor {
	return NewBase()
}
