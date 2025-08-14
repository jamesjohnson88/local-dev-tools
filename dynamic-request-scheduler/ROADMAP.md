## Dynamic Request Scheduler – Tasks and Decisions

### Progress Summary
- **Phase 0**: ✅ **COMPLETED** - Baseline refactor and scaffolding
- **Phase 1**: ✅ **COMPLETED** - Config-first loading  
- **Phase 2**: ✅ **COMPLETED** - Dynamic value representation and evaluation
- **Phase 3**: ✅ **COMPLETED** - Scheduling strategies
- **Phase 4**: ✅ **COMPLETED** - Execution engine
- **Phase 5**: ✅ **COMPLETED** - CLI and UX
- **Phase 6**: ✅ **COMPLETED** - Testing and examples
- **Phase 7**: ⏳ **PENDING** - Documentation

### Context and goals
- **Goal**: Make `ScheduledRequest` declarative, config-driven, and support dynamic values (time, URL, headers, payload) via templates/expressions, with multiple scheduling strategies.
- **Outcome**: Easily add many request types for testing without code changes; evaluate dynamic fields at runtime; support per-request schedules.

## Phased plan (with tasks)

### Phase 0: Baseline refactor and scaffolding ✅
- [x] Extract and rename core types:
  - [x] Introduce `HttpRequestSpec` (method, url, headers, body)
  - [x] Introduce `ScheduleSpec` (one-of: epoch | relative | expression | cron | jitter)
  - [x] Make `ScheduledRequest` a composition `{ name, schedule, http }`
- [x] Export currently unexported fields in `local-dev-tools/dynamic-request-scheduler/main.go` or move them into new package(s):
  - [x] Create `internal/spec` for type definitions
  - [x] Create `internal/engine` for evaluation and execution
- [x] Remove hardcoded defaults in `NewScheduledRequest`/`NewRequestBody`; defaults will be provided via config or functional options

### Phase 1: Config-first loading ✅
- [x] Decide on config format (YAML recommended, JSON also supported)
- [x] Implement `LoadConfig(path string) ([]ScheduledRequest, error)`
- [x] Validate config: method, URL, headers, schedule one-of, body serializability
- [x] Add `--config <path>` flag in `main` to load requests

### Phase 2: Dynamic value representation and evaluation (COMPLETED ✅)
- [x] Introduce dynamic value types:
  - [x] `DynamicString`, `DynamicInt64`, `DynamicAny` (JSON-unmarshal accepts literal or `{template: "..."}` or `{expr: "..."}`)
- [x] Implement template engine support:
  - [x] Start with Go `text/template` + function map (Sprig-like helpers)
  - [x] Functions: `now`, `unix`, `rfc3339`, `addSeconds`, `addMinutes`, `addHours`, `uuid`, `randInt`, `randFloat`, `env`, `jitter`
- [x] Provide evaluation context and resolver:
  - [x] `Evaluate(spec, ctx) (ResolvedRequest, error)` to resolve all dynamic fields before send
  - [x] Add `--var k=v` to inject user variables into the context
- [x] Optional: expression engine (CEL or `antonmedv/expr`) if we want `dyn.Unix.AddMinutes(11)` style

### Phase 3: Scheduling strategies ✅
- [x] Implement `ScheduleSpec` variants:
  - [x] `epoch: int64` (literal)
  - [x] `relative: duration` (e.g., "11m")
  - [x] `template/expr` (evaluated to epoch)
  - [x] `cron: string` (compute next time)
  - [x] Optional `jitter: duration` (± jitter)
- [x] `ComputeNextRun(now, schedule) (time.Time, error)`
- [x] Validate mutual exclusivity and sensible bounds

### Phase 4: Execution engine ✅
- [x] Replace single global ticker with per-request schedulers
- [x] For each request:
  - [x] Compute next run from `ScheduleSpec`
  - [x] At fire time: resolve dynamic fields → build HTTP request → send → record result → compute next
- [x] Concurrency controls: `--concurrency N` (worker pool), per-request max in-flight
- [x] Retries and backoff: `retries`, `backoff` (exponential with jitter)
- [x] Timeout per request (client-side)
- [x] `--once` and `--limit N` modes

### Phase 5: CLI and UX ✅
- [x] Flags: `--config`, `--scenario <name|regex>`, `--var k=v`, `--dry-run`, `--once`, `--concurrency`, `--limit`, `--log-json`
- [x] Dry run prints fully resolved requests and next run times without sending
- [x] Clear logging (name, resolved URL, status, latency, retries)

### Phase 6: Testing and examples ✅
- [x] Unit tests for:
  - [x] Template function map and dynamic evaluation
  - [x] Schedule computations (epoch, relative, cron, jitter)
  - [x] HTTP build and header/payload resolution
  - [x] Retry/backoff behavior
- [x] Integration smoke test against a local mock server
- [x] Provide `examples/` with YAML configs demonstrating:
  - [x] Literal epoch
  - [x] Relative time
  - [x] Template-based scheduled time
  - [x] Cron definition
  - [x] Dynamic URL/headers/payload

### Phase 7: Documentation
- [ ] `README.md` updates: quickstart, flags, examples
- [ ] `docs/DYNAMIC_VALUES.md`: templating syntax, functions, variables, safety notes
- [ ] `docs/SCHEDULING.md`: strategy matrix, cron details, jitter

## Decisions to make (with recommendations)

### Config format
- Options: YAML, JSON, or both
- Recommendation: **YAML primary**, JSON also supported (use `yaml.v3` with JSON tags)

### Dynamic values: template vs expression
- Options:
  - A) `text/template` + function map (Sprig-like)
  - B) Expression engine (CEL or `antonmedv/expr`)
- Recommendation: **Start with templates** (safer, simpler, battle-tested). Add optional expression engine later if dot-call syntax is desired.

### Where dynamic values are allowed
- Options: Only some fields vs all fields (url, headers, payload, schedule)
- Recommendation: **Allow everywhere**. Resolve just-in-time before sending.

### Schedule strategies
- Options: epoch, relative duration, template/expr to epoch, cron; optional jitter
- Recommendation: **Support all four**; enforce one-of per request; include jitter

### Function set for templates
- Options: minimal vs extended
- Recommendation: **Extended minimal**: `now`, `unix`, `rfc3339`, `addSeconds`, `addMinutes`, `addHours`, `uuid`, `randInt`, `randFloat`, `env`, `jitter`

### Safety and determinism
- Considerations: template sandboxing, environment access, reproducibility
- Recommendation: **No file/network access in templates**; allow `env` opt-in; document determinism limits (rand); support `--seed` for deterministic rand if needed

### Concurrency and retries
- Options: global worker pool vs per-request goroutines
- Recommendation: **Global worker pool** bounded by `--concurrency`; per-request queueing; exponential backoff with jitter

### Timezone and clock source
- Options: system local vs UTC; now() vs injected clock for tests
- Recommendation: **UTC** everywhere; use injectable clock in engine for tests

### Cron implementation
- Options: `robfig/cron/v3` or lightweight next-time calculator
- Recommendation: **robfig/cron/v3** (mature, supports time zones); store TZ per schedule if needed

## Acceptance criteria
- Config can define multiple requests with different schedule strategies
- Dynamic fields resolve correctly across URL, headers, and payload
- `--dry-run` prints resolved requests and next run times
- Engine schedules and sends requests according to spec with concurrency and retries
- Unit and integration tests passing; examples provided

## Example config (YAML)
```yaml
requests:
  - name: "Run once"
    schedule:
      relative: "11m"
      jitter: "±30s"
    http:
      method: "POST"
      url: "https://localhost:10001/core/scheduler/tasks/run-once"
      headers:
        Content-Type: "application/json"
        X-Trace: "{{ uuid }}"
      body:
        task_request_method: "GET"
        task_request_url: "https://localhost:10001/fad/health"
        task_request_headers: null
        task_request_payload: null
        scheduled_for: "{{ now | addMinutes 11 | unix }}"
```

## Proposed package layout
- `local-dev-tools/dynamic-request-scheduler/cmd/` – main CLI
- `local-dev-tools/dynamic-request-scheduler/internal/spec/` – types and validation
- `local-dev-tools/dynamic-request-scheduler/internal/template/` – template funcs and evaluators
- `local-dev-tools/dynamic-request-scheduler/internal/schedule/` – schedule computations (relative, cron, jitter)
- `local-dev-tools/dynamic-request-scheduler/internal/engine/` – execution, concurrency, retries, logging
- `local-dev-tools/dynamic-request-scheduler/examples/` – sample YAMLs
- `local-dev-tools/dynamic-request-scheduler/docs/` – deeper docs

## Open questions
- Do we need persistence of runs (e.g., state across restarts) or is in-memory sufficient?
- Should we support per-request rate limits or global QPS caps?
- Do we need per-request auth helpers (e.g., OAuth2, signed headers) in templates?
- Should we support capturing and reusing previous response data in subsequent requests (scenarios)?


