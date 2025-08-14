# Dynamic Request Scheduler

A flexible, configurable tool for scheduling and sending HTTP requests with dynamic values. Define requests in YAML/JSON configuration files and let the scheduler handle timing, dynamic field resolution, and execution.

## Features

- **Multiple Scheduling Strategies**: Epoch timestamps, relative durations, template-based calculations, and cron expressions (coming soon)
- **Dynamic Values**: Use Go templates to generate UUIDs, timestamps, random values, and more at runtime
- **Flexible Configuration**: YAML or JSON configuration files with validation
- **Template Engine**: Rich function library for time manipulation, ID generation, and data transformation
- **Jitter Support**: Add randomness to schedules to prevent thundering herd problems
- **Environment Integration**: Access environment variables and user-defined variables in templates

## Quick Start

### 1. Create Configuration

Create a `config.yaml` file:

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
        X-Timestamp: "{{ now | unix }}"
      body: null

  - name: "Data Sync"
    schedule:
      relative: "5m"
      jitter: "±30s"
    http:
      method: "POST"
      url: "https://api.example.com/sync"
      headers:
        Content-Type: "application/json"
        X-Request-ID: "{{ uuid }}"
      body:
        sync_id: "{{ uuid }}"
        timestamp: "{{ now | unix }}"
        source: "dynamic-scheduler"
```

### 2. Run the Scheduler

```bash
# Run with configuration file
./dynamic-request-scheduler --config config.yaml

# Run in legacy mode (single request, fixed interval)
./dynamic-request-scheduler --interval 60
```

## Configuration Structure

Each request has three main sections:

- **`name`**: Human-readable identifier for the request
- **`schedule`**: When to run the request (choose one strategy)
- **`http`**: HTTP request details (method, URL, headers, body)

### Scheduling Strategies

| Strategy | Description | Example |
|----------|-------------|---------|
| `epoch` | Specific Unix timestamp | `epoch: 1704067200` |
| `relative` | Duration from now | `relative: "5m"` |
| `template` | Computed time | `template: "{{ addHours 1 now \| unix }}"` |
| `cron` | Cron expression (coming soon) | `cron: "*/5 * * * *"` |

### Dynamic Values

Use Go templates to generate values at runtime:

```yaml
# Time functions
timestamp: "{{ now | unix }}"
future_time: "{{ addMinutes 30 now | unix }}"

# ID generation
request_id: "{{ uuid }}"
sequence: "{{ seq }}"

# Random values
random_delay: "{{ randInt 5 15 }}"
probability: "{{ randFloat }}"

# Environment and variables
api_key: "{{ env 'API_TOKEN' }}"
user_id: "{{ var 'user_id' }}"
```

## Documentation

- **[User Guide](docs/USER_GUIDE.md)**: Comprehensive usage instructions and examples
- **[Function Reference](docs/FUNCTIONS.md)**: Complete template function documentation
- **[Scheduling Guide](docs/SCHEDULING.md)**: Detailed scheduling strategy explanations
- **[Roadmap](ROADMAP.md)**: Development progress and future plans

## Examples

See the `examples/` directory for configuration examples:

- `example-config.yaml` - Basic configuration with various scheduling strategies
- `health-check.yaml` - Simple health check monitoring
- `data-sync.yaml` - Data synchronization patterns
- `business-hours.yaml` - Business hours scheduling

## Command Line Options

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

## Development Status

- **Phase 0**: ✅ Baseline refactor and scaffolding
- **Phase 1**: ✅ Config-first loading
- **Phase 2**: ✅ Dynamic value representation and evaluation
- **Phase 3**: ✅ Scheduling strategies
- **Phase 4**: ✅ Execution engine
- **Phase 5**: ✅ CLI and UX
- **Phase 6**: ✅ Testing and examples
- **Phase 7**: ⏳ Documentation

## Building

### Prerequisites

- Go 1.21 or later

### Build

```bash
go build .
```

### Test

```bash
go test ./internal/spec/...
```

### Run

```bash
./dynamic-request-scheduler --config config.yaml
```

## Architecture

The project is organized into several packages:

- **`cmd/`**: Main CLI application
- **`internal/spec/`**: Type definitions and configuration loading
- **`internal/template/`**: Template engine and function library
- **`internal/schedule/`**: Scheduling logic and computations
- **`internal/engine/`**: Execution engine and HTTP handling

## Contributing

1. Check the [roadmap](ROADMAP.md) for current development priorities
2. Follow Go coding standards and conventions
3. Add tests for new functionality
4. Update documentation for user-facing changes

## License

This project is part of the local development tools collection.

## Support

- Check the documentation in the `docs/` directory
- Review configuration examples in the `examples/` directory
- Examine the [roadmap](ROADMAP.md) for development status
- Run with `--help` for current command line options
