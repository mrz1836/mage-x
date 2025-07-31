// Package utils provides utility functions for mage tasks
package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// LogLevelDebug is for detailed debugging information
	LogLevelDebug LogLevel = iota
	// LogLevelInfo is for general informational messages
	LogLevelInfo
	// LogLevelWarn is for warning messages
	LogLevelWarn
	// LogLevelError is for error messages
	LogLevelError
)

// Logger provides native logging with colors and formatting
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	prefix   string
	useColor bool
	output   io.Writer
	spinner  *Spinner
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// Emoji constants for different message types
const (
	emojiSuccess = "✅"
	emojiError   = "❌"
	emojiWarning = "⚠️"
	emojiInfo    = "ℹ️"
	emojiDebug   = "🔍"
	emojiRocket  = "🚀"
	emojiBuild   = "🔨"
	emojiTest    = "🧪"
	emojiClean   = "🧹"
	emojiPackage = "📦"
	emojiLink    = "🔗"
	emojiArt     = "🎨"
	emojiCoffee  = "☕"
	emojiParty   = "🎉"
	emojiTarget  = "🎯"
	emojiBulb    = "💡"
	emojiBook    = "📚"
	emojiChart   = "📊"
	emojiShield  = "🛡️"
	emojiClock   = "⏱️"
)

var (
	// DefaultLogger is the global logger instance
	DefaultLogger = NewLogger() //nolint:gochecknoglobals // Package-level default

	// Contextual messages based on time and state
	contextualMessages = map[string][]string{ //nolint:gochecknoglobals // Package-level configuration
		"morning": {
			"☕ Time to build something great!",
			"🌅 Fresh build, fresh start!",
			"☕ Good morning! Let's ship some code!",
		},
		"afternoon": {
			"🚀 Afternoon productivity boost!",
			"💪 Keep pushing forward!",
			"🔥 Let's make progress!",
		},
		"evening": {
			"🌙 Burning the midnight oil!",
			"✨ Evening coding session!",
			"🌃 Night owl mode activated!",
		},
		"friday": {
			"🎉 Ship it before the weekend!",
			"📦 Feature Friday!",
			"🚀 Friday deployment time!",
		},
		"monday": {
			"💪 Monday motivation!",
			"🚀 Fresh week, fresh code!",
			"☕ Monday morning build!",
		},
		"fast": {
			"⚡ Blazing fast build!",
			"🏎️ Speed demon!",
			"🚄 Express build complete!",
		},
		"slow": {
			"🐌 Taking our time...",
			"⏳ Good things take time...",
			"🧘 Patience is a virtue...",
		},
		"success": {
			"✨ All green! You're a wizard!",
			"🎯 Nailed it!",
			"🎉 Success! High five!",
			"💯 Perfect execution!",
		},
		"fixed": {
			"🔧 Fixed! Back in business!",
			"✨ Problem solved!",
			"💪 Bug squashed!",
		},
	}
)

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		level:    LogLevelInfo,
		useColor: shouldUseColor(),
		output:   os.Stdout,
	}
}

// WithPrefix creates a new logger with a prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		level:    l.level,
		prefix:   prefix,
		useColor: l.useColor,
		output:   l.output,
	}
	return newLogger
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetColorEnabled enables or disables color output
func (l *Logger) SetColorEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.useColor = enabled
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// Success logs a success message with emoji
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.logWithEmoji(LogLevelInfo, emojiSuccess, msg)
}

// Fail logs a failure message with emoji
func (l *Logger) Fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.logWithEmoji(LogLevelError, emojiError, msg)
}

// Header prints a formatted header
func (l *Logger) Header(text string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner != nil {
		l.spinner.Stop()
	}

	line := strings.Repeat("=", 60)
	if l.useColor {
		if _, err := fmt.Fprintf(l.output, "\n%s%s%s\n", colorBold+colorBlue, line, colorReset); err != nil {
			// Continue if write fails
			log.Printf("failed to write separator header: %v", err)
		}
		if _, err := fmt.Fprintf(l.output, "%s%s %s %s%s\n", colorBold+colorBlue, "===", text, "===", colorReset); err != nil {
			// Continue if write fails
			log.Printf("failed to write separator text: %v", err)
		}
		if _, err := fmt.Fprintf(l.output, "%s%s%s\n\n", colorBold+colorBlue, line, colorReset); err != nil {
			// Continue if write fails
			log.Printf("failed to write separator footer: %v", err)
		}
	} else {
		if _, err := fmt.Fprintf(l.output, "\n%s\n=== %s ===\n%s\n\n", line, text, line); err != nil {
			// Continue if write fails
			log.Printf("failed to write plain separator: %v", err)
		}
	}
}

// StartSpinner starts a progress spinner with a message
func (l *Logger) StartSpinner(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner != nil {
		l.spinner.Stop()
	}

	l.spinner = NewSpinner(message)
	l.spinner.Start()
}

// StopSpinner stops the current spinner
func (l *Logger) StopSpinner() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner != nil {
		l.spinner.Stop()
		l.spinner = nil
	}
}

// UpdateSpinner updates the spinner message
func (l *Logger) UpdateSpinner(message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.spinner != nil {
		l.spinner.UpdateMessage(message)
	}
}

// GetContextualMessage returns a contextual message based on current time/state
func (l *Logger) GetContextualMessage(context string) string {
	messages, ok := contextualMessages[context]
	if !ok {
		return ""
	}

	// Simple rotation based on current time
	index := time.Now().Unix() % int64(len(messages))
	return messages[index]
}

// GetTimeContext returns the current time context (morning, afternoon, evening)
func (l *Logger) GetTimeContext() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	default:
		return "evening"
	}
}

// GetDayContext returns the current day context (monday, friday, etc)
func (l *Logger) GetDayContext() string {
	weekday := time.Now().Weekday()
	switch weekday {
	case time.Monday:
		return "monday"
	case time.Friday:
		return "friday"
	default:
		return ""
	}
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Stop spinner while logging
	if l.spinner != nil {
		l.spinner.Pause()
		defer l.spinner.Resume()
	}

	msg := fmt.Sprintf(format, args...)

	var levelStr, color string
	switch level {
	case LogLevelDebug:
		levelStr = "DEBUG"
		color = colorGray
	case LogLevelInfo:
		levelStr = "INFO "
		color = colorBlue
	case LogLevelWarn:
		levelStr = "WARN "
		color = colorYellow
	case LogLevelError:
		levelStr = "ERROR"
		color = colorRed
	}

	timestamp := time.Now().Format("15:04:05")

	if l.prefix != "" {
		msg = fmt.Sprintf("[%s] %s", l.prefix, msg)
	}

	if l.useColor {
		if _, err := fmt.Fprintf(l.output, "%s%s [%s]%s %s\n", color, timestamp, levelStr, colorReset, msg); err != nil {
			// Continue if write fails
			log.Printf("failed to write colored log message: %v", err)
		}
	} else {
		if _, err := fmt.Fprintf(l.output, "%s [%s] %s\n", timestamp, levelStr, msg); err != nil {
			// Continue if write fails
			log.Printf("failed to write plain log message: %v", err)
		}
	}
}

// logWithEmoji logs a message with an emoji prefix
func (l *Logger) logWithEmoji(level LogLevel, emoji, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Stop spinner while logging
	if l.spinner != nil {
		l.spinner.Pause()
		defer l.spinner.Resume()
	}

	if l.prefix != "" {
		msg = fmt.Sprintf("[%s] %s", l.prefix, msg)
	}

	var color string
	switch level {
	case LogLevelInfo:
		color = colorGreen
	case LogLevelWarn:
		color = colorYellow
	case LogLevelError:
		color = colorRed
	default:
		color = colorReset
	}

	if l.useColor {
		if _, err := fmt.Fprintf(l.output, "%s %s%s%s\n", emoji, color, msg, colorReset); err != nil {
			// Continue if write fails
			log.Printf("failed to write colored emoji message: %v", err)
		}
	} else {
		if _, err := fmt.Fprintf(l.output, "%s %s\n", emoji, msg); err != nil {
			// Continue if write fails
			log.Printf("failed to write plain emoji message: %v", err)
		}
	}
}

// shouldUseColor determines if color output should be enabled
func shouldUseColor() bool {
	// Disable color in CI environments
	if os.Getenv("CI") != "" {
		return false
	}

	// Disable color if NO_COLOR is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Disable color if not a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false // Assume no color on error
	}
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Disable color on Windows unless using Windows Terminal
	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") == "" {
			return false
		}
	}

	return true
}

// Progress represents a progress bar
type Progress struct {
	mu          sync.Mutex
	total       int
	current     int
	width       int
	message     string
	startTime   time.Time
	useColor    bool
	showPercent bool
	showTime    bool
}

// NewProgress creates a new progress bar
func NewProgress(total int, message string) *Progress {
	return &Progress{
		total:       total,
		message:     message,
		width:       40,
		startTime:   time.Now(),
		useColor:    shouldUseColor(),
		showPercent: true,
		showTime:    true,
	}
}

// Update updates the progress bar
func (p *Progress) Update(current int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = current
	p.render()
}

// Increment increments the progress by 1
func (p *Progress) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current++
	p.render()
}

// Finish completes the progress bar
func (p *Progress) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current = p.total
	p.render()
	if _, err := fmt.Fprintln(os.Stdout, ""); err != nil {
		// Continue if write fails
		log.Printf("failed to write progress newline: %v", err)
	}
}

// render draws the progress bar
func (p *Progress) render() {
	if p.total <= 0 {
		return
	}

	percent := float64(p.current) / float64(p.total)
	filled := int(percent * float64(p.width))

	// Build progress bar
	bar := strings.Builder{}
	bar.WriteString("[")

	for i := 0; i < p.width; i++ {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	bar.WriteString("]")

	// Build status
	status := ""
	if p.showPercent {
		status += fmt.Sprintf(" %3.0f%%", percent*100)
	}

	if p.showTime && p.current > 0 {
		elapsed := time.Since(p.startTime)
		if p.current < p.total {
			estimated := time.Duration(float64(elapsed) / float64(p.current) * float64(p.total))
			remaining := estimated - elapsed
			status += fmt.Sprintf(" (ETA: %s)", formatDuration(remaining))
		} else {
			status += fmt.Sprintf(" (%s)", formatDuration(elapsed))
		}
	}

	// Add emoji based on completion
	var emoji string
	if p.current >= p.total {
		emoji = " ✓"
	} else if p.current > 0 {
		emoji = " ⚡"
	} else {
		emoji = " 🔄"
	}

	// Clear line and print progress
	if _, err := fmt.Fprintf(os.Stdout, "\r%-30s %s%s%s", p.message, bar.String(), status, emoji); err != nil {
		// Continue if write fails
		log.Printf("failed to write progress bar: %v", err)
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	return fmt.Sprintf("%.1fm", d.Minutes())
}

// Package-level convenience functions that use the default logger

// Debug logs a debug message using the default logger
func Debug(format string, args ...interface{}) {
	DefaultLogger.Debug(format, args...)
}

// Info logs an informational message using the default logger
func Info(format string, args ...interface{}) {
	DefaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...interface{}) {
	DefaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...interface{}) {
	DefaultLogger.Error(format, args...)
}

// Success logs a success message using the default logger
func Success(format string, args ...interface{}) {
	DefaultLogger.Success(format, args...)
}

// Fail logs a failure message using the default logger
func Fail(format string, args ...interface{}) {
	DefaultLogger.Fail(format, args...)
}

// Header prints a formatted header using the default logger
func Header(text string) {
	DefaultLogger.Header(text)
}

// StartSpinner starts a progress spinner using the default logger
func StartSpinner(message string) {
	DefaultLogger.StartSpinner(message)
}

// StopSpinner stops the current spinner using the default logger
func StopSpinner() {
	DefaultLogger.StopSpinner()
}

// UpdateSpinner updates the spinner message using the default logger
func UpdateSpinner(message string) {
	DefaultLogger.UpdateSpinner(message)
}
