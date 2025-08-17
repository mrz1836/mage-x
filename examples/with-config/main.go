package main

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// These variables will be injected at build time via ldflags in .mage.yaml
var (
	version   = "dev"
	commit    = "unknown" //nolint:gochecknoglobals // build-time injection
	buildDate = "unknown" //nolint:gochecknoglobals // build-time injection
)

// ErrDivisionByZero is returned when attempting to divide by zero
var ErrDivisionByZero = errors.New("division by zero")

func main() {
	// Check for version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("config-demo %s (commit: %s, built: %s)\n", version, commit, buildDate)
		return
	}

	// Read environment variables configured in .mage.yaml
	logLevel := getEnv("LOG_LEVEL", "info")
	appEnv := getEnv("APP_ENV", "development")

	log.Printf("Starting config-demo application")
	log.Printf("Version: %s", version)
	log.Printf("Log Level: %s", logLevel)
	log.Printf("Environment: %s", appEnv)

	// Demonstrate some basic functionality
	calculator := NewCalculator()
	result := calculator.Add(10, 20)

	fmt.Printf("Calculator result: %d\n", result)
	log.Println("Application completed successfully!")
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Calculator demonstrates a simple struct for testing
type Calculator struct{}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Add performs addition of two integers
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// Multiply performs multiplication of two integers
func (c *Calculator) Multiply(a, b int) int {
	return a * b
}

// Divide performs division of two integers
func (c *Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}
