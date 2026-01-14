//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// This magefile demonstrates how custom commands work seamlessly
// alongside all built-in MAGE-X commands when using magex

// Static errors for deployment operations
var (
	ErrDeploymentFailed = errors.New("deployment failed")
	ErrStagingFailed    = errors.New("staging deployment failed")
)

// Deploy performs a custom deployment process
// This is a project-specific command that extends MAGE-X
func Deploy() error {
	utils.Info("🚀 Deploying application...")

	// Your custom deployment logic here
	utils.Info("📦 Building production artifacts...")
	time.Sleep(1 * time.Second)

	// Simulate potential deployment error
	if os.Getenv("MAGE_X_DEPLOY_FAIL") != "" {
		return fmt.Errorf("%w: %s", ErrDeploymentFailed, os.Getenv("MAGE_X_DEPLOY_FAIL"))
	}

	utils.Info("🔄 Uploading to servers...")
	time.Sleep(1 * time.Second)

	utils.Info("✅ Deployment complete!")
	return nil
}

// Stage deploys to staging environment
func Stage() error {
	utils.Info("🎯 Deploying to staging...")

	// Staging-specific logic
	if err := os.Setenv("ENV", "staging"); err != nil {
		return fmt.Errorf("failed to set staging environment: %w", err)
	}

	// Simulate staging deployment that could fail
	if os.Getenv("MAGE_X_STAGE_FAIL") != "" {
		return fmt.Errorf("%w: %s", ErrStagingFailed, os.Getenv("MAGE_X_STAGE_FAIL"))
	}

	utils.Info("✅ Staged successfully!")
	return nil
}

// Rollback reverts the last deployment
func Rollback() error {
	utils.Info("⏮️  Rolling back deployment...")

	// Rollback logic here - could fail in real scenarios
	if os.Getenv("MAGE_X_ROLLBACK_FAIL") != "" {
		return fmt.Errorf("%w: %s", ErrStagingFailed, os.Getenv("MAGE_X_ROLLBACK_FAIL"))
	}

	utils.Info("✅ Rollback complete!")
	return nil
}

// Pipeline represents a custom namespace for CI/CD operations
type Pipeline struct{}

// CI runs the custom CI pipeline
func (Pipeline) CI() error {
	utils.Info("🔄 Running custom CI pipeline...")

	// This could call MAGE-X commands internally
	// For example:
	// - magex lint
	// - magex test:race
	// - magex build:all

	// Simulate potential CI failure
	if os.Getenv("MAGE_X_CI_FAIL") != "" {
		return fmt.Errorf("%w: %s", ErrDeploymentFailed, os.Getenv("MAGE_X_CI_FAIL"))
	}

	utils.Info("✅ CI pipeline complete!")
	return nil
}

// CD runs the custom CD pipeline
func (Pipeline) CD() error {
	utils.Info("🚀 Running custom CD pipeline...")

	// Custom continuous deployment logic
	// Simulate potential CD failure
	if os.Getenv("MAGE_X_CD_FAIL") != "" {
		return fmt.Errorf("%w: %s", ErrDeploymentFailed, os.Getenv("MAGE_X_CD_FAIL"))
	}

	utils.Info("✅ CD pipeline complete!")
	return nil
}

// Note: When using magex, you get:
// 1. All 174 built-in MAGE-X commands (build, test, lint, etc.)
// 2. Plus these custom commands (deploy, stage, rollback, pipeline:ci, pipeline:cd)
// 3. All accessible through: magex <command>

// Example usage:
//   magex build          # Built-in MAGE-X command
//   magex test:race      # Built-in MAGE-X command
//   magex deploy         # Custom command from this file
//   magex pipeline:ci    # Custom namespace command
