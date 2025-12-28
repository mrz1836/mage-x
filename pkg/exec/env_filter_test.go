package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvFilteringExecutor_FilterEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		env         []string
		commandName string
		prefixes    []string
		whitelist   map[string][]string
		want        []string
	}{
		{
			name:        "filters AWS_SECRET_KEY",
			env:         []string{"PATH=/usr/bin", "AWS_SECRET_KEY=secret123", "HOME=/home/user"},
			commandName: "go",
			prefixes:    []string{"AWS_SECRET"},
			whitelist:   nil,
			want:        []string{"PATH=/usr/bin", "HOME=/home/user"},
		},
		{
			name:        "filters multiple sensitive vars",
			env:         []string{"PATH=/usr/bin", "AWS_SECRET_KEY=secret", "GITHUB_TOKEN=token123", "HOME=/home/user"},
			commandName: "go",
			prefixes:    []string{"AWS_SECRET", "GITHUB_TOKEN"},
			whitelist:   nil,
			want:        []string{"PATH=/usr/bin", "HOME=/home/user"},
		},
		{
			name:        "whitelisted variable for specific command",
			env:         []string{"PATH=/usr/bin", "GITHUB_TOKEN=token123"},
			commandName: "goreleaser",
			prefixes:    []string{"GITHUB_TOKEN"},
			whitelist:   map[string][]string{"goreleaser": {"GITHUB_TOKEN"}},
			want:        []string{"PATH=/usr/bin", "GITHUB_TOKEN=token123"},
		},
		{
			name:        "non-whitelisted command filters the variable",
			env:         []string{"PATH=/usr/bin", "GITHUB_TOKEN=token123"},
			commandName: "go",
			prefixes:    []string{"GITHUB_TOKEN"},
			whitelist:   map[string][]string{"goreleaser": {"GITHUB_TOKEN"}},
			want:        []string{"PATH=/usr/bin"},
		},
		{
			name:        "keeps variables that don't match prefix",
			env:         []string{"PATH=/usr/bin", "AWS_REGION=us-east-1", "AWS_SECRET_KEY=secret"},
			commandName: "go",
			prefixes:    []string{"AWS_SECRET"},
			whitelist:   nil,
			want:        []string{"PATH=/usr/bin", "AWS_REGION=us-east-1"},
		},
		{
			name:        "handles malformed env var without equals",
			env:         []string{"PATH=/usr/bin", "MALFORMED", "HOME=/home/user"},
			commandName: "go",
			prefixes:    []string{"SECRET"},
			whitelist:   nil,
			want:        []string{"PATH=/usr/bin", "MALFORMED", "HOME=/home/user"},
		},
		{
			name:        "prefix match requires underscore or end",
			env:         []string{"SECRETS_CONFIG=value", "SECRET_KEY=hidden", "SECRETIVE=ok"},
			commandName: "go",
			prefixes:    []string{"SECRET"},
			whitelist:   nil,
			want:        []string{"SECRETS_CONFIG=value", "SECRETIVE=ok"},
		},
		{
			name:        "empty env list",
			env:         []string{},
			commandName: "go",
			prefixes:    []string{"SECRET"},
			whitelist:   nil,
			want:        []string{},
		},
		{
			name:        "empty prefix list keeps all",
			env:         []string{"PATH=/usr/bin", "SECRET_KEY=hidden"},
			commandName: "go",
			prefixes:    []string{},
			whitelist:   nil,
			want:        []string{"PATH=/usr/bin", "SECRET_KEY=hidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &EnvFilteringExecutor{
				SensitivePrefixes: tt.prefixes,
				Whitelist:         tt.whitelist,
			}

			got := executor.FilterEnvironment(tt.env, tt.commandName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewEnvFilteringExecutor(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		base := NewBase()
		executor := NewEnvFilteringExecutor(base)

		assert.Equal(t, DefaultSensitivePrefixes, executor.SensitivePrefixes)
		assert.Equal(t, DefaultEnvWhitelist, executor.Whitelist)
		assert.Equal(t, base, executor.wrapped)
	})

	t.Run("custom sensitive prefixes", func(t *testing.T) {
		base := NewBase()
		customPrefixes := []string{"CUSTOM_SECRET", "MY_TOKEN"}
		executor := NewEnvFilteringExecutor(base, WithSensitivePrefixes(customPrefixes))

		assert.Equal(t, customPrefixes, executor.SensitivePrefixes)
	})

	t.Run("custom whitelist", func(t *testing.T) {
		base := NewBase()
		customWhitelist := map[string][]string{
			"myapp": {"MY_TOKEN", "API_KEY"},
		}
		executor := NewEnvFilteringExecutor(base, WithEnvWhitelist(customWhitelist))

		assert.Equal(t, customWhitelist, executor.Whitelist)
	})
}

func TestDefaultSensitivePrefixes(t *testing.T) {
	// Verify we have sensible defaults
	expectedPrefixes := []string{
		"AWS_SECRET",
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"NPM_TOKEN",
		"DOCKER_PASSWORD",
		"DATABASE_PASSWORD",
		"API_KEY",
		"SECRET",
		"PRIVATE_KEY",
	}

	require.Equal(t, expectedPrefixes, DefaultSensitivePrefixes)
}

func TestDefaultEnvWhitelist(t *testing.T) {
	// Verify goreleaser has access to common tokens
	goreleaserTokens, ok := DefaultEnvWhitelist["goreleaser"]
	require.True(t, ok, "goreleaser should be in default whitelist")
	assert.Contains(t, goreleaserTokens, "GITHUB_TOKEN")
	assert.Contains(t, goreleaserTokens, "GITLAB_TOKEN")
	assert.Contains(t, goreleaserTokens, "GITEA_TOKEN")
}

func TestEnvFilteringExecutor_ImplementsInterfaces(t *testing.T) {
	base := NewBase()
	executor := NewEnvFilteringExecutor(base)

	// Verify interface implementations
	var _ Executor = executor
	var _ ExecutorWithEnv = executor
	var _ StreamingExecutor = executor
}

func BenchmarkFilterEnvironment(b *testing.B) {
	executor := &EnvFilteringExecutor{
		SensitivePrefixes: DefaultSensitivePrefixes,
		Whitelist:         DefaultEnvWhitelist,
	}

	// Simulate realistic environment
	env := []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=/home/user",
		"GOPATH=/home/user/go",
		"GOROOT=/usr/local/go",
		"AWS_SECRET_KEY=somesecret",
		"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
		"GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"NPM_TOKEN=npm_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"EDITOR=vim",
		"SHELL=/bin/bash",
		"USER=testuser",
		"LANG=en_US.UTF-8",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.FilterEnvironment(env, "go")
	}
}
