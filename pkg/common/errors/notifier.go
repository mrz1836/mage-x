// Package errors provides comprehensive error notification capabilities
package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultErrorNotifier implements the ErrorNotifier interface
type DefaultErrorNotifier struct {
	mu             sync.RWMutex
	channels       map[string]NotificationChannel
	threshold      Severity
	rateLimit      time.Duration
	rateLimitCount int
	lastNotified   map[string]time.Time
	notifyCount    map[string]int
	enabled        bool
}

// NewErrorNotifier creates a new error notifier with default settings
func NewErrorNotifier() ErrorNotifier {
	return &DefaultErrorNotifier{
		channels:       make(map[string]NotificationChannel),
		threshold:      SeverityError,
		rateLimit:      time.Minute,
		rateLimitCount: 10,
		lastNotified:   make(map[string]time.Time),
		notifyCount:    make(map[string]int),
		enabled:        true,
	}
}

// Notify sends a notification for the given error
func (n *DefaultErrorNotifier) Notify(err error) error {
	return n.NotifyWithContext(context.Background(), err)
}

// NotifyWithContext sends a notification with context
func (n *DefaultErrorNotifier) NotifyWithContext(ctx context.Context, err error) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.enabled || err == nil {
		return nil
	}

	// Convert to MageError if possible
	var mageErr MageError
	var me MageError
	if errors.As(err, &me) {
		mageErr = me
	} else {
		builder := NewErrorBuilder()
		builder.WithMessage("%s", err.Error())
		builder.WithCode(ErrUnknown)
		builder.WithSeverity(SeverityError)
		mageErr = builder.Build()
	}

	// Check if error meets threshold
	if mageErr.Severity() < n.threshold {
		return nil
	}

	// Check rate limiting
	errorKey := fmt.Sprintf("%s:%s", mageErr.Code(), mageErr.Error())
	if n.isRateLimited(errorKey) {
		return nil
	}

	// Create notification
	notification := ErrorNotification{
		Error:       mageErr,
		Timestamp:   time.Now(),
		Environment: "mage-build-system",
		Hostname:    "localhost",
		Service:     "mage",
		Metadata:    map[string]interface{}{"severity": mageErr.Severity()},
	}

	// Send to all enabled channels
	var errors []error
	for _, channel := range n.channels {
		if channel.IsEnabled() {
			if err := channel.Send(ctx, notification); err != nil {
				errors = append(errors, fmt.Errorf("failed to send notification via %s: %w", channel.Name(), err))
			}
		}
	}

	// Update rate limiting counters
	n.updateRateLimit(errorKey)

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}

	return nil
}

// SetThreshold sets the minimum severity for notifications
func (n *DefaultErrorNotifier) SetThreshold(severity Severity) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.threshold = severity
}

// SetRateLimit sets the rate limiting parameters
func (n *DefaultErrorNotifier) SetRateLimit(duration time.Duration, count int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.rateLimit = duration
	n.rateLimitCount = count
}

// AddChannel adds a notification channel
func (n *DefaultErrorNotifier) AddChannel(channel NotificationChannel) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if channel == nil {
		return fmt.Errorf("channel cannot be nil")
	}

	n.channels[channel.Name()] = channel
	return nil
}

// RemoveChannel removes a notification channel
func (n *DefaultErrorNotifier) RemoveChannel(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.channels[name]; !exists {
		return fmt.Errorf("channel %s not found", name)
	}

	delete(n.channels, name)
	return nil
}

// GetChannels returns all registered channels
func (n *DefaultErrorNotifier) GetChannels() []NotificationChannel {
	n.mu.RLock()
	defer n.mu.RUnlock()

	channels := make([]NotificationChannel, 0, len(n.channels))
	for _, channel := range n.channels {
		channels = append(channels, channel)
	}

	return channels
}

// SetEnabled enables or disables notifications
func (n *DefaultErrorNotifier) SetEnabled(enabled bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.enabled = enabled
}

// isRateLimited checks if an error type is rate limited
func (n *DefaultErrorNotifier) isRateLimited(errorKey string) bool {
	now := time.Now()

	// Check if we're within the rate limit window
	if lastTime, exists := n.lastNotified[errorKey]; exists {
		if now.Sub(lastTime) < n.rateLimit {
			// Check if we've exceeded the count limit
			if count, countExists := n.notifyCount[errorKey]; countExists && count >= n.rateLimitCount {
				return true
			}
		} else {
			// Reset count for new window
			n.notifyCount[errorKey] = 0
		}
	}

	return false
}

// updateRateLimit updates the rate limiting counters
func (n *DefaultErrorNotifier) updateRateLimit(errorKey string) {
	now := time.Now()
	n.lastNotified[errorKey] = now
	n.notifyCount[errorKey]++
}

// Note: ErrorNotification is defined in interfaces.go

// EmailChannel implements NotificationChannel for email notifications
type EmailChannel struct {
	name     string
	smtpHost string
	smtpPort int
	smtpUser string
	smtpPass string
	from     string
	to       []string
	enabled  bool
	template string
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(name, smtpHost string, smtpPort int, smtpUser, smtpPass, from string, to []string) *EmailChannel {
	return &EmailChannel{
		name:     name,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		smtpUser: smtpUser,
		smtpPass: smtpPass,
		from:     from,
		to:       to,
		enabled:  true,
		template: defaultEmailTemplate,
	}
}

func (e *EmailChannel) Name() string {
	return e.name
}

func (e *EmailChannel) Send(ctx context.Context, notification ErrorNotification) error {
	if !e.enabled {
		return nil
	}

	// Format email content
	subject := fmt.Sprintf("[MAGE ERROR] %s - %s", notification.Error.Severity().String(), notification.Error.Code())
	body := e.formatEmailBody(notification)

	// Send email
	auth := smtp.PlainAuth("", e.smtpUser, e.smtpPass, e.smtpHost)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(e.to, ","), subject, body)

	addr := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)
	return smtp.SendMail(addr, auth, e.from, e.to, []byte(msg))
}

func (e *EmailChannel) IsEnabled() bool {
	return e.enabled
}

func (e *EmailChannel) SetEnabled(enabled bool) {
	e.enabled = enabled
}

func (e *EmailChannel) formatEmailBody(notification ErrorNotification) string {
	return fmt.Sprintf(`
Error Details:
- Code: %s
- Message: %s
- Severity: %s
- Timestamp: %s
- Environment: %s
- Service: %s

Context:
%s

Stack Trace:
%s
`,
		notification.Error.Code(),
		notification.Error.Error(),
		notification.Error.Severity().String(),
		notification.Timestamp.Format(time.RFC3339),
		notification.Environment,
		notification.Service,
		formatErrorContext(notification.Error.Context()),
		notification.Error.Error(), // Use Error() since StackTrace might not exist
	)
}

const defaultEmailTemplate = `
Error Code: {{.Error.Code}}
Message: {{.Error.Message}}
Severity: {{.Severity}}
Timestamp: {{.Timestamp}}
Source: {{.Source}}
`

// WebhookChannel implements NotificationChannel for webhook notifications
type WebhookChannel struct {
	name    string
	url     string
	headers map[string]string
	enabled bool
	client  *http.Client
}

// NewWebhookChannel creates a new webhook notification channel
func NewWebhookChannel(name, url string, headers map[string]string) *WebhookChannel {
	return &WebhookChannel{
		name:    name,
		url:     url,
		headers: headers,
		enabled: true,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WebhookChannel) Name() string {
	return w.name
}

func (w *WebhookChannel) Send(ctx context.Context, notification ErrorNotification) error {
	if !w.enabled {
		return nil
	}

	// Create JSON payload
	payload := fmt.Sprintf(`{
		"error_code": "%s",
		"message": "%s",
		"severity": "%s",
		"timestamp": "%s",
		"environment": "%s",
		"service": "%s"
	}`,
		notification.Error.Code(),
		escapeJSON(notification.Error.Error()),
		notification.Error.Severity().String(),
		notification.Timestamp.Format(time.RFC3339),
		notification.Environment,
		notification.Service,
	)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", w.url, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for key, value := range w.headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log close error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

func (w *WebhookChannel) IsEnabled() bool {
	return w.enabled
}

func (w *WebhookChannel) SetEnabled(enabled bool) {
	w.enabled = enabled
}

// ConsoleChannel implements NotificationChannel for console output
type ConsoleChannel struct {
	name      string
	enabled   bool
	formatter ErrorFormatter
}

// NewConsoleChannel creates a new console notification channel
func NewConsoleChannel(name string) *ConsoleChannel {
	return &ConsoleChannel{
		name:      name,
		enabled:   true,
		formatter: NewFormatter(),
	}
}

func (c *ConsoleChannel) Name() string {
	return c.name
}

func (c *ConsoleChannel) Send(ctx context.Context, notification ErrorNotification) error {
	if !c.enabled {
		return nil
	}

	formatted := c.formatter.Format(notification.Error)
	fmt.Printf("[NOTIFICATION] %s: %s\n", notification.Timestamp.Format("15:04:05"), formatted)

	return nil
}

func (c *ConsoleChannel) IsEnabled() bool {
	return c.enabled
}

func (c *ConsoleChannel) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Helper functions

func formatErrorContext(context ErrorContext) string {
	if len(context.Fields) == 0 {
		return "No additional context"
	}

	var lines []string
	for key, value := range context.Fields {
		lines = append(lines, fmt.Sprintf("  %s: %v", key, value))
	}

	return strings.Join(lines, "\n")
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// MockErrorNotifier implements ErrorNotifier for testing
type MockErrorNotifier struct {
	NotifyCalls            []error
	NotifyWithContextCalls []MockNotifyWithContextCall
	SetThresholdCalls      []Severity
	SetRateLimitCalls      []MockSetRateLimitCall
	AddChannelCalls        []NotificationChannel
	RemoveChannelCalls     []string
	ShouldError            bool
	Threshold              Severity
	Enabled                bool
	Channels               map[string]NotificationChannel
}

type MockNotifyWithContextCall struct {
	Error error
}

type MockSetRateLimitCall struct {
	Duration time.Duration
	Count    int
}

func NewMockErrorNotifier() *MockErrorNotifier {
	return &MockErrorNotifier{
		NotifyCalls:            make([]error, 0),
		NotifyWithContextCalls: make([]MockNotifyWithContextCall, 0),
		SetThresholdCalls:      make([]Severity, 0),
		SetRateLimitCalls:      make([]MockSetRateLimitCall, 0),
		AddChannelCalls:        make([]NotificationChannel, 0),
		RemoveChannelCalls:     make([]string, 0),
		Threshold:              SeverityError,
		Enabled:                true,
		Channels:               make(map[string]NotificationChannel),
	}
}

func (m *MockErrorNotifier) Notify(err error) error {
	m.NotifyCalls = append(m.NotifyCalls, err)
	if m.ShouldError {
		return fmt.Errorf("mock notify error")
	}
	return nil
}

func (m *MockErrorNotifier) NotifyWithContext(ctx context.Context, err error) error {
	m.NotifyWithContextCalls = append(m.NotifyWithContextCalls, MockNotifyWithContextCall{
		Error: err,
	})
	if m.ShouldError {
		return fmt.Errorf("mock notify with context error")
	}
	return nil
}

func (m *MockErrorNotifier) SetThreshold(severity Severity) {
	m.SetThresholdCalls = append(m.SetThresholdCalls, severity)
	m.Threshold = severity
}

func (m *MockErrorNotifier) SetRateLimit(duration time.Duration, count int) {
	m.SetRateLimitCalls = append(m.SetRateLimitCalls, MockSetRateLimitCall{
		Duration: duration,
		Count:    count,
	})
}

func (m *MockErrorNotifier) AddChannel(channel NotificationChannel) error {
	m.AddChannelCalls = append(m.AddChannelCalls, channel)
	if m.ShouldError {
		return fmt.Errorf("mock add channel error")
	}
	m.Channels[channel.Name()] = channel
	return nil
}

func (m *MockErrorNotifier) RemoveChannel(name string) error {
	m.RemoveChannelCalls = append(m.RemoveChannelCalls, name)
	if m.ShouldError {
		return fmt.Errorf("mock remove channel error")
	}
	delete(m.Channels, name)
	return nil
}
