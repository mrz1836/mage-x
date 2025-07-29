package channels

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mrz1836/go-mage/pkg/common/errors"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// FileStore implements Store interface using filesystem
type FileStore struct {
	mu      sync.RWMutex
	baseDir string
	fileOps fileops.FileOperator
}

// NewFileStore creates a new file-based store
func NewFileStore(baseDir string, fileOps fileops.FileOperator) (*FileStore, error) {
	store := &FileStore{
		baseDir: baseDir,
		fileOps: fileOps,
	}

	// Initialize directory structure
	if err := store.initialize(); err != nil {
		return nil, err
	}

	return store, nil
}

// initialize creates the necessary directory structure
func (s *FileStore) initialize() error {
	dirs := []string{
		"channels",
		"releases",
		"releases/stable",
		"releases/beta",
		"releases/edge",
		"releases/nightly",
		"releases/lts",
		"promotions",
		"configs",
	}

	for _, dir := range dirs {
		path := filepath.Join(s.baseDir, dir)
		if err := s.fileOps.MkdirAll(path, 0o755); err != nil {
			return errors.Wrap(err, "failed to create directory")
		}
	}

	return nil
}

// GetRelease retrieves a release from storage
func (s *FileStore) GetRelease(channel Channel, version string) (*Release, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.getReleasePath(channel, version)

	if !s.fileOps.Exists(path) {
		return nil, nil
	}

	data, err := s.fileOps.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read release file")
	}

	var release Release
	if err := json.Unmarshal(data, &release); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal release")
	}

	return &release, nil
}

// ListReleases lists all releases in a channel
func (s *FileStore) ListReleases(channel Channel) ([]*Release, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listReleasesUnlocked(channel)
}

// listReleasesUnlocked lists all releases in a channel without acquiring locks
func (s *FileStore) listReleasesUnlocked(channel Channel) ([]*Release, error) {
	channelDir := filepath.Join(s.baseDir, "releases", channel.String())

	if !s.fileOps.Exists(channelDir) {
		return []*Release{}, nil
	}

	entries, err := s.fileOps.ReadDir(channelDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read channel directory")
	}

	var releases []*Release
	for _, entry := range entries {
		if entry.IsDir() || !hasJSONExtension(entry.Name()) {
			continue
		}

		path := filepath.Join(channelDir, entry.Name())
		data, err := s.fileOps.ReadFile(path)
		if err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to read release file %s: %v\n", path, err)
			continue
		}

		var release Release
		if err := json.Unmarshal(data, &release); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to unmarshal release %s: %v\n", path, err)
			continue
		}

		releases = append(releases, &release)
	}

	return releases, nil
}

// SaveRelease saves a release to storage
func (s *FileStore) SaveRelease(release *Release) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := release.Validate(); err != nil {
		return errors.Wrap(err, "invalid release")
	}

	path := s.getReleasePath(release.Channel, release.Version)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := s.fileOps.MkdirAll(dir, 0o755); err != nil {
		return errors.Wrap(err, "failed to create release directory")
	}

	data, err := json.MarshalIndent(release, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal release")
	}

	if err := s.fileOps.WriteFile(path, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write release file")
	}

	// Also save to index for faster lookups
	if err := s.updateChannelIndex(release.Channel); err != nil {
		// Log error but don't fail the save
		fmt.Printf("Warning: failed to update channel index: %v\n", err)
	}

	return nil
}

// DeleteRelease removes a release from storage
func (s *FileStore) DeleteRelease(channel Channel, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getReleasePath(channel, version)

	if !s.fileOps.Exists(path) {
		return errors.WithCode(errors.ErrNotFound, "release not found")
	}

	if err := s.fileOps.Remove(path); err != nil {
		return errors.Wrap(err, "failed to delete release file")
	}

	// Update index
	if err := s.updateChannelIndex(channel); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to update channel index: %v\n", err)
	}

	return nil
}

// GetChannelConfig retrieves channel configuration
func (s *FileStore) GetChannelConfig(channel Channel) (*ChannelConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, "configs", fmt.Sprintf("%s.json", channel))

	if !s.fileOps.Exists(path) {
		return nil, nil
	}

	data, err := s.fileOps.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	var config ChannelConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	return &config, nil
}

// SaveChannelConfig saves channel configuration
func (s *FileStore) SaveChannelConfig(config *ChannelConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.baseDir, "configs", fmt.Sprintf("%s.json", config.Name))

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	if err := s.fileOps.WriteFile(path, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}

	return nil
}

// GetPromotionHistory retrieves promotion history for a version
func (s *FileStore) GetPromotionHistory(version string) ([]*PromotionRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getPromotionHistoryUnlocked(version)
}

// getPromotionHistoryUnlocked retrieves promotion history for a version without acquiring locks
func (s *FileStore) getPromotionHistoryUnlocked(version string) ([]*PromotionRequest, error) {
	// Sanitize version for filename
	safeVersion := sanitizeVersion(version)
	path := filepath.Join(s.baseDir, "promotions", fmt.Sprintf("%s.json", safeVersion))

	if !s.fileOps.Exists(path) {
		return []*PromotionRequest{}, nil
	}

	data, err := s.fileOps.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read promotion history")
	}

	var history []*PromotionRequest
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal promotion history")
	}

	return history, nil
}

// SavePromotionRequest saves a promotion request
func (s *FileStore) SavePromotionRequest(request *PromotionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing history
	history, err := s.getPromotionHistoryUnlocked(request.Version)
	if err != nil {
		return err
	}

	// Append new request
	history = append(history, request)

	// Save updated history
	safeVersion := sanitizeVersion(request.Version)
	path := filepath.Join(s.baseDir, "promotions", fmt.Sprintf("%s.json", safeVersion))

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal promotion history")
	}

	if err := s.fileOps.WriteFile(path, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write promotion history")
	}

	return nil
}

// Helper methods

// getReleasePath returns the file path for a release
func (s *FileStore) getReleasePath(channel Channel, version string) string {
	// Sanitize version for filename (replace / with -)
	safeVersion := sanitizeVersion(version)
	return filepath.Join(s.baseDir, "releases", channel.String(), fmt.Sprintf("%s.json", safeVersion))
}

// updateChannelIndex updates the channel index file
func (s *FileStore) updateChannelIndex(channel Channel) error {
	indexPath := filepath.Join(s.baseDir, "channels", fmt.Sprintf("%s-index.json", channel))

	releases, err := s.listReleasesUnlocked(channel)
	if err != nil {
		return err
	}

	// Create index with basic metadata
	type indexEntry struct {
		Version     string `json:"version"`
		PublishedAt string `json:"published_at"`
		Deprecated  bool   `json:"deprecated"`
	}

	var index []indexEntry
	for _, release := range releases {
		index = append(index, indexEntry{
			Version:     release.Version,
			PublishedAt: release.PublishedAt.Format("2006-01-02T15:04:05Z"),
			Deprecated:  release.Deprecated,
		})
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return s.fileOps.WriteFile(indexPath, data, 0o644)
}

// sanitizeVersion makes version safe for use in filenames
func sanitizeVersion(version string) string {
	// Replace characters that might cause issues in filenames
	replacer := map[rune]rune{
		'/':  '-',
		'\\': '-',
		':':  '-',
		'*':  '-',
		'?':  '-',
		'"':  '-',
		'<':  '-',
		'>':  '-',
		'|':  '-',
	}

	result := make([]rune, 0, len(version))
	for _, char := range version {
		if replacement, ok := replacer[char]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, char)
		}
	}

	return string(result)
}

// hasJSONExtension checks if a filename has .json extension
func hasJSONExtension(name string) bool {
	return filepath.Ext(name) == ".json"
}
