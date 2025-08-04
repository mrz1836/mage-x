// Package providers contains cloud provider specific implementations and utilities.
package providers

import (
	"errors"
	"fmt"

	"github.com/mrz1836/mage-x/pkg/common/config"
	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/common/paths"
)

// Error definitions for provider operations
var (
	ErrUnknownProviderType = errors.New("unknown provider type")
)

// CloudProvider defines the interface for cloud/platform providers
type CloudProvider interface {
	Config() config.MageLoader
	Environment() env.Environment
	FileOperator() fileops.FileOperator
	PathBuilder() paths.PathBuilder
}

// cloudProviderImpl implements the CloudProvider interface
type cloudProviderImpl struct {
	name         string
	configLoader config.MageLoader
	environment  env.Environment
	fileOperator fileops.FileOperator
	pathBuilder  paths.PathBuilder
}

// NewCloudProvider creates a new cloud provider instance
func NewCloudProvider(providerType string) (CloudProvider, error) {
	switch providerType {
	case "local":
		return newLocalProvider(), nil
	case "aws":
		return newAWSProvider(), nil
	case "azure":
		return newAzureProvider(), nil
	case "gcp":
		return newGCPProvider(), nil
	case "docker":
		return newDockerProvider(), nil
	case "kubernetes":
		return newKubernetesProvider(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownProviderType, providerType)
	}
}

// Config returns the configuration loader
func (p *cloudProviderImpl) Config() config.MageLoader {
	return p.configLoader
}

// Environment returns the environment manager
func (p *cloudProviderImpl) Environment() env.Environment {
	return p.environment
}

// FileOperator returns the file operator
func (p *cloudProviderImpl) FileOperator() fileops.FileOperator {
	return p.fileOperator
}

// PathBuilder returns the path builder
func (p *cloudProviderImpl) PathBuilder() paths.PathBuilder {
	return p.pathBuilder
}

// Local provider implementation
func newLocalProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "local",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// AWS provider implementation
func newAWSProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "aws",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Azure provider implementation
func newAzureProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "azure",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// GCP provider implementation
func newGCPProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "gcp",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Docker provider implementation
func newDockerProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "docker",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Kubernetes provider implementation
func newKubernetesProvider() CloudProvider {
	return &cloudProviderImpl{
		name:         "kubernetes",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}
