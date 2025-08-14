package spec

import (
	"testing"
	"time"
)

func TestNewScheduleEngine(t *testing.T) {
	engine := NewScheduleEngine()
	if engine == nil {
		t.Fatal("NewScheduleEngine() returned nil")
	}
}

func TestScheduleEngine_ComputeNextRun(t *testing.T) {
	engine := NewScheduleEngine()
	fixedTime := time.Unix(1000, 0)

	tests := []struct {
		name     string
		schedule ScheduleSpec
		want     time.Time
		wantErr  bool
	}{
		{
			name: "epoch schedule",
			schedule: ScheduleSpec{
				Epoch: int64Ptr(2000),
			},
			want: time.Unix(2000, 0),
		},
		{
			name: "relative schedule",
			schedule: ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			want: fixedTime.Add(5 * time.Minute),
		},
		{
			name: "cron schedule",
			schedule: ScheduleSpec{
				Cron: stringPtr("0 * * * *"), // Every hour at minute 0
			},
			wantErr: false, // We'll check it's after the current time
		},
		{
			name: "template schedule without engine",
			schedule: ScheduleSpec{
				Template: stringPtr("{{ now | unix }}"),
			},
			wantErr: true, // Should fail without template engine
		},
		{
			name:     "no schedule",
			schedule: ScheduleSpec{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ComputeNextRun(fixedTime, tt.schedule)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeNextRun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.name == "cron schedule" {
					// For cron, just check it's after the current time
					if result.Before(fixedTime) {
						t.Errorf("Cron schedule result %v is before current time %v", result, fixedTime)
					}
				} else if !result.Equal(tt.want) {
					t.Errorf("ComputeNextRun() = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestScheduleEngine_ComputeNextRunWithTemplate(t *testing.T) {
	engine := NewScheduleEngine()
	fixedTime := time.Unix(1000, 0)
	templateEngine := NewTemplateEngine(&EvaluationContext{
		Clock: &MockClock{now: fixedTime},
	})

	tests := []struct {
		name     string
		schedule ScheduleSpec
		want     time.Time
		wantErr  bool
	}{
		{
			name: "template schedule",
			schedule: ScheduleSpec{
				Template: stringPtr("{{ addMinutes 15 now | unix }}"),
			},
			want: fixedTime.Add(15 * time.Minute),
		},
		{
			name: "template schedule with jitter",
			schedule: ScheduleSpec{
				Template: stringPtr("{{ addMinutes 10 now | unix }}"),
				Jitter:   stringPtr("30s"),
			},
			wantErr: false, // We'll check it's within range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ComputeNextRunWithTemplate(fixedTime, tt.schedule, templateEngine)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeNextRunWithTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.name == "template schedule with jitter" {
					// Check that jitter is within expected range
					baseTime := fixedTime.Add(10 * time.Minute)
					jitterRange := 30 * time.Second
					if result.Before(baseTime) || result.After(baseTime.Add(jitterRange)) {
						t.Errorf("Jittered time %v is outside expected range [%v, %v]",
							result, baseTime, baseTime.Add(jitterRange))
					}
				} else if !result.Equal(tt.want) {
					t.Errorf("ComputeNextRunWithTemplate() = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestScheduleEngine_ValidateSchedule(t *testing.T) {
	engine := NewScheduleEngine()

	tests := []struct {
		name     string
		schedule ScheduleSpec
		wantErr  bool
	}{
		{
			name: "valid epoch schedule",
			schedule: ScheduleSpec{
				Epoch: int64Ptr(1000),
			},
			wantErr: false,
		},
		{
			name: "valid relative schedule",
			schedule: ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			wantErr: false,
		},
		{
			name: "valid cron schedule",
			schedule: ScheduleSpec{
				Cron: stringPtr("*/5 * * * *"),
			},
			wantErr: false,
		},
		{
			name: "valid template schedule",
			schedule: ScheduleSpec{
				Template: stringPtr("{{ now | unix }}"),
			},
			wantErr: false,
		},
		{
			name: "multiple strategies",
			schedule: ScheduleSpec{
				Epoch:    int64Ptr(1000),
				Relative: stringPtr("5m"),
			},
			wantErr: true,
		},
		{
			name:     "no strategy",
			schedule: ScheduleSpec{},
			wantErr:  true,
		},
		{
			name: "invalid relative duration",
			schedule: ScheduleSpec{
				Relative: stringPtr("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid cron expression",
			schedule: ScheduleSpec{
				Cron: stringPtr("invalid cron"),
			},
			wantErr: true,
		},
		{
			name: "negative epoch",
			schedule: ScheduleSpec{
				Epoch: int64Ptr(-1),
			},
			wantErr: true,
		},
		{
			name: "valid schedule with jitter",
			schedule: ScheduleSpec{
				Relative: stringPtr("5m"),
				Jitter:   stringPtr("30s"),
			},
			wantErr: false,
		},
		{
			name: "invalid jitter",
			schedule: ScheduleSpec{
				Relative: stringPtr("5m"),
				Jitter:   stringPtr("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.ValidateSchedule(tt.schedule)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSchedule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScheduleEngine_ApplyJitter(t *testing.T) {
	engine := NewScheduleEngine()
	baseTime := time.Unix(1000, 0)

	tests := []struct {
		name      string
		jitterStr string
		wantRange time.Duration
	}{
		{
			name:      "30 second jitter",
			jitterStr: "30s",
			wantRange: 30 * time.Second,
		},
		{
			name:      "2 minute jitter",
			jitterStr: "2m",
			wantRange: 2 * time.Minute,
		},
		{
			name:      "invalid jitter",
			jitterStr: "invalid",
			wantRange: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.applyJitter(baseTime, tt.jitterStr)

			if tt.wantRange > 0 {
				// Check that result is within expected range
				if result.Before(baseTime) || result.After(baseTime.Add(tt.wantRange)) {
					t.Errorf("Jittered time %v is outside expected range [%v, %v]",
						result, baseTime, baseTime.Add(tt.wantRange))
				}
			} else {
				// For invalid jitter, should return base time unchanged
				if !result.Equal(baseTime) {
					t.Errorf("Invalid jitter should return base time unchanged: %v, want %v", result, baseTime)
				}
			}
		})
	}
}
