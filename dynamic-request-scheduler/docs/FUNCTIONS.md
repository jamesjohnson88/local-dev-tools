# Template Functions Reference

This document provides a comprehensive reference for all available template functions in the Dynamic Request Scheduler.

## Function Categories

- [Time Functions](#time-functions)
- [ID and Random Functions](#id-and-random-functions)
- [Environment and Variables](#environment-and-variables)
- [Sequence and Iteration](#sequence-and-iteration)
- [Utility Functions](#utility-functions)

## Time Functions

### `now`

Returns the current time as a `time.Time` value.

**Signature:** `now() time.Time`

**Example:**
```yaml
# Get current time
timestamp: "{{ now }}"

# Use in time calculations
next_hour: "{{ addHours 1 now }}"
```

**Common Use Cases:**
- Base time for calculations
- Current timestamp in requests
- Reference point for scheduling

### `unix`

Converts a time value to Unix timestamp (seconds since epoch).

**Signature:** `unix(t time.Time) int64`

**Example:**
```yaml
# Current Unix timestamp
current_time: "{{ now | unix }}"

# Specific time as Unix timestamp
scheduled_time: "{{ addMinutes 30 now | unix }}"
```

**Common Use Cases:**
- API endpoints expecting Unix timestamps
- Database timestamp fields
- Logging and monitoring

### `rfc3339`

Formats a time value as RFC3339 string.

**Signature:** `rfc3339(t time.Time) string`

**Example:**
```yaml
# Current time in RFC3339 format
created_at: "{{ now | rfc3339 }}"

# Future time in RFC3339 format
expires_at: "{{ addHours 24 now | rfc3339 }}"
```

**Common Use Cases:**
- JSON API requests
- ISO 8601 compliant timestamps
- Human-readable time formats

### `addSeconds`

Adds a specified number of seconds to a time value.

**Signature:** `addSeconds(seconds int, t time.Time) time.Time`

**Example:**
```yaml
# 30 seconds from now
short_delay: "{{ addSeconds 30 now | unix }}"

# 5 seconds from specific time
adjusted_time: "{{ addSeconds 5 (parseTime "2006-01-02" "2024-01-01") | unix }}"
```

**Common Use Cases:**
- Short delays and timeouts
- Precise time adjustments
- Rate limiting calculations

### `addMinutes`

Adds a specified number of minutes to a time value.

**Signature:** `addMinutes(minutes int, t time.Time) time.Time`

**Example:**
```yaml
# 15 minutes from now
quarter_hour: "{{ addMinutes 15 now | unix }}"

# 2.5 hours from now (150 minutes)
half_day: "{{ addMinutes 150 now | unix }}"
```

**Common Use Cases:**
- Common scheduling intervals
- Business hour calculations
- Cache expiration times

### `addHours`

Adds a specified number of hours to a time value.

**Signature:** `addHours(hours int, t time.Time) time.Time`

**Example:**
```yaml
# 1 hour from now
next_hour: "{{ addHours 1 now | unix }}"

# 12 hours from now
half_day: "{{ addHours 12 now | unix }}"
```

**Common Use Cases:**
- Daily scheduling patterns
- Business day calculations
- Long-term planning

### `parseTime`

Parses a time string according to a layout.

**Signature:** `parseTime(layout, value string) (time.Time, error)`

**Example:**
```yaml
# Parse specific date
start_date: "{{ parseTime "2006-01-02" "2024-01-01" | unix }}"

# Parse date and time
event_time: "{{ parseTime "2006-01-02 15:04:05" "2024-01-01 14:30:00" | unix }}"
```

**Common Use Cases:**
- Fixed schedule dates
- Configuration-based timing
- Historical data references

**Note:** This function can return an error if parsing fails. Use with caution in templates.

## ID and Random Functions

### `uuid`

Generates a UUID v4 string.

**Signature:** `uuid() string`

**Example:**
```yaml
# Generate unique identifier
request_id: "{{ uuid }}"

# Use in headers
X-Request-ID: "{{ uuid }}"
```

**Common Use Cases:**
- Unique request identifiers
- Trace and correlation IDs
- Database record IDs

**Implementation Details:**
- Uses `crypto/rand` for cryptographically secure randomness
- Falls back to timestamp-based ID if crypto/rand fails
- Format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### `randInt`

Generates a random integer within a specified range.

**Signature:** `randInt(min, max int) int`

**Example:**
```yaml
# Random number between 1 and 100
random_value: "{{ randInt 1 100 }}"

# Random delay between 5 and 15 seconds
delay: "{{ randInt 5 15 }}"
```

**Common Use Cases:**
- Randomized delays and jitter
- Load testing with varied parameters
- A/B testing scenarios

**Behavior:**
- Returns `min` if `min >= max`
- Uses seeded random source if seed is set (deterministic)
- Falls back to time-based random if no seed

### `randFloat`

Generates a random float between 0.0 and 1.0.

**Signature:** `randFloat() float64`

**Example:**
```yaml
# Random probability
probability: "{{ randFloat }}"

# Random percentage (0-100)
percentage: "{{ mul 100 (randFloat) }}"
```

**Common Use Cases:**
- Probability-based decisions
- Randomized sampling
- Load distribution

**Behavior:**
- Returns value in range [0.0, 1.0)
- Uses seeded random source if seed is set (deterministic)
- Falls back to time-based random if no seed

## Environment and Variables

### `env`

Retrieves an environment variable value.

**Signature:** `env(key string) string`

**Example:**
```yaml
# Get API key from environment
Authorization: "Bearer {{ env 'API_TOKEN' }}"

# Get configuration from environment
base_url: "{{ env 'BASE_URL' }}"
```

**Common Use Cases:**
- API credentials and tokens
- Configuration values
- Environment-specific settings

**Security Considerations:**
- Environment variables are visible in process lists
- Use for non-sensitive configuration
- Consider using `--var` for sensitive data

### `var`

Retrieves a user-defined variable value.

**Signature:** `var(key string) interface{}`

**Example:**
```yaml
# Reference user variable
user_id: "{{ var 'user_id' }}"

# Use in headers
X-User-ID: "{{ var 'user_id' }}"
```

**Common Use Cases:**
- User-provided parameters
- Runtime configuration
- Test-specific values

**Usage:**
```bash
./dynamic-request-scheduler --config config.yaml --var "user_id=123" --var "api_key=secret"
```

## Sequence and Iteration

### `seq`

Returns an incremental sequence number.

**Signature:** `seq() int64`

**Example:**
```yaml
# Incremental counter
iteration: "{{ seq }}"

# Use in metadata
metadata:
  sequence: "{{ seq }}"
  source: "dynamic-scheduler"
```

**Common Use Cases:**
- Request numbering
- Iteration tracking
- Unique sequence identifiers

**Behavior:**
- Starts at 1 and increments with each call
- Resets when new evaluator is created
- Useful for tracking request order

## Utility Functions

### `jitter`

Adds random jitter to a time value within a specified duration range.

**Signature:** `jitter(base time.Time, duration string) time.Time`

**Example:**
```yaml
# Add jitter to scheduled time
schedule:
  relative: "5m"
  jitter: "±30s"

# Manual jitter calculation
jittered_time: "{{ jitter now '1m' | unix }}"
```

**Common Use Cases:**
- Preventing thundering herd problems
- Distributed system coordination
- Load balancing

**Duration Format:**
- Supports Go duration syntax: "30s", "1m", "2h"
- Jitter is applied within the specified range
- Uses seeded random if available

### `upper`

Converts a string to uppercase.

**Signature:** `upper(s string) string`

**Example:**
```yaml
# Convert to uppercase
method: "{{ upper 'get' }}"

# Dynamic case conversion
header_value: "{{ upper 'hello world' }}"
```

**Common Use Cases:**
- HTTP method normalization
- Header value formatting
- Case-insensitive comparisons

### `lower`

Converts a string to lowercase.

**Signature:** `lower(s string) string`

**Example:**
```yaml
# Convert to lowercase
content_type: "{{ lower 'APPLICATION/JSON' }}"

# Normalize header names
header_name: "{{ lower 'Content-Type' }}"
```

**Common Use Cases:**
- Header normalization
- Case-insensitive matching
- Data standardization

### `trim`

Removes leading and trailing whitespace from a string.

**Signature:** `trim(s string) string`

**Example:**
```yaml
# Clean up whitespace
clean_value: "{{ trim '  hello world  ' }}"

# Normalize user input
user_input: "{{ trim .Variables.user_name }}"
```

**Common Use Cases:**
- Data cleaning and normalization
- User input processing
- String comparison

## Function Composition and Piping

### Basic Piping

Go templates support piping results from one function to another:

```yaml
# Chain multiple functions
timestamp: "{{ now | addMinutes 30 | unix }}"

# Complex calculations
future_time: "{{ addHours 2 (addMinutes 30 now) | unix }}"
```

### Function Order

When using pipes, the result of the left function becomes the **last** argument to the right function:

```yaml
# Correct: addMinutes takes (minutes, time) arguments
"{{ addMinutes 5 now | unix }}"

# Incorrect: this won't work as expected
"{{ now | addMinutes 5 | unix }}"
```

### Nested Function Calls

For complex calculations, use parentheses to control evaluation order:

```yaml
# Complex time calculation
scheduled_time: "{{ addMinutes 15 (addHours 2 now) | unix }}"

# Multiple operations
final_time: "{{ addSeconds 30 (addMinutes 45 (addHours 1 now)) | unix }}"
```

## Error Handling

### Template Execution Errors

Templates that fail to execute will cause the entire request evaluation to fail:

```yaml
# This will fail if 'invalid_function' doesn't exist
bad_template: "{{ invalid_function }}"

# This will fail if parsing fails
bad_time: "{{ parseTime 'invalid' 'format' }}"
```

### Safe Template Design

Design templates to be robust:

```yaml
# Good: Simple, clear template
timestamp: "{{ now | unix }}"

# Better: Add error handling when possible
timestamp: "{{ now | unix }}"
fallback: "{{ now | unix }}"
```

## Performance Considerations

### Function Efficiency

- **Time functions**: Very fast, minimal overhead
- **Random functions**: Fast with seeded source, slower with time-based fallback
- **UUID generation**: Fast with crypto/rand, very fast with fallback
- **String functions**: Very fast, minimal overhead

### Template Complexity

- Keep templates simple and readable
- Avoid deeply nested function calls
- Use intermediate variables for complex calculations when possible

## Best Practices

### 1. Function Selection

- Use `now` as the base for time calculations
- Use `unix` for API compatibility
- Use `rfc3339` for human-readable formats
- Use `uuid` for unique identifiers

### 2. Template Design

- Keep templates simple and readable
- Use meaningful variable names
- Document complex calculations
- Test templates with various inputs

### 3. Error Prevention

- Validate template syntax before deployment
- Use `--dry-run` to test templates (when available)
- Provide fallback values where appropriate
- Test edge cases and boundary conditions

### 4. Performance

- Avoid unnecessary function calls
- Use seeded random for deterministic results
- Cache complex calculations when possible
- Monitor template execution time

## Examples by Use Case

### Health Check Endpoints

```yaml
requests:
  - name: "Health Check"
    schedule:
      relative: "1m"
    http:
      method: "GET"
      url: "https://api.example.com/health"
      headers:
        User-Agent: "HealthChecker/{{ seq }}"
        X-Timestamp: "{{ now | unix }}"
      body: null
```

### Data Collection

```yaml
requests:
  - name: "Data Collection"
    schedule:
      relative: "5m"
    http:
      method: "POST"
      url: "https://api.example.com/data"
      headers:
        Content-Type: "application/json"
        X-Request-ID: "{{ uuid }}"
        X-Timestamp: "{{ now | rfc3339 }}"
      body:
        collection_id: "{{ uuid }}"
        timestamp: "{{ now | unix }}"
        sequence: "{{ seq }}"
        metadata:
          source: "dynamic-scheduler"
          version: "1.0"
```

### Scheduled Tasks

```yaml
requests:
  - name: "Daily Cleanup"
    schedule:
      template: "{{ addHours 24 now | unix }}"
      jitter: "±5m"
    http:
      method: "POST"
      url: "https://api.example.com/tasks/cleanup"
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer {{ env 'CLEANUP_TOKEN' }}"
      body:
        task_name: "daily-cleanup"
        scheduled_for: "{{ now | unix }}"
        parameters:
          batch_size: "{{ randInt 100 1000 }}"
          priority: "low"
```

## Troubleshooting

### Common Issues

1. **Function Not Found**
   - Check function name spelling
   - Verify function exists in current version
   - Check function signature and arguments

2. **Template Syntax Errors**
   - Ensure balanced `{{` and `}}`
   - Check pipe syntax and spacing
   - Validate function argument order

3. **Unexpected Results**
   - Verify function behavior with simple examples
   - Check argument types and ranges
   - Test with known inputs

4. **Performance Issues**
   - Profile template execution time
   - Simplify complex templates
   - Use appropriate random sources

### Debug Tips

- Start with simple templates
- Test functions individually
- Use `--dry-run` when available
- Check function documentation
- Validate configuration files
