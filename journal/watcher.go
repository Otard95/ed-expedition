package journal

import (
	"ed-expedition/lib/channels"
	"ed-expedition/lib/slice"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/fsnotify/fsnotify"
	wailsLogger "github.com/wailsapp/wails/v2/pkg/logger"
)

var journalFilePattern = regexp.MustCompile(`Journal\.(\d{4}-\d{2}-\d{2}T\d{6})\.(\d+)\.json`)

const FanoutChannelTimeout = 200 * time.Millisecond

type Watcher struct {
	dir           string
	watcher       *fsnotify.Watcher
	currentFile   string
	seek          int64
	lastTimestamp time.Time
	started       bool
	logger        wailsLogger.Logger

	Loadout   *channels.FanoutChannel[*LoadoutEvent]
	FSDJump   *channels.FanoutChannel[*FSDJumpEvent]
	FSDTarget *channels.FanoutChannel[*FSDTargetEvent]
	Location  *channels.FanoutChannel[*LocationEvent]
}

func NewWatcher(dir string, logger wailsLogger.Logger) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(dir)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return &Watcher{
		dir:         dir,
		watcher:     watcher,
		currentFile: "",
		logger:      logger,

		Loadout:   channels.NewFanoutChannel[*LoadoutEvent]("Loadout", 32, FanoutChannelTimeout, logger),
		FSDJump:   channels.NewFanoutChannel[*FSDJumpEvent]("FSDJump", 32, FanoutChannelTimeout, logger),
		FSDTarget: channels.NewFanoutChannel[*FSDTargetEvent]("FSDTarget", 32, FanoutChannelTimeout, logger),
		Location:  channels.NewFanoutChannel[*LocationEvent]("Location", 32, FanoutChannelTimeout, logger),
	}, nil
}

func (jw *Watcher) Start() {
	jw.started = true
	go func() {
		for e := range jw.watcher.Events {
			file := path.Base(e.Name)

			if file == jw.currentFile {
				jw.handleJournalUpdate()
				continue
			}

			if !journalFilePattern.MatchString(file) {
				continue
			}

			jw.currentFile = file
			jw.seek = 0
			jw.handleJournalUpdate()
		}
	}()
}

func (jw *Watcher) Close() {
	jw.watcher.Close()
}

func (jw *Watcher) handleJournalUpdate() error {
	file, err := os.Open(path.Join(jw.dir, jw.currentFile))
	if err != nil {
		return err
	}
	defer file.Close()

	pos, err := file.Seek(jw.seek, 0)
	if err != nil {
		panic(err)
	}
	if pos != jw.seek {
		panic(fmt.Sprintf("Failed to seek! got to %d aimed for %d", pos, jw.seek))
	}

	buf := []byte{}
	for {
		b := make([]byte, 50000)
		n, err := file.Read(b)
		jw.seek += int64(n)
		if len(buf) == 0 {
			buf = b[:n]
		} else {
			buf = append(buf, b[:n]...)
		}

		if err == io.EOF {
			break
		}
	}

	err = jw.processData(buf)
	if err != nil {
		panic(fmt.Sprintf("Failed to process journal data: %v", err))
	}

	return nil
}

func (jw *Watcher) processData(data []byte) error {
	lines := slice.Split(data, '\n')
	jw.logger.Trace(fmt.Sprintf("[processData] Processing %d lines, lastTimestamp=%v", len(lines), jw.lastTimestamp))

	for i, line := range lines {
		if len(line) == 0 {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: empty, skipping", i))
			continue
		}

		jw.logger.Trace(fmt.Sprintf("[processData] Line %d: %s", i, string(line)))

		var base struct {
			Timestamp time.Time `json:"timestamp"`
			Event     EventType `json:"event"`
		}
		err := json.Unmarshal(line, &base)
		if err != nil {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: unmarshal error: %v", i, err))
			return err
		}

		jw.logger.Trace(fmt.Sprintf("[processData] Line %d: event=%s timestamp=%v", i, base.Event, base.Timestamp))

		if base.Timestamp.Before(jw.lastTimestamp) {
			jw.logger.Trace(fmt.Sprintf("[processData] Line %d: timestamp %v before lastTimestamp %v, skipping", i, base.Timestamp, jw.lastTimestamp))
			continue
		}
		jw.lastTimestamp = base.Timestamp

		switch base.Event {
		case Loadout:
			var event LoadoutEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace("[processData] Publishing Loadout")
				jw.Loadout.Publish(&event)
			}
		case FSDJump:
			var event FSDJumpEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace("[processData] Publishing FSDJump")
				jw.FSDJump.Publish(&event)
			}
		case FSDTarget:
			var event FSDTargetEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace(fmt.Sprintf("[processData] Publishing FSDTarget: %s", event.Name))
				jw.FSDTarget.Publish(&event)
			} else {
				jw.logger.Trace(fmt.Sprintf("[processData] FSDTarget unmarshal error: %v", err))
			}
		case Location:
			var event LocationEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				jw.logger.Trace("[processData] Publishing Location")
				jw.Location.Publish(&event)
			}
		}
	}

	return nil
}

// Looks over all logs at and after the provided timestamp and emits configured
// events from these
func (jw *Watcher) Sync(since time.Time) error {
	jw.logger.Trace(fmt.Sprintf("[Sync] Called with since=%v", since))
	if jw.started {
		return errors.New("Cannot call journal.Watcher.Sync() after the watcher has been started")
	}

	entries, err := os.ReadDir(jw.dir)
	if err != nil {
		return err
	}

	journals := make([]*JournalName, 0, 16)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !journalFilePattern.MatchString(entry.Name()) {
			continue
		}
		name, err := parseJournalName(entry.Name())
		if err != nil {
			return err
		}
		journals = append(journals, name)
	}

	slices.SortFunc(journals, func(a, b *JournalName) int {
		if timeDiff := a.time.Compare(b.time); timeDiff != 0 {
			return timeDiff
		}
		return a.part - b.part
	})

	jw.logger.Trace(fmt.Sprintf("[Sync] Found %d journals", len(journals)))
	for i, j := range journals {
		jw.logger.Trace(fmt.Sprintf("[Sync]   %d: %s", i, j.String()))
	}

	if len(journals) == 0 {
		jw.logger.Trace("[Sync] No journals found, returning")
		return nil
	}

	// TODO: Look for optimization potential
	//  - Maybe [Heading entry](https://elite-journal.readthedocs.io/en/latest/File%20Format.html#heading-entry)
	//
	// Since the timestamp in the journal's filename is from when it was started,
	// and we get the first one AFTER the provided timestamp 'since', we need to
	// go back to the journals with next recent timestamp, as these might include
	// event that are after our provided 'since'
	// Regardless, since each timestamp of every event handled by 'processData'
	// is checked against 'since', we'll never process event's we should not.
	cutoff := slices.IndexFunc(journals, func(j *JournalName) bool { return j.time.After(since) })
	jw.logger.Trace(fmt.Sprintf("[Sync] IndexFunc returned: %d", cutoff))
	if cutoff < 0 {
		cutoff = len(journals) - 1
		jw.logger.Trace(fmt.Sprintf("[Sync] No journal after since, cutoff set to last: %d", cutoff))
	} else if cutoff > 0 {
		cutoff--
		jw.logger.Trace(fmt.Sprintf("[Sync] Decremented cutoff to: %d", cutoff))
	}
	jw.logger.Trace(fmt.Sprintf("[Sync] Before part adjustment, cutoff=%d, journal part=%d", cutoff, journals[cutoff].part))
	cutoff -= journals[cutoff].part - 1
	jw.logger.Trace(fmt.Sprintf("[Sync] After part adjustment, cutoff=%d", cutoff))
	// If journal files have been deleted the first journal might be a part > 1
	if cutoff < 0 {
		jw.logger.Trace("[Sync] Cutoff negative, setting to 0")
		cutoff = 0
	}

	journals = journals[cutoff:]
	jw.logger.Trace(fmt.Sprintf("[Sync] Processing %d journals starting from index %d", len(journals), cutoff))

	jw.lastTimestamp = since
	for i, journal := range journals {
		jw.logger.Trace(fmt.Sprintf("[Sync] Processing journal %d: %s", i, journal.name))
		content, err := os.ReadFile(path.Join(jw.dir, journal.name))
		if err != nil {
			return err
		}
		jw.logger.Trace(fmt.Sprintf("[Sync] Read %d bytes from %s", len(content), journal.name))
		err = jw.processData(content)
		if err != nil {
			return err
		}
	}

	jw.logger.Trace("[Sync] Complete")
	return nil
}

type JournalName struct {
	name string
	time time.Time
	part int
}

func (j *JournalName) String() string {
	return fmt.Sprintf("%s (time: %s, part: %d)", j.name, j.time.Format("2006-01-02T15:04:05"), j.part)
}

func parseJournalName(name string) (*JournalName, error) {
	matches := journalFilePattern.FindStringSubmatch(name)
	if matches == nil {
		return nil, errors.New("File name is not a journal name")
	}

	timestamp, err := time.Parse("2006-01-02T150405", matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	part, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse part number: %w", err)
	}

	return &JournalName{
		name: name,
		time: timestamp,
		part: part,
	}, nil
}
