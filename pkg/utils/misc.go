package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// errParallelExecution is returned when parallel execution fails
var errParallelExecution = errors.New("parallel execution failed")

// Parallel runs functions in parallel
func Parallel(fns ...func() error) error {
	type result struct {
		err error
	}

	ch := make(chan result, len(fns))

	for _, fn := range fns {
		go func(f func() error) {
			ch <- result{err: f()}
		}(fn)
	}

	var errs []error
	for range fns {
		res := <-ch
		if res.err != nil {
			errs = append(errs, res.err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %v", errParallelExecution, errs)
	}

	return nil
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	return formatDuration(d) // Use the shared implementation from logger.go
}

// FormatBytes formats bytes in a human-readable way
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// promptInput is the io.Reader used by PromptForInput. Tests may swap this via
// SetPromptInput; production code resolves to os.Stdin at call time so existing
// tests that pipe os.Stdin keep working when promptInput is nil.
//
//nolint:gochecknoglobals // package-level seam, same pattern as packageCommandRunnerProvider
var promptInput io.Reader

// SetPromptInput swaps the input source used by PromptForInput. Returns the
// previous reader. Pass nil to restore the os.Stdin default. Intended for tests.
func SetPromptInput(r io.Reader) io.Reader {
	prev := promptInput
	promptInput = r
	return prev
}

// ResetPromptInput restores the os.Stdin default.
func ResetPromptInput() { promptInput = nil }

// PromptForInput prompts the user for input and returns the response
func PromptForInput(prompt string) (string, error) {
	if prompt != "" {
		fmt.Printf("%s: ", prompt)
	}

	src := promptInput
	if src == nil {
		src = os.Stdin
	}

	scanner := bufio.NewScanner(src)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", nil // EOF
	}

	return strings.TrimSpace(scanner.Text()), nil
}

// defaultReadBufferSize is the buffer size for reading files (128KB)
const defaultReadBufferSize = 128 * 1024

// CheckFileLineLength checks if any line in a file exceeds maxLen bytes
// Returns: (hasLongLines, lineNumber, lineLength, error)
// Uses bufio.Reader with a large buffer to avoid token size limits
func CheckFileLineLength(path string, maxLen int) (bool, int, int, error) {
	file, err := os.Open(path) // #nosec G304 -- path is validated before use
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Best-effort close, errors not actionable in validation context
	}()

	// Use a reader with a large buffer to handle lines larger than default
	reader := bufio.NewReaderSize(file, defaultReadBufferSize)
	lineNum := 0
	maxLineLen := 0

	for {
		lineNum++
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// If we get "token too long" or similar, report what we know
			return true, lineNum, maxLineLen, nil
		}

		lineLen := len(line)

		// If isPrefix is true, the line is longer than our buffer
		// Keep reading until we get the full line
		for isPrefix {
			var part []byte
			part, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
			lineLen += len(part)
		}

		if lineLen > maxLineLen {
			maxLineLen = lineLen
		}

		// Early return if we find a line exceeding the limit
		if lineLen > maxLen {
			return true, lineNum, lineLen, nil
		}
	}

	return false, 0, maxLineLen, nil
}
