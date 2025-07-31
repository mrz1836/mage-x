// Package channels provides release channel management for software distribution
package channels

import (
	"fmt"
	"time"
)

// Channel represents a release channel
type Channel string

const (
	// Stable channel for production-ready releases
	Stable Channel = "stable"

	// Beta channel for pre-release testing
	Beta Channel = "beta"

	// Edge channel for cutting-edge development builds
	Edge Channel = "edge"

	// Nightly channel for automated nightly builds
	Nightly Channel = "nightly"

	// LTS channel for long-term support releases
	LTS Channel = "lts"
)

// Release represents a software release in a channel
type Release struct {
	Version      string            `json:"version" yaml:"version"`
	Channel      Channel           `json:"channel" yaml:"channel"`
	PublishedAt  time.Time         `json:"published_at" yaml:"published_at"`
	ReleasedBy   string            `json:"released_by" yaml:"released_by"`
	Changelog    string            `json:"changelog" yaml:"changelog"`
	Artifacts    []Artifact        `json:"artifacts" yaml:"artifacts"`
	Dependencies []Dependency      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Promoted     bool              `json:"promoted" yaml:"promoted"`
	PromotedFrom Channel           `json:"promoted_from,omitempty" yaml:"promoted_from,omitempty"`
	PromotedAt   *time.Time        `json:"promoted_at,omitempty" yaml:"promoted_at,omitempty"`
	Deprecated   bool              `json:"deprecated" yaml:"deprecated"`
	DeprecatedAt *time.Time        `json:"deprecated_at,omitempty" yaml:"deprecated_at,omitempty"`
}

// Artifact represents a release artifact
type Artifact struct {
	Name        string            `json:"name" yaml:"name"`
	Platform    string            `json:"platform" yaml:"platform"`
	Arch        string            `json:"arch" yaml:"arch"`
	URL         string            `json:"url" yaml:"url"`
	Size        int64             `json:"size" yaml:"size"`
	Checksum    string            `json:"checksum" yaml:"checksum"`
	ChecksumAlg string            `json:"checksum_alg" yaml:"checksum_alg"`
	Signature   string            `json:"signature,omitempty" yaml:"signature,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Dependency represents a release dependency
type Dependency struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// ChannelConfig defines configuration for a release channel
type ChannelConfig struct {
	Name          Channel           `json:"name" yaml:"name"`
	Description   string            `json:"description" yaml:"description"`
	PromotionPath []Channel         `json:"promotion_path,omitempty" yaml:"promotion_path,omitempty"`
	RetentionDays int               `json:"retention_days" yaml:"retention_days"`
	AutoPromotion bool              `json:"auto_promotion" yaml:"auto_promotion"`
	RequiredTests []string          `json:"required_tests,omitempty" yaml:"required_tests,omitempty"`
	Approvers     []string          `json:"approvers,omitempty" yaml:"approvers,omitempty"`
	WebhookURL    string            `json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// PromotionRequest represents a request to promote a release between channels
type PromotionRequest struct {
	Version     string            `json:"version" yaml:"version"`
	FromChannel Channel           `json:"from_channel" yaml:"from_channel"`
	ToChannel   Channel           `json:"to_channel" yaml:"to_channel"`
	RequestedBy string            `json:"requested_by" yaml:"requested_by"`
	RequestedAt time.Time         `json:"requested_at" yaml:"requested_at"`
	ApprovedBy  string            `json:"approved_by,omitempty" yaml:"approved_by,omitempty"`
	ApprovedAt  *time.Time        `json:"approved_at,omitempty" yaml:"approved_at,omitempty"`
	TestResults []TestResult      `json:"test_results,omitempty" yaml:"test_results,omitempty"`
	Notes       string            `json:"notes,omitempty" yaml:"notes,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// TestResult represents the result of a test run for promotion
type TestResult struct {
	Name     string    `json:"name" yaml:"name"`
	Status   string    `json:"status" yaml:"status"`
	Duration int64     `json:"duration" yaml:"duration"` // milliseconds
	Output   string    `json:"output,omitempty" yaml:"output,omitempty"`
	RunAt    time.Time `json:"run_at" yaml:"run_at"`
}

// ChannelStats provides statistics for a channel
type ChannelStats struct {
	Channel         Channel   `json:"channel" yaml:"channel"`
	TotalReleases   int       `json:"total_releases" yaml:"total_releases"`
	LatestVersion   string    `json:"latest_version" yaml:"latest_version"`
	LatestRelease   time.Time `json:"latest_release" yaml:"latest_release"`
	ActiveDownloads int64     `json:"active_downloads" yaml:"active_downloads"`
	TotalDownloads  int64     `json:"total_downloads" yaml:"total_downloads"`
}

// IsValid checks if the channel is valid
func (c Channel) IsValid() bool {
	switch c {
	case Stable, Beta, Edge, Nightly, LTS:
		return true
	default:
		return false
	}
}

// String returns the string representation of the channel
func (c Channel) String() string {
	return string(c)
}

// CanPromoteTo checks if a release can be promoted from this channel to another
func (c Channel) CanPromoteTo(target Channel) bool {
	// Define promotion paths
	promotionPaths := map[Channel][]Channel{
		Edge:    {Beta, Nightly},
		Nightly: {Beta},
		Beta:    {Stable, LTS},
		Stable:  {LTS},
		LTS:     {}, // LTS cannot be promoted further
	}

	allowed, exists := promotionPaths[c]
	if !exists {
		return false
	}

	for _, ch := range allowed {
		if ch == target {
			return true
		}
	}

	return false
}

// GetDefaultRetention returns the default retention period for the channel
func (c Channel) GetDefaultRetention() int {
	retentionDays := map[Channel]int{
		Edge:    7,    // 1 week
		Nightly: 14,   // 2 weeks
		Beta:    30,   // 1 month
		Stable:  365,  // 1 year
		LTS:     1825, // 5 years
	}

	if days, exists := retentionDays[c]; exists {
		return days
	}

	return 30 // default to 30 days
}

// Validate checks if the release is valid
func (r *Release) Validate() error {
	if r.Version == "" {
		return fmt.Errorf("release version is required")
	}

	if !r.Channel.IsValid() {
		return fmt.Errorf("invalid channel: %s", r.Channel)
	}

	if len(r.Artifacts) == 0 {
		return fmt.Errorf("release must have at least one artifact")
	}

	for i := range r.Artifacts {
		artifact := &r.Artifacts[i]
		if artifact.Name == "" {
			return fmt.Errorf("artifact %d: name is required", i)
		}
		if artifact.URL == "" {
			return fmt.Errorf("artifact %d: URL is required", i)
		}
		if artifact.Checksum == "" {
			return fmt.Errorf("artifact %d: checksum is required", i)
		}
	}

	return nil
}

// IsExpired checks if the release has exceeded its retention period
func (r *Release) IsExpired(retentionDays int) bool {
	if retentionDays <= 0 {
		return false // No expiration
	}

	expirationTime := r.PublishedAt.Add(time.Duration(retentionDays) * 24 * time.Hour)
	return time.Now().After(expirationTime)
}

// GetArtifact returns the artifact for a specific platform and architecture
func (r *Release) GetArtifact(platform, arch string) *Artifact {
	for i := range r.Artifacts {
		artifact := &r.Artifacts[i]
		if artifact.Platform == platform && artifact.Arch == arch {
			return artifact
		}
	}
	return nil
}
