// Package mage provides AWS credential management commands
package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// AWS credential management constants
const (
	awsDefaultDuration   = 43200 // 12 hours in seconds
	awsDefaultProfile    = "default"
	awsBaseProfileSuffix = "-base"
	awsCredentialsFile   = "credentials"
	awsConfigFile        = "config"
	awsBackupSuffix      = ".bak"
	mfaTokenLength       = 6
	awsInstallURL        = "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"

	// awsSessionCredsRejection is how STS refuses GetSessionToken when called with
	// temporary credentials: only long-term IAM keys can start a new session
	awsSessionCredsRejection = "Cannot call GetSessionToken with session credentials"

	// awsTemporaryKeyPrefix starts every temporary STS access key ID; long-term
	// IAM user keys start with "AKIA"
	awsTemporaryKeyPrefix = "ASIA"
)

// Static errors for AWS operations
var (
	errMFASerialNotFound          = errors.New("MFA serial not found in config. Run 'magex aws:setup' first")
	errInvalidMFAToken            = errors.New("MFA token must be exactly 6 digits")
	errEmptyInput                 = errors.New("input cannot be empty")
	errSTSCallFailed              = errors.New("AWS STS get-session-token failed")
	errAWSCredBackupFailed        = errors.New("failed to backup credentials file")
	errAWSCLINotFound             = errors.New("AWS CLI not found in PATH")
	errBaseHasSessionCreds        = errors.New("base profile holds temporary session credentials instead of long-term IAM keys")
	errWouldOverwriteLongTermKeys = errors.New("refusing to overwrite long-term IAM keys with session credentials")
)

// mfaTokenPattern validates 6-digit MFA tokens (compiled once at package level)
//

var mfaTokenPattern = regexp.MustCompile(`^\d{6}$`)

// getConfigSectionName returns the INI section name for a profile
// In AWS config files, non-default profiles are prefixed with "profile "
func getConfigSectionName(profile string) string {
	if profile == awsDefaultProfile {
		return profile
	}
	return "profile " + profile
}

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
// #nosec G117 -- SessionToken field name matches secret pattern by design (it is a credential field)
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
		baseProfile = profileParam + awsBaseProfileSuffix
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

	awsDir, err := getAWSDir()
	if err != nil {
		return err
	}
	credPath := filepath.Join(awsDir, awsCredentialsFile)

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

	// Writing the session into the base itself would replace its long-term keys
	// with temporary ones, which STS can't start a new session from
	if baseProfile == profile && hasLongTermKeys(loadAWSCredentialsSection(credPath, profile)) {
		sessionProfile, linkErr := sessionProfileForBase(baseProfile)
		if linkErr != nil {
			return linkErr
		}
		utils.Info("'%s' holds your long-term keys; refreshing its session profile '%s' instead", baseProfile, sessionProfile)
		profile = sessionProfile
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

	// STS would reject the call, so don't ask for an MFA code it can't use
	if hasTemporaryCreds(loadAWSCredentialsSection(credPath, baseProfile)) {
		return reportBaseHasSessionCreds(credPath, profile, baseProfile)
	}

	// Prompt for MFA token
	mfaToken, err := promptForMFAToken()
	if err != nil {
		return err
	}

	// Call AWS STS using BASE profile (which has the long-term credentials)
	utils.Info("Getting session token from AWS STS...")
	creds, err := getAWSSessionToken(baseProfile, mfaSerial, mfaToken, duration)
	if err != nil {
		// Temporary credentials resolved from elsewhere (e.g. environment
		// variables or the config file) only show up as this rejection
		if strings.Contains(err.Error(), awsSessionCredsRejection) {
			utils.Println("")
			return reportBaseHasSessionCreds(credPath, profile, baseProfile)
		}
		return err
	}

	// Backup and update credentials - write to SESSION profile
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

	// Resolve session -> base so a filter like "profile=mrz-ro" also matches
	// credentials stored under "mrz-ro-base".
	resolvedBase := ""
	if profileFilter != "" {
		resolvedBase = getSourceProfile(profileFilter)
	}

	found := false
	for _, section := range ini.Sections {
		if profileFilter != "" {
			matchesLiteral := section.Name == profileFilter
			matchesBase := resolvedBase != "" && section.Name == resolvedBase
			if !matchesLiteral && !matchesBase {
				continue
			}
		}
		found = true
		displayAWSProfileStatus(section, configINI)
	}

	if profileFilter != "" && !found {
		if resolvedBase != "" {
			utils.Warn("Profile '%s' (base '%s') not found in credentials file", profileFilter, resolvedBase)
		} else {
			utils.Warn("Profile '%s' not found in credentials file", profileFilter)
		}
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

// hasValidAWSSetup checks if a profile has valid setup: the resolved base
// profile must have aws_access_key_id in credentials AND mfa_serial in config.
// The argument may be either a session profile (resolved via source_profile)
// or a legacy single-profile setup (falls back to literal name).
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

	// Resolve session -> base profile (legacy single-profile setups fall back to literal name).
	baseProfile := getSourceProfile(profile)
	if baseProfile == "" {
		baseProfile = profile
	}

	// Verify long-term access key under the BASE profile section in credentials.
	data, err := os.ReadFile(credPath) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return false
	}

	ini := parseAWSINI(data)

	hasKey := false
	for _, section := range ini.Sections {
		if section.Name == baseProfile {
			if _, ok := section.Values["aws_access_key_id"]; ok {
				hasKey = true
			}
			break
		}
	}
	if !hasKey {
		return false
	}

	// Setup is only valid if MFA serial is configured for the base profile.
	if _, mfaErr := getMFASerial(baseProfile); mfaErr != nil {
		return false
	}

	return true
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
	sectionName := getConfigSectionName(profile)

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
	sectionName := getConfigSectionName(profile)

	for _, section := range ini.Sections {
		if section.Name == sectionName {
			if sourceProfile, ok := section.Values["source_profile"]; ok {
				return sourceProfile
			}
		}
	}

	return ""
}

// linkedSessionProfiles returns the profiles whose source_profile points at
// baseProfile, as aws:setup links them. Assume-role profiles are skipped: the
// AWS CLI ignores static credentials stored for them.
func linkedSessionProfiles(baseProfile string) []string {
	awsDir, err := getAWSDir()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(awsDir, awsConfigFile)) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return nil
	}

	var linked []string
	for _, section := range parseAWSINI(data).Sections {
		name, isProfile := strings.CutPrefix(section.Name, "profile ")
		if !isProfile && section.Name != awsDefaultProfile {
			continue // sso-session and other non-profile sections
		}
		if name == baseProfile || section.Values["source_profile"] != baseProfile || section.Values["role_arn"] != "" {
			continue
		}
		linked = append(linked, name)
	}

	return linked
}

// sessionProfileForBase returns the session profile to refresh when the user
// names a base profile that holds long-term keys. Without exactly one linked
// session profile it explains the alternatives and returns an error.
func sessionProfileForBase(baseProfile string) (string, error) {
	linked := linkedSessionProfiles(baseProfile)
	if len(linked) == 1 {
		return linked[0], nil
	}

	utils.Warn("Profile '%s' holds long-term IAM keys; refreshing it in place would replace them with temporary credentials that can't be refreshed again", baseProfile)
	if len(linked) == 0 {
		utils.Info("Store the session in its own profile instead: magex aws:refresh profile=<session-profile> base=%s", baseProfile)
	} else {
		utils.Info("Refresh one of its session profiles instead (%s), e.g.: magex aws:refresh profile=%s", strings.Join(linked, ", "), linked[0])
	}

	return "", fmt.Errorf("%w: %s", errWouldOverwriteLongTermKeys, baseProfile)
}

// loadAWSCredentialsSection returns a profile's section of a credentials file,
// or nil when the file or section doesn't exist
func loadAWSCredentialsSection(path, profile string) *awsINISection {
	data, err := os.ReadFile(path) //nolint:gosec // path is constructed from known safe components
	if err != nil {
		return nil
	}

	for _, section := range parseAWSINI(data).Sections {
		if section.Name == profile {
			return section
		}
	}

	return nil
}

// hasTemporaryCreds reports whether a credentials section holds temporary STS
// credentials. The access key prefix still identifies them after the session
// token line has been deleted by hand.
func hasTemporaryCreds(section *awsINISection) bool {
	if section == nil {
		return false
	}
	return section.Values["aws_session_token"] != "" ||
		strings.HasPrefix(section.Values["aws_access_key_id"], awsTemporaryKeyPrefix)
}

// hasLongTermKeys reports whether a credentials section holds long-term IAM keys
func hasLongTermKeys(section *awsINISection) bool {
	return section != nil && section.Values["aws_access_key_id"] != "" && !hasTemporaryCreds(section)
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

	// Verify all digits using package-level compiled regex
	if !mfaTokenPattern.MatchString(token) {
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
	// #nosec G117 -- SessionToken field name matches secret pattern by design (it is a credential field)
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
		fmt.Fprintf(&buf, "[%s]\n", section.Name)

		// Write keys in order
		for _, key := range section.KeyOrder {
			if value, ok := section.Values[key]; ok {
				fmt.Fprintf(&buf, "%s = %s\n", key, value)
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

// deleteINIValue removes a key from a section, keeping the order of the rest
func deleteINIValue(section *awsINISection, key string) {
	delete(section.Values, key)
	section.KeyOrder = slices.DeleteFunc(section.KeyOrder, func(k string) bool { return k == key })
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
	if err := os.WriteFile(backupPath, data, fileops.PermFileSensitive); err != nil { // #nosec G703 -- backupPath is constructed from validated config
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
	} else {
		// A leftover token would be sent along with the new long-term keys, and AWS rejects that pair
		deleteINIValue(section, "aws_session_token")
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

	sectionName := getConfigSectionName(profile)

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

// reportBaseHasSessionCreds handles a refresh that STS can't serve because the
// base profile's credentials are temporary. It isn't an error while the requested
// profile is still logged in; either way it explains how to recover.
func reportBaseHasSessionCreds(credPath, profile, baseProfile string) error {
	accountID, _, active := checkAWSSession(profile)
	if active {
		utils.Success("Already logged in: profile '%s' has an active session (Account: %s)", profile, accountID)
	}

	if hasTemporaryCreds(loadAWSCredentialsSection(credPath, baseProfile)) {
		utils.Warn("Profile '%s' holds temporary session credentials (ASIA… keys) instead of long-term IAM keys (AKIA… keys), so AWS can't start a new session from it", baseProfile)
		printRestoreKeysHint(credPath, profile, baseProfile)
	} else {
		utils.Warn("The AWS CLI resolved temporary session credentials for '%s' instead of long-term IAM keys, so AWS can't start a new session from them", baseProfile)
		listCmd := "aws configure list"
		if baseProfile != awsDefaultProfile {
			listCmd += " --profile " + baseProfile
		}
		utils.Info("See where they come from with: %s", listCmd)
	}

	if active {
		return nil
	}
	return fmt.Errorf("%w: %s", errBaseHasSessionCreds, baseProfile)
}

// printRestoreKeysHint explains how to put long-term keys back into baseProfile,
// pointing at the credentials backup when it still holds them
func printRestoreKeysHint(credPath, profile, baseProfile string) {
	setupCmd := "magex aws:setup profile=" + setupProfileName(profile, baseProfile)

	backupPath := credPath + awsBackupSuffix
	if hasLongTermKeys(loadAWSCredentialsSection(backupPath, baseProfile)) {
		utils.Info("%s still has the long-term keys for '%s'", backupPath, baseProfile)
		utils.Info("Replace both aws_access_key_id and aws_secret_access_key under [%s] in %s with the ones from that backup, and delete any aws_session_token line there",
			baseProfile, credPath)
		utils.Info("Do it before running aws:setup or refreshing another profile, since both overwrite that backup")
		utils.Info("Or re-enter your keys with: %s", setupCmd)
		return
	}

	utils.Info("Re-enter your long-term keys with: %s", setupCmd)
}

// setupProfileName returns the profile= value for aws:setup, which stores the
// long-term keys under "<profile>-base": baseProfile itself when it follows that
// naming, otherwise a new base for profile
func setupProfileName(profile, baseProfile string) string {
	if name := strings.TrimSuffix(baseProfile, awsBaseProfileSuffix); name != baseProfile && name != "" {
		return name
	}
	return profile
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
