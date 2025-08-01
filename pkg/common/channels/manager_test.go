package channels

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mageErrors "github.com/mrz1836/go-mage/pkg/common/errors"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// Test-specific static errors
var errMockHookFailed = errors.New("hook error")

// MockStore implements Store interface for testing
type MockStore struct {
	releases    map[string]map[string]*Release // channel -> version -> release
	configs     map[Channel]*ChannelConfig
	promotions  map[string][]*PromotionRequest // version -> requests
	shouldError bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		releases:   make(map[string]map[string]*Release),
		configs:    make(map[Channel]*ChannelConfig),
		promotions: make(map[string][]*PromotionRequest),
	}
}

func (m *MockStore) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockStore) GetRelease(channel Channel, version string) (*Release, error) {
	if m.shouldError {
		return nil, ErrMockOperation
	}

	channelReleases, exists := m.releases[channel.String()]
	if !exists {
		return nil, ErrReleaseNotFound
	}

	release, exists := channelReleases[version]
	if !exists {
		return nil, ErrReleaseNotFound
	}

	return release, nil
}

func (m *MockStore) ListReleases(channel Channel) ([]*Release, error) {
	if m.shouldError {
		return nil, ErrMockOperation
	}

	channelReleases, exists := m.releases[channel.String()]
	if !exists {
		return []*Release{}, nil
	}

	releases := make([]*Release, 0, len(channelReleases))
	for _, release := range channelReleases {
		releases = append(releases, release)
	}

	return releases, nil
}

func (m *MockStore) SaveRelease(release *Release) error {
	if m.shouldError {
		return ErrMockOperation
	}

	if err := release.Validate(); err != nil {
		return err
	}

	channelKey := release.Channel.String()
	if _, exists := m.releases[channelKey]; !exists {
		m.releases[channelKey] = make(map[string]*Release)
	}

	m.releases[channelKey][release.Version] = release
	return nil
}

func (m *MockStore) DeleteRelease(channel Channel, version string) error {
	if m.shouldError {
		return ErrMockOperation
	}

	channelReleases, exists := m.releases[channel.String()]
	if !exists {
		return mageErrors.WithCode(mageErrors.ErrNotFound, "release not found")
	}

	if _, exists := channelReleases[version]; !exists {
		return mageErrors.WithCode(mageErrors.ErrNotFound, "release not found")
	}

	delete(channelReleases, version)
	return nil
}

func (m *MockStore) GetChannelConfig(channel Channel) (*ChannelConfig, error) {
	if m.shouldError {
		return nil, ErrMockOperation
	}

	config, exists := m.configs[channel]
	if !exists {
		return nil, ErrConfigNotFound
	}

	return config, nil
}

func (m *MockStore) SaveChannelConfig(config *ChannelConfig) error {
	if m.shouldError {
		return ErrMockOperation
	}

	m.configs[config.Name] = config
	return nil
}

func (m *MockStore) GetPromotionHistory(version string) ([]*PromotionRequest, error) {
	if m.shouldError {
		return nil, ErrMockOperation
	}

	history, exists := m.promotions[version]
	if !exists {
		return []*PromotionRequest{}, nil
	}

	return history, nil
}

func (m *MockStore) SavePromotionRequest(request *PromotionRequest) error {
	if m.shouldError {
		return ErrMockOperation
	}

	if _, exists := m.promotions[request.Version]; !exists {
		m.promotions[request.Version] = []*PromotionRequest{}
	}

	m.promotions[request.Version] = append(m.promotions[request.Version], request)
	return nil
}

// MockValidator for testing
type MockValidator struct {
	shouldFail bool
	errorMsg   string
}

func (v *MockValidator) Validate(_ *Release) error {
	if v.shouldFail {
		return fmt.Errorf("%w: %s", ErrMockOperation, v.errorMsg)
	}
	return nil
}

// MockHook for testing
type MockHook struct {
	publishCalled   bool
	promoteCalled   bool
	deprecateCalled bool
	shouldFail      bool
}

func (h *MockHook) OnPublish(_ *Release) error {
	h.publishCalled = true
	if h.shouldFail {
		return errMockHookFailed
	}
	return nil
}

func (h *MockHook) OnPromote(_ *Release, _ Channel) error {
	h.promoteCalled = true
	if h.shouldFail {
		return errMockHookFailed
	}
	return nil
}

func (h *MockHook) OnDeprecate(_ *Release) error {
	h.deprecateCalled = true
	if h.shouldFail {
		return errMockHookFailed
	}
	return nil
}

func TestNewManager(t *testing.T) {
	store := NewMockStore()
	fileOps := fileops.New().File

	manager := NewManager(store, fileOps)

	assert.NotNil(t, manager)
	assert.Equal(t, store, manager.store)
	assert.Equal(t, fileOps, manager.fileOps)
	assert.NotNil(t, manager.configs)
	assert.NotNil(t, manager.validators)
	assert.NotNil(t, manager.hooks)
}

func TestManager_Initialize(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)

		err := manager.Initialize()
		require.NoError(t, err)

		// Check that default configs were created
		expectedChannels := []Channel{Edge, Nightly, Beta, Stable, LTS}
		for _, channel := range expectedChannels {
			config, exists := manager.configs[channel]
			assert.True(t, exists, "Config should exist for channel %s", channel)
			assert.Equal(t, channel, config.Name)
			assert.NotEmpty(t, config.Description)
		}
	})

	t.Run("store error", func(t *testing.T) {
		store := NewMockStore()
		store.SetError(true)
		manager := NewManager(store, fileops.New().File)

		err := manager.Initialize()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save channel config")
	})

	t.Run("existing configs preserved", func(t *testing.T) {
		store := NewMockStore()

		// Pre-populate with custom config
		customConfig := &ChannelConfig{
			Name:        Beta,
			Description: "Custom beta config",
		}
		require.NoError(t, store.SaveChannelConfig(customConfig))

		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		// Should find existing config and use it instead of default
		retrievedConfig := manager.configs[Beta]
		assert.Equal(t, customConfig.Description, retrievedConfig.Description)
		assert.Equal(t, customConfig.Name, retrievedConfig.Name)
	})
}

func TestManager_PublishRelease(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)
	err := manager.Initialize()
	require.NoError(t, err)

	validRelease := &Release{
		Version:    "1.0.0",
		Channel:    Stable,
		ReleasedBy: "developer",
		Changelog:  "Initial release",
		Artifacts: []Artifact{
			{
				Name:        "binary-linux-amd64",
				Platform:    "linux",
				Arch:        "amd64",
				URL:         "https://example.com/binary",
				Size:        1024,
				Checksum:    "abc123",
				ChecksumAlg: "sha256",
			},
		},
	}

	t.Run("successful publish", func(t *testing.T) {
		release := *validRelease // Copy

		err := manager.PublishRelease(&release)
		require.NoError(t, err)

		// Check that timestamp was set
		assert.False(t, release.PublishedAt.IsZero())

		// Verify it was saved
		saved, err := store.GetRelease(Stable, "1.0.0")
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, "1.0.0", saved.Version)
	})

	t.Run("invalid channel", func(t *testing.T) {
		release := *validRelease
		release.Channel = Channel("invalid")

		err := manager.PublishRelease(&release)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid channel")
	})

	t.Run("invalid release", func(t *testing.T) {
		release := *validRelease
		release.Version = "" // Invalid

		err := manager.PublishRelease(&release)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid release")
	})

	t.Run("custom validator failure", func(t *testing.T) {
		validator := &MockValidator{
			shouldFail: true,
			errorMsg:   "custom validation failed",
		}
		manager.AddValidator(validator)

		release := *validRelease

		err := manager.PublishRelease(&release)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "custom validation failed")
	})

	t.Run("version already exists", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		// First publish should succeed
		release1 := *validRelease
		err = manager.PublishRelease(&release1)
		require.NoError(t, err)

		// Second publish with same version should fail
		release2 := *validRelease
		err = manager.PublishRelease(&release2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("store error", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		store.SetError(true)

		release := *validRelease
		release.Version = "1.1.0" // Different version

		err = manager.PublishRelease(&release)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save release")
	})

	t.Run("hook called", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		hook := &MockHook{}
		manager.AddHook(hook)

		release := *validRelease
		release.Version = "1.2.0"

		err = manager.PublishRelease(&release)
		require.NoError(t, err)

		assert.True(t, hook.publishCalled)
	})

	t.Run("hook error doesn't fail publish", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		hook := &MockHook{shouldFail: true}
		manager.AddHook(hook)

		release := *validRelease
		release.Version = "1.3.0"

		err = manager.PublishRelease(&release)
		assert.NoError(t, err) // Should not fail even with hook error
	})
}

func TestManager_PromoteRelease(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)
	err := manager.Initialize()
	require.NoError(t, err)

	// Create a source release
	sourceRelease := &Release{
		Version:    "1.0.0",
		Channel:    Beta,
		ReleasedBy: "developer",
		Changelog:  "Beta release",
		Artifacts: []Artifact{
			{
				Name:        "binary-linux-amd64",
				Platform:    "linux",
				Arch:        "amd64",
				URL:         "https://example.com/binary",
				Size:        1024,
				Checksum:    "abc123",
				ChecksumAlg: "sha256",
			},
		},
	}
	err = manager.PublishRelease(sourceRelease)
	require.NoError(t, err)

	t.Run("successful promotion", func(t *testing.T) {
		request := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
			TestResults: []TestResult{
				{Name: "unit", Status: "passed"},
				{Name: "integration", Status: "passed"},
				{Name: "smoke", Status: "passed"},
				{Name: "performance", Status: "passed"},
				{Name: "security", Status: "passed"},
			},
		}

		err := manager.PromoteRelease(request)
		require.NoError(t, err)

		// Check that promoted release exists
		promoted, err := store.GetRelease(Stable, "1.0.0")
		require.NoError(t, err)
		require.NotNil(t, promoted)

		assert.Equal(t, Stable, promoted.Channel)
		assert.True(t, promoted.Promoted)
		assert.Equal(t, Beta, promoted.PromotedFrom)
		assert.NotNil(t, promoted.PromotedAt)
	})

	t.Run("invalid channels", func(t *testing.T) {
		request := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Channel("invalid"),
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err := manager.PromoteRelease(request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid channel")
	})

	t.Run("invalid promotion path", func(t *testing.T) {
		request := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Stable,
			ToChannel:   Beta, // Cannot promote backwards
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err := manager.PromoteRelease(request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot promote between channels")
	})

	t.Run("source release not found", func(t *testing.T) {
		request := &PromotionRequest{
			Version:     "2.0.0", // Doesn't exist
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err := manager.PromoteRelease(request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("target already exists", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		// First publish a release to Beta channel
		betaRelease := &Release{
			Version:     "1.0.0",
			Channel:     Beta,
			ReleasedBy:  "developer",
			Changelog:   "Test release",
			Artifacts:   []Artifact{{Name: "test", Platform: "linux", Arch: "amd64", URL: "test", Size: 100, Checksum: "abc", ChecksumAlg: "sha256"}},
			PublishedAt: time.Now(),
		}
		err = manager.PublishRelease(betaRelease)
		require.NoError(t, err)

		// First promotion should succeed
		request1 := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
			TestResults: []TestResult{
				{Name: "unit", Status: "passed"},
				{Name: "integration", Status: "passed"},
				{Name: "smoke", Status: "passed"},
				{Name: "performance", Status: "passed"},
				{Name: "security", Status: "passed"},
			},
		}
		err = manager.PromoteRelease(request1)
		require.NoError(t, err)

		// Second promotion to same target should fail
		request2 := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err = manager.PromoteRelease(request2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("missing required tests", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		// First publish a release to Beta channel
		betaRelease := &Release{
			Version:     "2.0.0", // Different version
			Channel:     Beta,
			ReleasedBy:  "developer",
			Changelog:   "Test release",
			Artifacts:   []Artifact{{Name: "test", Platform: "linux", Arch: "amd64", URL: "test", Size: 100, Checksum: "abc", ChecksumAlg: "sha256"}},
			PublishedAt: time.Now(),
		}
		err = manager.PublishRelease(betaRelease)
		require.NoError(t, err)

		request := &PromotionRequest{
			Version:     "2.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
			// Missing test results
		}

		err = manager.PromoteRelease(request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test results required")
	})

	t.Run("failed required test", func(t *testing.T) {
		// Create fresh manager for this test
		store := NewMockStore()
		manager := NewManager(store, fileops.New().File)
		err := manager.Initialize()
		require.NoError(t, err)

		// First publish a release to Beta channel
		betaRelease := &Release{
			Version:     "3.0.0", // Different version
			Channel:     Beta,
			ReleasedBy:  "developer",
			Changelog:   "Test release",
			Artifacts:   []Artifact{{Name: "test", Platform: "linux", Arch: "amd64", URL: "test", Size: 100, Checksum: "abc", ChecksumAlg: "sha256"}},
			PublishedAt: time.Now(),
		}
		err = manager.PublishRelease(betaRelease)
		require.NoError(t, err)

		request := &PromotionRequest{
			Version:     "3.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
			TestResults: []TestResult{
				{Name: "unit", Status: "passed"},
				{Name: "integration", Status: "failed"}, // Failed test
			},
		}

		err = manager.PromoteRelease(request)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required test not passed")
	})
}

func TestManager_GetLatestRelease(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	t.Run("no releases", func(t *testing.T) {
		latest, err := manager.GetLatestRelease(Stable)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoReleasesFound)
		assert.Nil(t, latest)
	})

	t.Run("single release", func(t *testing.T) {
		release := &Release{
			Version:     "1.0.0",
			Channel:     Stable,
			PublishedAt: time.Now(),
			Artifacts: []Artifact{
				{
					Name:     "binary",
					URL:      "https://example.com/binary",
					Checksum: "abc123",
				},
			},
		}

		err := store.SaveRelease(release)
		require.NoError(t, err)

		latest, err := manager.GetLatestRelease(Stable)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, "1.0.0", latest.Version)
	})

	t.Run("multiple releases", func(t *testing.T) {
		now := time.Now()

		releases := []*Release{
			{
				Version:     "1.0.0",
				Channel:     Beta,
				PublishedAt: now.Add(-2 * time.Hour),
				Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
			},
			{
				Version:     "1.1.0",
				Channel:     Beta,
				PublishedAt: now.Add(-1 * time.Hour), // More recent
				Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
			},
			{
				Version:     "0.9.0",
				Channel:     Beta,
				PublishedAt: now.Add(-3 * time.Hour),
				Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "ghi789"}},
			},
		}

		for _, release := range releases {
			err := store.SaveRelease(release)
			require.NoError(t, err)
		}

		latest, err := manager.GetLatestRelease(Beta)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, "1.1.0", latest.Version) // Should be the most recent
	})

	t.Run("store error", func(t *testing.T) {
		store.SetError(true)

		_, err := manager.GetLatestRelease(Stable)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list releases")

		store.SetError(false)
	})
}

func TestManager_ListReleases(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	// Create test releases
	releases := []*Release{
		{
			Version:     "1.0.0",
			Channel:     Stable,
			PublishedAt: time.Now(),
			Deprecated:  false,
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		},
		{
			Version:     "0.9.0",
			Channel:     Stable,
			PublishedAt: time.Now().Add(-time.Hour),
			Deprecated:  true,
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
		},
	}

	for _, release := range releases {
		err := store.SaveRelease(release)
		require.NoError(t, err)
	}

	t.Run("include deprecated", func(t *testing.T) {
		list, err := manager.ListReleases(Stable, true)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("exclude deprecated", func(t *testing.T) {
		list, err := manager.ListReleases(Stable, false)
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "1.0.0", list[0].Version)
		assert.False(t, list[0].Deprecated)
	})

	t.Run("empty channel", func(t *testing.T) {
		list, err := manager.ListReleases(Edge, false)
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("store error", func(t *testing.T) {
		store.SetError(true)

		_, err := manager.ListReleases(Stable, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list releases")

		store.SetError(false)
	})
}

func TestManager_DeprecateRelease(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	// Create a test release
	release := &Release{
		Version:     "1.0.0",
		Channel:     Stable,
		PublishedAt: time.Now(),
		Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
	}
	err := store.SaveRelease(release)
	require.NoError(t, err)

	t.Run("successful deprecation", func(t *testing.T) {
		err := manager.DeprecateRelease(Stable, "1.0.0")
		require.NoError(t, err)

		// Verify it was deprecated
		deprecated, err := store.GetRelease(Stable, "1.0.0")
		require.NoError(t, err)
		require.NotNil(t, deprecated)
		assert.True(t, deprecated.Deprecated)
		assert.NotNil(t, deprecated.DeprecatedAt)
	})

	t.Run("release not found", func(t *testing.T) {
		err := manager.DeprecateRelease(Stable, "2.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("already deprecated", func(t *testing.T) {
		// Release should already be deprecated from previous test
		err := manager.DeprecateRelease(Stable, "1.0.0")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already deprecated")
	})

	t.Run("hook called", func(t *testing.T) {
		// Create a new release for this test
		newRelease := &Release{
			Version:     "1.1.0",
			Channel:     Stable,
			PublishedAt: time.Now(),
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
		}
		err := store.SaveRelease(newRelease)
		require.NoError(t, err)

		hook := &MockHook{}
		manager.AddHook(hook)

		err = manager.DeprecateRelease(Stable, "1.1.0")
		require.NoError(t, err)

		assert.True(t, hook.deprecateCalled)
	})
}

func TestManager_CleanupExpiredReleases(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)
	err := manager.Initialize()
	require.NoError(t, err)

	now := time.Now()

	// Create releases with different ages
	releases := []*Release{
		{
			Version:     "1.0.0",
			Channel:     Edge,
			PublishedAt: now.Add(-10 * 24 * time.Hour), // 10 days ago (should be expired for Edge)
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		},
		{
			Version:     "1.1.0",
			Channel:     Edge,
			PublishedAt: now.Add(-5 * 24 * time.Hour), // 5 days ago (should not be expired for Edge)
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
		},
	}

	for _, release := range releases {
		err := store.SaveRelease(release)
		require.NoError(t, err)
	}

	t.Run("successful cleanup", func(t *testing.T) {
		err := manager.CleanupExpiredReleases()
		require.NoError(t, err)

		// Check that expired release was removed
		expired, err := store.GetRelease(Edge, "1.0.0")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrReleaseNotFound)
		assert.Nil(t, expired)

		// Check that non-expired release still exists
		current, err := store.GetRelease(Edge, "1.1.0")
		require.NoError(t, err)
		assert.NotNil(t, current)
	})

	t.Run("cleanup errors", func(t *testing.T) {
		// Create another expired release
		expiredRelease := &Release{
			Version:     "1.2.0",
			Channel:     Edge,
			PublishedAt: now.Add(-15 * 24 * time.Hour),
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "ghi789"}},
		}
		err := store.SaveRelease(expiredRelease)
		require.NoError(t, err)

		// Set store to error when deleting
		store.SetError(true)

		err = manager.CleanupExpiredReleases()
		require.Error(t, err)

		store.SetError(false)
	})
}

func TestManager_GetChannelStats(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	t.Run("empty channel", func(t *testing.T) {
		stats, err := manager.GetChannelStats(Stable)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, Stable, stats.Channel)
		assert.Equal(t, 0, stats.TotalReleases)
		assert.Empty(t, stats.LatestVersion)
	})

	t.Run("channel with releases", func(t *testing.T) {
		now := time.Now()

		releases := []*Release{
			{
				Version:     "1.0.0",
				Channel:     Beta,
				PublishedAt: now.Add(-2 * time.Hour),
				Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
			},
			{
				Version:     "1.1.0",
				Channel:     Beta,
				PublishedAt: now.Add(-1 * time.Hour), // Most recent
				Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
			},
		}

		for _, release := range releases {
			err := store.SaveRelease(release)
			require.NoError(t, err)
		}

		stats, err := manager.GetChannelStats(Beta)
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, Beta, stats.Channel)
		assert.Equal(t, 2, stats.TotalReleases)
		assert.Equal(t, "1.1.0", stats.LatestVersion)
		assert.WithinDuration(t, now.Add(-time.Hour), stats.LatestRelease, time.Minute)
	})

	t.Run("store error", func(t *testing.T) {
		store.SetError(true)

		_, err := manager.GetChannelStats(Stable)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list releases")

		store.SetError(false)
	})
}

func TestManager_AddValidator(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	validator := &MockValidator{}
	manager.AddValidator(validator)

	assert.Len(t, manager.validators, 1)
	assert.Equal(t, validator, manager.validators[0])
}

func TestManager_AddHook(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	hook := &MockHook{}
	manager.AddHook(hook)

	assert.Len(t, manager.hooks, 1)
	assert.Equal(t, hook, manager.hooks[0])
}

func TestManager_GetPromotionHistory(t *testing.T) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	t.Run("no history", func(t *testing.T) {
		history, err := manager.GetPromotionHistory("1.0.0")
		require.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("with history", func(t *testing.T) {
		request := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err := store.SavePromotionRequest(request)
		require.NoError(t, err)

		history, err := manager.GetPromotionHistory("1.0.0")
		require.NoError(t, err)
		assert.Len(t, history, 1)
		assert.Equal(t, request.Version, history[0].Version)
	})

	t.Run("store error", func(t *testing.T) {
		store.SetError(true)

		_, err := manager.GetPromotionHistory("1.0.0")
		require.Error(t, err)

		store.SetError(false)
	})
}

// Benchmark tests
func BenchmarkManager_PublishRelease(b *testing.B) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)
	if err := manager.Initialize(); err != nil {
		b.Fatal(err)
	}

	release := &Release{
		Version:    "1.0.0",
		Channel:    Stable,
		ReleasedBy: "developer",
		Changelog:  "Benchmark release",
		Artifacts: []Artifact{
			{
				Name:        "binary-linux-amd64",
				Platform:    "linux",
				Arch:        "amd64",
				URL:         "https://example.com/binary",
				Size:        1024,
				Checksum:    "abc123",
				ChecksumAlg: "sha256",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testRelease := *release
		testRelease.Version = fmt.Sprintf("1.0.%d", i)

		err := manager.PublishRelease(&testRelease)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManager_GetLatestRelease(b *testing.B) {
	store := NewMockStore()
	manager := NewManager(store, fileops.New().File)

	// Pre-populate with releases
	for i := 0; i < 100; i++ {
		release := &Release{
			Version:     fmt.Sprintf("1.0.%d", i),
			Channel:     Stable,
			PublishedAt: time.Now().Add(time.Duration(i) * time.Minute),
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}
		if err := store.SaveRelease(release); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetLatestRelease(Stable)
		if err != nil {
			b.Fatal(err)
		}
	}
}
