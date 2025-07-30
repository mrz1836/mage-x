package channels

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

func TestNewFileStore(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		tempDir := t.TempDir()
		fileOps := fileops.New()

		store, err := NewFileStore(tempDir, fileOps.File)
		require.NoError(t, err)
		require.NotNil(t, store)

		assert.Equal(t, tempDir, store.baseDir)
		assert.Equal(t, fileOps.File, store.fileOps)

		// Check that directories were created
		expectedDirs := []string{
			"channels", "releases", "releases/stable", "releases/beta",
			"releases/edge", "releases/nightly", "releases/lts",
			"promotions", "configs",
		}

		for _, dir := range expectedDirs {
			dirPath := filepath.Join(tempDir, dir)
			assert.DirExists(t, dirPath, "Directory %s should exist", dir)
		}
	})

	t.Run("initialization failure", func(t *testing.T) {
		// Try to create store in a read-only location
		invalidDir := "/invalid/readonly/path"
		fileOps := fileops.New()

		_, err := NewFileStore(invalidDir, fileOps.File)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestFileStore_Release_Operations(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	validRelease := &Release{
		Version:     "1.0.0",
		Channel:     Stable,
		PublishedAt: time.Now(),
		ReleasedBy:  "developer",
		Changelog:   "Initial release",
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

	t.Run("SaveRelease and GetRelease", func(t *testing.T) {
		// Save release
		err := store.SaveRelease(validRelease)
		require.NoError(t, err)

		// Verify file was created
		expectedPath := filepath.Join(tempDir, "releases", "stable", "1.0.0.json")
		assert.FileExists(t, expectedPath)

		// Get release
		retrieved, err := store.GetRelease(Stable, "1.0.0")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, validRelease.Version, retrieved.Version)
		assert.Equal(t, validRelease.Channel, retrieved.Channel)
		assert.Equal(t, validRelease.ReleasedBy, retrieved.ReleasedBy)
		assert.Equal(t, validRelease.Changelog, retrieved.Changelog)
		assert.Equal(t, len(validRelease.Artifacts), len(retrieved.Artifacts))
	})

	t.Run("GetRelease non-existent", func(t *testing.T) {
		release, err := store.GetRelease(Stable, "2.0.0")
		require.NoError(t, err)
		assert.Nil(t, release)
	})

	t.Run("SaveRelease invalid release", func(t *testing.T) {
		invalidRelease := &Release{
			Version: "", // Invalid - empty version
			Channel: Stable,
		}

		err := store.SaveRelease(invalidRelease)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid release")
	})

	t.Run("version sanitization", func(t *testing.T) {
		unsafeRelease := &Release{
			Version:     "1.0.0/beta:test",
			Channel:     Beta,
			PublishedAt: time.Now(),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}

		err := store.SaveRelease(unsafeRelease)
		require.NoError(t, err)

		// Verify file was created with sanitized filename
		expectedPath := filepath.Join(tempDir, "releases", "beta", "1.0.0-beta-test.json")
		assert.FileExists(t, expectedPath)

		// Should be able to retrieve it
		retrieved, err := store.GetRelease(Beta, "1.0.0/beta:test")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "1.0.0/beta:test", retrieved.Version) // Original version preserved
	})
}

func TestFileStore_ListReleases(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	t.Run("empty channel", func(t *testing.T) {
		releases, err := store.ListReleases(Edge)
		require.NoError(t, err)
		assert.Len(t, releases, 0)
	})

	t.Run("channel with releases", func(t *testing.T) {
		// Create multiple releases
		releases := []*Release{
			{
				Version:     "1.0.0",
				Channel:     Stable,
				PublishedAt: time.Now(),
				ReleasedBy:  "dev1",
				Artifacts:   []Artifact{{Name: "binary1", URL: "https://example.com/binary1", Checksum: "abc123"}},
			},
			{
				Version:     "1.1.0",
				Channel:     Stable,
				PublishedAt: time.Now().Add(time.Hour),
				ReleasedBy:  "dev2",
				Artifacts:   []Artifact{{Name: "binary2", URL: "https://example.com/binary2", Checksum: "def456"}},
			},
		}

		for _, release := range releases {
			err := store.SaveRelease(release)
			require.NoError(t, err)
		}

		// List releases
		listed, err := store.ListReleases(Stable)
		require.NoError(t, err)
		assert.Len(t, listed, 2)

		// Verify contents (order may vary)
		versions := make([]string, len(listed))
		for i, release := range listed {
			versions[i] = release.Version
		}
		assert.Contains(t, versions, "1.0.0")
		assert.Contains(t, versions, "1.1.0")
	})

	t.Run("ignore non-json files", func(t *testing.T) {
		// Create a non-JSON file in the channel directory
		channelDir := filepath.Join(tempDir, "releases", "beta")
		err := os.MkdirAll(channelDir, 0o755)
		require.NoError(t, err)

		nonJSONFile := filepath.Join(channelDir, "README.txt")
		err = os.WriteFile(nonJSONFile, []byte("This is not a release"), 0o644)
		require.NoError(t, err)

		// Should ignore the non-JSON file
		releases, err := store.ListReleases(Beta)
		require.NoError(t, err)
		assert.Len(t, releases, 0)
	})

	t.Run("handle corrupted files gracefully", func(t *testing.T) {
		// Create a valid release first
		validRelease := &Release{
			Version:     "1.0.0",
			Channel:     Nightly,
			PublishedAt: time.Now(),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}
		err := store.SaveRelease(validRelease)
		require.NoError(t, err)

		// Create a corrupted JSON file
		channelDir := filepath.Join(tempDir, "releases", "nightly")
		corruptedFile := filepath.Join(channelDir, "corrupted.json")
		err = os.WriteFile(corruptedFile, []byte("invalid json content"), 0o644)
		require.NoError(t, err)

		// Should handle corrupted file gracefully and return valid releases
		releases, err := store.ListReleases(Nightly)
		require.NoError(t, err)
		assert.Len(t, releases, 1) // Only the valid release
		assert.Equal(t, "1.0.0", releases[0].Version)
	})
}

func TestFileStore_DeleteRelease(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	// Create a test release
	release := &Release{
		Version:     "1.0.0",
		Channel:     Stable,
		PublishedAt: time.Now(),
		ReleasedBy:  "developer",
		Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
	}
	err = store.SaveRelease(release)
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		err := store.DeleteRelease(Stable, "1.0.0")
		require.NoError(t, err)

		// Verify file was deleted
		expectedPath := filepath.Join(tempDir, "releases", "stable", "1.0.0.json")
		assert.NoFileExists(t, expectedPath)

		// Verify release is no longer retrievable
		retrieved, err := store.GetRelease(Stable, "1.0.0")
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent release", func(t *testing.T) {
		err := store.DeleteRelease(Stable, "2.0.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestFileStore_ChannelConfig_Operations(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	config := &ChannelConfig{
		Name:          Beta,
		Description:   "Beta testing channel",
		PromotionPath: []Channel{Stable},
		RetentionDays: 30,
		AutoPromotion: false,
		RequiredTests: []string{"unit", "integration"},
		Approvers:     []string{"admin1", "admin2"},
		WebhookURL:    "https://hooks.example.com/webhook",
		Metadata: map[string]string{
			"owner": "platform-team",
		},
	}

	t.Run("SaveChannelConfig and GetChannelConfig", func(t *testing.T) {
		// Save config
		err := store.SaveChannelConfig(config)
		require.NoError(t, err)

		// Verify file was created
		expectedPath := filepath.Join(tempDir, "configs", "beta.json")
		assert.FileExists(t, expectedPath)

		// Get config
		retrieved, err := store.GetChannelConfig(Beta)
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, config.Name, retrieved.Name)
		assert.Equal(t, config.Description, retrieved.Description)
		assert.Equal(t, config.PromotionPath, retrieved.PromotionPath)
		assert.Equal(t, config.RetentionDays, retrieved.RetentionDays)
		assert.Equal(t, config.AutoPromotion, retrieved.AutoPromotion)
		assert.Equal(t, config.RequiredTests, retrieved.RequiredTests)
		assert.Equal(t, config.Approvers, retrieved.Approvers)
		assert.Equal(t, config.WebhookURL, retrieved.WebhookURL)
		assert.Equal(t, config.Metadata, retrieved.Metadata)
	})

	t.Run("GetChannelConfig non-existent", func(t *testing.T) {
		config, err := store.GetChannelConfig(LTS)
		require.NoError(t, err)
		assert.Nil(t, config)
	})
}

func TestFileStore_PromotionHistory_Operations(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	promotionRequest := &PromotionRequest{
		Version:     "1.0.0",
		FromChannel: Beta,
		ToChannel:   Stable,
		RequestedBy: "admin",
		RequestedAt: time.Now(),
		ApprovedBy:  "senior-admin",
		ApprovedAt:  func() *time.Time { t := time.Now().Add(time.Hour); return &t }(),
		TestResults: []TestResult{
			{
				Name:     "unit",
				Status:   "passed",
				Duration: 5000,
				Output:   "All tests passed",
				RunAt:    time.Now(),
			},
		},
		Notes: "Ready for stable release",
		Metadata: map[string]string{
			"pr": "1234",
		},
	}

	t.Run("SavePromotionRequest and GetPromotionHistory", func(t *testing.T) {
		// Save promotion request
		err := store.SavePromotionRequest(promotionRequest)
		require.NoError(t, err)

		// Verify file was created
		expectedPath := filepath.Join(tempDir, "promotions", "1.0.0.json")
		assert.FileExists(t, expectedPath)

		// Get promotion history
		history, err := store.GetPromotionHistory("1.0.0")
		require.NoError(t, err)
		require.Len(t, history, 1)

		retrieved := history[0]
		assert.Equal(t, promotionRequest.Version, retrieved.Version)
		assert.Equal(t, promotionRequest.FromChannel, retrieved.FromChannel)
		assert.Equal(t, promotionRequest.ToChannel, retrieved.ToChannel)
		assert.Equal(t, promotionRequest.RequestedBy, retrieved.RequestedBy)
		assert.Equal(t, promotionRequest.ApprovedBy, retrieved.ApprovedBy)
		assert.Equal(t, promotionRequest.Notes, retrieved.Notes)
		assert.Equal(t, promotionRequest.Metadata, retrieved.Metadata)
		assert.Len(t, retrieved.TestResults, 1)
	})

	t.Run("multiple promotion requests", func(t *testing.T) {
		// Add another promotion request for the same version
		secondRequest := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Stable,
			ToChannel:   LTS,
			RequestedBy: "release-manager",
			RequestedAt: time.Now().Add(2 * time.Hour),
			Notes:       "Promoting to LTS",
		}

		err := store.SavePromotionRequest(secondRequest)
		require.NoError(t, err)

		// Get promotion history - should have both requests
		history, err := store.GetPromotionHistory("1.0.0")
		require.NoError(t, err)
		require.Len(t, history, 2)

		// Verify both requests are present
		versions := make([]string, len(history))
		fromChannels := make([]Channel, len(history))
		toChannels := make([]Channel, len(history))

		for i, req := range history {
			versions[i] = req.Version
			fromChannels[i] = req.FromChannel
			toChannels[i] = req.ToChannel
		}

		// Both should be for version 1.0.0
		assert.Equal(t, "1.0.0", versions[0])
		assert.Equal(t, "1.0.0", versions[1])

		// Should have both promotion paths
		assert.Contains(t, fromChannels, Beta)
		assert.Contains(t, fromChannels, Stable)
		assert.Contains(t, toChannels, Stable)
		assert.Contains(t, toChannels, LTS)
	})

	t.Run("GetPromotionHistory non-existent", func(t *testing.T) {
		history, err := store.GetPromotionHistory("2.0.0")
		require.NoError(t, err)
		assert.Len(t, history, 0)
	})

	t.Run("version sanitization in promotion history", func(t *testing.T) {
		unsafeRequest := &PromotionRequest{
			Version:     "1.0.0/beta:special",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "admin",
			RequestedAt: time.Now(),
		}

		err := store.SavePromotionRequest(unsafeRequest)
		require.NoError(t, err)

		// Verify file was created with sanitized filename
		expectedPath := filepath.Join(tempDir, "promotions", "1.0.0-beta-special.json")
		assert.FileExists(t, expectedPath)

		// Should be able to retrieve it with original version
		history, err := store.GetPromotionHistory("1.0.0/beta:special")
		require.NoError(t, err)
		require.Len(t, history, 1)
		assert.Equal(t, "1.0.0/beta:special", history[0].Version)
	})
}

func TestFileStore_IndexUpdates(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	t.Run("index updated on save", func(t *testing.T) {
		release := &Release{
			Version:     "1.0.0",
			Channel:     Stable,
			PublishedAt: time.Now(),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}

		err := store.SaveRelease(release)
		require.NoError(t, err)

		// Check that index file was created
		indexPath := filepath.Join(tempDir, "channels", "stable-index.json")
		assert.FileExists(t, indexPath)

		// Read and verify index content
		data, err := os.ReadFile(indexPath)
		require.NoError(t, err)

		indexContent := string(data)
		assert.Contains(t, indexContent, "1.0.0")
		assert.Contains(t, indexContent, "false") // Not deprecated
	})

	t.Run("index updated on delete", func(t *testing.T) {
		// Add another release
		release2 := &Release{
			Version:     "1.1.0",
			Channel:     Stable,
			PublishedAt: time.Now().Add(time.Hour),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "def456"}},
		}

		err := store.SaveRelease(release2)
		require.NoError(t, err)

		// Delete the first release
		err = store.DeleteRelease(Stable, "1.0.0")
		require.NoError(t, err)

		// Index should be updated
		indexPath := filepath.Join(tempDir, "channels", "stable-index.json")
		data, err := os.ReadFile(indexPath)
		require.NoError(t, err)

		indexContent := string(data)
		assert.NotContains(t, indexContent, "1.0.0") // Should be removed
		assert.Contains(t, indexContent, "1.1.0")    // Should still be there
	})
}

func TestSanitizeVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"clean version", "1.0.0", "1.0.0"},
		{"with slash", "1.0.0/beta", "1.0.0-beta"},
		{"with backslash", "1.0.0\\beta", "1.0.0-beta"},
		{"with colon", "1.0.0:beta", "1.0.0-beta"},
		{"with asterisk", "1.0.*", "1.0.-"},
		{"with question mark", "1.0.?", "1.0.-"},
		{"with quotes", "1.0.0\"beta\"", "1.0.0-beta-"},
		{"with angle brackets", "1.0.0<beta>", "1.0.0-beta-"},
		{"with pipe", "1.0.0|beta", "1.0.0-beta"},
		{"multiple special chars", "1.0.0/beta:test*", "1.0.0-beta-test-"},
		{"only special chars", "/*?", "---"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeVersion(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasJSONExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"json file", "release.json", true},
		{"JSON file uppercase", "release.JSON", false}, // Case sensitive
		{"text file", "readme.txt", false},
		{"no extension", "release", false},
		{"multiple dots", "release.1.0.json", true},
		{"dot json in middle", "release.json.txt", false},
		{"empty string", "", false},
		{"just extension", ".json", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasJSONExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileStore_getReleasePath(t *testing.T) {
	tempDir := t.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	require.NoError(t, err)

	tests := []struct {
		name     string
		channel  Channel
		version  string
		expected string
	}{
		{
			"simple version",
			Stable,
			"1.0.0",
			filepath.Join(tempDir, "releases", "stable", "1.0.0.json"),
		},
		{
			"version with special chars",
			Beta,
			"1.0.0/beta:test",
			filepath.Join(tempDir, "releases", "beta", "1.0.0-beta-test.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := store.getReleasePath(tt.channel, tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkFileStore_SaveRelease(b *testing.B) {
	tempDir := b.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	if err != nil {
		b.Fatal(err)
	}

	release := &Release{
		Version:     "1.0.0",
		Channel:     Stable,
		PublishedAt: time.Now(),
		ReleasedBy:  "developer",
		Changelog:   "Benchmark release",
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

		err := store.SaveRelease(&testRelease)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileStore_GetRelease(b *testing.B) {
	tempDir := b.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate with releases
	for i := 0; i < 100; i++ {
		release := &Release{
			Version:     fmt.Sprintf("1.0.%d", i),
			Channel:     Stable,
			PublishedAt: time.Now(),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}
		if err := store.SaveRelease(release); err != nil {
			b.Fatalf("Failed to save release for benchmark setup: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version := fmt.Sprintf("1.0.%d", i%100)
		_, err := store.GetRelease(Stable, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileStore_ListReleases(b *testing.B) {
	tempDir := b.TempDir()
	fileOps := fileops.New()
	store, err := NewFileStore(tempDir, fileOps.File)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate with releases
	for i := 0; i < 50; i++ {
		release := &Release{
			Version:     fmt.Sprintf("1.0.%d", i),
			Channel:     Stable,
			PublishedAt: time.Now(),
			ReleasedBy:  "developer",
			Artifacts:   []Artifact{{Name: "binary", URL: "https://example.com/binary", Checksum: "abc123"}},
		}
		if err := store.SaveRelease(release); err != nil {
			b.Fatalf("Failed to save release for benchmark setup: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.ListReleases(Stable)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSanitizeVersion(b *testing.B) {
	versions := []string{
		"1.0.0",
		"1.0.0/beta",
		"1.0.0:alpha*test",
		"complex/version:with*many?special<chars>",
		"simple",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version := versions[i%len(versions)]
		_ = sanitizeVersion(version)
	}
}
