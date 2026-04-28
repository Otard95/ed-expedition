package job

import (
	"context"
	"ed-expedition/lib/channels"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

type NoCtx = struct{}

type PhaseType int
type JobState string

const (
	PhaseTypeObservable PhaseType = iota
	PhaseTypeEstimated
	PhaseTypeIndeterminate

	JobStatePending  JobState = "pending"
	JobStateRunning  JobState = "running"
	JobStateComplete JobState = "complete"
	JobStateError    JobState = "error"
)

type PhaseConfig[C any] struct {
	Name             string
	Label            string
	Type             PhaseType
	Callback         func(ctx context.Context, state *C, tracker *ProgressTracker) error
	EstimateCallback func(completed map[string]time.Duration) time.Duration
}

type JobConfig[C any, R any] struct {
	Name     string
	Context  *C
	Phases   []PhaseConfig[C]
	Finalize func(state C) (R, error)
}

type PhaseStatus struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Index int    `json:"index"`
	Total int    `json:"total"`
}
type ProgressInfo struct {
	Fraction    float64 `json:"fraction"`
	Label       string  `json:"label"`
	Determinate bool    `json:"determinate"`
}

type JobStatus struct {
	ID       string        `json:"id"`
	State    JobState      `json:"status"`
	Name     string        `json:"name"`
	Phase    *PhaseStatus  `json:"phase,omitempty"`
	Progress *ProgressInfo `json:"progress,omitempty"`
}

type JobResult[R any] struct {
	Ok    bool   `json:"ok"`
	Value R      `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type Job[C any, R any] struct {
	mu                  sync.RWMutex
	id                  string
	config              JobConfig[C, R]
	durations           map[string]time.Duration
	currentPhase        int
	currentPhaseTracker *ProgressTracker
	state               JobState
	result              *JobResult[R]
	statusChange        *channels.FanoutChannel[JobStatus]
}

func New[C any, R any](
	name string,
	context C,
	phases []PhaseConfig[C],
	finalize func(state C) (R, error),
	logger wailsLogger.Logger,
) *Job[C, R] {
	return &Job[C, R]{
		id: uuid.New().String(),
		config: JobConfig[C, R]{
			Name:     name,
			Context:  &context,
			Phases:   phases,
			Finalize: finalize,
		},
		durations:    make(map[string]time.Duration),
		currentPhase: -1,
		state:        JobStatePending,
		statusChange: channels.NewFanoutChannel[JobStatus](
			"StatusChange", 5, 5*time.Millisecond, logger),
	}
}

func (j *Job[C, R]) Id() string {
	return j.id
}

func (j *Job[C, R]) StatusChange() *channels.FanoutChannel[JobStatus] {
	return j.statusChange
}

func (j *Job[C, R]) Status() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()

	status := JobStatus{
		ID:    j.id,
		State: j.state,
		Name:  j.config.Name,
	}

	if j.currentPhase > -1 {
		status.Phase = &PhaseStatus{
			Name:  j.config.Phases[j.currentPhase].Name,
			Label: j.config.Phases[j.currentPhase].Label,
			Index: j.currentPhase,
			Total: len(j.config.Phases),
		}
		status.Progress = &ProgressInfo{
			Fraction:    j.currentPhaseTracker.Fraction(),
			Label:       j.currentPhaseTracker.Label(),
			Determinate: j.currentPhaseTracker.kind != PhaseTypeIndeterminate,
		}
	}

	return status
}

func (j *Job[C, R]) IsDone() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.state == JobStateComplete || j.state == JobStateError
}

func (j *Job[C, R]) Result() (*JobResult[R], error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	if j.state == JobStatePending || j.state == JobStateRunning {
		return nil, errors.New("Cannot get result of a job that is not done")
	}
	return j.result, nil
}

func (j *Job[C, R]) setError(err string) {
	j.mu.Lock()
	j.result = &JobResult[R]{Ok: false, Error: err}
	j.state = JobStateError
	j.mu.Unlock()
}

func (j *Job[C, R]) createPhaseTracker(
	phase *PhaseConfig[C],
	onChange func(t *ProgressTracker),
) *ProgressTracker {
	switch phase.Type {
	case PhaseTypeObservable:
		return NewObservableProgressTracker(onChange)
	case PhaseTypeEstimated:
		return NewEstimatedProgressTracker(phase.EstimateCallback(j.durations), onChange)
	case PhaseTypeIndeterminate:
		return NewIndeterminateProgressTracker(onChange)
	}
	panic("[Job.createPhaseTracker] Unreachable")
}

func (j *Job[C, R]) Run(ctx context.Context) *JobResult[R] {
	if j.IsDone() {
		return j.result
	}
	defer j.statusChange.Close()

	j.mu.Lock()
	j.state = JobStateRunning
	j.currentPhase = 0
	j.mu.Unlock()

	trackerOnChange := func(t *ProgressTracker) {
		j.statusChange.Publish(j.Status())
	}

	for i, phase := range j.config.Phases {

		j.mu.Lock()
		tracker := j.createPhaseTracker(&phase, trackerOnChange)
		j.currentPhaseTracker = tracker
		j.currentPhase = i
		j.mu.Unlock()

		j.statusChange.Publish(j.Status())

		err := phase.Callback(ctx, j.config.Context, tracker)

		if err != nil {
			j.setError(err.Error())
			tracker.Done()
			j.statusChange.Publish(j.Status())
			return j.result
		}

		tracker.Done()

		j.mu.Lock()
		j.durations[phase.Name] = tracker.Duration()
		j.mu.Unlock()

		select {
		case <-ctx.Done():
			j.setError("Job was canceled")
			j.statusChange.Publish(j.Status())
			return j.result
		default:
		}

	}

	j.mu.Lock()
	value, err := j.config.Finalize(*j.config.Context)

	if err != nil {
		j.state = JobStateError
		j.result = &JobResult[R]{Ok: false, Error: err.Error()}
	} else {
		j.state = JobStateComplete
		j.result = &JobResult[R]{Ok: true, Value: value}
	}
	j.mu.Unlock()

	j.statusChange.Publish(j.Status())

	return j.result
}
