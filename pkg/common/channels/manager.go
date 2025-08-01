package channels

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	mageErrors "github.com/mrz1836/go-mage/pkg/common/errors"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/common/paths"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Sentinel errors
var (
	ErrNoReleasesFound          = errors.New("no releases found")
	ErrInvalidChannel           = errors.New("invalid channel")
	ErrVersionAlreadyExists     = errors.New("version already exists in channel")
	ErrInvalidArgument          = errors.New("invalid argument")
	ErrCannotPromote            = errors.New("cannot promote between channels")
	ErrReleaseNotFoundInChannel = errors.New("release not found in channel")
	ErrChannelConfigNotFound    = errors.New("channel config not found")
	ErrTestResultsRequired      = errors.New("test results required for promotion")
	ErrTestFailed               = errors.New("required test not passed")
	ErrApprovalRequired         = errors.New("approval required for promotion")
	ErrAlreadyDeprecated        = errors.New("release already deprecated")
	ErrCleanupFailed            = errors.New("cleanup failed with errors")
	ErrMockOperation            = errors.New("mock error")
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
				return mageErrors.Wrap(err, "failed to save channel config")
			}
			m.configs[config.Name] = config
		} else {
			// Use existing config
			m.configs[config.Name] = existing
		}
	}

	return nil
}

// PublishRelease publishes a new release to a channel
func (m *Manager) PublishRelease(release *Release) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate channel
	if !release.Channel.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidChannel, release.Channel)
	}

	// Validate release
	if err := release.Validate(); err != nil {
		return mageErrors.Wrap(err, "invalid release")
	}

	// Run custom validators
	for _, validator := range m.validators {
		if err := validator.Validate(release); err != nil {
			return mageErrors.Wrap(err, "validation failed")
		}
	}

	// Check if version already exists in channel
	if existing, err := m.store.GetRelease(release.Channel, release.Version); err == nil && existing != nil {
		return fmt.Errorf("%w: version %s in channel %s", ErrVersionAlreadyExists, release.Version, release.Channel)
	}

	// Set publication timestamp
	release.PublishedAt = time.Now()

	// Save release
	if err := m.store.SaveRelease(release); err != nil {
		return mageErrors.Wrap(err, "failed to save release")
	}

	// Run hooks
	for _, hook := range m.hooks {
		if err := hook.OnPublish(release); err != nil {
			// Log error but don't fail the publish
			utils.Warn("hook failed: %v", err)
		}
	}

	return nil
}

// PromoteRelease promotes a release from one channel to another
func (m *Manager) PromoteRelease(request *PromotionRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.validatePromotionRequest(request); err != nil {
		return err
	}

	sourceRelease, err := m.getSourceRelease(request)
	if err != nil {
		return err
	}

	if err := m.checkTargetChannel(request); err != nil {
		return err
	}

	if err := m.validateTargetConfig(request); err != nil {
		return err
	}

	promotedRelease := m.createPromotedRelease(sourceRelease, request)

	if err := m.store.SaveRelease(promotedRelease); err != nil {
		return mageErrors.Wrap(err, "failed to save promoted release")
	}

	m.savePromotionRequest(request)
	m.runPromotionHooks(promotedRelease, request.FromChannel)

	return nil
}

// validatePromotionRequest validates the basic promotion request parameters
func (m *Manager) validatePromotionRequest(request *PromotionRequest) error {
	if !request.FromChannel.IsValid() || !request.ToChannel.IsValid() {
		return ErrInvalidChannel
	}

	if !request.FromChannel.CanPromoteTo(request.ToChannel) {
		return fmt.Errorf("%w: from %s to %s", ErrCannotPromote, request.FromChannel, request.ToChannel)
	}

	return nil
}

// getSourceRelease retrieves and validates the source release
func (m *Manager) getSourceRelease(request *PromotionRequest) (*Release, error) {
	sourceRelease, err := m.store.GetRelease(request.FromChannel, request.Version)
	if err != nil {
		return nil, mageErrors.Wrap(err, "failed to get source release")
	}
	if sourceRelease == nil {
		return nil, fmt.Errorf("%w: release %s in channel %s", ErrReleaseNotFoundInChannel, request.Version, request.FromChannel)
	}
	return sourceRelease, nil
}

// checkTargetChannel validates the target channel doesn't already have the version
func (m *Manager) checkTargetChannel(request *PromotionRequest) error {
	if existing, err := m.store.GetRelease(request.ToChannel, request.Version); err == nil && existing != nil {
		return fmt.Errorf("%w: version %s in channel %s", ErrVersionAlreadyExists, request.Version, request.ToChannel)
	}
	return nil
}

// validateTargetConfig gets and validates the target channel configuration
func (m *Manager) validateTargetConfig(request *PromotionRequest) error {
	targetConfig, exists := m.configs[request.ToChannel]
	if !exists {
		return fmt.Errorf("%w: %s", ErrChannelConfigNotFound, request.ToChannel)
	}

	if err := m.validateRequiredTests(targetConfig, request); err != nil {
		return err
	}

	return m.validateApprovers(targetConfig, request)
}

// validateRequiredTests checks if all required tests have passed
func (m *Manager) validateRequiredTests(config *ChannelConfig, request *PromotionRequest) error {
	if len(config.RequiredTests) == 0 {
		return nil
	}

	if len(request.TestResults) == 0 {
		return ErrTestResultsRequired
	}

	testMap := make(map[string]bool)
	for _, result := range request.TestResults {
		if result.Status == "passed" || result.Status == "success" {
			testMap[result.Name] = true
		}
	}

	for _, required := range config.RequiredTests {
		if !testMap[required] {
			return fmt.Errorf("%w: %s", ErrTestFailed, required)
		}
	}

	return nil
}

// validateApprovers checks if approval is required and provided
func (m *Manager) validateApprovers(config *ChannelConfig, request *PromotionRequest) error {
	if len(config.Approvers) > 0 && request.ApprovedBy == "" {
		return ErrApprovalRequired
	}
	return nil
}

// createPromotedRelease creates a new release for the target channel
func (m *Manager) createPromotedRelease(sourceRelease *Release, request *PromotionRequest) *Release {
	return &Release{
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
}

// savePromotionRequest saves the promotion request with timestamp
func (m *Manager) savePromotionRequest(request *PromotionRequest) {
	request.ApprovedAt = &time.Time{}
	*request.ApprovedAt = time.Now()
	if err := m.store.SavePromotionRequest(request); err != nil {
		// Log error but don't fail the promotion
		utils.Warn("failed to save promotion request: %v", err)
	}
}

// runPromotionHooks executes promotion hooks
func (m *Manager) runPromotionHooks(promotedRelease *Release, fromChannel Channel) {
	for _, hook := range m.hooks {
		if err := hook.OnPromote(promotedRelease, fromChannel); err != nil {
			// Log error but don't fail
			utils.Warn("hook failed: %v", err)
		}
	}
}

// GetLatestRelease returns the latest release for a channel
func (m *Manager) GetLatestRelease(channel Channel) (*Release, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	releases, err := m.store.ListReleases(channel)
	if err != nil {
		return nil, mageErrors.Wrap(err, "failed to list releases")
	}

	if len(releases) == 0 {
		return nil, ErrNoReleasesFound
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
		return nil, mageErrors.Wrap(err, "failed to list releases")
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
		return mageErrors.Wrap(err, "failed to get release")
	}
	if release == nil {
		return fmt.Errorf("%w: release %s in channel %s", ErrReleaseNotFoundInChannel, version, channel)
	}

	if release.Deprecated {
		return ErrAlreadyDeprecated
	}

	release.Deprecated = true
	now := time.Now()
	release.DeprecatedAt = &now

	if err := m.store.SaveRelease(release); err != nil {
		return mageErrors.Wrap(err, "failed to save release")
	}

	// Run hooks
	for _, hook := range m.hooks {
		if err := hook.OnDeprecate(release); err != nil {
			// Log error but don't fail
			utils.Warn("hook failed: %v", err)
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
					utils.Info("Cleaned up expired release: %s/%s", channel, release.Version)
				}
			}
		}
	}

	if len(cleanupErrors) > 0 {
		return fmt.Errorf("%w: %d errors, first: %w", ErrCleanupFailed, len(cleanupErrors), cleanupErrors[0])
	}

	return nil
}

// GetChannelStats returns statistics for a channel
func (m *Manager) GetChannelStats(channel Channel) (*ChannelStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	releases, err := m.store.ListReleases(channel)
	if err != nil {
		return nil, mageErrors.Wrap(err, "failed to list releases")
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
