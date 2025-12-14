package mage

import (
	"testing"
)

func TestCIDetector_IsCI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected bool
	}{
		{
			name:     "no CI environment",
			env:      map[string]string{},
			expected: false,
		},
		{
			name:     "generic CI=true",
			env:      map[string]string{"CI": "true"},
			expected: true,
		},
		{
			name:     "GitHub Actions",
			env:      map[string]string{"GITHUB_ACTIONS": "true"},
			expected: true,
		},
		{
			name:     "GitLab CI",
			env:      map[string]string{"GITLAB_CI": "true"},
			expected: true,
		},
		{
			name:     "CircleCI",
			env:      map[string]string{"CIRCLECI": "true"},
			expected: true,
		},
		{
			name:     "Travis CI",
			env:      map[string]string{"TRAVIS": "true"},
			expected: true,
		},
		{
			name:     "Jenkins",
			env:      map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			expected: true,
		},
		{
			name:     "Azure Pipelines",
			env:      map[string]string{"TF_BUILD": "True"},
			expected: true,
		},
		{
			name:     "Buildkite",
			env:      map[string]string{"BUILDKITE": "true"},
			expected: true,
		},
		{
			name:     "Drone",
			env:      map[string]string{"DRONE": "true"},
			expected: true,
		},
		{
			name:     "AWS CodeBuild",
			env:      map[string]string{"CODEBUILD_CI": "true"},
			expected: true,
		},
		{
			name:     "TeamCity",
			env:      map[string]string{"TEAMCITY_VERSION": "2023.1"},
			expected: true,
		},
		{
			name:     "Bitbucket Pipelines",
			env:      map[string]string{"BITBUCKET_BUILD_NUMBER": "123"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector := newCIDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})
			if got := detector.IsCI(); got != tt.expected {
				t.Errorf("IsCI() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCIDetector_Platform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "local - no CI",
			env:      map[string]string{},
			expected: "local",
		},
		{
			name:     "generic CI",
			env:      map[string]string{"CI": "true"},
			expected: "generic",
		},
		{
			name:     "GitHub Actions",
			env:      map[string]string{"GITHUB_ACTIONS": "true"},
			expected: "github",
		},
		{
			name:     "GitLab CI",
			env:      map[string]string{"GITLAB_CI": "true"},
			expected: "gitlab",
		},
		{
			name:     "CircleCI",
			env:      map[string]string{"CIRCLECI": "true"},
			expected: "circleci",
		},
		{
			name:     "Travis CI",
			env:      map[string]string{"TRAVIS": "true"},
			expected: "travis",
		},
		{
			name:     "Jenkins",
			env:      map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			expected: "jenkins",
		},
		{
			name:     "Azure Pipelines",
			env:      map[string]string{"TF_BUILD": "True"},
			expected: "azure",
		},
		{
			name:     "Buildkite",
			env:      map[string]string{"BUILDKITE": "true"},
			expected: "buildkite",
		},
		{
			name:     "Drone",
			env:      map[string]string{"DRONE": "true"},
			expected: "drone",
		},
		{
			name:     "AWS CodeBuild",
			env:      map[string]string{"CODEBUILD_CI": "true"},
			expected: "codebuild",
		},
		{
			name:     "TeamCity",
			env:      map[string]string{"TEAMCITY_VERSION": "2023.1"},
			expected: "teamcity",
		},
		{
			name:     "Bitbucket Pipelines",
			env:      map[string]string{"BITBUCKET_BUILD_NUMBER": "123"},
			expected: "bitbucket",
		},
		{
			name: "GitHub takes priority over generic CI",
			env: map[string]string{
				"CI":             "true",
				"GITHUB_ACTIONS": "true",
			},
			expected: "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector := newCIDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})
			if got := detector.Platform(); got != tt.expected {
				t.Errorf("Platform() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCIDetector_GetConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		env        map[string]string
		params     map[string]string
		cfg        *Config
		wantEnable bool
		wantFormat CIFormat
	}{
		{
			name:       "default - no CI",
			env:        map[string]string{},
			params:     nil,
			cfg:        nil,
			wantEnable: false,
			wantFormat: CIFormatJSON,
		},
		{
			name:       "auto-enable in GitHub Actions",
			env:        map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"},
			params:     nil,
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatGitHub,
		},
		{
			name:       "auto-enable in generic CI",
			env:        map[string]string{"CI": "true"},
			params:     nil,
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatJSON,
		},
		{
			name:       "explicit enable via param",
			env:        map[string]string{},
			params:     map[string]string{"ci": "true"},
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatJSON,
		},
		{
			name:       "explicit enable via param - empty value",
			env:        map[string]string{},
			params:     map[string]string{"ci": ""},
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatJSON,
		},
		{
			name:       "explicit disable via param in CI",
			env:        map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"},
			params:     map[string]string{"ci": "false"},
			cfg:        nil,
			wantEnable: false,
			wantFormat: CIFormatGitHub,
		},
		{
			name:       "env override - enable",
			env:        map[string]string{"MAGE_X_CI_MODE": "true"},
			params:     nil,
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatJSON,
		},
		{
			name:       "env override - format",
			env:        map[string]string{"CI": "true", "MAGE_X_CI_FORMAT": "github"},
			params:     nil,
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatGitHub,
		},
		{
			name:   "config file settings",
			env:    map[string]string{},
			params: nil,
			cfg: &Config{
				Test: TestConfig{
					CIMode: CIMode{
						Enabled:      true,
						Format:       CIFormatGitHub,
						ContextLines: 30,
						MaxMemoryMB:  200,
					},
				},
			},
			wantEnable: true,
			wantFormat: CIFormatGitHub,
		},
		{
			name:   "param overrides config",
			env:    map[string]string{},
			params: map[string]string{"ci": "false"},
			cfg: &Config{
				Test: TestConfig{
					CIMode: CIMode{
						Enabled: true,
						Format:  CIFormatGitHub,
					},
				},
			},
			wantEnable: false,
			wantFormat: CIFormatGitHub,
		},
		{
			name:       "param format override",
			env:        map[string]string{},
			params:     map[string]string{"ci": "true", "ci_format": "json"},
			cfg:        nil,
			wantEnable: true,
			wantFormat: CIFormatJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector := newCIDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			got := detector.GetConfig(tt.params, tt.cfg)

			if got.Enabled != tt.wantEnable {
				t.Errorf("GetConfig().Enabled = %v, want %v", got.Enabled, tt.wantEnable)
			}
			if got.Format != tt.wantFormat {
				t.Errorf("GetConfig().Format = %v, want %v", got.Format, tt.wantFormat)
			}
		})
	}
}

func TestCIDetector_GetConfig_ContextLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected int
	}{
		{
			name:     "default",
			env:      map[string]string{},
			expected: 20,
		},
		{
			name:     "env override",
			env:      map[string]string{"MAGE_X_CI_CONTEXT": "50"},
			expected: 50,
		},
		{
			name:     "env override - invalid (too high)",
			env:      map[string]string{"MAGE_X_CI_CONTEXT": "200"},
			expected: 20, // Should use default
		},
		{
			name:     "env override - invalid (negative)",
			env:      map[string]string{"MAGE_X_CI_CONTEXT": "-5"},
			expected: 20, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector := newCIDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			got := detector.GetConfig(nil, nil)

			if got.ContextLines != tt.expected {
				t.Errorf("GetConfig().ContextLines = %d, want %d", got.ContextLines, tt.expected)
			}
		})
	}
}

func TestCIDetector_GetConfig_MaxMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      map[string]string
		expected int
	}{
		{
			name:     "default",
			env:      map[string]string{},
			expected: 100,
		},
		{
			name:     "env override - plain number",
			env:      map[string]string{"MAGE_X_CI_MAX_MEMORY": "200"},
			expected: 200,
		},
		{
			name:     "env override - with MB suffix",
			env:      map[string]string{"MAGE_X_CI_MAX_MEMORY": "300MB"},
			expected: 300,
		},
		{
			name:     "env override - with mb suffix lowercase",
			env:      map[string]string{"MAGE_X_CI_MAX_MEMORY": "400mb"},
			expected: 400,
		},
		{
			name:     "env override - invalid (too low)",
			env:      map[string]string{"MAGE_X_CI_MAX_MEMORY": "5"},
			expected: 100, // Should use default
		},
		{
			name:     "env override - invalid (too high)",
			env:      map[string]string{"MAGE_X_CI_MAX_MEMORY": "2000"},
			expected: 100, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			detector := newCIDetectorWithEnv(func(key string) string {
				return tt.env[key]
			})

			got := detector.GetConfig(nil, nil)

			if got.MaxMemoryMB != tt.expected {
				t.Errorf("GetConfig().MaxMemoryMB = %d, want %d", got.MaxMemoryMB, tt.expected)
			}
		})
	}
}

func TestCIDetector_GetMetadata(t *testing.T) {
	t.Parallel()

	t.Run("GitHub Actions metadata", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"GITHUB_ACTIONS":  "true",
			"GITHUB_REF_NAME": "main",
			"GITHUB_SHA":      "abc123",
			"GITHUB_RUN_ID":   "12345",
			"GITHUB_WORKFLOW": "CI",
		}
		d := newCIDetectorWithEnv(func(key string) string {
			return env[key]
		})
		detector, ok := d.(*ciDetector)
		if !ok {
			t.Fatal("failed to cast to *ciDetector")
		}

		metadata := detector.GetMetadata()

		if metadata.Platform != string(CIFormatGitHub) {
			t.Errorf("GetMetadata().Platform = %q, want %q", metadata.Platform, string(CIFormatGitHub))
		}
		if metadata.Branch != "main" {
			t.Errorf("GetMetadata().Branch = %q, want %q", metadata.Branch, "main")
		}
		if metadata.Commit != "abc123" {
			t.Errorf("GetMetadata().Commit = %q, want %q", metadata.Commit, "abc123")
		}
		if metadata.RunID != "12345" {
			t.Errorf("GetMetadata().RunID = %q, want %q", metadata.RunID, "12345")
		}
		if metadata.Workflow != "CI" {
			t.Errorf("GetMetadata().Workflow = %q, want %q", metadata.Workflow, "CI")
		}
	})

	t.Run("GitLab CI metadata", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"GITLAB_CI":          "true",
			"CI_COMMIT_REF_NAME": "feature-branch",
			"CI_COMMIT_SHA":      "def456",
			"CI_PIPELINE_ID":     "67890",
			"CI_JOB_NAME":        "test",
		}
		d := newCIDetectorWithEnv(func(key string) string {
			return env[key]
		})
		detector, ok := d.(*ciDetector)
		if !ok {
			t.Fatal("failed to cast to *ciDetector")
		}

		metadata := detector.GetMetadata()

		if metadata.Platform != "gitlab" {
			t.Errorf("GetMetadata().Platform = %q, want %q", metadata.Platform, "gitlab")
		}
		if metadata.Branch != "feature-branch" {
			t.Errorf("GetMetadata().Branch = %q, want %q", metadata.Branch, "feature-branch")
		}
		if metadata.Commit != "def456" {
			t.Errorf("GetMetadata().Commit = %q, want %q", metadata.Commit, "def456")
		}
		if metadata.RunID != "67890" {
			t.Errorf("GetMetadata().RunID = %q, want %q", metadata.RunID, "67890")
		}
		if metadata.Workflow != "test" {
			t.Errorf("GetMetadata().Workflow = %q, want %q", metadata.Workflow, "test")
		}
	})

	t.Run("Azure Pipelines metadata", func(t *testing.T) {
		t.Parallel()
		env := map[string]string{
			"TF_BUILD":             "True",
			"BUILD_SOURCEBRANCH":   "refs/heads/develop",
			"BUILD_SOURCEVERSION":  "789abc",
			"BUILD_BUILDNUMBER":    "42",
			"BUILD_DEFINITIONNAME": "Build and Test",
		}
		d := newCIDetectorWithEnv(func(key string) string {
			return env[key]
		})
		detector, ok := d.(*ciDetector)
		if !ok {
			t.Fatal("failed to cast to *ciDetector")
		}

		metadata := detector.GetMetadata()

		if metadata.Platform != "azure" {
			t.Errorf("GetMetadata().Platform = %q, want %q", metadata.Platform, "azure")
		}
		if metadata.Branch != "develop" {
			t.Errorf("GetMetadata().Branch = %q, want %q", metadata.Branch, "develop")
		}
		if metadata.Commit != "789abc" {
			t.Errorf("GetMetadata().Commit = %q, want %q", metadata.Commit, "789abc")
		}
		if metadata.RunID != "42" {
			t.Errorf("GetMetadata().RunID = %q, want %q", metadata.RunID, "42")
		}
		if metadata.Workflow != "Build and Test" {
			t.Errorf("GetMetadata().Workflow = %q, want %q", metadata.Workflow, "Build and Test")
		}
	})
}

func TestNewCIDetector(t *testing.T) {
	t.Parallel()

	detector := NewCIDetector()
	if detector == nil {
		t.Error("NewCIDetector() returned nil")
	}
}
