package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testState struct {
	Value    string
	Counter  int
	Recorded []string
}

func simpleJob(state testState, onChange func(JobStatus)) *Job[testState, string] {
	return NewJob(
		"test-job",
		state,
		[]PhaseConfig[testState]{
			{
				Name:  "work",
				Label: "Doing work",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					tracker.SetTotal(10)
					for range 10 {
						tracker.Increment(1)
						state.Counter++
					}
					return nil
				},
			},
		},
		func(state testState) (string, error) {
			return state.Value, nil
		},
		onChange,
	)
}

// --- Constructor ---

func TestNewJob_HasUUID(t *testing.T) {
	j := simpleJob(testState{}, nil)
	assert.NotEmpty(t, j.Id())
	assert.Len(t, j.Id(), 36)
}

func TestNewJob_StartsAsPending(t *testing.T) {
	j := simpleJob(testState{}, nil)

	assert.False(t, j.IsDone())

	status := j.Status()
	assert.Equal(t, JobStatePending, status.State)
	assert.Nil(t, status.Phase)
	assert.Nil(t, status.Progress)

	_, err := j.Result()
	assert.Error(t, err)
}

// --- Single Phase ---

func TestRun_SinglePhase_Success(t *testing.T) {
	j := simpleJob(testState{Value: "hello"}, nil)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "hello", result.Value)
	assert.Empty(t, result.Error)
}

func TestRun_SinglePhase_MutatesState(t *testing.T) {
	j := simpleJob(testState{Value: "result"}, nil)
	j.Run(context.Background())

	result, err := j.Result()
	require.NoError(t, err)
	assert.True(t, result.Ok)
	assert.Equal(t, "result", result.Value)
}

func TestRun_SinglePhase_Error(t *testing.T) {
	j := NewJob(
		"failing-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "fail",
				Label: "Will fail",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					return errors.New("something broke")
				},
			},
		},
		func(state testState) (string, error) { return "", nil },
		nil,
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "something broke", result.Error)
	assert.True(t, j.IsDone())
}

func TestRun_FinalizeError(t *testing.T) {
	j := NewJob(
		"finalize-fail",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "work",
				Label: "Work",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					return nil
				},
			},
		},
		func(state testState) (string, error) {
			return "", errors.New("finalize failed")
		},
		nil,
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "finalize failed", result.Error)
}

// --- Multi Phase ---

func TestRun_MultiPhase_ExecutesInOrder(t *testing.T) {
	j := NewJob(
		"multi-phase",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "first",
				Label: "First",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "first")
					return nil
				},
			},
			{
				Name:  "second",
				Label: "Second",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "second")
					return nil
				},
			},
			{
				Name:  "third",
				Label: "Third",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "third")
					return nil
				},
			},
		},
		func(state testState) (string, error) {
			return state.Recorded[0] + "+" + state.Recorded[1] + "+" + state.Recorded[2], nil
		},
		nil,
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "first+second+third", result.Value)
}

func TestRun_MultiPhase_ErrorStopsExecution(t *testing.T) {
	state := testState{}
	j := NewJob(
		"multi-fail",
		state,
		[]PhaseConfig[testState]{
			{
				Name:  "first",
				Label: "First",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "first")
					return nil
				},
			},
			{
				Name:  "second",
				Label: "Second",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "second")
					return errors.New("phase 2 failed")
				},
			},
			{
				Name:  "third",
				Label: "Third",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "third")
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "", nil },
		nil,
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "phase 2 failed", result.Error)
	assert.Equal(t, []string{"first", "second"}, j.config.Context.Recorded)
}

// --- Cancellation ---

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	j := NewJob(
		"cancel-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "first",
				Label: "First",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "first")
					cancel()
					return nil
				},
			},
			{
				Name:  "second",
				Label: "Should not run",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Recorded = append(state.Recorded, "second")
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "", nil },
		nil,
	)

	result := j.Run(ctx)

	assert.False(t, result.Ok)
	assert.Equal(t, "Job was canceled", result.Error)
	assert.Equal(t, []string{"first"}, j.config.Context.Recorded)
}

// --- State Pipeline ---

func TestRun_StatePipelineBetweenPhases(t *testing.T) {
	j := NewJob(
		"pipeline",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "produce",
				Label: "Produce",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Value = "produced-value"
					return nil
				},
			},
			{
				Name:  "consume",
				Label: "Consume",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					state.Value = state.Value + "+consumed"
					return nil
				},
			},
		},
		func(state testState) (string, error) {
			return state.Value, nil
		},
		nil,
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "produced-value+consumed", result.Value)
}

// --- onChange ---

func TestRun_OnChangeEmitsOnCompletion(t *testing.T) {
	var statuses []JobStatus
	var mu sync.Mutex

	j := simpleJob(testState{Value: "done"}, func(s JobStatus) {
		mu.Lock()
		statuses = append(statuses, s)
		mu.Unlock()
	})

	j.Run(context.Background())

	mu.Lock()
	defer mu.Unlock()

	require.NotEmpty(t, statuses)
	last := statuses[len(statuses)-1]
	assert.Equal(t, JobStateComplete, last.State)
}

func TestRun_OnChangeEmitsOnError(t *testing.T) {
	var statuses []JobStatus
	var mu sync.Mutex

	j := NewJob(
		"error-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "fail",
				Label: "Fail",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					return errors.New("boom")
				},
			},
		},
		func(state testState) (string, error) { return "", nil },
		func(s JobStatus) {
			mu.Lock()
			statuses = append(statuses, s)
			mu.Unlock()
		},
	)

	j.Run(context.Background())

	mu.Lock()
	defer mu.Unlock()

	require.NotEmpty(t, statuses)
	last := statuses[len(statuses)-1]
	assert.Equal(t, JobStateError, last.State)
}

func TestRun_OnChangeEmitsProgressUpdates(t *testing.T) {
	var statuses []JobStatus
	var mu sync.Mutex

	j := NewJob(
		"progress-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "work",
				Label: "Working",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					tracker.SetTotal(2)
					tracker.SetProgress(1)
					tracker.SetProgress(2)
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "ok", nil },
		func(s JobStatus) {
			mu.Lock()
			statuses = append(statuses, s)
			mu.Unlock()
		},
	)

	j.Run(context.Background())

	mu.Lock()
	defer mu.Unlock()

	// SetTotal + 2x SetProgress + Done + completion = 5 updates
	assert.Equal(t, 5, len(statuses))
}

// --- Estimated Phase with EstimateCallback ---

func TestRun_EstimatedPhaseUsesCallback(t *testing.T) {
	var callbackReceived map[string]time.Duration

	j := NewJob(
		"estimated-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "timed-work",
				Label: "Timed",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			},
			{
				Name:  "estimated-work",
				Label: "Estimated",
				Type:  PhaseTypeEstimated,
				EstimateCallback: func(completed map[string]time.Duration) time.Duration {
					callbackReceived = completed
					return completed["timed-work"] * 2
				},
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "ok", nil },
		nil,
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	require.NotNil(t, callbackReceived)
	assert.Contains(t, callbackReceived, "timed-work")
	assert.GreaterOrEqual(t, callbackReceived["timed-work"], 10*time.Millisecond)
}

// --- Result() ---

func TestResult_BeforeRun_ReturnsError(t *testing.T) {
	j := simpleJob(testState{}, nil)

	_, err := j.Result()
	assert.Error(t, err)
}

func TestResult_AfterRun_ReturnsResult(t *testing.T) {
	j := simpleJob(testState{Value: "final"}, nil)
	j.Run(context.Background())

	result, err := j.Result()
	require.NoError(t, err)
	assert.True(t, result.Ok)
	assert.Equal(t, "final", result.Value)
}

// --- Double Run ---

func TestRun_DoubleRunIsNoop(t *testing.T) {
	var count int
	j := NewJob(
		"double-run",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "work",
				Label: "Work",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					count++
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "ok", nil },
		nil,
	)

	j.Run(context.Background())
	j.Run(context.Background())

	assert.Equal(t, 1, count)
}

// --- Status ---

func TestStatus_DuringRun_ReportsPhase(t *testing.T) {
	phaseReached := make(chan bool, 1)
	statusChecked := make(chan bool, 1)

	j := NewJob(
		"status-check",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "first",
				Label: "First Phase",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					return nil
				},
			},
			{
				Name:  "second",
				Label: "Second Phase",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					phaseReached <- true
					<-statusChecked
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "ok", nil },
		nil,
	)

	go j.Run(context.Background())

	<-phaseReached
	status := j.Status()
	statusChecked <- true

	assert.Equal(t, JobStateRunning, status.State)
	assert.Equal(t, "second", status.Phase.Name)
	assert.Equal(t, 1, status.Phase.Index)
	assert.Equal(t, 2, status.Phase.Total)
}
