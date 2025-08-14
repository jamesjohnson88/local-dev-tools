# Dynamic Request Scheduler - User Guide

## Overview

The Dynamic Request Scheduler is a flexible tool for scheduling and sending HTTP requests with dynamic values. It allows you to define requests in configuration files (YAML or JSON) and automatically resolves dynamic fields like timestamps, UUIDs, and environment variables at runtime.

## Quick Start

### 1. Basic Configuration

Create a configuration file (e.g., `config.yaml`):

```yaml
requests:
  - name: "Health Check"
    schedule:
      relative: "5m"
    http:
      method: "GET"
      url: "https://api.example.com/health"
      headers:
        User-Agent: "HealthChecker/1.0"
      body: null
```

### 2. Run the Scheduler

```bash
# Run with configuration file
./dynamic-request-scheduler --config config.yaml

# Run in legacy mode (single request, fixed interval)
./dynamic-request-scheduler --interval 60
```

## Configuration Structure

### Request Definition

Each request in your configuration has three main sections:

```yaml
requests:
  - name: "Request Name"           # Human-readable identifier
    schedule: { ... }              # When to run this request
    http: { ... }                  # HTTP request details
```

### Schedule Specification

The `schedule` section defines when the request should run. You can use one of these strategies:

```yaml
schedule:
  # Option 1: Run at specific Unix timestamp
  epoch: 1704067200
  
  # Option 2: Run relative to current time
  relative: "10m"  # 10 minutes from now
  
  # Option 3: Use template to compute time
  template: "{{ addMinutes 15 now | unix }}"
  
  # Option 4: Cron expression (coming in Phase 3)
  cron: "*/5 * * * *"
  
  # Optional: Add random jitter to avoid thundering herd
  jitter: "±30s"
```

**Note**: Only one scheduling strategy can be specified per request.

### HTTP Request Specification

The `http` section defines the actual HTTP request:

```yaml
http:
  method: "POST"                    # HTTP method (GET, POST, PUT, DELETE, etc.)
  url: "https://api.example.com/endpoint"
  headers:                          # HTTP headers
    Content-Type: "application/json"
    Authorization: "Bearer {{ env 'API_TOKEN' }}"
  body:                             # Request body (null for GET requests)
    user_id: "{{ uuid }}"
    timestamp: "{{ now | rfc3339 }}"
```

## Dynamic Values and Templates

### Template Syntax

Dynamic values use Go's `text/template` syntax with double curly braces:

```yaml
# Basic template
"{{ now | unix }}"

# Template with parameters
"{{ addMinutes 5 now | unix }}"

# Variable substitution
"Bearer {{ .Variables.api_key }}"
```

### Available Functions

#### Time Functions

| Function | Description | Example |
|----------|-------------|---------|
| `now` | Current time | `{{ now }}` |
| `unix` | Unix timestamp | `{{ now \| unix }}` |
| `rfc3339` | RFC3339 formatted time | `{{ now \| rfc3339 }}` |
| `addSeconds` | Add seconds to time | `{{ addSeconds 30 now }}` |
| `addMinutes` | Add minutes to time | `{{ addMinutes 5 now }}` |
| `addHours` | Add hours to time | `{{ addHours 2 now }}` |

#### ID and Random Functions

| Function | Description | Example |
|----------|-------------|---------|
| `uuid` | Generate UUID v4 | `{{ uuid }}` |
| `randInt` | Random integer | `{{ randInt 1 100 }}` |
| `randFloat` | Random float 0-1 | `{{ randFloat }}` |
| `seq` | Incremental sequence | `{{ seq }}` |

#### Environment and Variables

| Function | Description | Example |
|----------|-------------|---------|
| `env` | Environment variable | `{{ env "API_KEY" }}` |
| `var` | User variable | `{{ var "user_id" }}` |

#### Utility Functions

| Function | Description | Example |
|----------|-------------|---------|
| `jitter` | Add random jitter | `{{ jitter now "30s" }}` |
| `upper` | Convert to uppercase | `{{ upper "hello" }}` |
| `lower` | Convert to lowercase | `{{ lower "WORLD" }}` |
| `trim` | Trim whitespace | `{{ trim "  text  " }}` |

### Variable Substitution

You can inject variables at runtime using the `--var` flag:

```bash
./dynamic-request-scheduler --config config.yaml --var "api_key=secret123" --var "user_id=456"
```

Then reference them in your config:

```yaml
headers:
  Authorization: "Bearer {{ .Variables.api_key }}"
  X-User-ID: "{{ .Variables.user_id }}"
```

## Configuration Examples

### Example 1: Simple Health Check

```yaml
requests:
  - name: "API Health Check"
    schedule:
      relative: "1m"
    http:
      method: "GET"
      url: "https://api.example.com/health"
      headers:
        User-Agent: "HealthChecker/1.0"
      body: null
```

### Example 2: Dynamic POST Request

```yaml
requests:
  - name: "Create User"
    schedule:
      relative: "5m"
      jitter: "±10s"
    http:
      method: "POST"
      url: "https://api.example.com/users"
      headers:
        Content-Type: "application/json"
        X-Trace-ID: "{{ uuid }}"
        X-Timestamp: "{{ now | unix }}"
      body:
        user_id: "{{ uuid }}"
        created_at: "{{ now | rfc3339 }}"
        metadata:
          source: "dynamic-scheduler"
          iteration: "{{ seq }}"
```

### Example 3: Template-Based Scheduling

```yaml
requests:
  - name: "Scheduled Task"
    schedule:
      template: "{{ addMinutes 30 now | unix }}"
    http:
      method: "POST"
      url: "https://api.example.com/tasks"
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer {{ env 'TASK_API_TOKEN' }}"
      body:
        task_name: "periodic-cleanup"
        scheduled_for: "{{ now | unix }}"
        parameters:
          batch_size: "{{ randInt 10 100 }}"
```

### Example 4: Multiple Requests with Different Strategies

```yaml
requests:
  # Run every 5 minutes
  - name: "Frequent Health Check"
    schedule:
      relative: "5m"
    http:
      method: "GET"
      url: "https://api.example.com/health"
      headers: {}
      body: null

  # Run at specific time
  - name: "Daily Report"
    schedule:
      epoch: 1704067200  # 2024-01-01 00:00:00 UTC
    http:
      method: "POST"
      url: "https://api.example.com/reports/daily"
      headers:
        Content-Type: "application/json"
      body:
        report_date: "{{ now | rfc3339 }}"
        type: "daily-summary"

  # Run with computed time
  - name: "Next Hour Task"
    schedule:
      template: "{{ addHours 1 now | unix }}"
      jitter: "±5m"
    http:
      method: "GET"
      url: "https://api.example.com/tasks/next-hour"
      headers: {}
      body: null
```

## Command Line Options

### Basic Options

```bash
./dynamic-request-scheduler [OPTIONS]
```

| Option | Description | Default |
|--------|-------------|---------|
| `--config <path>` | Path to configuration file | None (legacy mode) |
| `--interval <seconds>` | Interval in seconds (legacy mode) | 60 |

### Available Options

| Option | Description | Default |
|--------|-------------|---------|
| `--config <path>` | Path to configuration file | None (legacy mode) |
| `--interval <seconds>` | Interval in seconds (legacy mode) | 60 |
| `--dry-run` | Show resolved requests without sending | false |
| `--once` | Run all requests once and exit | false |
| `--workers <N>` | Number of worker goroutines | 1 |
| `--concurrency <N>` | Maximum concurrent requests | 10 |
| `--timeout <duration>` | HTTP request timeout | 30s |

### Planned Options (Future)

| Option | Description | Status |
|--------|-------------|--------|
| `--var <key=value>` | Set template variables | Coming Soon |
| `--seed <number>` | Seed for deterministic random values | Coming Soon |
| `--limit <N>` | Maximum number of requests to run | Coming Soon |

## Best Practices

### 1. Naming Conventions

- Use descriptive names for requests
- Include the target service/endpoint in the name
- Add frequency indicators (e.g., "Daily", "Hourly", "On-Demand")

### 2. Error Handling

- Always validate your configuration files
- Test templates with `--dry-run` (when available)
- Use environment variables for sensitive data

### 3. Scheduling Strategy Selection

- **`epoch`**: For one-time, specific time events
- **`relative`**: For recurring events with simple intervals
- **`template`**: For complex time calculations
- **`cron`**: For traditional cron-like scheduling (coming soon)

### 4. Dynamic Values

- Use `{{ uuid }}` for unique identifiers
- Use `{{ now | unix }}` for current timestamps
- Use `{{ seq }}` for incremental counters
- Use `{{ env "VAR" }}` for environment variables

### 5. Jitter Usage

- Add jitter to avoid thundering herd problems
- Use small jitter values (e.g., "±30s") for most cases
- Larger jitter for distributed systems

## Troubleshooting

### Common Issues

1. **Template Syntax Errors**
   - Check for balanced `{{` and `}}`
   - Verify function names and parameters
   - Ensure proper spacing around pipes

2. **Invalid Schedules**
   - Only one scheduling strategy per request
   - Valid duration formats: "5m", "2h", "30s"
   - Unix timestamps must be positive integers

3. **Missing Variables**
   - Check environment variable names
   - Verify `--var` flag usage
   - Ensure variables are defined before use

### Debug Mode

When available, use `--dry-run` to see resolved requests without sending them:

```bash
./dynamic-request-scheduler --config config.yaml --dry-run
```

This will show you:
- Resolved URLs
- Final headers and body
- Computed schedule times
- Any template resolution errors

## Next Steps

This guide covers the current functionality. Future versions will include:

- **Variable Injection**: `--var` flag for runtime variable injection
- **Seeded Randomness**: `--seed` flag for deterministic results
- **Request Chaining**: Dependent request sequences
- **Response Handling**: Capture and reuse response data
- **Metrics and Monitoring**: Request success rates and timing
- **Timezone Support**: Per-request timezone specification

## Getting Help

- Check the configuration examples in the `examples/` directory
- Review the detailed function reference in `docs/FUNCTIONS.md`
- Examine the scheduling strategies in `docs/SCHEDULING.md`
- Run with `--help` for current command line options
