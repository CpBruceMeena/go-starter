package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/CpBruceMeena/go-starter/internal/consumer"
	"github.com/CpBruceMeena/go-starter/internal/logger"
)

// TaskFunc is the function signature for worker tasks
type TaskFunc func(ctx context.Context) error

// Task represents a scheduled task
type Task struct {
	Name     string
	Interval time.Duration
	Fn       TaskFunc
	Timeout  time.Duration
}

// Worker manages background tasks (cron jobs) and message consumers
type Worker struct {
	// Cron tasks
	tasks []Task

	// Message consumers
	consumers []consumer.Consumer

	log     *logger.Logger
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
}

// New creates a new worker instance
func New(log *logger.Logger) *Worker {
	return &Worker{
		tasks:     []Task{},
		consumers: []consumer.Consumer{},
		log:       log,
		stopCh:    make(chan struct{}),
		running:   false,
	}
}

// RegisterTask registers a new scheduled cron task
func (w *Worker) RegisterTask(task Task) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if task.Timeout == 0 {
		task.Timeout = 30 * time.Second // Default timeout
	}

	w.tasks = append(w.tasks, task)
	w.log.Info("task registered", "name", task.Name, "interval", task.Interval.String())
}

// RegisterConsumer registers a new message consumer
func (w *Worker) RegisterConsumer(c consumer.Consumer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.consumers = append(w.consumers, c)
	w.log.Info("consumer registered", "total_consumers", len(w.consumers))
}

// Start starts the worker and all registered tasks and consumers
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker is already running")
	}
	w.running = true
	w.mu.Unlock()

	w.log.Info("starting worker",
		"tasks", len(w.tasks),
		"consumers", len(w.consumers),
	)

	// Start cron tasks
	for _, task := range w.tasks {
		task := task // Capture for closure
		w.wg.Add(1)
		go w.runTask(ctx, task)
	}

	// Start consumers
	for _, c := range w.consumers {
		consumer := c // Capture for closure
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			if err := consumer.Start(ctx); err != nil {
				w.log.Error("consumer error", "error", err.Error())
			}
		}()
	}

	return nil
}

// Stop stops the worker gracefully
func (w *Worker) Stop(ctx context.Context) error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return fmt.Errorf("worker is not running")
	}
	w.running = false
	w.mu.Unlock()

	w.log.Info("stopping worker")

	// Stop all consumers
	for _, c := range w.consumers {
		if err := c.Stop(ctx); err != nil {
			w.log.Error("error stopping consumer", "error", err.Error())
		}
	}

	// Signal all tasks to stop
	close(w.stopCh)

	// Wait for all tasks and consumers to complete with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.log.Info("worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.log.Warn("worker stop timeout exceeded")
		return ctx.Err()
	}
}

// runTask runs a task on a schedule
func (w *Worker) runTask(ctx context.Context, task Task) {
	defer w.wg.Done()

	ticker := time.NewTicker(task.Interval)
	defer ticker.Stop()

	// Run immediately on start
	w.executeTask(ctx, task)

	for {
		select {
		case <-w.stopCh:
			w.log.Info("task stopped", "name", task.Name)
			return
		case <-ticker.C:
			w.executeTask(ctx, task)
		}
	}
}

// executeTask executes a task with timeout and error handling
func (w *Worker) executeTask(ctx context.Context, task Task) {
	// Create context with timeout
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	startTime := time.Now()
	w.log.Info("task started", "name", task.Name)

	err := task.Fn(taskCtx)

	duration := time.Since(startTime)

	if err != nil {
		w.log.Error("task failed",
			"name", task.Name,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
		return
	}

	w.log.Info("task completed",
		"name", task.Name,
		"duration_ms", duration.Milliseconds(),
	)
}

// IsRunning returns whether the worker is running
func (w *Worker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// TaskCount returns the number of registered tasks
func (w *Worker) TaskCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.tasks)
}

// ConsumerCount returns the number of registered consumers
func (w *Worker) ConsumerCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.consumers)
}
