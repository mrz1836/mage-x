//go:build unit
// +build unit

package mage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// TestParseAWSINI tests the INI file parser
func TestParseAWSINI(t *testing.T) {
	t.Run("parse valid credentials file", func(t *testing.T) {
		data := []byte(`[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[production]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
aws_session_token = AQoDYXdzEJr...
`)
		ini := parseAWSINI(data)
		require.Len(t, ini.Sections, 2)

		// Check default section
		assert.Equal(t, "default", ini.Sections[0].Name)
		assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", ini.Sections[0].Values["aws_access_key_id"])
		assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", ini.Sections[0].Values["aws_secret_access_key"])

		// Check production section
		assert.Equal(t, "production", ini.Sections[1].Name)
		assert.Equal(t, "AKIAI44QH8DHBEXAMPLE", ini.Sections[1].Values["aws_access_key_id"])
		assert.Contains(t, ini.Sections[1].Values, "aws_session_token")
	})

	t.Run("parse config file with profile prefix", func(t *testing.T) {
		data := []byte(`[default]
region = us-east-1
mfa_serial = arn:aws:iam::123456789012:mfa/user

[profile production]
region = us-west-2
mfa_serial = arn:aws:iam::123456789012:mfa/admin
`)
		ini := parseAWSINI(data)
		require.Len(t, ini.Sections, 2)

		assert.Equal(t, "default", ini.Sections[0].Name)
		assert.Equal(t, "arn:aws:iam::123456789012:mfa/user", ini.Sections[0].Values["mfa_serial"])

		assert.Equal(t, "profile production", ini.Sections[1].Name)
		assert.Equal(t, "arn:aws:iam::123456789012:mfa/admin", ini.Sections[1].Values["mfa_serial"])
	})

	t.Run("skip comments and empty lines", func(t *testing.T) {
		data := []byte(`# This is a comment
; Another comment style

[default]
aws_access_key_id = AKIAEXAMPLE

# Comment between values
aws_secret_access_key = SECRET
`)
		ini := parseAWSINI(data)
		require.Len(t, ini.Sections, 1)
		assert.Len(t, ini.Sections[0].Values, 2)
	})

	t.Run("empty file", func(t *testing.T) {
		data := []byte(``)
		ini := parseAWSINI(data)
		assert.Empty(t, ini.Sections)
	})

	t.Run("preserve key order", func(t *testing.T) {
		data := []byte(`[default]
aws_access_key_id = AKIA
aws_secret_access_key = SECRET
aws_session_token = TOKEN
`)
		ini := parseAWSINI(data)
		require.Len(t, ini.Sections, 1)

		expected := []string{"aws_access_key_id", "aws_secret_access_key", "aws_session_token"}
		assert.Equal(t, expected, ini.Sections[0].KeyOrder)
	})
}

// TestWriteAWSINI tests the INI file writer
func TestWriteAWSINI(t *testing.T) {
	t.Run("write single section", func(t *testing.T) {
		ini := &awsINIFile{
			Sections: []*awsINISection{
				{
					Name: "default",
					Values: map[string]string{
						"aws_access_key_id":     "AKIAEXAMPLE",
						"aws_secret_access_key": "SECRET",
					},
					KeyOrder: []string{"aws_access_key_id", "aws_secret_access_key"},
				},
			},
		}

		output := writeAWSINI(ini)
		assert.Contains(t, string(output), "[default]")
		assert.Contains(t, string(output), "aws_access_key_id = AKIAEXAMPLE")
		assert.Contains(t, string(output), "aws_secret_access_key = SECRET")
	})

	t.Run("write multiple sections", func(t *testing.T) {
		ini := &awsINIFile{
			Sections: []*awsINISection{
				{
					Name: "default",
					Values: map[string]string{
						"aws_access_key_id": "AKIA1",
					},
					KeyOrder: []string{"aws_access_key_id"},
				},
				{
					Name: "production",
					Values: map[string]string{
						"aws_access_key_id": "AKIA2",
					},
					KeyOrder: []string{"aws_access_key_id"},
				},
			},
		}

		output := writeAWSINI(ini)
		assert.Contains(t, string(output), "[default]")
		assert.Contains(t, string(output), "[production]")
	})

	t.Run("roundtrip parse and write", func(t *testing.T) {
		original := `[default]
aws_access_key_id = AKIAEXAMPLE
aws_secret_access_key = SECRET
`
		ini := parseAWSINI([]byte(original))

		output := writeAWSINI(ini)

		// Parse again to verify
		ini2 := parseAWSINI(output)
		assert.Equal(t, ini.Sections[0].Values["aws_access_key_id"], ini2.Sections[0].Values["aws_access_key_id"])
	})
}

// TestGetOrCreateSection tests section creation/retrieval
func TestGetOrCreateSection(t *testing.T) {
	t.Run("get existing section", func(t *testing.T) {
		ini := &awsINIFile{
			Sections: []*awsINISection{
				{Name: "default", Values: map[string]string{"key": "value"}},
			},
		}

		section := getOrCreateSection(ini, "default")
		assert.Equal(t, "value", section.Values["key"])
	})

	t.Run("create new section", func(t *testing.T) {
		ini := &awsINIFile{Sections: []*awsINISection{}}

		section := getOrCreateSection(ini, "new-profile")
		assert.Equal(t, "new-profile", section.Name)
		assert.NotNil(t, section.Values)
		assert.Len(t, ini.Sections, 1)
	})
}

// TestSetINIValue tests setting values in sections
func TestSetINIValue(t *testing.T) {
	t.Run("set new value", func(t *testing.T) {
		section := &awsINISection{
			Name:     "default",
			Values:   make(map[string]string),
			KeyOrder: []string{},
		}

		setINIValue(section, "new_key", "new_value")
		assert.Equal(t, "new_value", section.Values["new_key"])
		assert.Contains(t, section.KeyOrder, "new_key")
	})

	t.Run("update existing value", func(t *testing.T) {
		section := &awsINISection{
			Name:     "default",
			Values:   map[string]string{"existing": "old"},
			KeyOrder: []string{"existing"},
		}

		setINIValue(section, "existing", "new")
		assert.Equal(t, "new", section.Values["existing"])
		// KeyOrder should not duplicate
		assert.Len(t, section.KeyOrder, 1)
	})
}

// TestMaskCredential tests credential masking
func TestMaskCredential(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard access key", "AKIAIOSFODNN7EXAMPLE", "AKIA************MPLE"},
		{"short string", "SHORT", "****"},
		{"exactly 8 chars", "12345678", "****"},
		{"9 chars", "123456789", "1234*6789"},
		{"empty string", "", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCredential(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateMFAToken tests MFA token validation via promptForMFAToken indirectly
func TestValidateMFAToken(t *testing.T) {
	// We test the validation logic through the error types
	t.Run("error type for invalid MFA", func(t *testing.T) {
		assert.Equal(t, "MFA token must be exactly 6 digits", errInvalidMFAToken.Error())
	})
}

// TestMFATokenPattern_EdgeCases tests MFA token pattern validation edge cases
func TestMFATokenPattern_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		token string
		valid bool
	}{
		{
			name:  "valid 6 digits",
			token: "123456",
			valid: true,
		},
		{
			name:  "leading zeros - 000000",
			token: "000000",
			valid: true,
		},
		{
			name:  "leading zeros - 000123",
			token: "000123",
			valid: true,
		},
		{
			name:  "all nines - 999999",
			token: "999999",
			valid: true,
		},
		{
			name:  "5 digits",
			token: "12345",
			valid: false,
		},
		{
			name:  "7 digits - 1000000",
			token: "1000000",
			valid: false,
		},
		{
			name:  "contains letters",
			token: "12a456",
			valid: false,
		},
		{
			name:  "all letters",
			token: "abcdef",
			valid: false,
		},
		{
			name:  "special characters",
			token: "123-56",
			valid: false,
		},
		{
			name:  "spaces in between",
			token: "123 456",
			valid: false,
		},
		{
			name:  "leading spaces",
			token: "  123456",
			valid: false,
		},
		{
			name:  "trailing spaces",
			token: "123456  ",
			valid: false,
		},
		{
			name:  "negative number",
			token: "-23456",
			valid: false,
		},
		{
			name:  "decimal number",
			token: "123.56",
			valid: false,
		},
		{
			name:  "empty string",
			token: "",
			valid: false,
		},
		{
			name:  "hex digits",
			token: "ABCDEF",
			valid: false,
		},
		{
			name:  "unicode digits",
			token: "①②③④⑤⑥",
			valid: false,
		},
		{
			name:  "tab character",
			token: "123\t456",
			valid: false,
		},
		{
			name:  "newline",
			token: "123456\n",
			valid: false,
		},
		{
			name:  "plus sign",
			token: "+23456",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mfaTokenPattern.MatchString(tt.token)
			assert.Equal(t, tt.valid, result, "MFA token validation mismatch")
		})
	}
}

// TestAWSConstants tests that constants are set correctly
func TestAWSConstants(t *testing.T) {
	t.Run("default duration is 12 hours", func(t *testing.T) {
		assert.Equal(t, 43200, awsDefaultDuration)
	})

	t.Run("default profile is 'default'", func(t *testing.T) {
		assert.Equal(t, "default", awsDefaultProfile)
	})

	t.Run("MFA token length is 6", func(t *testing.T) {
		assert.Equal(t, 6, mfaTokenLength)
	})
}

// TestAWSErrors tests error message content
func TestAWSErrors(t *testing.T) {
	t.Run("MFA serial not found error", func(t *testing.T) {
		assert.Contains(t, errMFASerialNotFound.Error(), "aws:setup")
	})
}

// TestBackupFile tests the backup functionality
func TestBackupFile(t *testing.T) {
	t.Run("backup non-existent file does nothing", func(t *testing.T) {
		err := backupFile("/non/existent/path/file.txt")
		assert.NoError(t, err)
	})

	t.Run("backup existing file", func(t *testing.T) {
		// Create temp dir and file
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test-creds")
		err := os.WriteFile(testFile, []byte("test content"), 0o600)
		require.NoError(t, err)

		// Backup
		err = backupFile(testFile)
		require.NoError(t, err)

		// Check backup exists
		backupPath := testFile + awsBackupSuffix
		_, err = os.Stat(backupPath)
		assert.NoError(t, err)

		// Verify content
		content, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})
}

// TestGetAWSDir tests the AWS directory path function
func TestGetAWSDir(t *testing.T) {
	t.Run("returns path ending with .aws", func(t *testing.T) {
		dir, err := getAWSDir()
		require.NoError(t, err)
		assert.Equal(t, ".aws", filepath.Base(dir))
	})

	t.Run("path is under home directory", func(t *testing.T) {
		dir, err := getAWSDir()
		require.NoError(t, err)

		home, err := os.UserHomeDir()
		require.NoError(t, err)

		assert.Equal(t, filepath.Join(home, ".aws"), dir)
	})
}

// TestWriteAWSCredentials tests credential file writing
func TestWriteAWSCredentials(t *testing.T) {
	t.Run("write new credentials file", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		err := writeAWSCredentials(credPath, "default", "AKIATEST", "SECRET123", "")
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(credPath)
		require.NoError(t, err)

		// Verify content
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "AKIATEST")
		assert.Contains(t, string(content), "SECRET123")
	})

	t.Run("write credentials with session token", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		err := writeAWSCredentials(credPath, "default", "AKIATEST", "SECRET", "SESSIONTOKEN")
		require.NoError(t, err)

		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "aws_session_token = SESSIONTOKEN")
	})

	t.Run("update existing profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Write initial credentials
		err := writeAWSCredentials(credPath, "default", "AKIA1", "SECRET1", "")
		require.NoError(t, err)

		// Update credentials
		err = writeAWSCredentials(credPath, "default", "AKIA2", "SECRET2", "")
		require.NoError(t, err)

		// Verify updated content
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "AKIA2")
		assert.NotContains(t, string(content), "AKIA1")
	})

	t.Run("add new profile to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Write first profile
		err := writeAWSCredentials(credPath, "default", "AKIA1", "SECRET1", "")
		require.NoError(t, err)

		// Write second profile
		err = writeAWSCredentials(credPath, "production", "AKIA2", "SECRET2", "")
		require.NoError(t, err)

		// Verify both profiles exist
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[default]")
		assert.Contains(t, string(content), "[production]")
		assert.Contains(t, string(content), "AKIA1")
		assert.Contains(t, string(content), "AKIA2")
	})

	t.Run("creates backup of existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Create initial file
		err := os.WriteFile(credPath, []byte("[default]\naws_access_key_id = OLD\n"), 0o600)
		require.NoError(t, err)

		// Write new credentials (should create backup)
		err = writeAWSCredentials(credPath, "default", "NEW", "SECRET", "")
		require.NoError(t, err)

		// Verify backup exists with old content
		backupPath := credPath + awsBackupSuffix
		backupContent, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Contains(t, string(backupContent), "OLD")
	})

	t.Run("file has correct permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		err := writeAWSCredentials(credPath, "default", "AKIA", "SECRET", "")
		require.NoError(t, err)

		info, err := os.Stat(credPath)
		require.NoError(t, err)
		// Check owner-only permissions (0600)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})
}

// TestWriteAWSConfig tests config file writing
func TestWriteAWSConfig(t *testing.T) {
	t.Run("write config for default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		err := writeAWSConfig(configPath, "default", "arn:aws:iam::123456:mfa/user")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[default]")
		assert.Contains(t, string(content), "mfa_serial = arn:aws:iam::123456:mfa/user")
	})

	t.Run("write config for non-default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		err := writeAWSConfig(configPath, "production", "arn:aws:iam::123456:mfa/admin")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		// Non-default profiles should be prefixed with "profile "
		assert.Contains(t, string(content), "[profile production]")
	})
}

// TestHasValidAWSSetup tests setup detection
func TestHasValidAWSSetup(t *testing.T) {
	t.Run("no aws directory returns false", func(t *testing.T) {
		// Save original HOME and restore after test
		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()

		// Set HOME to non-existent directory
		tmpDir := t.TempDir()
		nonExistent := filepath.Join(tmpDir, "nonexistent")
		_ = os.Setenv("HOME", nonExistent)

		result := hasValidAWSSetup("default")
		assert.False(t, result)
	})
}

// TestAWSNamespaceType tests that AWS is a valid namespace type
func TestAWSNamespaceType(t *testing.T) {
	t.Run("AWS type exists", func(t *testing.T) {
		var aws AWS
		_ = aws // Verify type exists
	})
}

// TestWriteOrUpdateAWSSessionCredentials tests session credential updates
func TestWriteOrUpdateAWSSessionCredentials(t *testing.T) {
	t.Run("update existing profile with session credentials", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Create initial credentials file
		initial := `[default]
aws_access_key_id = AKIA_ORIGINAL
aws_secret_access_key = SECRET_ORIGINAL
`
		err := os.WriteFile(credPath, []byte(initial), 0o600)
		require.NoError(t, err)

		// Update with session credentials
		creds := &awsSTSCredentials{
			AccessKeyID:     "ASIA_SESSION",
			SecretAccessKey: "SECRET_SESSION",
			SessionToken:    "TOKEN_SESSION",
			Expiration:      "2024-01-01T00:00:00Z",
		}

		err = writeOrUpdateAWSSessionCredentials(credPath, "default", creds)
		require.NoError(t, err)

		// Verify updated content
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "ASIA_SESSION")
		assert.Contains(t, string(content), "SECRET_SESSION")
		assert.Contains(t, string(content), "TOKEN_SESSION")
	})

	t.Run("create new profile if not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Create credentials with only default profile
		initial := `[default]
aws_access_key_id = AKIA
`
		err := os.WriteFile(credPath, []byte(initial), 0o600)
		require.NoError(t, err)

		creds := &awsSTSCredentials{
			AccessKeyID:     "ASIA",
			SecretAccessKey: "SECRET",
			SessionToken:    "TOKEN",
		}

		// Should create new profile
		err = writeOrUpdateAWSSessionCredentials(credPath, "newprofile", creds)
		require.NoError(t, err)

		// Verify both profiles exist
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[default]")
		assert.Contains(t, string(content), "[newprofile]")
	})

	t.Run("create file if not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "newcredentials")

		creds := &awsSTSCredentials{
			AccessKeyID:     "ASIA",
			SecretAccessKey: "SECRET",
			SessionToken:    "TOKEN",
		}

		err := writeOrUpdateAWSSessionCredentials(credPath, "default", creds)
		require.NoError(t, err)

		// Verify file was created
		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[default]")
		assert.Contains(t, string(content), "ASIA")
	})
}

// TestGetSourceProfile tests source_profile retrieval
func TestGetSourceProfile(t *testing.T) {
	t.Run("get source_profile for non-default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		// Write config with source_profile
		configContent := `[profile mrz]
source_profile = mrz-base
region = us-east-1
`
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0o600)
		require.NoError(t, err)

		// Override HOME for this test
		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		sourceProfile := getSourceProfile("mrz")
		assert.Equal(t, "mrz-base", sourceProfile)
	})

	t.Run("return empty when source_profile not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		// Write config without source_profile
		configContent := `[profile mrz]
region = us-east-1
`
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0o600)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		sourceProfile := getSourceProfile("mrz")
		assert.Empty(t, sourceProfile)
	})

	t.Run("return empty when config file missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		sourceProfile := getSourceProfile("mrz")
		assert.Empty(t, sourceProfile)
	})
}

// TestWriteAWSConfigSourceProfile tests writing source_profile
func TestWriteAWSConfigSourceProfile(t *testing.T) {
	t.Run("write source_profile for non-default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		err := writeAWSConfigSourceProfile(configPath, "mrz", "mrz-base")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[profile mrz]")
		assert.Contains(t, string(content), "source_profile = mrz-base")
	})

	t.Run("write source_profile for default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		err := writeAWSConfigSourceProfile(configPath, "default", "default-base")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[default]")
		assert.Contains(t, string(content), "source_profile = default-base")
	})

	t.Run("update existing profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config")

		// Write initial config
		initial := `[profile mrz]
region = us-east-1
`
		err := os.WriteFile(configPath, []byte(initial), 0o600)
		require.NoError(t, err)

		// Add source_profile
		err = writeAWSConfigSourceProfile(configPath, "mrz", "mrz-base")
		require.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "[profile mrz]")
		assert.Contains(t, string(content), "source_profile = mrz-base")
		assert.Contains(t, string(content), "region = us-east-1")
	})
}

// TestGetMFASerial tests MFA serial retrieval
func TestGetMFASerial(t *testing.T) {
	t.Run("get MFA serial for default profile", func(t *testing.T) {
		// Create temp AWS directory structure
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		// Write config file
		configContent := `[default]
mfa_serial = arn:aws:iam::123456789012:mfa/testuser
region = us-east-1
`
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0o600)
		require.NoError(t, err)

		// Override HOME for this test
		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		serial, err := getMFASerial("default")
		require.NoError(t, err)
		assert.Equal(t, "arn:aws:iam::123456789012:mfa/testuser", serial)
	})

	t.Run("get MFA serial for non-default profile", func(t *testing.T) {
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		// Write config with profile prefix
		configContent := `[profile production]
mfa_serial = arn:aws:iam::123456789012:mfa/admin
`
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0o600)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		serial, err := getMFASerial("production")
		require.NoError(t, err)
		assert.Equal(t, "arn:aws:iam::123456789012:mfa/admin", serial)
	})

	t.Run("error when MFA serial not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		awsDir := filepath.Join(tmpDir, ".aws")
		err := os.MkdirAll(awsDir, 0o700)
		require.NoError(t, err)

		// Write config without MFA serial
		configContent := `[default]
region = us-east-1
`
		err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0o600)
		require.NoError(t, err)

		originalHome := os.Getenv("HOME")
		defer func() {
			_ = os.Setenv("HOME", originalHome)
		}()
		_ = os.Setenv("HOME", tmpDir)

		_, err = getMFASerial("default")
		assert.ErrorIs(t, err, errMFASerialNotFound)
	})
}

// TestAWSSTSCredentials tests the STS credentials struct
func TestAWSSTSCredentials(t *testing.T) {
	t.Run("struct fields", func(t *testing.T) {
		creds := awsSTSCredentials{
			AccessKeyID:     "ASIAXXX",
			SecretAccessKey: "SECRET",
			SessionToken:    "TOKEN",
			Expiration:      "2024-01-01T12:00:00Z",
		}

		assert.Equal(t, "ASIAXXX", creds.AccessKeyID)
		assert.Equal(t, "SECRET", creds.SecretAccessKey)
		assert.Equal(t, "TOKEN", creds.SessionToken)
		assert.Equal(t, "2024-01-01T12:00:00Z", creds.Expiration)
	})
}

// TestPromptForNonEmpty tests the non-empty validation prompt
func TestPromptForNonEmpty(t *testing.T) {
	// This function reads from stdin, so we test the error message format
	t.Run("error message format", func(t *testing.T) {
		// The function returns errEmptyInput when input is empty
		// We verify the error exists and has a meaningful message
		assert.NotEmpty(t, errEmptyInput.Error())
		assert.Contains(t, errEmptyInput.Error(), "empty")
	})
}

// awsTestRunner is a simple mock runner for AWS tests
type awsTestRunner struct {
	runCmdOutputFunc func(cmd string, args ...string) (string, error)
	capturedArgs     []string
	allCalls         [][]string // each entry is [cmd, args...] in invocation order
}

func (r *awsTestRunner) RunCmd(name string, args ...string) error {
	call := append([]string{name}, args...)
	r.allCalls = append(r.allCalls, call)
	return nil
}

func (r *awsTestRunner) RunCmdOutput(name string, args ...string) (string, error) {
	r.capturedArgs = args
	call := append([]string{name}, args...)
	r.allCalls = append(r.allCalls, call)
	if r.runCmdOutputFunc != nil {
		return r.runCmdOutputFunc(name, args...)
	}
	return "", nil
}

// TestCheckAWSSession tests AWS session validation
func TestCheckAWSSession(t *testing.T) {
	t.Run("valid session returns account and ARN", func(t *testing.T) {
		// Save original runner and restore after test
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }()

		// Create mock runner that returns valid STS response
		mockRunner := &awsTestRunner{
			runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
				if cmd == "aws" && len(args) > 0 && args[0] == "sts" {
					return `{"Account": "123456789012", "Arn": "arn:aws:sts::123456789012:assumed-role/TestRole/session", "UserId": "AROA123456:session"}`, nil
				}
				return "", nil
			},
		}
		_ = SetRunner(mockRunner)

		accountID, arn, isValid := checkAWSSession("default")
		assert.True(t, isValid)
		assert.Equal(t, "123456789012", accountID)
		assert.Equal(t, "arn:aws:sts::123456789012:assumed-role/TestRole/session", arn)
	})

	t.Run("invalid session returns false", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }()

		mockRunner := &awsTestRunner{
			runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
				if cmd == "aws" && len(args) > 0 && args[0] == "sts" {
					return "", fmt.Errorf("ExpiredToken: The security token included in the request is expired")
				}
				return "", nil
			},
		}
		_ = SetRunner(mockRunner)

		accountID, arn, isValid := checkAWSSession("expired-profile")
		assert.False(t, isValid)
		assert.Empty(t, accountID)
		assert.Empty(t, arn)
	})

	t.Run("invalid JSON response returns false", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }()

		mockRunner := &awsTestRunner{
			runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
				return "not valid json", nil
			},
		}
		_ = SetRunner(mockRunner)

		accountID, arn, isValid := checkAWSSession("bad-json-profile")
		assert.False(t, isValid)
		assert.Empty(t, accountID)
		assert.Empty(t, arn)
	})

	t.Run("uses profile flag for non-default profiles", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }()

		mockRunner := &awsTestRunner{
			runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
				return `{"Account": "123", "Arn": "arn", "UserId": "user"}`, nil
			},
		}
		_ = SetRunner(mockRunner)

		checkAWSSession("production")
		assert.Contains(t, mockRunner.capturedArgs, "--profile")
		assert.Contains(t, mockRunner.capturedArgs, "production")
	})

	t.Run("does not use profile flag for default", func(t *testing.T) {
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }()

		mockRunner := &awsTestRunner{
			runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
				return `{"Account": "123", "Arn": "arn", "UserId": "user"}`, nil
			},
		}
		_ = SetRunner(mockRunner)

		checkAWSSession("default")
		assert.NotContains(t, mockRunner.capturedArgs, "--profile")
	})
}

// TestNoArgsPlaceholders tests the NoArgs placeholder methods
func TestNoArgsPlaceholders(t *testing.T) {
	// These methods just call the main methods without args
	// We verify they exist and have the right signature
	var aws AWS

	t.Run("LoginNoArgs exists", func(t *testing.T) {
		// Just verify the method exists with correct signature
		fn := aws.LoginNoArgs
		_ = fn
	})

	t.Run("SetupNoArgs exists", func(t *testing.T) {
		fn := aws.SetupNoArgs
		_ = fn
	})

	t.Run("RefreshNoArgs exists", func(t *testing.T) {
		fn := aws.RefreshNoArgs
		_ = fn
	})

	t.Run("StatusNoArgs exists", func(t *testing.T) {
		fn := aws.StatusNoArgs
		_ = fn
	})
}

// TestCheckAWSCLI tests AWS CLI detection
func TestCheckAWSCLI(t *testing.T) {
	t.Run("AWS CLI found in PATH", func(t *testing.T) {
		// Skip if AWS CLI not actually installed
		if !utils.CommandExists("aws") {
			t.Skip("AWS CLI not installed")
		}

		err := checkAWSCLI()
		assert.NoError(t, err)
	})

	t.Run("AWS CLI not found returns helpful error", func(t *testing.T) {
		err := getAWSCLINotFoundError()
		assert.Error(t, err)

		errMsg := err.Error()
		assert.Contains(t, errMsg, "AWS CLI not found")
		assert.Contains(t, errMsg, "Install it using")

		// Should contain platform-specific guidance
		if utils.IsWindows() {
			assert.Contains(t, errMsg, "awscli.amazonaws.com/AWSCLIV2.msi")
		} else if utils.IsMac() {
			assert.Contains(t, errMsg, "brew install awscli")
		} else {
			assert.Contains(t, errMsg, "apt-get install awscli")
		}
	})

	t.Run("error message includes PATH guidance", func(t *testing.T) {
		err := getAWSCLINotFoundError()
		assert.Contains(t, err.Error(), "PATH")
	})

	t.Run("error message includes official docs", func(t *testing.T) {
		err := getAWSCLINotFoundError()
		assert.Contains(t, err.Error(), awsInstallURL)
	})
}

// TestAWSCLIErrorMessageFormat tests error message formatting
func TestAWSCLIErrorMessageFormat(t *testing.T) {
	t.Run("error is multiline for readability", func(t *testing.T) {
		err := getAWSCLINotFoundError()
		assert.Contains(t, err.Error(), "\n")
	})

	t.Run("error message not too long", func(t *testing.T) {
		err := getAWSCLINotFoundError()
		// Should be helpful but concise (under 500 chars)
		assert.Less(t, len(err.Error()), 500)
	})
}

// ============================================================================
// Test helpers for end-to-end AWS flow tests
//
// Note: file I/O on ~/.aws/credentials is not serialized inside aws.go, so
// concurrent calls to Setup/Refresh are not tested here.
// ============================================================================

// Sentinel errors used by mock runners. Defined statically to satisfy err113.
var (
	errTestSTSUnavailable           = errors.New("test: STS unavailable in unit test")
	errTestUnexpectedRunnerCall     = errors.New("test: unexpected runner call")
	errTestUnexpectedSTSDuringSetup = errors.New("test: unexpected STS call during Setup")
	errTestSetupRunnerInvoked       = errors.New("test: Setup should not invoke runner")
	errTestSTSDuringInvalidMFA      = errors.New("test: STS should not be invoked when MFA token is malformed")
)

// stsSessionExpiry is a fixed expiration string used in canned STS responses.
const stsSessionExpiry = "2030-01-01T00:00:00Z"

// linesReader is an io.Reader that returns one queued line per Read call,
// preventing bufio.Scanner from over-buffering when PromptForInput creates a
// fresh Scanner for each prompt.
type linesReader struct {
	pending []byte
	lines   []string
	idx     int
}

func (lr *linesReader) Read(p []byte) (int, error) {
	if len(lr.pending) == 0 {
		if lr.idx >= len(lr.lines) {
			return 0, io.EOF
		}
		lr.pending = []byte(lr.lines[lr.idx] + "\n")
		lr.idx++
	}
	n := copy(p, lr.pending)
	lr.pending = lr.pending[n:]
	return n, nil
}

// withMockedHome seeds ~/.aws/credentials and ~/.aws/config in a temp HOME.
// Empty strings are skipped (the corresponding file is not written).
func withMockedHome(t *testing.T, credsContent, configContent string) {
	t.Helper()
	tmpDir := t.TempDir()
	awsDir := filepath.Join(tmpDir, ".aws")
	require.NoError(t, os.MkdirAll(awsDir, 0o700))
	if credsContent != "" {
		require.NoError(t, os.WriteFile(filepath.Join(awsDir, awsCredentialsFile), []byte(credsContent), 0o600))
	}
	if configContent != "" {
		require.NoError(t, os.WriteFile(filepath.Join(awsDir, awsConfigFile), []byte(configContent), 0o600))
	}
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() { _ = os.Setenv("HOME", originalHome) }) //nolint:errcheck // test cleanup
	require.NoError(t, os.Setenv("HOME", tmpDir))
}

// withMockedPrompts queues stdin inputs for utils.PromptForInput. Each call
// to PromptForInput consumes one line in order.
func withMockedPrompts(t *testing.T, inputs []string) {
	t.Helper()
	prev := utils.SetPromptInput(&linesReader{lines: inputs})
	t.Cleanup(func() { utils.SetPromptInput(prev) })
}

// withMockedRunner installs an awsTestRunner and restores the original on
// cleanup. Returns the mock for argument inspection.
func withMockedRunner(t *testing.T, fn func(cmd string, args ...string) (string, error)) *awsTestRunner {
	t.Helper()
	original := GetRunner()
	mock := &awsTestRunner{runCmdOutputFunc: fn}
	require.NoError(t, SetRunner(mock))
	t.Cleanup(func() { _ = SetRunner(original) }) //nolint:errcheck // test cleanup; SetRunner only fails on nil
	return mock
}

// stsSessionTokenJSON returns a canned STS get-session-token response.
func stsSessionTokenJSON(accessKey, secret, token string) string {
	return fmt.Sprintf(`{
  "Credentials": {
    "AccessKeyId": "%s",
    "SecretAccessKey": "%s",
    "SessionToken": "%s",
    "Expiration": "%s"
  }
}`, accessKey, secret, token, stsSessionExpiry)
}

// readCredentialsINI reads and parses ~/.aws/credentials under the mocked HOME.
func readCredentialsINI(t *testing.T) *awsINIFile {
	t.Helper()
	awsDir, err := getAWSDir()
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(awsDir, awsCredentialsFile)) //nolint:gosec // path constructed from test temp dir
	require.NoError(t, err)
	return parseAWSINI(data)
}

// readConfigINI reads and parses ~/.aws/config under the mocked HOME.
func readConfigINI(t *testing.T) *awsINIFile {
	t.Helper()
	awsDir, err := getAWSDir()
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(awsDir, awsConfigFile)) //nolint:gosec // path constructed from test temp dir
	require.NoError(t, err)
	return parseAWSINI(data)
}

// findSection returns the named section from a parsed INI, or nil if absent.
func findSection(ini *awsINIFile, name string) *awsINISection {
	if ini == nil {
		return nil
	}
	for _, section := range ini.Sections {
		if section.Name == name {
			return section
		}
	}
	return nil
}

// argsContain returns true when target appears as a contiguous subsequence of args.
func argsContain(args []string, target ...string) bool {
	if len(target) == 0 || len(target) > len(args) {
		return false
	}
	for i := 0; i+len(target) <= len(args); i++ {
		match := true
		for j, want := range target {
			if args[i+j] != want {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// skipIfNoAWSCLI matches the existing skip pattern in this file. The Login,
// Setup, and Refresh entry points all call checkAWSCLI which probes $PATH.
func skipIfNoAWSCLI(t *testing.T) {
	t.Helper()
	if !utils.CommandExists("aws") {
		t.Skip("AWS CLI not installed")
	}
}

// ============================================================================
// hasValidAWSSetup — table-driven coverage of session/base resolution
// ============================================================================

const baseProfileMFA = "arn:aws:iam::123456789012:mfa/test"

func TestHasValidAWSSetup_Table(t *testing.T) {
	tests := []struct {
		name       string
		makeAWSDir bool
		credsFile  string
		config     string
		profile    string
		want       bool
	}{
		{
			name:       "no .aws directory",
			makeAWSDir: false,
			profile:    "default",
			want:       false,
		},
		{
			name:       "credentials file missing",
			makeAWSDir: true,
			config:     "[default]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "default",
			want:       false,
		},
		{
			name:       "profile absent from credentials",
			makeAWSDir: true,
			credsFile:  "[other]\n" + "aws_access_key_id = AKIATEST\n",
			profile:    "default",
			want:       false,
		},
		{
			name:       "base half-done: key present but no mfa_serial in config",
			makeAWSDir: true,
			credsFile:  "[mrz-ro-base]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile mrz-ro]\nsource_profile = mrz-ro-base\n",
			profile:    "mrz-ro",
			want:       false,
		},
		{
			name:       "post-setup pre-refresh (primary bug-fix case)",
			makeAWSDir: true,
			credsFile: "[mrz-ro-base]\n" +
				"aws_access_key_id = AKIATEST\n" +
				"aws_secret_access_key = SECRETTEST\n",
			config: "[profile mrz-ro-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile mrz-ro]\nsource_profile = mrz-ro-base\n",
			profile: "mrz-ro",
			want:    true,
		},
		{
			name:       "post-refresh has both base and session sections",
			makeAWSDir: true,
			credsFile: "[mrz-ro-base]\naws_access_key_id = AKIATEST\n\n" +
				"[mrz-ro]\naws_access_key_id = ASIASESSION\naws_session_token = TOK\n",
			config: "[profile mrz-ro-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile mrz-ro]\nsource_profile = mrz-ro-base\n",
			profile: "mrz-ro",
			want:    true,
		},
		{
			name:       "legacy single-profile default",
			makeAWSDir: true,
			credsFile:  "[default]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[default]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "default",
			want:       true,
		},
		{
			name:       "legacy single-profile non-default",
			makeAWSDir: true,
			credsFile:  "[prod]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile prod]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "prod",
			want:       true,
		},
		{
			name:       "dangling source_profile points to missing base",
			makeAWSDir: true,
			credsFile:  "[other-base]\n" + "aws_access_key_id = AKIATEST\n",
			config: "[profile mrz-ro]\nsource_profile = mrz-ro-base\n\n" +
				"[profile other-base]\nmfa_serial = " + baseProfileMFA + "\n",
			profile: "mrz-ro",
			want:    false,
		},
		{
			name:       "user passes base profile name directly",
			makeAWSDir: true,
			credsFile:  "[mrz-ro-base]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile mrz-ro-base]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "mrz-ro-base",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalHome := os.Getenv("HOME")
			t.Cleanup(func() { _ = os.Setenv("HOME", originalHome) }) //nolint:errcheck // test cleanup

			tmpDir := t.TempDir()
			require.NoError(t, os.Setenv("HOME", tmpDir))

			if tt.makeAWSDir {
				awsDir := filepath.Join(tmpDir, ".aws")
				require.NoError(t, os.MkdirAll(awsDir, 0o700))
				if tt.credsFile != "" {
					require.NoError(t, os.WriteFile(filepath.Join(awsDir, awsCredentialsFile), []byte(tt.credsFile), 0o600))
				}
				if tt.config != "" {
					require.NoError(t, os.WriteFile(filepath.Join(awsDir, awsConfigFile), []byte(tt.config), 0o600))
				}
			}

			assert.Equal(t, tt.want, hasValidAWSSetup(tt.profile))
		})
	}
}

// ============================================================================
// Status — base/session profile resolution via runner-arg assertions
//
// Status displays each matched section by calling checkAWSSession(name),
// which routes through GetRunner().RunCmdOutput. By making the mock runner
// fail the STS call, displayAWSProfileStatus completes without panicking;
// the captured args list tells us which sections were matched.
// ============================================================================

func TestStatusBaseProfileResolution(t *testing.T) {
	const credsBothSections = "[mrz-ro-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n\n" +
		"[mrz-ro]\n" +
		"aws_access_key_id = ASIASESSION\n" +
		"aws_secret_access_key = SECRETSESSION\n" +
		"aws_session_token = TOKSESSION\n"

	const credsBaseOnly = "[mrz-ro-base]\n" +
		"aws_access_key_id = AKIABASE\n"

	const configBoth = "[profile mrz-ro-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile mrz-ro]\n" +
		"source_profile = mrz-ro-base\n"

	// extractProfilesFromCalls returns the profile names checkAWSSession was
	// invoked for, derived from "--profile <name>" pairs in captured args.
	extractProfiles := func(calls [][]string) []string {
		var profiles []string
		for _, call := range calls {
			for i := 0; i < len(call)-1; i++ {
				if call[i] == "--profile" {
					profiles = append(profiles, call[i+1])
					break
				}
			}
		}
		return profiles
	}

	t.Run("filter by session profile shows base section post-setup", func(t *testing.T) {
		withMockedHome(t, credsBaseOnly, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=mrz-ro")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "mrz-ro-base", "Status should display base profile when filtered by session name")
	})

	t.Run("filter by session profile post-refresh shows both sections", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=mrz-ro")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "mrz-ro-base")
		assert.Contains(t, profiles, "mrz-ro")
	})

	t.Run("filter by base profile directly shows only base", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=mrz-ro-base")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "mrz-ro-base")
		assert.NotContains(t, profiles, "mrz-ro", "filtering by base name should not double-print session section")
	})

	t.Run("no filter displays every section", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status()
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.ElementsMatch(t, []string{"mrz-ro-base", "mrz-ro"}, profiles)
	})

	t.Run("filter that matches nothing returns nil and skips STS calls", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=ghost")
		require.NoError(t, err)
		assert.Empty(t, mock.allCalls, "no sections matched, so no STS calls expected")
	})

	t.Run("missing credentials file returns nil without panic", func(t *testing.T) {
		withMockedHome(t, "", "")
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=mrz-ro")
		require.NoError(t, err)
		assert.Empty(t, mock.allCalls)
	})
}

// ============================================================================
// Login — end-to-end flow tests
// ============================================================================

func TestLoginFlow(t *testing.T) {
	const credsBaseOnly = "[mrz-ro-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"

	const credsBothSections = "[mrz-ro-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n\n" +
		"[mrz-ro]\n" +
		"aws_access_key_id = ASIASTALE\n" +
		"aws_secret_access_key = SECRETSTALE\n" +
		"aws_session_token = TOKSTALE\n"

	const configBoth = "[profile mrz-ro-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile mrz-ro]\n" +
		"source_profile = mrz-ro-base\n"

	t.Run("login_after_setup_before_refresh_enters_refresh_path (regression)", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, configBoth)
		withMockedPrompts(t, []string{"123456"})

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				return stsSessionTokenJSON("ASIANEW", "SECRETNEW", "TOKNEW"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		err := AWS{}.Login("profile=mrz-ro")
		require.NoError(t, err)

		// The STS call must target the BASE profile (long-term keys).
		require.NotEmpty(t, mock.allCalls)
		stsArgs := mock.allCalls[len(mock.allCalls)-1]
		assert.True(t, argsContain(stsArgs, "--profile", "mrz-ro-base"),
			"Refresh path must invoke STS with --profile mrz-ro-base; got %v", stsArgs)
		assert.True(t, argsContain(stsArgs, "--token-code", "123456"))

		// Credentials file should now contain the session section with the new token.
		ini := readCredentialsINI(t)
		session := findSection(ini, "mrz-ro")
		require.NotNil(t, session, "session profile [mrz-ro] must be written after Refresh")
		assert.Equal(t, "TOKNEW", session.Values["aws_session_token"])
		assert.Equal(t, "ASIANEW", session.Values["aws_access_key_id"])
	})

	t.Run("login_with_missing_setup_enters_setup_path", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, "", "")
		withMockedPrompts(t, []string{
			"AKIATEST",
			"SECRETTEST",
			"arn:aws:iam::1:mfa/u",
		})
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestUnexpectedSTSDuringSetup
		})

		err := AWS{}.Login("profile=mrz-ro")
		require.NoError(t, err)

		assert.Empty(t, mock.allCalls, "Setup path must not invoke STS")

		creds := readCredentialsINI(t)
		base := findSection(creds, "mrz-ro-base")
		require.NotNil(t, base, "[mrz-ro-base] must be written to credentials")
		assert.Equal(t, "AKIATEST", base.Values["aws_access_key_id"])
		assert.Equal(t, "SECRETTEST", base.Values["aws_secret_access_key"])
		// Session profile section must NOT be written by Setup.
		assert.Nil(t, findSection(creds, "mrz-ro"))

		config := readConfigINI(t)
		baseCfg := findSection(config, "profile mrz-ro-base")
		require.NotNil(t, baseCfg, "[profile mrz-ro-base] must be written to config")
		assert.Equal(t, "arn:aws:iam::1:mfa/u", baseCfg.Values["mfa_serial"])

		sessionCfg := findSection(config, "profile mrz-ro")
		require.NotNil(t, sessionCfg, "[profile mrz-ro] must be written to config")
		assert.Equal(t, "mrz-ro-base", sessionCfg.Values["source_profile"])
	})

	t.Run("login_with_healthy_setup_refreshes_session_token", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBothSections, configBoth)
		withMockedPrompts(t, []string{"123456"})

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				return stsSessionTokenJSON("ASIAFRESH", "SECRETFRESH", "TOKFRESH"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		err := AWS{}.Login("profile=mrz-ro")
		require.NoError(t, err)
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "mrz-ro")
		require.NotNil(t, session)
		assert.Equal(t, "TOKFRESH", session.Values["aws_session_token"])
	})
}

// ============================================================================
// Setup — exercises the interactive prompt path end-to-end
// ============================================================================

func TestSetupWritesBothProfileSections(t *testing.T) {
	skipIfNoAWSCLI(t)

	withMockedHome(t, "", "")
	withMockedPrompts(t, []string{
		"AKIATEST",
		"SECRETTEST",
		"arn:aws:iam::123456789012:mfa/test",
	})
	withMockedRunner(t, func(cmd string, args ...string) (string, error) {
		return "", errTestSetupRunnerInvoked
	})

	require.NoError(t, AWS{}.Setup("profile=mrz-ro"))

	creds := readCredentialsINI(t)
	base := findSection(creds, "mrz-ro-base")
	require.NotNil(t, base)
	assert.Equal(t, "AKIATEST", base.Values["aws_access_key_id"])
	assert.Equal(t, "SECRETTEST", base.Values["aws_secret_access_key"])
	assert.Nil(t, findSection(creds, "mrz-ro"), "Setup must not write the session profile to credentials")

	config := readConfigINI(t)
	baseCfg := findSection(config, "profile mrz-ro-base")
	require.NotNil(t, baseCfg)
	assert.Equal(t, "arn:aws:iam::123456789012:mfa/test", baseCfg.Values["mfa_serial"])

	sessionCfg := findSection(config, "profile mrz-ro")
	require.NotNil(t, sessionCfg)
	assert.Equal(t, "mrz-ro-base", sessionCfg.Values["source_profile"])

	// Re-running Setup creates .bak files for both credentials and config.
	withMockedPrompts(t, []string{
		"AKIATWO",
		"SECRETTWO",
		"arn:aws:iam::123456789012:mfa/test",
	})
	require.NoError(t, AWS{}.Setup("profile=mrz-ro"))

	awsDir, err := getAWSDir()
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(awsDir, awsCredentialsFile+awsBackupSuffix))
	assert.FileExists(t, filepath.Join(awsDir, awsConfigFile+awsBackupSuffix))

	creds = readCredentialsINI(t)
	base = findSection(creds, "mrz-ro-base")
	require.NotNil(t, base)
	assert.Equal(t, "AKIATWO", base.Values["aws_access_key_id"])
}

// ============================================================================
// Refresh — exercises MFA prompt, STS call, and session-credential write
// ============================================================================

func TestRefreshFlow(t *testing.T) {
	const credsBaseOnly = "[mrz-ro-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const configBoth = "[profile mrz-ro-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile mrz-ro]\n" +
		"source_profile = mrz-ro-base\n"

	t.Run("writes session creds under session profile", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, configBoth)
		withMockedPrompts(t, []string{"123456"})

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				assert.True(t, argsContain(args, "--serial-number", baseProfileMFA))
				assert.True(t, argsContain(args, "--token-code", "123456"))
				assert.True(t, argsContain(args, "--profile", "mrz-ro-base"))
				return stsSessionTokenJSON("ASIASESSION", "SECSESSION", "TOKSESSION"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		require.NoError(t, AWS{}.Refresh("profile=mrz-ro"))
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "mrz-ro")
		require.NotNil(t, session)
		assert.Equal(t, "ASIASESSION", session.Values["aws_access_key_id"])
		assert.Equal(t, "SECSESSION", session.Values["aws_secret_access_key"])
		assert.Equal(t, "TOKSESSION", session.Values["aws_session_token"])
	})

	t.Run("invalid MFA format errors before STS is invoked", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, configBoth)
		withMockedPrompts(t, []string{"12345"}) // 5 digits

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSDuringInvalidMFA
		})

		err := AWS{}.Refresh("profile=mrz-ro")
		require.Error(t, err)
		require.ErrorIs(t, err, errInvalidMFAToken, "expected errInvalidMFAToken, got: %v", err)
		assert.Empty(t, mock.allCalls, "runner must not be invoked when validation fails")
	})

	t.Run("explicit base= overrides source_profile lookup", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t,
			"[custom-base]\naws_access_key_id = AKIACUSTOM\n",
			"[profile custom-base]\nmfa_serial = "+baseProfileMFA+"\n",
		)
		withMockedPrompts(t, []string{"123456"})

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				assert.True(t, argsContain(args, "--profile", "custom-base"))
				return stsSessionTokenJSON("ASIANEW", "SECRETNEW", "TOKNEW"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		require.NoError(t, AWS{}.Refresh("profile=mrz-ro", "base=custom-base"))
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "mrz-ro")
		require.NotNil(t, session)
		assert.Equal(t, "TOKNEW", session.Values["aws_session_token"])
	})
}

// ============================================================================
// Edge cases
// ============================================================================

func TestEdgeCases(t *testing.T) {
	t.Run("source_profile is resolved only one level deep", func(t *testing.T) {
		// a -> b -> c (creds + MFA only on c). getSourceProfile is non-recursive,
		// so hasValidAWSSetup("a") looks for credentials on "b" and finds none.
		withMockedHome(t,
			"[c]\naws_access_key_id = AKIATEST\n",
			"[profile a]\nsource_profile = b\n\n"+
				"[profile b]\nsource_profile = c\n\n"+
				"[profile c]\nmfa_serial = "+baseProfileMFA+"\n",
		)
		assert.False(t, hasValidAWSSetup("a"))
	})
}
