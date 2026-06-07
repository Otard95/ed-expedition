package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testLogger struct{}

func (l *testLogger) Print(message string)   {}
func (l *testLogger) Trace(message string)   {}
func (l *testLogger) Debug(message string)   {}
func (l *testLogger) Info(message string)    {}
func (l *testLogger) Warning(message string) {}
func (l *testLogger) Error(message string)   {}
func (l *testLogger) Fatal(message string)   {}

type testState struct {
	Value    string
	Counter  int
	Recorded []string
}

func simpleJob(state testState) *Job[testState, string] {
	return New(
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
		&testLogger{},
	)
}

func collectStatuses(j *Job[testState, string]) *[]JobStatus {
	statuses := &[]JobStatus{}
	ch := j.StatusChange().Subscribe()
	go func() {
		for s := range ch {
			*statuses = append(*statuses, s)
		}
	}()
	return statuses
}

// --- Constructor ---

func TestNewJob_HasUUID(t *testing.T) {
	j := simpleJob(testState{})
	assert.NotEmpty(t, j.Id())
	assert.Len(t, j.Id(), 36)
}

func TestNewJob_StartsAsPending(t *testing.T) {
	j := simpleJob(testState{})

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
	j := simpleJob(testState{Value: "hello"})

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "hello", result.Value)
	assert.Empty(t, result.Error)
}

func TestRun_SinglePhase_MutatesState(t *testing.T) {
	j := simpleJob(testState{Value: "result"})
	j.Run(context.Background())

	result, err := j.Result()
	require.NoError(t, err)
	assert.True(t, result.Ok)
	assert.Equal(t, "result", result.Value)
}

func TestRun_SinglePhase_Error(t *testing.T) {
	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "something broke", result.Error)
	assert.True(t, j.IsDone())
}

func TestRun_FinalizeError(t *testing.T) {
	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "finalize failed", result.Error)
}

// --- Multi Phase ---

func TestRun_MultiPhase_ExecutesInOrder(t *testing.T) {
	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "first+second+third", result.Value)
}

func TestRun_MultiPhase_ErrorStopsExecution(t *testing.T) {
	state := testState{}
	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.False(t, result.Ok)
	assert.Equal(t, "phase 2 failed", result.Error)
	assert.Equal(t, []string{"first", "second"}, j.config.Context.Recorded)
}

// --- Cancellation ---

func TestRun_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	j := New(
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
		&testLogger{},
	)

	result := j.Run(ctx)

	assert.False(t, result.Ok)
	assert.Equal(t, "Job was canceled", result.Error)
	assert.Equal(t, []string{"first"}, j.config.Context.Recorded)
}

// --- State Pipeline ---

func TestRun_StatePipelineBetweenPhases(t *testing.T) {
	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	assert.Equal(t, "produced-value+consumed", result.Value)
}

// --- Status Updates via FanoutChannel ---

func TestRun_StatusChange_EmitsOnCompletion(t *testing.T) {
	j := simpleJob(testState{Value: "done"})
	statuses := collectStatuses(j)

	j.Run(context.Background())
	time.Sleep(10 * time.Millisecond)

	require.NotEmpty(t, *statuses)
	last := (*statuses)[len(*statuses)-1]
	assert.Equal(t, JobStateComplete, last.State)
}

func TestRun_StatusChange_EmitsOnError(t *testing.T) {
	j := New(
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
		&testLogger{},
	)
	statuses := collectStatuses(j)

	j.Run(context.Background())
	time.Sleep(10 * time.Millisecond)

	require.NotEmpty(t, *statuses)
	last := (*statuses)[len(*statuses)-1]
	assert.Equal(t, JobStateError, last.State)
}

func TestRun_StatusChange_EmitsProgressUpdates(t *testing.T) {
	j := New(
		"progress-job",
		testState{},
		[]PhaseConfig[testState]{
			{
				Name:  "work",
				Label: "Working",
				Type:  PhaseTypeObservable,
				Callback: func(ctx context.Context, state *testState, tracker *ProgressTracker) error {
					tracker.SetTotal(2)
					time.Sleep(110 * time.Millisecond)
					tracker.SetProgress(1)
					time.Sleep(110 * time.Millisecond)
					tracker.SetProgress(2)
					return nil
				},
			},
		},
		func(state testState) (string, error) { return "ok", nil },
		&testLogger{},
	)
	statuses := collectStatuses(j)

	j.Run(context.Background())
	time.Sleep(10 * time.Millisecond)

	// Phase start + SetTotal + SetProgress(1) + SetProgress(2) + completion
	// Throttle is 100ms, sleeps ensure each update gets through
	assert.GreaterOrEqual(t, len(*statuses), 4)
}

// --- Estimated Phase with EstimateCallback ---

func TestRun_EstimatedPhaseUsesCallback(t *testing.T) {
	var callbackReceived map[string]time.Duration

	j := New(
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
		&testLogger{},
	)

	result := j.Run(context.Background())

	assert.True(t, result.Ok)
	require.NotNil(t, callbackReceived)
	assert.Contains(t, callbackReceived, "timed-work")
	assert.GreaterOrEqual(t, callbackReceived["timed-work"], 10*time.Millisecond)
}

// --- Result() ---

func TestResult_BeforeRun_ReturnsError(t *testing.T) {
	j := simpleJob(testState{})

	_, err := j.Result()
	assert.Error(t, err)
}

func TestResult_AfterRun_ReturnsResult(t *testing.T) {
	j := simpleJob(testState{Value: "final"})
	j.Run(context.Background())

	result, err := j.Result()
	require.NoError(t, err)
	assert.True(t, result.Ok)
	assert.Equal(t, "final", result.Value)
}

// --- Double Run ---

func TestRun_DoubleRunIsNoop(t *testing.T) {
	var count int
	j := New(
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
		&testLogger{},
	)

	j.Run(context.Background())
	j.Run(context.Background())

	assert.Equal(t, 1, count)
}

// --- Status ---

func TestStatus_DuringRun_ReportsPhase(t *testing.T) {
	phaseReached := make(chan bool, 1)
	statusChecked := make(chan bool, 1)

	j := New(
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
		&testLogger{},
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
