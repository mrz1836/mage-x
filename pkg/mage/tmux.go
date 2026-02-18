// Package mage provides tmux session management commands
package mage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// tmux session management constants
const (
	tmuxDefaultModel  = "sonnet"
	tmuxClaudeCommand = "claude"
	tmuxInstallURL    = "https://github.com/tmux/tmux/wiki/Installing"
	claudeInstallURL  = "https://docs.anthropic.com/en/docs/claude-code"
)

// Static errors for tmux operations
var (
	errTmuxNotFound       = errors.New("tmux not found in PATH")
	errSessionNotFound    = errors.New("session not found")
	errInvalidSessionName = errors.New("session name cannot be empty")
)

// Tmux namespace for tmux session management
type Tmux mg.Namespace

// getSupportedModels returns the map of supported Claude model aliases
func getSupportedModels() map[string]string {
	return map[string]string{
		// Anthropic
		"opus":   "claude-opus-4-5",
		"sonnet": "claude-sonnet-4-5",
		"haiku":  "claude-haiku-4-5",
		// OpenAI
		"gpt": "gpt-4o",
		"o1":  "o1",
		"o3":  "o3-mini",
		// Google
		"gemini": "gemini-3-flash",
	}
}

// checkTmux verifies tmux is installed and available
func checkTmux() error {
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("%w. Install from: %s", errTmuxNotFound, tmuxInstallURL)
	}
	return nil
}

// validateModel checks if the model is in the supported list and returns the full model name
func validateModel(model string) string {
	supportedModels := getSupportedModels()
	fullModel, ok := supportedModels[model]
	if !ok {
		// Allow unknown models with a warning
		utils.Warn("Unknown model '%s'. Using as-is. Supported models: %v",
			model, getSupportedModelNames())
		return model
	}
	return fullModel
}

// getSupportedModelNames returns a sorted list of supported model aliases
func getSupportedModelNames() []string {
	supportedModels := getSupportedModels()
	names := make([]string, 0, len(supportedModels))
	for name := range supportedModels {
		names = append(names, name)
	}
	return names
}

// getSessionName determines the session name from parameters or current directory
func getSessionName(params map[string]string) (string, error) {
	if name, ok := params["name"]; ok && name != "" {
		return name, nil
	}

	// Get working directory (from dir param or current)
	dir := params["dir"]
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Use basename of directory as session name
	return filepath.Base(dir), nil
}

// getExpandedDir gets and expands the directory path from params
func getExpandedDir(params map[string]string) (string, error) {
	dir := utils.GetParam(params, "dir", "")

	// If no dir provided, use current directory
	if dir == "" {
		return os.Getwd()
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(home, dir[2:])
	}

	// Make absolute
	return filepath.Abs(dir)
}

// sessionExists checks if a tmux session exists
func sessionExists(name string) bool {
	ctx := context.Background()
	// #nosec G204 -- fixed tmux command with session name parameter
	cmd := exec.CommandContext(ctx, "tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// List shows all tmux sessions with status
func (Tmux) List() error {
	utils.Header("Tmux Sessions")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Run tmux ls to get session list
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "tmux", "ls")
	output, err := cmd.CombinedOutput()
	// Handle no sessions case gracefully
	if err != nil {
		if strings.Contains(string(output), "no server running") ||
			strings.Contains(string(output), "no sessions") ||
			cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			utils.Info("No tmux sessions found")
			return nil
		}
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Parse and display sessions
	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 0 || (len(sessions) == 1 && sessions[0] == "") {
		utils.Info("No tmux sessions found")
		return nil
	}

	utils.Info("Found %d session(s):\n", len(sessions))
	for _, session := range sessions {
		if session == "" {
			continue
		}
		// tmux ls output format: "name: X windows (created TIME) [WIDTHxHEIGHT] (attached)"
		// Display as-is for now, can enhance parsing later
		fmt.Printf("  • %s\n", session)
	}

	return nil
}

// Start starts a Claude Code session in tmux
func (Tmux) Start(args ...string) error {
	utils.Header("Start Tmux Session")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Parse parameters
	params := utils.ParseParams(args)
	model := utils.GetParam(params, "model", tmuxDefaultModel)

	// Get and expand directory
	dir, err := getExpandedDir(params)
	if err != nil {
		return err
	}

	// Get session name
	sessionName, err := getSessionName(params)
	if err != nil {
		return err
	}

	// Check if session already exists
	if sessionExists(sessionName) {
		utils.Info("Session '%s' already exists. Attaching...", sessionName)
		return attachToSession(sessionName)
	}

	// Validate and get full model name
	fullModel := validateModel(model)

	// Build Claude Code command
	claudeCmd := fmt.Sprintf("%s --dangerously-skip-permissions --model %s",
		tmuxClaudeCommand, fullModel)

	utils.Info("Creating session '%s' in directory: %s", sessionName, dir)
	utils.Info("Using model: %s", fullModel)

	// Create new tmux session in detached mode first
	ctx := context.Background()
	//nolint:gosec // claudeCmd is constructed from validated inputs
	createCmd := exec.CommandContext(ctx, "tmux", "new", "-d", "-s", sessionName, "-c", dir, claudeCmd)

	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	utils.Success("Session '%s' created. Attaching...", sessionName)

	// Now attach to the session with proper TTY
	return attachToSession(sessionName)
}

// attachToSession attaches to an existing tmux session
func attachToSession(name string) error {
	ctx := context.Background()
	// #nosec G204 -- fixed tmux command with session name parameter
	cmd := exec.CommandContext(ctx, "tmux", "attach", "-t", name)

	// Connect to terminal for interactive session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach to session: %w", err)
	}

	return nil
}

// Attach attaches to an existing tmux session
func (Tmux) Attach(args ...string) error {
	utils.Header("Attach to Tmux Session")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Parse parameters
	params := utils.ParseParams(args)
	sessionName := utils.GetParam(params, "name", "")

	// Session name is required
	if sessionName == "" {
		return errInvalidSessionName
	}

	// Check if session exists
	if !sessionExists(sessionName) {
		return fmt.Errorf("%w: %s", errSessionNotFound, sessionName)
	}

	utils.Info("Attaching to session '%s'...", sessionName)
	return attachToSession(sessionName)
}

// Kill kills a specific tmux session
func (Tmux) Kill(args ...string) error {
	utils.Header("Kill Tmux Session")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Parse parameters
	params := utils.ParseParams(args)
	sessionName := utils.GetParam(params, "name", "")

	// Session name is required
	if sessionName == "" {
		return errInvalidSessionName
	}

	// Check if session exists
	if !sessionExists(sessionName) {
		return fmt.Errorf("%w: %s", errSessionNotFound, sessionName)
	}

	utils.Info("Killing session '%s'...", sessionName)

	// Kill the session
	ctx := context.Background()
	//nolint:gosec // sessionName is validated and passed safely to tmux
	cmd := exec.CommandContext(ctx, "tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill session: %w", err)
	}

	utils.Success("Session '%s' killed successfully!", sessionName)
	return nil
}

// KillAll kills all tmux sessions with confirmation
func (Tmux) KillAll() error {
	utils.Header("Kill All Tmux Sessions")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Get list of sessions
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "tmux", "ls")
	output, err := cmd.CombinedOutput()
	// Handle no sessions case gracefully
	if err != nil {
		if strings.Contains(string(output), "no server running") ||
			strings.Contains(string(output), "no sessions") ||
			cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			utils.Info("No tmux sessions to kill")
			return nil
		}
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Parse sessions
	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 0 || (len(sessions) == 1 && sessions[0] == "") {
		utils.Info("No tmux sessions to kill")
		return nil
	}

	// Show sessions that will be killed
	utils.Warn("This will kill %d tmux session(s):", len(sessions))
	for _, session := range sessions {
		if session == "" {
			continue
		}
		// Extract just the session name (before the colon)
		name := strings.Split(session, ":")[0]
		fmt.Printf("  • %s\n", name)
	}

	// Prompt for confirmation
	fmt.Print("\nContinue? (y/N): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// If no input (just Enter), treat as "N"
		response = "n"
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "y" && response != "yes" {
		utils.Info("Canceled")
		return nil
	}

	// Kill all sessions (kill-server stops tmux entirely)
	utils.Info("Killing all sessions...")
	cmd = exec.CommandContext(ctx, "tmux", "kill-server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to kill sessions: %w", err)
	}

	utils.Success("All tmux sessions killed successfully!")
	return nil
}

// Status shows session health/status
func (Tmux) Status(args ...string) error {
	utils.Header("Tmux Session Status")

	// Check if tmux is installed
	if err := checkTmux(); err != nil {
		return err
	}

	// Get list of sessions
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "tmux", "ls")
	output, err := cmd.CombinedOutput()
	// Handle no sessions case gracefully
	if err != nil {
		if strings.Contains(string(output), "no server running") ||
			strings.Contains(string(output), "no sessions") ||
			cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 1 {
			utils.Info("No tmux sessions found")
			utils.Info("\nTo start a new session:")
			utils.Info("  magex tmux:start")
			return nil
		}
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	// Parse and display sessions with enhanced status
	sessions := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(sessions) == 0 || (len(sessions) == 1 && sessions[0] == "") {
		utils.Info("No tmux sessions found")
		return nil
	}

	utils.Info("Found %d session(s):\n", len(sessions))
	for _, session := range sessions {
		if session == "" {
			continue
		}

		// Parse session info
		// Format: "name: X windows (created TIME) [WIDTHxHEIGHT] (attached)"
		parts := strings.Split(session, ":")
		if len(parts) < 2 {
			fmt.Printf("  • %s\n", session)
			continue
		}

		name := strings.TrimSpace(parts[0])
		info := strings.TrimSpace(parts[1])

		// Check if attached
		attached := strings.Contains(info, "(attached)")
		status := "🟢 Running"
		if attached {
			status = "🔵 Attached"
		}

		// Extract window count
		windowCount := "?"
		if idx := strings.Index(info, " window"); idx > 0 {
			// Find the number before " window"
			numStart := strings.LastIndex(info[:idx], " ")
			if numStart >= 0 {
				windowCount = strings.TrimSpace(info[numStart:idx])
			}
		}

		fmt.Printf("  • %s  [%s]  %s window(s)\n", name, status, windowCount)
	}

	return nil
}
