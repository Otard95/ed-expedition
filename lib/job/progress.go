package job

import (
	"math"
	"sync"
	"time"
)

type ProgressTracker struct {
	mu        sync.RWMutex
	kind      PhaseType
	startTime time.Time
	endTime   time.Time
	total     float64
	current   float64
	label     string
	OnChange  func(t *ProgressTracker)
}

func NewObservableProgressTracker(onChange func(t *ProgressTracker)) *ProgressTracker {
	return &ProgressTracker{
		kind:      PhaseTypeObservable,
		startTime: time.Now(),
		endTime:   time.Time{},
		OnChange:  onChange,
	}
}

func NewEstimatedProgressTracker(estimated time.Duration, onChange func(t *ProgressTracker)) *ProgressTracker {
	end := make(chan bool, 1)
	t := &ProgressTracker{
		kind:      PhaseTypeEstimated,
		startTime: time.Now(),
		endTime:   time.Time{},
		total:     1,
		OnChange: func(t *ProgressTracker) {
			if t.IsDone() {
				end <- true
			}
			onChange(t)
		},
	}

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				elapsed := float64(time.Since(t.startTime).Milliseconds())
				estimated := float64(estimated.Milliseconds())
				t.SetProgress(1 - math.Pow(math.E, -2*(elapsed/estimated)))
			case <-end:
				return
			}
		}
	}()

	return t
}

func NewIndeterminateProgressTracker(onChange func(t *ProgressTracker)) *ProgressTracker {
	return &ProgressTracker{
		kind:      PhaseTypeIndeterminate,
		startTime: time.Now(),
		endTime:   time.Time{},
		OnChange:  onChange,
	}
}

func (t *ProgressTracker) Kind() PhaseType {
	return t.kind
}

func (t *ProgressTracker) SetTotal(v float64) {
	if t.IsDone() || t.kind == PhaseTypeIndeterminate {
		return
	}

	t.mu.Lock()
	t.total = v
	t.mu.Unlock()
	t.OnChange(t)
}

func (t *ProgressTracker) SetProgress(v float64) {
	if t.IsDone() || t.kind == PhaseTypeIndeterminate {
		return
	}

	t.mu.Lock()
	t.current = v
	t.mu.Unlock()
	t.OnChange(t)
}

func (t *ProgressTracker) SetLabel(v string) {
	t.mu.Lock()
	t.label = v
	t.mu.Unlock()
	t.OnChange(t)
}

func (t *ProgressTracker) Increment(v float64) {
	if t.IsDone() || t.kind == PhaseTypeIndeterminate {
		return
	}

	t.mu.Lock()
	t.current += v
	t.mu.Unlock()
	t.OnChange(t)
}

func (t *ProgressTracker) Done() {
	if t.IsDone() {
		return
	}

	t.mu.Lock()
	t.current = t.total
	t.endTime = time.Now()
	t.mu.Unlock()
	t.OnChange(t)
}

func (t *ProgressTracker) Label() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.label
}

func (t *ProgressTracker) Fraction() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.total == 0 {
		return 0
	}
	return float64(t.current) / float64(t.total)
}

func (t *ProgressTracker) IsDone() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return !t.endTime.IsZero()
}

func (t *ProgressTracker) Duration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.endTime.Sub(t.startTime)
}
