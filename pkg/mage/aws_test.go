//go:build unit
// +build unit

package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		assert.Len(t, ini.Sections, 0)
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
	t.Run("AWS CLI not found error", func(t *testing.T) {
		assert.Contains(t, errAWSCLINotFound.Error(), "AWS CLI not found")
		assert.Contains(t, errAWSCLINotFound.Error(), "https://")
	})

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
		assert.True(t, filepath.Base(dir) == ".aws")
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
		assert.Equal(t, "", sourceProfile)
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
		assert.Equal(t, "", sourceProfile)
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
}

func (r *awsTestRunner) RunCmd(name string, args ...string) error {
	return nil
}

func (r *awsTestRunner) RunCmdOutput(name string, args ...string) (string, error) {
	r.capturedArgs = args
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
		var fn func() error = aws.LoginNoArgs
		_ = fn
	})

	t.Run("SetupNoArgs exists", func(t *testing.T) {
		var fn func() error = aws.SetupNoArgs
		_ = fn
	})

	t.Run("RefreshNoArgs exists", func(t *testing.T) {
		var fn func() error = aws.RefreshNoArgs
		_ = fn
	})

	t.Run("StatusNoArgs exists", func(t *testing.T) {
		var fn func() error = aws.StatusNoArgs
		_ = fn
	})
}
