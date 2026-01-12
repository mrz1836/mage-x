package security_test

import (
	"fmt"

	"github.com/mrz1836/mage-x/pkg/security"
)

func ExampleValidateVersion() {
	// Valid semantic versions
	validVersions := []string{"1.0.0", "v1.2.3", "2.0.0-beta.1", "1.0.0+build.123"}

	for _, v := range validVersions {
		if err := security.ValidateVersion(v); err == nil {
			fmt.Printf("%s: valid\n", v)
		}
	}
	// Output:
	// 1.0.0: valid
	// v1.2.3: valid
	// 2.0.0-beta.1: valid
	// 1.0.0+build.123: valid
}

func ExampleValidateVersion_invalid() {
	// Invalid versions are rejected
	err := security.ValidateVersion("not-a-version")
	if err != nil {
		fmt.Println("Invalid version rejected")
	}
	// Output: Invalid version rejected
}

func ExampleValidateURL() {
	// Valid URLs are accepted
	if err := security.ValidateURL("https://github.com/mrz1836/mage-x"); err == nil {
		fmt.Println("GitHub URL: valid")
	}

	// Invalid protocols are rejected
	if err := security.ValidateURL("ftp://example.com"); err != nil {
		fmt.Println("FTP URL: rejected (invalid protocol)")
	}
	// Output:
	// GitHub URL: valid
	// FTP URL: rejected (invalid protocol)
}

func ExampleValidateFilename() {
	// Safe filenames are accepted
	safeNames := []string{"main.go", "config-v2.yaml", "README.md"}

	for _, name := range safeNames {
		if err := security.ValidateFilename(name); err == nil {
			fmt.Printf("%s: valid\n", name)
		}
	}
	// Output:
	// main.go: valid
	// config-v2.yaml: valid
	// README.md: valid
}

func ExampleValidateFilename_pathTraversal() {
	// Path traversal attempts are blocked
	err := security.ValidateFilename("../../../etc/passwd")
	if err != nil {
		fmt.Println("Path traversal blocked")
	}
	// Output: Path traversal blocked
}

func ExampleValidateGitRef() {
	// Valid git refs
	validRefs := []string{"main", "feature/new-feature", "v1.0.0"}

	for _, ref := range validRefs {
		if err := security.ValidateGitRef(ref); err == nil {
			fmt.Printf("%s: valid\n", ref)
		}
	}
	// Output:
	// main: valid
	// feature/new-feature: valid
	// v1.0.0: valid
}

func ExampleValidateEnvVar() {
	// Valid environment variable names
	validNames := []string{"PATH", "HOME", "MY_VAR", "_INTERNAL"}

	for _, name := range validNames {
		if err := security.ValidateEnvVar(name); err == nil {
			fmt.Printf("%s: valid\n", name)
		}
	}
	// Output:
	// PATH: valid
	// HOME: valid
	// MY_VAR: valid
	// _INTERNAL: valid
}

func ExampleValidateEnvVar_invalid() {
	// Invalid names are rejected
	invalidNames := []string{"123_VAR", "MY-VAR", ""}

	for _, name := range invalidNames {
		if err := security.ValidateEnvVar(name); err != nil {
			fmt.Printf("%q: rejected\n", name)
		}
	}
	// Output:
	// "123_VAR": rejected
	// "MY-VAR": rejected
	// "": rejected
}

func ExampleValidatePort() {
	// Valid non-privileged ports
	if err := security.ValidatePort(8080); err == nil {
		fmt.Println("Port 8080: valid")
	}

	if err := security.ValidatePort(3000); err == nil {
		fmt.Println("Port 3000: valid")
	}

	// Invalid port
	if err := security.ValidatePort(0); err != nil {
		fmt.Println("Port 0: rejected")
	}
	// Output:
	// Port 8080: valid
	// Port 3000: valid
	// Port 0: rejected
}

func ExampleValidateEmail() {
	// Valid emails
	if err := security.ValidateEmail("user@example.com"); err == nil {
		fmt.Println("user@example.com: valid")
	}

	// Invalid emails
	if err := security.ValidateEmail("not-an-email"); err != nil {
		fmt.Println("not-an-email: rejected")
	}
	// Output:
	// user@example.com: valid
	// not-an-email: rejected
}
