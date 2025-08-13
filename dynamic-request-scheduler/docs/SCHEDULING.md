# Scheduling Strategies Guide

This document explains the different scheduling strategies available in the Dynamic Request Scheduler, how they work, and when to use each one.

## Overview

The scheduler supports multiple strategies for determining when requests should run:

- **Epoch**: Run at a specific Unix timestamp
- **Relative**: Run relative to the current time
- **Template**: Run at a time computed by a template
- **Cron**: Run according to cron expressions (coming in Phase 3)
- **Jitter**: Add randomness to any schedule (optional)

## Schedule Specification

Each request can specify exactly one scheduling strategy in the `schedule` section:

```yaml
requests:
  - name: "Example Request"
    schedule:
      # Choose ONE of these:
      epoch: 1704067200        # Specific timestamp
      relative: "5m"           # Relative to now
      template: "{{ ... }}"    # Computed time
      cron: "*/5 * * * *"     # Cron expression (coming soon)
      
      # Optional: Add jitter to any schedule
      jitter: "±30s"
    http:
      # ... HTTP request details
```

**Important**: Only one scheduling strategy can be specified per request. The scheduler will validate this and return an error if multiple strategies are specified.

## Epoch Scheduling

### How It Works

Epoch scheduling runs a request at a specific Unix timestamp (seconds since January 1, 1970 UTC).

### Syntax

```yaml
schedule:
  epoch: 1704067200  # Unix timestamp
```

### Examples

```yaml
# Run on January 1, 2024 at 00:00:00 UTC
requests:
  - name: "New Year Task"
    schedule:
      epoch: 1704067200
    http:
      method: "POST"
      url: "https://api.example.com/tasks/new-year"
      body:
        event: "new-year-2024"
        timestamp: "{{ now | unix }}"

# Run at a specific historical time
requests:
  - name: "Historical Data"
    schedule:
      epoch: 1640995200  # January 1, 2022 00:00:00 UTC
    http:
      method: "GET"
      url: "https://api.example.com/data/2022-01-01"
```

### Use Cases

- **One-time events**: Specific dates and times
- **Historical processing**: Backfill operations
- **Fixed schedules**: Known future events
- **Testing**: Reproducible execution times

### Considerations

- **Time zones**: Epoch timestamps are always UTC
- **Past times**: Requests with past timestamps will run immediately
- **Precision**: Unix timestamps have second-level precision
- **Human readability**: Hard to read and modify manually

## Relative Scheduling

### How It Works

Relative scheduling runs a request after a specified duration from the current time. The duration is parsed using Go's duration syntax.

### Syntax

```yaml
schedule:
  relative: "5m"      # 5 minutes from now
  relative: "2h30m"   # 2 hours 30 minutes from now
  relative: "1d"      # 1 day from now
```

### Duration Format

Go duration syntax supports these units:

| Unit | Description | Examples |
|------|-------------|----------|
| `ns` | Nanoseconds | `100ns`, `1.5ns` |
| `us` | Microseconds | `100us`, `1.5us` |
| `ms` | Milliseconds | `100ms`, `1.5ms` |
| `s` | Seconds | `30s`, `1.5s` |
| `m` | Minutes | `5m`, `1.5m` |
| `h` | Hours | `2h`, `1.5h` |
| `d` | Days | `1d`, `1.5d` |

### Examples

```yaml
# Simple intervals
requests:
  - name: "Health Check"
    schedule:
      relative: "1m"    # Every minute
    http:
      method: "GET"
      url: "https://api.example.com/health"

  - name: "Data Sync"
    schedule:
      relative: "5m"    # Every 5 minutes
    http:
      method: "POST"
      url: "https://api.example.com/sync"

# Complex durations
requests:
  - name: "Daily Report"
    schedule:
      relative: "24h"   # Every 24 hours
    http:
      method: "POST"
      url: "https://api.example.com/reports/daily"

  - name: "Weekly Cleanup"
    schedule:
      relative: "168h"  # 7 days (7 * 24 hours)
    http:
      method: "POST"
      url: "https://api.example.com/cleanup/weekly"

# Precise timing
requests:
  - name: "Quarterly Task"
    schedule:
      relative: "2160h" # 90 days (90 * 24 hours)
    http:
      method: "POST"
      url: "https://api.example.com/tasks/quarterly"
```

### Use Cases

- **Recurring tasks**: Health checks, data syncs, reports
- **Simple intervals**: Minutes, hours, days
- **Development and testing**: Quick iteration cycles
- **Monitoring**: Regular status checks

### Considerations

- **Recurrence**: Relative schedules run once per interval
- **Drift**: No automatic correction for execution delays
- **Precision**: Duration parsing is exact
- **Human readability**: Easy to understand and modify

## Template Scheduling

### How It Works

Template scheduling uses Go templates to compute the execution time. The template must evaluate to a Unix timestamp.

### Syntax

```yaml
schedule:
  template: "{{ addMinutes 30 now | unix }}"
```

### Examples

```yaml
# Simple time calculations
requests:
  - name: "Next Hour Task"
    schedule:
      template: "{{ addHours 1 now | unix }}"
    http:
      method: "GET"
      url: "https://api.example.com/tasks/next-hour"

  - name: "Quarter Past"
    schedule:
      template: "{{ addMinutes 15 now | unix }}"
    http:
      method: "POST"
      url: "https://api.example.com/tasks/quarter-past"

# Complex time calculations
requests:
  - name: "Business Hours Task"
    schedule:
      template: "{{ addHours 9 (parseTime '2006-01-02' (now | rfc3339 | slice 0 10)) | unix }}"
    http:
      method: "POST"
      url: "https://api.example.com/tasks/business-hours"

# Conditional scheduling
requests:
  - name: "Adaptive Task"
    schedule:
      template: "{{ if gt (now | hour) 12 }}{{ addHours 2 now | unix }}{{ else }}{{ addHours 1 now | unix }}{{ end }}"
    http:
      method: "GET"
      url: "https://api.example.com/tasks/adaptive"
```

### Use Cases

- **Complex timing**: Business hours, conditional schedules
- **Dynamic intervals**: Time-based calculations
- **Integration**: Combine with external time sources
- **Custom logic**: Application-specific scheduling rules

### Considerations

- **Template complexity**: Can become hard to read
- **Error handling**: Template failures cause scheduling failures
- **Performance**: Templates are evaluated at scheduling time
- **Debugging**: Hard to predict exact execution times

## Cron Scheduling (Coming in Phase 3)

### How It Works

Cron scheduling uses traditional cron expressions to define recurring schedules with minute, hour, day, month, and weekday precision.

### Syntax

```yaml
schedule:
  cron: "*/5 * * * *"  # Every 5 minutes
  cron: "0 2 * * *"    # Every day at 2:00 AM
  cron: "0 9 * * 1"    # Every Monday at 9:00 AM
```

### Examples

```yaml
# Frequent tasks
requests:
  - name: "Health Check"
    schedule:
      cron: "*/1 * * * *"    # Every minute
    http:
      method: "GET"
      url: "https://api.example.com/health"

  - name: "Data Sync"
    schedule:
      cron: "*/5 * * * *"    # Every 5 minutes
    http:
      method: "POST"
      url: "https://api.example.com/sync"

# Daily tasks
requests:
  - name: "Morning Report"
    schedule:
      cron: "0 9 * * *"      # 9:00 AM daily
    http:
      method: "POST"
      url: "https://api.example.com/reports/morning"

  - name: "Nightly Backup"
    schedule:
      cron: "0 2 * * *"      # 2:00 AM daily
    http:
      method: "POST"
      url: "https://api.example.com/backup"

# Weekly tasks
requests:
  - name: "Weekly Summary"
    schedule:
      cron: "0 10 * * 1"     # Monday 10:00 AM
    http:
      method: "POST"
      url: "https://api.example.com/reports/weekly"

# Monthly tasks
requests:
  - name: "Monthly Cleanup"
    schedule:
      cron: "0 3 1 * *"      # 1st of month at 3:00 AM
    http:
      method: "POST"
      url: "https://api.example.com/cleanup/monthly"
```

### Use Cases

- **Traditional scheduling**: Unix/Linux cron-like behavior
- **Business schedules**: Regular business hours, weekly patterns
- **System maintenance**: Backups, cleanup, reports
- **Integration**: Compatible with existing cron-based systems

### Considerations

- **Complexity**: Cron expressions can be hard to read
- **Timezone handling**: Will support timezone specification
- **Precision**: Minute-level precision
- **Standard format**: Compatible with standard cron syntax

## Jitter

### How It Works

Jitter adds random variation to any schedule to prevent multiple instances from running simultaneously (thundering herd problem).

### Syntax

```yaml
schedule:
  relative: "5m"
  jitter: "±30s"      # Add/subtract up to 30 seconds
  jitter: "±2m"       # Add/subtract up to 2 minutes
  jitter: "±1h"       # Add/subtract up to 1 hour
```

### Examples

```yaml
# Health checks with jitter
requests:
  - name: "Health Check"
    schedule:
      relative: "1m"
      jitter: "±10s"    # Run between 50s and 70s from now
    http:
      method: "GET"
      url: "https://api.example.com/health"

# Data sync with jitter
requests:
  - name: "Data Sync"
    schedule:
      relative: "5m"
      jitter: "±30s"    # Run between 4m30s and 5m30s from now
    http:
      method: "POST"
      url: "https://api.example.com/sync"

# Daily tasks with jitter
requests:
  - name: "Daily Report"
    schedule:
      relative: "24h"
      jitter: "±1h"     # Run between 23h and 25h from now
    http:
      method: "POST"
      url: "https://api.example.com/reports/daily"
```

### Use Cases

- **Distributed systems**: Prevent coordinated execution
- **Load balancing**: Spread requests across time
- **Resource contention**: Avoid simultaneous resource usage
- **Monitoring**: Prevent alert storms

### Considerations

- **Randomness**: Uses seeded random for determinism
- **Range**: Jitter is applied within the specified duration
- **Predictability**: With fixed seed, jitter is reproducible
- **Overlap**: Jitter can cause schedules to overlap

## Schedule Validation

### Mutual Exclusivity

The scheduler enforces that only one scheduling strategy is specified per request:

```yaml
# ❌ Invalid: Multiple strategies
schedule:
  epoch: 1704067200
  relative: "5m"

# ✅ Valid: Single strategy
schedule:
  relative: "5m"
```

### Validation Rules

1. **Exactly one strategy**: Must specify exactly one of `epoch`, `relative`, `template`, or `cron`
2. **Valid values**: All values must be valid for their type
3. **Jitter optional**: Jitter can be specified with any strategy
4. **Template evaluation**: Templates must evaluate to valid Unix timestamps

### Error Messages

The scheduler provides clear error messages for validation failures:

```
Error: request "Health Check": schedule must specify exactly one strategy (epoch, relative, template, or cron)
Error: request "Data Sync": invalid relative duration "invalid-duration"
Error: request "Template Task": template evaluation failed: function "invalid_func" not defined
```

## Schedule Computation

### How It Works

1. **Parse schedule**: Extract scheduling strategy and parameters
2. **Compute base time**: Calculate the base execution time
3. **Apply jitter**: Add random variation if jitter is specified
4. **Validate result**: Ensure the final time is valid
5. **Return time**: Return the computed execution time

### Examples

```yaml
# Epoch scheduling
schedule:
  epoch: 1704067200
  jitter: "±30s"
# Base time: 1704067200 (Jan 1, 2024 00:00:00 UTC)
# Jitter range: 1704067170 to 1704067230
# Final time: Random value in range

# Relative scheduling
schedule:
  relative: "5m"
  jitter: "±1m"
# Base time: now + 5 minutes
# Jitter range: now + 4 minutes to now + 6 minutes
# Final time: Random value in range

# Template scheduling
schedule:
  template: "{{ addHours 1 now | unix }}"
  jitter: "±15m"
# Base time: now + 1 hour
# Jitter range: now + 45 minutes to now + 1 hour 15 minutes
# Final time: Random value in range
```

## Best Practices

### 1. Strategy Selection

- **Use `epoch`** for one-time, specific events
- **Use `relative`** for simple, recurring intervals
- **Use `template`** for complex time calculations
- **Use `cron`** for traditional cron-like scheduling (when available)

### 2. Jitter Usage

- **Small jitter** (seconds/minutes) for frequent tasks
- **Medium jitter** (minutes) for regular tasks
- **Large jitter** (hours) for daily/weekly tasks
- **No jitter** for time-critical operations

### 3. Schedule Design

- **Avoid conflicts**: Don't schedule multiple heavy tasks simultaneously
- **Consider resources**: Account for system capacity and API limits
- **Plan for failures**: Allow time for retries and error handling
- **Document schedules**: Make schedules clear and maintainable

### 4. Testing and Validation

- **Test templates**: Verify template syntax and logic
- **Validate schedules**: Check that schedules make sense
- **Use dry-run**: Test scheduling without execution (when available)
- **Monitor execution**: Track actual vs. expected execution times

## Common Patterns

### Health Check Pattern

```yaml
requests:
  - name: "Health Check"
    schedule:
      relative: "1m"
      jitter: "±10s"
    http:
      method: "GET"
      url: "https://api.example.com/health"
      headers:
        User-Agent: "HealthChecker/{{ seq }}"
```

### Data Sync Pattern

```yaml
requests:
  - name: "Data Sync"
    schedule:
      relative: "5m"
      jitter: "±30s"
    http:
      method: "POST"
      url: "https://api.example.com/sync"
      headers:
        Content-Type: "application/json"
        X-Sync-ID: "{{ uuid }}"
      body:
        sync_id: "{{ uuid }}"
        timestamp: "{{ now | unix }}"
        source: "dynamic-scheduler"
```

### Daily Report Pattern

```yaml
requests:
  - name: "Daily Report"
    schedule:
      relative: "24h"
      jitter: "±1h"
    http:
      method: "POST"
      url: "https://api.example.com/reports/daily"
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer {{ env 'REPORT_TOKEN' }}"
      body:
        report_date: "{{ now | rfc3339 }}"
        report_type: "daily-summary"
        generated_at: "{{ now | unix }}"
```

### Business Hours Pattern

```yaml
requests:
  - name: "Business Hours Task"
    schedule:
      template: "{{ addHours 9 (parseTime '2006-01-02' (now | rfc3339 | slice 0 10)) | unix }}"
      jitter: "±15m"
    http:
      method: "POST"
      url: "https://api.example.com/tasks/business-hours"
      headers:
        Content-Type: "application/json"
      body:
        task_name: "business-hours-task"
        scheduled_for: "{{ now | unix }}"
        business_date: "{{ now | rfc3339 | slice 0 10 }}"
```

## Troubleshooting

### Common Issues

1. **Invalid duration format**
   - Check Go duration syntax
   - Use valid units (s, m, h, d)
   - Avoid invalid combinations

2. **Template evaluation failures**
   - Verify template syntax
   - Check function names and arguments
   - Test templates independently

3. **Schedule conflicts**
   - Review all request schedules
   - Consider resource constraints
   - Use jitter to spread load

4. **Unexpected execution times**
   - Check timezone handling
   - Verify jitter calculations
   - Review template logic

### Debug Tips

- **Validate configuration**: Check for syntax and validation errors
- **Test templates**: Verify template evaluation with simple examples
- **Check timezones**: Ensure consistent timezone handling
- **Monitor execution**: Track actual vs. expected timing
- **Use logging**: Enable detailed scheduling logs (when available)

## Future Enhancements

### Phase 3 Features

- **Cron expressions**: Full cron syntax support
- **Timezone handling**: Per-request timezone specification
- **Advanced jitter**: Distribution-based jitter algorithms
- **Schedule dependencies**: Request chaining and dependencies

### Phase 4+ Features

- **Dynamic scheduling**: Runtime schedule modifications
- **Conditional execution**: Schedule based on system state
- **Load-based scheduling**: Adaptive intervals based on load
- **Schedule optimization**: Automatic schedule optimization

## Getting Help

- **User Guide**: See `docs/USER_GUIDE.md` for general usage
- **Function Reference**: See `docs/FUNCTIONS.md` for template functions
- **Examples**: Check the `examples/` directory for configuration examples
- **Validation**: Use configuration validation to catch errors early
- **Testing**: Test schedules with `--dry-run` when available
