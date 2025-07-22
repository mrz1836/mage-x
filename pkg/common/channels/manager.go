package channels

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/errors"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/common/paths"
)

// Manager handles release channel operations
type Manager struct {
	mu         sync.RWMutex
	store      Store
	fileOps    fileops.FileOperator
	paths      paths.PathBuilder
	configs    map[Channel]*ChannelConfig
	validators []ReleaseValidator
	hooks      []ReleaseHook
}

// Store interface for channel data persistence
type Store interface {
	// Release operations
	GetRelease(channel Channel, version string) (*Release, error)
	ListReleases(channel Channel) ([]*Release, error)
	SaveRelease(release *Release) error
	DeleteRelease(channel Channel, version string) error

	// Channel operations
	GetChannelConfig(channel Channel) (*ChannelConfig, error)
	SaveChannelConfig(config *ChannelConfig) error

	// Promotion operations
	GetPromotionHistory(version string) ([]*PromotionRequest, error)
	SavePromotionRequest(request *PromotionRequest) error
}

// ReleaseValidator validates releases before publishing
type ReleaseValidator interface {
	Validate(release *Release) error
}

// ReleaseHook is called after release events
type ReleaseHook interface {
	OnPublish(release *Release) error
	OnPromote(release *Release, from Channel) error
	OnDeprecate(release *Release) error
}

// NewManager creates a new channel manager
func NewManager(store Store, fileOps fileops.FileOperator) *Manager {
	return &Manager{
		store:      store,
		fileOps:    fileOps,
		paths:      paths.NewPathBuilder(""),
		configs:    make(map[Channel]*ChannelConfig),
		validators: []ReleaseValidator{},
		hooks:      []ReleaseHook{},
	}
}

// Initialize sets up the channel manager with default configurations
func (m *Manager) Initialize() error {
	// Set up default channel configurations
	defaultConfigs := []*ChannelConfig{
		{
			Name:          Edge,
			Description:   "Cutting-edge development builds",
			PromotionPath: []Channel{Beta, Nightly},
			RetentionDays: 7,
			AutoPromotion: false,
		},
		{
			Name:          Nightly,
			Description:   "Automated nightly builds",
			PromotionPath: []Channel{Beta},
			RetentionDays: 14,
			AutoPromotion: false,
		},
		{
			Name:          Beta,
			Description:   "Pre-release testing builds",
			PromotionPath: []Channel{Stable},
			RetentionDays: 30,
			AutoPromotion: false,
			RequiredTests: []string{"unit", "integration", "smoke"},
		},
		{
			Name:          Stable,
			Description:   "Production-ready releases",
			PromotionPath: []Channel{LTS},
			RetentionDays: 365,
			AutoPromotion: false,
			RequiredTests: []string{"unit", "integration", "smoke", "performance", "security"},
		},
		{
			Name:          LTS,
			Description:   "Long-term support releases",
			PromotionPath: []Channel{},
			RetentionDays: 1825, // 5 years
			AutoPromotion: false,
			RequiredTests: []string{"unit", "integration", "smoke", "performance", "security", "compatibility"},
		},
	}

	for _, config := range defaultConfigs {
		// Only save if config doesn't exist
		if existing, err := m.store.GetChannelConfig(config.Name); err != nil || existing == nil {
			if err := m.store.SaveChannelConfig(config); err != nil {
				return errors.Wrap(err, "failed to save channel config")
			}
		}
		m.configs[config.Name] = config
	}

	return nil
}

// PublishRelease publishes a new release to a channel
func (m *Manager) PublishRelease(release *Release) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate channel
	if !release.Channel.IsValid() {
		return errors.WithCodef(errors.ErrInvalidArgument, "invalid channel: %s", release.Channel)
	}

	// Validate release
	if err := release.Validate(); err != nil {
		return errors.Wrap(err, "invalid release")
	}

	// Run custom validators
	for _, validator := range m.validators {
		if err := validator.Validate(release); err != nil {
			return errors.Wrap(err, "validation failed")
		}
	}

	// Check if version already exists in channel
	if existing, err := m.store.GetRelease(release.Channel, release.Version); err == nil && existing != nil {
		return errors.WithCodef(errors.ErrAlreadyExists,
			"version %s already exists in channel %s", release.Version, release.Channel)
	}

	// Set publication timestamp
	release.PublishedAt = time.Now()

	// Save release
	if err := m.store.SaveRelease(release); err != nil {
		return errors.Wrap(err, "failed to save release")
	}

	// Run hooks
	for _, hook := range m.hooks {
		if err := hook.OnPublish(release); err != nil {
			// Log error but don't fail the publish
			fmt.Printf("Warning: hook failed: %v\n", err)
		}
	}

	return nil
}

// PromoteRelease promotes a release from one channel to another
func (m *Manager) PromoteRelease(request *PromotionRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate channels
	if !request.FromChannel.IsValid() || !request.ToChannel.IsValid() {
		return errors.WithCode(errors.ErrInvalidArgument, "invalid channel")
	}

	// Check promotion path
	if !request.FromChannel.CanPromoteTo(request.ToChannel) {
		return errors.WithCodef(errors.ErrInvalidArgument,
			"cannot promote from %s to %s", request.FromChannel, request.ToChannel)
	}

	// Get source release
	sourceRelease, err := m.store.GetRelease(request.FromChannel, request.Version)
	if err != nil {
		return errors.Wrap(err, "failed to get source release")
	}
	if sourceRelease == nil {
		return errors.WithCodef(errors.ErrNotFound,
			"release %s not found in channel %s", request.Version, request.FromChannel)
	}

	// Check if already exists in target channel
	if existing, err := m.store.GetRelease(request.ToChannel, request.Version); err == nil && existing != nil {
		return errors.WithCodef(errors.ErrAlreadyExists,
			"version %s already exists in channel %s", request.Version, request.ToChannel)
	}

	// Get target channel config
	targetConfig, exists := m.configs[request.ToChannel]
	if !exists {
		return errors.WithCodef(errors.ErrNotFound, "channel config not found: %s", request.ToChannel)
	}

	// Check required tests
	if len(targetConfig.RequiredTests) > 0 {
		if len(request.TestResults) == 0 {
			return errors.WithCode(errors.ErrInvalidArgument, "test results required for promotion")
		}

		// Verify all required tests passed
		testMap := make(map[string]bool)
		for _, result := range request.TestResults {
			if result.Status == "passed" || result.Status == "success" {
				testMap[result.Name] = true
			}
		}

		for _, required := range targetConfig.RequiredTests {
			if !testMap[required] {
				return errors.WithCodef(errors.ErrTestFailed,
					"required test not passed: %s", required)
			}
		}
	}

	// Check approvers if required
	if len(targetConfig.Approvers) > 0 && request.ApprovedBy == "" {
		return errors.WithCode(errors.ErrUnauthorized, "approval required for promotion")
	}

	// Create promoted release
	promotedRelease := &Release{
		Version:      sourceRelease.Version,
		Channel:      request.ToChannel,
		PublishedAt:  time.Now(),
		ReleasedBy:   request.RequestedBy,
		Changelog:    sourceRelease.Changelog,
		Artifacts:    sourceRelease.Artifacts,
		Dependencies: sourceRelease.Dependencies,
		Metadata:     sourceRelease.Metadata,
		Promoted:     true,
		PromotedFrom: request.FromChannel,
		PromotedAt:   &request.RequestedAt,
	}

	// Save promoted release
	if err := m.store.SaveRelease(promotedRelease); err != nil {
		return errors.Wrap(err, "failed to save promoted release")
	}

	// Save promotion request
	request.ApprovedAt = &time.Time{}
	*request.ApprovedAt = time.Now()
	if err := m.store.SavePromotionRequest(request); err != nil {
		// Log error but don't fail the promotion
		fmt.Printf("Warning: failed to save promotion request: %v\n", err)
	}

	// Run hooks
	for _, hook := range m.hooks {
		if err := hook.OnPromote(promotedRelease, request.FromChannel); err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: hook failed: %v\n", err)
		}
	}

	return nil
}

// GetLatestRelease returns the latest release for a channel
func (m *Manager) GetLatestRelease(channel Channel) (*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	releases, err := m.store.ListReleases(channel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list releases")
	}

	if len(releases) == 0 {
		return nil, nil
	}

	// Sort by published date descending
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].PublishedAt.After(releases[j].PublishedAt)
	})

	return releases[0], nil
}

// ListReleases returns all releases for a channel
func (m *Manager) ListReleases(channel Channel, includeDeprecated bool) ([]*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	releases, err := m.store.ListReleases(channel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list releases")
	}

	if includeDeprecated {
		return releases, nil
	}

	// Filter out deprecated releases
	var active []*Release
	for _, release := range releases {
		if !release.Deprecated {
			active = append(active, release)
		}
	}

	return active, nil
}

// DeprecateRelease marks a release as deprecated
func (m *Manager) DeprecateRelease(channel Channel, version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	release, err := m.store.GetRelease(channel, version)
	if err != nil {
		return errors.Wrap(err, "failed to get release")
	}
	if release == nil {
		return errors.WithCodef(errors.ErrNotFound,
			"release %s not found in channel %s", version, channel)
	}

	if release.Deprecated {
		return errors.WithCode(errors.ErrAlreadyExists, "release already deprecated")
	}

	release.Deprecated = true
	now := time.Now()
	release.DeprecatedAt = &now

	if err := m.store.SaveRelease(release); err != nil {
		return errors.Wrap(err, "failed to save release")
	}

	// Run hooks
	for _, hook := range m.hooks {
		if err := hook.OnDeprecate(release); err != nil {
			// Log error but don't fail
			fmt.Printf("Warning: hook failed: %v\n", err)
		}
	}

	return nil
}

// CleanupExpiredReleases removes releases that have exceeded retention period
func (m *Manager) CleanupExpiredReleases() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var cleanupErrors []error

	for channel, config := range m.configs {
		releases, err := m.store.ListReleases(channel)
		if err != nil {
			cleanupErrors = append(cleanupErrors,
				fmt.Errorf("failed to list releases for %s: %w", channel, err))
			continue
		}

		for _, release := range releases {
			if release.IsExpired(config.RetentionDays) {
				if err := m.store.DeleteRelease(channel, release.Version); err != nil {
					cleanupErrors = append(cleanupErrors,
						fmt.Errorf("failed to delete %s/%s: %w", channel, release.Version, err))
				} else {
					fmt.Printf("Cleaned up expired release: %s/%s\n", channel, release.Version)
				}
			}
		}
	}

	if len(cleanupErrors) > 0 {
		chain := errors.NewChain()
		for _, err := range cleanupErrors {
			chain.Add(err)
		}
		return chain
	}

	return nil
}

// GetChannelStats returns statistics for a channel
func (m *Manager) GetChannelStats(channel Channel) (*ChannelStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	releases, err := m.store.ListReleases(channel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list releases")
	}

	stats := &ChannelStats{
		Channel:       channel,
		TotalReleases: len(releases),
	}

	if len(releases) > 0 {
		// Find latest release
		latest := releases[0]
		for _, release := range releases {
			if release.PublishedAt.After(latest.PublishedAt) {
				latest = release
			}
		}

		stats.LatestVersion = latest.Version
		stats.LatestRelease = latest.PublishedAt
	}

	// Note: Download stats would need to be tracked separately
	// This is a placeholder for the interface

	return stats, nil
}

// AddValidator adds a custom release validator
func (m *Manager) AddValidator(validator ReleaseValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validators = append(m.validators, validator)
}

// AddHook adds a release event hook
func (m *Manager) AddHook(hook ReleaseHook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, hook)
}

// GetPromotionHistory returns the promotion history for a version
func (m *Manager) GetPromotionHistory(version string) ([]*PromotionRequest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.GetPromotionHistory(version)
}
