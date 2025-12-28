// Package env provides environment variable utilities
package env

// IsVerbose returns true if verbose mode is enabled via environment variables.
// It checks both VERBOSE and V environment variables.
// This is the canonical source for environment-based verbose detection.
// For config-file based verbose settings, see pkg/mage/config.go.
func IsVerbose() bool {
	return GetBool("VERBOSE", false) || GetBool("V", false)
}

// IsCI returns true if running in a CI/CD environment.
// It checks common CI environment variables from major providers:
// - CI (generic)
// - CONTINUOUS_INTEGRATION (generic)
// - GITHUB_ACTIONS (GitHub Actions)
// - GITLAB_CI (GitLab CI)
// - TRAVIS (Travis CI)
// - CIRCLECI (CircleCI)
// - JENKINS_URL (Jenkins)
// - CODEBUILD_BUILD_ID (AWS CodeBuild)
// - BUILDKITE (Buildkite)
// - AZURE_PIPELINES (Azure DevOps)
// - TEAMCITY_VERSION (TeamCity)
// - BITBUCKET_BUILD_NUMBER (Bitbucket Pipelines)
func IsCI() bool {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"CODEBUILD_BUILD_ID",
		"BUILDKITE",
		"AZURE_PIPELINES",
		"TEAMCITY_VERSION",
		"BITBUCKET_BUILD_NUMBER",
	}

	for _, v := range ciVars {
		if Exists(v) {
			return true
		}
	}
	return false
}

// IsDebug returns true if debug mode is enabled via environment variables.
// It checks the DEBUG environment variable.
func IsDebug() bool {
	return GetBool("DEBUG", false)
}

// IsQuiet returns true if quiet mode is enabled via environment variables.
// It checks the QUIET and Q environment variables.
func IsQuiet() bool {
	return GetBool("QUIET", false) || GetBool("Q", false)
}

// IsDryRun returns true if dry-run mode is enabled via environment variables.
// It checks the DRY_RUN and DRYRUN environment variables.
func IsDryRun() bool {
	return GetBool("DRY_RUN", false) || GetBool("DRYRUN", false)
}

// IsTest returns true if running in a test environment.
// It checks the GO_TEST environment variable (set by go test).
func IsTest() bool {
	return Exists("GO_TEST") || GetBool("TESTING", false)
}

// GetLogLevel returns the configured log level from environment.
// It checks LOG_LEVEL and defaults to "info".
// Valid values: debug, info, warn, error
func GetLogLevel() string {
	return GetString("LOG_LEVEL", "info")
}

// GetParallelism returns the number of parallel jobs to run.
// It checks GOMAXPROCS and PARALLEL environment variables.
// Defaults to 0 (use runtime default).
func GetParallelism() int {
	if p := GetInt("PARALLEL", 0); p > 0 {
		return p
	}
	return GetInt("GOMAXPROCS", 0)
}

// GetTimeout returns the default timeout from environment.
// It checks TIMEOUT environment variable.
// Defaults to empty string (no timeout).
func GetTimeout() string {
	return GetString("TIMEOUT", "")
}
