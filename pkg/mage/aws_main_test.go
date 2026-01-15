//go:build integration
// +build integration

package mage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors
var (
	errAWSLoginTestFailed   = errors.New("login test failed")
	errAWSSetupTestFailed   = errors.New("setup test failed")
	errAWSRefreshTestFailed = errors.New("refresh test failed")
	errAWSStatusTestFailed  = errors.New("status test failed")
)

// AWSMainTestSuite defines the test suite for AWS main methods
type AWSMainTestSuite struct {
	suite.Suite
	env    *testutil.TestEnvironment
	aws    AWS
	awsDir string
}

// SetupTest runs before each test
func (ts *AWSMainTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.awsDir = filepath.Join(ts.env.TempDir, ".aws")
	os.MkdirAll(ts.awsDir, 0o700)

	// Override HOME
	os.Setenv("HOME", ts.env.TempDir)

	ts.aws = AWS{}
}

// TearDownTest runs after each test
func (ts *AWSMainTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestStatus_NoCredentialsFile tests Status with missing credentials
func (ts *AWSMainTestSuite) TestStatus_NoCredentialsFile() {
	// No credentials file exists
	err := ts.aws.Status()

	// Should not error, just warn
	ts.Assert().NoError(err)
}

// TestStatus_SingleProfile tests Status with one profile
func (ts *AWSMainTestSuite) TestStatus_SingleProfile() {
	// Create credentials file
	credContent := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
`
	credPath := filepath.Join(ts.awsDir, awsCredentialsFile)
	err := os.WriteFile(credPath, []byte(credContent), 0o600)
	ts.Require().NoError(err)

	// Mock checkAWSSession to return invalid (no actual AWS CLI call)
	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
	}).Return("", errAWSCommandFailed)

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.aws.Status()
		},
	)

	ts.Assert().NoError(err)
}

// TestStatus_MultipleProfiles tests Status with multiple profiles
func (ts *AWSMainTestSuite) TestStatus_MultipleProfiles() {
	// Create credentials file with multiple profiles
	credContent := `[default]
aws_access_key_id = KEY1

[production]
aws_access_key_id = KEY2

[staging]
aws_access_key_id = KEY3
`
	credPath := filepath.Join(ts.awsDir, awsCredentialsFile)
	err := os.WriteFile(credPath, []byte(credContent), 0o600)
	ts.Require().NoError(err)

	// Mock session checks
	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
	}).Return("", errAWSCommandFailed).Maybe()

	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
		"--profile", "production",
	}).Return("", errAWSCommandFailed).Maybe()

	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
		"--profile", "staging",
	}).Return("", errAWSCommandFailed).Maybe()

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.aws.Status()
		},
	)

	ts.Assert().NoError(err)
}

// TestStatus_FilterByProfile tests Status with profile filter
func (ts *AWSMainTestSuite) TestStatus_FilterByProfile() {
	// Create credentials file
	credContent := `[default]
aws_access_key_id = KEY1

[production]
aws_access_key_id = KEY2
`
	credPath := filepath.Join(ts.awsDir, awsCredentialsFile)
	err := os.WriteFile(credPath, []byte(credContent), 0o600)
	ts.Require().NoError(err)

	// Mock session check for production only
	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
		"--profile", "production",
	}).Return("", errAWSCommandFailed)

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.aws.Status("profile=production")
		},
	)

	ts.Assert().NoError(err)
}

// TestWriteAWSCredentials tests credential file writing
func (ts *AWSMainTestSuite) TestWriteAWSCredentials() {
	ts.Run("write new credentials", func() {
		credPath := filepath.Join(ts.awsDir, awsCredentialsFile)

		err := writeAWSCredentials(credPath, "default", "AKIATEST", "SECRET", "")
		ts.Assert().NoError(err)

		// Verify file exists and has correct content
		content, err := os.ReadFile(credPath)
		ts.Require().NoError(err)

		ts.Assert().Contains(string(content), "[default]")
		ts.Assert().Contains(string(content), "aws_access_key_id = AKIATEST")
		ts.Assert().Contains(string(content), "aws_secret_access_key = SECRET")
	})

	ts.Run("update existing credentials", func() {
		credPath := filepath.Join(ts.awsDir, awsCredentialsFile)

		// Write initial credentials
		err := writeAWSCredentials(credPath, "default", "KEY1", "SECRET1", "")
		ts.Require().NoError(err)

		// Update credentials
		err = writeAWSCredentials(credPath, "default", "KEY2", "SECRET2", "")
		ts.Assert().NoError(err)

		// Verify updated
		content, err := os.ReadFile(credPath)
		ts.Require().NoError(err)

		ts.Assert().Contains(string(content), "KEY2")
		ts.Assert().NotContains(string(content), "KEY1")
	})

	ts.Run("write with session token", func() {
		credPath := filepath.Join(ts.awsDir, awsCredentialsFile)

		err := writeAWSCredentials(credPath, "default", "KEY", "SECRET", "TOKEN123")
		ts.Assert().NoError(err)

		content, err := os.ReadFile(credPath)
		ts.Require().NoError(err)

		ts.Assert().Contains(string(content), "aws_session_token = TOKEN123")
	})

	ts.Run("create backup", func() {
		credPath := filepath.Join(ts.awsDir, awsCredentialsFile)

		// Write initial
		initialContent := `[default]
aws_access_key_id = ORIGINAL
`
		err := os.WriteFile(credPath, []byte(initialContent), 0o600)
		ts.Require().NoError(err)

		// Update (should create backup)
		err = writeAWSCredentials(credPath, "default", "UPDATED", "SECRET", "")
		ts.Assert().NoError(err)

		// Verify backup exists
		backupPath := credPath + awsBackupSuffix
		backupContent, err := os.ReadFile(backupPath)
		ts.Assert().NoError(err)
		ts.Assert().Contains(string(backupContent), "ORIGINAL")
	})
}

// TestWriteAWSConfig tests config file writing
func (ts *AWSMainTestSuite) TestWriteAWSConfig() {
	ts.Run("write new config", func() {
		configPath := filepath.Join(ts.awsDir, awsConfigFile)

		err := writeAWSConfig(configPath, "default", "arn:aws:iam::123456789012:mfa/test")
		ts.Assert().NoError(err)

		content, err := os.ReadFile(configPath)
		ts.Require().NoError(err)

		ts.Assert().Contains(string(content), "[default]")
		ts.Assert().Contains(string(content), "mfa_serial = arn:aws:iam::123456789012:mfa/test")
	})

	ts.Run("write with profile prefix", func() {
		configPath := filepath.Join(ts.awsDir, awsConfigFile)

		err := writeAWSConfig(configPath, "production", "arn:aws:iam::123456789012:mfa/prod")
		ts.Assert().NoError(err)

		content, err := os.ReadFile(configPath)
		ts.Require().NoError(err)

		// Non-default profiles get "profile " prefix
		ts.Assert().Contains(string(content), "[profile production]")
		ts.Assert().Contains(string(content), "mfa_serial = arn:aws:iam::123456789012:mfa/prod")
	})
}

// TestGetSourceProfile tests source profile retrieval
func (ts *AWSMainTestSuite) TestGetSourceProfile() {
	ts.Run("source profile found", func() {
		configContent := `[profile mrz]
source_profile = mrz-base
region = us-east-1
`
		configPath := filepath.Join(ts.awsDir, awsConfigFile)
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		ts.Require().NoError(err)

		sourceProfile := getSourceProfile("mrz")
		ts.Assert().Equal("mrz-base", sourceProfile)
	})

	ts.Run("source profile not found", func() {
		configContent := `[profile test]
region = us-east-1
`
		configPath := filepath.Join(ts.awsDir, awsConfigFile)
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		ts.Require().NoError(err)

		sourceProfile := getSourceProfile("test")
		ts.Assert().Empty(sourceProfile)
	})

	ts.Run("config file missing", func() {
		sourceProfile := getSourceProfile("any")
		ts.Assert().Empty(sourceProfile)
	})
}

// TestHasValidAWSSetup tests setup detection
func (ts *AWSMainTestSuite) TestHasValidAWSSetup() {
	ts.Run("valid setup exists", func() {
		credContent := `[default]
aws_access_key_id = KEY1
aws_secret_access_key = SECRET1
`
		credPath := filepath.Join(ts.awsDir, awsCredentialsFile)
		err := os.WriteFile(credPath, []byte(credContent), 0o600)
		ts.Require().NoError(err)

		hasSetup := hasValidAWSSetup("default")
		ts.Assert().True(hasSetup)
	})

	ts.Run("no setup exists", func() {
		hasSetup := hasValidAWSSetup("nonexistent")
		ts.Assert().False(hasSetup)
	})

	ts.Run("credentials file missing", func() {
		// Remove credentials file
		os.Remove(filepath.Join(ts.awsDir, awsCredentialsFile))

		hasSetup := hasValidAWSSetup("default")
		ts.Assert().False(hasSetup)
	})
}

// TestDisplayAWSProfileStatus tests profile status display
func (ts *AWSMainTestSuite) TestDisplayAWSProfileStatus() {
	section := &awsINISection{
		Name: "testprofile",
		Values: map[string]string{
			"aws_access_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
		KeyOrder: []string{"aws_access_key_id", "aws_secret_access_key"},
	}

	// Mock session check
	mockResponse := map[string]string{
		"Account": "123456789012",
		"Arn":     "arn:aws:iam::123456789012:user/testuser",
		"UserId":  "AIDAI1234567890EXAMPLE",
	}
	mockJSON, _ := json.Marshal(mockResponse)

	ts.env.Runner.On("RunCmdOutput", "aws", []string{
		"sts", "get-caller-identity",
		"--output", "json",
		"--profile", "testprofile",
	}).Return(string(mockJSON), nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			// This function prints output, just verify it doesn't panic
			ts.Assert().NotPanics(func() {
				displayAWSProfileStatus(section, nil)
			})
			return nil
		},
	)

	ts.Assert().NoError(err)
}

// TestAWSConfigSectionName tests section name formatting
func (ts *AWSMainTestSuite) TestAWSConfigSectionName() {
	tests := []struct {
		profile string
		want    string
	}{
		{"default", "default"},
		{"production", "profile production"},
		{"staging", "profile staging"},
		{"dev", "profile dev"},
	}

	for _, tt := range tests {
		ts.Run(tt.profile, func() {
			result := getConfigSectionName(tt.profile)
			ts.Assert().Equal(tt.want, result)
		})
	}
}

// TestAWSMainTestSuite runs the test suite
func TestAWSMainTestSuite(t *testing.T) {
	suite.Run(t, new(AWSMainTestSuite))
}
