package spec

import (
	"fmt"
	"reflect"
	"time"
)

// Evaluator resolves all dynamic fields in a request specification
type Evaluator struct {
	engine *TemplateEngine
}

// NewEvaluator creates a new evaluator with the given template engine
func NewEvaluator(engine *TemplateEngine) *Evaluator {
	return &Evaluator{engine: engine}
}

// EvaluateRequest resolves all dynamic fields in a ScheduledRequest
func (e *Evaluator) EvaluateRequest(req *ScheduledRequest) (*ResolvedRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	resolved := &ResolvedRequest{
		Name:   req.Name,
		Method: req.HTTP.Method,
		URL:    req.HTTP.URL,
	}

	// Resolve URL if it contains templates
	if IsTemplateString(resolved.URL) {
		resolvedURL, err := e.engine.EvaluateTemplate(resolved.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve URL template: %w", err)
		}
		resolved.URL = resolvedURL
	}

	// Resolve headers
	resolved.Headers = make(map[string]string)
	for key, value := range req.HTTP.Headers {
		resolvedKey := key
		resolvedValue := value

		// Resolve header key if it contains templates
		if IsTemplateString(key) {
			var err error
			resolvedKey, err = e.engine.EvaluateTemplate(key)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve header key template: %w", err)
			}
		}

		// Resolve header value if it contains templates
		if IsTemplateString(value) {
			var err error
			resolvedValue, err = e.engine.EvaluateTemplate(value)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve header value template: %w", err)
			}
		}

		resolved.Headers[resolvedKey] = resolvedValue
	}

	// Resolve body recursively
	if req.HTTP.Body != nil {
		resolvedBody, err := e.resolveValue(req.HTTP.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve body: %w", err)
		}
		resolved.Body = resolvedBody
	}

	// Compute scheduled time from schedule specification
	scheduledTime, err := e.computeScheduledTime(req.Schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to compute scheduled time: %w", err)
	}
	resolved.ScheduledFor = scheduledTime

	return resolved, nil
}

// resolveValue recursively resolves templates in any value
func (e *Evaluator) resolveValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case string:
		if IsTemplateString(val) {
			return e.engine.EvaluateTemplate(val)
		}
		return val, nil

	case map[string]interface{}:
		resolved := make(map[string]interface{})
		for key, value := range val {
			resolvedKey := key
			resolvedValue := value

			// Resolve key if it contains templates
			if IsTemplateString(key) {
				resolvedKeyStr, err := e.engine.EvaluateTemplate(key)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve map key template: %w", err)
				}
				resolvedKey = resolvedKeyStr
			}

			// Resolve value recursively
			resolvedValue, err := e.resolveValue(value)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve map value: %w", err)
			}

			resolved[resolvedKey] = resolvedValue
		}
		return resolved, nil

	case []interface{}:
		resolved := make([]interface{}, len(val))
		for i, item := range val {
			resolvedItem, err := e.resolveValue(item)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve array item %d: %w", i, err)
			}
			resolved[i] = resolvedItem
		}
		return resolved, nil

	case DynamicString:
		if val.IsTemplate() {
			return e.engine.EvaluateTemplate(val.GetTemplate())
		}
		return val.GetValue(), nil

	case DynamicInt64:
		if val.IsTemplate() {
			return e.engine.EvaluateTemplateToInt64(val.GetTemplate())
		}
		return val.GetValue(), nil

	case DynamicAny:
		if val.IsTemplate() {
			return e.engine.EvaluateTemplate(val.GetTemplate())
		}
		return val.GetValue(), nil

	default:
		// For other types, try to use reflection to handle nested structs
		return e.resolveReflectedValue(v)
	}
}

// resolveReflectedValue uses reflection to resolve templates in struct fields
func (e *Evaluator) resolveReflectedValue(v interface{}) (interface{}, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return v, nil
	}

	// Create a new instance of the same type
	result := reflect.New(val.Type()).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the field value
		var fieldValue interface{}
		if field.CanInterface() {
			fieldValue = field.Interface()
		} else {
			// For unexported fields, try to get the value safely
			switch field.Kind() {
			case reflect.String:
				fieldValue = field.String()
			case reflect.Int, reflect.Int64:
				fieldValue = field.Int()
			case reflect.Bool:
				fieldValue = field.Bool()
			default:
				fieldValue = field.Interface()
			}
		}

		// Resolve the field value
		resolvedValue, err := e.resolveValue(fieldValue)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve field %s: %w", fieldType.Name, err)
		}

		// Set the resolved value
		if resolvedValue != nil {
			fieldResult := reflect.ValueOf(resolvedValue)
			if fieldResult.Type().AssignableTo(field.Type()) {
				result.Field(i).Set(fieldResult)
			} else {
				// Try to convert the type
				if fieldResult.CanConvert(field.Type()) {
					result.Field(i).Set(fieldResult.Convert(field.Type()))
				}
			}
		}
	}

	return result.Interface(), nil
}

// computeScheduledTime computes the actual scheduled time from a ScheduleSpec
func (e *Evaluator) computeScheduledTime(schedule ScheduleSpec) (time.Time, error) {
	now := e.engine.ctx.Clock.Now()

	switch {
	case schedule.Epoch != nil:
		return time.Unix(*schedule.Epoch, 0), nil

	case schedule.Relative != nil:
		duration, err := time.ParseDuration(*schedule.Relative)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative duration '%s': %w", *schedule.Relative, err)
		}
		scheduledTime := now.Add(duration)

		// Apply jitter if specified
		if schedule.Jitter != nil {
			scheduledTime = e.engine.jitter(scheduledTime, *schedule.Jitter)
		}

		return scheduledTime, nil

	case schedule.Template != nil:
		epoch, err := e.engine.EvaluateTemplateToInt64(*schedule.Template)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to evaluate schedule template: %w", err)
		}
		scheduledTime := time.Unix(epoch, 0)

		// Apply jitter if specified
		if schedule.Jitter != nil {
			scheduledTime = e.engine.jitter(scheduledTime, *schedule.Jitter)
		}

		return scheduledTime, nil

	case schedule.Cron != nil:
		// TODO: Implement cron parsing and next run calculation
		// For now, return an error
		return time.Time{}, fmt.Errorf("cron scheduling not yet implemented")

	default:
		return time.Time{}, fmt.Errorf("no valid schedule strategy found")
	}
}

// SetVariable sets a variable in the template engine context
func (e *Evaluator) SetVariable(key string, value interface{}) {
	e.engine.SetVariable(key, value)
}

// SetSeed sets the seed for deterministic random functions
func (e *Evaluator) SetSeed(seed int64) {
	e.engine.SetSeed(seed)
}
