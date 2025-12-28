// Package utils provides utility functions for mage tasks.
//
// This package is organized into focused modules:
//   - cmd.go: Command execution (RunCmd, RunCmdV, RunCmdOutput, RunCmdPipe)
//   - env.go: Environment variable helpers (GetEnv, GetEnvBool, GetEnvInt, IsVerbose, IsCI)
//   - fs.go: File system operations (FileExists, DirExists, EnsureDir, CleanDir, CopyFile)
//   - platform.go: Platform detection (GetCurrentPlatform, ParsePlatform, IsWindows, IsMac, IsLinux)
//   - golang.go: Go tooling utilities (GoList, GetModuleName, GetGoVersion)
//   - misc.go: Miscellaneous utilities (Parallel, FormatDuration, FormatBytes, PromptForInput)
//   - logger.go: Logging functions (Header, Success, Error, Info, Warn)
//   - download.go: Download utilities with retry support
//   - spinner.go: Progress spinner for long-running operations
//   - metrics.go: Metrics collection and reporting
//   - params.go: Parameter parsing and validation
//   - build_tags.go: Build tag parsing utilities
//   - system_memory.go: System memory detection
package utils
