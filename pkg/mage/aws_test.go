//go:build unit
// +build unit

package mage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkglog "github.com/mrz1836/mage-x/pkg/log"
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

// TestDeleteINIValue tests removing values from sections
func TestDeleteINIValue(t *testing.T) {
	newSection := func() *awsINISection {
		return &awsINISection{
			Name:     "default",
			Values:   map[string]string{"a": "1", "b": "2", "c": "3"},
			KeyOrder: []string{"a", "b", "c"},
		}
	}

	t.Run("removes the key and keeps the order of the rest", func(t *testing.T) {
		section := newSection()

		deleteINIValue(section, "b")
		assert.NotContains(t, section.Values, "b")
		assert.Equal(t, []string{"a", "c"}, section.KeyOrder)
	})

	t.Run("missing key is a no-op", func(t *testing.T) {
		section := newSection()

		deleteINIValue(section, "missing")
		assert.Len(t, section.Values, 3)
		assert.Equal(t, []string{"a", "b", "c"}, section.KeyOrder)
	})

	t.Run("re-adding a removed key writes it once", func(t *testing.T) {
		section := newSection()

		deleteINIValue(section, "b")
		setINIValue(section, "b", "4")
		ini := &awsINIFile{Sections: []*awsINISection{section}}
		assert.Equal(t, "[default]\na = 1\nc = 3\nb = 4\n", string(writeAWSINI(ini)))
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
		backups := listBackups(testFile)
		require.Len(t, backups, 1)
		assert.True(t, strings.HasPrefix(filepath.Base(backups[0]), "test-creds.bak."), "got %s", backups[0])

		// Verify content and owner-only permissions
		content, err := os.ReadFile(backups[0])
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
		info, err := os.Stat(backups[0])
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("keeps only the newest backups", func(t *testing.T) {
		testFile := filepath.Join(t.TempDir(), "credentials")
		for i := 1; i <= awsBackupsKept+2; i++ {
			require.NoError(t, os.WriteFile(testFile, []byte(fmt.Sprintf("v%d", i)), 0o600))
			require.NoError(t, backupFile(testFile))
		}

		backups := listBackups(testFile)
		require.Len(t, backups, awsBackupsKept)
		for i, backup := range backups {
			content, err := os.ReadFile(backup) //nolint:gosec // path constructed from test temp dir
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("v%d", awsBackupsKept+2-i), string(content), "backups are listed newest first")
		}
	})

	t.Run("never removes files it didn't create", func(t *testing.T) {
		dir := t.TempDir()
		testFile := filepath.Join(dir, "credentials")
		legacy := testFile + awsBackupSuffix
		handMade := testFile + awsBackupSuffix + ".old"
		for _, path := range []string{testFile, legacy, handMade} {
			require.NoError(t, os.WriteFile(path, []byte("keep"), 0o600))
		}

		for i := 0; i < awsBackupsKept+2; i++ {
			require.NoError(t, backupFile(testFile))
		}

		assert.Len(t, listBackups(testFile), awsBackupsKept)
		assert.FileExists(t, legacy)
		assert.FileExists(t, handMade)
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

	t.Run("long-term keys drop a leftover session token", func(t *testing.T) {
		tmpDir := t.TempDir()
		credPath := filepath.Join(tmpDir, "credentials")

		// Session credentials an earlier refresh wrote over the long-term keys
		err := writeAWSCredentials(credPath, "dev-base", "ASIATEMP", "SECRETTEMP", "TOKTEMP")
		require.NoError(t, err)

		err = writeAWSCredentials(credPath, "dev-base", "AKIALONG", "SECRETLONG", "")
		require.NoError(t, err)

		content, err := os.ReadFile(credPath)
		require.NoError(t, err)
		assert.Equal(t, "[dev-base]\naws_access_key_id = AKIALONG\naws_secret_access_key = SECRETLONG\n", string(content))
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
		backups := listBackups(credPath)
		require.NotEmpty(t, backups)
		backupContent, err := os.ReadFile(backups[0])
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

// TestGetBaseProfile tests resolving a session profile to its base profile
func TestGetBaseProfile(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{
			name:   "magex_base_profile link",
			config: "[profile dev]\nmagex_base_profile = team-base\n",
			want:   "team-base",
		},
		{
			name:   "magex_base_profile wins over source_profile",
			config: "[profile dev]\nsource_profile = old-base\nmagex_base_profile = team-base\n",
			want:   "team-base",
		},
		{
			name:   "legacy source_profile without role_arn",
			config: "[profile dev]\nsource_profile = team-base\nregion = us-east-1\n",
			want:   "team-base",
		},
		{
			name:   "source_profile with role_arn is an assume-role profile, not a link",
			config: "[profile dev]\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = team-base\n",
		},
		{
			name:   "no link falls back to <profile>-base when it has an MFA serial",
			config: "[profile dev]\nregion = us-east-1\n\n[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n",
			want:   "dev-base",
		},
		{
			name:   "no link and <profile>-base has no MFA serial",
			config: "[profile dev]\nregion = us-east-1\n\n[profile dev-base]\nregion = us-east-1\n",
		},
		{
			name: "config file missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedHome(t, "", tt.config)
			assert.Equal(t, tt.want, getBaseProfile("dev"))
		})
	}
}

// TestSetBaseProfileLink tests the line-level config edit that links a session profile
func TestSetBaseProfileLink(t *testing.T) {
	tests := []struct {
		name    string
		content string
		section string
		want    string
	}{
		{
			name:    "empty file",
			section: "profile dev",
			want:    "[profile dev]\nmagex_base_profile = dev-base\n",
		},
		{
			name:    "appends a missing section",
			content: "[default]\nregion = us-east-1\n",
			section: "profile dev",
			want:    "[default]\nregion = us-east-1\n\n[profile dev]\nmagex_base_profile = dev-base\n",
		},
		{
			name:    "appends after a file without a trailing newline",
			content: "[default]\nregion = us-east-1",
			section: "profile dev",
			want:    "[default]\nregion = us-east-1\n\n[profile dev]\nmagex_base_profile = dev-base\n",
		},
		{
			name: "replaces a legacy source_profile and keeps every other line",
			content: "# work account\n[profile dev]\nregion = us-east-1\nsource_profile = dev-base\n" +
				"s3 =\n  max_concurrent_requests = 20\n\n[profile other]\nsource_profile = dev-base\n",
			section: "profile dev",
			want: "# work account\n[profile dev]\nregion = us-east-1\nmagex_base_profile = dev-base\n" +
				"s3 =\n  max_concurrent_requests = 20\n\n[profile other]\nsource_profile = dev-base\n",
		},
		{
			name:    "keeps source_profile on an assume-role profile",
			content: "[profile dev]\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = dev-base\n",
			section: "profile dev",
			want:    "[profile dev]\nmagex_base_profile = dev-base\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = dev-base\n",
		},
		{
			name:    "updates an existing link and drops a leftover source_profile",
			content: "[profile dev]\nsource_profile = old-base\nmagex_base_profile = old-base\n",
			section: "profile dev",
			want:    "[profile dev]\nmagex_base_profile = dev-base\n",
		},
		{
			name:    "adds the link under an existing header",
			content: "[profile dev]\nregion = us-east-1\n",
			section: "profile dev",
			want:    "[profile dev]\nmagex_base_profile = dev-base\nregion = us-east-1\n",
		},
		{
			name:    "default profile section",
			content: "[default]\nsource_profile = dev-base\n",
			section: "default",
			want:    "[default]\nmagex_base_profile = dev-base\n",
		},
		{
			name:    "keeps CRLF line endings",
			content: "[profile dev]\r\nsource_profile = dev-base\r\nregion = us-east-1\r\n",
			section: "profile dev",
			want:    "[profile dev]\r\nmagex_base_profile = dev-base\r\nregion = us-east-1\r\n",
		},
		{
			name:    "ignores an indented key with the same name",
			content: "[profile dev]\nservices =\n  source_profile = x\n",
			section: "profile dev",
			want:    "[profile dev]\nmagex_base_profile = dev-base\nservices =\n  source_profile = x\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, setBaseProfileLink(tt.content, tt.section, "dev-base"))
		})
	}
}

// TestWriteAWSConfigBaseProfile tests writing the base profile link to the config file
func TestWriteAWSConfigBaseProfile(t *testing.T) {
	t.Run("creates a missing file with owner-only permissions", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config")

		require.NoError(t, writeAWSConfigBaseProfile(configPath, "dev", "dev-base"))

		content, err := os.ReadFile(configPath) //nolint:gosec // path constructed from test temp dir
		require.NoError(t, err)
		assert.Equal(t, "[profile dev]\nmagex_base_profile = dev-base\n", string(content))
		info, err := os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("updates an existing file", func(t *testing.T) {
		configPath := filepath.Join(t.TempDir(), "config")
		require.NoError(t, os.WriteFile(configPath, []byte("[profile dev]\nsource_profile = dev-base\nregion = us-east-1\n"), 0o600))

		require.NoError(t, writeAWSConfigBaseProfile(configPath, "dev", "dev-base"))

		content, err := os.ReadFile(configPath) //nolint:gosec // path constructed from test temp dir
		require.NoError(t, err)
		assert.Equal(t, "[profile dev]\nmagex_base_profile = dev-base\nregion = us-east-1\n", string(content))
	})

	t.Run("fails when the config path can't be read", func(t *testing.T) {
		require.Error(t, writeAWSConfigBaseProfile(t.TempDir(), "dev", "dev-base"))
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
			credsFile:  "[dev-base]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile dev]\nsource_profile = dev-base\n",
			profile:    "dev",
			want:       false,
		},
		{
			name:       "post-setup pre-refresh (primary bug-fix case)",
			makeAWSDir: true,
			credsFile: "[dev-base]\n" +
				"aws_access_key_id = AKIATEST\n" +
				"aws_secret_access_key = SECRETTEST\n",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile dev]\nsource_profile = dev-base\n",
			profile: "dev",
			want:    true,
		},
		{
			name:       "post-refresh has both base and session sections",
			makeAWSDir: true,
			credsFile: "[dev-base]\naws_access_key_id = AKIATEST\n\n" +
				"[dev]\naws_access_key_id = ASIASESSION\naws_session_token = TOK\n",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile dev]\nsource_profile = dev-base\n",
			profile: "dev",
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
			config: "[profile dev]\nsource_profile = dev-base\n\n" +
				"[profile other-base]\nmfa_serial = " + baseProfileMFA + "\n",
			profile: "dev",
			want:    false,
		},
		{
			name:       "user passes base profile name directly",
			makeAWSDir: true,
			credsFile:  "[dev-base]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "dev-base",
			want:       true,
		},
		{
			name:       "session linked with magex_base_profile",
			makeAWSDir: true,
			credsFile:  "[dev-base]\n" + "aws_access_key_id = AKIATEST\n",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile dev]\nmagex_base_profile = dev-base\n",
			profile: "dev",
			want:    true,
		},
		{
			name:       "session without a link pairs with <profile>-base",
			makeAWSDir: true,
			credsFile:  "[dev-base]\n" + "aws_access_key_id = AKIATEST\n",
			config:     "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:    "dev",
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
	const credsBothSections = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n\n" +
		"[dev]\n" +
		"aws_access_key_id = ASIASESSION\n" +
		"aws_secret_access_key = SECRETSESSION\n" +
		"aws_session_token = TOKSESSION\n"

	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n"

	const configBoth = "[profile dev-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile dev]\n" +
		"source_profile = dev-base\n"

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

		err := AWS{}.Status("profile=dev")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "dev-base", "Status should display base profile when filtered by session name")
	})

	t.Run("filter by session profile post-refresh shows both sections", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=dev")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "dev-base")
		assert.Contains(t, profiles, "dev")
	})

	t.Run("filter by base profile directly shows only base", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status("profile=dev-base")
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.Contains(t, profiles, "dev-base")
		assert.NotContains(t, profiles, "dev", "filtering by base name should not double-print session section")
	})

	t.Run("no filter displays every section", func(t *testing.T) {
		withMockedHome(t, credsBothSections, configBoth)
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			return "", errTestSTSUnavailable
		})

		err := AWS{}.Status()
		require.NoError(t, err)

		profiles := extractProfiles(mock.allCalls)
		assert.ElementsMatch(t, []string{"dev-base", "dev"}, profiles)
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

		err := AWS{}.Status("profile=dev")
		require.NoError(t, err)
		assert.Empty(t, mock.allCalls)
	})
}

// ============================================================================
// Login — end-to-end flow tests
// ============================================================================

func TestLoginFlow(t *testing.T) {
	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"

	const credsBothSections = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n\n" +
		"[dev]\n" +
		"aws_access_key_id = ASIASTALE\n" +
		"aws_secret_access_key = SECRETSTALE\n" +
		"aws_session_token = TOKSTALE\n"

	const configBoth = "[profile dev-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile dev]\n" +
		"source_profile = dev-base\n"

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

		err := AWS{}.Login("profile=dev")
		require.NoError(t, err)

		// The STS call must target the BASE profile (long-term keys).
		require.NotEmpty(t, mock.allCalls)
		stsArgs := mock.allCalls[len(mock.allCalls)-1]
		assert.True(t, argsContain(stsArgs, "--profile", "dev-base"),
			"Refresh path must invoke STS with --profile dev-base; got %v", stsArgs)
		assert.True(t, argsContain(stsArgs, "--token-code", "123456"))

		// Credentials file should now contain the session section with the new token.
		ini := readCredentialsINI(t)
		session := findSection(ini, "dev")
		require.NotNil(t, session, "session profile [dev] must be written after Refresh")
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

		err := AWS{}.Login("profile=dev")
		require.NoError(t, err)

		assert.Empty(t, mock.allCalls, "Setup path must not invoke STS")

		creds := readCredentialsINI(t)
		base := findSection(creds, "dev-base")
		require.NotNil(t, base, "[dev-base] must be written to credentials")
		assert.Equal(t, "AKIATEST", base.Values["aws_access_key_id"])
		assert.Equal(t, "SECRETTEST", base.Values["aws_secret_access_key"])
		// Session profile section must NOT be written by Setup.
		assert.Nil(t, findSection(creds, "dev"))

		config := readConfigINI(t)
		baseCfg := findSection(config, "profile dev-base")
		require.NotNil(t, baseCfg, "[profile dev-base] must be written to config")
		assert.Equal(t, "arn:aws:iam::1:mfa/u", baseCfg.Values["mfa_serial"])

		sessionCfg := findSection(config, "profile dev")
		require.NotNil(t, sessionCfg, "[profile dev] must be written to config")
		assert.Equal(t, "dev-base", sessionCfg.Values[awsBaseProfileKey])
		assert.NotContains(t, sessionCfg.Values, "source_profile", "the AWS SDKs for Go reject source_profile without role_arn")
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

		err := AWS{}.Login("profile=dev")
		require.NoError(t, err)
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "dev")
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

	require.NoError(t, AWS{}.Setup("profile=dev"))

	creds := readCredentialsINI(t)
	base := findSection(creds, "dev-base")
	require.NotNil(t, base)
	assert.Equal(t, "AKIATEST", base.Values["aws_access_key_id"])
	assert.Equal(t, "SECRETTEST", base.Values["aws_secret_access_key"])
	assert.Nil(t, findSection(creds, "dev"), "Setup must not write the session profile to credentials")

	config := readConfigINI(t)
	baseCfg := findSection(config, "profile dev-base")
	require.NotNil(t, baseCfg)
	assert.Equal(t, "arn:aws:iam::123456789012:mfa/test", baseCfg.Values["mfa_serial"])

	sessionCfg := findSection(config, "profile dev")
	require.NotNil(t, sessionCfg)
	assert.Equal(t, "dev-base", sessionCfg.Values[awsBaseProfileKey])
	assert.NotContains(t, sessionCfg.Values, "source_profile", "the AWS SDKs for Go reject source_profile without role_arn")

	// Re-running Setup creates .bak files for both credentials and config.
	withMockedPrompts(t, []string{
		"AKIATWO",
		"SECRETTWO",
		"arn:aws:iam::123456789012:mfa/test",
	})
	require.NoError(t, AWS{}.Setup("profile=dev"))

	awsDir, err := getAWSDir()
	require.NoError(t, err)
	assert.NotEmpty(t, listBackups(filepath.Join(awsDir, awsCredentialsFile)))
	assert.NotEmpty(t, listBackups(filepath.Join(awsDir, awsConfigFile)))

	creds = readCredentialsINI(t)
	base = findSection(creds, "dev-base")
	require.NotNil(t, base)
	assert.Equal(t, "AKIATWO", base.Values["aws_access_key_id"])
}

// ============================================================================
// Refresh — exercises MFA prompt, STS call, and session-credential write
// ============================================================================

func TestRefreshFlow(t *testing.T) {
	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const configBoth = "[profile dev-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile dev]\n" +
		"source_profile = dev-base\n"

	t.Run("writes session creds under session profile", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, configBoth)
		withMockedPrompts(t, []string{"123456"})

		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				assert.True(t, argsContain(args, "--serial-number", baseProfileMFA))
				assert.True(t, argsContain(args, "--token-code", "123456"))
				assert.True(t, argsContain(args, "--profile", "dev-base"))
				return stsSessionTokenJSON("ASIASESSION", "SECSESSION", "TOKSESSION"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		require.NoError(t, AWS{}.Refresh("profile=dev"))
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "dev")
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

		err := AWS{}.Refresh("profile=dev")
		require.Error(t, err)
		require.ErrorIs(t, err, errInvalidMFAToken, "expected errInvalidMFAToken, got: %v", err)
		assert.Empty(t, mock.allCalls, "runner must not be invoked when validation fails")
	})

	t.Run("explicit base= overrides source_profile lookup", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(
			t,
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

		require.NoError(t, AWS{}.Refresh("profile=dev", "base=custom-base"))
		require.NotEmpty(t, mock.allCalls)

		ini := readCredentialsINI(t)
		session := findSection(ini, "dev")
		require.NotNil(t, session)
		assert.Equal(t, "TOKNEW", session.Values["aws_session_token"])
	})
}

// ============================================================================
// Refresh — base profiles that can't start a new session, and never
// overwriting long-term keys
// ============================================================================

var (
	errTestCallerIdentityFailed = errors.New("test: sts get-caller-identity failed")
	// errTestSTSRejectsSessionCreds is the error the runner returns when the base
	// profile's credentials are temporary (captured from a real AWS CLI run)
	errTestSTSRejectsSessionCreds = errors.New("command failed [aws sts get-session-token]: exit status 254\n" +
		"aws: [ERROR]: An error occurred (AccessDenied) when calling the GetSessionToken operation: " +
		"Cannot call GetSessionToken with session credentials")
)

// callerIdentityJSON is a canned sts get-caller-identity response.
const callerIdentityJSON = `{"UserId": "AIDATEST", "Account": "123456789012", "Arn": "arn:aws:iam::123456789012:user/test"}`

// captureAWSLog redirects the shared CLI logger into a buffer while fn runs and
// returns what it wrote.
func captureAWSLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	logger := pkglog.Default()
	logger.SetOutput(&buf)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })
	fn()
	return buf.String()
}

// awsTestPath returns the path of a file in ~/.aws under the mocked HOME.
func awsTestPath(t *testing.T, name string) string {
	t.Helper()
	awsDir, err := getAWSDir()
	require.NoError(t, err)
	return filepath.Join(awsDir, name)
}

// assertAWSFileUntouched fails when ~/.aws/<name> differs from want or was backed up.
func assertAWSFileUntouched(t *testing.T, name, want string) {
	t.Helper()
	data, err := os.ReadFile(awsTestPath(t, name)) //nolint:gosec // path constructed from test temp dir
	require.NoError(t, err)
	assert.Equal(t, want, string(data), "%s must not be rewritten", name)
	backups, err := filepath.Glob(awsTestPath(t, name+awsBackupSuffix+"*"))
	require.NoError(t, err)
	assert.Empty(t, backups, "%s must not be backed up", name)
}

func TestRefreshBaseHoldsSessionCreds(t *testing.T) {
	// What an older refresh left behind after writing a session over the base
	// profile's long-term keys
	const credsSessionInBase = "[dev-base]\n" +
		"aws_access_key_id = ASIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n" +
		"aws_session_token = TOKBASE\n"
	// The same credentials after deleting the token line by hand, which leaves
	// a temporary key that can't be used at all
	const credsTokenDeleted = "[dev-base]\n" +
		"aws_access_key_id = ASIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const configBoth = "[profile dev-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile dev]\n" +
		"magex_base_profile = dev-base\n"

	tests := []struct {
		name      string
		credsFile string
		profile   string
		loggedIn  bool
		wantErr   error
	}{
		{name: "session profile still logged in", credsFile: credsSessionInBase, profile: "dev", loggedIn: true},
		{name: "session profile logged out", credsFile: credsSessionInBase, profile: "dev", wantErr: errBaseHasSessionCreds},
		{name: "base profile refreshed directly while logged in", credsFile: credsSessionInBase, profile: "dev-base", loggedIn: true},
		{name: "base profile refreshed directly after it expired", credsFile: credsSessionInBase, profile: "dev-base", wantErr: errBaseHasSessionCreds},
		{name: "session token line deleted by hand", credsFile: credsTokenDeleted, profile: "dev", wantErr: errBaseHasSessionCreds},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipIfNoAWSCLI(t)

			withMockedHome(t, tt.credsFile, configBoth)
			withMockedPrompts(t, nil) // an MFA prompt would hit EOF
			mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
				if tt.loggedIn && argsContain(args, "sts", "get-caller-identity") {
					return callerIdentityJSON, nil
				}
				return "", errTestCallerIdentityFailed
			})

			var err error
			output := captureAWSLog(t, func() { err = AWS{}.Refresh("profile=" + tt.profile) })

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotContains(t, output, "Already logged in")
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, "Already logged in: profile '"+tt.profile+"' has an active session (Account: 123456789012)")
			}
			assert.Contains(t, output, "Profile 'dev-base' holds temporary session credentials")
			assert.Contains(t, output, "Re-enter your long-term keys with: magex aws:setup profile=dev")

			// Only the session check reaches AWS: no MFA code, no STS token call, no write
			require.Len(t, mock.allCalls, 1)
			assert.True(t, argsContain(mock.allCalls[0], "get-caller-identity"), "got %v", mock.allCalls[0])
			assert.True(t, argsContain(mock.allCalls[0], "--profile", tt.profile), "got %v", mock.allCalls[0])
			assertAWSFileUntouched(t, awsCredentialsFile, tt.credsFile)
			assertAWSFileUntouched(t, awsConfigFile, configBoth)
		})
	}

	const credsLongTerm = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const olderBackup = awsCredentialsFile + awsBackupSuffix + ".20260101T000000.000Z"
	const newerBackup = awsCredentialsFile + awsBackupSuffix + ".20260201T000000.000Z"
	const legacyBackup = awsCredentialsFile + awsBackupSuffix

	backups := []struct {
		name  string
		files map[string]string // backup file name in ~/.aws -> content
		want  string            // backup the hint should name, "" for none
	}{
		{
			name:  "points at a .bak from older versions that has the long-term keys",
			files: map[string]string{legacyBackup: credsLongTerm},
			want:  legacyBackup,
		},
		{
			name:  "ignores a backup that only has temporary keys",
			files: map[string]string{legacyBackup: credsTokenDeleted},
		},
		{
			name:  "skips newer backups that only have temporary keys",
			files: map[string]string{newerBackup: credsSessionInBase, olderBackup: credsLongTerm},
			want:  olderBackup,
		},
		{
			name:  "prefers the newest timestamped backup",
			files: map[string]string{newerBackup: credsLongTerm, olderBackup: credsLongTerm, legacyBackup: credsLongTerm},
			want:  newerBackup,
		},
	}

	for _, tt := range backups {
		t.Run(tt.name, func(t *testing.T) {
			skipIfNoAWSCLI(t)

			withMockedHome(t, credsSessionInBase, configBoth)
			for name, content := range tt.files {
				require.NoError(t, os.WriteFile(awsTestPath(t, name), []byte(content), 0o600))
			}
			withMockedPrompts(t, nil)
			withMockedRunner(t, func(cmd string, args ...string) (string, error) {
				return "", errTestCallerIdentityFailed
			})

			var err error
			output := captureAWSLog(t, func() { err = AWS{}.Refresh("profile=dev") })

			require.ErrorIs(t, err, errBaseHasSessionCreds)
			if tt.want != "" {
				assert.Contains(t, output, awsTestPath(t, tt.want)+" still has the long-term keys for 'dev-base'")
				assert.Contains(t, output, "Replace both aws_access_key_id and aws_secret_access_key under [dev-base]")
				assert.Contains(t, output, "Or re-enter your keys with: magex aws:setup profile=dev")
			} else {
				assert.NotContains(t, output, "still has the long-term keys")
				assert.Contains(t, output, "Re-enter your long-term keys with: magex aws:setup profile=dev")
			}
		})
	}
}

func TestRefreshSTSRejectsSessionCreds(t *testing.T) {
	// The credentials file holds long-term keys, but the AWS CLI resolved
	// temporary credentials from somewhere else
	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const configBoth = "[profile dev-base]\n" +
		"mfa_serial = " + baseProfileMFA + "\n\n" +
		"[profile dev]\n" +
		"magex_base_profile = dev-base\n"

	tests := []struct {
		name     string
		loggedIn bool
		wantErr  error
	}{
		{name: "session profile still logged in", loggedIn: true},
		{name: "session profile logged out", wantErr: errBaseHasSessionCreds},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skipIfNoAWSCLI(t)

			withMockedHome(t, credsBaseOnly, configBoth)
			withMockedPrompts(t, []string{"123456"})
			withMockedRunner(t, func(cmd string, args ...string) (string, error) {
				switch {
				case argsContain(args, "sts", "get-session-token"):
					return "", errTestSTSRejectsSessionCreds
				case tt.loggedIn && argsContain(args, "sts", "get-caller-identity"):
					return callerIdentityJSON, nil
				default:
					return "", errTestCallerIdentityFailed
				}
			})

			var err error
			output := captureAWSLog(t, func() { err = AWS{}.Refresh("profile=dev") })

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotContains(t, err.Error(), "command failed", "the raw AWS CLI error must not reach the user")
			} else {
				require.NoError(t, err)
				assert.Contains(t, output, "Already logged in: profile 'dev'")
			}
			assert.Contains(t, output, "The AWS CLI resolved temporary session credentials for 'dev-base'")
			assert.Contains(t, output, "aws configure list --profile dev-base")
			assertAWSFileUntouched(t, awsCredentialsFile, credsBaseOnly)
			assertAWSFileUntouched(t, awsConfigFile, configBoth)
		})
	}
}

func TestRefreshBaseProfileWithLongTermKeys(t *testing.T) {
	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const credsProd = "[prod]\n" +
		"aws_access_key_id = AKIAPROD\n" +
		"aws_secret_access_key = SECRETPROD\n"

	t.Run("writes the session to the linked session profile", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, "[profile dev-base]\n"+
			"mfa_serial = "+baseProfileMFA+"\n\n"+
			"[profile dev]\n"+
			"magex_base_profile = dev-base\n\n"+
			"[profile dev-admin]\n"+
			"role_arn = arn:aws:iam::123456789012:role/admin\n"+
			"source_profile = dev-base\n")
		withMockedPrompts(t, []string{"123456"})
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				return stsSessionTokenJSON("ASIASESSION", "SECSESSION", "TOKSESSION"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		require.NoError(t, AWS{}.Refresh("profile=dev-base"))

		require.Len(t, mock.allCalls, 1)
		assert.True(t, argsContain(mock.allCalls[0], "--profile", "dev-base"),
			"STS must still be called with the base profile's long-term keys; got %v", mock.allCalls[0])

		ini := readCredentialsINI(t)
		base := findSection(ini, "dev-base")
		require.NotNil(t, base)
		assert.Equal(t, "AKIABASE", base.Values["aws_access_key_id"], "long-term keys must survive the refresh")
		assert.NotContains(t, base.Values, "aws_session_token")

		session := findSection(ini, "dev")
		require.NotNil(t, session, "the session belongs in the linked session profile")
		assert.Equal(t, "ASIASESSION", session.Values["aws_access_key_id"])
		assert.Equal(t, "TOKSESSION", session.Values["aws_session_token"])
		assert.Nil(t, findSection(ini, "dev-admin"), "assume-role profiles are not session targets")
	})

	t.Run("pairs the base with <name> when the session profile has no link", func(t *testing.T) {
		skipIfNoAWSCLI(t)

		withMockedHome(t, credsBaseOnly, "[profile dev-base]\n"+
			"mfa_serial = "+baseProfileMFA+"\n\n"+
			"[profile dev]\n"+
			"region = us-east-1\n")
		withMockedPrompts(t, []string{"123456"})
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				return stsSessionTokenJSON("ASIASESSION", "SECSESSION", "TOKSESSION"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		require.NoError(t, AWS{}.Refresh("profile=dev-base"))

		require.Len(t, mock.allCalls, 1)
		assert.True(t, argsContain(mock.allCalls[0], "--profile", "dev-base"), "got %v", mock.allCalls[0])
		ini := readCredentialsINI(t)
		assert.Equal(t, "AKIABASE", findSection(ini, "dev-base").Values["aws_access_key_id"])
		session := findSection(ini, "dev")
		require.NotNil(t, session)
		assert.Equal(t, "TOKSESSION", session.Values["aws_session_token"])
	})

	refusals := []struct {
		name      string
		credsFile string
		config    string
		profile   string
	}{
		{
			name:      "no linked session profile",
			credsFile: credsProd,
			config:    "[profile prod]\nmfa_serial = " + baseProfileMFA + "\n",
			profile:   "prod",
		},
		{
			name:      "profile linked to itself",
			credsFile: credsProd,
			config:    "[profile prod]\nmfa_serial = " + baseProfileMFA + "\nsource_profile = prod\n",
			profile:   "prod",
		},
		{
			name:      "several linked session profiles",
			credsFile: credsBaseOnly,
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n\n" +
				"[profile dev]\nmagex_base_profile = dev-base\n\n" +
				"[profile dev-rw]\nsource_profile = dev-base\n",
			profile: "dev-base",
		},
	}

	for _, tt := range refusals {
		t.Run("refuses with "+tt.name, func(t *testing.T) {
			skipIfNoAWSCLI(t)

			withMockedHome(t, tt.credsFile, tt.config)
			withMockedPrompts(t, nil)
			mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
				return "", errTestUnexpectedRunnerCall
			})

			err := AWS{}.Refresh("profile=" + tt.profile)

			require.ErrorIs(t, err, errWouldOverwriteLongTermKeys)
			assert.Empty(t, mock.allCalls, "nothing may reach AWS before the refusal")
			assertAWSFileUntouched(t, awsCredentialsFile, tt.credsFile)
		})
	}
}

func TestRefreshBaseProfileLink(t *testing.T) {
	const credsBaseOnly = "[dev-base]\n" +
		"aws_access_key_id = AKIABASE\n" +
		"aws_secret_access_key = SECRETBASE\n"
	const baseConfig = "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n"

	// refresh runs a successful MFA refresh and returns the log output and the STS call
	refresh := func(t *testing.T, args ...string) (string, []string) {
		t.Helper()
		withMockedPrompts(t, []string{"123456"})
		mock := withMockedRunner(t, func(cmd string, args ...string) (string, error) {
			if argsContain(args, "sts", "get-session-token") {
				return stsSessionTokenJSON("ASIASESSION", "SECSESSION", "TOKSESSION"), nil
			}
			return "", errTestUnexpectedRunnerCall
		})

		var err error
		output := captureAWSLog(t, func() { err = AWS{}.Refresh(args...) })
		require.NoError(t, err)
		require.Len(t, mock.allCalls, 1)
		return output, mock.allCalls[0]
	}

	t.Run("replaces a legacy source_profile link and keeps the rest of the file", func(t *testing.T) {
		skipIfNoAWSCLI(t)
		withMockedHome(t, credsBaseOnly, baseConfig+"\n# edited by hand\n[profile dev]\nsource_profile = dev-base\nregion = us-east-1\n")

		output, stsCall := refresh(t, "profile=dev")

		assert.True(t, argsContain(stsCall, "--profile", "dev-base"), "got %v", stsCall)
		content, err := os.ReadFile(awsTestPath(t, awsConfigFile)) //nolint:gosec // path constructed from test temp dir
		require.NoError(t, err)
		assert.Equal(t, baseConfig+"\n# edited by hand\n[profile dev]\nmagex_base_profile = dev-base\nregion = us-east-1\n", string(content))
		backups, err := filepath.Glob(awsTestPath(t, awsConfigFile+awsBackupSuffix+"*"))
		require.NoError(t, err)
		assert.NotEmpty(t, backups, "config must be backed up before it's edited")
		assert.Contains(t, output, "Linked 'dev' to 'dev-base' with magex_base_profile instead of source_profile")
	})

	t.Run("saves an explicit base= when the profile has no link", func(t *testing.T) {
		skipIfNoAWSCLI(t)
		withMockedHome(t, "[team-base]\naws_access_key_id = AKIATEAM\naws_secret_access_key = SECRETTEAM\n",
			"[profile team-base]\nmfa_serial = "+baseProfileMFA+"\n")

		output, stsCall := refresh(t, "profile=dev", "base=team-base")

		assert.True(t, argsContain(stsCall, "--profile", "team-base"), "got %v", stsCall)
		assert.Equal(t, "team-base", getBaseProfile("dev"), "later refreshes must find the base without base=")
		assert.Contains(t, output, "Saved 'team-base' as the base profile for 'dev'")
	})

	unchanged := []struct {
		name   string
		config string
		args   []string
	}{
		{
			name:   "already linked with magex_base_profile",
			config: baseConfig + "\n[profile dev]\nmagex_base_profile = dev-base\n",
			args:   []string{"profile=dev"},
		},
		{
			name:   "base found by name without a link",
			config: baseConfig,
			args:   []string{"profile=dev"},
		},
		{
			name:   "explicit base= that matches the name aws:setup uses",
			config: baseConfig,
			args:   []string{"profile=dev", "base=dev-base"},
		},
		{
			name: "explicit base= while a link exists",
			config: baseConfig + "\n[profile dev]\nmagex_base_profile = dev-base\n\n" +
				"[profile other-base]\nmfa_serial = " + baseProfileMFA + "\n",
			args: []string{"profile=dev", "base=other-base"},
		},
	}

	for _, tt := range unchanged {
		t.Run("leaves config alone: "+tt.name, func(t *testing.T) {
			skipIfNoAWSCLI(t)
			withMockedHome(t, credsBaseOnly, tt.config)

			refresh(t, tt.args...)

			assertAWSFileUntouched(t, awsConfigFile, tt.config)
		})
	}
}

func TestStatusWarnsAboutLegacyLink(t *testing.T) {
	const creds = "[dev]\naws_access_key_id = ASIASESSION\naws_secret_access_key = SECSESSION\naws_session_token = TOK\n"

	tests := []struct {
		name   string
		config string
		warn   bool
	}{
		{name: "source_profile without role_arn", config: "[profile dev]\nsource_profile = dev-base\n", warn: true},
		{name: "magex_base_profile", config: "[profile dev]\nmagex_base_profile = dev-base\n"},
		{name: "assume-role profile", config: "[profile dev]\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = dev-base\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockedHome(t, creds, tt.config)
			withMockedRunner(t, func(cmd string, args ...string) (string, error) {
				return "", errTestSTSUnavailable
			})

			var err error
			output := captureAWSLog(t, func() { err = AWS{}.Status("profile=dev") })

			require.NoError(t, err)
			if tt.warn {
				assert.Contains(t, output, "source_profile without role_arn")
				assert.Contains(t, output, "magex aws:refresh profile=dev")
			} else {
				assert.NotContains(t, output, "source_profile without role_arn")
			}
		})
	}
}

func TestLinkedSessionProfiles(t *testing.T) {
	t.Run("only plain profiles linked to the base", func(t *testing.T) {
		withMockedHome(t, "", "[default]\nsource_profile = shared-base\n\n"+
			"[profile a]\nsource_profile = shared-base\n\n"+
			"[profile b]\nsource_profile = other-base\n\n"+
			"[profile c]\nmagex_base_profile = shared-base\n\n"+
			"[profile d]\nmagex_base_profile = other-base\nsource_profile = shared-base\n\n"+
			"[profile admin]\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = shared-base\n\n"+
			"[sso-session corp]\nsource_profile = shared-base\n\n"+
			"[profile shared-base]\nmfa_serial = "+baseProfileMFA+"\nsource_profile = shared-base\n")

		assert.Equal(t, []string{"default", "a", "c"}, linkedSessionProfiles("shared-base"))
		assert.Empty(t, linkedSessionProfiles("unknown-base"))
	})

	pairing := []struct {
		name   string
		creds  string
		config string
		want   []string
	}{
		{
			name:   "pairs <name> with <name>-base when it has no link",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n\n[profile dev]\nregion = us-east-1\n",
			want:   []string{"dev"},
		},
		{
			name:   "pairs a session profile found only in the credentials file",
			creds:  "[dev]\naws_access_key_id = ASIASESSION\n",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n",
			want:   []string{"dev"},
		},
		{
			name:   "no pairing when <name> links to another base",
			config: "[profile dev]\nmagex_base_profile = other-base\n",
		},
		{
			name:   "no pairing with an assume-role profile",
			config: "[profile dev]\nrole_arn = arn:aws:iam::123456789012:role/admin\nsource_profile = x\n",
		},
		{
			name:   "no pairing when <name> doesn't exist",
			config: "[profile dev-base]\nmfa_serial = " + baseProfileMFA + "\n",
		},
	}

	for _, tt := range pairing {
		t.Run(tt.name, func(t *testing.T) {
			withMockedHome(t, tt.creds, tt.config)
			assert.Equal(t, tt.want, linkedSessionProfiles("dev-base"))
		})
	}

	t.Run("missing config file", func(t *testing.T) {
		withMockedHome(t, "", "")
		assert.Empty(t, linkedSessionProfiles("shared-base"))
	})
}

func TestCredentialSectionKinds(t *testing.T) {
	tests := []struct {
		name          string
		section       *awsINISection
		wantTemporary bool
		wantLongTerm  bool
	}{
		{name: "missing section"},
		{
			name:         "long-term keys",
			section:      &awsINISection{Values: map[string]string{"aws_access_key_id": "AKIA", "aws_secret_access_key": "S"}},
			wantLongTerm: true,
		},
		{
			name:          "session credentials",
			section:       &awsINISection{Values: map[string]string{"aws_access_key_id": "ASIA", "aws_session_token": "T"}},
			wantTemporary: true,
		},
		{
			name:          "temporary key with its token line deleted",
			section:       &awsINISection{Values: map[string]string{"aws_access_key_id": "ASIAX", "aws_secret_access_key": "S"}},
			wantTemporary: true,
		},
		{
			name:         "empty session token is ignored",
			section:      &awsINISection{Values: map[string]string{"aws_access_key_id": "AKIA", "aws_session_token": ""}},
			wantLongTerm: true,
		},
		{
			name:    "no access key",
			section: &awsINISection{Values: map[string]string{"region": "us-east-1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTemporary, hasTemporaryCreds(tt.section))
			assert.Equal(t, tt.wantLongTerm, hasLongTermKeys(tt.section))
		})
	}
}

func TestLoadAWSINISection(t *testing.T) {
	credPath := filepath.Join(t.TempDir(), awsCredentialsFile)
	require.NoError(t, os.WriteFile(credPath, []byte("[a]\naws_access_key_id = AKIAA\n\n[b]\naws_access_key_id = AKIAB\n"), 0o600))

	section := loadAWSINISection(credPath, "b")
	require.NotNil(t, section)
	assert.Equal(t, "AKIAB", section.Values["aws_access_key_id"])
	assert.Nil(t, loadAWSINISection(credPath, "missing"))
	assert.Nil(t, loadAWSINISection(credPath+".missing", "a"))
}

func TestSetupProfileName(t *testing.T) {
	tests := []struct {
		profile     string
		baseProfile string
		want        string
	}{
		{profile: "dev", baseProfile: "dev-base", want: "dev"},
		{profile: "dev-base", baseProfile: "dev-base", want: "dev"},
		{profile: "prod", baseProfile: "prod", want: "prod"},
		{profile: "dev", baseProfile: "custom", want: "dev"},
		{profile: "odd", baseProfile: "-base", want: "odd"},
	}

	for _, tt := range tests {
		t.Run(tt.profile+"/"+tt.baseProfile, func(t *testing.T) {
			assert.Equal(t, tt.want, setupProfileName(tt.profile, tt.baseProfile))
		})
	}
}

// ============================================================================
// Edge cases
// ============================================================================

func TestEdgeCases(t *testing.T) {
	t.Run("source_profile is resolved only one level deep", func(t *testing.T) {
		// a -> b -> c (creds + MFA only on c). getSourceProfile is non-recursive,
		// so hasValidAWSSetup("a") looks for credentials on "b" and finds none.
		withMockedHome(
			t,
			"[c]\naws_access_key_id = AKIATEST\n",
			"[profile a]\nsource_profile = b\n\n"+
				"[profile b]\nsource_profile = c\n\n"+
				"[profile c]\nmfa_serial = "+baseProfileMFA+"\n",
		)
		assert.False(t, hasValidAWSSetup("a"))
	})
}
