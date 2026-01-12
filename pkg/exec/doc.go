// Package exec provides command execution utilities with security
// and reliability features for the MAGE-X build system.
//
// # Executor Interface
//
// The package defines composable executor interfaces:
//   - Executor: Basic command execution
//   - ExecutorWithEnv: Execution with custom environment
//   - ExecutorWithDir: Execution in specific directory
//   - StreamingExecutor: Real-time output streaming
//   - FullExecutor: Combines all capabilities
//
// # Decorator Pattern
//
// Executors can be composed using decorators:
//
//	base := exec.NewBase()
//	secured := exec.WithValidation(base)
//	retrying := exec.WithRetry(secured, classifier)
//	final := exec.WithTimeout(retrying, 5*time.Minute)
//
// # Security Features
//
// The package includes:
//   - Argument validation against injection patterns
//   - Sensitive environment variable filtering
//   - Path traversal prevention
//
// # Usage
//
// For basic command execution:
//
//	executor := exec.NewBase()
//	err := executor.Execute(ctx, "go", "build", "./...")
//
// For secured execution with validation:
//
//	executor := exec.NewSecure()
//	output, err := executor.ExecuteOutput(ctx, "git", "status")
package exec
