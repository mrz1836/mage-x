// Package security provides secure command execution and input validation
// for the MAGE-X build system.
//
// # Input Validation
//
// The package offers comprehensive validation for user inputs:
//   - ValidateURL: URL safety checks including SSRF protection
//   - ValidateVersion: Semantic version validation with injection prevention
//   - ValidatePath: Path traversal and injection prevention
//   - ValidateFilename: Safe filename validation
//   - ValidateEmail: Email format validation
//   - ValidatePort: Network port range validation
//   - ValidateGitRef: Git reference safety validation
//
// # Command Execution
//
// SecureExecutor provides safe command execution with:
//   - Argument validation and sanitization
//   - Environment variable filtering
//   - Command injection prevention
//
// # Error Types
//
// All validation errors are defined as package-level variables
// enabling use with errors.Is() for specific error handling.
//
// Example:
//
//	if err := security.ValidateURL(url); err != nil {
//	    if errors.Is(err, security.ErrURLInvalidProtocol) {
//	        // Handle invalid protocol
//	    }
//	}
package security
