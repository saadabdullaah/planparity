package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/saadabdullaah/planparity/internal/model"
)

// JobStatus is an in-memory analysis lifecycle state.
type JobStatus string

const (
	JobQueued    JobStatus = "QUEUED"
	JobRunning   JobStatus = "RUNNING"
	JobCompleted JobStatus = "COMPLETED"
	JobFailed    JobStatus = "FAILED"
	JobCancelled JobStatus = "CANCELLED"
)

// JobEvent is replayable progress for the SSE endpoint.
type JobEvent struct {
	ID      int64     `json:"id"`
	Type    string    `json:"type"`
	Message string    `json:"message,omitempty"`
	At      time.Time `json:"at"`
}

// JobSnapshot is the public status view; results are fetched separately.
type JobSnapshot struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind"`
	Status      JobStatus            `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	StartedAt   *time.Time           `json:"started_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	LastEvent   *JobEvent            `json:"last_event,omitempty"`
	Failure     *model.PublicFailure `json:"failure,omitempty"`
}

// JobFunc performs analysis and emits public progress events.
type JobFunc func(context.Context, func(string, string)) (any, error)

type job struct {
	mu          sync.RWMutex
	snapshot    JobSnapshot
	events      []JobEvent
	result      any
	cancel      context.CancelFunc
	subscribers map[chan JobEvent]struct{}
}

// JobManager owns memory-resident jobs and results for one server process.
type JobManager struct {
	mu      sync.RWMutex
	jobs    map[string]*job
	ctx     context.Context
	cancel  context.CancelFunc
	maxJobs int
}

// NewJobManager creates a process-local manager.
func NewJobManager(parent context.Context, maxJobs int) *JobManager {
	if maxJobs <= 0 {
		maxJobs = 100
	}
	ctx, cancel := context.WithCancel(parent)
	return &JobManager{jobs: make(map[string]*job), ctx: ctx, cancel: cancel, maxJobs: maxJobs}
}

// Start queues a job and returns its unguessable ID.
func (manager *JobManager) Start(kind string, function JobFunc) (string, error) {
	id, err := randomID(16)
	if err != nil {
		return "", err
	}
	manager.mu.Lock()
	manager.pruneLocked()
	ctx, cancel := context.WithCancel(manager.ctx)
	item := &job{
		snapshot: JobSnapshot{ID: id, Kind: kind, Status: JobQueued, CreatedAt: time.Now().UTC()},
		cancel:   cancel, subscribers: make(map[chan JobEvent]struct{}), events: []JobEvent{},
	}
	manager.jobs[id] = item
	manager.mu.Unlock()
	go item.run(ctx, function)
	return id, nil
}

func (item *job) run(ctx context.Context, function JobFunc) {
	now := time.Now().UTC()
	item.mu.Lock()
	item.snapshot.Status = JobRunning
	item.snapshot.StartedAt = &now
	item.mu.Unlock()
	item.emit("job.started", "Analysis started")
	result, err := function(ctx, item.emit)
	status := JobCompleted
	var failure *model.PublicFailure
	switch {
	case errors.Is(ctx.Err(), context.Canceled):
		status = JobCancelled
		failure = &model.PublicFailure{Code: "JOB_CANCELLED", Message: "Analysis was cancelled."}
	case err != nil:
		status = JobFailed
		failure = &model.PublicFailure{Code: "JOB_FAILED", Message: "Analysis failed without exposing engine or state detail."}
	}
	item.finish(status, result, failure)
}

func (item *job) emit(eventType, message string) {
	item.mu.Lock()
	event := JobEvent{ID: int64(len(item.events) + 1), Type: eventType, Message: message, At: time.Now().UTC()}
	item.events = append(item.events, event)
	item.snapshot.LastEvent = &event
	for subscriber := range item.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
	item.mu.Unlock()
}

// Get returns a status snapshot.
func (manager *JobManager) Get(id string) (JobSnapshot, bool) {
	item, exists := manager.job(id)
	if !exists {
		return JobSnapshot{}, false
	}
	item.mu.RLock()
	defer item.mu.RUnlock()
	return item.snapshot, true
}

// Result returns a completed job's sanitized result.
func (manager *JobManager) Result(id string) (any, JobStatus, bool) {
	item, exists := manager.job(id)
	if !exists {
		return nil, "", false
	}
	item.mu.RLock()
	defer item.mu.RUnlock()
	return item.result, item.snapshot.Status, true
}

// Cancel requests cancellation for a live job.
func (manager *JobManager) Cancel(id string) bool {
	item, exists := manager.job(id)
	if !exists {
		return false
	}
	item.cancel()
	return true
}

// Subscribe replays events after afterID and streams future events.
func (manager *JobManager) Subscribe(id string, afterID int64) (<-chan JobEvent, func(), bool) {
	item, exists := manager.job(id)
	if !exists {
		return nil, func() {}, false
	}
	channel := make(chan JobEvent, 64)
	item.mu.Lock()
	for _, event := range item.events {
		if event.ID > afterID {
			channel <- event
		}
	}
	terminal := item.snapshot.Status == JobCompleted || item.snapshot.Status == JobFailed || item.snapshot.Status == JobCancelled
	if !terminal {
		item.subscribers[channel] = struct{}{}
	} else {
		close(channel)
	}
	item.mu.Unlock()
	var once sync.Once
	cancel := func() {
		once.Do(func() {
			item.mu.Lock()
			if _, exists := item.subscribers[channel]; exists {
				delete(item.subscribers, channel)
				close(channel)
			}
			item.mu.Unlock()
		})
	}
	return channel, cancel, true
}

// Close cancels all jobs and releases results.
func (manager *JobManager) Close() {
	manager.cancel()
	manager.mu.Lock()
	for _, item := range manager.jobs {
		item.cancel()
	}
	manager.jobs = make(map[string]*job)
	manager.mu.Unlock()
}

func (manager *JobManager) job(id string) (*job, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	item, exists := manager.jobs[id]
	return item, exists
}

func (manager *JobManager) pruneLocked() {
	if len(manager.jobs) < manager.maxJobs {
		return
	}
	type candidate struct {
		id        string
		completed time.Time
	}
	candidates := make([]candidate, 0)
	for id, item := range manager.jobs {
		item.mu.RLock()
		if item.snapshot.CompletedAt != nil {
			candidates = append(candidates, candidate{id: id, completed: *item.snapshot.CompletedAt})
		}
		item.mu.RUnlock()
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].completed.Before(candidates[j].completed) })
	for _, candidate := range candidates {
		if len(manager.jobs) < manager.maxJobs {
			break
		}
		delete(manager.jobs, candidate.id)
	}
}

func (item *job) finish(status JobStatus, result any, failure *model.PublicFailure) {
	item.mu.Lock()
	completed := time.Now().UTC()
	item.snapshot.Status = status
	item.snapshot.CompletedAt = &completed
	item.snapshot.Failure = failure
	if status == JobCompleted {
		item.result = result
	}
	eventType, message := "job.completed", "Analysis completed"
	if status == JobCancelled {
		eventType, message = "job.cancelled", "Analysis cancelled"
	}
	if status == JobFailed {
		eventType, message = "job.failed", "Analysis failed"
	}
	event := JobEvent{ID: int64(len(item.events) + 1), Type: eventType, Message: message, At: completed}
	item.events = append(item.events, event)
	item.snapshot.LastEvent = &event
	for subscriber := range item.subscribers {
		select {
		case subscriber <- event:
		default:
		}
		close(subscriber)
		delete(item.subscribers, subscriber)
	}
	item.mu.Unlock()
}

func randomID(bytesLength int) (string, error) {
	buffer := make([]byte, bytesLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
