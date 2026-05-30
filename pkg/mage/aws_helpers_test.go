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
	errAWSCommandFailed = errors.New("aws command failed")
	errAWSInvalidRunner = errors.New("invalid command runner")
	errAWSSTSTestFailed = errors.New("sts test failed")
)

func setTestCommandRunner(r any) error {
	runner, ok := r.(CommandRunner)
	if !ok {
		return errAWSInvalidRunner
	}
	return SetRunner(runner)
}

// AWSHelpersTestSuite defines the test suite for AWS helper functions
type AWSHelpersTestSuite struct {
	suite.Suite

	env    *testutil.TestEnvironment
	awsDir string
}

// SetupTest runs before each test
func (ts *AWSHelpersTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.awsDir = filepath.Join(ts.env.TempDir, ".aws")
	ts.Require().NoError(os.MkdirAll(ts.awsDir, 0o700))

	// Override HOME for AWS directory detection
	ts.T().Setenv("HOME", ts.env.TempDir)
}

// TearDownTest runs after each test
func (ts *AWSHelpersTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGetConfigSectionName tests INI section name generation
func (ts *AWSHelpersTestSuite) TestGetConfigSectionName() {
	tests := []struct {
		name    string
		profile string
		want    string
	}{
		{
			name:    "default profile",
			profile: "default",
			want:    "default",
		},
		{
			name:    "custom profile",
			profile: "production",
			want:    "profile production",
		},
		{
			name:    "another custom profile",
			profile: "staging",
			want:    "profile staging",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			result := getConfigSectionName(tt.profile)
			ts.Equal(tt.want, result)
		})
	}
}

// TestParseAWSINI tests INI file parsing
func (ts *AWSHelpersTestSuite) TestParseAWSINI() {
	tests := []struct {
		name         string
		iniContent   string
		wantSections int
		checkSection string
		checkKey     string
		checkValue   string
	}{
		{
			name: "single section",
			// #nosec G101 -- test fixture credentials
			iniContent: `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
`,
			wantSections: 1,
			checkSection: "default",
			checkKey:     "aws_access_key_id",
			checkValue:   "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name: "multiple sections",
			iniContent: `[default]
aws_access_key_id = KEY1

[production]
aws_access_key_id = KEY2
`,
			wantSections: 2,
			checkSection: "production",
			checkKey:     "aws_access_key_id",
			checkValue:   "KEY2",
		},
		{
			name: "with comments and empty lines",
			iniContent: `# This is a comment
[default]
aws_access_key_id = KEY1

# Another comment
[staging]
aws_access_key_id = KEY2
; Semicolon comment
region = us-west-2
`,
			wantSections: 2,
			checkSection: "staging",
			checkKey:     "region",
			checkValue:   "us-west-2",
		},
		{
			name: "with whitespace",
			iniContent: `[default]
  aws_access_key_id   =   SPACED_KEY
  region=us-east-1
`,
			wantSections: 1,
			checkSection: "default",
			checkKey:     "region",
			checkValue:   "us-east-1",
		},
		{
			name:         "empty content",
			iniContent:   "",
			wantSections: 0,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			ini := parseAWSINI([]byte(tt.iniContent))

			ts.Len(ini.Sections, tt.wantSections)

			if tt.checkSection != "" {
				found := false
				for _, section := range ini.Sections {
					if section.Name == tt.checkSection {
						found = true
						if tt.checkKey != "" {
							value, ok := section.Values[tt.checkKey]
							ts.True(ok, "key should exist")
							ts.Equal(tt.checkValue, value)
						}
						break
					}
				}
				ts.True(found, "section should exist")
			}
		})
	}
}

// TestWriteAWSINI tests INI file serialization
func (ts *AWSHelpersTestSuite) TestWriteAWSINI() {
	ini := &awsINIFile{
		Sections: []*awsINISection{
			{
				Name: "default",
				Values: map[string]string{
					"aws_access_key_id":     "KEY1",
					"aws_secret_access_key": "SECRET1",
				},
				KeyOrder: []string{"aws_access_key_id", "aws_secret_access_key"},
			},
			{
				Name: "production",
				Values: map[string]string{
					"region": "us-west-2",
				},
				KeyOrder: []string{"region"},
			},
		},
	}

	data := writeAWSINI(ini)
	content := string(data)

	// Verify sections are present
	ts.Contains(content, "[default]")
	ts.Contains(content, "[production]")

	// Verify key-value pairs
	ts.Contains(content, "aws_access_key_id = KEY1")
	ts.Contains(content, "region = us-west-2")

	// Parse it back and verify
	parsedINI := parseAWSINI(data)
	ts.Len(parsedINI.Sections, 2)
}

// TestGetOrCreateSection tests section creation and retrieval
func (ts *AWSHelpersTestSuite) TestGetOrCreateSection() {
	ini := &awsINIFile{
		Sections: []*awsINISection{
			{
				Name:     "existing",
				Values:   make(map[string]string),
				KeyOrder: []string{},
			},
		},
	}

	ts.Run("get existing section", func() {
		section := getOrCreateSection(ini, "existing")
		ts.NotNil(section)
		ts.Equal("existing", section.Name)
		ts.Len(ini.Sections, 1)
	})

	ts.Run("create new section", func() {
		section := getOrCreateSection(ini, "new")
		ts.NotNil(section)
		ts.Equal("new", section.Name)
		ts.Len(ini.Sections, 2)
	})
}

// TestSetINIValue tests setting values in INI sections
func (ts *AWSHelpersTestSuite) TestSetINIValue() {
	section := &awsINISection{
		Name:     "test",
		Values:   make(map[string]string),
		KeyOrder: []string{},
	}

	ts.Run("set new value", func() {
		setINIValue(section, "key1", "value1")
		ts.Equal("value1", section.Values["key1"])
		ts.Contains(section.KeyOrder, "key1")
		ts.Len(section.KeyOrder, 1)
	})

	ts.Run("update existing value", func() {
		setINIValue(section, "key1", "updated")
		ts.Equal("updated", section.Values["key1"])
		ts.Len(section.KeyOrder, 1, "should not duplicate in key order")
	})

	ts.Run("set multiple values", func() {
		setINIValue(section, "key2", "value2")
		setINIValue(section, "key3", "value3")
		ts.Len(section.Values, 3)
		ts.Len(section.KeyOrder, 3)
	})
}

// TestMaskCredential tests credential masking
func (ts *AWSHelpersTestSuite) TestMaskCredential() {
	tests := []struct {
		name       string
		credential string
		want       string
	}{
		{
			name: "standard access key",
			// #nosec G101 -- test fixture credential
			credential: "AKIAIOSFODNN7EXAMPLE",
			want:       "AKIA************MPLE",
		},
		{
			name:       "short string",
			credential: "SHORT",
			want:       "****",
		},
		{
			name:       "very short",
			credential: "AB",
			want:       "****",
		},
		{
			name:       "exactly 8 chars",
			credential: "12345678",
			want:       "****",
		},
		{
			name:       "9 chars",
			credential: "123456789",
			want:       "1234*6789",
		},
		{
			name: "long secret",
			// #nosec G101 -- test fixture credential
			credential: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			want:       "wJal********************************EKEY",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			result := maskCredential(tt.credential)
			ts.Equal(tt.want, result)

			// Verify first 4 and last 4 are preserved for long strings
			if len(tt.credential) > 8 {
				ts.Equal(tt.credential[:4], result[:4])
				ts.Equal(tt.credential[len(tt.credential)-4:], result[len(result)-4:])
			}
		})
	}
}

// TestBackupFile tests file backup creation
func (ts *AWSHelpersTestSuite) TestBackupFile() {
	ts.Run("backup existing file", func() {
		// Create original file
		filePath := filepath.Join(ts.awsDir, "credentials")
		originalContent := []byte("original content")
		err := os.WriteFile(filePath, originalContent, 0o600)
		ts.Require().NoError(err)

		// Create backup
		err = backupFile(filePath)
		ts.Require().NoError(err)

		// Verify backup exists
		backupPath := filePath + awsBackupSuffix
		// #nosec G304 -- test reads backup path derived from temp fixture
		backupContent, err := os.ReadFile(backupPath)
		ts.Require().NoError(err)
		ts.Equal(originalContent, backupContent)
	})

	ts.Run("backup non-existent file", func() {
		// Should not error
		err := backupFile(filepath.Join(ts.awsDir, "nonexistent"))
		ts.Require().NoError(err)
	})
}

// TestGetAWSSessionToken tests STS token retrieval
func (ts *AWSHelpersTestSuite) TestGetAWSSessionToken() {
	ts.Run("successful token retrieval", func() {
		// Create mock STS response
		mockResponse := map[string]any{
			"Credentials": map[string]string{
				"AccessKeyId":     "ASIATESTACCESSKEY",
				"SecretAccessKey": "TestSecretKey123",
				"SessionToken":    "TestSessionToken456",
				"Expiration":      "2024-12-31T23:59:59Z",
			},
		}
		mockJSON, err := json.Marshal(mockResponse)
		ts.Require().NoError(err)

		// Mock the runner
		ts.env.Runner.On("RunCmdOutput", "aws", []string{
			"sts", "get-session-token",
			"--serial-number", "arn:aws:iam::123456789012:mfa/test",
			"--token-code", "123456",
			"--duration-seconds", "43200",
			"--output", "json",
		}).Return(string(mockJSON), nil)

		err = ts.env.WithMockRunner(
			setTestCommandRunner,
			func() any { return GetRunner() },
			func() error {
				creds, tokenErr := getAWSSessionToken("default", "arn:aws:iam::123456789012:mfa/test", "123456", 43200)
				if tokenErr != nil {
					return tokenErr
				}

				// Verify credentials
				ts.Equal("ASIATESTACCESSKEY", creds.AccessKeyID)
				ts.Equal("TestSecretKey123", creds.SecretAccessKey)
				ts.Equal("TestSessionToken456", creds.SessionToken)
				ts.Equal("2024-12-31T23:59:59Z", creds.Expiration)

				return nil
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("invalid JSON response", func() {
		// Mock invalid JSON. Use a distinct token-code so this expectation has
		// unique args: SetupTest runs once per top-level test (not per ts.Run),
		// so all subtests here share ts.env.Runner. testify's findExpectedCall
		// returns the FIRST registered call matching the args, so reusing the
		// "successful" subtest's token-code ("123456") would match its valid-JSON
		// expectation instead of this one.
		ts.env.Runner.On("RunCmdOutput", "aws", []string{
			"sts", "get-session-token",
			"--serial-number", "arn:aws:iam::123456789012:mfa/test",
			"--token-code", "654321",
			"--duration-seconds", "43200",
			"--output", "json",
		}).Return("invalid json{{{", nil)

		err := ts.env.WithMockRunner(
			setTestCommandRunner,
			func() any { return GetRunner() },
			func() error {
				_, err := getAWSSessionToken("default", "arn:aws:iam::123456789012:mfa/test", "654321", 43200)
				return err
			},
		)

		ts.Require().Error(err)
		ts.Contains(err.Error(), "failed to parse STS response")
	})

	ts.Run("STS call fails", func() {
		// Mock AWS CLI error
		ts.env.Runner.On("RunCmdOutput", "aws", []string{
			"sts", "get-session-token",
			"--serial-number", "arn:aws:iam::123456789012:mfa/test",
			"--token-code", "invalid",
			"--duration-seconds", "43200",
			"--output", "json",
		}).Return("", errAWSSTSTestFailed)

		err := ts.env.WithMockRunner(
			setTestCommandRunner,
			func() any { return GetRunner() },
			func() error {
				_, err := getAWSSessionToken("default", "arn:aws:iam::123456789012:mfa/test", "invalid", 43200)
				return err
			},
		)

		ts.Require().Error(err)
		ts.ErrorIs(err, errSTSCallFailed)
	})
}

// TestCheckAWSSession tests session validation
func (ts *AWSHelpersTestSuite) TestCheckAWSSession() {
	ts.Run("valid session", func() {
		// Mock successful get-caller-identity response
		mockResponse := map[string]string{
			"Account": "123456789012",
			"Arn":     "arn:aws:iam::123456789012:user/testuser",
			"UserId":  "AIDAI1234567890EXAMPLE",
		}
		mockJSON, err := json.Marshal(mockResponse)
		ts.Require().NoError(err)

		ts.env.Runner.On("RunCmdOutput", "aws", []string{
			"sts", "get-caller-identity",
			"--output", "json",
		}).Return(string(mockJSON), nil)

		err = ts.env.WithMockRunner(
			setTestCommandRunner,
			func() any { return GetRunner() },
			func() error {
				accountID, userARN, isValid := checkAWSSession("default")

				ts.True(isValid)
				ts.Equal("123456789012", accountID)
				ts.Equal("arn:aws:iam::123456789012:user/testuser", userARN)

				return nil
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("invalid session", func() {
		ts.env.Runner.On("RunCmdOutput", "aws", []string{
			"sts", "get-caller-identity",
			"--output", "json",
			"--profile", "expired",
		}).Return("", errAWSCommandFailed)

		err := ts.env.WithMockRunner(
			setTestCommandRunner,
			func() any { return GetRunner() },
			func() error {
				_, _, isValid := checkAWSSession("expired")
				ts.False(isValid)
				return nil
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGetMFASerial tests MFA serial retrieval
func (ts *AWSHelpersTestSuite) TestGetMFASerial() {
	ts.Run("MFA serial found", func() {
		// Create config file with MFA serial
		configContent := `[default]
region = us-east-1
mfa_serial = arn:aws:iam::123456789012:mfa/testuser
`
		configPath := filepath.Join(ts.awsDir, awsConfigFile)
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		ts.Require().NoError(err)

		serial, err := getMFASerial("default")
		ts.Require().NoError(err)
		ts.Equal("arn:aws:iam::123456789012:mfa/testuser", serial)
	})

	ts.Run("MFA serial not found", func() {
		// Create config without MFA serial
		configContent := `[default]
region = us-east-1
`
		configPath := filepath.Join(ts.awsDir, awsConfigFile)
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		ts.Require().NoError(err)

		_, err = getMFASerial("default")
		ts.Require().Error(err)
		ts.ErrorIs(err, errMFASerialNotFound)
	})

	ts.Run("config file missing", func() {
		// Remove config file
		ts.Require().NoError(os.Remove(filepath.Join(ts.awsDir, awsConfigFile)))

		_, err := getMFASerial("default")
		ts.Require().Error(err)
		ts.ErrorIs(err, errMFASerialNotFound)
	})
}

// TestAWSHelpersTestSuite runs the test suite
func TestAWSHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(AWSHelpersTestSuite))
}
