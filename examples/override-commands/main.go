package main

import (
	"fmt"
	"os"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Version can be injected at build time via ldflags
const Version = "development"

func main() {
	fmt.Printf("Example Override Commands Application\n")
	fmt.Printf("Version: %s\n", Version)

	if len(os.Args) > 1 {
		fmt.Printf("Args: %v\n", os.Args[1:])
	}

	// This demonstrates the custom build process
	utils.Info("✅ Application built with custom override commands!")
}
