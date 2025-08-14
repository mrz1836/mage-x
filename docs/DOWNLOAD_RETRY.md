# Download Retry Configuration

The mage-x project includes comprehensive retry logic for all external binary downloads and tool installations to handle network failures, timeouts, and other transient issues.

## Overview

The retry system provides:
- **Exponential backoff** with configurable multiplier
- **Maximum retry attempts** with sensible defaults
- **Timeout handling** for individual download attempts
- **Resumable downloads** using HTTP Range headers
- **Checksum verification** for downloaded files
- **Multiple fallback strategies** for tool installation
- **Environment variable overrides** for CI/CD environments

## Configuration

### YAML Configuration

Add a `download` section to your `mage.yaml` configuration file:

```yaml
download:
  max_retries: 5              # Maximum number of retry attempts (default: 5)
  initial_delay_ms: 1000      # Initial delay between retries in ms (default: 1000)
  max_delay_ms: 30000         # Maximum delay between retries in ms (default: 30000)
  timeout_ms: 60000           # Timeout per attempt in ms (default: 60000)
  backoff_multiplier: 2.0     # Exponential backoff multiplier (default: 2.0)
  enable_resume: true         # Enable resumable downloads (default: true)
  user_agent: "mage-x/1.0"    # HTTP User-Agent string (default: "mage-x-downloader/1.0")
```

### Environment Variable Overrides

All configuration options can be overridden using environment variables:

```bash
export MAGE_DOWNLOAD_RETRIES=3              # Override max_retries
export MAGE_DOWNLOAD_TIMEOUT=30000          # Override timeout_ms
export MAGE_DOWNLOAD_INITIAL_DELAY=500      # Override initial_delay_ms
export MAGE_DOWNLOAD_MAX_DELAY=15000        # Override max_delay_ms
export MAGE_DOWNLOAD_BACKOFF=1.5            # Override backoff_multiplier
export MAGE_DOWNLOAD_RESUME=true            # Override enable_resume
export MAGE_DOWNLOAD_USER_AGENT="ci/1.0"    # Override user_agent
```

## Supported Tools

The retry logic is automatically applied to all tool installations including:

### Go Tools
- **gofumpt**: Stricter Go formatter
- **goimports**: Automatic import management
- **govulncheck**: Vulnerability scanner
- **mockgen**: Mock generator (if configured)
- **swag**: Swagger documentation generator (if configured)

### External Tools
- **golangci-lint**: Comprehensive Go linter (via curl script)
- **Custom tools**: Any tools defined in `tools.custom` configuration

### Homebrew Integration (macOS)
- Homebrew installations also include retry logic
- Fallback to direct installation if Homebrew fails

## Retry Behavior

### Exponential Backoff

The retry system uses exponential backoff to avoid overwhelming servers:

```
Attempt 1: Immediate
Attempt 2: Wait 1000ms (initial_delay_ms)
Attempt 3: Wait 2000ms (initial_delay_ms * backoff_multiplier)
Attempt 4: Wait 4000ms (previous_delay * backoff_multiplier)
Attempt 5: Wait 8000ms (capped at max_delay_ms)
```

### Retriable vs Non-Retriable Errors

The system automatically distinguishes between retriable and non-retriable errors:

#### Retriable Errors (will trigger retry):
- **Network timeouts**: `connection timeout`, `i/o timeout`
- **Connection issues**: `connection refused`, `connection reset`
- **DNS problems**: `no such host`, `temporary failure in name resolution`
- **Network unreachable**: `network is unreachable`, `host is unreachable`
- **TLS handshake failures**: `tls handshake timeout`
- **Go module errors**: `go: downloading`, `verifying module`, `sumdb verification`
- **HTTP 5xx server errors**: Temporary server issues

#### Non-Retriable Errors (will fail immediately):
- **HTTP 4xx client errors**: `404 Not Found`, `403 Forbidden`, `401 Unauthorized`
- **File system errors**: `permission denied`, `file not found`
- **Validation errors**: Invalid checksums, malformed URLs
- **Context cancellation**: User-initiated cancellation

### Resumable Downloads

When `enable_resume` is true (default), the system supports resuming interrupted downloads:

1. If a download is interrupted, the partial file is preserved
2. On retry, an HTTP `Range` header is sent to request only the remaining bytes
3. The server response is validated to ensure proper resume support
4. If resume fails, the download restarts from the beginning

## Tool-Specific Behavior

### golangci-lint Installation

golangci-lint uses a multi-step installation process with fallbacks:

1. **Homebrew (macOS)**: Try `brew install golangci-lint` with retry
2. **Download Script**: Download and execute the official installation script
3. **Direct Command**: Fall back to direct curl execution if script download fails
4. **Each step includes retry logic** with the configured parameters

### Go Tool Installation

Go tools installed via `go install` include additional fallback mechanisms:

1. **Standard Installation**: `go install module@version` with retry
2. **Direct Proxy**: Fall back to `GOPROXY=direct` if standard installation fails
3. **Version Handling**: Automatic `@latest` suffix for unversioned tools

## Error Handling and Logging

### Retry Logging

The system provides detailed logging during retry operations:

```
INFO: Installing gofumpt with retry logic...
WARN: Installation failed: connection refused, trying direct proxy...
INFO: Command go attempt 2/4 failed: timeout. Retrying in 2s...
SUCCESS: gofumpt installed successfully
```

### Error Context

Failed installations provide comprehensive error messages:

```
failed to install gofumpt after 5 retries and fallback: 
  last error: context deadline exceeded
  attempted methods: go install, direct proxy
  total duration: 45.2s
```

## CI/CD Optimizations

### Environment-Specific Settings

For CI/CD environments, consider these optimizations:

```bash
# Faster retries for CI (shorter delays, fewer attempts)
export MAGE_DOWNLOAD_RETRIES=3
export MAGE_DOWNLOAD_INITIAL_DELAY=250
export MAGE_DOWNLOAD_MAX_DELAY=5000
export MAGE_DOWNLOAD_TIMEOUT=30000

# Disable resume for clean CI environments
export MAGE_DOWNLOAD_RESUME=false

# Custom user agent for telemetry
export MAGE_DOWNLOAD_USER_AGENT="ci-system/1.0"
```

### Parallel Builds

The retry system is thread-safe and supports parallel tool installations:

```bash
# Multiple tools can be installed concurrently
mage tools:install &
mage lint:install &
wait
```

## Monitoring and Metrics

### Success Metrics

Track installation success rates in your monitoring:

- **Retry attempts per tool**: Monitor if certain tools consistently require retries
- **Installation duration**: Track how retry logic affects build times
- **Failure patterns**: Identify network issues or problematic package sources

### Debugging

Enable verbose logging to debug retry behavior:

```bash
export VERBOSE=true
mage tools:install
```

This will show detailed retry attempts, backoff delays, and fallback strategies.

## Best Practices

### Development

1. **Test network resilience**: Use the integration tests to verify retry behavior
2. **Configure for your environment**: Adjust timeouts based on your network conditions
3. **Monitor retry patterns**: Track which tools require retries most frequently

### Production/CI

1. **Use shorter delays**: CI environments can use shorter retry delays
2. **Set appropriate timeouts**: Balance between reliability and build speed
3. **Monitor failure rates**: Set up alerts for high retry rates or persistent failures
4. **Cache dependencies**: Consider using dependency caching to reduce download needs

### Network-Constrained Environments

1. **Increase timeouts**: `MAGE_DOWNLOAD_TIMEOUT=120000` for slower networks
2. **Reduce retry attempts**: `MAGE_DOWNLOAD_RETRIES=2` to fail faster
3. **Enable resume**: Ensure `MAGE_DOWNLOAD_RESUME=true` for interrupted downloads
4. **Use proxy settings**: Configure corporate proxy settings if needed

## Security Considerations

### Download Verification

- **Checksum validation**: SHA256 checksums are verified when available
- **HTTPS enforcement**: All downloads use HTTPS where possible
- **User agent identification**: Custom user agents help identify legitimate requests
- **Path validation**: Download paths are validated to prevent directory traversal

### Secure Execution

- **Command validation**: All executed commands go through security validation
- **Environment filtering**: Sensitive environment variables are filtered
- **Timeout enforcement**: Prevents infinite-running processes
- **Audit logging**: All retry attempts are logged for security monitoring

## Troubleshooting

### Common Issues

#### "Maximum retries exceeded"
- **Cause**: Network connectivity issues or server problems
- **Solution**: Check network connectivity, increase retry count, or check server status

#### "Checksum verification failed"
- **Cause**: Corrupted download or man-in-the-middle attack
- **Solution**: Retry download, check network security, verify source integrity

#### "Context deadline exceeded"
- **Cause**: Individual download attempts timing out
- **Solution**: Increase `timeout_ms` value or check network speed

#### "Resumable download failed"
- **Cause**: Server doesn't support HTTP Range requests
- **Solution**: Set `MAGE_DOWNLOAD_RESUME=false` to disable resume

### Debug Commands

```bash
# Test download configuration
mage tools:verify

# Force tool reinstallation with verbose logging
export VERBOSE=true
mage tools:install

# Test network connectivity
curl -v https://github.com/golangci/golangci-lint/releases/latest

# Check proxy settings
echo $HTTP_PROXY $HTTPS_PROXY $NO_PROXY
```

## API Reference

### DownloadConfig Structure

```go
type DownloadConfig struct {
    MaxRetries        int     `yaml:"max_retries"`        // Maximum retry attempts
    InitialDelayMs    int     `yaml:"initial_delay_ms"`   // Initial delay in milliseconds
    MaxDelayMs        int     `yaml:"max_delay_ms"`       // Maximum delay in milliseconds
    TimeoutMs         int     `yaml:"timeout_ms"`         // Per-attempt timeout in milliseconds
    BackoffMultiplier float64 `yaml:"backoff_multiplier"` // Exponential backoff multiplier
    EnableResume      bool    `yaml:"enable_resume"`      // Enable resumable downloads
    UserAgent         string  `yaml:"user_agent"`         // HTTP User-Agent string
}
```

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `MAGE_DOWNLOAD_RETRIES` | Maximum retry attempts | 5 | `3` |
| `MAGE_DOWNLOAD_TIMEOUT` | Timeout per attempt (ms) | 60000 | `30000` |
| `MAGE_DOWNLOAD_INITIAL_DELAY` | Initial delay (ms) | 1000 | `500` |
| `MAGE_DOWNLOAD_MAX_DELAY` | Maximum delay (ms) | 30000 | `15000` |
| `MAGE_DOWNLOAD_BACKOFF` | Backoff multiplier | 2.0 | `1.5` |
| `MAGE_DOWNLOAD_RESUME` | Enable resume | true | `false` |
| `MAGE_DOWNLOAD_USER_AGENT` | User agent string | "mage-x-downloader/1.0" | "ci/1.0" |