package log

import "context"

// getRequestIDKeys returns the context keys checked for request ID extraction.
// These cover common conventions used across different frameworks and standards.
func getRequestIDKeys() []interface{} {
	return []interface{}{"request_id", "requestId", "req_id", "trace_id"}
}

// GetRequestIDFromContext extracts a request ID from the given context.
// It checks multiple common context keys used for request/trace identification.
// Returns an empty string if no valid request ID is found.
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	for _, key := range getRequestIDKeys() {
		if value := ctx.Value(key); value != nil {
			if requestID, ok := value.(string); ok && requestID != "" {
				return requestID
			}
		}
	}

	return ""
}
