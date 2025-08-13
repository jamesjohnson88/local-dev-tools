# Documentation Index

Welcome to the Dynamic Request Scheduler documentation! This index will help you find the information you need to get started and become proficient with the tool.

## Getting Started

### 🚀 [User Guide](USER_GUIDE.md)
**Start here if you're new to the tool!**

The comprehensive user guide covers:
- Quick start instructions
- Configuration file structure
- Basic usage examples
- Command line options
- Best practices and troubleshooting

**Perfect for**: New users, first-time setup, understanding the basics

## Reference Documentation

### 📚 [Function Reference](FUNCTIONS.md)
**Complete documentation of all available template functions**

Detailed coverage of:
- Time manipulation functions (`now`, `addMinutes`, `unix`, etc.)
- ID and random generation (`uuid`, `randInt`, `seq`)
- Environment and variable access (`env`, `var`)
- Utility functions (`jitter`, `upper`, `lower`, `trim`)
- Function composition and piping examples

**Perfect for**: Writing templates, understanding function capabilities, troubleshooting

### ⏰ [Scheduling Guide](SCHEDULING.md)
**Deep dive into scheduling strategies and patterns**

Comprehensive coverage of:
- Epoch scheduling (specific timestamps)
- Relative scheduling (duration-based)
- Template-based scheduling (computed times)
- Cron scheduling (coming in Phase 3)
- Jitter and randomization
- Best practices and common patterns

**Perfect for**: Understanding scheduling options, designing complex schedules, troubleshooting timing issues

## Examples and Templates

### 📁 [Examples Directory](../examples/)
**Ready-to-use configuration examples**

Available examples:
- `example-config.yaml` - Basic configuration with various strategies
- `health-check.yaml` - Health monitoring patterns
- `data-sync.yaml` - Data collection and synchronization
- `business-hours.yaml` - Business hours scheduling

**Perfect for**: Learning by example, copying and modifying configurations, understanding patterns

## Development and Contributing

### 🗺️ [Roadmap](../ROADMAP.md)
**Development progress and future plans**

Current status:
- **Phase 0**: ✅ Baseline refactor and scaffolding
- **Phase 1**: ✅ Config-first loading
- **Phase 2**: ✅ Dynamic value representation and evaluation
- **Phase 3**: ⏳ Scheduling strategies (in progress)
- **Phase 4**: ⏳ Execution engine
- **Phase 5**: ⏳ CLI and UX
- **Phase 6**: ⏳ Testing and examples
- **Phase 7**: ⏳ Documentation

**Perfect for**: Contributors, understanding development priorities, feature planning

### 📖 [Main README](../README.md)
**Project overview and quick reference**

High-level information about:
- Features and capabilities
- Quick start guide
- Building and testing
- Architecture overview

**Perfect for**: Project overview, quick reference, understanding the big picture

## Documentation Structure

```
docs/
├── README.md              # This index file
├── USER_GUIDE.md          # Comprehensive user guide
├── FUNCTIONS.md           # Template function reference
└── SCHEDULING.md          # Scheduling strategies guide

examples/
├── example-config.yaml    # Basic configuration
├── health-check.yaml      # Health monitoring
├── data-sync.yaml         # Data synchronization
└── business-hours.yaml    # Business hours scheduling
```

## How to Use This Documentation

### 🆕 **New Users**
1. Start with the [User Guide](USER_GUIDE.md)
2. Look at the [examples](../examples/) for inspiration
3. Refer to the [Function Reference](FUNCTIONS.md) when writing templates
4. Use the [Scheduling Guide](SCHEDULING.md) for timing questions

### 🔧 **Advanced Users**
1. Use the [Function Reference](FUNCTIONS.md) for detailed function information
2. Check the [Scheduling Guide](SCHEDULING.md) for advanced patterns
3. Review the [examples](../examples/) for complex use cases
4. Check the [roadmap](../ROADMAP.md) for upcoming features

### 👨‍💻 **Contributors**
1. Review the [roadmap](../ROADMAP.md) for development priorities
2. Check the [main README](../README.md) for architecture overview
3. Use the [User Guide](USER_GUIDE.md) to understand user needs
4. Reference the [Function Reference](FUNCTIONS.md) for implementation details

## Quick Reference

### Common Template Functions
```yaml
# Time functions
"{{ now | unix }}"                    # Current Unix timestamp
"{{ addMinutes 30 now | unix }}"     # 30 minutes from now
"{{ now | rfc3339 }}"                # Current time in RFC3339 format

# ID generation
"{{ uuid }}"                          # Generate UUID v4
"{{ seq }}"                           # Incremental sequence number

# Random values
"{{ randInt 1 100 }}"                # Random integer 1-100
"{{ randFloat }}"                     # Random float 0.0-1.0

# Environment and variables
"{{ env 'API_TOKEN' }}"              # Environment variable
"{{ var 'user_id' }}"                # User variable
```

### Common Scheduling Patterns
```yaml
# Every minute
schedule:
  relative: "1m"
  jitter: "±10s"

# Every hour
schedule:
  relative: "1h"
  jitter: "±5m"

# Specific time (9 AM)
schedule:
  template: "{{ addHours 9 (parseTime '2006-01-02' (now | rfc3339 | slice 0 10)) | unix }}"

# Fixed timestamp
schedule:
  epoch: 1704067200
```

## Getting Help

### 📖 **Documentation Issues**
- Check that you're using the correct version of the documentation
- Verify that the examples match your version
- Look for any "Coming Soon" or "Phase X" notes

### 🐛 **Configuration Problems**
- Validate your YAML/JSON syntax
- Check template syntax (balanced `{{` and `}}`)
- Verify function names and arguments
- Test templates with simple examples first

### ⚡ **Performance Issues**
- Review scheduling patterns for conflicts
- Check jitter settings for appropriate values
- Monitor template complexity and execution time
- Use seeded random for deterministic results

### 🔮 **Feature Requests**
- Check the [roadmap](../ROADMAP.md) for planned features
- Look for "Coming in Phase X" notes in documentation
- Consider if your use case can be solved with current features

## Documentation Updates

This documentation is updated as the tool evolves:

- **Phase 2**: ✅ Complete (current)
- **Phase 3**: Will add cron scheduling documentation
- **Phase 4**: Will add execution engine documentation
- **Phase 5**: Will add CLI and UX documentation

## Contributing to Documentation

When contributing to the project:

1. **Update user-facing docs** for any user-visible changes
2. **Add examples** for new features and patterns
3. **Update this index** when adding new documentation
4. **Test examples** to ensure they work correctly
5. **Follow the style** of existing documentation

---

**Happy scheduling!** 🎯

If you find any issues with the documentation or have suggestions for improvement, please contribute to the project or open an issue.
