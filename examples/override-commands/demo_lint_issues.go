package main

import (
	"github.com/mrz1836/mage-x/pkg/utils"
)

// This file contains intentional lint issues for demonstration

// DemoFunction shows issues that custom lint overrides can catch
func DemoFunction() {
	// This was fmt.Println but fixed to use utils.Info
	utils.Info("This will be flagged by strict linting")

	// Remove this comment - caught by custom pre-lint checks

	// This is fine for regular lint
	var unused string
	_ = unused // Avoid unused variable warning
}

// Another comment for testing
