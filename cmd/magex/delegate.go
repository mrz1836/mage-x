package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/mrz1836/mage-x/pkg/utils"
)

const (
	magefileFilename = "magefile.go"
	magefilesDirname = "magefiles"
	// DefaultDelegateTimeout is the default timeout for delegated mage commands.
	// A value of 0 means no timeout — the command runs until completion or signal.
	DefaultDelegateTimeout = 0 * time.Second
	// stderrDrainDelay is how long a delegated command's stderr may stay open after
	// the command exits, which happens when it leaves a background process running.
	stderrDrainDelay = 2 * time.Second
	// stderrPartialDelay is how long a delegated command's stderr must go quiet
	// before a line without its newline yet, such as a prompt, is displayed.
	stderrPartialDelay = 150 * time.Millisecond
	// interruptKillDelay is how long a delegated command may keep running after
	// magex is interrupted before it is killed. mage gives a target 5 seconds to
	// clean up after Ctrl-C, so this leaves room for that.
	interruptKillDelay = 10 * time.Second
	// unknownTargetMessage starts the line mage prints for a target it can't find.
	unknownTargetMessage = "Unknown target specified:"
)

var (
	// ErrCommandNotFound is returned when command is not found and no magefile.go exists
	ErrCommandNotFound = errors.New("command not found and no magefile.go exists")
	// ErrGoCommandNotFound is returned when go command is not available for delegation
	ErrGoCommandNotFound = errors.New("go command not found - required for custom magefile.go commands")
	// ErrCommandFailed is returned when a custom command execution fails
	ErrCommandFailed = errors.New("custom command failed")
	// ErrMagefileRestoreFailed is returned when the temporary magefile.go cannot be restored
	ErrMagefileRestoreFailed = errors.New("failed to restore magefile.go after command execution")
	// ErrCommandTimeout is returned when a delegated command times out
	ErrCommandTimeout = errors.New("command timed out")
	// ErrCommandCanceled is returned when magex is interrupted (Ctrl-C or SIGTERM)
	// while a custom command runs, however the command then exits
	ErrCommandCanceled = errors.New("custom command canceled")
)

// stopReason records why a delegated command was stopped before it finished.
type stopReason int

const (
	// stopNone means the command ran until it exited on its own.
	stopNone stopReason = iota
	// stopCanceled means the parent context was canceled: magex got Ctrl-C or SIGTERM.
	stopCanceled
	// stopTimedOut means the -t timeout expired.
	stopTimedOut
)

// delegateOptions holds the settings of a delegated run that tests shorten or
// replace.
type delegateOptions struct {
	// killDelay is how long the command may keep running after the parent
	// context is canceled before it is killed.
	killDelay time.Duration
	// processes lists the running processes' PIDs keyed by their parent's PID,
	// to reach the processes the command started. nil means processChildren.
	processes func() (map[int][]int, error)
	// notices receives what magex itself reports about the run, such as that it
	// killed the command. nil means os.Stderr.
	notices io.Writer
}

// noticeWriter returns where magex reports on the run.
func (o delegateOptions) noticeWriter() io.Writer {
	if o.notices == nil {
		return os.Stderr
	}
	return o.notices
}

// processTree lists the running processes with o.processes, or with
// processChildren if it is nil. A failure gives an empty list: reaching the
// command's own processes is a best effort.
func (o delegateOptions) processTree() map[int][]int {
	list := o.processes
	if list == nil {
		list = processChildren
	}
	children, err := list()
	if err != nil {
		return nil
	}
	return children
}

// DelegateResult contains the result of a delegated mage command execution
type DelegateResult struct {
	// ExitCode is the exit code from the delegated command (0 for success)
	ExitCode int
	// Err contains any error that occurred during execution
	Err error
}

// convertToMageFormat converts command names to mage's expected format
// Mage expects colon-separated format for namespaced commands (e.g., ci:static)
// and performs case-insensitive lookup, so we just lowercase the entire command
func convertToMageFormat(command string) string {
	return strings.ToLower(command)
}

// DelegateToMage executes a custom command using mage or go run.
// Returns DelegateResult with the exit code and any error.
// This provides seamless execution of user-defined commands without plugin compilation.
func DelegateToMage(ctx context.Context, command string, args ...string) (result DelegateResult) {
	return DelegateToMageWithTimeout(ctx, command, DefaultDelegateTimeout, args...)
}

// DelegateToMageWithTimeout executes a custom command with a configurable timeout.
// Returns DelegateResult with the exit code and any error.
//
// Canceling ctx (magex got Ctrl-C or SIGTERM) lets the command stop on its own:
// it gets os.Interrupt, and is killed only if it is still running 10 seconds
// later. After a SIGTERM, the processes the command started, such as the
// compiled magefile, get an interrupt too, since mage passes nothing on. The
// result is then ErrCommandCanceled, however the command exited. A timeout
// greater than 0 is a hard limit: the command is killed as soon as it expires,
// and the result is ErrCommandTimeout with exit code 124. Killing the command
// kills the processes it started too.
func DelegateToMageWithTimeout(ctx context.Context, command string, timeout time.Duration, args ...string) DelegateResult {
	return delegateToMage(ctx, command, timeout, delegateOptions{killDelay: interruptKillDelay}, args...)
}

// delegateToMage is DelegateToMageWithTimeout with the settings tests override.
func delegateToMage(ctx context.Context, command string, timeout time.Duration, opts delegateOptions, args ...string) (result DelegateResult) {
	// Convert colon-separated format to mage's camelCase format
	mageCommand := convertToMageFormat(command)

	// Check for magefiles/ directory first (preferred by standard mage)
	magefilesDir := magefilesDirname
	var targetPath string
	var useDirectory bool

	if info, err := os.Stat(magefilesDir); err == nil && info.IsDir() {
		targetPath = magefilesDir
		useDirectory = true
	} else {
		// Fallback to root magefile.go
		magefilePath := magefileFilename
		if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
			return DelegateResult{ExitCode: 1, Err: fmt.Errorf("%w: %s", ErrCommandNotFound, command)}
		}
		targetPath = magefilePath
		useDirectory = false
	}

	var cmd *exec.Cmd

	// Check if we have both magefiles/ directory and magefile.go (conflict situation)
	hasRootMagefile := false
	if _, err := os.Stat(magefileFilename); err == nil {
		hasRootMagefile = true
	}
	hasConflict := useDirectory && hasRootMagefile

	// Handle conflict by temporarily renaming magefile.go
	if hasConflict {
		tempName := magefileFilename + ".tmp"
		if err := os.Rename(magefileFilename, tempName); err != nil {
			return DelegateResult{ExitCode: 1, Err: fmt.Errorf("failed to temporarily rename magefile.go: %w", err)}
		}
		restore := &magefileRestore{tempName: tempName}
		// A second Ctrl-C force-exits magex, which skips the deferred restore
		// below, so the forced exit restores magefile.go too
		removeHook := forceExitHooksFrom(ctx).add(restore.onForcedExit)
		// Use named return to properly capture restore errors
		defer func() {
			restoreErr := restore.restore()
			removeHook()
			if restoreErr != nil {
				// Combine errors if command also failed
				if result.Err != nil {
					result.Err = fmt.Errorf("%w; additionally, %w: %w", result.Err, ErrMagefileRestoreFailed, restoreErr)
				} else {
					// Command succeeded but restore failed - this is still an error
					result.Err = fmt.Errorf("%w: %w", ErrMagefileRestoreFailed, restoreErr)
					result.ExitCode = 1
				}
			}
		}()
	}

	// The commands below are not tied to ctx: exec.CommandContext would kill
	// them the moment it is canceled, giving the target no chance to clean up.
	// waitDelegated stops them instead.

	// Try to use mage first if available
	if magePath, err := exec.LookPath("mage"); err == nil {
		// Use mage binary - mage handles both directory and file automatically
		// NOTE: mage binary does NOT support command-line arguments for custom functions
		// Arguments must be passed via environment variables (MAGE_ARGS, MAGE_ARGS_JSON)
		cmdArgs := []string{mageCommand}
		// #nosec G204,G702 -- This is necessary for dynamic command execution with user-defined commands
		cmd = exec.Command(magePath, cmdArgs...) //nolint:noctx // waitDelegated stops it gracefully; CommandContext would kill it
	} else {
		// Fallback to go run with mage tags
		cmdArgs := make([]string, 0, 4+len(args))
		if useDirectory {
			// For directory, we need to run from within the directory using go run .
			cmdArgs = append(cmdArgs, "run", "-tags=mage", ".", mageCommand)
			cmdArgs = append(cmdArgs, args...)
			// #nosec G204,G702 -- This is necessary for dynamic command execution with user-defined commands
			cmd = exec.Command("go", cmdArgs...) //nolint:noctx // waitDelegated stops it gracefully; CommandContext would kill it
			// Set the working directory to the magefiles directory
			cmd.Dir = targetPath
		} else {
			// For single file, specify the file path
			cmdArgs = append(cmdArgs, "run", "-tags=mage", targetPath, mageCommand)
			cmdArgs = append(cmdArgs, args...)
			// #nosec G204,G702 -- This is necessary for dynamic command execution with user-defined commands
			cmd = exec.Command("go", cmdArgs...) //nolint:noctx // waitDelegated stops it gracefully; CommandContext would kill it
		}
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Make arguments available via environment variables for magefile functions to
	// access (utils.MageArgs reads them): MAGE_ARGS joins them with spaces, and
	// MAGE_ARGS_JSON keeps arguments that contain spaces intact
	if len(args) > 0 {
		cmd.Env = append(cmd.Env, utils.EnvMageArgs+"="+strings.Join(args, " "))
		if encoded, err := json.Marshal(args); err == nil {
			cmd.Env = append(cmd.Env, utils.EnvMageArgsJSON+"="+string(encoded))
		}
	}

	// Set working directory if not already set (directory case sets it specifically)
	if cmd.Dir == "" {
		if dir, err := os.Getwd(); err == nil {
			cmd.Dir = dir
		}
	}

	// Set up stdio with error filtering for stderr. os/exec owns the stderr pipe:
	// Wait returns once the filter has seen all of it, unlike cmd.StderrPipe, which
	// Wait closes as soon as the process exits, dropping output not yet read.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	stderr := &stderrFilter{out: os.Stderr}
	cmd.Stderr = stderr
	// A background process the command started may keep stderr open; stop waiting
	// for it once the command itself has exited and this delay has passed.
	cmd.WaitDelay = stderrDrainDelay

	// Don't start the command once magex is already shutting down
	if ctx.Err() != nil {
		return DelegateResult{ExitCode: exitCodeSIGINT, Err: fmt.Errorf("%w '%s'", ErrCommandCanceled, command)}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return DelegateResult{ExitCode: 1, Err: fmt.Errorf("failed to start custom command '%s': %w", command, err)}
	}

	// Wait for the command to finish and its stderr to be fully captured
	stopped, waitErr := waitDelegated(ctx, cmd, command, timeout, opts)
	stderr.Flush()

	// ErrWaitDelay means the command succeeded but something it left running still
	// held stderr; that is not a failure of the command.
	if errors.Is(waitErr, exec.ErrWaitDelay) {
		waitErr = nil
	}

	// Interrupted: magex already said so, and however the command exited, even
	// killed after ignoring the interrupt, was its response, not a failure
	if stopped == stopCanceled {
		if waitErr != nil {
			return DelegateResult{ExitCode: exitCodeSIGINT, Err: fmt.Errorf("%w '%s': %w", ErrCommandCanceled, command, waitErr)}
		}
		return DelegateResult{ExitCode: exitCodeSIGINT, Err: fmt.Errorf("%w '%s'", ErrCommandCanceled, command)}
	}

	if waitErr != nil {
		// Extract actual exit code from the error
		exitCode := 1
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}

		// Killed because the timeout expired
		if stopped == stopTimedOut {
			return DelegateResult{
				ExitCode: 124, // Standard timeout exit code
				Err:      fmt.Errorf("%w: '%s' after %v", ErrCommandTimeout, command, timeout),
			}
		}

		// Stderr was shown as it arrived, so the error only adds what was held back
		if hidden := strings.TrimSpace(stderr.Hidden()); hidden != "" {
			return DelegateResult{
				ExitCode: exitCode,
				Err:      fmt.Errorf("%w '%s':\n%s", ErrCommandFailed, command, hidden),
			}
		}
		return DelegateResult{
			ExitCode: exitCode,
			Err:      fmt.Errorf("%w '%s': %w", ErrCommandFailed, command, waitErr),
		}
	}

	return DelegateResult{ExitCode: 0, Err: nil}
}

// waitDelegated waits for cmd, which has been started, and returns why it was
// stopped early, if it was, along with Wait's error.
//
// Canceling ctx means magex got Ctrl-C or SIGTERM. A terminal's Ctrl-C has
// already reached the command and the target it runs, since they share the
// foreground process group, and mage gives the target 5 seconds to clean up.
// So cmd gets os.Interrupt, and is killed only if it is still running after
// opts.killDelay, with a notice naming command. A timeout greater than 0 is a
// hard limit: cmd is killed as soon as it expires.
func waitDelegated(ctx context.Context, cmd *exec.Cmd, command string, timeout time.Duration, opts delegateOptions) (stopReason, error) {
	var timeoutC <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timeoutC = timer.C
	}

	exited := make(chan struct{})
	stopped := make(chan stopReason, 1)
	go func() {
		stopped <- stopDelegated(ctx, cmd.Process, command, opts, timeoutC, exited)
	}()

	err := cmd.Wait()
	close(exited)
	return <-stopped, err
}

// stopDelegated watches a running command for waitDelegated until exited
// closes, stopping the command if ctx is canceled or timeoutC fires first.
// command names it in the notice of a kill after an interrupt.
func stopDelegated(ctx context.Context, proc *os.Process, command string, opts delegateOptions, timeoutC <-chan time.Time, exited <-chan struct{}) stopReason {
	select {
	case <-exited:
		return stopNone
	case <-timeoutC:
		killDelegated(proc, opts)
		return stopTimedOut
	case <-ctx.Done():
	}

	// A SIGTERM was sent to magex alone: mage doesn't catch SIGTERM and passes
	// nothing on, so the compiled magefile it runs never hears of it. Give that
	// process the interrupt the magefile's generated main handles, as Ctrl-C
	// would. After a SIGINT nothing more is sent: a terminal's Ctrl-C has already
	// reached it, and a second interrupt makes mage give up the cleanup at once.
	// The command's processes are listed before it is signaled, in case that
	// ends it and they are reparented out of reach.
	var children []int
	if canceledBySignal(ctx, syscall.SIGTERM) {
		children = opts.processTree()[proc.Pid]
	}

	// Best effort: this fails on Windows, and mage ignores it while the target runs
	_ = proc.Signal(os.Interrupt) //nolint:errcheck // a command that missed it is killed after killDelay
	signalPIDs(children, os.Interrupt)

	grace := time.NewTimer(opts.killDelay)
	defer grace.Stop()
	select {
	case <-exited:
	case <-grace.C:
		killDelegated(proc, opts)
		// Say so: the target's own output stops mid-cleanup, and only the exit
		// status would tell. The timeout below reports itself as an error.
		_, _ = fmt.Fprintf(opts.noticeWriter(), "⚠  Custom command '%s' was still running %s after the interrupt — killed it.\n", command, opts.killDelay) //nolint:errcheck // notice on the shutdown path; nothing to do if the write fails
	case <-timeoutC:
		killDelegated(proc, opts)
	}
	return stopCanceled
}

// killDelegated kills a delegated command that would not stop on its own, and
// every process it started, so that none is left running without it, such as
// the compiled magefile mage runs. Those processes are listed before the
// command is killed, while they are still its own.
func killDelegated(proc *os.Process, opts delegateOptions) {
	descendants := descendantPIDs(opts.processTree(), proc.Pid)
	signalPIDs(descendants, os.Kill)
	_ = proc.Kill() //nolint:errcheck // it fails only if the command has already exited, which is the goal
}

// magefileRestore moves magefile.go back after it was renamed out of the way
// so that mage builds only magefiles/. Both the deferred restore and a forced
// exit may run it, even at the same time; the rename happens once, and every
// caller gets its result.
type magefileRestore struct {
	tempName string

	once sync.Once
	err  error
}

// restore renames tempName back to magefile.go, the first time it is called.
func (r *magefileRestore) restore() error {
	r.once.Do(func() {
		r.err = os.Rename(r.tempName, magefileFilename)
	})
	return r.err
}

// onForcedExit restores magefile.go before magex force-exits, reporting a
// failure since nothing else will.
func (r *magefileRestore) onForcedExit() {
	if err := r.restore(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v: %v\n", ErrMagefileRestoreFailed, err)
	}
}

// stderrFilter receives a delegated command's stderr and passes it through to
// out: each line as it completes, and a partial line, such as a prompt with no
// trailing newline, once stderrPartialDelay passes without more output. Lines
// reporting an unknown target are held back from the display and kept for the
// error message instead. os/exec calls Write from one goroutine, but the
// partial-line timer runs on another, so mu guards the fields.
type stderrFilter struct {
	out io.Writer

	mu      sync.Mutex
	hidden  strings.Builder // lines held back from the display, for the error message
	partial []byte          // the start of a line whose '\n' hasn't arrived yet
	shown   int             // how much of partial has been displayed already
	timer   *time.Timer     // displays partial once the output goes quiet
}

// Write handles each complete line in p, and schedules a trailing partial line
// to be displayed if nothing follows it.
func (f *stderrFilter) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.partial = append(f.partial, p...)
	for {
		i := bytes.IndexByte(f.partial, '\n')
		if i < 0 {
			break
		}
		f.line(f.partial[:i])
		f.partial = f.partial[i+1:]
		f.shown = 0
	}

	if len(f.partial) > f.shown {
		if f.timer == nil {
			f.timer = time.AfterFunc(stderrPartialDelay, f.showPartial)
		} else {
			f.timer.Reset(stderrPartialDelay)
		}
	}

	return len(p), nil
}

// Flush handles a final line that ended without a '\n'.
func (f *stderrFilter) Flush() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.timer != nil {
		f.timer.Stop()
	}
	if len(f.partial) > 0 {
		f.line(f.partial)
		f.partial, f.shown = nil, 0
	}
}

// Hidden returns the lines held back from the display.
func (f *stderrFilter) Hidden() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.hidden.String()
}

// showPartial displays the part of the pending line not shown yet. A line that
// may still turn out to report an unknown target waits for its newline.
func (f *stderrFilter) showPartial() {
	f.mu.Lock()
	defer f.mu.Unlock()

	pending := f.partial[f.shown:]
	if len(pending) == 0 || (f.shown == 0 && mayReportUnknownTarget(string(f.partial))) {
		return
	}
	_, _ = f.out.Write(pending) //nolint:errcheck // #nosec G705 -- intentionally printing subprocess stderr to user; nothing to do if it fails
	f.shown = len(f.partial)
}

// line handles a complete line (f.mu held), displaying whatever part of it
// showPartial hasn't already.
func (f *stderrFilter) line(raw []byte) {
	line := strings.TrimSuffix(string(raw), "\r")

	// mage's "Unknown target specified" goes in the error message instead of the display
	if f.shown == 0 && strings.Contains(line, unknownTargetMessage) {
		f.hidden.WriteString(line)
		f.hidden.WriteString("\n")
		return
	}

	rest := line[min(f.shown, len(line)):]
	_, _ = fmt.Fprintln(f.out, rest) //nolint:errcheck // #nosec G705 -- intentionally printing subprocess stderr to user; nothing to do if it fails
}

// mayReportUnknownTarget reports whether an incomplete line is, or may still
// become, mage's unknown-target message.
func mayReportUnknownTarget(partial string) bool {
	return strings.Contains(partial, unknownTargetMessage) || strings.HasPrefix(unknownTargetMessage, partial)
}

// HasMagefile checks if magefiles/ directory or magefile.go exists in the current directory
func HasMagefile() bool {
	return GetMagefilePath() != ""
}

// GetMagefilePath returns the path to magefiles/ directory or magefile.go if it exists
func GetMagefilePath() string {
	// Check for magefiles/ directory first
	if info, err := os.Stat(magefilesDirname); err == nil && info.IsDir() {
		if abs, err := filepath.Abs(magefilesDirname); err == nil {
			return abs
		}
		return magefilesDirname
	}
	// Fallback to magefile.go
	if _, err := os.Stat(magefileFilename); err == nil {
		if abs, err := filepath.Abs("magefile.go"); err == nil {
			return abs
		}
		return magefileFilename
	}
	return ""
}

// ValidateGoEnvironment checks if Go is available for delegation
func ValidateGoEnvironment() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("%w", ErrGoCommandNotFound)
	}
	return nil
}
