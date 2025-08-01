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

	if opts.CompactMode {
		return f.formatCompactError(err, opts)
	}

	return f.formatFullError(err, opts, depth)
}

// formatCompactError formats an error in compact single-line format
func (f *DefaultErrorFormatter) formatCompactError(err MageError, opts FormatOptions) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s/%s] %s", err.Code(), err.Severity(), err.Error()))

	if opts.IncludeFields && len(err.Context().Fields) > 0 {
		fields := f.formatFields(err.Context().Fields, opts)
		sb.WriteString(fmt.Sprintf(" {%s}", fields))
	}

	return sb.String()
}

// formatFullError formats an error in full detailed format
func (f *DefaultErrorFormatter) formatFullError(err MageError, opts FormatOptions, depth int) string {
	var sb strings.Builder
	indent := f.getIndent(depth)

	f.writeErrorHeader(&sb, err, opts, indent)
	f.writeErrorDetails(&sb, err, opts, indent, depth)

	return sb.String()
}

// writeErrorHeader writes the error header line
func (f *DefaultErrorFormatter) writeErrorHeader(sb *strings.Builder, err MageError, opts FormatOptions, indent string) {
	headerText := fmt.Sprintf("%s[%s] %s", indent, err.Code(), err.Error())
	if opts.UseColor {
		sb.WriteString(f.colorize(err.Severity(), headerText))
	} else {
		sb.WriteString(headerText)
	}

	// Add severity on new line
	fmt.Fprintf(sb, "\n%s  Severity: %s", indent, err.Severity())
}

// writeErrorDetails writes all the error details
func (f *DefaultErrorFormatter) writeErrorDetails(sb *strings.Builder, err MageError, opts FormatOptions, indent string, depth int) {
	f.writeTimestamp(sb, err, opts, indent)
	f.writeContext(sb, err, opts, indent)
	f.writeFields(sb, err, opts, indent)
	f.writeCause(sb, err, opts, indent, depth)
	f.writeStackTrace(sb, err, opts, indent)
}

// writeTimestamp writes the timestamp if enabled
func (f *DefaultErrorFormatter) writeTimestamp(sb *strings.Builder, err MageError, opts FormatOptions, indent string) {
	if opts.IncludeTimestamp && !err.Context().Timestamp.IsZero() {
		fmt.Fprintf(sb, "\n%s  Time: %s", indent,
			err.Context().Timestamp.Format(opts.TimeFormat))
	}
}

// writeContext writes the context information if enabled
func (f *DefaultErrorFormatter) writeContext(sb *strings.Builder, err MageError, opts FormatOptions, indent string) {
	if opts.IncludeContext {
		ctx := err.Context()
		f.formatContext(sb, &ctx, indent)
	}
}

// writeFields writes the context fields if enabled
func (f *DefaultErrorFormatter) writeFields(sb *strings.Builder, err MageError, opts FormatOptions, indent string) {
	if !opts.IncludeFields || len(err.Context().Fields) == 0 {
		return
	}

	fmt.Fprintf(sb, "\n%s  Fields:", indent)
	for k, v := range err.Context().Fields {
		fmt.Fprintf(sb, "\n%s    %s: %v", indent, k, v)
	}
}

// writeCause writes the error cause if enabled
func (f *DefaultErrorFormatter) writeCause(sb *strings.Builder, err MageError, opts FormatOptions, indent string, depth int) {
	if !opts.IncludeCause || err.Cause() == nil {
		return
	}

	fmt.Fprintf(sb, "\n%s  Caused by:", indent)
	var causeMageErr MageError
	if errors.As(err.Cause(), &causeMageErr) {
		sb.WriteString("\n")
		sb.WriteString(f.formatMageError(causeMageErr, opts, depth+1))
	} else {
		fmt.Fprintf(sb, "\n%s    %v", indent, err.Cause())
	}
}

// writeStackTrace writes the stack trace if enabled
func (f *DefaultErrorFormatter) writeStackTrace(sb *strings.Builder, err MageError, opts FormatOptions, indent string) {
	if !opts.IncludeStack || err.Context().StackTrace == "" {
		return
	}

	fmt.Fprintf(sb, "\n%s  Stack trace:", indent)
	stackLines := strings.Split(err.Context().StackTrace, "\n")
	for _, line := range stackLines {
		if strings.TrimSpace(line) != "" {
			fmt.Fprintf(sb, "\n%s  %s", indent, line)
		}
	}
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
