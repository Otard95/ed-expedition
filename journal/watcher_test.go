package journal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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

func fsdTargetEvent(timestamp, name string, systemAddress int) string {
	return fmt.Sprintf(
		`{"timestamp":"%s","event":"FSDTarget","Name":"%s","SystemAddress":%d,"StarClass":"G"}`,
		timestamp, name, systemAddress,
	)
}

func fsdJumpEvent(timestamp, system string, systemAddress int, fuelUsed, fuelLevel, jumpDist float64) string {
	return fmt.Sprintf(
		`{"timestamp":"%s","event":"FSDJump","StarSystem":"%s","SystemAddress":%d,"StarPos":[0,0,0],"FuelUsed":%.1f,"FuelLevel":%.1f,"JumpDist":%.1f}`,
		timestamp, system, systemAddress, fuelUsed, fuelLevel, jumpDist,
	)
}

func writeJournal(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content+"\n"), 0644); err != nil {
		t.Fatalf("Failed to create journal file %s: %v", filename, err)
	}
}

func appendJournal(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("Failed to open journal file %s: %v", filename, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content + "\n"); err != nil {
		t.Fatalf("Failed to append to journal file %s: %v", filename, err)
	}
	if err := f.Sync(); err != nil {
		t.Fatalf("Failed to sync journal file %s: %v", filename, err)
	}
}

func collectTargetEvents(ch chan *FSDTargetEvent, expected int, timeout time.Duration) []*FSDTargetEvent {
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

	for {
		select {
		case event := <-ch:
			events = append(events, event)
		case <-time.After(10 * time.Millisecond):
			return events
		}
	}
}

func collectJumpEvents(ch chan *FSDJumpEvent, expected int, timeout time.Duration) []*FSDJumpEvent {
	events := make([]*FSDJumpEvent, 0, expected)
	deadline := time.After(timeout)

	for range expected {
		select {
		case event := <-ch:
			events = append(events, event)
		case <-deadline:
			return events
		}
	}

	for {
		select {
		case event := <-ch:
			events = append(events, event)
		case <-time.After(10 * time.Millisecond):
			return events
		}
	}
}

// --- Sync Tests ---

type SyncTestSuite struct {
	suite.Suite
	tmpDir  string
	watcher *Watcher
}

func (s *SyncTestSuite) SetupTest() {
	var err error
	s.tmpDir, err = os.MkdirTemp("", "journal-sync-test-*")
	s.Require().NoError(err)
}

func (s *SyncTestSuite) TearDownTest() {
	if s.watcher != nil {
		s.watcher.Close()
	}
	if s.tmpDir != "" {
		os.RemoveAll(s.tmpDir)
	}
}

func (s *SyncTestSuite) createWatcher() {
	var err error
	s.watcher, err = NewWatcher(s.tmpDir, &TestLogger{})
	s.Require().NoError(err)
}

func (s *SyncTestSuite) createWatcherWithRecorder() *RecordingLogger {
	logger := &RecordingLogger{}
	var err error
	s.watcher, err = NewWatcher(s.tmpDir, logger)
	s.Require().NoError(err)
	return logger
}

func (s *SyncTestSuite) TestDeletedEarlyParts() {
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.03.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "Sol", 1))
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.04.log",
		fsdTargetEvent("2024-12-19T10:10:00Z", "Alpha Centauri", 2))

	s.createWatcher()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 9, 0, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	events := collectTargetEvents(ch, 2, time.Second)
	s.Require().Len(events, 2)
	s.Equal("Sol", events[0].Name)
	s.Equal("Alpha Centauri", events[1].Name)
}

func (s *SyncTestSuite) TestMiddleOfMultipleParts() {
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "Early", 1))
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.02.log",
		fsdTargetEvent("2024-12-19T10:15:00Z", "Middle", 2))
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T110000.01.log",
		fsdTargetEvent("2024-12-19T11:05:00Z", "Late", 3))

	logger := s.createWatcherWithRecorder()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 10, 20, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	events := collectTargetEvents(ch, 1, time.Second)
	s.Require().Len(events, 1)
	s.Equal("Late", events[0].Name)

	eventsSkipped := 0
	for _, msg := range logger.Messages {
		if strings.Contains(msg, "before lastTimestamp") && strings.Contains(msg, "skipping") {
			eventsSkipped++
		}
	}
	s.Equal(2, eventsSkipped)
}

func (s *SyncTestSuite) TestAllJournalsBeforeSince() {
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "Old1", 1))
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T110000.01.log",
		fsdTargetEvent("2024-12-19T11:05:00Z", "Old2", 2)+"\n"+
			fsdTargetEvent("2024-12-19T13:30:00Z", "Recent", 3))

	s.createWatcher()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 13, 0, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	events := collectTargetEvents(ch, 1, time.Second)
	s.Require().Len(events, 1)
	s.Equal("Recent", events[0].Name)
}

func (s *SyncTestSuite) TestAllJournalsAfterSince() {
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "First", 1))
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T110000.01.log",
		fsdTargetEvent("2024-12-19T11:05:00Z", "Second", 2))

	s.createWatcher()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 18, 0, 0, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	events := collectTargetEvents(ch, 2, time.Second)
	s.Require().Len(events, 2)
	s.Equal("First", events[0].Name)
	s.Equal("Second", events[1].Name)
}

func (s *SyncTestSuite) TestEmptyDirectory() {
	s.createWatcher()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 10, 0, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	events := collectTargetEvents(ch, 0, time.Second)
	s.Len(events, 0)
}

func (s *SyncTestSuite) TestExactTimestampIncluded() {
	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:00:00Z", "Before", 1)+"\n"+
			fsdTargetEvent("2024-12-19T10:05:00Z", "ExactMatch", 2)+"\n"+
			fsdTargetEvent("2024-12-19T10:10:00Z", "After", 3))

	s.createWatcher()
	ch := s.watcher.FSDTarget.Subscribe()

	since := time.Date(2024, 12, 19, 10, 5, 0, 0, time.UTC)
	s.Require().NoError(s.watcher.Sync(since))

	// TODO: Exact timestamp dedup requires hash-based approach (see #39).
	// Events at exactly `since` are currently included (potential duplicate).
	events := collectTargetEvents(ch, 2, time.Second)
	s.Require().Len(events, 2)
	s.Equal("ExactMatch", events[0].Name)
	s.Equal("After", events[1].Name)
}

func TestSyncTestSuite(t *testing.T) {
	suite.Run(t, new(SyncTestSuite))
}

// --- Live Watcher Tests ---

type LiveTestSuite struct {
	suite.Suite
	tmpDir  string
	watcher *Watcher
}

func (s *LiveTestSuite) SetupTest() {
	var err error
	s.tmpDir, err = os.MkdirTemp("", "journal-live-test-*")
	s.Require().NoError(err)

	s.watcher, err = NewWatcher(s.tmpDir, &TestLogger{})
	s.Require().NoError(err)
	s.watcher.Start()
}

func (s *LiveTestSuite) TearDownTest() {
	if s.watcher != nil {
		s.watcher.Close()
	}
	if s.tmpDir != "" {
		os.RemoveAll(s.tmpDir)
	}
}

func (s *LiveTestSuite) TestNewJournalFile() {
	ch := s.watcher.FSDTarget.Subscribe()

	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "Sol", 1))

	events := collectTargetEvents(ch, 1, time.Second)
	s.Require().Len(events, 1)
	s.Equal("Sol", events[0].Name)
}

func (s *LiveTestSuite) TestAppendToExistingJournal() {
	ch := s.watcher.FSDTarget.Subscribe()

	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "First", 1))

	events := collectTargetEvents(ch, 1, time.Second)
	s.Require().Len(events, 1)
	s.Equal("First", events[0].Name)

	appendJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:10:00Z", "Second", 2))

	events = collectTargetEvents(ch, 1, time.Second)
	s.Require().Len(events, 1)
	s.Equal("Second", events[0].Name)
}

func (s *LiveTestSuite) TestMultipleEventTypes() {
	targetCh := s.watcher.FSDTarget.Subscribe()
	jumpCh := s.watcher.FSDJump.Subscribe()

	writeJournal(s.T(), s.tmpDir, "Journal.2024-12-19T100000.01.log",
		fsdTargetEvent("2024-12-19T10:05:00Z", "Sol", 1)+"\n"+
			fsdJumpEvent("2024-12-19T10:06:00Z", "Sol", 1, 1.5, 14.5, 10.5))

	targets := collectTargetEvents(targetCh, 1, time.Second)
	s.Require().Len(targets, 1)
	s.Equal("Sol", targets[0].Name)

	jumps := collectJumpEvents(jumpCh, 1, time.Second)
	s.Require().Len(jumps, 1)
	s.Equal("Sol", jumps[0].StarSystem)
}

func TestLiveTestSuite(t *testing.T) {
	suite.Run(t, new(LiveTestSuite))
}
