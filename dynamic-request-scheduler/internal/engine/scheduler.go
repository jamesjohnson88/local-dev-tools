package engine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

// Scheduler manages request execution
type Scheduler struct {
	requests    []spec.ScheduledRequest
	workers     int
	concurrency int
	once        bool
	dryRun      bool
	httpClient  *HTTPClient
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.Mutex
	running     bool
}

// SchedulerConfig holds configuration for the scheduler
type SchedulerConfig struct {
	Workers     int
	Concurrency int
	Once        bool
	DryRun      bool
	Timeout     time.Duration
}

// NewScheduler creates a new scheduler with the given configuration
func NewScheduler(requests []spec.ScheduledRequest, config SchedulerConfig) *Scheduler {
	if config.Workers <= 0 {
		config.Workers = 1
	}
	if config.Concurrency <= 0 {
		config.Concurrency = 10
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		requests:    requests,
		workers:     config.Workers,
		concurrency: config.Concurrency,
		once:        config.Once,
		dryRun:      config.DryRun,
		httpClient:  NewHTTPClient(config.Timeout),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the scheduling loop
func (s *Scheduler) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Printf("Starting scheduler with %d requests, %d workers, concurrency: %d",
		len(s.requests), s.workers, s.concurrency)

	if s.dryRun {
		return s.runDryRun()
	}

	if s.once {
		return s.runOnce()
	}

	return s.runContinuous()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		log.Println("Stopping scheduler...")
		s.cancel()
		s.running = false
	}
}

// runDryRun shows what would be executed without actually running
func (s *Scheduler) runDryRun() error {
	log.Println("DRY RUN MODE - No requests will be sent")

	evaluator := spec.NewEvaluator(spec.NewTemplateEngine(&spec.EvaluationContext{
		Variables: make(map[string]interface{}),
		Clock:     &spec.RealClock{},
	}))

	for _, req := range s.requests {
		resolved, err := evaluator.EvaluateRequest(&req)
		if err != nil {
			log.Printf("Error evaluating request '%s': %v", req.Name, err)
			continue
		}

		log.Printf("Request: %s", resolved.Name)
		log.Printf("  Method: %s", resolved.Method)
		log.Printf("  URL: %s", resolved.URL)
		log.Printf("  Scheduled for: %s", resolved.ScheduledFor.Format(time.RFC3339))
		log.Printf("  Headers: %v", resolved.Headers)
		if resolved.Body != nil {
			log.Printf("  Body: %v", resolved.Body)
		}
		log.Println()
	}

	return nil
}

// runOnce executes all requests once and exits
func (s *Scheduler) runOnce() error {
	log.Println("Running all requests once...")

	evaluator := spec.NewEvaluator(spec.NewTemplateEngine(&spec.EvaluationContext{
		Variables: make(map[string]interface{}),
		Clock:     &spec.RealClock{},
	}))

	// Create a worker pool for concurrent execution
	semaphore := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup

	for _, req := range s.requests {
		wg.Add(1)
		go func(request spec.ScheduledRequest) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Evaluate and execute request
			s.executeRequest(&request, evaluator)
		}(req)
	}

	wg.Wait()
	log.Println("All requests completed")
	return nil
}

// runContinuous runs the scheduler continuously
func (s *Scheduler) runContinuous() error {
	log.Println("Starting continuous scheduling...")

	// Create evaluator with context
	evaluator := spec.NewEvaluator(spec.NewTemplateEngine(&spec.EvaluationContext{
		Variables: make(map[string]interface{}),
		Clock:     &spec.RealClock{},
	}))

	// Create a worker pool for concurrent execution
	semaphore := make(chan struct{}, s.concurrency)

	// Start worker goroutines
	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(i, evaluator, semaphore)
	}

	// Wait for context cancellation
	<-s.ctx.Done()

	// Wait for all workers to finish
	s.wg.Wait()

	log.Println("Scheduler stopped")
	return nil
}

// worker runs in a loop, processing scheduled requests
func (s *Scheduler) worker(id int, evaluator *spec.Evaluator, semaphore chan struct{}) {
	defer s.wg.Done()

	log.Printf("Worker %d started", id)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Worker %d stopping", id)
			return
		default:
			// Process all requests
			for _, req := range s.requests {
				select {
				case <-s.ctx.Done():
					return
				default:
					// Check if it's time to run this request
					if s.shouldRunRequest(&req, evaluator) {
						// Acquire semaphore for concurrency control
						semaphore <- struct{}{}

						// Execute request in a goroutine to allow concurrent execution
						go func(request spec.ScheduledRequest) {
							defer func() { <-semaphore }()
							s.executeRequest(&request, evaluator)
						}(req)
					}
				}
			}

			// Sleep before next iteration
			time.Sleep(1 * time.Second)
		}
	}
}

// shouldRunRequest determines if a request should be executed now
func (s *Scheduler) shouldRunRequest(req *spec.ScheduledRequest, evaluator *spec.Evaluator) bool {
	// For now, we'll use a simple approach: run relative schedules immediately
	// In a full implementation, this would track last run times and compute next runs

	if req.Schedule.Relative != nil {
		// For relative schedules, we'll run them immediately for now
		// TODO: Implement proper scheduling logic with last run tracking
		return true
	}

	if req.Schedule.Epoch != nil {
		// For epoch schedules, check if it's time
		now := time.Now().Unix()
		return *req.Schedule.Epoch <= now
	}

	// For template and cron schedules, we need more sophisticated logic
	// TODO: Implement proper scheduling for these types
	return false
}

// executeRequest evaluates and executes a single request
func (s *Scheduler) executeRequest(req *spec.ScheduledRequest, evaluator *spec.Evaluator) {
	start := time.Now()

	// Evaluate the request
	resolved, err := evaluator.EvaluateRequest(req)
	if err != nil {
		log.Printf("Error evaluating request '%s': %v", req.Name, err)
		return
	}

	log.Printf("Executing request '%s' at %s", resolved.Name, start.Format(time.RFC3339))

	// Execute the HTTP request
	status, duration, err := s.sendHTTPRequest(resolved)

	if err != nil {
		log.Printf("Request '%s' failed: %v (duration: %v)", resolved.Name, err, duration)
	} else {
		log.Printf("Request '%s' completed: %s (duration: %v)", resolved.Name, status, duration)
	}
}

// sendHTTPRequest sends an HTTP request and returns status, duration, and error
func (s *Scheduler) sendHTTPRequest(resolved *spec.ResolvedRequest) (string, time.Duration, error) {
	resp, err := s.httpClient.SendRequest(resolved)
	if err != nil {
		return "", 0, err
	}

	return resp.Status, resp.Duration, nil
}
