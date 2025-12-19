package journal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLogger implements wails Logger interface for testing
type TestLogger struct{}

func (l *TestLogger) Print(message string)   {}
func (l *TestLogger) Trace(message string)   {}
func (l *TestLogger) Debug(message string)   {}
func (l *TestLogger) Info(message string)    {}
func (l *TestLogger) Warning(message string) {}
func (l *TestLogger) Error(message string)   {}
func (l *TestLogger) Fatal(message string)   {}

// RecordingLogger implements wails Logger interface and records trace messages
type RecordingLogger struct {
	Messages []string
}

func (l *RecordingLogger) Print(message string)   {}
func (l *RecordingLogger) Trace(message string)   { l.Messages = append(l.Messages, message) }
func (l *RecordingLogger) Debug(message string)   {}
func (l *RecordingLogger) Info(message string)    {}
func (l *RecordingLogger) Warning(message string) {}
func (l *RecordingLogger) Error(message string)   {}
func (l *RecordingLogger) Fatal(message string)   {}

func TestWatcher_Sync_DeletedEarlyParts(t *testing.T) {
	// Test case: First journal is part 3 (parts 1-2 were deleted)
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create journals starting from part 3 (simulating deleted parts 1-2)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.03.json",
		`{"timestamp":"2024-12-19T10:05:00Z","event":"FSDTarget","Name":"Sol","SystemAddress":10477373803,"StarClass":"G"}`)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.04.json",
		`{"timestamp":"2024-12-19T10:10:00Z","event":"FSDTarget","Name":"Alpha Centauri","SystemAddress":123456,"StarClass":"K"}`)

	watcher, err := NewWatcher(tmpDir, &TestLogger{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Subscribe to FSDTarget events
	targetChan := watcher.FSDTarget.Subscribe()

	// Sync from before all journals
	since := time.Date(2024, 12, 19, 9, 0, 0, 0, time.UTC)
	if err := watcher.Sync(since); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should receive both events
	events := collectEvents(targetChan, 2, time.Millisecond*1000)
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	if events[0].Name != "Sol" {
		t.Errorf("Expected first event to be Sol, got %s", events[0].Name)
	}
	if events[1].Name != "Alpha Centauri" {
		t.Errorf("Expected second event to be Alpha Centauri, got %s", events[1].Name)
	}
}

func TestWatcher_Sync_MiddleOfMultipleParts(t *testing.T) {
	// Test case: Sync in the middle of journal parts
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple parts for same timestamp
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.01.json",
		`{"timestamp":"2024-12-19T10:05:00Z","event":"FSDTarget","Name":"Early","SystemAddress":1,"StarClass":"G"}`)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.02.json",
		`{"timestamp":"2024-12-19T10:15:00Z","event":"FSDTarget","Name":"Middle","SystemAddress":2,"StarClass":"K"}`)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T110000.01.json",
		`{"timestamp":"2024-12-19T11:05:00Z","event":"FSDTarget","Name":"Late","SystemAddress":3,"StarClass":"M"}`)

	logger := &RecordingLogger{}
	watcher, err := NewWatcher(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	targetChan := watcher.FSDTarget.Subscribe()

	// Sync from 10:20 (should start from beginning to catch events after 10:20 in earlier journals)
	since := time.Date(2024, 12, 19, 10, 20, 0, 0, time.UTC)
	if err := watcher.Sync(since); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should receive only events after 10:20
	events := collectEvents(targetChan, 1, time.Millisecond*1000)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Name != "Late" {
		t.Errorf("Expected Late event, got %s", events[0].Name)
	}

	// Verify all 3 journals were read but 2 events were skipped due to timestamp
	eventsSkipped := 0
	for _, msg := range logger.Messages {
		if strings.Contains(msg, "before lastTimestamp") && strings.Contains(msg, "skipping") {
			eventsSkipped++
		}
	}
	if eventsSkipped != 2 {
		t.Errorf("Expected 2 events to be skipped due to timestamp, got %d", eventsSkipped)
	}
}

func TestWatcher_Sync_AllJournalsBeforeSince(t *testing.T) {
	// Test case: All journals are before the 'since' timestamp
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create journals with events before 'since'
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.01.json",
		`{"timestamp":"2024-12-19T10:05:00Z","event":"FSDTarget","Name":"Old1","SystemAddress":1,"StarClass":"G"}`)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T110000.01.json",
		`{"timestamp":"2024-12-19T11:05:00Z","event":"FSDTarget","Name":"Old2","SystemAddress":2,"StarClass":"K"}
{"timestamp":"2024-12-19T13:30:00Z","event":"FSDTarget","Name":"Recent","SystemAddress":3,"StarClass":"M"}`)

	watcher, err := NewWatcher(tmpDir, &TestLogger{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	targetChan := watcher.FSDTarget.Subscribe()

	// Sync from 13:00 (last journal started at 11:00, but has event at 13:30)
	since := time.Date(2024, 12, 19, 13, 0, 0, 0, time.UTC)
	if err := watcher.Sync(since); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should receive only the recent event
	events := collectEvents(targetChan, 1, time.Millisecond*1000)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Name != "Recent" {
		t.Errorf("Expected Recent event, got %s", events[0].Name)
	}
}

func TestWatcher_Sync_AllJournalsAfterSince(t *testing.T) {
	// Test case: All journals are after the 'since' timestamp
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T100000.01.json",
		`{"timestamp":"2024-12-19T10:05:00Z","event":"FSDTarget","Name":"First","SystemAddress":1,"StarClass":"G"}`)
	createJournalWithEvent(t, tmpDir, "Journal.2024-12-19T110000.01.json",
		`{"timestamp":"2024-12-19T11:05:00Z","event":"FSDTarget","Name":"Second","SystemAddress":2,"StarClass":"K"}`)

	watcher, err := NewWatcher(tmpDir, &TestLogger{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	targetChan := watcher.FSDTarget.Subscribe()

	// Sync from way before all journals
	since := time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC)
	if err := watcher.Sync(since); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Should receive both events
	events := collectEvents(targetChan, 2, time.Millisecond*1000)
	t.Logf("Received %d events", len(events))
	for i, event := range events {
		t.Logf("Event %d: %s at %v", i, event.Name, event.Timestamp)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	if events[0].Name != "First" {
		t.Errorf("Expected first event to be First, got %s", events[0].Name)
	}
	if events[1].Name != "Second" {
		t.Errorf("Expected second event to be Second, got %s", events[1].Name)
	}
}

func TestWatcher_Sync_EmptyDirectory(t *testing.T) {
	// Test case: No journal files
	tmpDir, err := os.MkdirTemp("", "journal-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	watcher, err := NewWatcher(tmpDir, &TestLogger{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	targetChan := watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 10, 0, 0, 0, time.UTC)
	if err := watcher.Sync(since); err != nil {
		t.Fatalf("Sync should not fail on empty directory: %v", err)
	}

	// Should receive no events
	events := collectEvents(targetChan, 0, time.Millisecond*1000)
	if len(events) != 0 {
		t.Fatalf("Expected 0 events, got %d", len(events))
	}
}

// Helper functions

func createJournalWithEvent(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create journal file %s: %v", filename, err)
	}
}

func collectEvents(ch chan *FSDTargetEvent, expected int, timeout time.Duration) []*FSDTargetEvent {
	events := make([]*FSDTargetEvent, 0, expected)
	deadline := time.After(timeout)

	for range expected {
		select {
		case event := <-ch:
			events = append(events, event)
		case <-deadline:
			return events
		}
	}

	// Drain any extra events
	for {
		select {
		case event := <-ch:
			events = append(events, event)
		case <-time.After(time.Millisecond * 10):
			return events
		}
	}
}
