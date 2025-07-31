// Package providers contains cloud provider specific implementations and utilities.
package providers

import (
	"fmt"

	"github.com/mrz1836/go-mage/pkg/common/config"
	"github.com/mrz1836/go-mage/pkg/common/env"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/common/paths"
)

// Provider defines the interface for cloud/platform providers
type Provider interface {
	Config() config.MageLoader
	Environment() env.Environment
	FileOperator() fileops.FileOperator
	PathBuilder() paths.PathBuilder
}

// providerImpl implements the Provider interface
type providerImpl struct {
	name         string
	configLoader config.MageLoader
	environment  env.Environment
	fileOperator fileops.FileOperator
	pathBuilder  paths.PathBuilder
}

// NewProvider creates a new provider instance
func NewProvider(providerType string) (Provider, error) {
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
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// Config returns the configuration loader
func (p *providerImpl) Config() config.MageLoader {
	return p.configLoader
}

// Environment returns the environment manager
func (p *providerImpl) Environment() env.Environment {
	return p.environment
}

// FileOperator returns the file operator
func (p *providerImpl) FileOperator() fileops.FileOperator {
	return p.fileOperator
}

// PathBuilder returns the path builder
func (p *providerImpl) PathBuilder() paths.PathBuilder {
	return p.pathBuilder
}

// Local provider implementation
func newLocalProvider() Provider {
	return &providerImpl{
		name:         "local",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// AWS provider implementation
func newAWSProvider() Provider {
	return &providerImpl{
		name:         "aws",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Azure provider implementation
func newAzureProvider() Provider {
	return &providerImpl{
		name:         "azure",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// GCP provider implementation
func newGCPProvider() Provider {
	return &providerImpl{
		name:         "gcp",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Docker provider implementation
func newDockerProvider() Provider {
	return &providerImpl{
		name:         "docker",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}

// Kubernetes provider implementation
func newKubernetesProvider() Provider {
	return &providerImpl{
		name:         "kubernetes",
		configLoader: config.NewLoader(),
		environment:  env.NewEnvironment(),
		fileOperator: fileops.NewFileOperator(),
		pathBuilder:  paths.NewPathBuilder(""),
	}
}
