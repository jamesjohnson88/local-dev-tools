package spec

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// ScheduleEngine handles schedule computations
type ScheduleEngine struct {
	cronParser cron.Parser
}

// NewScheduleEngine creates a new schedule engine
func NewScheduleEngine() *ScheduleEngine {
	return &ScheduleEngine{
		cronParser: cron.NewParser(
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		),
	}
}

// ComputeNextRun calculates the next execution time for a schedule
func (s *ScheduleEngine) ComputeNextRun(now time.Time, schedule ScheduleSpec) (time.Time, error) {
	var baseTime time.Time

	switch {
	case schedule.Epoch != nil:
		// Epoch scheduling - run at specific Unix timestamp
		baseTime = time.Unix(*schedule.Epoch, 0)
		// If the epoch time is in the past, return it as-is (will run immediately)
		if baseTime.Before(now) {
			return baseTime, nil
		}

	case schedule.Relative != nil:
		// Relative scheduling - run after specified duration
		duration, err := time.ParseDuration(*schedule.Relative)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative duration '%s': %w", *schedule.Relative, err)
		}
		baseTime = now.Add(duration)

	case schedule.Template != nil:
		// Template scheduling - evaluate template to get Unix timestamp
		// Note: This requires a template engine, so we'll return an error
		// The actual evaluation should be done in the evaluator
		return time.Time{}, fmt.Errorf("template scheduling requires template evaluation context")

	case schedule.Cron != nil:
		// Cron scheduling - parse cron expression and find next run
		cronSchedule, err := s.cronParser.Parse(*schedule.Cron)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid cron expression '%s': %w", *schedule.Cron, err)
		}
		baseTime = cronSchedule.Next(now)

	default:
		return time.Time{}, fmt.Errorf("no valid schedule strategy found")
	}

	// Apply jitter if specified
	if schedule.Jitter != nil {
		baseTime = s.applyJitter(baseTime, *schedule.Jitter)
	}

	return baseTime, nil
}

// ComputeNextRunWithTemplate calculates the next execution time for a schedule with template evaluation
func (s *ScheduleEngine) ComputeNextRunWithTemplate(now time.Time, schedule ScheduleSpec, templateEngine *TemplateEngine) (time.Time, error) {
	var baseTime time.Time

	switch {
	case schedule.Epoch != nil:
		// Epoch scheduling - run at specific Unix timestamp
		baseTime = time.Unix(*schedule.Epoch, 0)
		// If the epoch time is in the past, return it as-is (will run immediately)
		if baseTime.Before(now) {
			return baseTime, nil
		}

	case schedule.Relative != nil:
		// Relative scheduling - run after specified duration
		duration, err := time.ParseDuration(*schedule.Relative)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative duration '%s': %w", *schedule.Relative, err)
		}
		baseTime = now.Add(duration)

	case schedule.Template != nil:
		// Template scheduling - evaluate template to get Unix timestamp
		epoch, err := templateEngine.EvaluateTemplateToInt64(*schedule.Template)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to evaluate schedule template: %w", err)
		}
		baseTime = time.Unix(epoch, 0)

	case schedule.Cron != nil:
		// Cron scheduling - parse cron expression and find next run
		cronSchedule, err := s.cronParser.Parse(*schedule.Cron)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid cron expression '%s': %w", *schedule.Cron, err)
		}
		baseTime = cronSchedule.Next(now)

	default:
		return time.Time{}, fmt.Errorf("no valid schedule strategy found")
	}

	// Apply jitter if specified
	if schedule.Jitter != nil {
		baseTime = s.applyJitter(baseTime, *schedule.Jitter)
	}

	return baseTime, nil
}

// applyJitter adds random variation to the scheduled time
func (s *ScheduleEngine) applyJitter(baseTime time.Time, jitterStr string) time.Time {
	var duration time.Duration
	var err error

	// Parse jitter format: "±30s", "±2m", etc.
	if len(jitterStr) > 1 && (jitterStr[0] == '±' || jitterStr[0] == '+' || jitterStr[0] == '-') {
		// Handle ± symbol and +/- prefixes
		if jitterStr[0] == '±' {
			duration, err = time.ParseDuration(jitterStr[1:])
		} else {
			duration, err = time.ParseDuration(jitterStr)
		}
	} else {
		duration, err = time.ParseDuration(jitterStr)
	}

	if err != nil {
		// If jitter parsing fails, return base time unchanged
		return baseTime
	}

	// Add random jitter within the duration range
	// For now, use a simple random approach - this could be enhanced with seeded random
	jitterNanos := duration.Nanoseconds()
	if jitterNanos > 0 {
		// Use time-based random for now - could be enhanced with seeded random
		randomJitter := time.Duration(time.Now().UnixNano() % jitterNanos)
		baseTime = baseTime.Add(randomJitter)
	}

	return baseTime
}

// ValidateSchedule validates a schedule specification
func (s *ScheduleEngine) ValidateSchedule(schedule ScheduleSpec) error {
	// Check mutual exclusivity
	count := 0
	if schedule.Epoch != nil {
		count++
	}
	if schedule.Relative != nil {
		count++
	}
	if schedule.Template != nil {
		count++
	}
	if schedule.Cron != nil {
		count++
	}

	if count != 1 {
		return fmt.Errorf("exactly one schedule strategy must be specified (epoch, relative, template, or cron)")
	}

	// Validate specific strategies
	if schedule.Epoch != nil {
		if *schedule.Epoch < 0 {
			return fmt.Errorf("epoch timestamp must be non-negative")
		}
	}

	if schedule.Relative != nil {
		if _, err := time.ParseDuration(*schedule.Relative); err != nil {
			return fmt.Errorf("invalid relative duration '%s': %w", *schedule.Relative, err)
		}
	}

	if schedule.Cron != nil {
		if _, err := s.cronParser.Parse(*schedule.Cron); err != nil {
			return fmt.Errorf("invalid cron expression '%s': %w", *schedule.Cron, err)
		}
	}

	// Validate jitter if specified
	if schedule.Jitter != nil {
		jitterStr := *schedule.Jitter
		var err error

		// Parse jitter format: "±30s", "±2m", etc.
		if len(jitterStr) > 1 && (jitterStr[0] == '±' || jitterStr[0] == '+' || jitterStr[0] == '-') {
			// Handle ± symbol and +/- prefixes
			if jitterStr[0] == '±' {
				_, err = time.ParseDuration(jitterStr[1:])
			} else {
				_, err = time.ParseDuration(jitterStr)
			}
		} else {
			_, err = time.ParseDuration(jitterStr)
		}

		if err != nil {
			return fmt.Errorf("invalid jitter duration '%s': %w", *schedule.Jitter, err)
		}
	}

	return nil
}
