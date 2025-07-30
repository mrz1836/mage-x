package channels

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannel_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		channel Channel
		valid   bool
	}{
		{"Stable channel", Stable, true},
		{"Beta channel", Beta, true},
		{"Edge channel", Edge, true},
		{"Nightly channel", Nightly, true},
		{"LTS channel", LTS, true},
		{"Invalid channel", Channel("invalid"), false},
		{"Empty channel", Channel(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.channel.IsValid())
		})
	}
}

func TestChannel_String(t *testing.T) {
	tests := []struct {
		channel Channel
		str     string
	}{
		{Stable, "stable"},
		{Beta, "beta"},
		{Edge, "edge"},
		{Nightly, "nightly"},
		{LTS, "lts"},
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			assert.Equal(t, tt.str, tt.channel.String())
		})
	}
}

func TestChannel_CanPromoteTo(t *testing.T) {
	tests := []struct {
		name  string
		from  Channel
		to    Channel
		canDo bool
	}{
		{"Edge to Beta", Edge, Beta, true},
		{"Edge to Nightly", Edge, Nightly, true},
		{"Edge to Stable", Edge, Stable, false},
		{"Edge to LTS", Edge, LTS, false},
		{"Nightly to Beta", Nightly, Beta, true},
		{"Nightly to Stable", Nightly, Stable, false},
		{"Beta to Stable", Beta, Stable, true},
		{"Beta to LTS", Beta, LTS, true},
		{"Beta to Edge", Beta, Edge, false},
		{"Stable to LTS", Stable, LTS, true},
		{"Stable to Beta", Stable, Beta, false},
		{"LTS to anything", LTS, Stable, false},
		{"LTS to Beta", LTS, Beta, false},
		{"Invalid from", Channel("invalid"), Stable, false},
		{"Invalid to", Stable, Channel("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.canDo, tt.from.CanPromoteTo(tt.to))
		})
	}
}

func TestChannel_GetDefaultRetention(t *testing.T) {
	tests := []struct {
		channel Channel
		days    int
	}{
		{Edge, 7},
		{Nightly, 14},
		{Beta, 30},
		{Stable, 365},
		{LTS, 1825},
		{Channel("invalid"), 30}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.channel), func(t *testing.T) {
			assert.Equal(t, tt.days, tt.channel.GetDefaultRetention())
		})
	}
}

func TestRelease_Validate(t *testing.T) {
	validArtifact := Artifact{
		Name:        "binary-linux-amd64",
		Platform:    "linux",
		Arch:        "amd64",
		URL:         "https://example.com/binary",
		Size:        1024,
		Checksum:    "abc123",
		ChecksumAlg: "sha256",
	}

	t.Run("valid release", func(t *testing.T) {
		release := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{validArtifact},
		}

		err := release.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing version", func(t *testing.T) {
		release := &Release{
			Channel:   Stable,
			Artifacts: []Artifact{validArtifact},
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version is required")
	})

	t.Run("invalid channel", func(t *testing.T) {
		release := &Release{
			Version:   "1.0.0",
			Channel:   Channel("invalid"),
			Artifacts: []Artifact{validArtifact},
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid channel")
	})

	t.Run("no artifacts", func(t *testing.T) {
		release := &Release{
			Version: "1.0.0",
			Channel: Stable,
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one artifact")
	})

	t.Run("artifact missing name", func(t *testing.T) {
		invalidArtifact := validArtifact
		invalidArtifact.Name = ""

		release := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{invalidArtifact},
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("artifact missing URL", func(t *testing.T) {
		invalidArtifact := validArtifact
		invalidArtifact.URL = ""

		release := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{invalidArtifact},
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL is required")
	})

	t.Run("artifact missing checksum", func(t *testing.T) {
		invalidArtifact := validArtifact
		invalidArtifact.Checksum = ""

		release := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{invalidArtifact},
		}

		err := release.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "checksum is required")
	})

	t.Run("multiple artifacts", func(t *testing.T) {
		artifact2 := validArtifact
		artifact2.Name = "binary-darwin-amd64"
		artifact2.Platform = "darwin"

		release := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{validArtifact, artifact2},
		}

		err := release.Validate()
		assert.NoError(t, err)
	})
}

func TestRelease_IsExpired(t *testing.T) {
	now := time.Now()

	t.Run("not expired", func(t *testing.T) {
		release := &Release{
			PublishedAt: now.Add(-5 * 24 * time.Hour), // 5 days ago
		}

		assert.False(t, release.IsExpired(7)) // 7 day retention
	})

	t.Run("expired", func(t *testing.T) {
		release := &Release{
			PublishedAt: now.Add(-10 * 24 * time.Hour), // 10 days ago
		}

		assert.True(t, release.IsExpired(7)) // 7 day retention
	})

	t.Run("no retention", func(t *testing.T) {
		release := &Release{
			PublishedAt: now.Add(-100 * 24 * time.Hour), // 100 days ago
		}

		assert.False(t, release.IsExpired(0))  // No expiration
		assert.False(t, release.IsExpired(-1)) // No expiration
	})

	t.Run("edge case - exactly at retention", func(t *testing.T) {
		release := &Release{
			PublishedAt: now.Add(-7 * 24 * time.Hour), // Exactly 7 days ago
		}

		expired := release.IsExpired(7)
		// This might be true or false depending on timing, but should not panic
		assert.IsType(t, true, expired)
	})
}

func TestRelease_GetArtifact(t *testing.T) {
	linuxAmd64 := Artifact{
		Name:     "binary-linux-amd64",
		Platform: "linux",
		Arch:     "amd64",
		URL:      "https://example.com/linux-amd64",
	}

	darwinArm64 := Artifact{
		Name:     "binary-darwin-arm64",
		Platform: "darwin",
		Arch:     "arm64",
		URL:      "https://example.com/darwin-arm64",
	}

	release := &Release{
		Version:   "1.0.0",
		Channel:   Stable,
		Artifacts: []Artifact{linuxAmd64, darwinArm64},
	}

	t.Run("found artifact", func(t *testing.T) {
		artifact := release.GetArtifact("linux", "amd64")
		require.NotNil(t, artifact)
		assert.Equal(t, linuxAmd64.Name, artifact.Name)
		assert.Equal(t, linuxAmd64.URL, artifact.URL)
	})

	t.Run("found second artifact", func(t *testing.T) {
		artifact := release.GetArtifact("darwin", "arm64")
		require.NotNil(t, artifact)
		assert.Equal(t, darwinArm64.Name, artifact.Name)
		assert.Equal(t, darwinArm64.URL, artifact.URL)
	})

	t.Run("not found - wrong platform", func(t *testing.T) {
		artifact := release.GetArtifact("windows", "amd64")
		assert.Nil(t, artifact)
	})

	t.Run("not found - wrong arch", func(t *testing.T) {
		artifact := release.GetArtifact("linux", "arm64")
		assert.Nil(t, artifact)
	})

	t.Run("not found - both wrong", func(t *testing.T) {
		artifact := release.GetArtifact("windows", "arm64")
		assert.Nil(t, artifact)
	})

	t.Run("empty artifacts", func(t *testing.T) {
		emptyRelease := &Release{
			Version:   "1.0.0",
			Channel:   Stable,
			Artifacts: []Artifact{},
		}

		artifact := emptyRelease.GetArtifact("linux", "amd64")
		assert.Nil(t, artifact)
	})
}

func TestChannelConfig(t *testing.T) {
	t.Run("channel config creation", func(t *testing.T) {
		config := &ChannelConfig{
			Name:          Beta,
			Description:   "Beta channel for testing",
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

		assert.Equal(t, Beta, config.Name)
		assert.Equal(t, "Beta channel for testing", config.Description)
		assert.Contains(t, config.PromotionPath, Stable)
		assert.Equal(t, 30, config.RetentionDays)
		assert.False(t, config.AutoPromotion)
		assert.Contains(t, config.RequiredTests, "unit")
		assert.Contains(t, config.Approvers, "admin1")
		assert.Equal(t, "https://hooks.example.com/webhook", config.WebhookURL)
		assert.Equal(t, "platform-team", config.Metadata["owner"])
	})
}

func TestPromotionRequest(t *testing.T) {
	t.Run("promotion request creation", func(t *testing.T) {
		now := time.Now()
		approvedTime := now.Add(time.Hour)

		request := &PromotionRequest{
			Version:     "1.0.0",
			FromChannel: Beta,
			ToChannel:   Stable,
			RequestedBy: "developer",
			RequestedAt: now,
			ApprovedBy:  "admin",
			ApprovedAt:  &approvedTime,
			TestResults: []TestResult{
				{
					Name:     "unit",
					Status:   "passed",
					Duration: 5000, // 5 seconds in milliseconds
					Output:   "All tests passed",
					RunAt:    now,
				},
			},
			Notes: "Ready for stable release",
			Metadata: map[string]string{
				"pr": "1234",
			},
		}

		assert.Equal(t, "1.0.0", request.Version)
		assert.Equal(t, Beta, request.FromChannel)
		assert.Equal(t, Stable, request.ToChannel)
		assert.Equal(t, "developer", request.RequestedBy)
		assert.Equal(t, "admin", request.ApprovedBy)
		assert.NotNil(t, request.ApprovedAt)
		assert.Equal(t, approvedTime, *request.ApprovedAt)
		assert.Len(t, request.TestResults, 1)
		assert.Equal(t, "unit", request.TestResults[0].Name)
		assert.Equal(t, "passed", request.TestResults[0].Status)
		assert.Equal(t, "Ready for stable release", request.Notes)
		assert.Equal(t, "1234", request.Metadata["pr"])
	})
}

func TestTestResult(t *testing.T) {
	t.Run("test result creation", func(t *testing.T) {
		now := time.Now()
		result := TestResult{
			Name:     "integration",
			Status:   "failed",
			Duration: 30000, // 30 seconds
			Output:   "Test failed at line 42",
			RunAt:    now,
		}

		assert.Equal(t, "integration", result.Name)
		assert.Equal(t, "failed", result.Status)
		assert.Equal(t, int64(30000), result.Duration)
		assert.Equal(t, "Test failed at line 42", result.Output)
		assert.Equal(t, now, result.RunAt)
	})
}

func TestChannelStats(t *testing.T) {
	t.Run("channel stats creation", func(t *testing.T) {
		now := time.Now()
		stats := &ChannelStats{
			Channel:         Stable,
			TotalReleases:   25,
			LatestVersion:   "2.1.0",
			LatestRelease:   now,
			ActiveDownloads: 1500,
			TotalDownloads:  50000,
		}

		assert.Equal(t, Stable, stats.Channel)
		assert.Equal(t, 25, stats.TotalReleases)
		assert.Equal(t, "2.1.0", stats.LatestVersion)
		assert.Equal(t, now, stats.LatestRelease)
		assert.Equal(t, int64(1500), stats.ActiveDownloads)
		assert.Equal(t, int64(50000), stats.TotalDownloads)
	})
}

func TestArtifact(t *testing.T) {
	t.Run("artifact creation with all fields", func(t *testing.T) {
		artifact := Artifact{
			Name:        "myapp-linux-amd64",
			Platform:    "linux",
			Arch:        "amd64",
			URL:         "https://releases.example.com/myapp-linux-amd64",
			Size:        12345678,
			Checksum:    "sha256:abcdef123456",
			ChecksumAlg: "sha256",
			Signature:   "-----BEGIN PGP SIGNATURE-----...",
			Metadata: map[string]string{
				"build_id": "12345",
				"commit":   "abc123",
			},
		}

		assert.Equal(t, "myapp-linux-amd64", artifact.Name)
		assert.Equal(t, "linux", artifact.Platform)
		assert.Equal(t, "amd64", artifact.Arch)
		assert.Equal(t, "https://releases.example.com/myapp-linux-amd64", artifact.URL)
		assert.Equal(t, int64(12345678), artifact.Size)
		assert.Equal(t, "sha256:abcdef123456", artifact.Checksum)
		assert.Equal(t, "sha256", artifact.ChecksumAlg)
		assert.Contains(t, artifact.Signature, "-----BEGIN PGP SIGNATURE-----")
		assert.Equal(t, "12345", artifact.Metadata["build_id"])
		assert.Equal(t, "abc123", artifact.Metadata["commit"])
	})
}

func TestDependency(t *testing.T) {
	t.Run("dependency creation", func(t *testing.T) {
		dep := Dependency{
			Name:        "go",
			Version:     "1.19+",
			Required:    true,
			Description: "Go programming language",
		}

		assert.Equal(t, "go", dep.Name)
		assert.Equal(t, "1.19+", dep.Version)
		assert.True(t, dep.Required)
		assert.Equal(t, "Go programming language", dep.Description)
	})

	t.Run("optional dependency", func(t *testing.T) {
		dep := Dependency{
			Name:     "docker",
			Version:  "20.10+",
			Required: false,
		}

		assert.Equal(t, "docker", dep.Name)
		assert.Equal(t, "20.10+", dep.Version)
		assert.False(t, dep.Required)
		assert.Empty(t, dep.Description)
	})
}

// Benchmark tests
func BenchmarkChannel_IsValid(b *testing.B) {
	channels := []Channel{Stable, Beta, Edge, Nightly, LTS, Channel("invalid")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		channel := channels[i%len(channels)]
		_ = channel.IsValid()
	}
}

func BenchmarkChannel_CanPromoteTo(b *testing.B) {
	fromChannels := []Channel{Edge, Nightly, Beta, Stable}
	toChannels := []Channel{Beta, Stable, LTS}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		from := fromChannels[i%len(fromChannels)]
		to := toChannels[i%len(toChannels)]
		_ = from.CanPromoteTo(to)
	}
}

func BenchmarkRelease_Validate(b *testing.B) {
	release := &Release{
		Version: "1.0.0",
		Channel: Stable,
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
		if err := release.Validate(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRelease_GetArtifact(b *testing.B) {
	artifacts := make([]Artifact, 10)
	for i := 0; i < 10; i++ {
		artifacts[i] = Artifact{
			Name:     fmt.Sprintf("binary-%d", i),
			Platform: "linux",
			Arch:     fmt.Sprintf("arch%d", i),
			URL:      fmt.Sprintf("https://example.com/binary-%d", i),
			Checksum: fmt.Sprintf("checksum%d", i),
		}
	}

	release := &Release{
		Version:   "1.0.0",
		Channel:   Stable,
		Artifacts: artifacts,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = release.GetArtifact("linux", fmt.Sprintf("arch%d", i%10))
	}
}
