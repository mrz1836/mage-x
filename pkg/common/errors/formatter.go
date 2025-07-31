package errors

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// DefaultErrorFormatter is the default implementation of ErrorFormatter
type DefaultErrorFormatter struct {
	defaultOptions FormatOptions
}

// NewFormatter creates a new error formatter
func NewFormatter() *DefaultErrorFormatter {
	return &DefaultErrorFormatter{
		defaultOptions: FormatOptions{
			IncludeStack:     false,
			IncludeContext:   true,
			IncludeCause:     true,
			IncludeTimestamp: true,
			IncludeFields:    true,
			IndentLevel:      2,
			MaxDepth:         10,
			TimeFormat:       time.RFC3339,
			FieldsSeparator:  ", ",
			UseColor:         false,
			CompactMode:      false,
		},
	}
}

// Format formats an error
func (f *DefaultErrorFormatter) Format(err error) string {
	return f.FormatWithOptions(err, f.defaultOptions)
}

// FormatMageError formats a MageError
func (f *DefaultErrorFormatter) FormatMageError(err MageError) string {
	opts := f.defaultOptions
	opts.IncludeStack = err.Severity() >= SeverityError
	return f.formatMageError(err, opts, 0)
}

// FormatChain formats an error chain
func (f *DefaultErrorFormatter) FormatChain(chain ErrorChain) string {
	if chain.Count() == 0 {
		return "no errors in chain"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Error chain (%d errors):\n", chain.Count()))

	for i, err := range chain.Errors() {
		sb.WriteString(fmt.Sprintf("\n[%d] ", i+1))
		var mageErr MageError
		if errors.As(err, &mageErr) {
			sb.WriteString(f.formatMageError(mageErr, f.defaultOptions, 1))
		} else {
			sb.WriteString(f.formatGenericError(err, f.defaultOptions, 1))
		}
	}

	return sb.String()
}

// FormatWithOptions formats an error with specific options
func (f *DefaultErrorFormatter) FormatWithOptions(err error, opts FormatOptions) string {
	if err == nil {
		return "<nil>"
	}

	// Handle MageError
	var mageErr MageError
	if errors.As(err, &mageErr) {
		return f.formatMageError(mageErr, opts, 0)
	}

	// Handle ErrorChain
	var chain ErrorChain
	if errors.As(err, &chain) {
		return f.FormatChain(chain)
	}

	// Handle generic error
	return f.formatGenericError(err, opts, 0)
}

// formatMageError formats a MageError with indentation
func (f *DefaultErrorFormatter) formatMageError(err MageError, opts FormatOptions, depth int) string {
	if depth > opts.MaxDepth {
		return f.indent("...(max depth reached)", depth)
	}

	var sb strings.Builder

	if opts.CompactMode {
		// Compact single-line format
		sb.WriteString(fmt.Sprintf("[%s/%s] %s", err.Code(), err.Severity(), err.Error()))

		if opts.IncludeFields && len(err.Context().Fields) > 0 {
			fields := f.formatFields(err.Context().Fields, opts)
			sb.WriteString(fmt.Sprintf(" {%s}", fields))
		}

		return sb.String()
	}

	// Full format
	indent := f.getIndent(depth)

	// Header line
	if opts.UseColor {
		sb.WriteString(f.colorize(err.Severity(), fmt.Sprintf("%s[%s] %s",
			indent, err.Code(), err.Error())))
	} else {
		sb.WriteString(fmt.Sprintf("%s[%s] %s", indent, err.Code(), err.Error()))
	}

	// Severity
	sb.WriteString(fmt.Sprintf("\n%s  Severity: %s", indent, err.Severity()))

	// Timestamp
	if opts.IncludeTimestamp && !err.Context().Timestamp.IsZero() {
		sb.WriteString(fmt.Sprintf("\n%s  Time: %s", indent,
			err.Context().Timestamp.Format(opts.TimeFormat)))
	}

	// Context
	if opts.IncludeContext {
		ctx := err.Context()
		f.formatContext(&sb, &ctx, indent)
	}

	// Fields
	if opts.IncludeFields && len(err.Context().Fields) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s  Fields:", indent))
		for k, v := range err.Context().Fields {
			sb.WriteString(fmt.Sprintf("\n%s    %s: %v", indent, k, v))
		}
	}

	// Cause
	if opts.IncludeCause && err.Cause() != nil {
		sb.WriteString(fmt.Sprintf("\n%s  Caused by:", indent))
		var causeMageErr MageError
		if errors.As(err.Cause(), &causeMageErr) {
			sb.WriteString("\n")
			sb.WriteString(f.formatMageError(causeMageErr, opts, depth+1))
		} else {
			sb.WriteString(fmt.Sprintf("\n%s    %v", indent, err.Cause()))
		}
	}

	// Stack trace
	if opts.IncludeStack && err.Context().StackTrace != "" {
		sb.WriteString(fmt.Sprintf("\n%s  Stack trace:", indent))
		stackLines := strings.Split(err.Context().StackTrace, "\n")
		for _, line := range stackLines {
			if strings.TrimSpace(line) != "" {
				sb.WriteString(fmt.Sprintf("\n%s  %s", indent, line))
			}
		}
	}

	return sb.String()
}

// formatGenericError formats a generic error
func (f *DefaultErrorFormatter) formatGenericError(err error, opts FormatOptions, depth int) string {
	if depth > opts.MaxDepth {
		return f.indent("...(max depth reached)", depth)
	}

	indent := f.getIndent(depth)

	if opts.CompactMode {
		return err.Error()
	}

	return fmt.Sprintf("%s%v", indent, err)
}

// formatContext formats context information with guard clauses to reduce nesting
func (f *DefaultErrorFormatter) formatContext(sb *strings.Builder, ctx *ErrorContext, indent string) {
	if ctx.Operation != "" {
		fmt.Fprintf(sb, "\n%s  Operation: %s", indent, ctx.Operation)
	}
	if ctx.Resource != "" {
		fmt.Fprintf(sb, "\n%s  Resource: %s", indent, ctx.Resource)
	}
	if ctx.User != "" {
		fmt.Fprintf(sb, "\n%s  User: %s", indent, ctx.User)
	}
	if ctx.RequestID != "" {
		fmt.Fprintf(sb, "\n%s  RequestID: %s", indent, ctx.RequestID)
	}
	if ctx.Environment != "" {
		fmt.Fprintf(sb, "\n%s  Environment: %s", indent, ctx.Environment)
	}
}

// formatFields formats context fields
func (f *DefaultErrorFormatter) formatFields(fields map[string]interface{}, opts FormatOptions) string {
	if len(fields) == 0 {
		return ""
	}

	parts := make([]string, 0, len(fields))
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(parts, opts.FieldsSeparator)
}

// getIndent returns the indentation string for a given depth
func (f *DefaultErrorFormatter) getIndent(depth int) string {
	return strings.Repeat(" ", depth*f.defaultOptions.IndentLevel)
}

// indent adds indentation to a string
func (f *DefaultErrorFormatter) indent(s string, depth int) string {
	indent := f.getIndent(depth)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// colorize adds color codes based on severity
func (f *DefaultErrorFormatter) colorize(severity Severity, text string) string {
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorGray   = "\033[90m"
	)

	var color string
	switch severity {
	case SeverityDebug:
		color = colorGray
	case SeverityInfo:
		color = colorCyan
	case SeverityWarning:
		color = colorYellow
	case SeverityError:
		color = colorRed
	case SeverityCritical:
		color = colorPurple
	case SeverityFatal:
		color = colorRed
	default:
		color = colorReset
	}

	return fmt.Sprintf("%s%s%s", color, text, colorReset)
}

// SetDefaultOptions sets the default formatting options
func (f *DefaultErrorFormatter) SetDefaultOptions(opts FormatOptions) {
	f.defaultOptions = opts
}

// GetDefaultOptions returns the current default formatting options
func (f *DefaultErrorFormatter) GetDefaultOptions() FormatOptions {
	return f.defaultOptions
}
