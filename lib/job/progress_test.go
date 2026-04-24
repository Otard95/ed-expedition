package job

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func noopOnChange(t *ProgressTracker) {}

func TestObservableTracker_Kind(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	assert.Equal(t, PhaseTypeObservable, tracker.Kind())
}

func TestObservableTracker_SetTotalAndProgress(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)

	tracker.SetTotal(100)
	tracker.SetProgress(50)

	assert.Equal(t, 0.5, tracker.Fraction())
}

func TestObservableTracker_FractionZeroTotal(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	assert.Equal(t, 0.0, tracker.Fraction())
}

func TestObservableTracker_Increment(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	tracker.SetTotal(10)

	tracker.Increment(3)
	tracker.Increment(3)
	tracker.Increment(4)

	assert.Equal(t, 1.0, tracker.Fraction())
}

func TestObservableTracker_Label(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	tracker.SetLabel("4.2 GB / 10.1 GB")
	assert.Equal(t, "4.2 GB / 10.1 GB", tracker.Label())
}

func TestObservableTracker_Done(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	tracker.SetTotal(100)
	tracker.SetProgress(50)

	tracker.Done()

	assert.True(t, tracker.IsDone())
	assert.Equal(t, 1.0, tracker.Fraction())
}

func TestObservableTracker_DoneSetsDuration(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)

	time.Sleep(10 * time.Millisecond)
	tracker.Done()

	assert.GreaterOrEqual(t, tracker.Duration(), 10*time.Millisecond)
}

func TestObservableTracker_DoubleDoneIsNoop(t *testing.T) {
	var count int
	tracker := NewObservableProgressTracker(func(t *ProgressTracker) {
		count++
	})
	tracker.SetTotal(10)

	tracker.Done()
	countAfterFirst := count
	tracker.Done()

	assert.Equal(t, countAfterFirst, count)
}

func TestObservableTracker_MutationsAfterDoneAreIgnored(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	tracker.SetTotal(100)
	tracker.SetProgress(50)
	tracker.Done()

	tracker.SetProgress(75)
	tracker.SetTotal(200)
	tracker.Increment(10)

	assert.Equal(t, 1.0, tracker.Fraction())
}

func TestObservableTracker_OnChangeFiresOnSetTotal(t *testing.T) {
	var count int
	tracker := NewObservableProgressTracker(func(t *ProgressTracker) {
		count++
	})

	tracker.SetTotal(100)

	assert.Equal(t, 1, count)
}

func TestObservableTracker_OnChangeFiresOnProgress(t *testing.T) {
	var count int
	tracker := NewObservableProgressTracker(func(t *ProgressTracker) {
		count++
	})
	tracker.SetTotal(100)

	tracker.SetProgress(25)
	tracker.SetProgress(50)
	tracker.Increment(10)

	assert.Equal(t, 4, count)
}

func TestObservableTracker_OnChangeFiresOnLabel(t *testing.T) {
	var count int
	tracker := NewObservableProgressTracker(func(t *ProgressTracker) {
		count++
	})

	tracker.SetLabel("test")

	assert.Equal(t, 1, count)
}

func TestObservableTracker_OnChangeFiresOnDone(t *testing.T) {
	var fired bool
	tracker := NewObservableProgressTracker(func(t *ProgressTracker) {
		fired = true
	})

	tracker.Done()

	assert.True(t, fired)
}

// --- Indeterminate Tracker ---

func TestIndeterminateTracker_Kind(t *testing.T) {
	tracker := NewIndeterminateProgressTracker(noopOnChange)
	assert.Equal(t, PhaseTypeIndeterminate, tracker.Kind())
}

func TestIndeterminateTracker_IgnoresMutations(t *testing.T) {
	var count int
	tracker := NewIndeterminateProgressTracker(func(t *ProgressTracker) {
		count++
	})

	tracker.SetTotal(100)
	tracker.SetProgress(50)
	tracker.Increment(10)

	assert.Equal(t, 0.0, tracker.Fraction())
	assert.Equal(t, 0, count)
}

func TestIndeterminateTracker_DoneStillWorks(t *testing.T) {
	tracker := NewIndeterminateProgressTracker(noopOnChange)
	tracker.Done()
	assert.True(t, tracker.IsDone())
}

func TestIndeterminateTracker_LabelStillWorks(t *testing.T) {
	tracker := NewIndeterminateProgressTracker(noopOnChange)
	tracker.SetLabel("waiting for API")
	assert.Equal(t, "waiting for API", tracker.Label())
}

// --- Estimated Tracker ---

func TestEstimatedTracker_Kind(t *testing.T) {
	tracker := NewEstimatedProgressTracker(time.Second, noopOnChange)
	defer tracker.Done()
	assert.Equal(t, PhaseTypeEstimated, tracker.Kind())
}

func TestEstimatedTracker_ProgressIncreases(t *testing.T) {
	tracker := NewEstimatedProgressTracker(2*time.Second, noopOnChange)
	defer tracker.Done()

	// Wait past the 500ms tick interval so at least one tick fires
	time.Sleep(600 * time.Millisecond)

	assert.Greater(t, tracker.Fraction(), 0.0)
}

func TestEstimatedTracker_NeverReachesOne(t *testing.T) {
	tracker := NewEstimatedProgressTracker(50*time.Millisecond, noopOnChange)
	defer tracker.Done()

	time.Sleep(200 * time.Millisecond)

	assert.Less(t, tracker.Fraction(), 1.0)
}

func TestEstimatedTracker_DoneStopsGoroutine(t *testing.T) {
	var calls atomic.Int32
	tracker := NewEstimatedProgressTracker(100*time.Millisecond, func(t *ProgressTracker) {
		calls.Add(1)
	})

	tracker.Done()
	countAtDone := calls.Load()

	time.Sleep(150 * time.Millisecond)

	assert.Equal(t, countAtDone, calls.Load())
}

func TestEstimatedTracker_DoneSetsFullProgress(t *testing.T) {
	tracker := NewEstimatedProgressTracker(time.Second, noopOnChange)
	tracker.Done()
	assert.Equal(t, 1.0, tracker.Fraction())
}

// --- Concurrency ---

func TestObservableTracker_ConcurrentAccess(t *testing.T) {
	tracker := NewObservableProgressTracker(noopOnChange)
	tracker.SetTotal(1000)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			tracker.Increment(1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = tracker.Fraction()
			_ = tracker.Label()
			_ = tracker.IsDone()
		}
	}()

	wg.Wait()

	assert.Equal(t, 1.0, tracker.Fraction())
}
