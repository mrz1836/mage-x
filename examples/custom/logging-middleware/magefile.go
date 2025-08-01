//go:build mage
// +build mage

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage"
)

// LoggingBuild wraps any BuildNamespace with comprehensive logging
type LoggingBuild struct {
	mage.BuildNamespace
	logger *log.Logger
}

// NewLoggingBuild creates a build namespace with logging middleware
func NewLoggingBuild(inner mage.BuildNamespace, logger *log.Logger) *LoggingBuild {
	return &LoggingBuild{
		BuildNamespace: inner,
		logger:         logger,
	}
}

func (l *LoggingBuild) Default() error {
	return l.withLogging("Default", func() error {
		return l.BuildNamespace.Default()
	})
}

func (l *LoggingBuild) All() error {
	return l.withLogging("All", func() error {
		return l.BuildNamespace.All()
	})
}

func (l *LoggingBuild) Platform(platform string) error {
	return l.withLogging(fmt.Sprintf("Platform(%s)", platform), func() error {
		return l.BuildNamespace.Platform(platform)
	})
}

func (l *LoggingBuild) Linux() error {
	return l.withLogging("Linux", func() error {
		return l.BuildNamespace.Linux()
	})
}

func (l *LoggingBuild) Darwin() error {
	return l.withLogging("Darwin", func() error {
		return l.BuildNamespace.Darwin()
	})
}

func (l *LoggingBuild) Windows() error {
	return l.withLogging("Windows", func() error {
		return l.BuildNamespace.Windows()
	})
}

func (l *LoggingBuild) Docker() error {
	return l.withLogging("Docker", func() error {
		return l.BuildNamespace.Docker()
	})
}

func (l *LoggingBuild) Clean() error {
	return l.withLogging("Clean", func() error {
		return l.BuildNamespace.Clean()
	})
}

func (l *LoggingBuild) Install() error {
	return l.withLogging("Install", func() error {
		return l.BuildNamespace.Install()
	})
}

func (l *LoggingBuild) Generate() error {
	return l.withLogging("Generate", func() error {
		return l.BuildNamespace.Generate()
	})
}

func (l *LoggingBuild) PreBuild() error {
	return l.withLogging("PreBuild", func() error {
		return l.BuildNamespace.PreBuild()
	})
}

func (l *LoggingBuild) withLogging(operation string, fn func() error) error {
	l.logger.Printf("🔨 Starting build operation: %s", operation)
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	if err != nil {
		l.logger.Printf("❌ Build operation %s failed after %v: %v", operation, duration, err)
	} else {
		l.logger.Printf("✅ Build operation %s completed successfully in %v", operation, duration)
	}

	return err
}

// LoggingTest wraps any TestNamespace with logging
type LoggingTest struct {
	mage.TestNamespace
	logger *log.Logger
}

func NewLoggingTest(inner mage.TestNamespace, logger *log.Logger) *LoggingTest {
	return &LoggingTest{
		TestNamespace: inner,
		logger:        logger,
	}
}

func (l *LoggingTest) Default() error {
	return l.withLogging("Default", func() error {
		return l.TestNamespace.Default()
	})
}

func (l *LoggingTest) Unit() error {
	return l.withLogging("Unit", func() error {
		return l.TestNamespace.Unit()
	})
}

func (l *LoggingTest) Integration() error {
	return l.withLogging("Integration", func() error {
		return l.TestNamespace.Integration()
	})
}

func (l *LoggingTest) Bench() error {
	return l.withLogging("Bench", func() error {
		return l.TestNamespace.Bench()
	})
}

func (l *LoggingTest) Coverage() error {
	return l.withLogging("Coverage", func() error {
		return l.TestNamespace.Coverage()
	})
}

func (l *LoggingTest) Race() error {
	return l.withLogging("Race", func() error {
		return l.TestNamespace.Race()
	})
}

func (l *LoggingTest) Short() error {
	return l.withLogging("Short", func() error {
		return l.TestNamespace.Short()
	})
}

func (l *LoggingTest) CI() error {
	return l.withLogging("CI", func() error {
		return l.TestNamespace.CI()
	})
}

func (l *LoggingTest) All() error {
	return l.withLogging("All", func() error {
		return l.TestNamespace.All()
	})
}

func (l *LoggingTest) withLogging(operation string, fn func() error) error {
	l.logger.Printf("🧪 Starting test operation: %s", operation)
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	if err != nil {
		l.logger.Printf("❌ Test operation %s failed after %v: %v", operation, duration, err)
	} else {
		l.logger.Printf("✅ Test operation %s completed successfully in %v", operation, duration)
	}

	return err
}

// Global logger setup
var (
	buildLogger = log.New(os.Stdout, "[BUILD] ", log.LstdFlags|log.Lshortfile) //nolint:gochecknoglobals // Example code demonstrating global loggers
	testLogger  = log.New(os.Stdout, "[TEST]  ", log.LstdFlags|log.Lshortfile) //nolint:gochecknoglobals // Example code demonstrating global loggers
)

// Build with comprehensive logging
func Build() error {
	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, buildLogger)
	return loggingBuild.Default()
}

// BuildAll builds for all platforms with logging
func BuildAll() error {
	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, buildLogger)
	return loggingBuild.All()
}

// BuildPlatform builds for specific platform with logging
func BuildPlatform(platform string) error {
	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, buildLogger)
	return loggingBuild.Platform(platform)
}

// Test with comprehensive logging
func Test() error {
	baseTest := mage.NewTestNamespace()
	loggingTest := NewLoggingTest(baseTest, testLogger)
	return loggingTest.Default()
}

// TestUnit runs unit tests with logging
func TestUnit() error {
	baseTest := mage.NewTestNamespace()
	loggingTest := NewLoggingTest(baseTest, testLogger)
	return loggingTest.Unit()
}

// TestCoverage runs coverage tests with logging
func TestCoverage() error {
	baseTest := mage.NewTestNamespace()
	loggingTest := NewLoggingTest(baseTest, testLogger)
	return loggingTest.Coverage()
}

// CI runs complete pipeline with detailed logging
func CI() error {
	fmt.Println("🚀 Starting CI Pipeline with detailed logging...")

	// Create loggers for different operations
	lintLogger := log.New(os.Stdout, "[LINT]  ", log.LstdFlags)
	localBuildLogger := log.New(os.Stdout, "[BUILD] ", log.LstdFlags)
	localTestLogger := log.New(os.Stdout, "[TEST]  ", log.LstdFlags)

	// Lint with logging
	lintLogger.Println("🔍 Starting linting operation...")
	lint := mage.NewLintNamespace()
	if err := lint.Default(); err != nil {
		lintLogger.Printf("❌ Linting failed: %v", err)
		return fmt.Errorf("linting failed: %w", err)
	}
	lintLogger.Println("✅ Linting completed successfully")

	// Test with logging middleware
	baseTest := mage.NewTestNamespace()
	loggingTest := NewLoggingTest(baseTest, localTestLogger)
	if err := loggingTest.Coverage(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Build with logging middleware
	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, localBuildLogger)
	if err := loggingBuild.Default(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("🎉 CI Pipeline completed successfully with full logging!")
	return nil
}

// Clean with logging
func Clean() error {
	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, buildLogger)
	return loggingBuild.Clean()
}

// DeployPipeline shows a complex deployment with logging
func DeployPipeline() error {
	deployLogger := log.New(os.Stdout, "[DEPLOY] ", log.LstdFlags)

	deployLogger.Println("🚀 Starting deployment pipeline...")

	// Build for multiple platforms
	platforms := []string{"linux/amd64", "darwin/amd64", "windows/amd64"}

	baseBuild := mage.NewBuildNamespace()
	loggingBuild := NewLoggingBuild(baseBuild, buildLogger)

	for _, platform := range platforms {
		deployLogger.Printf("📦 Building for platform: %s", platform)
		if err := loggingBuild.Platform(platform); err != nil {
			deployLogger.Printf("❌ Platform build failed for %s: %v", platform, err)
			return fmt.Errorf("platform build failed for %s: %w", platform, err)
		}
	}

	deployLogger.Println("✅ All platform builds completed successfully")

	// Docker build if available
	if _, err := os.Stat("Dockerfile"); err == nil {
		dockerLogger := log.New(os.Stdout, "[DOCKER] ", log.LstdFlags)
		dockerLogger.Println("🐳 Building Docker image...")

		build := mage.NewBuildNamespace()
		if err := build.Docker(); err != nil {
			dockerLogger.Printf("❌ Docker build failed: %v", err)
			return fmt.Errorf("docker build failed: %w", err)
		}
		dockerLogger.Println("✅ Docker image built successfully")
	}

	deployLogger.Println("🎉 Deployment pipeline completed successfully!")
	return nil
}
