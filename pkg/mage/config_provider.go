// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"gopkg.in/yaml.v3"
)

// ConfigProvider interface defines methods for configuration management
type ConfigProvider interface {
	// GetConfig returns the current configuration
	GetConfig() (*Config, error)
	// ResetConfig resets the configuration (primarily for testing)
	ResetConfig()
	// SetConfig sets a specific configuration (primarily for testing)
	SetConfig(cfg *Config)
}

// DefaultConfigProvider implements ConfigProvider using the existing singleton pattern
type DefaultConfigProvider struct {
	once   sync.Once
	config *Config
	err    error
	mu     sync.RWMutex
}

// NewDefaultConfigProvider creates a new default config provider
func NewDefaultConfigProvider() *DefaultConfigProvider {
	return &DefaultConfigProvider{}
}

// GetConfig loads and returns the configuration
func (p *DefaultConfigProvider) GetConfig() (*Config, error) {
	p.mu.RLock()
	// Check if config is already set manually (for tests)
	if p.config != nil {
		config := p.config
		p.mu.RUnlock()
		return config, nil
	}
	p.mu.RUnlock()

	p.once.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()

		// Load .env files before processing configuration
		// This ensures environment variables from .env files are available
		// when applyEnvOverrides is called later
		_ = env.LoadEnvFiles() //nolint:errcheck // .env files are optional

		p.config = defaultConfig()
		configFile := ".mage.yaml"

		// Try multiple config file names
		configFiles := []string{".mage.yaml", ".mage.yml", "mage.yaml", "mage.yml"}

		for _, cf := range configFiles {
			if _, err := os.Stat(cf); err == nil {
				configFile = cf
				break
			}
		}

		// Check for enterprise configuration file
		enterpriseConfigFile := ".mage.enterprise.yaml"
		fileOps := fileops.New()
		if fileOps.File.Exists(enterpriseConfigFile) {
			// Load enterprise configuration
			var enterpriseConfig EnterpriseConfiguration
			if err := fileOps.YAML.ReadYAML(enterpriseConfigFile, &enterpriseConfig); err == nil {
				p.config.Enterprise = &enterpriseConfig
			}
		}

		if fileOps.File.Exists(configFile) {
			// Config file exists, read and parse it
			data, err := fileOps.File.ReadFile(configFile)
			if err != nil {
				p.err = fmt.Errorf("failed to read config: %w", err)
				return
			}

			if err := yaml.Unmarshal(data, p.config); err != nil {
				p.err = fmt.Errorf("failed to parse config: %w", err)
				return
			}
		}
		// If config file doesn't exist, we continue with defaults

		// Clean all loaded values to remove inline comments and trim whitespace
		cleanConfigValues(p.config)

		// Apply environment variable overrides (always apply, regardless of config file existence)
		applyEnvOverrides(p.config)
	})

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config, p.err
}

// ResetConfig resets the configuration
func (p *DefaultConfigProvider) ResetConfig() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.once = sync.Once{}
	p.config = nil
	p.err = nil
}

// SetConfig sets a specific configuration (for testing)
func (p *DefaultConfigProvider) SetConfig(cfg *Config) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = cfg
	p.err = nil
}

// ProviderRegistry manages configuration providers
type ProviderRegistry struct {
	mu       sync.RWMutex
	provider ConfigProvider
}

// NewProviderRegistry creates a new provider registry with a default provider
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		provider: NewDefaultConfigProvider(),
	}
}

// SetProvider sets a custom configuration provider
func (r *ProviderRegistry) SetProvider(provider ConfigProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.provider = provider
}

// GetProvider returns the current configuration provider
func (r *ProviderRegistry) GetProvider() ConfigProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.provider
}

// PackageProviderRegistryProvider interface defines methods for managing the package-level provider registry
type PackageProviderRegistryProvider interface {
	// GetRegistry returns the current provider registry
	GetRegistry() *ProviderRegistry
	// SetRegistry sets a new provider registry (primarily for testing)
	SetRegistry(registry *ProviderRegistry)
	// ResetRegistry resets to a new default registry (primarily for testing)
	ResetRegistry()
}

// DefaultPackageProviderRegistryProvider implements PackageProviderRegistryProvider
type DefaultPackageProviderRegistryProvider struct {
	once     sync.Once
	registry *ProviderRegistry
	mu       sync.RWMutex
}

// NewDefaultPackageProviderRegistryProvider creates a new default package provider registry provider
func NewDefaultPackageProviderRegistryProvider() *DefaultPackageProviderRegistryProvider {
	return &DefaultPackageProviderRegistryProvider{}
}

// GetRegistry returns the provider registry, creating it if necessary
func (p *DefaultPackageProviderRegistryProvider) GetRegistry() *ProviderRegistry {
	p.mu.RLock()
	if p.registry != nil {
		registry := p.registry
		p.mu.RUnlock()
		return registry
	}
	p.mu.RUnlock()

	p.once.Do(func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.registry == nil {
			p.registry = NewProviderRegistry()
		}
	})

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.registry
}

// SetRegistry sets a new provider registry (primarily for testing)
func (p *DefaultPackageProviderRegistryProvider) SetRegistry(registry *ProviderRegistry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.registry = registry
}

// ResetRegistry resets to a new default registry (primarily for testing)
func (p *DefaultPackageProviderRegistryProvider) ResetRegistry() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.once = sync.Once{}
	p.registry = nil
}

// Thread-safe singleton instances for package-level provider registry management
var (
	defaultPackageProviderRegistryProviderOnce     sync.Once                       //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultPackageProviderRegistryProviderInstance PackageProviderRegistryProvider //nolint:gochecknoglobals // Required for thread-safe singleton pattern
)

// getDefaultPackageProviderRegistryProvider returns the default package provider registry provider instance
func getDefaultPackageProviderRegistryProvider() PackageProviderRegistryProvider {
	defaultPackageProviderRegistryProviderOnce.Do(func() {
		defaultPackageProviderRegistryProviderInstance = NewDefaultPackageProviderRegistryProvider()
	})
	return defaultPackageProviderRegistryProviderInstance
}

// Package-level default instance for backward compatibility.
// This variable provides the global API while internally using thread-safe singletons.
var (
	packageProviderRegistryProviderInstance PackageProviderRegistryProvider = getDefaultPackageProviderRegistryProvider() //nolint:gochecknoglobals // Required for backward compatibility API
)

// SetPackageProviderRegistryProvider sets a custom package provider registry provider (primarily for testing)
func SetPackageProviderRegistryProvider(provider PackageProviderRegistryProvider) {
	packageProviderRegistryProviderInstance = provider
}

// getPackageProviderRegistryProvider returns the current provider registry provider
func getPackageProviderRegistryProvider() PackageProviderRegistryProvider {
	return packageProviderRegistryProviderInstance
}

// getPackageProviderRegistry returns the current provider registry
func getPackageProviderRegistry() *ProviderRegistry {
	return getPackageProviderRegistryProvider().GetRegistry()
}

// SetConfigProvider sets a custom configuration provider
// This function maintains backward compatibility
func SetConfigProvider(provider ConfigProvider) {
	getPackageProviderRegistry().SetProvider(provider)
}

// GetConfigProvider returns the current configuration provider
// This function maintains backward compatibility
func GetConfigProvider() ConfigProvider {
	return getPackageProviderRegistry().GetProvider()
}

// MockConfigProvider is a simple mock implementation for testing
type MockConfigProvider struct {
	Config *Config
	Err    error
}

// GetConfig returns the mock configuration
func (m *MockConfigProvider) GetConfig() (*Config, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Config == nil {
		return defaultConfig(), nil
	}
	return m.Config, nil
}

// ResetConfig resets the mock configuration
func (m *MockConfigProvider) ResetConfig() {
	m.Config = nil
	m.Err = nil
}

// SetConfig sets the mock configuration
func (m *MockConfigProvider) SetConfig(cfg *Config) {
	m.Config = cfg
}

// Static errors for MockConfigProvider
var (
	ErrMockConfigFailure = errors.New("mock config failure")
)
