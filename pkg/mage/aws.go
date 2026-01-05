// Package mage provides AWS credential management commands
package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// AWS credential management constants
const (
	awsDefaultDuration = 43200 // 12 hours in seconds
	awsDefaultProfile  = "default"
	awsCredentialsFile = "credentials"
	awsConfigFile      = "config"
	awsBackupSuffix    = ".bak"
	mfaTokenLength     = 6
	awsInstallURL      = "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
)

// Static errors for AWS operations
var (
	errMFASerialNotFound   = errors.New("MFA serial not found in config. Run 'magex aws:setup' first")
	errInvalidMFAToken     = errors.New("MFA token must be exactly 6 digits")
	errEmptyInput          = errors.New("input cannot be empty")
	errSTSCallFailed       = errors.New("AWS STS get-session-token failed")
	errAWSCredBackupFailed = errors.New("failed to backup credentials file")
	errAWSCLINotFound      = errors.New("AWS CLI not found in PATH")
)

// AWS namespace for AWS credential management
type AWS mg.Namespace

// awsINISection represents a section in an INI file
type awsINISection struct {
	Name   string
	Values map[string]string
	// Preserve order of keys for consistent output
	KeyOrder []string
}

// awsINIFile represents a parsed INI file
type awsINIFile struct {
	Sections []*awsINISection
}

// awsSTSCredentials represents the response from AWS STS
type awsSTSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      string
}

// Login performs smart AWS login - detects if setup needed, then does MFA refresh
func (AWS) Login(args ...string) error {
	utils.Header("AWS Login")

	if err := checkAWSCLI(); err != nil {
		return err
	}

	params := utils.ParseParams(args)
	profile := utils.GetParam(params, "profile", awsDefaultProfile)

	// Check if setup exists
	if !hasValidAWSSetup(profile) {
		utils.Info("No existing setup detected for profile '%s'. Running setup...", profile)
		return (AWS{}).Setup(args...)
	}

	// Existing setup - do refresh
	utils.Info("Existing setup detected. Refreshing credentials...")
	return (AWS{}).Refresh(args...)
}

// Setup performs interactive AWS credential setup with base/session profile pattern
func (AWS) Setup(args ...string) error {
	utils.Header("AWS Setup")

	if err := checkAWSCLI(); err != nil {
		return err
	}

	params := utils.ParseParams(args)
	profileParam := utils.GetParam(params, "profile", "")

	var baseProfile, sessionProfile string

	if profileParam != "" {
		// If profile given, assume it's the session profile, derive base
		sessionProfile = profileParam
		baseProfile = profileParam + "-base"
		utils.Info("Setting up profiles: base='%s', session='%s'", baseProfile, sessionProfile)
	} else {
		// Prompt for both profile names
		utils.Info("This will set up a base profile (for long-term IAM keys) and a session profile (for temporary MFA credentials).")
		utils.Println("")

		var err error
		baseProfile, err = promptForNonEmpty("Base profile name (for long-term keys, e.g., mrz-base)")
		if err != nil {
			return err
		}

		sessionProfile, err = promptForNonEmpty("Session profile name (for temp creds, e.g., mrz)")
		if err != nil {
			return err
		}
	}

	utils.Println("")

	// Prompt for credentials
	accessKeyID, err := promptForNonEmpty("AWS Access Key ID")
	if err != nil {
		return err
	}

	secretKey, err := promptForNonEmpty("AWS Secret Access Key")
	if err != nil {
		return err
	}

	mfaSerial, err := promptForNonEmpty("MFA Serial ARN (e.g., arn:aws:iam::123456789012:mfa/username)")
	if err != nil {
		return err
	}

	// Validate MFA serial format
	if !strings.HasPrefix(mfaSerial, "arn:aws:iam::") {
		utils.Warn("MFA serial doesn't look like an ARN - please verify it's correct")
	}

	// Get AWS directory path
	awsDir, err := getAWSDir()
	if err != nil {
		return err
	}

	// Ensure ~/.aws directory exists
	if err := os.MkdirAll(awsDir, fileops.PermDirPrivate); err != nil {
		return fmt.Errorf("failed to create AWS directory: %w", err)
	}

	// Write credentials to BASE profile (long-term IAM keys)
	credPath := filepath.Join(awsDir, awsCredentialsFile)
	if err := writeAWSCredentials(credPath, baseProfile, accessKeyID, secretKey, ""); err != nil {
		return err
	}
	utils.Info("Wrote long-term credentials to profile '%s'", baseProfile)

	// Write MFA serial to BASE profile's config
	configPath := filepath.Join(awsDir, awsConfigFile)
	if err := writeAWSConfig(configPath, baseProfile, mfaSerial); err != nil {
		return err
	}

	// Write source_profile to SESSION profile's config (links session -> base)
	if err := writeAWSConfigSourceProfile(configPath, sessionProfile, baseProfile); err != nil {
		return err
	}
	utils.Info("Configured session profile '%s' with source_profile='%s'", sessionProfile, baseProfile)

	utils.Println("")
	utils.Success("AWS setup complete!")
	utils.Info("Base profile: %s (long-term credentials + MFA serial)", baseProfile)
	utils.Info("Session profile: %s (will store temporary credentials)", sessionProfile)
	utils.Println("")
	utils.Info("To refresh credentials, run: magex aws:refresh profile=%s", sessionProfile)

	return nil
}

// Refresh refreshes AWS session credentials using MFA
func (AWS) Refresh(args ...string) error {
	utils.Header("AWS MFA Refresh")

	if err := checkAWSCLI(); err != nil {
		return err
	}

	params := utils.ParseParams(args)
	profile := utils.GetParam(params, "profile", awsDefaultProfile)
	durationStr := utils.GetParam(params, "duration", "")
	baseParamExplicit := utils.GetParam(params, "base", "")

	duration := awsDefaultDuration
	if durationStr != "" {
		if _, err := fmt.Sscanf(durationStr, "%d", &duration); err != nil {
			utils.Warn("Invalid duration '%s', using default %d seconds", durationStr, awsDefaultDuration)
			duration = awsDefaultDuration
		}
	}

	// Determine base profile (where long-term credentials and MFA serial are stored)
	baseProfile := baseParamExplicit
	if baseProfile == "" {
		// Try to read source_profile from config
		baseProfile = getSourceProfile(profile)
	}
	if baseProfile == "" {
		// Fallback: same profile for base and session (backward compatibility)
		baseProfile = profile
	}

	// Get MFA serial from BASE profile's config
	mfaSerial, err := getMFASerial(baseProfile)
	if err != nil {
		return err
	}

	utils.Info("Session profile: %s", profile)
	if baseProfile != profile {
		utils.Info("Base profile: %s (source of long-term credentials)", baseProfile)
	}
	utils.Info("MFA Device: %s", mfaSerial)
	utils.Println("")

	// Prompt for MFA token
	mfaToken, err := promptForMFAToken()
	if err != nil {
		return err
	}

	// Call AWS STS using BASE profile (which has the long-term credentials)
	utils.Info("Getting session token from AWS STS...")
	creds, err := getAWSSessionToken(baseProfile, mfaSerial, mfaToken, duration)
	if err != nil {
		return err
	}

	// Backup and update credentials - write to SESSION profile
	awsDir, err := getAWSDir()
	if err != nil {
		return err
	}

	credPath := filepath.Join(awsDir, awsCredentialsFile)
	if err := writeOrUpdateAWSSessionCredentials(credPath, profile, creds); err != nil {
		return err
	}

	utils.Println("")
	utils.Success("Session credentials refreshed for profile '%s'", profile)
	utils.Info("Credentials valid until: %s", creds.Expiration)

	return nil
}

// Status shows AWS credential status
func (AWS) Status(args ...string) error {
	utils.Header("AWS Credential Status")

	params := utils.ParseParams(args)
	profileFilter := utils.GetParam(params, "profile", "")

	awsDir, err := getAWSDir()
	if err != nil {
		return err
	}

	// Check credentials file
	credPath := filepath.Join(awsDir, awsCredentialsFile)
	if _, statErr := os.Stat(credPath); os.IsNotExist(statErr) {
		utils.Warn("No credentials file found at %s", credPath)
		utils.Info("Run 'magex aws:setup' to configure credentials")
		return nil
	}

	// Parse credentials file
	data, err := os.ReadFile(credPath) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return fmt.Errorf("failed to read credentials: %w", err)
	}

	ini := parseAWSINI(data)

	// Check config file for MFA serials
	configPath := filepath.Join(awsDir, awsConfigFile)
	var configINI *awsINIFile
	if configData, configErr := os.ReadFile(configPath); configErr == nil { //nolint:gosec // path is constructed from known safe components
		configINI = parseAWSINI(configData)
	}

	found := false
	for _, section := range ini.Sections {
		if profileFilter != "" && section.Name != profileFilter {
			continue
		}
		found = true
		displayAWSProfileStatus(section, configINI)
	}

	if profileFilter != "" && !found {
		utils.Warn("Profile '%s' not found in credentials file", profileFilter)
	}

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// checkAWSCLI verifies the AWS CLI is installed
func checkAWSCLI() error {
	if !utils.CommandExists("aws") {
		return getAWSCLINotFoundError()
	}
	return nil
}

// getAWSCLINotFoundError returns an OS-specific error message for missing AWS CLI
func getAWSCLINotFoundError() error {
	msg := "AWS CLI not found in PATH.\n\n"

	if utils.IsWindows() {
		msg += "Install it using:\n"
		msg += "  MSI Installer: https://awscli.amazonaws.com/AWSCLIV2.msi\n"
		msg += "  Official Guide: " + awsInstallURL + "\n\n"
		msg += "After installation, restart your terminal and ensure 'aws.exe' is in your PATH."
	} else if utils.IsMac() {
		msg += "Install it using:\n"
		msg += "  Homebrew:    brew install awscli\n"
		msg += "  Official:    " + awsInstallURL + "\n\n"
		msg += "If already installed, ensure 'aws' is in your PATH."
	} else { // Linux
		msg += "Install it using:\n"
		msg += "  Ubuntu/Debian: sudo apt-get install awscli\n"
		msg += "  RHEL/CentOS:   sudo yum install aws-cli\n"
		msg += "  Official:      " + awsInstallURL + "\n\n"
		msg += "If already installed, ensure 'aws' is in your PATH."
	}

	return fmt.Errorf("%w: %s", errAWSCLINotFound, msg)
}

// getAWSDir returns the AWS configuration directory path
func getAWSDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".aws"), nil
}

// hasValidAWSSetup checks if a profile has valid setup (credentials and MFA serial)
func hasValidAWSSetup(profile string) bool {
	awsDir, err := getAWSDir()
	if err != nil {
		return false
	}

	// Check credentials file exists
	credPath := filepath.Join(awsDir, awsCredentialsFile)
	if _, statErr := os.Stat(credPath); os.IsNotExist(statErr) {
		return false
	}

	// Check if profile exists in credentials
	data, err := os.ReadFile(credPath) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return false
	}

	ini := parseAWSINI(data)

	// Look for profile section with access key
	for _, section := range ini.Sections {
		if section.Name == profile {
			// If access key exists, setup is detected
			// MFA serial check happens later in refresh flow
			if _, ok := section.Values["aws_access_key_id"]; ok {
				return true
			}
			break
		}
	}

	return false
}

// getMFASerial retrieves the MFA serial ARN from the config file
func getMFASerial(profile string) (string, error) {
	awsDir, err := getAWSDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(awsDir, awsConfigFile)
	data, err := os.ReadFile(configPath) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return "", errMFASerialNotFound
	}

	ini := parseAWSINI(data)

	// Look for profile section (in config, non-default profiles are prefixed with "profile ")
	sectionName := profile
	if profile != awsDefaultProfile {
		sectionName = "profile " + profile
	}

	for _, section := range ini.Sections {
		if section.Name == sectionName {
			if serial, ok := section.Values["mfa_serial"]; ok {
				return serial, nil
			}
		}
	}

	return "", errMFASerialNotFound
}

// getSourceProfile retrieves the source_profile from the config file for a given profile
func getSourceProfile(profile string) string {
	awsDir, err := getAWSDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(awsDir, awsConfigFile)
	data, err := os.ReadFile(configPath) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return ""
	}

	ini := parseAWSINI(data)

	// Look for profile section (in config, non-default profiles are prefixed with "profile ")
	sectionName := profile
	if profile != awsDefaultProfile {
		sectionName = "profile " + profile
	}

	for _, section := range ini.Sections {
		if section.Name == sectionName {
			if sourceProfile, ok := section.Values["source_profile"]; ok {
				return sourceProfile
			}
		}
	}

	return ""
}

// promptForNonEmpty prompts for input and validates it's not empty
func promptForNonEmpty(prompt string) (string, error) {
	value, err := utils.PromptForInput(prompt)
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: %s", errEmptyInput, prompt)
	}

	return value, nil
}

// promptForMFAToken prompts for and validates a 6-digit MFA token
func promptForMFAToken() (string, error) {
	token, err := utils.PromptForInput("Enter 6-digit MFA code")
	if err != nil {
		return "", err
	}

	token = strings.TrimSpace(token)
	if len(token) != mfaTokenLength {
		return "", errInvalidMFAToken
	}

	// Verify all digits using compiled regex
	mfaPattern := regexp.MustCompile(`^\d{6}$`)
	if !mfaPattern.MatchString(token) {
		return "", errInvalidMFAToken
	}

	return token, nil
}

// getAWSSessionToken calls AWS STS to get temporary credentials
func getAWSSessionToken(profile, mfaSerial, mfaToken string, duration int) (*awsSTSCredentials, error) {
	args := []string{
		"sts", "get-session-token",
		"--serial-number", mfaSerial,
		"--token-code", mfaToken,
		"--duration-seconds", fmt.Sprintf("%d", duration),
		"--output", "json",
	}

	if profile != awsDefaultProfile {
		args = append(args, "--profile", profile)
	}

	output, err := GetRunner().RunCmdOutput("aws", args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSTSCallFailed, err)
	}

	// Parse JSON response
	var response struct {
		Credentials struct {
			AccessKeyID     string `json:"AccessKeyId"`
			SecretAccessKey string `json:"SecretAccessKey"`
			SessionToken    string `json:"SessionToken"`
			Expiration      string `json:"Expiration"`
		} `json:"Credentials"`
	}

	if err := json.Unmarshal([]byte(output), &response); err != nil {
		return nil, fmt.Errorf("failed to parse STS response: %w", err)
	}

	return &awsSTSCredentials{
		AccessKeyID:     response.Credentials.AccessKeyID,
		SecretAccessKey: response.Credentials.SecretAccessKey,
		SessionToken:    response.Credentials.SessionToken,
		Expiration:      response.Credentials.Expiration,
	}, nil
}

// ============================================================================
// INI File Operations
// ============================================================================

// parseAWSINI parses an INI file from byte content
func parseAWSINI(data []byte) *awsINIFile {
	ini := &awsINIFile{Sections: []*awsINISection{}}
	var currentSection *awsINISection

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Section header: [section_name]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			name = strings.TrimSpace(name)
			currentSection = &awsINISection{
				Name:     name,
				Values:   make(map[string]string),
				KeyOrder: []string{},
			}
			ini.Sections = append(ini.Sections, currentSection)
			continue
		}

		// Key-value pair: key = value
		if currentSection != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			currentSection.Values[key] = value
			currentSection.KeyOrder = append(currentSection.KeyOrder, key)
		}
	}

	return ini
}

// writeAWSINI serializes an INI file to bytes
func writeAWSINI(ini *awsINIFile) []byte {
	var buf strings.Builder

	for i, section := range ini.Sections {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("[%s]\n", section.Name))

		// Write keys in order
		for _, key := range section.KeyOrder {
			if value, ok := section.Values[key]; ok {
				buf.WriteString(fmt.Sprintf("%s = %s\n", key, value))
			}
		}
	}

	return []byte(buf.String())
}

// getOrCreateSection finds or creates a section in the INI file
func getOrCreateSection(ini *awsINIFile, name string) *awsINISection {
	for _, section := range ini.Sections {
		if section.Name == name {
			return section
		}
	}

	// Create new section
	section := &awsINISection{
		Name:     name,
		Values:   make(map[string]string),
		KeyOrder: []string{},
	}
	ini.Sections = append(ini.Sections, section)
	return section
}

// setINIValue sets a value in a section, maintaining key order
func setINIValue(section *awsINISection, key, value string) {
	if _, exists := section.Values[key]; !exists {
		section.KeyOrder = append(section.KeyOrder, key)
	}
	section.Values[key] = value
}

// backupFile creates a backup of a file
func backupFile(path string) error {
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		return nil // Nothing to backup
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is from known safe source
	if err != nil {
		return fmt.Errorf("%w: %w", errAWSCredBackupFailed, err)
	}

	backupPath := path + awsBackupSuffix
	if err := os.WriteFile(backupPath, data, fileops.PermFileSensitive); err != nil {
		return fmt.Errorf("%w: %w", errAWSCredBackupFailed, err)
	}

	utils.Info("Backup created: %s", backupPath)
	return nil
}

// writeAWSCredentials writes credentials to the credentials file
func writeAWSCredentials(path, profile, accessKeyID, secretKey, sessionToken string) error {
	// Backup existing file
	if err := backupFile(path); err != nil {
		return err
	}

	// Load existing or create new
	var ini *awsINIFile
	if data, readErr := os.ReadFile(path); readErr == nil { //nolint:gosec // path is from known safe source
		ini = parseAWSINI(data)
	}
	if ini == nil {
		ini = &awsINIFile{Sections: []*awsINISection{}}
	}

	// Get or create profile section
	section := getOrCreateSection(ini, profile)
	setINIValue(section, "aws_access_key_id", accessKeyID)
	setINIValue(section, "aws_secret_access_key", secretKey)
	if sessionToken != "" {
		setINIValue(section, "aws_session_token", sessionToken)
	}

	// Write file with sensitive permissions
	if err := os.WriteFile(path, writeAWSINI(ini), fileops.PermFileSensitive); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// writeAWSConfig writes config to the config file
func writeAWSConfig(path, profile, mfaSerial string) error {
	// Backup existing file
	if err := backupFile(path); err != nil {
		return err
	}

	// Load existing or create new
	var ini *awsINIFile
	if data, readErr := os.ReadFile(path); readErr == nil { //nolint:gosec // path is from known safe source
		ini = parseAWSINI(data)
	}
	if ini == nil {
		ini = &awsINIFile{Sections: []*awsINISection{}}
	}

	// In config file, non-default profiles are prefixed with "profile "
	sectionName := profile
	if profile != awsDefaultProfile {
		sectionName = "profile " + profile
	}

	// Get or create profile section
	section := getOrCreateSection(ini, sectionName)
	setINIValue(section, "mfa_serial", mfaSerial)

	// Write file with sensitive permissions
	if err := os.WriteFile(path, writeAWSINI(ini), fileops.PermFileSensitive); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// writeAWSConfigSourceProfile writes source_profile to link a session profile to a base profile
func writeAWSConfigSourceProfile(path, sessionProfile, baseProfile string) error {
	// Load existing or create new (no backup - writeAWSConfig already did backup)
	var ini *awsINIFile
	if data, readErr := os.ReadFile(path); readErr == nil { //nolint:gosec // path is from known safe source
		ini = parseAWSINI(data)
	}
	if ini == nil {
		ini = &awsINIFile{Sections: []*awsINISection{}}
	}

	// In config file, non-default profiles are prefixed with "profile "
	sectionName := sessionProfile
	if sessionProfile != awsDefaultProfile {
		sectionName = "profile " + sessionProfile
	}

	// Get or create profile section
	section := getOrCreateSection(ini, sectionName)
	setINIValue(section, "source_profile", baseProfile)

	// Write file with sensitive permissions
	if err := os.WriteFile(path, writeAWSINI(ini), fileops.PermFileSensitive); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// writeOrUpdateAWSSessionCredentials writes session credentials, creating the profile if needed
func writeOrUpdateAWSSessionCredentials(path, profile string, creds *awsSTSCredentials) error {
	// Backup existing file
	if err := backupFile(path); err != nil {
		return err
	}

	// Load existing or create new
	var ini *awsINIFile
	if data, readErr := os.ReadFile(path); readErr == nil { //nolint:gosec // path is from known safe source
		ini = parseAWSINI(data)
	}
	if ini == nil {
		ini = &awsINIFile{Sections: []*awsINISection{}}
	}

	// Get or create profile section
	section := getOrCreateSection(ini, profile)

	// Update with session credentials
	setINIValue(section, "aws_access_key_id", creds.AccessKeyID)
	setINIValue(section, "aws_secret_access_key", creds.SecretAccessKey)
	setINIValue(section, "aws_session_token", creds.SessionToken)

	// Write file with sensitive permissions
	if err := os.WriteFile(path, writeAWSINI(ini), fileops.PermFileSensitive); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// maskCredential masks a credential for display
func maskCredential(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

// checkAWSSession checks if an AWS session is valid by calling sts get-caller-identity
// Returns (accountID, userARN, isValid)
func checkAWSSession(profile string) (string, string, bool) {
	args := []string{"sts", "get-caller-identity", "--output", "json"}
	if profile != awsDefaultProfile {
		args = append(args, "--profile", profile)
	}

	output, err := GetRunner().RunCmdOutput("aws", args...)
	if err != nil {
		return "", "", false
	}

	// Parse JSON response
	var response struct {
		Account string `json:"Account"`
		Arn     string `json:"Arn"`
		UserID  string `json:"UserId"`
	}

	if err := json.Unmarshal([]byte(output), &response); err != nil {
		return "", "", false
	}

	return response.Account, response.Arn, true
}

// displayAWSProfileStatus displays the status of a single profile
func displayAWSProfileStatus(section *awsINISection, configINI *awsINIFile) {
	utils.Println("")
	utils.Info("Profile: %s", section.Name)

	// Show masked access key
	if accessKey, ok := section.Values["aws_access_key_id"]; ok {
		utils.Info("  Access Key ID: %s", maskCredential(accessKey))
	}

	// Check for session token
	hasSessionToken := false
	if _, ok := section.Values["aws_session_token"]; ok {
		hasSessionToken = true
		utils.Info("  Session Token: Present")
	} else {
		utils.Info("  Session Token: Not present (long-term credentials)")
	}

	// Show MFA serial if available
	if configINI != nil {
		sectionName := section.Name
		if section.Name != awsDefaultProfile {
			sectionName = "profile " + section.Name
		}
		for _, configSection := range configINI.Sections {
			if configSection.Name == sectionName {
				if mfaSerial, ok := configSection.Values["mfa_serial"]; ok {
					utils.Info("  MFA Device: %s", mfaSerial)
				}
				break
			}
		}
	}

	// Check if session is valid by calling AWS
	accountID, userARN, isValid := checkAWSSession(section.Name)
	if isValid {
		utils.Success("  Session Status: ✓ Active (Account: %s)", accountID)
		// Extract username from ARN for cleaner display
		if parts := strings.Split(userARN, "/"); len(parts) > 1 {
			utils.Info("  Identity: %s", parts[len(parts)-1])
		}
	} else {
		if hasSessionToken {
			utils.Error("  Session Status: ✗ Expired or Invalid")
			utils.Info("  Run 'magex aws:refresh profile=%s' to refresh", section.Name)
		} else {
			utils.Warn("  Session Status: ✗ Not authenticated (needs MFA)")
		}
	}
}

// ============================================================================
// Placeholder methods for commands without args (required for registry)
// ============================================================================

// LoginNoArgs is a placeholder that directs users to use Login with args
func (a AWS) LoginNoArgs() error {
	return a.Login()
}

// SetupNoArgs is a placeholder that directs users to use Setup with args
func (a AWS) SetupNoArgs() error {
	return a.Setup()
}

// RefreshNoArgs is a placeholder that directs users to use Refresh with args
func (a AWS) RefreshNoArgs() error {
	return a.Refresh()
}

// StatusNoArgs is a placeholder that directs users to use Status with args
func (a AWS) StatusNoArgs() error {
	return a.Status()
}
